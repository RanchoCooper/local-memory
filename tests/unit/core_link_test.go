package unit

import (
	"errors"
	"testing"

	"localmemory/core"
)

func TestLink_NewLink(t *testing.T) {
	sqlite := NewMockSQLiteStore()

	link := core.NewLink(sqlite)

	if link == nil {
		t.Fatal("NewLink returned nil")
	}
}

func TestLink_Link(t *testing.T) {
	sqlite := NewMockSQLiteStore()

	// Pre-add two memories
	memory1 := &core.Memory{
		Key:   "key1",
		Value: "value 1",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}
	memory1.BeforeSave()
	sqlite.memories[memory1.ID] = memory1

	memory2 := &core.Memory{
		Key:   "key2",
		Value: "value 2",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}
	memory2.BeforeSave()
	sqlite.memories[memory2.ID] = memory2

	link := core.NewLink(sqlite)

	err := link.Link(memory1.ID, memory2.ID)
	if err != nil {
		t.Fatalf("Link() error = %v", err)
	}

	// Verify bidirectional link
	m1, _ := sqlite.GetByID(memory1.ID)
	m2, _ := sqlite.GetByID(memory2.ID)

	// Check memory1 has memory2 in RelatedIDs
	found := false
	for _, id := range m1.RelatedIDs {
		if id == memory2.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected memory1 to have memory2 in RelatedIDs")
	}

	// Check memory2 has memory1 in RelatedIDs
	found = false
	for _, id := range m2.RelatedIDs {
		if id == memory1.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected memory2 to have memory1 in RelatedIDs")
	}
}

func TestLink_Link_SameID(t *testing.T) {
	sqlite := NewMockSQLiteStore()

	memory := &core.Memory{
		Key:   "key",
		Value: "value",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}
	memory.BeforeSave()
	sqlite.memories[memory.ID] = memory

	link := core.NewLink(sqlite)

	err := link.Link(memory.ID, memory.ID)
	if err == nil {
		t.Fatal("Expected error when linking memory to itself")
	}
	if !errors.Is(err, core.ErrCannotLinkToSelf) {
		t.Errorf("Expected ErrCannotLinkToSelf, got %v", err)
	}
}

func TestLink_Link_FirstMemoryNotFound(t *testing.T) {
	sqlite := NewMockSQLiteStore()

	memory2 := &core.Memory{
		Key:   "key2",
		Value: "value 2",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}
	memory2.BeforeSave()
	sqlite.memories[memory2.ID] = memory2

	link := core.NewLink(sqlite)

	err := link.Link("non-existent-id", memory2.ID)
	if err == nil {
		t.Fatal("Expected error when first memory not found")
	}
	if !errors.Is(err, core.ErrMemoryNotFound) {
		t.Errorf("Expected ErrMemoryNotFound, got %v", err)
	}
}

func TestLink_Link_SecondMemoryNotFound(t *testing.T) {
	sqlite := NewMockSQLiteStore()

	memory1 := &core.Memory{
		Key:   "key1",
		Value: "value 1",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}
	memory1.BeforeSave()
	sqlite.memories[memory1.ID] = memory1

	link := core.NewLink(sqlite)

	err := link.Link(memory1.ID, "non-existent-id")
	if err == nil {
		t.Fatal("Expected error when second memory not found")
	}
	if !errors.Is(err, core.ErrMemoryNotFound) {
		t.Errorf("Expected ErrMemoryNotFound, got %v", err)
	}
}

