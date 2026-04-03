package unit

import (
	"encoding/json"
	"fmt"
	"testing"

	"localmemory/core"
)

func TestMemory_BeforeSave(t *testing.T) {
	tests := []struct {
		name      string
		memory    *core.Memory
		checkFunc func(*core.Memory) error
	}{
		{
			name: "generates ID when empty",
			memory: &core.Memory{
				Value: "test value",
			},
			checkFunc: func(m *core.Memory) error {
				if m.ID == "" {
					return errExpected("ID to be generated")
				}
				return nil
			},
		},
		{
			name: "preserves existing ID",
			memory: &core.Memory{
				ID:    "existing-id-123",
				Value: "test value",
			},
			checkFunc: func(m *core.Memory) error {
				if m.ID != "existing-id-123" {
					return errExpected("existing ID to be preserved")
				}
				return nil
			},
		},
		{
			name: "sets CreatedAt when zero",
			memory: &core.Memory{
				Value: "test value",
			},
			checkFunc: func(m *core.Memory) error {
				if m.CreatedAt == 0 {
					return errExpected("CreatedAt to be set")
				}
				return nil
			},
		},
		{
			name: "preserves existing CreatedAt",
			memory: &core.Memory{
				Value:     "test value",
				CreatedAt: 1000,
			},
			checkFunc: func(m *core.Memory) error {
				if m.CreatedAt != 1000 {
					return errExpected("existing CreatedAt to be preserved")
				}
				return nil
			},
		},
		{
			name: "sets UpdatedAt when zero",
			memory: &core.Memory{
				Value: "test value",
			},
			checkFunc: func(m *core.Memory) error {
				if m.UpdatedAt == 0 {
					return errExpected("UpdatedAt to be set")
				}
				return nil
			},
		},
		{
			name: "preserves existing UpdatedAt",
			memory: &core.Memory{
				Value:     "test value",
				UpdatedAt: 2000,
			},
			checkFunc: func(m *core.Memory) error {
				if m.UpdatedAt != 2000 {
					return errExpected("existing UpdatedAt to be preserved")
				}
				return nil
			},
		},
		{
			name: "sets default Confidence to 1.0",
			memory: &core.Memory{
				Value: "test value",
			},
			checkFunc: func(m *core.Memory) error {
				if m.Confidence != 1.0 {
					return errExpectedf("Confidence to be 1.0, got %f", m.Confidence)
				}
				return nil
			},
		},
		{
			name: "preserves non-zero Confidence",
			memory: &core.Memory{
				Value:      "test value",
				Confidence: 0.5,
			},
			checkFunc: func(m *core.Memory) error {
				if m.Confidence != 0.5 {
					return errExpectedf("Confidence to be 0.5, got %f", m.Confidence)
				}
				return nil
			},
		},
		{
			name: "sets default MediaType to text",
			memory: &core.Memory{
				Value: "test value",
			},
			checkFunc: func(m *core.Memory) error {
				if m.MediaType != core.MediaText {
					return errExpectedf("MediaType to be text, got %s", m.MediaType)
				}
				return nil
			},
		},
		{
			name: "preserves existing MediaType",
			memory: &core.Memory{
				Value:     "test value",
				MediaType: core.MediaImage,
			},
			checkFunc: func(m *core.Memory) error {
				if m.MediaType != core.MediaImage {
					return errExpectedf("MediaType to be image, got %s", m.MediaType)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.memory.BeforeSave()
			if err := tt.checkFunc(tt.memory); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestMemory_MarshalMetadata(t *testing.T) {
	tests := []struct {
		name      string
		memory    *core.Memory
		wantEmpty bool
		wantErr   bool
	}{
		{
			name:      "empty metadata returns empty string",
			memory:    &core.Memory{},
			wantEmpty: true,
		},
		{
			name: "metadata with source",
			memory: &core.Memory{
				Metadata: core.Metadata{
					Source: "cli",
				},
			},
			wantEmpty: false,
		},
		{
			name: "metadata with language",
			memory: &core.Memory{
				Metadata: core.Metadata{
					Language: "en",
				},
			},
			wantEmpty: false,
		},
		{
			name: "metadata with extra fields",
			memory: &core.Memory{
				Metadata: core.Metadata{
					Extra: map[string]any{"key": "value"},
				},
			},
			wantEmpty: false,
		},
		{
			name: "metadata with multiple fields",
			memory: &core.Memory{
				Metadata: core.Metadata{
					Source:   "cli",
					Language: "en",
					Extra:    map[string]any{"custom": "data"},
				},
			},
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.memory.MarshalMetadata()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantEmpty && result != "" {
				t.Errorf("MarshalMetadata() = %q, want empty", result)
			}
			if !tt.wantEmpty && result == "" {
				t.Error("MarshalMetadata() returned empty string, want non-empty")
			}
		})
	}
}

func TestUnmarshalMetadata(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    core.Metadata
		wantErr bool
	}{
		{
			name:    "empty string returns empty metadata",
			data:    "",
			want:    core.Metadata{},
			wantErr: false,
		},
		{
			name:    "valid JSON with source",
			data:    `{"source":"cli"}`,
			want:    core.Metadata{Source: "cli"},
			wantErr: false,
		},
		{
			name:    "valid JSON with language",
			data:    `{"language":"en"}`,
			want:    core.Metadata{Language: "en"},
			wantErr: false,
		},
		{
			name: "valid JSON with multiple fields",
			data:  `{"source":"cli","language":"en","file_path":"/tmp/test"}`,
			want: core.Metadata{
				Source:   "cli",
				Language: "en",
				FilePath: "/tmp/test",
			},
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			data:    `{invalid}`,
			want:    core.Metadata{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := core.UnmarshalMetadata(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Source != tt.want.Source {
				t.Errorf("Source = %v, want %v", got.Source, tt.want.Source)
			}
			if got.Language != tt.want.Language {
				t.Errorf("Language = %v, want %v", got.Language, tt.want.Language)
			}
			if got.FilePath != tt.want.FilePath {
				t.Errorf("FilePath = %v, want %v", got.FilePath, tt.want.FilePath)
			}
		})
	}
}

func TestMemory_MarshalRelatedIDs(t *testing.T) {
	tests := []struct {
		name      string
		memory    *core.Memory
		wantEmpty bool
		wantErr   bool
	}{
		{
			name:      "nil RelatedIDs returns empty",
			memory:    &core.Memory{RelatedIDs: nil},
			wantEmpty: true,
		},
		{
			name:      "empty RelatedIDs returns empty",
			memory:    &core.Memory{RelatedIDs: []string{}},
			wantEmpty: true,
		},
		{
			name: "single ID",
			memory: &core.Memory{
				RelatedIDs: []string{"id1"},
			},
			wantEmpty: false,
		},
		{
			name: "multiple IDs",
			memory: &core.Memory{
				RelatedIDs: []string{"id1", "id2", "id3"},
			},
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.memory.MarshalRelatedIDs()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalRelatedIDs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantEmpty && result != "" {
				t.Errorf("MarshalRelatedIDs() = %q, want empty", result)
			}
			if !tt.wantEmpty && result == "" {
				t.Error("MarshalRelatedIDs() returned empty, want non-empty")
			}
		})
	}
}

func TestUnmarshalRelatedIDs(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    []string
		wantErr bool
	}{
		{
			name:    "empty string returns nil",
			data:    "",
			want:    nil,
			wantErr: false,
		},
		{
			name:    "valid single ID",
			data:    `["id1"]`,
			want:    []string{"id1"},
			wantErr: false,
		},
		{
			name:    "valid multiple IDs",
			data:    `["id1","id2","id3"]`,
			want:    []string{"id1", "id2", "id3"},
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			data:    `invalid`,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := core.UnmarshalRelatedIDs(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalRelatedIDs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("UnmarshalRelatedIDs() got %d items, want %d", len(got), len(tt.want))
			}
		})
	}
}

func TestMemory_MarshalTags(t *testing.T) {
	tests := []struct {
		name      string
		memory    *core.Memory
		wantEmpty bool
		wantErr   bool
	}{
		{
			name:      "nil Tags returns empty",
			memory:    &core.Memory{Tags: nil},
			wantEmpty: true,
		},
		{
			name:      "empty Tags returns empty",
			memory:    &core.Memory{Tags: []string{}},
			wantEmpty: true,
		},
		{
			name: "single tag",
			memory: &core.Memory{
				Tags: []string{"tag1"},
			},
			wantEmpty: false,
		},
		{
			name: "multiple tags",
			memory: &core.Memory{
				Tags: []string{"tag1", "tag2", "tag3"},
			},
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.memory.MarshalTags()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantEmpty && result != "" {
				t.Errorf("MarshalTags() = %q, want empty", result)
			}
			if !tt.wantEmpty && result == "" {
				t.Error("MarshalTags() returned empty, want non-empty")
			}
		})
	}
}

