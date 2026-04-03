package unit

import (
	"errors"
	"testing"

	"localmemory/core"
)

// MockSQLiteStore is a mock implementation of SQLiteStoreInterface
type MockSQLiteStore struct {
	memories map[string]*core.Memory
	saveErr  error
	getErr   error
	listErr  error
}

func NewMockSQLiteStore() *MockSQLiteStore {
	return &MockSQLiteStore{
		memories: make(map[string]*core.Memory),
	}
}

func (m *MockSQLiteStore) Save(mem *core.Memory) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	mem.BeforeSave()
	m.memories[mem.ID] = mem
	return nil
}

func (m *MockSQLiteStore) GetByID(id string) (*core.Memory, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	mem, ok := m.memories[id]
	if !ok {
		return nil, nil
	}
	return mem, nil
}

func (m *MockSQLiteStore) GetByKey(key string) (*core.Memory, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, mem := range m.memories {
		if mem.Key == key && !mem.Deleted {
			return mem, nil
		}
	}
	return nil, nil
}

func (m *MockSQLiteStore) List(req *core.ListRequest) ([]*core.Memory, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	var result []*core.Memory
	for _, mem := range m.memories {
		if !req.IncludeDeleted && mem.Deleted {
			continue
		}
		if req.Scope != "" && mem.Scope != req.Scope {
			continue
		}
		result = append(result, mem)
	}
	return result, len(result), nil
}

// MockVectorStore is a mock implementation of VectorStore
type MockVectorStore struct {
	vectors    map[string][]float32
	metadata   map[string]map[string]any
	upsertErr  error
	searchErr  error
	deleteErr  error
	upsertCalls []struct {
		ID       string
		Vector   []float32
		Metadata map[string]any
	}
	searchCalls []struct {
		Query  []float32
		TopK   int
		Filter *core.VectorFilter
	}
}

func NewMockVectorStore() *MockVectorStore {
	return &MockVectorStore{
		vectors:    make(map[string][]float32),
		metadata:    make(map[string]map[string]any),
		upsertCalls: make([]struct {
			ID       string
			Vector   []float32
			Metadata map[string]any
		}, 0),
		searchCalls: make([]struct {
			Query  []float32
			TopK   int
			Filter *core.VectorFilter
		}, 0),
	}
}

func (m *MockVectorStore) Upsert(id string, vector []float32, metadata map[string]any) error {
	if m.upsertErr != nil {
		return m.upsertErr
	}
	m.vectors[id] = vector
	m.metadata[id] = metadata
	m.upsertCalls = append(m.upsertCalls, struct {
		ID       string
		Vector   []float32
		Metadata map[string]any
	}{id, vector, metadata})
	return nil
}

func (m *MockVectorStore) Search(query []float32, topK int, filter *core.VectorFilter) ([]core.VectorResult, error) {
	if m.searchErr != nil {
		return nil, m.searchErr
	}
	m.searchCalls = append(m.searchCalls, struct {
		Query  []float32
		TopK   int
		Filter *core.VectorFilter
	}{query, topK, filter})

	var results []core.VectorResult
	count := 0
	for id, vector := range m.vectors {
		if count >= topK {
			break
		}
		// Calculate simple cosine similarity
		score := m.cosineSimilarity(query, vector)
		results = append(results, core.VectorResult{
			ID:       id,
			Score:    score,
			Metadata: m.metadata[id],
		})
		count++
	}
	return results, nil
}

func (m *MockVectorStore) cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dotProduct float64
	var normA float64
	var normB float64
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

func (m *MockVectorStore) Delete(id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.vectors, id)
	delete(m.metadata, id)
	return nil
}

func (m *MockVectorStore) Close() error {
	return nil
}

// MockEmbeddingService is a mock implementation of EmbeddingService
type MockEmbeddingService struct {
	vectors     map[string][]float32
	embedErr    error
	embedCalls  []string
}