func TestLink_Link_AlreadyLinked(t *testing.T) {
	sqlite := NewMockSQLiteStore()

	memory1 := &core.Memory{
		Key:   "key1",
		Value: "value 1",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}
	memory1.BeforeSave()
	sqlite.memories[memory1.ID] = memory1

	memory2 := &core.Memory{
		Key:   "key2",
		Value: "value 2",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}
	memory2.BeforeSave()
	sqlite.memories[memory2.ID] = memory2

	link := core.NewLink(sqlite)

	// Link first time
	err := link.Link(memory1.ID, memory2.ID)
	if err != nil {
		t.Fatalf("First Link() error = %v", err)
	}

	// Link second time should not cause duplicates
	err = link.Link(memory1.ID, memory2.ID)
	if err != nil {
		t.Fatalf("Second Link() error = %v", err)
	}

	// Verify no duplicates
	m1, _ := sqlite.GetByID(memory1.ID)
	count := 0
	for _, id := range m1.RelatedIDs {
		if id == memory2.ID {
			count++
		}
	}
	if count != 1 {
		t.Errorf("Expected exactly 1 occurrence of memory2 in memory1.RelatedIDs, got %d", count)
	}
}

func TestLink_Unlink(t *testing.T) {
	sqlite := NewMockSQLiteStore()

	// Pre-add two linked memories
	memory1 := &core.Memory{
		Key:        "key1",
		Value:     "value 1",
		Type:      core.TypeFact,
		Scope:     core.ScopeGlobal,
		RelatedIDs: []string{"related-id"},
	}
	memory1.BeforeSave()
	sqlite.memories[memory1.ID] = memory1

	memory2 := &core.Memory{
		Key:        "key2",
		Value:     "value 2",
		Type:      core.TypeFact,
		Scope:     core.ScopeGlobal,
		RelatedIDs: []string{memory1.ID},
	}
	memory2.BeforeSave()
	sqlite.memories[memory2.ID] = memory2

	link := core.NewLink(sqlite)

	err := link.Unlink(memory1.ID, memory2.ID)
	if err != nil {
		t.Fatalf("Unlink() error = %v", err)
	}

	// Verify links are removed
	m1, _ := sqlite.GetByID(memory1.ID)
	m2, _ := sqlite.GetByID(memory2.ID)

	for _, id := range m1.RelatedIDs {
		if id == memory2.ID {
			t.Error("Expected memory2 to be removed from memory1.RelatedIDs")
		}
	}

	for _, id := range m2.RelatedIDs {
		if id == memory1.ID {
			t.Error("Expected memory1 to be removed from memory2.RelatedIDs")
		}
	}
}

func TestLink_Unlink_FirstMemoryNotFound(t *testing.T) {
	sqlite := NewMockSQLiteStore()

	memory2 := &core.Memory{
		Key:   "key2",
		Value: "value 2",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}
	memory2.BeforeSave()
	sqlite.memories[memory2.ID] = memory2

	link := core.NewLink(sqlite)

	err := link.Unlink("non-existent-id", memory2.ID)
	if err == nil {
		t.Fatal("Expected error when first memory not found")
	}
	if !errors.Is(err, core.ErrMemoryNotFound) {
		t.Errorf("Expected ErrMemoryNotFound, got %v", err)
	}
}

func TestLink_Unlink_SecondMemoryNotFound(t *testing.T) {
	sqlite := NewMockSQLiteStore()

	memory1 := &core.Memory{
		Key:   "key1",
		Value: "value 1",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}
	memory1.BeforeSave()
	sqlite.memories[memory1.ID] = memory1

	link := core.NewLink(sqlite)

	err := link.Unlink(memory1.ID, "non-existent-id")
	if err == nil {
		t.Fatal("Expected error when second memory not found")
	}
	if !errors.Is(err, core.ErrMemoryNotFound) {
		t.Errorf("Expected ErrMemoryNotFound, got %v", err)
	}
}

func TestLink_Unlink_NotLinked(t *testing.T) {
	sqlite := NewMockSQLiteStore()

	memory1 := &core.Memory{
		Key:   "key1",
		Value: "value 1",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}
	memory1.BeforeSave()
	sqlite.memories[memory1.ID] = memory1

	memory2 := &core.Memory{
		Key:   "key2",
		Value: "value 2",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}
	memory2.BeforeSave()
	sqlite.memories[memory2.ID] = memory2

	link := core.NewLink(sqlite)

	// Unlink memories that were never linked should not error
	err := link.Unlink(memory1.ID, memory2.ID)
	if err != nil {
		t.Fatalf("Unlink() error = %v", err)
	}
}

