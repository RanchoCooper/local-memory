package core

import (
	"encoding/json"
	"time"
)

// Memory is the fundamental memory unit for AI agents.
// It contains type, scope, content, associations, and other metadata.
type Memory struct {
	ID         string     `json:"id"`                      // Unique identifier, UUID format
	Type       MemoryType `json:"type"`                    // Memory type: preference | fact | event
	Scope      Scope      `json:"scope"`                   // Scope: global | session | agent
	MediaType  MediaType  `json:"media_type"`             // Media type: text | image | audio | video
	Key        string     `json:"key"`                     // Memory key for unique identification and retrieval
	Value      string     `json:"value"`                  // Memory value, text content or media path
	Confidence float64    `json:"confidence"`              // Confidence score, 0.0~1.0
	RelatedIDs []string   `json:"related_ids"`            // Related memory ID list, supports graph associations
	Tags       []string   `json:"tags"`                   // Tag list for categorization and retrieval
	Metadata   Metadata   `json:"metadata"`                // Extended metadata: source, language, file info, etc.
	Deleted    bool       `json:"deleted"`                // Soft delete flag, true means deleted
	DeletedAt  int64      `json:"deleted_at"`             // Delete timestamp, Unix time
	Embedding  []float32 `json:"-"`                      // Vector embedding, in-memory only, not persisted to database
	CreatedAt  int64     `json:"created_at"`             // Creation timestamp, Unix time
	UpdatedAt  int64     `json:"updated_at"`             // Update timestamp, Unix time
}

// MemoryType represents the type of memory.
// Used to distinguish different types of memories for categorization and retrieval.
type MemoryType string

const (
	TypePreference   MemoryType = "preference"   // User preference, e.g., programming language, design style
	TypeFact         MemoryType = "fact"         // Objective fact, e.g., project architecture, tech stack
	TypeEvent        MemoryType = "event"        // Event record, e.g., completed features, fixed bugs
	TypeSkill        MemoryType = "skill"         // Skill/ability, e.g., frameworks, tools used
	TypeGoal         MemoryType = "goal"          // Goal/intent, e.g., features to implement
	TypeRelationship MemoryType = "relationship"  // Relationship, e.g., association with a module
)

// Scope represents the visibility and sharing scope of memory.
type Scope string

const (
	ScopeGlobal  Scope = "global"  // Global shared, visible to all agents and sessions
	ScopeSession Scope = "session" // Session level, only visible in current session
	ScopeAgent   Scope = "agent"  // Agent private, only visible to current agent
)

// MediaType represents the type of media content.
// MVP supports text and image, audio/video are reserved for future.
type MediaType string

const (
	MediaText   MediaType = "text"   // Text (default)
	MediaImage  MediaType = "image"  // Image (MVP supported)
	MediaAudio  MediaType = "audio"  // Audio (reserved)
	MediaVideo  MediaType = "video"  // Video (reserved)
)

// Metadata contains extended information about the memory.
// Such as source, associated files, etc.
type Metadata struct {
	Source     string         `json:"source,omitempty"`      // Source: claude_code, user_input, api, etc.
	Language   string         `json:"language,omitempty"`    // Content language, e.g., zh, en
	FilePath   string         `json:"file_path,omitempty"`   // Associated file path
	FileSize   int64          `json:"file_size,omitempty"`   // File size in bytes
	MimeType   string         `json:"mime_type,omitempty"`   // MIME type, e.g., image/png
	AgentID    string         `json:"agent_id,omitempty"`    // Source agent identifier
	SessionID  string         `json:"session_id,omitempty"`  // Session identifier
	Extra      map[string]any `json:"extra,omitempty"`       // Extension fields, custom key-value pairs
}

// QueryRequest represents a semantic search request.
type QueryRequest struct {
	Query string   `json:"query"`                      // Query text, natural language description
	TopK  int      `json:"topk"`                       // Maximum number of results to return
	Scope Scope    `json:"scope,omitempty"`            // Optional: scope filter
	Tags  []string `json:"tags,omitempty"`             // Optional: tag filter
}

