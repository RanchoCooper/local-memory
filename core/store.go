package core

import (
	"fmt"
	"time"
)

// Store handles memory storage.
// Responsible for writing memories to SQLite and vector store.
type Store struct {
	sqliteStore  SQLiteStoreInterface
	vectorStore  VectorStore
	embeddingSvc EmbeddingService
}

// SQLiteStoreInterface is the SQLite storage interface.
// Abstracts SQLite storage implementation, supports testing and replacement.
type SQLiteStoreInterface interface {
	Save(m *Memory) error
	GetByID(id string) (*Memory, error)
	GetByKey(key string) (*Memory, error)
	List(req *ListRequest) ([]*Memory, int, error)
}

// VectorStore is the vector storage interface.
// Abstracts vector storage implementation, supports Qdrant, USearch, etc.
type VectorStore interface {
	Upsert(id string, vector []float32, metadata map[string]any) error
	Search(query []float32, topK int, filter *VectorFilter) ([]VectorResult, error)
	Delete(id string) error
	Close() error
}

// VectorFilter represents vector search filter conditions.
type VectorFilter struct {
	Scope string
	Type  string
	Tags  []string
}

// VectorResult represents a vector search result.
type VectorResult struct {
	ID       string
	Score    float64
	Metadata map[string]any
}

// EmbeddingService is the embedding service interface.
// Abstracts embedding implementation, supports local models or API.
type EmbeddingService interface {
	Embed(text string) ([]float32, error)
	EmbedBatch(texts []string) ([][]float32, error)
}

// NewStore creates a Store instance.
func NewStore(sqliteStore SQLiteStoreInterface, vectorStore VectorStore, embeddingSvc EmbeddingService) *Store {
	return &Store{
		sqliteStore:  sqliteStore,
		vectorStore:  vectorStore,
		embeddingSvc: embeddingSvc,
	}
}

// Save saves a memory.
// 1. Generate vector embedding
// 2. Save to SQLite
// 3. Save to vector store
func (s *Store) Save(m *Memory) error {
	// Check if Evolve is enabled: whether memory with same key already exists
	existing, err := s.sqliteStore.GetByKey(m.Key)
	if err != nil {
		return fmt.Errorf("failed to check existing memory: %w", err)
	}

	if existing != nil {
		// Memory with same key exists, merge via Evolve
		return s.evolveAndSave(existing, m)
	}

	// Generate vector embedding
	if s.embeddingSvc != nil && m.Value != "" {
		vector, err := s.embeddingSvc.Embed(m.Value)
		if err != nil {
			return fmt.Errorf("failed to generate embedding: %w", err)
		}
		m.Embedding = vector
	}

	// Save to SQLite
	if err := s.sqliteStore.Save(m); err != nil {
		return fmt.Errorf("failed to save to sqlite: %w", err)
	}

	// Save to vector store
	if s.vectorStore != nil && len(m.Embedding) > 0 {
		meta := s.memoryToMetadata(m)
		if err := s.vectorStore.Upsert(m.ID, m.Embedding, meta); err != nil {
			return fmt.Errorf("failed to save to vector store: %w", err)
		}
	}

	return nil
}

// evolveAndSave evolves and saves.
// When memory with same key exists, merges the two.
func (s *Store) evolveAndSave(existing, new *Memory) error {
	// Update existing memory's value (append new information)
	existing.Value = existing.Value + "\n" + new.Value

	// Update confidence: take higher value, but not exceeding 1.0
	existing.Confidence = minFloat64(1.0, existing.Confidence+0.1)

	// Update timestamp
	existing.UpdatedAt = time.Now().Unix()

	// Merge tags
	existing.Tags = mergeTags(existing.Tags, new.Tags)

	// Merge related memories
	existing.RelatedIDs = mergeIDs(existing.RelatedIDs, new.RelatedIDs)

	// Regenerate embedding
	if s.embeddingSvc != nil && existing.Value != "" {
		vector, err := s.embeddingSvc.Embed(existing.Value)
		if err != nil {
			return fmt.Errorf("failed to regenerate embedding: %w", err)
		}
		existing.Embedding = vector
	}

	// Save updated memory
	if err := s.sqliteStore.Save(existing); err != nil {
		return fmt.Errorf("failed to save evolved memory: %w", err)
	}

	// Update vector store
	if s.vectorStore != nil && len(existing.Embedding) > 0 {
		meta := s.memoryToMetadata(existing)
		if err := s.vectorStore.Upsert(existing.ID, existing.Embedding, meta); err != nil {
			return fmt.Errorf("failed to update vector store: %w", err)
		}
	}

	return nil
}

// memoryToMetadata converts Memory to vector store metadata.
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

// mergeTags merges two tag lists.
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

// mergeIDs merges two ID lists.
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

// minFloat64 returns the smaller of two floats.
func minFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
