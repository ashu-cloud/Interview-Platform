// Package adaptive implements the multi-dimensional competency tracking engine.
// It explicitly separates:
// 1. Topic Coverage: Cumulative exploration of the topic across the interview (monotonic, non-decaying).
// 2. Demonstrated Competence: Quality and depth of answers across multiple turns (single mention capped at 0.50).
// 3. Recency of Discussion: Exponentially decaying conversational focus indicating active short-term context.
package adaptive

import (
	"context"
	"fmt"
	"math"

	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/embedding"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/model"
)

const (
	// CoveredThreshold is the score at or above which a criterion is considered adequately explored.
	CoveredThreshold = 0.70

	// RelevanceThreshold is the minimum cosine similarity to contribute to topic coverage.
	RelevanceThreshold = 0.40

	// SubstantiveThreshold is the minimum cosine similarity to count toward demonstrated competence.
	SubstantiveThreshold = 0.55

	// SingleMentionCap is the maximum competence score attainable from a single answer.
	SingleMentionCap = 0.50

	// RecencyDecayFactor is the weight retained for active conversational recency.
	RecencyDecayFactor = 0.75

	// DecayFactor is an alias for RecencyDecayFactor for simulation benchmarks.
	DecayFactor = RecencyDecayFactor

	// CoverageStepFactor determines the accumulation rate per relevant answer.
	CoverageStepFactor = 0.35
)

// CriterionState stores multi-dimensional tracking metrics for a single rubric criterion.
type CriterionState struct {
	// TopicCoverage represents cumulative interview exploration [0.0, 1.0].
	// Monotonic: does not decay when the conversation transitions to other topics.
	TopicCoverage float64 `json:"topic_coverage"`

	// DemonstratedCompetence represents demonstrated technical ability [0.0, 1.0].
	// Requires multiple substantive turns; a single high-similarity answer is capped at SingleMentionCap.
	DemonstratedCompetence float64 `json:"demonstrated_competence"`

	// Recency represents immediate conversational focus [0.0, 1.0].
	// Exponentially decays when the topic is not actively discussed.
	Recency float64 `json:"recency"`

	// SubstantiveTurns counts the number of turns with similarity >= SubstantiveThreshold.
	SubstantiveTurns int `json:"substantive_turns"`

	// CumulativeSim accumulates similarity values across substantive turns.
	CumulativeSim float64 `json:"cumulative_sim"`
}

// Engine tracks per-criterion coverage and competence of participant answers.
type Engine struct {
	embeddingClient embedding.EmbeddingClient
}

// NewEngine creates an Engine backed by the given embedding client.
func NewEngine(client embedding.EmbeddingClient) *Engine {
	return &Engine{embeddingClient: client}
}

// UpdateState embeds the participant answer and computes updated multi-dimensional state.
func (e *Engine) UpdateState(ctx context.Context,
	participantAnswer string,
	criteriaEmbeddings map[string][]float32,
	currentState map[string]CriterionState) (map[string]CriterionState, error) {

	answerVec, err := e.embeddingClient.Embed(ctx, participantAnswer)
	if err != nil {
		return currentState, fmt.Errorf("embed answer: %w", err)
	}

	updated := make(map[string]CriterionState, len(criteriaEmbeddings))
	for id, critVec := range criteriaEmbeddings {
		state := currentState[id]
		sim := float64(embedding.CosineSimilarity(answerVec, critVec))
		if sim < 0.0 {
			sim = 0.0
		} else if sim > 1.0 {
			sim = 1.0
		}

		// 1. Topic Coverage (Monotonic Accumulation)
		// Only relevant answers add to coverage; once covered, it does not evaporate when other topics are discussed.
		if sim >= RelevanceThreshold {
			addedCoverage := (sim - RelevanceThreshold*0.5) * CoverageStepFactor
			state.TopicCoverage += addedCoverage
			if state.TopicCoverage > 1.0 {
				state.TopicCoverage = 1.0
			}
		}

		// 2. Demonstrated Competence (Evidence of Mastery)
		// A single high similarity score does NOT prove mastery (capped at SingleMentionCap).
		// Sustained depth across multiple substantive turns enables higher competence ratings.
		if sim >= SubstantiveThreshold {
			state.SubstantiveTurns++
			state.CumulativeSim += sim
			avgSim := state.CumulativeSim / float64(state.SubstantiveTurns)

			if state.SubstantiveTurns == 1 {
				state.DemonstratedCompetence = math.Min(SingleMentionCap, avgSim*0.55)
			} else {
				depthMultiplier := math.Min(1.0, 0.75+0.15*float64(state.SubstantiveTurns-1))
				state.DemonstratedCompetence = math.Min(1.0, avgSim*depthMultiplier)
			}
		}

		// 3. Recency of Discussion (Decaying Focus)
		state.Recency = state.Recency*RecencyDecayFactor + sim*(1.0-RecencyDecayFactor)
		if state.Recency > 1.0 {
			state.Recency = 1.0
		} else if state.Recency < 0.0 {
			state.Recency = 0.0
		}

		updated[id] = state
	}

	return updated, nil
}

