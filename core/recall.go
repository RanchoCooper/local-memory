package core

// Recall handles memory retrieval.
// Responsible for retrieving memories from vector store and SQLite.
type Recall struct {
	store        *Store
	sqliteStore  SQLiteStoreInterface
	vectorStore  VectorStore
	embeddingSvc EmbeddingService
	ranker       *Ranker
}

// NewRecall creates a Recall instance.
func NewRecall(sqliteStore SQLiteStoreInterface, vectorStore VectorStore, embeddingSvc EmbeddingService, ranker *Ranker) *Recall {
	return &Recall{
		store:        NewStore(sqliteStore, vectorStore, embeddingSvc),
		sqliteStore:  sqliteStore,
		vectorStore:  vectorStore,
		embeddingSvc: embeddingSvc,
		ranker:       ranker,
	}
}

// Query performs semantic search.
// Flow: Query Text → Embedding → Vector Search → Ranking → TopK Results
func (r *Recall) Query(req *QueryRequest) (*QueryResponse, error) {
	// Return error if no embedding service
	if r.embeddingSvc == nil {
		return nil, ErrEmbeddingServiceRequired
	}

	// Return error if no vector store
	if r.vectorStore == nil {
		return nil, ErrVectorStoreRequired
	}

	// 1. Convert query text to vector
	queryVector, err := r.embeddingSvc.Embed(req.Query)
	if err != nil {
		return nil, err
	}

	// 2. Build filter conditions
	filter := &VectorFilter{
		Scope: string(req.Scope),
		Tags:  req.Tags,
	}

	// 3. Vector search
	vectorResults, err := r.vectorStore.Search(queryVector, req.TopK, filter)
	if err != nil {
		return nil, err
	}

	// 4. Get complete memories from SQLite and rank
	var results []*QueryResult
	for _, vr := range vectorResults {
		memory, err := r.sqliteStore.GetByID(vr.ID)
		if err != nil {
			continue
		}
		if memory == nil || memory.Deleted {
			continue
		}

		// Calculate final score
		finalScore := r.ranker.CalcScore(vr.Score, memory)

		results = append(results, &QueryResult{
			Memory: memory,
			Score:  finalScore,
		})
	}

	// 5. Sort by score descending
	r.ranker.ScoreSort(results)

	// Limit result count
	if len(results) > req.TopK {
		results = results[:req.TopK]
	}

	return &QueryResponse{Results: results}, nil
}

// GetByID retrieves a memory by ID.
func (r *Recall) GetByID(id string) (*Memory, error) {
	return r.sqliteStore.GetByID(id)
}

// List lists memories.
func (r *Recall) List(req *ListRequest) (*ListResponse, error) {
	// Set defaults
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	memories, total, err := r.sqliteStore.List(req)
	if err != nil {
		return nil, err
	}

	return &ListResponse{
		Memories: memories,
		Total:    total,
	}, nil
}
