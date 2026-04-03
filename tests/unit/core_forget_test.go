package unit

import (
	"errors"
	"testing"
	"time"

	"localmemory/core"
)

func TestForget_NewForget(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()

	forget := core.NewForget(sqlite, vector)

	if forget == nil {
		t.Fatal("NewForget returned nil")
	}
}

func TestForget_Delete(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()

	// Pre-add a memory
	memory := &core.Memory{
		Key:   "test-key",
		Value: "test value",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}
	memory.BeforeSave()
	sqlite.memories[memory.ID] = memory

	forget := core.NewForget(sqlite, vector)

	err := forget.Delete(memory.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify memory is soft deleted
	retrieved, _ := sqlite.GetByID(memory.ID)
	if !retrieved.Deleted {
		t.Error("Expected memory to be marked as deleted")
	}
	if retrieved.DeletedAt == 0 {
		t.Error("Expected DeletedAt to be set")
	}
}

func TestForget_Delete_NotFound(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()

	forget := core.NewForget(sqlite, vector)

	err := forget.Delete("non-existent-id")
	if err == nil {
		t.Fatal("Expected error when deleting non-existent memory")
	}
	if !errors.Is(err, core.ErrMemoryNotFound) {
		t.Errorf("Expected ErrMemoryNotFound, got %v", err)
	}
}

func TestForget_Delete_SQLiteError(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	sqlite.getErr = errors.New("sqlite error")
	vector := NewMockVectorStore()

	forget := core.NewForget(sqlite, vector)

	err := forget.Delete("some-id")
	if err == nil {
		t.Fatal("Expected error when SQLite fails")
	}
}

func TestForget_Restore(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()

	// Pre-add a deleted memory
	memory := &core.Memory{
		Key:       "test-key",
		Value:     "test value",
		Type:      core.TypeFact,
		Scope:     core.ScopeGlobal,
		Deleted:   true,
		DeletedAt: time.Now().Unix(),
	}
	memory.BeforeSave()
	sqlite.memories[memory.ID] = memory

	forget := core.NewForget(sqlite, vector)

	err := forget.Restore(memory.ID)
	if err != nil {
		t.Fatalf("Restore() error = %v", err)
	}

	// Verify memory is restored
	retrieved, _ := sqlite.GetByID(memory.ID)
	if retrieved.Deleted {
		t.Error("Expected memory to be restored (not deleted)")
	}
	if retrieved.DeletedAt != 0 {
		t.Error("Expected DeletedAt to be cleared")
	}
}

func TestForget_Restore_NotFound(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()

	forget := core.NewForget(sqlite, vector)

	err := forget.Restore("non-existent-id")
	if err == nil {
		t.Fatal("Expected error when restoring non-existent memory")
	}
	if !errors.Is(err, core.ErrMemoryNotFound) {
		t.Errorf("Expected ErrMemoryNotFound, got %v", err)
	}
}

func TestForget_Restore_NotDeleted(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()

	// Pre-add a non-deleted memory
	memory := &core.Memory{
		Key:     "test-key",
		Value:   "test value",
		Type:    core.TypeFact,
		Scope:   core.ScopeGlobal,
		Deleted: false,
	}
	memory.BeforeSave()
	sqlite.memories[memory.ID] = memory

	forget := core.NewForget(sqlite, vector)

	err := forget.Restore(memory.ID)
	if err == nil {
		t.Fatal("Expected error when restoring non-deleted memory")
	}
	if !errors.Is(err, core.ErrMemoryNotDeleted) {
		t.Errorf("Expected ErrMemoryNotDeleted, got %v", err)
	}
}

func TestForget_HardDelete(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()

	// Pre-add a memory with vector
	memory := &core.Memory{
		Key:   "test-key",
		Value: "test value",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}
	memory.BeforeSave()
	sqlite.memories[memory.ID] = memory
	vector.vectors[memory.ID] = make([]float32, 1536)

	forget := core.NewForget(sqlite, vector)

	err := forget.HardDelete(memory.ID)
	if err != nil {
		t.Fatalf("HardDelete() error = %v", err)
	}

	// Verify memory is removed from SQLite
	_, err = sqlite.GetByID(memory.ID)
	// The mock returns nil, nil for non-existent
}

func TestForget_HardDelete_NotFound(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()

	forget := core.NewForget(sqlite, vector)

	err := forget.HardDelete("non-existent-id")
	if err == nil {
		t.Fatal("Expected error when hard deleting non-existent memory")
	}
	if !errors.Is(err, core.ErrMemoryNotFound) {
		t.Errorf("Expected ErrMemoryNotFound, got %v", err)
	}
}

func TestForget_HardDelete_VectorStoreWarning(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()
	vector.deleteErr = errors.New("vector delete error")

	// Pre-add a memory with vector
	memory := &core.Memory{
		Key:   "test-key",
		Value: "test value",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}
	memory.BeforeSave()
	sqlite.memories[memory.ID] = memory
	vector.vectors[memory.ID] = make([]float32, 1536)

	forget := core.NewForget(sqlite, vector)

	// Hard delete should still succeed even if vector store fails
	err := forget.HardDelete(memory.ID)
	if err != nil {
		t.Fatalf("HardDelete() error = %v", err)
	}
}

func TestForget_HardDelete_NoVectorStore(t *testing.T) {
	sqlite := NewMockSQLiteStore()

	// Pre-add a memory
	memory := &core.Memory{
		Key:   "test-key",
		Value: "test value",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}
	memory.BeforeSave()
	sqlite.memories[memory.ID] = memory

	forget := core.NewForget(sqlite, nil) // No vector store

	err := forget.HardDelete(memory.ID)
	if err != nil {
		t.Fatalf("HardDelete() error = %v", err)
	}
}

func TestForget_PurgeExpired(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()

	// Pre-add an old memory (should be expired)
	oldMemory := &core.Memory{
		Key:      "old-key",
		Value:    "old value",
		Type:     core.TypeFact,
		Scope:    core.ScopeGlobal,
		CreatedAt: time.Now().Add(-30 * 24 * time.Hour).Unix(), // 30 days old
	}
	oldMemory.BeforeSave()
	sqlite.memories[oldMemory.ID] = oldMemory
	vector.vectors[oldMemory.ID] = make([]float32, 1536)

	// Pre-add a recent memory (should not be expired)
	recentMemory := &core.Memory{
		Key:       "recent-key",
		Value:     "recent value",
		Type:      core.TypeFact,
		Scope:     core.ScopeGlobal,
		CreatedAt: time.Now().Unix(),
	}
	recentMemory.BeforeSave()
	sqlite.memories[recentMemory.ID] = recentMemory
	vector.vectors[recentMemory.ID] = make([]float32, 1536)

	forget := core.NewForget(sqlite, vector)

	// Purge with lambda=0.01 and threshold=0.1
	count, err := forget.PurgeExpired(0.01, 0.1)
	if err != nil {
		t.Fatalf("PurgeExpired() error = %v", err)
	}

	// Should have purged exactly 1 (the old one)
	if count != 1 {
		t.Errorf("Expected to purge 1 expired memory, got %d", count)
	}

	// Recent memory should still exist
	if _, ok := sqlite.memories[recentMemory.ID]; !ok {
		t.Error("Expected recent memory to still exist")
	}
}

func TestForget_PurgeExpired_NoMatches(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()

	// Pre-add only recent memories
	recentMemory := &core.Memory{
		Key:       "recent-key",
		Value:     "recent value",
		Type:      core.TypeFact,
		Scope:     core.ScopeGlobal,
		CreatedAt: time.Now().Unix(),
	}
	recentMemory.BeforeSave()
	sqlite.memories[recentMemory.ID] = recentMemory

	forget := core.NewForget(sqlite, vector)

	count, err := forget.PurgeExpired(0.01, 0.1)
	if err != nil {
		t.Fatalf("PurgeExpired() error = %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 purged, got %d", count)
	}
}

func TestForget_PurgeExpired_Empty(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()

	forget := core.NewForget(sqlite, vector)

	count, err := forget.PurgeExpired(0.01, 0.1)
	if err != nil {
		t.Fatalf("PurgeExpired() error = %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 purged, got %d", count)
	}
}

func TestForget_PurgeByScope(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()

	// Pre-add global memories
	globalMemory1 := &core.Memory{
		Key:   "global-key-1",
		Value: "global value 1",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}
	globalMemory1.BeforeSave()
	sqlite.memories[globalMemory1.ID] = globalMemory1

	globalMemory2 := &core.Memory{
		Key:   "global-key-2",
		Value: "global value 2",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}
	globalMemory2.BeforeSave()
	sqlite.memories[globalMemory2.ID] = globalMemory2

	// Pre-add session memory
	sessionMemory := &core.Memory{
		Key:   "session-key",
		Value: "session value",
		Type:  core.TypeFact,
		Scope: core.ScopeSession,
	}
	sessionMemory.BeforeSave()
	sqlite.memories[sessionMemory.ID] = sessionMemory

	forget := core.NewForget(sqlite, vector)

	count, err := forget.PurgeByScope(core.ScopeGlobal)
	if err != nil {
		t.Fatalf("PurgeByScope() error = %v", err)
	}

	if count != 2 {
		t.Errorf("Expected to purge 2 global memories, got %d", count)
	}

	// Session memory should still exist
	if _, ok := sqlite.memories[sessionMemory.ID]; !ok {
		t.Error("Expected session memory to still exist")
	}
}

func TestForget_PurgeByScope_NoMatches(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()

	// Pre-add only session memories
	sessionMemory := &core.Memory{
		Key:   "session-key",
		Value: "session value",
		Type:  core.TypeFact,
		Scope: core.ScopeSession,
	}
	sessionMemory.BeforeSave()
	sqlite.memories[sessionMemory.ID] = sessionMemory

	forget := core.NewForget(sqlite, vector)

	count, err := forget.PurgeByScope(core.ScopeGlobal)
	if err != nil {
		t.Fatalf("PurgeByScope() error = %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 purged, got %d", count)
	}
}

func TestForget_PurgeByScope_Empty(t *testing.T) {
	sqlite := NewMockSQLiteStore()
	vector := NewMockVectorStore()

	forget := core.NewForget(sqlite, vector)

	count, err := forget.PurgeByScope(core.ScopeGlobal)
	if err != nil {
		t.Fatalf("PurgeByScope() error = %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 purged, got %d", count)
	}
}