// UpdateCoverage provides backward compatibility for callers tracking Topic Coverage via map[string]float64.
// It preserves monotonic topic coverage so explored topics do NOT vanish when other topics are discussed.
func (e *Engine) UpdateCoverage(ctx context.Context,
	participantAnswer string,
	criteriaEmbeddings map[string][]float32,
	currentCoverage map[string]float64) (map[string]float64, error) {

	answerVec, err := e.embeddingClient.Embed(ctx, participantAnswer)
	if err != nil {
		return currentCoverage, fmt.Errorf("embed answer: %w", err)
	}

	updated := make(map[string]float64, len(criteriaEmbeddings))
	for id, critVec := range criteriaEmbeddings {
		current := currentCoverage[id]
		sim := float64(embedding.CosineSimilarity(answerVec, critVec))
		if sim < 0.0 {
			sim = 0.0
		} else if sim > 1.0 {
			sim = 1.0
		}

		if sim >= RelevanceThreshold {
			addedCoverage := (sim - RelevanceThreshold*0.5) * CoverageStepFactor
			current += addedCoverage
			if current > 1.0 {
				current = 1.0
			}
		}
		updated[id] = current
	}
	return updated, nil
}

// NextDirective returns the evaluation criterion with the largest coverage deficit.
// If all criteria have reached CoveredThreshold, returns ("", "").
// Ties resolve to the earliest criterion in the slice.
func (e *Engine) NextDirective(coverage map[string]float64,
	criteria []model.RubricCriterion) (name, description string) {
	maxDeficit := 0.0
	for _, c := range criteria {
		score := coverage[c.ID]
		if score >= CoveredThreshold {
			continue
		}
		deficit := 1.0 - score
		if deficit > maxDeficit {
			maxDeficit = deficit
			name = c.Name
			description = c.Description
		}
	}
	return name, description
}

// NextDirectiveFromState selects the next directive using multi-dimensional state.
// Prioritizes topic coverage deficits, using demonstrated competence as a tie-breaker.
func (e *Engine) NextDirectiveFromState(states map[string]CriterionState,
	criteria []model.RubricCriterion) (name, description string) {
	maxDeficit := 0.0
	lowestCompetence := 1.0

	for _, c := range criteria {
		state := states[c.ID]
		if state.TopicCoverage >= CoveredThreshold {
			continue
		}
		deficit := 1.0 - state.TopicCoverage
		// Prioritize larger coverage deficit; break ties with lower demonstrated competence
		if deficit > maxDeficit || (math.Abs(deficit-maxDeficit) < 0.05 && state.DemonstratedCompetence < lowestCompetence) {
			maxDeficit = deficit
			lowestCompetence = state.DemonstratedCompetence
			name = c.Name
			description = c.Description
		}
	}
	return name, description
}