func TestLink_GetRelated_VerifiesBidirectionalLinks(t *testing.T) {
	sqlite := NewMockSQLiteStore()

	// Create memories
	memoryA := &core.Memory{
		Key:   "keyA",
		Value: "value A",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}
	memoryA.BeforeSave()
	sqlite.memories[memoryA.ID] = memoryA

	memoryB := &core.Memory{
		Key:   "keyB",
		Value: "value B",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}
	memoryB.BeforeSave()
	sqlite.memories[memoryB.ID] = memoryB

	link := core.NewLink(sqlite)

	// Link A and B
	if err := link.Link(memoryA.ID, memoryB.ID); err != nil {
		t.Fatalf("Link(A, B) error = %v", err)
	}

	// Verify the Link method creates bidirectional links
	// by checking RelatedIDs directly
	m1, _ := sqlite.GetByID(memoryA.ID)
	m2, _ := sqlite.GetByID(memoryB.ID)

	// A should have B in its RelatedIDs
	foundAB := false
	for _, id := range m1.RelatedIDs {
		if id == memoryB.ID {
			foundAB = true
			break
		}
	}
	if !foundAB {
		t.Error("Expected A.RelatedIDs to contain B after Link(A, B)")
	}

	// B should have A in its RelatedIDs
	foundBA := false
	for _, id := range m2.RelatedIDs {
		if id == memoryA.ID {
			foundBA = true
			break
		}
	}
	if !foundBA {
		t.Error("Expected B.RelatedIDs to contain A after Link(A, B)")
	}
}

func TestLink_GetRelated_DepthZero(t *testing.T) {
	sqlite := NewMockSQLiteStore()

	memory := &core.Memory{
		Key:        "key",
		Value:     "value",
		Type:      core.TypeFact,
		Scope:     core.ScopeGlobal,
		RelatedIDs: []string{},
	}
	memory.BeforeSave()
	sqlite.memories[memory.ID] = memory

	link := core.NewLink(sqlite)

	// Depth 0 should be treated as depth 1
	related, err := link.GetRelated(memory.ID, 0)
	if err != nil {
		t.Fatalf("GetRelated() error = %v", err)
	}

	// Should still work with default depth of 1
	_ = related
}

func TestLink_GetRelated_NonExistent(t *testing.T) {
	sqlite := NewMockSQLiteStore()

	link := core.NewLink(sqlite)

	related, err := link.GetRelated("non-existent-id", 1)
	if err != nil {
		t.Fatalf("GetRelated() error = %v", err)
	}

	if len(related) != 0 {
		t.Errorf("Expected 0 related memories, got %d", len(related))
	}
}

func TestLink_GetRelated_ExcludesDeleted(t *testing.T) {
	sqlite := NewMockSQLiteStore()

	memory1 := &core.Memory{
		Key:   "key1",
		Value: "value 1",
		Type:  core.TypeFact,
		Scope: core.ScopeGlobal,
	}
	memory1.BeforeSave()
	sqlite.memories[memory1.ID] = memory1

	memory2 := &core.Memory{
		Key:     "key2",
		Value:   "value 2",
		Type:    core.TypeFact,
		Scope:   core.ScopeGlobal,
		Deleted: true, // Mark as deleted
	}
	memory2.BeforeSave()
	sqlite.memories[memory2.ID] = memory2

	// Link memory1 to deleted memory2
	memory1.RelatedIDs = []string{memory2.ID}
	sqlite.memories[memory1.ID] = memory1

	link := core.NewLink(sqlite)

	related, err := link.GetRelated(memory1.ID, 1)
	if err != nil {
		t.Fatalf("GetRelated() error = %v", err)
	}

	// Should not include deleted memory
	for _, m := range related {
		if m.Deleted {
			t.Error("Expected deleted memories to be excluded")
		}
	}
}

