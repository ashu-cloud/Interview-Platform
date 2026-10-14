package adaptive

import (
	"context"
	"math"
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

func TestUpdateCoverage(t *testing.T) {
	// Fixed vector for answer: [1, 0]
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

	// c1: 0.0*0.85 + 1.0*(0.15) = 0.15
	if math.Abs(updated["c1"]-0.15) > 1e-6 {
		t.Errorf("expected c1 coverage 0.15, got %f", updated["c1"])
	}

	// c2: 0.5*0.85 + 0.0*(0.15) = 0.425
	if math.Abs(updated["c2"]-0.425) > 1e-6 {
		t.Errorf("expected c2 coverage 0.425, got %f", updated["c2"])
	}
}

func TestNextDirective(t *testing.T) {
	engine := NewEngine(fixedEmbeddingClient{})

	criteria := []model.RubricCriterion{
		{ID: "c1", Name: "Architecture", Description: "System design"},
		{ID: "c2", Name: "Coding", Description: "Data structures"},
		{ID: "c3", Name: "Testing", Description: "Unit testing"},
	}

	// c1 has score 0.6 (deficit 0.4), c2 has score 0.2 (deficit 0.8), c3 has score 0.8 (covered >= 0.7)
	coverage := map[string]float64{
		"c1": 0.6,
		"c2": 0.2,
		"c3": 0.8,
	}

	name, desc := engine.NextDirective(coverage, criteria)
	if name != "Coding" || desc != "Data structures" {
		t.Errorf("expected NextDirective to pick Coding (largest deficit), got %s (%s)", name, desc)
	}

	// When all criteria >= CoveredThreshold (0.70)
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
