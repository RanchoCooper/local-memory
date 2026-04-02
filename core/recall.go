package core

// Recall 记忆检索模块
// 负责从向量存储和 SQLite 检索记忆
type Recall struct {
	store        *Store
	sqliteStore  SQLiteStoreInterface
	vectorStore  VectorStore
	embeddingSvc EmbeddingService
	ranker       *Ranker
}

// NewRecall 创建 Recall 实例
func NewRecall(sqliteStore SQLiteStoreInterface, vectorStore VectorStore, embeddingSvc EmbeddingService, ranker *Ranker) *Recall {
	return &Recall{
		store:        NewStore(sqliteStore, vectorStore, embeddingSvc),
		sqliteStore:  sqliteStore,
		vectorStore:  vectorStore,
		embeddingSvc: embeddingSvc,
		ranker:       ranker,
	}
}

// Query 语义检索
// 流程：Query Text → Embedding → Vector Search → Ranking → TopK Results
func (r *Recall) Query(req *QueryRequest) (*QueryResponse, error) {
	// 如果没有 embedding 服务，返回错误
	if r.embeddingSvc == nil {
		return nil, ErrEmbeddingServiceRequired
	}

	// 如果没有向量存储，返回错误
	if r.vectorStore == nil {
		return nil, ErrVectorStoreRequired
	}

	// 1. 将查询文本转换为向量
	queryVector, err := r.embeddingSvc.Embed(req.Query)
	if err != nil {
		return nil, err
	}

	// 2. 构建过滤条件
	filter := &VectorFilter{
		Scope: string(req.Scope),
		Tags:  req.Tags,
	}

	// 3. 向量搜索
	vectorResults, err := r.vectorStore.Search(queryVector, req.TopK, filter)
	if err != nil {
		return nil, err
	}

	// 4. 从 SQLite 获取完整记忆并排序
	var results []*QueryResult
	for _, vr := range vectorResults {
		memory, err := r.sqliteStore.GetByID(vr.ID)
		if err != nil {
			continue
		}
		if memory == nil || memory.Deleted {
			continue
		}

		// 计算最终得分
		finalScore := r.ranker.CalcScore(vr.Score, memory)

		results = append(results, &QueryResult{
			Memory: memory,
			Score:  finalScore,
		})
	}

	// 5. 按得分降序排序
	r.ranker.ScoreSort(results)

	// 限制返回数量
	if len(results) > req.TopK {
		results = results[:req.TopK]
	}

	return &QueryResponse{Results: results}, nil
}

// GetByID 根据 ID 获取记忆
func (r *Recall) GetByID(id string) (*Memory, error) {
	return r.sqliteStore.GetByID(id)
}

// List 列出记忆
func (r *Recall) List(req *ListRequest) (*ListResponse, error) {
	// 设置默认值
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
