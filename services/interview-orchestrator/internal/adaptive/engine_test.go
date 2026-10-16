package adaptive

import (
	"context"
	"testing"

	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/model"
)

type fixedEmbeddingClient struct {
	vec []float32
	err error
}

func (f fixedEmbeddingClient) Embed(ctx context.Context, text string) ([]float32, error) {
	return f.vec, f.err
}

// dynamicEmbeddingClient returns vector by lookup.
type dynamicEmbeddingClient struct {
	embeddings map[string][]float32
}

func (d dynamicEmbeddingClient) Embed(ctx context.Context, text string) ([]float32, error) {
	if vec, ok := d.embeddings[text]; ok {
		return vec, nil
	}
	return make([]float32, 2), nil
}

func TestUpdateCoverage_BackwardCompatibility(t *testing.T) {
	client := fixedEmbeddingClient{vec: []float32{1, 0}}
	engine := NewEngine(client)

	criteriaEmbeddings := map[string][]float32{
		"c1": {1, 0}, // sim = 1.0
		"c2": {0, 1}, // sim = 0.0
	}
	currentCoverage := map[string]float64{
		"c1": 0.0,
		"c2": 0.5,
	}

	updated, err := engine.UpdateCoverage(context.Background(), "my answer", criteriaEmbeddings, currentCoverage)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// c1: sim = 1.0 >= 0.40 -> added = (1.0 - 0.20)*0.35 = 0.28.
	if updated["c1"] < 0.27 || updated["c1"] > 0.29 {
		t.Errorf("expected c1 coverage ~0.28, got %f", updated["c1"])
	}

	// c2: sim = 0.0 < 0.40 -> retains 0.5 (does NOT decay to 0.425!)
	if updated["c2"] != 0.5 {
		t.Errorf("expected c2 coverage to retain 0.5, got %f", updated["c2"])
	}
}

func TestTopicSwitching_NonDecayingCoverage(t *testing.T) {
	dyn := dynamicEmbeddingClient{
		embeddings: map[string][]float32{
			"answer_c1": {1, 0},
			"answer_c2": {0, 1},
		},
	}
	engine := NewEngine(dyn)

	critEmbeddings := map[string][]float32{
		"c1": {1, 0},
		"c2": {0, 1},
	}

	state := make(map[string]CriterionState)

	// Turn 1: Candidate speaks thoroughly about C1
	state, err := engine.UpdateState(context.Background(), "answer_c1", critEmbeddings, state)
	if err != nil {
		t.Fatal(err)
	}

	c1Turn1Coverage := state["c1"].TopicCoverage
	if c1Turn1Coverage <= 0.20 {
		t.Fatalf("expected C1 coverage to increase, got %f", c1Turn1Coverage)
	}

	// Turn 2: Candidate switches completely to C2
	state, err = engine.UpdateState(context.Background(), "answer_c2", critEmbeddings, state)
	if err != nil {
		t.Fatal(err)
	}

	// Crucial check: C1 coverage MUST NOT decay when C2 is discussed!
	if state["c1"].TopicCoverage != c1Turn1Coverage {
		t.Errorf("topic switching caused C1 coverage to change! Before: %f, After: %f",
			c1Turn1Coverage, state["c1"].TopicCoverage)
	}

	// Recency of C1 should have decayed because C2 was discussed
	if state["c1"].Recency >= 0.30 {
		t.Errorf("expected C1 recency to decay when switching to C2, got %f", state["c1"].Recency)
	}
}

func TestIrrelevantAnswers_NoCoverageGrowth(t *testing.T) {
	dyn := dynamicEmbeddingClient{
		embeddings: map[string][]float32{
			"irrelevant": {0.1, 0.1}, // very low similarity to orthogonal basis
		},
	}
	engine := NewEngine(dyn)

	critEmbeddings := map[string][]float32{
		"c1": {1, 0},
		"c2": {0, 1},
	}

	state := map[string]CriterionState{
		"c1": {TopicCoverage: 0.3, DemonstratedCompetence: 0.2},
		"c2": {TopicCoverage: 0.4, DemonstratedCompetence: 0.3},
	}

	updated, err := engine.UpdateState(context.Background(), "irrelevant", critEmbeddings, state)
	if err != nil {
		t.Fatal(err)
	}

	if updated["c1"].TopicCoverage != 0.3 || updated["c2"].TopicCoverage != 0.4 {
		t.Errorf("irrelevant answer should not increase topic coverage: %+v", updated)
	}
	if updated["c1"].DemonstratedCompetence != 0.2 || updated["c2"].DemonstratedCompetence != 0.3 {
		t.Errorf("irrelevant answer should not increase competence: %+v", updated)
	}
}