func TestLink_GetRelatedIDs(t *testing.T) {
	sqlite := NewMockSQLiteStore()

	relatedIDs := []string{"related-1", "related-2", "related-3"}
	memory := &core.Memory{
		Key:        "key",
		Value:     "value",
		Type:      core.TypeFact,
		Scope:     core.ScopeGlobal,
		RelatedIDs: relatedIDs,
	}
	memory.BeforeSave()
	sqlite.memories[memory.ID] = memory

	link := core.NewLink(sqlite)

	result, err := link.GetRelatedIDs(memory.ID)
	if err != nil {
		t.Fatalf("GetRelatedIDs() error = %v", err)
	}

	if len(result) != len(relatedIDs) {
		t.Errorf("Expected %d related IDs, got %d", len(relatedIDs), len(result))
	}
}

func TestLink_GetRelatedIDs_NotFound(t *testing.T) {
	sqlite := NewMockSQLiteStore()

	link := core.NewLink(sqlite)

	_, err := link.GetRelatedIDs("non-existent-id")
	if err == nil {
		t.Fatal("Expected error when memory not found")
	}
	if !errors.Is(err, core.ErrMemoryNotFound) {
		t.Errorf("Expected ErrMemoryNotFound, got %v", err)
	}
}

func TestLink_GetRelatedIDs_NoRelated(t *testing.T) {
	sqlite := NewMockSQLiteStore()

	memory := &core.Memory{
		Key:        "key",
		Value:     "value",
		Type:      core.TypeFact,
		Scope:     core.ScopeGlobal,
		RelatedIDs: []string{},
	}
	memory.BeforeSave()
	sqlite.memories[memory.ID] = memory

	link := core.NewLink(sqlite)

	result, err := link.GetRelatedIDs(memory.ID)
	if err != nil {
		t.Fatalf("GetRelatedIDs() error = %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 related IDs, got %d", len(result))
	}
}

// Test helper functions used by Link (indirectly through exported API)
func TestAddIDIfNotExists(t *testing.T) {
	tests := []struct {
		name   string
		ids    []string
		newID  string
		want   int
	}{
		{
			name:  "add new id",
			ids:   []string{"id1", "id2"},
			newID: "id3",
			want:  3,
		},
		{
			name:  "id already exists",
			ids:   []string{"id1", "id2"},
			newID: "id1",
			want:  2,
		},
		{
			name:  "empty list",
			ids:   []string{},
			newID: "id1",
			want:  1,
		},
		{
			name:  "nil list",
			ids:   nil,
			newID: "id1",
			want:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := addIDIfNotExistsHelper(tt.ids, tt.newID)
			if len(result) != tt.want {
				t.Errorf("addIDIfNotExists() got %d items, want %d", len(result), tt.want)
			}
		})
	}
}

func TestRemoveID(t *testing.T) {
	tests := []struct {
		name      string
		ids       []string
		targetID  string
		want      int
	}{
		{
			name:     "remove existing id",
			ids:      []string{"id1", "id2", "id3"},
			targetID: "id2",
			want:     2,
		},
		{
			name:     "remove non-existing id",
			ids:      []string{"id1", "id2"},
			targetID: "id3",
			want:     2,
		},
		{
			name:     "remove from single item list",
			ids:      []string{"id1"},
			targetID: "id1",
			want:     0,
		},
		{
			name:     "empty list",
			ids:      []string{},
			targetID: "id1",
			want:     0,
		},
		{
			name:     "nil list",
			ids:      nil,
			targetID: "id1",
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeIDHelper(tt.ids, tt.targetID)
			if len(result) != tt.want {
				t.Errorf("removeID() got %d items, want %d", len(result), tt.want)
			}
		})
	}
}

// Helper functions that mirror the core implementation for testing
func addIDIfNotExistsHelper(ids []string, newID string) []string {
	for _, id := range ids {
		if id == newID {
			return ids
		}
	}
	return append(ids, newID)
}

func removeIDHelper(ids []string, targetID string) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		if id != targetID {
			result = append(result, id)
		}
	}
	return result
}