// QueryResult represents a single search result.
type QueryResult struct {
	Memory *Memory `json:"memory"` // Matched memory
	Score  float64 `json:"score"`  // Relevance score, 0.0~1.0
}

// QueryResponse represents a semantic search response.
type QueryResponse struct {
	Results []*QueryResult `json:"results"` // Result list, sorted by score descending
}

// ListRequest represents a list query request.
type ListRequest struct {
	Scope          Scope    `json:"scope,omitempty"`            // Optional: scope filter
	Tags           []string `json:"tags,omitempty"`             // Optional: tag filter
	Limit          int      `json:"limit"`                      // Page size
	Offset         int      `json:"offset"`                     // Offset
	IncludeDeleted bool     `json:"include_deleted"`            // Whether to include deleted memories
}

// ListResponse represents a list query response.
type ListResponse struct {
	Memories []*Memory `json:"memories"` // Memory list
	Total    int       `json:"total"`    // Total count
}

// StatsResponse represents statistics information.
type StatsResponse struct {
	Total   int            `json:"total"`    // Total memory count
	ByType  map[string]int `json:"by_type"`  // Statistics by type
	ByScope map[string]int `json:"by_scope"` // Statistics by scope
	ByMedia map[string]int `json:"by_media"` // Statistics by media type
	Deleted int            `json:"deleted"`  // Deleted memory count
}

// BeforeSave performs preprocessing before saving.
// Automatically fills ID, timestamps, default values, etc.
func (m *Memory) BeforeSave() {
	now := time.Now().Unix()
	if m.ID == "" {
		m.ID = GenerateID() // Generate UUID for empty ID
	}
	if m.CreatedAt == 0 {
		m.CreatedAt = now // Set creation time on first save
	}
	if m.UpdatedAt == 0 {
		m.UpdatedAt = now // Set update time on first save
	}
	if m.Confidence == 0 {
		m.Confidence = 1.0 // Default confidence is 1.0
	}
	if m.MediaType == "" {
		m.MediaType = MediaText // Default media type is text
	}
}

// MarshalMetadata serializes Metadata struct to JSON string.
func (m *Memory) MarshalMetadata() (string, error) {
	if len(m.Metadata.Extra) == 0 && m.Metadata.Source == "" && m.Metadata.Language == "" {
		return "", nil // Skip serialization for empty Metadata
	}
	data, err := json.Marshal(m.Metadata)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// UnmarshalMetadata deserializes JSON string to Metadata struct.
func UnmarshalMetadata(data string) (Metadata, error) {
	if data == "" {
		return Metadata{}, nil
	}
	var m Metadata
	err := json.Unmarshal([]byte(data), &m)
	return m, err
}

// MarshalRelatedIDs serializes related memory ID list to JSON string.
func (m *Memory) MarshalRelatedIDs() (string, error) {
	if len(m.RelatedIDs) == 0 {
		return "", nil
	}
	data, err := json.Marshal(m.RelatedIDs)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// UnmarshalRelatedIDs deserializes JSON string to related memory ID list.
func UnmarshalRelatedIDs(data string) ([]string, error) {
	if data == "" {
		return nil, nil
	}
	var ids []string
	err := json.Unmarshal([]byte(data), &ids)
	return ids, err
}

// MarshalTags serializes tag list to JSON string.
func (m *Memory) MarshalTags() (string, error) {
	if len(m.Tags) == 0 {
		return "", nil
	}
	data, err := json.Marshal(m.Tags)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// UnmarshalTags deserializes JSON string to tag list.
func UnmarshalTags(data string) ([]string, error) {
	if data == "" {
		return nil, nil
	}
	var tags []string
	err := json.Unmarshal([]byte(data), &tags)
	return tags, err
}
