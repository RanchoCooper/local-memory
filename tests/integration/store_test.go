//go:build cgo

package integration

import (
	"os"
	"testing"

	"localmemory/core"
	"localmemory/storage"
)

// Integration test for Store module
// Tests the full flow: Save -> Recall -> Evolve -> Forget

func setupIntegrationDB(t *testing.T) (*storage.SQLiteStore, func()) {
	tmpfile, err := os.CreateTemp("", "integration_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpfile.Close()

	store, err := storage.NewSQLiteStore(tmpfile.Name())
	if err != nil {
		os.Remove(tmpfile.Name())
		t.Fatalf("Failed to create SQLite store: %v", err)
	}

	cleanup := func() {
		store.Close()
		os.Remove(tmpfile.Name())
	}

	return store, cleanup
}

func TestIntegration_StoreSaveAndRecall(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	store, cleanup := setupIntegrationDB(t)
	defer cleanup()

	// Create a simple store without embedding service (MVP mode)
	s := core.NewStore(store, nil, nil)

	// Save a memory
	m := &core.Memory{
		Type:  core.TypePreference,
		Scope: core.ScopeGlobal,
		Key:   "integration_test_key",
		Value: "integration_test_value",
	}

	err := s.Save(m)
	if err != nil {
		t.Fatalf("Failed to save memory: %v", err)
	}

	// Get the memory back
	got, err := store.GetByID(m.ID)
	if err != nil {
		t.Fatalf("Failed to get memory: %v", err)
	}
	if got == nil {
		t.Fatal("Expected to get memory")
	}
	if got.Key != m.Key {
		t.Errorf("Expected key '%s', got '%s'", m.Key, got.Key)
	}
}

func TestIntegration_EvolveMemory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	store, cleanup := setupIntegrationDB(t)
	defer cleanup()

	s := core.NewStore(store, nil, nil)

	// Save first memory
	m1 := &core.Memory{
		Type:  core.TypePreference,
		Scope: core.ScopeGlobal,
		Key:   "evolve_test_key",
		Value: "first_value",
	}

	err := s.Save(m1)
	if err != nil {
		t.Fatalf("Failed to save first memory: %v", err)
	}

	// Save second memory with same key - should trigger evolve
	m2 := &core.Memory{
		Type:  core.TypePreference,
		Scope: core.ScopeGlobal,
		Key:   "evolve_test_key",
		Value: "second_value",
	}

	err = s.Save(m2)
	if err != nil {
		t.Fatalf("Failed to save second memory: %v", err)
	}

	// Get the memory - should have merged values
	got, err := store.GetByKey("evolve_test_key", "default")
	if err != nil {
		t.Fatalf("Failed to get memory: %v", err)
	}
	if got == nil {
		t.Fatal("Expected to get memory")
	}

	// Value should contain both values
	if got.Value != "first_value\nsecond_value" {
		t.Errorf("Expected merged value, got '%s'", got.Value)
	}

	// Confidence should be higher (capped at 1.0)
	if got.Confidence < 1.0 {
		t.Errorf("Expected confidence >= 1.0 after evolve, got %f", got.Confidence)
	}
}

func TestIntegration_ForgetMemory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	store, cleanup := setupIntegrationDB(t)
	defer cleanup()

	forget := core.NewForget(store, nil)

	// Save a memory
	m := &core.Memory{
		Type:  core.TypeEvent,
		Scope: core.ScopeSession,
		Key:   "forget_test_key",
		Value: "to_be_forgotten",
	}
	err := store.Save(m)
	if err != nil {
		t.Fatalf("Failed to save memory: %v", err)
	}

	// Soft delete
	err = forget.Delete(m.ID)
	if err != nil {
		t.Fatalf("Failed to delete memory: %v", err)
	}

	// Verify it's marked as deleted
	got, err := store.GetByID(m.ID)
	if err != nil {
		t.Fatalf("Failed to get memory: %v", err)
	}
	if !got.Deleted {
		t.Error("Expected memory to be marked as deleted")
	}
}

func TestIntegration_ListMemories(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	store, cleanup := setupIntegrationDB(t)
	defer cleanup()

	// Save multiple memories
	for i := 0; i < 10; i++ {
		m := &core.Memory{
			Type:  core.TypePreference,
			Scope: core.ScopeGlobal,
			Key:   "list_test_key",
			Value: "list_test_value",
		}
		store.Save(m)
	}

	// List with pagination
	ranker := core.NewRanker(0.01)
	recall := core.NewRecall(store, nil, nil, ranker)

	resp, err := recall.List(&core.ListRequest{
		Scope:  core.ScopeGlobal,
		Limit:  5,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("Failed to list memories: %v", err)
	}

	if resp.Total != 10 {
		t.Errorf("Expected 10 total, got %d", resp.Total)
	}
	if len(resp.Memories) != 5 {
		t.Errorf("Expected 5 memories in page, got %d", len(resp.Memories))
	}
}
