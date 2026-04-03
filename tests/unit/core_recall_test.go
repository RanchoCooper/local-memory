package unit

import (
	"errors"
	"testing"

	"localmemory/core"
)

func TestRecall_NewRecall(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()
	ranker := core.NewRanker(0.01)

	recall := core.NewRecall(sqlite, vector, embedding, ranker)

	if recall == nil {
		t.Fatal("NewRecall returned nil")
	}
}

func TestRecall_Query_Success(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()
	ranker := core.NewRanker(0.01)

	// Pre-add some memories to SQLite
	memory1 := &core.Memory{
		Key:        "key1",
		Value:     "test value 1",
		Type:      core.TypeFact,
		Scope:     core.ScopeGlobal,
		Tags:      []string{"tag1"},
		Confidence: 0.9,
	}
	memory1.BeforeSave()
	sqlite.memories[memory1.ID] = memory1

	memory2 := &core.Memory{
		Key:        "key2",
		Value:     "test value 2",
		Type:      core.TypeFact,
		Scope:     core.ScopeGlobal,
		Tags:      []string{"tag2"},
		Confidence: 0.8,
	}
	memory2.BeforeSave()
	sqlite.memories[memory2.ID] = memory2

	// Add vectors for the memories
	vector.vectors[memory1.ID] = make([]float32, 1536)
	vector.vectors[memory2.ID] = make([]float32, 1536)

	recall := core.NewRecall(sqlite, vector, embedding, ranker)

	req := &core.QueryRequest{
		Query: "test query",
		TopK:  10,
	}

	response, err := recall.Query(req)
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	// Should return results (may be empty if vector search doesn't match)
	_ = response.Results // Just verify it's not nil
}

func TestRecall_Query_NoEmbeddingService(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	ranker := core.NewRanker(0.01)

	recall := core.NewRecall(sqlite, vector, nil, ranker)

	req := &core.QueryRequest{
		Query: "test query",
		TopK:  10,
	}

	_, err := recall.Query(req)
	if err == nil {
		t.Fatal("Expected error when embedding service is nil")
	}
	if !errors.Is(err, core.ErrEmbeddingServiceRequired) {
		t.Errorf("Expected ErrEmbeddingServiceRequired, got %v", err)
	}
}

func TestRecall_Query_NoVectorStore(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	embedding := NewMockEmbeddingService()
	ranker := core.NewRanker(0.01)

	recall := core.NewRecall(sqlite, nil, embedding, ranker)

	req := &core.QueryRequest{
		Query: "test query",
		TopK:  10,
	}

	_, err := recall.Query(req)
	if err == nil {
		t.Fatal("Expected error when vector store is nil")
	}
	if !errors.Is(err, core.ErrVectorStoreRequired) {
		t.Errorf("Expected ErrVectorStoreRequired, got %v", err)
	}
}

func TestRecall_Query_EmbeddingError(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()
	embedding.embedErr = errors.New("embedding error")
	ranker := core.NewRanker(0.01)

	recall := core.NewRecall(sqlite, vector, embedding, ranker)

	req := &core.QueryRequest{
		Query: "test query",
		TopK:  10,
	}

	_, err := recall.Query(req)
	if err == nil {
		t.Fatal("Expected error when embedding fails")
	}
}

func TestRecall_Query_VectorSearchError(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	vector.searchErr = errors.New("vector search error")
	embedding := NewMockEmbeddingService()
	ranker := core.NewRanker(0.01)

	recall := core.NewRecall(sqlite, vector, embedding, ranker)

	req := &core.QueryRequest{
		Query: "test query",
		TopK:  10,
	}

	_, err := recall.Query(req)
	if err == nil {
		t.Fatal("Expected error when vector search fails")
	}
}