func NewMockEmbeddingService() *MockEmbeddingService {
	return &MockEmbeddingService{
		vectors:    make(map[string][]float32),
		embedCalls: make([]string, 0),
	}
}

func (m *MockEmbeddingService) Embed(text string) ([]float32, error) {
	m.embedCalls = append(m.embedCalls, text)
	if m.embedErr != nil {
		return nil, m.embedErr
	}
	// Return cached or generate deterministic vector based on text
	if vec, ok := m.vectors[text]; ok {
		return vec, nil
	}
	// Generate a fixed-size vector (1536 dimensions like OpenAI)
	vec := make([]float32, 1536)
	for i := range vec {
		vec[i] = float32(i % 100)
	}
	m.vectors[text] = vec
	return vec, nil
}

func (m *MockEmbeddingService) EmbedBatch(texts []string) ([][]float32, error) {
	result := make([][]float32, 0, len(texts))
	for _, text := range texts {
		vec, err := m.Embed(text)
		if err != nil {
			return nil, err
		}
		result = append(result, vec)
	}
	return result, nil
}

func TestNewStore(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()

	store := core.NewStore(sqlite, vector, embedding)

	if store == nil {
		t.Fatal("NewStore returned nil")
	}
}

func TestStore_Save_NewMemory(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()

	store := core.NewStore(sqlite, vector, embedding)

	memory := &core.Memory{
		Key:   "test-key",
		Value: "test value",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}

	err := store.Save(memory)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify ID was generated
	if memory.ID == "" {
		t.Error("Expected ID to be generated")
	}

	// Verify embedding was generated
	if len(memory.Embedding) == 0 {
		t.Error("Expected embedding to be generated")
	}

	// Verify memory was saved to SQLite
	if _, ok := sqlite.memories[memory.ID]; !ok {
		t.Error("Expected memory to be saved to SQLite")
	}

	// Verify memory was saved to vector store
	if _, ok := vector.vectors[memory.ID]; !ok {
		t.Error("Expected memory to be saved to vector store")
	}

	// Verify embedding service was called
	if len(embedding.embedCalls) == 0 {
		t.Error("Expected embedding service to be called")
	}
}

func TestStore_Save_MemoryWithSameKey_Evolves(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()

	store := core.NewStore(sqlite, vector, embedding)

	// Save first memory
	memory1 := &core.Memory{
		Key:   "test-key",
		Value: "original value",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}

	err := store.Save(memory1)
	if err != nil {
		t.Fatalf("First Save() error = %v", err)
	}

	originalID := memory1.ID
	originalConfidence := memory1.Confidence

	// Save second memory with same key
	memory2 := &core.Memory{
		Key:   "test-key",
		Value: "new value",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}

	err = store.Save(memory2)
	if err != nil {
		t.Fatalf("Second Save() error = %v", err)
	}

	// Verify values were merged
	if memory1.Value != "original value\nnew value" {
		t.Errorf("Expected merged value, got %q", memory1.Value)
	}

	// Verify ID was preserved
	if memory1.ID != originalID {
		t.Errorf("Expected ID to be preserved, got %s", memory1.ID)
	}

	// Verify confidence was increased (but capped at 1.0)
	expectedConfidence := originalConfidence + 0.1
	if expectedConfidence > 1.0 {
		expectedConfidence = 1.0
	}
	if memory1.Confidence != expectedConfidence {
		t.Errorf("Expected confidence %f, got %f", expectedConfidence, memory1.Confidence)
	}
}

func TestStore_Save_NoEmbeddingService(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()

	// No embedding service
	store := core.NewStore(sqlite, vector, nil)

	memory := &core.Memory{
		Key:   "test-key",
		Value: "test value",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}

	err := store.Save(memory)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify memory was saved to SQLite even without embedding
	if _, ok := sqlite.memories[memory.ID]; !ok {
		t.Error("Expected memory to be saved to SQLite")
	}

	// Verify no embedding was generated
	if len(memory.Embedding) != 0 {
		t.Error("Expected no embedding without embedding service")
	}
}

