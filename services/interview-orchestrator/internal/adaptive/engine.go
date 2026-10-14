// Package adaptive implements the deterministic coverage engine that decides
// which evaluation criterion the assessor should probe next.
package adaptive

import (
	"context"
	"fmt"

	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/embedding"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/model"
)

const (
	// DecayFactor is the weight kept from the previous coverage score each turn.
	DecayFactor = 0.85
	// CoveredThreshold is the score at or above which a criterion counts as covered.
	CoveredThreshold = 0.70
)

// Engine tracks per-criterion coverage of the participant's answers.
type Engine struct {
	embeddingClient embedding.EmbeddingClient
}

// NewEngine creates an Engine backed by the given embedding client.
func NewEngine(client embedding.EmbeddingClient) *Engine {
	return &Engine{embeddingClient: client}
}

// UpdateCoverage embeds the participant's answer and updates the coverage matrix.
// On embedding failure, returns the unchanged matrix (graceful degradation).
func (e *Engine) UpdateCoverage(ctx context.Context,
	participantAnswer string,
	criteriaEmbeddings map[string][]float32,
	currentCoverage map[string]float64) (map[string]float64, error) {

	answerVec, err := e.embeddingClient.Embed(ctx, participantAnswer)
	if err != nil {
		return currentCoverage, fmt.Errorf("embed answer: %w", err)
	}

	updated := make(map[string]float64, len(currentCoverage))
	for id, current := range currentCoverage {
		critVec, ok := criteriaEmbeddings[id]
		if !ok {
			updated[id] = current
			continue
		}
		sim := embedding.CosineSimilarity(answerVec, critVec)
		score := current*DecayFactor + sim*(1.0-DecayFactor)
		if score > 1.0 {
			score = 1.0
		}
		if score < 0.0 {
			score = 0.0
		}
		updated[id] = score
	}
	return updated, nil
}

// NextDirective returns the evaluation criterion with the largest coverage deficit.
// Returns ("", "") if all criteria are at or above CoveredThreshold.
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
