//go:build cgo

package unit

import (
	"os"
	"testing"

	"localmemory/core"
	"localmemory/storage"
)

type testDB struct {
	store *storage.SQLiteStore
	path  string
}

func setupTestDB(t *testing.T) *testDB {
	tmpfile, err := os.CreateTemp("", "test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpfile.Close()
	path := tmpfile.Name()

	store, err := storage.NewSQLiteStore(path)
	if err != nil {
		os.Remove(path)
		t.Fatalf("Failed to create SQLite store: %v", err)
	}

	return &testDB{store: store, path: path}
}

func (tdb *testDB) cleanup() {
	tdb.store.Close()
	os.Remove(tdb.path)
}

func TestSQLiteStore_SaveAndGet(t *testing.T) {
	tdb := setupTestDB(t)
	defer tdb.cleanup()

	// Test save
	m := &core.Memory{
		Type:      core.TypePreference,
		Scope:     core.ScopeGlobal,
		Key:       "test_key",
		Value:     "test_value",
		Confidence: 0.9,
	}

	err := tdb.store.Save(m)
	if err != nil {
		t.Fatalf("Failed to save memory: %v", err)
	}

	if m.ID == "" {
		t.Error("Expected ID to be set after save")
	}

	// Test get by ID
	got, err := tdb.store.GetByID(m.ID)
	if err != nil {
		t.Fatalf("Failed to get memory: %v", err)
	}
	if got == nil {
		t.Fatal("Expected to get memory, got nil")
	}
	if got.Key != m.Key {
		t.Errorf("Expected key '%s', got '%s'", m.Key, got.Key)
	}
	if got.Value != m.Value {
		t.Errorf("Expected value '%s', got '%s'", m.Value, got.Value)
	}
}

func TestSQLiteStore_SaveAndGetByKey(t *testing.T) {
	tdb := setupTestDB(t)
	defer tdb.cleanup()

	m := &core.Memory{
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
		Key:   "unique_key_123",
		Value: "fact_value",
	}

	err := tdb.store.Save(m)
	if err != nil {
		t.Fatalf("Failed to save memory: %v", err)
	}

	got, err := tdb.store.GetByKey("unique_key_123", "default")
	if err != nil {
		t.Fatalf("Failed to get by key: %v", err)
	}
	if got == nil {
		t.Fatal("Expected to get memory by key, got nil")
	}
	if got.ID != m.ID {
		t.Errorf("Expected ID '%s', got '%s'", m.ID, got.ID)
	}
}

func TestSQLiteStore_Delete(t *testing.T) {
	tdb := setupTestDB(t)
	defer tdb.cleanup()

	m := &core.Memory{
		Type:  core.TypeEvent,
		Scope: core.ScopeSession,
		Key:   "delete_key",
		Value: "to_be_deleted",
	}

	err := tdb.store.Save(m)
	if err != nil {
		t.Fatalf("Failed to save memory: %v", err)
	}

	// Soft delete
	err = tdb.store.Delete(m.ID)
	if err != nil {
		t.Fatalf("Failed to delete memory: %v", err)
	}

	// Get should return memory but marked as deleted
	got, err := tdb.store.GetByID(m.ID)
	if err != nil {
		t.Fatalf("Failed to get memory: %v", err)
	}
	if got == nil {
		t.Fatal("Expected to get memory, got nil")
	}
	if !got.Deleted {
		t.Error("Expected memory to be marked as deleted")
	}
}

func TestSQLiteStore_List(t *testing.T) {
	tdb := setupTestDB(t)
	defer tdb.cleanup()

	// Save multiple memories
	for i := 0; i < 5; i++ {
		m := &core.Memory{
			Type:  core.TypePreference,
			Scope:  core.ScopeGlobal,
			Key:   "list_key",
			Value: "list_value",
		}
		tdb.store.Save(m)
	}

	req := &core.ListRequest{
		Scope:  core.ScopeGlobal,
		Limit: 10,
	}

	memories, total, err := tdb.store.List(req)
	if err != nil {
		t.Fatalf("Failed to list memories: %v", err)
	}

	if total < 5 {
		t.Errorf("Expected at least 5 memories, got %d", total)
	}
	if len(memories) != total {
		t.Errorf("Expected %d memories in list, got %d", total, len(memories))
	}
}

func TestSQLiteStore_GetStats(t *testing.T) {
	tdb := setupTestDB(t)
	defer tdb.cleanup()

	// Save memories of different types
	memories := []*core.Memory{
		{Type: core.TypePreference, Scope: core.ScopeGlobal, Key: "p1", Value: "v1"},
		{Type: core.TypeFact, Scope: core.ScopeGlobal, Key: "f1", Value: "v2"},
		{Type: core.TypePreference, Scope: core.ScopeAgent, Key: "p2", Value: "v3"},
	}

	for _, m := range memories {
		tdb.store.Save(m)
	}

	stats, err := tdb.store.GetStats()
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats.Total != 3 {
		t.Errorf("Expected 3 total, got %d", stats.Total)
	}
	if stats.ByType[string(core.TypePreference)] != 2 {
		t.Errorf("Expected 2 preferences, got %d", stats.ByType[string(core.TypePreference)])
	}
	if stats.ByScope[string(core.ScopeGlobal)] != 2 {
		t.Errorf("Expected 2 global scope, got %d", stats.ByScope[string(core.ScopeGlobal)])
	}
}