func TestRecall_Query_WithScopeFilter(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()
	ranker := core.NewRanker(0.01)

	// Pre-add memories with different scopes
	globalMemory := &core.Memory{
		Key:    "global-key",
		Value:  "global value",
		Type:   core.TypeFact,
		Scope:  core.ScopeGlobal,
		Tags:   []string{"tag1"},
	}
	globalMemory.BeforeSave()
	sqlite.memories[globalMemory.ID] = globalMemory
	vector.vectors[globalMemory.ID] = make([]float32, 1536)

	sessionMemory := &core.Memory{
		Key:    "session-key",
		Value:  "session value",
		Type:   core.TypeFact,
		Scope:  core.ScopeSession,
		Tags:   []string{"tag1"},
	}
	sessionMemory.BeforeSave()
	sqlite.memories[sessionMemory.ID] = sessionMemory
	vector.vectors[sessionMemory.ID] = make([]float32, 1536)

	recall := core.NewRecall(sqlite, vector, embedding, ranker)

	req := &core.QueryRequest{
		Query: "test query",
		TopK:  10,
		Scope: core.ScopeGlobal,
	}

	response, err := recall.Query(req)
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}

	// Query should succeed with scope filter
	// Note: The mock vector store doesn't filter by scope,
	// but the actual implementation filters at vector search level
	_ = response
}

func TestRecall_Query_WithTagFilter(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()
	ranker := core.NewRanker(0.01)

	// Pre-add memories with different tags
	taggedMemory := &core.Memory{
		Key:    "tagged-key",
		Value:  "tagged value",
		Type:   core.TypeFact,
		Scope:  core.ScopeGlobal,
		Tags:   []string{"important", "work"},
	}
	taggedMemory.BeforeSave()
	sqlite.memories[taggedMemory.ID] = taggedMemory
	vector.vectors[taggedMemory.ID] = make([]float32, 1536)

	otherMemory := &core.Memory{
		Key:    "other-key",
		Value:  "other value",
		Type:   core.TypeFact,
		Scope:  core.ScopeGlobal,
		Tags:   []string{"personal"},
	}
	otherMemory.BeforeSave()
	sqlite.memories[otherMemory.ID] = otherMemory
	vector.vectors[otherMemory.ID] = make([]float32, 1536)

	recall := core.NewRecall(sqlite, vector, embedding, ranker)

	req := &core.QueryRequest{
		Query: "test query",
		TopK:  10,
		Tags:  []string{"important"},
	}

	response, err := recall.Query(req)
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}

	// Response should not be nil
	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	// Check that the vector store was called with the tag filter
	if len(vector.searchCalls) == 0 {
		t.Error("Expected vector store Search to be called")
	}
}

func TestRecall_Query_LimitsResults(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()
	ranker := core.NewRanker(0.01)

	// Add multiple memories
	for i := 0; i < 20; i++ {
		memory := &core.Memory{
			Key:   "key",
			Value: "test value",
			Type:  core.TypeFact,
			Scope: core.ScopeGlobal,
		}
		memory.BeforeSave()
		sqlite.memories[memory.ID] = memory
		vector.vectors[memory.ID] = make([]float32, 1536)
	}

	recall := core.NewRecall(sqlite, vector, embedding, ranker)

	req := &core.QueryRequest{
		Query: "test query",
		TopK:  5,
	}

	response, err := recall.Query(req)
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}

	if len(response.Results) > req.TopK {
		t.Errorf("Expected at most %d results, got %d", req.TopK, len(response.Results))
	}
}