func TestUnmarshalTags(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    []string
		wantErr bool
	}{
		{
			name:    "empty string returns nil",
			data:    "",
			want:    nil,
			wantErr: false,
		},
		{
			name:    "valid single tag",
			data:    `["tag1"]`,
			want:    []string{"tag1"},
			wantErr: false,
		},
		{
			name:    "valid multiple tags",
			data:    `["tag1","tag2","tag3"]`,
			want:    []string{"tag1", "tag2", "tag3"},
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			data:    `invalid`,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := core.UnmarshalTags(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("UnmarshalTags() got %d items, want %d", len(got), len(tt.want))
			}
		})
	}
}

func TestMemory_MarshalRoundTrip(t *testing.T) {
	memory := &core.Memory{
		ID:         "test-id-123",
		Type:       core.TypePreference,
		Scope:      core.ScopeGlobal,
		MediaType:  core.MediaText,
		Key:        "test-key",
		Value:      "test value content",
		Confidence: 0.85,
		RelatedIDs: []string{"related-1", "related-2"},
		Tags:       []string{"tag1", "tag2"},
		Metadata: core.Metadata{
			Source:   "test",
			Language: "en",
			Extra:    map[string]any{"custom": "value"},
		},
	}

	// Marshal RelatedIDs
	idsJSON, err := memory.MarshalRelatedIDs()
	if err != nil {
		t.Fatalf("MarshalRelatedIDs() error = %v", err)
	}

	ids, err := core.UnmarshalRelatedIDs(idsJSON)
	if err != nil {
		t.Fatalf("UnmarshalRelatedIDs() error = %v", err)
	}
	if len(ids) != len(memory.RelatedIDs) {
		t.Errorf("RelatedIDs round trip: got %d, want %d", len(ids), len(memory.RelatedIDs))
	}

	// Marshal Tags
	tagsJSON, err := memory.MarshalTags()
	if err != nil {
		t.Fatalf("MarshalTags() error = %v", err)
	}

	tags, err := core.UnmarshalTags(tagsJSON)
	if err != nil {
		t.Fatalf("UnmarshalTags() error = %v", err)
	}
	if len(tags) != len(memory.Tags) {
		t.Errorf("Tags round trip: got %d, want %d", len(tags), len(memory.Tags))
	}

	// Marshal Metadata
	metadataJSON, err := memory.MarshalMetadata()
	if err != nil {
		t.Fatalf("MarshalMetadata() error = %v", err)
	}

	metadata, err := core.UnmarshalMetadata(metadataJSON)
	if err != nil {
		t.Fatalf("UnmarshalMetadata() error = %v", err)
	}
	if metadata.Source != memory.Metadata.Source {
		t.Errorf("Metadata.Source round trip: got %s, want %s", metadata.Source, memory.Metadata.Source)
	}
}

