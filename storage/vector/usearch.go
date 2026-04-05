package vector

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"localmemory/core"
)

// USearchConfig holds USearch configuration.
type USearchConfig struct {
	Path        string
	VectorSize int
	Metric     string
}

// USearchStore is an in-memory vector store implementation using pure Go (no external dependencies).
// For production, Qdrant is recommended.
type USearchStore struct {
	mu         sync.RWMutex
	vectors    map[string][]float32
	metadata   map[string]map[string]any
	vectorSize int
}

func NewUSearchStore(cfg any) (*USearchStore, error) {
	c, ok := cfg.(USearchConfig)
	if !ok {
		return nil, fmt.Errorf("invalid usearch config")
	}

	dir := filepath.Dir(c.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	return &USearchStore{
		vectors:    make(map[string][]float32),
		metadata:   make(map[string]map[string]any),
		vectorSize: c.VectorSize,
	}, nil
}

func (s *USearchStore) Upsert(id string, vector []float32, metadata map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(vector) != s.vectorSize {
		return fmt.Errorf("vector size mismatch: got %d, want %d", len(vector), s.vectorSize)
	}

	vecCopy := make([]float32, len(vector))
	copy(vecCopy, vector)
	s.vectors[id] = vecCopy

	metaCopy := make(map[string]any)
	for k, v := range metadata {
		metaCopy[k] = v
	}
	s.metadata[id] = metaCopy

	return nil
}

func (s *USearchStore) Search(query []float32, topK int, filter *Filter) ([]Result, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(query) != s.vectorSize {
		return nil, fmt.Errorf("query vector size mismatch: got %d, want %d", len(query), s.vectorSize)
	}

	var scores []scoredID
	for id, vec := range s.vectors {
		meta := s.metadata[id]
		if filter != nil {
			if filter.Scope != "" {
				if scope, ok := meta["scope"].(string); !ok || scope != filter.Scope {
					continue
				}
			}
			if filter.Type != "" {
				if typ, ok := meta["type"].(string); !ok || typ != filter.Type {
					continue
				}
			}
			if filter.ProfileID != "" {
				if pid, ok := meta["profile_id"].(string); !ok || pid != filter.ProfileID {
					continue
				}
			}
		}

		score := cosineSimilarity(query, vec)
		scores = append(scores, scoredID{id, score})
	}

	quickSort(scores)

	if topK > len(scores) {
		topK = len(scores)
	}

	results := make([]Result, topK)
	for i := 0; i < topK; i++ {
		id := scores[i].id
		results[i] = Result{
			ID:       id,
			Score:    scores[i].score,
			Metadata: s.metadata[id],
		}
	}

	return results, nil
}

func (s *USearchStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.vectors, id)
	delete(s.metadata, id)
	return nil
}

func (s *USearchStore) Close() error {
	return nil
}

func cosineSimilarity(a, b []float32) float64 {
	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (normA * normB)
}

type scoredID struct {
	id    string
	score float64
}

func quickSort(scores []scoredID) {
	if len(scores) <= 1 {
		return
	}

	pivot := scores[len(scores)/2]
	i, j := 0, len(scores)-1

	for i <= j {
		for scores[i].score > pivot.score {
			i++
		}
		for scores[j].score < pivot.score {
			j--
		}
		if i <= j {
			scores[i], scores[j] = scores[j], scores[i]
			i++
			j--
		}
	}

	if j > 0 {
		quickSort(scores[:j+1])
	}
	if i < len(scores) {
		quickSort(scores[i:])
	}
}

// Ensure core import is used (for compatibility)
var _ = core.Memory{}