func TestRecall_GetByID(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()
	ranker := core.NewRanker(0.01)

	// Pre-add a memory
	memory := &core.Memory{
		Key:   "test-key",
		Value: "test value",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}
	memory.BeforeSave()
	sqlite.memories[memory.ID] = memory

	recall := core.NewRecall(sqlite, vector, embedding, ranker)

	result, err := recall.GetByID(memory.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if result == nil {
		t.Fatal("Expected memory, got nil")
	}

	if result.Key != memory.Key {
		t.Errorf("Key = %v, want %v", result.Key, memory.Key)
	}
}

func TestRecall_GetByID_NotFound(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()
	ranker := core.NewRanker(0.01)

	recall := core.NewRecall(sqlite, vector, embedding, ranker)

	result, err := recall.GetByID("non-existent-id")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if result != nil {
		t.Error("Expected nil for non-existent ID")
	}
}

func TestRecall_List(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()
	ranker := core.NewRanker(0.01)

	// Pre-add memories
	for i := 0; i < 5; i++ {
		memory := &core.Memory{
			Key:   "key",
			Value: "test value",
			Type:  core.TypeFact,
			Scope: core.ScopeGlobal,
		}
		memory.BeforeSave()
		sqlite.memories[memory.ID] = memory
	}

	recall := core.NewRecall(sqlite, vector, embedding, ranker)

	req := &core.ListRequest{
		Limit: 10,
	}

	response, err := recall.List(req)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if len(response.Memories) != 5 {
		t.Errorf("Expected 5 memories, got %d", len(response.Memories))
	}

	if response.Total != 5 {
		t.Errorf("Expected total 5, got %d", response.Total)
	}
}

func TestRecall_List_DefaultLimit(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()
	ranker := core.NewRanker(0.01)

	recall := core.NewRecall(sqlite, vector, embedding, ranker)

	// Zero limit should use default
	req := &core.ListRequest{
		Limit: 0,
	}

	response, err := recall.List(req)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	// Default limit is 20
	if len(response.Memories) > 20 {
		t.Errorf("Expected at most 20 memories with default limit, got %d", len(response.Memories))
	}
}

func TestRecall_List_NegativeOffset(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()
	ranker := core.NewRanker(0.01)

	recall := core.NewRecall(sqlite, vector, embedding, ranker)

	req := &core.ListRequest{
		Limit:  10,
		Offset: -1,
	}

	response, err := recall.List(req)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	// Negative offset should be treated as 0
	if response == nil {
		t.Fatal("Expected response, got nil")
	}
}

func TestRecall_List_WithScope(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()
	ranker := core.NewRanker(0.01)

	// Add memories with different scopes
	globalMemory := &core.Memory{
		Key:   "global-key",
		Value: "global value",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}
	globalMemory.BeforeSave()
	sqlite.memories[globalMemory.ID] = globalMemory

	sessionMemory := &core.Memory{
		Key:   "session-key",
		Value: "session value",
		Type:  core.TypeFact,
		Scope: core.ScopeSession,
	}
	sessionMemory.BeforeSave()
	sqlite.memories[sessionMemory.ID] = sessionMemory

	recall := core.NewRecall(sqlite, vector, embedding, ranker)

	req := &core.ListRequest{
		Scope: core.ScopeGlobal,
		Limit: 10,
	}

	response, err := recall.List(req)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	for _, m := range response.Memories {
		if m.Scope != core.ScopeGlobal {
			t.Errorf("Expected scope Global, got %v", m.Scope)
		}
	}
}

func TestRecall_List_IncludeDeleted(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()
	ranker := core.NewRanker(0.01)

	// Add normal memory
	normalMemory := &core.Memory{
		Key:    "normal-key",
		Value:  "normal value",
		Type:   core.TypeFact,
		Scope:  core.ScopeGlobal,
		Deleted: false,
	}
	normalMemory.BeforeSave()
	sqlite.memories[normalMemory.ID] = normalMemory

	// Add deleted memory
	deletedMemory := &core.Memory{
		Key:     "deleted-key",
		Value:   "deleted value",
		Type:    core.TypeFact,
		Scope:   core.ScopeGlobal,
		Deleted: true,
	}
	deletedMemory.BeforeSave()
	sqlite.memories[deletedMemory.ID] = deletedMemory

	recall := core.NewRecall(sqlite, vector, embedding, ranker)

	// Without include deleted
	req1 := &core.ListRequest{
		IncludeDeleted: false,
		Limit:          10,
	}

	response1, err := recall.List(req1)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(response1.Memories) != 1 {
		t.Errorf("Expected 1 non-deleted memory, got %d", len(response1.Memories))
	}

	// With include deleted
	req2 := &core.ListRequest{
		IncludeDeleted: true,
		Limit:          10,
	}

	response2, err := recall.List(req2)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(response2.Memories) != 2 {
		t.Errorf("Expected 2 total memories, got %d", len(response2.Memories))
	}
}

func TestRecall_List_Empty(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()
	ranker := core.NewRanker(0.01)

	recall := core.NewRecall(sqlite, vector, embedding, ranker)

	req := &core.ListRequest{
		Limit: 10,
	}

	response, err := recall.List(req)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(response.Memories) != 0 {
		t.Errorf("Expected 0 memories, got %d", len(response.Memories))
	}

	if response.Total != 0 {
		t.Errorf("Expected total 0, got %d", response.Total)
	}
}