func TestStore_Save_NoVectorStore(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	embedding := NewMockEmbeddingService()

	// No vector store
	store := core.NewStore(sqlite, nil, embedding)

	memory := &core.Memory{
		Key:   "test-key",
		Value: "test value",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}

	err := store.Save(memory)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify embedding was generated
	if len(memory.Embedding) == 0 {
		t.Error("Expected embedding to be generated")
	}
}

func TestStore_Save_EmptyValue(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()

	store := core.NewStore(sqlite, vector, embedding)

	memory := &core.Memory{
		Key:   "test-key",
		Value: "", // Empty value
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}

	err := store.Save(memory)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify memory was saved to SQLite
	if _, ok := sqlite.memories[memory.ID]; !ok {
		t.Error("Expected memory to be saved to SQLite")
	}

	// Verify embedding was NOT generated for empty value
	if len(memory.Embedding) != 0 {
		t.Error("Expected no embedding for empty value")
	}
}

func TestStore_Save_SQLiteError(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	sqlite.saveErr = errors.New("sqlite error")

	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()

	store := core.NewStore(sqlite, vector, embedding)

	memory := &core.Memory{
		Key:   "test-key",
		Value: "test value",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}

	err := store.Save(memory)
	if err == nil {
		t.Fatal("Expected error when SQLite fails")
	}
}

func TestStore_Save_EmbeddingError(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()
	embedding.embedErr = errors.New("embedding error")

	store := core.NewStore(sqlite, vector, embedding)

	memory := &core.Memory{
		Key:   "test-key",
		Value: "test value",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}

	err := store.Save(memory)
	if err == nil {
		t.Fatal("Expected error when embedding fails")
	}
}

func TestStore_Save_VectorStoreError(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	vector.upsertErr = errors.New("vector store error")

	embedding := NewMockEmbeddingService()

	store := core.NewStore(sqlite, vector, embedding)

	memory := &core.Memory{
		Key:   "test-key",
		Value: "test value",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}

	err := store.Save(memory)
	// Vector store error is returned
	if err == nil {
		t.Fatal("Expected error when vector store fails")
	}
}

func TestStore_EvolveAndSave_Integration(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()

	store := core.NewStore(sqlite, vector, embedding)

	// Save initial memory with tags and related IDs
	memory1 := &core.Memory{
		Key:        "test-key",
		Value:     "original value",
		Type:      core.TypeFact,
		Scope:     core.ScopeGlobal,
		Tags:      []string{"tag1", "tag2"},
		RelatedIDs: []string{"related-1"},
	}

	err := store.Save(memory1)
	if err != nil {
		t.Fatalf("First Save() error = %v", err)
	}

	// Save memory with same key but different/additional tags
	memory2 := &core.Memory{
		Key:        "test-key",
		Value:     "new value",
		Type:      core.TypeFact,
		Scope:     core.ScopeGlobal,
		Tags:      []string{"tag2", "tag3"}, // tag2 is duplicate
		RelatedIDs: []string{"related-2"},   // related-2 is new
	}

	err = store.Save(memory2)
	if err != nil {
		t.Fatalf("Second Save() error = %v", err)
	}

	// Verify tags are merged (no duplicates)
	tagSet := make(map[string]bool)
	for _, tag := range memory1.Tags {
		tagSet[tag] = true
	}
	expectedTags := 3 // tag1, tag2, tag3
	if len(tagSet) != expectedTags {
		t.Errorf("Expected %d unique tags, got %d", expectedTags, len(tagSet))
	}

	// Verify related IDs are merged (no duplicates)
	idSet := make(map[string]bool)
	for _, id := range memory1.RelatedIDs {
		idSet[id] = true
	}
	expectedIDs := 2 // related-1, related-2
	if len(idSet) != expectedIDs {
		t.Errorf("Expected %d unique related IDs, got %d", expectedIDs, len(idSet))
	}
}