func TestGenerateID(t *testing.T) {
	// Test that GenerateID returns non-empty string
	id1 := core.GenerateID()
	if id1 == "" {
		t.Error("Expected non-empty ID")
	}

	// Test uniqueness
	id2 := core.GenerateID()
	if id1 == id2 {
		t.Error("Expected unique IDs, got same value")
	}

	// Test UUID format (36 characters with hyphens)
	if len(id1) != 36 {
		t.Errorf("Expected UUID v4 format (36 chars), got %d chars", len(id1))
	}

	// Verify it's valid JSON (can be marshaled)
	var idJSON map[string]string
	if err := json.Unmarshal([]byte(`{"id":"`+id1+`"}`), &idJSON); err != nil {
		t.Errorf("Generated ID is not valid JSON string: %v", err)
	}
}

func TestMemoryType_Constants(t *testing.T) {
	tests := []struct {
		name  string
		mtype core.MemoryType
		want  string
	}{
		{"TypePreference", core.TypePreference, "preference"},
		{"TypeFact", core.TypeFact, "fact"},
		{"TypeEvent", core.TypeEvent, "event"},
		{"TypeSkill", core.TypeSkill, "skill"},
		{"TypeGoal", core.TypeGoal, "goal"},
		{"TypeRelationship", core.TypeRelationship, "relationship"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.mtype) != tt.want {
				t.Errorf("MemoryType = %v, want %v", tt.mtype, tt.want)
			}
		})
	}
}

func TestScope_Constants(t *testing.T) {
	tests := []struct {
		name  string
		scope core.Scope
		want  string
	}{
		{"ScopeGlobal", core.ScopeGlobal, "global"},
		{"ScopeSession", core.ScopeSession, "session"},
		{"ScopeAgent", core.ScopeAgent, "agent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.scope) != tt.want {
				t.Errorf("Scope = %v, want %v", tt.scope, tt.want)
			}
		})
	}
}

func TestMediaType_Constants(t *testing.T) {
	tests := []struct {
		name  string
		media core.MediaType
		want  string
	}{
		{"MediaText", core.MediaText, "text"},
		{"MediaImage", core.MediaImage, "image"},
		{"MediaAudio", core.MediaAudio, "audio"},
		{"MediaVideo", core.MediaVideo, "video"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.media) != tt.want {
				t.Errorf("MediaType = %v, want %v", tt.media, tt.want)
			}
		})
	}
}

// Helper functions for error messages
func errExpected(msg string) error {
	return &testError{msg: msg}
}

func errExpectedf(format string, args ...any) error {
	return &testError{msg: fmt.Sprintf(format, args...)}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
