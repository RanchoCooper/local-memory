package core

import (
	"fmt"
	"math"
)

// ConsolidationAction represents the action taken during consolidation.
type ConsolidationAction string

const (
	ConsolidationActionNOOP   ConsolidationAction = "NOOP"   // Skip - exact duplicate
	ConsolidationActionUPDATE ConsolidationAction = "UPDATE" // Merge with existing
	ConsolidationActionADD   ConsolidationAction = "ADD"   // Create new entry
)

// ConsolidationResult contains the result of consolidation decision.
type ConsolidationResult struct {
	Action      ConsolidationAction
	Similarity  float64
	ExistingMem *Memory
	NewMem      *Memory
}

// Consolidator handles memory deduplication.
type Consolidator struct {
	embeddingSvc EmbeddingService
}

// NewConsolidator creates a Consolidator.
func NewConsolidator(embeddingSvc EmbeddingService) *Consolidator {
	return &Consolidator{
		embeddingSvc: embeddingSvc,
	}
}

// ComputeSimilarity computes cosine similarity between two texts.
func (c *Consolidator) ComputeSimilarity(text1, text2 string) (float64, error) {
	vec1, err := c.embeddingSvc.Embed(text1)
	if err != nil {
		return 0, err
	}

	vec2, err := c.embeddingSvc.Embed(text2)
	if err != nil {
		return 0, err
	}

	return CosineSimilarity(vec1, vec2), nil
}

// CosineSimilarity computes cosine similarity between two vectors.
func CosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// DecideConsolidation decides what action to take for a new memory.
func (c *Consolidator) DecideConsolidation(existing, new *Memory) (*ConsolidationResult, error) {
	// If no embedding service, fall back to UPDATE by default
	if c.embeddingSvc == nil {
		return &ConsolidationResult{
			Action:      ConsolidationActionUPDATE,
			Similarity:  0.0,
			ExistingMem: existing,
			NewMem:      new,
		}, nil
	}

	// Compute similarity
	similarity, err := c.ComputeSimilarity(existing.Value, new.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to compute similarity: %w", err)
	}

	result := &ConsolidationResult{
		Similarity:  similarity,
		ExistingMem: existing,
		NewMem:      new,
	}

	// Decide action based on thresholds
	// Note: NOOP is only used for EXACT duplicate content (similarity ~1.0 with same text)
	// Since different texts rarely have similarity > 0.95, we treat as UPDATE for safety
	// This preserves the original evolve behavior: always merge when key matches
	switch {
	case similarity > 0.95 && existing.Value == new.Value:
		// Exact duplicate (same text) - NOOP
		result.Action = ConsolidationActionNOOP
	case similarity > 0.80:
		// Similar - UPDATE
		result.Action = ConsolidationActionUPDATE
	default:
		// Different fact - ADD with new key
		result.Action = ConsolidationActionADD
	}

	return result, nil
}
