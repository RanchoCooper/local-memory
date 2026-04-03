package unit

import (
	"errors"
	"testing"
	"time"

	"localmemory/core"
)

func TestEvolve_NewEvolve(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()
	store := core.NewStore(sqlite, vector, embedding)

	evolve := core.NewEvolve(store)

	if evolve == nil {
		t.Fatal("NewEvolve returned nil")
	}
}

func TestEvolve_Merge_AppendStrategy(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()
	store := core.NewStore(sqlite, vector, embedding)
	evolve := core.NewEvolve(store)

	existing := &core.Memory{
		Key:        "test-key",
		Value:     "existing value",
		Type:      core.TypeFact,
		Scope:     core.ScopeGlobal,
		Confidence: 0.8,
		Tags:      []string{"tag1"},
		RelatedIDs: []string{"related-1"},
		Metadata: core.Metadata{
			Source: "original",
		},
	}

	new := &core.Memory{
		Key:        "test-key",
		Value:     "new value",
		Type:      core.TypeFact,
		Scope:     core.ScopeGlobal,
		Confidence: 0.9,
		Tags:      []string{"tag2"},
		RelatedIDs: []string{"related-2"},
		Metadata: core.Metadata{
			Source: "update",
		},
	}

	result, err := evolve.Merge(existing, new, nil)
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// Value should be appended
	expectedValue := "existing value\nnew value"
	if result.Value != expectedValue {
		t.Errorf("Value = %v, want %v", result.Value, expectedValue)
	}

	// Confidence should be increased (0.8 + 0.1 = 0.9, capped at 1.0)
	if result.Confidence != 0.9 {
		t.Errorf("Confidence = %v, want 0.9", result.Confidence)
	}

	// Tags should be merged
	if len(result.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(result.Tags))
	}

	// RelatedIDs should be merged
	if len(result.RelatedIDs) != 2 {
		t.Errorf("Expected 2 related IDs, got %d", len(result.RelatedIDs))
	}

	// Metadata source should be updated
	if result.Metadata.Source != "update" {
		t.Errorf("Metadata.Source = %v, want update", result.Metadata.Source)
	}
}

func TestEvolve_Merge_ReplaceStrategy(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()
	store := core.NewStore(sqlite, vector, embedding)
	evolve := core.NewEvolve(store)

	existing := &core.Memory{
		Key:   "test-key",
		Value: "existing value",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}

	new := &core.Memory{
		Key:   "test-key",
		Value: "new value",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}

	opts := &core.MergeOption{
		Strategy: "replace",
	}

	result, err := evolve.Merge(existing, new, opts)
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// Value should be replaced
	if result.Value != "new value" {
		t.Errorf("Value = %v, want new value", result.Value)
	}
}

func TestEvolve_Merge_MaxStrategy(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()
	store := core.NewStore(sqlite, vector, embedding)
	evolve := core.NewEvolve(store)

	existing := &core.Memory{
		Key:   "test-key",
		Value: "short",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}

	new := &core.Memory{
		Key:   "test-key",
		Value: "much longer new value",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}

	opts := &core.MergeOption{
		Strategy: "max",
	}

	result, err := evolve.Merge(existing, new, opts)
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// Value should be the longer one
	if result.Value != "much longer new value" {
		t.Errorf("Value = %v, want much longer new value", result.Value)
	}
}

func TestEvolve_Merge_MaxStrategy_ExistingLonger(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()
	store := core.NewStore(sqlite, vector, embedding)
	evolve := core.NewEvolve(store)

	existing := &core.Memory{
		Key:   "test-key",
		Value: "much longer existing value",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}

	new := &core.Memory{
		Key:   "test-key",
		Value: "short",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}

	opts := &core.MergeOption{
		Strategy: "max",
	}

	result, err := evolve.Merge(existing, new, opts)
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// Value should be the existing (longer) one
	if result.Value != "much longer existing value" {
		t.Errorf("Value = %v, want much longer existing value", result.Value)
	}
}

func TestEvolve_Merge_Append_EmptyValues(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()
	store := core.NewStore(sqlite, vector, embedding)
	evolve := core.NewEvolve(store)

	existing := &core.Memory{
		Key:   "test-key",
		Value: "existing",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}

	new := &core.Memory{
		Key:   "test-key",
		Value: "", // Empty value
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}

	result, err := evolve.Merge(existing, new, nil)
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// Should keep existing value when new is empty
	if result.Value != "existing" {
		t.Errorf("Value = %v, want existing", result.Value)
	}
}

func TestEvolve_Merge_Append_BothEmpty(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()
	store := core.NewStore(sqlite, vector, embedding)
	evolve := core.NewEvolve(store)

	existing := &core.Memory{
		Key:   "test-key",
		Value: "",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}

	new := &core.Memory{
		Key:   "test-key",
		Value: "",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}

	result, err := evolve.Merge(existing, new, nil)
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// Should be empty
	if result.Value != "" {
		t.Errorf("Value = %v, want empty", result.Value)
	}
}