func TestRepeatedMentionsVsSingleMention(t *testing.T) {
	dyn := dynamicEmbeddingClient{
		embeddings: map[string][]float32{
			"strong_turn_1": {1, 0},
			"strong_turn_2": {1, 0},
		},
	}
	engine := NewEngine(dyn)

	critEmbeddings := map[string][]float32{
		"c1": {1, 0},
	}

	state := make(map[string]CriterionState)

	// Single turn with perfect similarity (sim = 1.0)
	state, err := engine.UpdateState(context.Background(), "strong_turn_1", critEmbeddings, state)
	if err != nil {
		t.Fatal(err)
	}

	// Even with sim = 1.0, a single turn MUST be capped at SingleMentionCap (0.50)
	if state["c1"].DemonstratedCompetence > SingleMentionCap {
		t.Errorf("single turn demonstrated competence must be capped at %v, got %f",
			SingleMentionCap, state["c1"].DemonstratedCompetence)
	}

	// Second turn with strong similarity confirms sustained mastery
	state, err = engine.UpdateState(context.Background(), "strong_turn_2", critEmbeddings, state)
	if err != nil {
		t.Fatal(err)
	}

	if state["c1"].DemonstratedCompetence <= SingleMentionCap {
		t.Errorf("repeated substantive turns should exceed single-mention cap, got %f",
			state["c1"].DemonstratedCompetence)
	}
	if state["c1"].DemonstratedCompetence < 0.70 {
		t.Errorf("two strong substantive turns should achieve competence >= 0.70, got %f",
			state["c1"].DemonstratedCompetence)
	}
}

func TestBalancedInterview(t *testing.T) {
	criteria := []model.RubricCriterion{
		{ID: "c1", Name: "Architecture"},
		{ID: "c2", Name: "Data"},
		{ID: "c3", Name: "Streaming"},
	}

	dyn := dynamicEmbeddingClient{
		embeddings: map[string][]float32{
			"ans_c1": {1, 0, 0},
			"ans_c2": {0, 1, 0},
			"ans_c3": {0, 0, 1},
		},
	}
	engine := NewEngine(dyn)

	critEmbeddings := map[string][]float32{
		"c1": {1, 0, 0},
		"c2": {0, 1, 0},
		"c3": {0, 0, 1},
	}

	state := make(map[string]CriterionState)

	// Balanced interview: 3 turns per topic
	for turn := 1; turn <= 9; turn++ {
		var ansKey string
		if turn <= 3 {
			ansKey = "ans_c1"
		} else if turn <= 6 {
			ansKey = "ans_c2"
		} else {
			ansKey = "ans_c3"
		}
		var err error
		state, err = engine.UpdateState(context.Background(), ansKey, critEmbeddings, state)
		if err != nil {
			t.Fatal(err)
		}
	}

	// All 3 criteria must reach CoveredThreshold (>= 0.70)
	for _, c := range criteria {
		cov := state[c.ID].TopicCoverage
		if cov < CoveredThreshold {
			t.Errorf("criterion %s failed to reach CoveredThreshold in balanced interview: %f", c.ID, cov)
		}
	}

	// When all covered, NextDirectiveFromState returns ("", "")
	name, _ := engine.NextDirectiveFromState(state, criteria)
	if name != "" {
		t.Errorf("expected empty directive when all criteria covered, got %s", name)
	}
}

func TestNextDirective(t *testing.T) {
	engine := NewEngine(fixedEmbeddingClient{})

	criteria := []model.RubricCriterion{
		{ID: "c1", Name: "Architecture", Description: "System design"},
		{ID: "c2", Name: "Coding", Description: "Data structures"},
		{ID: "c3", Name: "Testing", Description: "Unit testing"},
	}

	coverage := map[string]float64{
		"c1": 0.6,
		"c2": 0.2,
		"c3": 0.8,
	}

	name, desc := engine.NextDirective(coverage, criteria)
	if name != "Coding" || desc != "Data structures" {
		t.Errorf("expected NextDirective to pick Coding (largest deficit), got %s (%s)", name, desc)
	}

	allCovered := map[string]float64{
		"c1": 0.75,
		"c2": 0.80,
		"c3": 0.70,
	}
	name, desc = engine.NextDirective(allCovered, criteria)
	if name != "" || desc != "" {
		t.Errorf("expected empty directive when all covered, got %s", name)
	}
}
