package vector

type Filter struct {
	Scope  string
	Tags   []string
	Type   string
}

type Result struct {
	ID       string
	Score    float64
	Metadata map[string]any
}

type VectorStore interface {
	Upsert(id string, vector []float32, metadata map[string]any) error
	Search(query []float32, topK int, filter *Filter) ([]Result, error)
	Delete(id string) error
	Close() error
}

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
