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
	GetByKey(key string, profileID string) (*Memory, error)
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
	Scope     string
	Type      string
	Tags      []string
	ProfileID string
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
	// Ensure profile_id is set before any operations
	if m.ProfileID == "" {
		m.ProfileID = "default"
	}

	// Check if Evolve is enabled: whether memory with same key already exists
	existing, err := s.sqliteStore.GetByKey(m.Key, m.ProfileID)
	if err != nil {
		return fmt.Errorf("failed to check existing memory: %w", err)
	}

	if existing != nil {
		// Memory with same key exists, merge via Evolve
		// (m.ProfileID already matches existing.ProfileID since we set it above)
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

// evolveAndSave evolves and saves using smart deduplication.
// When memory with same key exists, computes similarity and decides action.
func (s *Store) evolveAndSave(existing, new *Memory) error {
	// If no embedding service, fall back to simple evolve
	if s.embeddingSvc == nil {
		return s.simpleEvolveAndSave(existing, new)
	}

	// Compute consolidation decision
	consolidator := NewConsolidator(s.embeddingSvc)
	decision, err := consolidator.DecideConsolidation(existing, new)
	if err != nil {
		// Fall back to simple evolve on error
		return s.simpleEvolveAndSave(existing, new)
	}

	switch decision.Action {
	case ConsolidationActionNOOP:
		// Exact duplicate - skip entirely
		return nil

	case ConsolidationActionUPDATE:
		// Merge with existing
		return s.mergeAndSave(existing, new, decision.Similarity)

	case ConsolidationActionADD:
		// Different fact - create new key with suffix
		new.Key = fmt.Sprintf("%s::%d", new.Key, existing.EvidenceCount)
		new.EvidenceCount = 1
		return s.Save(new)
	}

	return nil
}

// simpleEvolveAndSave performs simple merge without similarity check.
func (s *Store) simpleEvolveAndSave(existing, new *Memory) error {
	// Update existing memory's value (append new information, avoid empty append)
	if new.Value != "" {
		if existing.Value != "" {
			existing.Value = existing.Value + "\n" + new.Value
		} else {
			existing.Value = new.Value
		}
	}

	// Update confidence: take higher value, but not exceeding 1.0
	existing.Confidence = minFloat64(1.0, existing.Confidence+0.1)

	// Update timestamp
	existing.UpdatedAt = time.Now().Unix()

	// Merge tags
	existing.Tags = mergeTags(existing.Tags, new.Tags)

	// Merge related memories
	existing.RelatedIDs = mergeIDs(existing.RelatedIDs, new.RelatedIDs)

	// Regenerate embedding only if embedding service is available
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

	// Update vector store only if we have an embedding
	if s.vectorStore != nil && len(existing.Embedding) > 0 {
		meta := s.memoryToMetadata(existing)
		if err := s.vectorStore.Upsert(existing.ID, existing.Embedding, meta); err != nil {
			return fmt.Errorf("failed to update vector store: %w", err)
		}
	}

	return nil
}

// mergeAndSave merges new memory into existing.
func (s *Store) mergeAndSave(existing, new *Memory, similarity float64) error {
	// Append new info to existing value
	if new.Value != "" {
		if existing.Value != "" {
			existing.Value = existing.Value + "\n" + new.Value
		} else {
			existing.Value = new.Value
		}
	}

	// Update confidence: keep max, increment evidence count
	existing.Confidence = minFloat64(1.0, maxFloat64(existing.Confidence, new.Confidence))
	existing.EvidenceCount++

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
		return fmt.Errorf("failed to save merged memory: %w", err)
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

// ExtractAndSave extracts atomic facts and saves them.
// Returns the parent memory and child fact memories.
func (s *Store) ExtractAndSave(m *Memory) (*Memory, []*Memory, error) {
	// Generate embedding for original memory
	if s.embeddingSvc != nil && m.Value != "" {
		vector, err := s.embeddingSvc.Embed(m.Value)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate embedding: %w", err)
		}
		m.Embedding = vector
	}

	// Save the parent memory
	if err := s.sqliteStore.Save(m); err != nil {
		return nil, nil, fmt.Errorf("failed to save parent memory: %w", err)
	}

	// Save to vector store
	if s.vectorStore != nil && len(m.Embedding) > 0 {
		meta := s.memoryToMetadata(m)
		if err := s.vectorStore.Upsert(m.ID, m.Embedding, meta); err != nil {
			return nil, nil, fmt.Errorf("failed to save to vector store: %w", err)
		}
	}

	// Extract atomic facts
	extractor := NewExtractor()
	facts, err := extractor.ExtractFacts(m.Value, m.Key)
	if err != nil {
		return m, nil, nil // Return parent but no facts on error
	}

	// Save each atomic fact
	factMemories := make([]*Memory, 0, len(facts))
	for i, fact := range facts {
		// Sub-key: parentKey::index
		subKey := fmt.Sprintf("%s::%d", m.Key, i)

		factMemory := &Memory{
			ProfileID:  m.ProfileID,
			Type:       TypeFact,
			Scope:      m.Scope,
			MediaType:  MediaText,
			Key:        subKey,
			Value:      fact.Content,
			Confidence: m.Confidence,
			Tags:       []string{string(fact.FactType), "atomic_fact"},
			Metadata: Metadata{
				Source: "extractor",
				Extra: map[string]any{
					"parent_key":  fact.ParentKey,
					"entities":    fact.Entities,
					"fact_type":   fact.FactType,
				},
			},
		}

		// Generate embedding for fact
		if s.embeddingSvc != nil {
			vector, err := s.embeddingSvc.Embed(fact.Content)
			if err == nil {
				factMemory.Embedding = vector
			}
		}

		if err := s.sqliteStore.Save(factMemory); err != nil {
			continue // Skip failed facts
		}

		// Save to vector store
		if s.vectorStore != nil && len(factMemory.Embedding) > 0 {
			meta := s.memoryToMetadata(factMemory)
			if err := s.vectorStore.Upsert(factMemory.ID, factMemory.Embedding, meta); err != nil {
				continue
			}
		}

		factMemories = append(factMemories, factMemory)
	}

	return m, factMemories, nil
}

// memoryToMetadata converts Memory to vector store metadata.
func (s *Store) memoryToMetadata(m *Memory) map[string]any {
	return map[string]any{
		"id":         m.ID,
		"profile_id": m.ProfileID,
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

// maxFloat64 returns the larger of two floats.
func maxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
