package unit

import (
	"testing"

	"localmemory/core"
)

func TestMemory_BeforeSave(t *testing.T) {
	// Test case 1: New memory with empty fields
	m := &core.Memory{
		Value: "test value",
	}
	m.BeforeSave()

	if m.ID == "" {
		t.Error("Expected ID to be generated")
	}
	if m.CreatedAt == 0 {
		t.Error("Expected CreatedAt to be set")
	}
	if m.UpdatedAt == 0 {
		t.Error("Expected UpdatedAt to be set")
	}
	if m.Confidence != 1.0 {
		t.Errorf("Expected default Confidence to be 1.0, got %f", m.Confidence)
	}
	if m.MediaType != core.MediaText {
		t.Errorf("Expected default MediaType to be text, got %s", m.MediaType)
	}

	// Test case 2: Memory with existing ID should not regenerate
	existingID := m.ID
	m2 := &core.Memory{
		ID:    existingID,
		Value: "test 2",
	}
	m2.BeforeSave()
	if m2.ID != existingID {
		t.Error("Existing ID should not be overwritten")
	}

	// Test case 3: Existing timestamps should not be overwritten
	m3 := &core.Memory{
		Value:     "test 3",
		CreatedAt: 1000,
		UpdatedAt: 1000,
	}
	before := m3.CreatedAt
	m3.BeforeSave()
	if m3.CreatedAt != before {
		t.Error("Existing CreatedAt should not be overwritten")
	}
}

func TestMemory_MarshalMetadata(t *testing.T) {
	// Test case 1: Empty metadata
	m := &core.Memory{}
	_, err := m.MarshalMetadata()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Test case 2: Metadata with values
	m2 := &core.Memory{
		Metadata: core.Metadata{
			Source:   "cli",
			Language: "en",
		},
	}
	result, err := m2.MarshalMetadata()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result == "" {
		t.Error("Expected non-empty metadata string")
	}
}

func TestMemory_UnmarshalMetadata(t *testing.T) {
	// Test case 1: Empty string
	m, err := core.UnmarshalMetadata("")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if m.Source != "" {
		t.Error("Expected empty source")
	}

	// Test case 2: Valid JSON
	m2, err := core.UnmarshalMetadata(`{"source":"cli","language":"en"}`)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if m2.Source != "cli" {
		t.Errorf("Expected source 'cli', got '%s'", m2.Source)
	}
	if m2.Language != "en" {
		t.Errorf("Expected language 'en', got '%s'", m2.Language)
	}
}

func TestMemory_RelatedIDs(t *testing.T) {
	// Test marshal
	m := &core.Memory{
		RelatedIDs: []string{"id1", "id2"},
	}
	result, err := m.MarshalRelatedIDs()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result == "" {
		t.Error("Expected non-empty related_ids string")
	}

	// Test unmarshal
	ids, err := core.UnmarshalRelatedIDs(`["id1","id2"]`)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(ids) != 2 {
		t.Errorf("Expected 2 IDs, got %d", len(ids))
	}

	// Test empty
	ids2, err := core.UnmarshalRelatedIDs("")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if ids2 != nil {
		t.Error("Expected nil for empty string")
	}
}

func TestMemory_Tags(t *testing.T) {
	// Test marshal
	m := &core.Memory{
		Tags: []string{"tag1", "tag2"},
	}
	result, err := m.MarshalTags()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result == "" {
		t.Error("Expected non-empty tags string")
	}

	// Test unmarshal
	tags, err := core.UnmarshalTags(`["tag1","tag2"]`)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(tags))
	}
}

func TestGenerateID(t *testing.T) {
	id1 := core.GenerateID()
	id2 := core.GenerateID()

	if id1 == "" {
		t.Error("Expected non-empty ID")
	}
	if id1 == id2 {
		t.Error("Expected unique IDs")
	}
	if len(id1) != 36 { // UUID v4 format
		t.Errorf("Expected UUID v4 format, got %d chars", len(id1))
	}
}