func TestEvolve_Merge_ConfidenceCap(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()
	store := core.NewStore(sqlite, vector, embedding)
	evolve := core.NewEvolve(store)

	existing := &core.Memory{
		Key:        "test-key",
		Value:     "existing",
		Type:      core.TypeFact,
		Scope:     core.ScopeGlobal,
		Confidence: 0.95, // High confidence
	}

	new := &core.Memory{
		Key:        "test-key",
		Value:     "new",
		Type:      core.TypeFact,
		Scope:     core.ScopeGlobal,
		Confidence: 0.9,
	}

	result, err := evolve.Merge(existing, new, nil)
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// Confidence should be capped at 1.0
	if result.Confidence != 1.0 {
		t.Errorf("Confidence = %v, want 1.0 (capped)", result.Confidence)
	}
}

func TestEvolve_Merge_UpdatesTimestamp(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()
	store := core.NewStore(sqlite, vector, embedding)
	evolve := core.NewEvolve(store)

	oldTimestamp := time.Now().Add(-24 * time.Hour).Unix()

	existing := &core.Memory{
		Key:       "test-key",
		Value:     "existing",
		Type:      core.TypeFact,
		Scope:     core.ScopeGlobal,
		UpdatedAt: oldTimestamp,
	}

	new := &core.Memory{
		Key:   "test-key",
		Value: "new",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}

	result, err := evolve.Merge(existing, new, nil)
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// UpdatedAt should be updated
	if result.UpdatedAt <= oldTimestamp {
		t.Error("Expected UpdatedAt to be updated to a newer timestamp")
	}
}

func TestEvolve_Merge_NilOptions(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()
	store := core.NewStore(sqlite, vector, embedding)
	evolve := core.NewEvolve(store)

	existing := &core.Memory{
		Key:   "test-key",
		Value: "existing",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}

	new := &core.Memory{
		Key:   "test-key",
		Value: "new",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}

	// nil options should use default (append)
	result, err := evolve.Merge(existing, new, nil)
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// Should use append strategy by default
	if result.Value != "existing\nnew" {
		t.Errorf("Value = %v, want 'existing\\nnew'", result.Value)
	}
}

func TestEvolve_Merge_UnknownStrategy(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()
	store := core.NewStore(sqlite, vector, embedding)
	evolve := core.NewEvolve(store)

	existing := &core.Memory{
		Key:   "test-key",
		Value: "existing",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}

	new := &core.Memory{
		Key:   "test-key",
		Value: "new",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}

	opts := &core.MergeOption{
		Strategy: "unknown",
	}

	result, err := evolve.Merge(existing, new, opts)
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// Unknown strategy should default to new value
	if result.Value != "new" {
		t.Errorf("Value = %v, want new", result.Value)
	}
}

func TestEvolve_EvolveExisting_NotFound(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()
	store := core.NewStore(sqlite, vector, embedding)
	evolve := core.NewEvolve(store)

	memory := &core.Memory{
		Key:   "non-existent-key",
		Value: "new value",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}

	result, merged, err := evolve.EvolveExisting(memory)
	if err != nil {
		t.Fatalf("EvolveExisting() error = %v", err)
	}

	if merged {
		t.Error("Expected merged to be false for new memory")
	}

	if result.Key != memory.Key {
		t.Errorf("Key = %v, want %v", result.Key, memory.Key)
	}
}

func TestEvolve_EvolveExisting_Found(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()
	store := core.NewStore(sqlite, vector, embedding)

	// First save a memory
	existing := &core.Memory{
		Key:   "test-key",
		Value: "existing value",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}
	existing.BeforeSave()
	sqlite.memories[existing.ID] = existing

	evolve := core.NewEvolve(store)

	new := &core.Memory{
		Key:   "test-key",
		Value: "new value",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}

	result, merged, err := evolve.EvolveExisting(new)
	if err != nil {
		t.Fatalf("EvolveExisting() error = %v", err)
	}

	if !merged {
		t.Error("Expected merged to be true")
	}

	// Should have merged value
	if result.Value != "existing value\nnew value" {
		t.Errorf("Value = %v, want 'existing value\\nnew value'", result.Value)
	}
}

func TestEvolve_EvolveExisting_GetByKeyError(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	sqlite.getErr = errors.New("database error")
	vector := NewMockVectorStore()
	embedding := NewMockEmbeddingService()
	store := core.NewStore(sqlite, vector, embedding)
	evolve := core.NewEvolve(store)

	memory := &core.Memory{
		Key:   "test-key",
		Value: "new value",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}

	_, _, err := evolve.EvolveExisting(memory)
	if err == nil {
		t.Fatal("Expected error when GetByKey fails")
	}
}

func TestDefaultMergeOption(t *testing.T) {
	opts := core.DefaultMergeOption

	if opts == nil {
		t.Fatal("DefaultMergeOption is nil")
	}

	if opts.Strategy != "append" {
		t.Errorf("Strategy = %v, want append", opts.Strategy)
	}
}