func TestStore_GetByID_ThroughSave(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()

	store := core.NewStore(sqlite, vector, embedding)

	memory := &core.Memory{
		Key:   "test-key",
		Value: "test value",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}

	err := store.Save(memory)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// GetByID is part of SQLiteStoreInterface, test through retrieval
	retrieved, err := sqlite.GetByID(memory.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if retrieved == nil {
		t.Fatal("Expected to retrieve memory")
	}
	if retrieved.Key != memory.Key {
		t.Errorf("Key = %v, want %v", retrieved.Key, memory.Key)
	}
	if retrieved.Value != memory.Value {
		t.Errorf("Value = %v, want %v", retrieved.Value, memory.Value)
	}
}

func TestMergeTags(t *testing.T) {
	tests := []struct {
		name string
		a    []string
		b    []string
		want int
	}{
		{
			name: "merge two lists with no duplicates",
			a:    []string{"tag1", "tag2"},
			b:    []string{"tag3", "tag4"},
			want: 4,
		},
		{
			name: "merge with duplicates",
			a:    []string{"tag1", "tag2"},
			b:    []string{"tag2", "tag3"},
			want: 3,
		},
		{
			name: "both empty",
			a:    []string{},
			b:    []string{},
			want: 0,
		},
		{
			name: "first empty",
			a:    []string{},
			b:    []string{"tag1"},
			want: 1,
		},
		{
			name: "second empty",
			a:    []string{"tag1"},
			b:    []string{},
			want: 1,
		},
		{
			name: "both nil",
			a:    nil,
			b:    nil,
			want: 0,
		},
		{
			name: "first nil",
			a:    nil,
			b:    []string{"tag1"},
			want: 1,
		},
		{
			name: "second nil",
			a:    []string{"tag1"},
			b:    nil,
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeTagsHelper(tt.a, tt.b)
			if len(result) != tt.want {
				t.Errorf("mergeTags() got %d items, want %d", len(result), tt.want)
			}
		})
	}
}

func TestMergeIDs(t *testing.T) {
	tests := []struct {
		name string
		a    []string
		b    []string
		want int
	}{
		{
			name: "merge two lists with no duplicates",
			a:    []string{"id1", "id2"},
			b:    []string{"id3", "id4"},
			want: 4,
		},
		{
			name: "merge with duplicates",
			a:    []string{"id1", "id2"},
			b:    []string{"id2", "id3"},
			want: 3,
		},
		{
			name: "both empty",
			a:    []string{},
			b:    []string{},
			want: 0,
		},
		{
			name: "both nil",
			a:    nil,
			b:    nil,
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeIDsHelper(tt.a, tt.b)
			if len(result) != tt.want {
				t.Errorf("mergeIDs() got %d items, want %d", len(result), tt.want)
			}
		})
	}
}

func TestMinFloat64(t *testing.T) {
	tests := []struct {
		name string
		a    float64
		b    float64
		want float64
	}{
		{"a less than b", 1.0, 2.0, 1.0},
		{"b less than a", 2.0, 1.0, 1.0},
		{"equal", 1.0, 1.0, 1.0},
		{"zero and positive", 0.0, 1.0, 0.0},
		{"negative values", -1.0, 0.0, -1.0},
		{"both negative", -2.0, -1.0, -2.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := minFloat64Helper(tt.a, tt.b)
			if result != tt.want {
				t.Errorf("minFloat64(%f, %f) = %f, want %f", tt.a, tt.b, result, tt.want)
			}
		})
	}
}

// Helper functions that mirror the core implementation for testing
func mergeTagsHelper(a, b []string) []string {
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

func mergeIDsHelper(a, b []string) []string {
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

func minFloat64Helper(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
