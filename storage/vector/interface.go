package vector

// Filter represents filter conditions for vector search.
type Filter struct {
	Scope string
	Type  string
	Tags  []string
}

// Result represents a vector search result.
type Result struct {
	ID       string
	Score    float64
	Metadata map[string]any
}

// VectorStore is the interface for vector storage implementations.
type VectorStore interface {
	Upsert(id string, vector []float32, metadata map[string]any) error
	Search(query []float32, topK int, filter *Filter) ([]Result, error)
	Delete(id string) error
	Close() error
}

// NewVectorStore creates a vector store instance based on type.
func NewVectorStore(storeType string, cfg any) (VectorStore, error) {
	switch storeType {
	case "qdrant":
		return NewQdrantStore(cfg)
	case "usearch":
		return NewUSearchStore(cfg)
	default:
		return NewQdrantStore(cfg)
	}
}
