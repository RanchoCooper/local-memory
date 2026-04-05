package unit

import (
	"math"
	"testing"

	"localmemory/core"
)

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a        []float32
		b        []float32
		expected float64
	}{
		{
			name:     "identical vectors",
			a:        []float32{1, 0, 0},
			b:        []float32{1, 0, 0},
			expected: 1.0,
		},
		{
			name:     "orthogonal vectors",
			a:        []float32{1, 0, 0},
			b:        []float32{0, 1, 0},
			expected: 0.0,
		},
		{
			name:     "opposite vectors",
			a:        []float32{1, 0, 0},
			b:        []float32{-1, 0, 0},
			expected: -1.0,
		},
		{
			name:     "2D vectors",
			a:        []float32{1, 1},
			b:        []float32{1, 1},
			expected: 1.0,
		},
		{
			name:     "2D orthogonal",
			a:        []float32{1, 0},
			b:        []float32{0, 1},
			expected: 0.0,
		},
		{
			name:     "partial similarity",
			a:        []float32{1, 1, 1},
			b:        []float32{1, 1, 0},
			expected: 2.0 / (math.Sqrt(3) * math.Sqrt(2)), // ~0.816
		},
		{
			name:     "empty vectors",
			a:        []float32{},
			b:        []float32{},
			expected: 0.0,
		},
		{
			name:     "different lengths vectors",
			a:        []float32{1, 0, 0},
			b:        []float32{1, 0},
			expected: 0.0,
		},
		{
			name:     "zero vector a",
			a:        []float32{0, 0, 0},
			b:        []float32{1, 1, 1},
			expected: 0.0,
		},
		{
			name:     "zero vector b",
			a:        []float32{1, 1, 1},
			b:        []float32{0, 0, 0},
			expected: 0.0,
		},
		{
			name:     "both zero vectors",
			a:        []float32{0, 0, 0},
			b:        []float32{0, 0, 0},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := core.CosineSimilarity(tt.a, tt.b)
			if math.Abs(got-tt.expected) > 1e-9 {
				t.Errorf("cosineSimilarity(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.expected)
			}
		})
	}
}

func TestDecideConsolidation_NoEmbeddingService(t *testing.T) {
	consolidator := core.NewConsolidator(nil)

	existing := &core.Memory{
		ID:    "existing-1",
		Key:   "test_key",
		Value: "existing value",
	}
	newMem := &core.Memory{
		ID:    "new-1",
		Key:   "test_key",
		Value: "new value",
	}

	result, err := consolidator.DecideConsolidation(existing, newMem)
	if err != nil {
		t.Fatalf("DecideConsolidation failed: %v", err)
	}

	if result.Action != core.ConsolidationActionUPDATE {
		t.Errorf("Expected action UPDATE, got %v", result.Action)
	}
	if result.Similarity != 0.0 {
		t.Errorf("Expected similarity 0.0, got %f", result.Similarity)
	}
}

func TestDecideConsolidation_WithMockEmbedding(t *testing.T) {
	mockSvc := &mockEmbeddingService{
		embedFunc: func(text string) ([]float32, error) {
			// Return vectors based on text content to control similarity
			switch text {
			case "same text":
				return []float32{1, 0, 0}, nil // identical to itself
			case "text a":
				return []float32{1, 0, 0}, nil // orthogonal to "text b"
			case "text b":
				return []float32{0, 1, 0}, nil
			case "similar to a":
				return []float32{0.9, 0.1, 0}, nil // high similarity to "text a"
			default:
				return []float32{0, 0, 1}, nil
			}
		},
	}

	consolidator := core.NewConsolidator(mockSvc)

	t.Run("exact duplicate triggers NOOP", func(t *testing.T) {
		existing := &core.Memory{ID: "1", Key: "key", Value: "same text"}
		newMem := &core.Memory{ID: "2", Key: "key", Value: "same text"}

		result, err := consolidator.DecideConsolidation(existing, newMem)
		if err != nil {
			t.Fatalf("DecideConsolidation failed: %v", err)
		}

		if result.Action != core.ConsolidationActionNOOP {
			t.Errorf("Expected action NOOP, got %v", result.Action)
		}
	})

	t.Run("similar text triggers UPDATE", func(t *testing.T) {
		existing := &core.Memory{ID: "1", Key: "key", Value: "text a"}
		newMem := &core.Memory{ID: "2", Key: "key", Value: "similar to a"}

		result, err := consolidator.DecideConsolidation(existing, newMem)
		if err != nil {
			t.Fatalf("DecideConsolidation failed: %v", err)
		}

		if result.Action != core.ConsolidationActionUPDATE {
			t.Errorf("Expected action UPDATE, got %v", result.Action)
		}
		if result.Similarity <= 0.80 {
			t.Errorf("Expected similarity > 0.80, got %f", result.Similarity)
		}
	})

	t.Run("different text triggers ADD", func(t *testing.T) {
		existing := &core.Memory{ID: "1", Key: "key", Value: "text a"}
		newMem := &core.Memory{ID: "2", Key: "key", Value: "text b"}

		result, err := consolidator.DecideConsolidation(existing, newMem)
		if err != nil {
			t.Fatalf("DecideConsolidation failed: %v", err)
		}

		if result.Action != core.ConsolidationActionADD {
			t.Errorf("Expected action ADD, got %v", result.Action)
		}
		if result.Similarity > 0.80 {
			t.Errorf("Expected similarity <= 0.80, got %f", result.Similarity)
		}
	})
}

// mockEmbeddingService is a mock for testing
type mockEmbeddingService struct {
	embedFunc func(text string) ([]float32, error)
}

func (m *mockEmbeddingService) Embed(text string) ([]float32, error) {
	return m.embedFunc(text)
}

func (m *mockEmbeddingService) EmbedBatch(texts []string) ([][]float32, error) {
	result := make([][]float32, len(texts))
	for i, text := range texts {
		vec, err := m.Embed(text)
		if err != nil {
			return nil, err
		}
		result[i] = vec
	}
	return result, nil
}