package core

import (
	"fmt"
	"time"
)

// Store 记忆存储模块
// 负责将记忆写入 SQLite 和向量存储
type Store struct {
	sqliteStore  SQLiteStoreInterface
	vectorStore  VectorStore
	embeddingSvc EmbeddingService
}

// SQLiteStoreInterface SQLite 存储接口
// 抽象 SQLite 存储实现，支持测试和替换
type SQLiteStoreInterface interface {
	Save(m *Memory) error
	GetByID(id string) (*Memory, error)
	GetByKey(key string) (*Memory, error)
	List(req *ListRequest) ([]*Memory, int, error)
}

// VectorStore 向量存储接口
// 抽象向量存储实现，支持 Qdrant、USearch 等
type VectorStore interface {
	Upsert(id string, vector []float32, metadata map[string]any) error
	Search(query []float32, topK int, filter *VectorFilter) ([]VectorResult, error)
	Delete(id string) error
	Close() error
}

// VectorFilter 向量搜索过滤条件
type VectorFilter struct {
	Scope string
	Type  string
	Tags  []string
}

// VectorResult 向量搜索结果
type VectorResult struct {
	ID       string
	Score    float64
	Metadata map[string]any
}

// EmbeddingService 嵌入服务接口
// 抽象 Embedding 实现，支持本地模型或 API
type EmbeddingService interface {
	Embed(text string) ([]float32, error)
	EmbedBatch(texts []string) ([][]float32, error)
}

// NewStore 创建 Store 实例
func NewStore(sqliteStore SQLiteStoreInterface, vectorStore VectorStore, embeddingSvc EmbeddingService) *Store {
	return &Store{
		sqliteStore:  sqliteStore,
		vectorStore:  vectorStore,
		embeddingSvc: embeddingSvc,
	}
}

// Save 保存记忆
// 1. 生成向量嵌入
// 2. 保存到 SQLite
// 3. 保存到向量存储
func (s *Store) Save(m *Memory) error {
	// 检查是否启用 Evolve：同 key 记忆是否已存在
	existing, err := s.sqliteStore.GetByKey(m.Key)
	if err != nil {
		return fmt.Errorf("failed to check existing memory: %w", err)
	}

	if existing != nil {
		// 同 key 记忆已存在，调用 Evolve 合并
		return s.evolveAndSave(existing, m)
	}

	// 生成向量嵌入
	if s.embeddingSvc != nil && m.Value != "" {
		vector, err := s.embeddingSvc.Embed(m.Value)
		if err != nil {
			return fmt.Errorf("failed to generate embedding: %w", err)
		}
		m.Embedding = vector
	}

	// 保存到 SQLite
	if err := s.sqliteStore.Save(m); err != nil {
		return fmt.Errorf("failed to save to sqlite: %w", err)
	}

	// 保存到向量存储
	if s.vectorStore != nil && len(m.Embedding) > 0 {
		meta := s.memoryToMetadata(m)
		if err := s.vectorStore.Upsert(m.ID, m.Embedding, meta); err != nil {
			return fmt.Errorf("failed to save to vector store: %w", err)
		}
	}

	return nil
}

// evolveAndSave 进化并保存
// 当同 key 记忆存在时，合并两者
func (s *Store) evolveAndSave(existing, new *Memory) error {
	// 更新已有记忆的 value（追加新信息）
	existing.Value = existing.Value + "\n" + new.Value

	// 更新置信度：取较高值，但不超过 1.0
	existing.Confidence = minFloat64(1.0, existing.Confidence+0.1)

	// 更新时间戳
	existing.UpdatedAt = time.Now().Unix()

	// 合并标签
	existing.Tags = mergeTags(existing.Tags, new.Tags)

	// 合并关联记忆
	existing.RelatedIDs = mergeIDs(existing.RelatedIDs, new.RelatedIDs)

	// 重新生成嵌入
	if s.embeddingSvc != nil && existing.Value != "" {
		vector, err := s.embeddingSvc.Embed(existing.Value)
		if err != nil {
			return fmt.Errorf("failed to regenerate embedding: %w", err)
		}
		existing.Embedding = vector
	}

	// 保存更新后的记忆
	if err := s.sqliteStore.Save(existing); err != nil {
		return fmt.Errorf("failed to save evolved memory: %w", err)
	}

	// 更新向量存储
	if s.vectorStore != nil && len(existing.Embedding) > 0 {
		meta := s.memoryToMetadata(existing)
		if err := s.vectorStore.Upsert(existing.ID, existing.Embedding, meta); err != nil {
			return fmt.Errorf("failed to update vector store: %w", err)
		}
	}

	return nil
}

// memoryToMetadata 将 Memory 转换为向量存储的 metadata
func (s *Store) memoryToMetadata(m *Memory) map[string]any {
	return map[string]any{
		"id":         m.ID,
		"key":        m.Key,
		"type":       string(m.Type),
		"scope":      string(m.Scope),
		"media_type": string(m.MediaType),
		"tags":       m.Tags,
	}
}

// mergeTags 合并标签列表
func mergeTags(a, b []string) []string {
	tagSet := make(map[string]bool)
	for _, tag := range a {
		tagSet[tag] = true
	}
	for _, tag := range b {
		tagSet[tag] = true
	}
	result := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		result = append(result, tag)
	}
	return result
}

// mergeIDs 合并 ID 列表
func mergeIDs(a, b []string) []string {
	idSet := make(map[string]bool)
	for _, id := range a {
		idSet[id] = true
	}
	for _, id := range b {
		idSet[id] = true
	}
	result := make([]string, 0, len(idSet))
	for id := range idSet {
		result = append(result, id)
	}
	return result
}

// minFloat64 返回两个浮点数的较小值
func minFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
