package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"localmemory/core"

	_ "modernc.org/sqlite"
)

// SQLiteStore is the SQLite database adapter.
// Handles Memory metadata persistence storage.
// Uses WAL mode for better concurrency, supports soft delete.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a SQLiteStore instance.
// Automatically creates database directory and initializes schema.
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	// Ensure data directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	// Open database connection with WAL mode and foreign key constraints
	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_foreign_keys=ON")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	s := &SQLiteStore{db: db}
	if err := s.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return s, nil
}

// initSchema initializes the database schema.
// Creates memories table and necessary indexes.
func (s *SQLiteStore) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS memories (
		id          TEXT PRIMARY KEY,
		type        TEXT NOT NULL,
		scope       TEXT NOT NULL,
		media_type  TEXT DEFAULT 'text',
		key         TEXT NOT NULL,
		value       TEXT NOT NULL,
		confidence  REAL DEFAULT 1.0,
		related_ids TEXT,
		tags        TEXT,
		metadata    TEXT,
		deleted     INTEGER DEFAULT 0,
		deleted_at  INTEGER,
		created_at  INTEGER NOT NULL,
		updated_at  INTEGER NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_memories_key ON memories(key);
	CREATE INDEX IF NOT EXISTS idx_memories_scope ON memories(scope);
	CREATE INDEX IF NOT EXISTS idx_memories_type ON memories(type);
	CREATE INDEX IF NOT EXISTS idx_memories_updated ON memories(updated_at);
	CREATE INDEX IF NOT EXISTS idx_memories_media ON memories(media_type);
	CREATE INDEX IF NOT EXISTS idx_memories_deleted ON memories(deleted);
	`

	_, err := s.db.Exec(schema)
	return err
}

// Save saves or updates a memory.
// Uses ON CONFLICT for upsert semantics.
// Automatically calls BeforeSave() to fill defaults.
func (s *SQLiteStore) Save(m *core.Memory) error {
	m.BeforeSave()

	// Serialize complex fields to JSON
	relatedIDs, _ := m.MarshalRelatedIDs()
	tags, _ := m.MarshalTags()
	metadata, _ := m.MarshalMetadata()

	query := `
	INSERT INTO memories (id, type, scope, media_type, key, value, confidence, related_ids, tags, metadata, deleted, deleted_at, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		type = excluded.type,
		scope = excluded.scope,
		media_type = excluded.media_type,
		key = excluded.key,
		value = excluded.value,
		confidence = excluded.confidence,
		related_ids = excluded.related_ids,
		tags = excluded.tags,
		metadata = excluded.metadata,
		deleted = excluded.deleted,
		deleted_at = excluded.deleted_at,
		updated_at = excluded.updated_at
	`

	_, err := s.db.Exec(query,
		m.ID, m.Type, m.Scope, m.MediaType, m.Key, m.Value,
		m.Confidence, relatedIDs, tags, metadata,
		m.Deleted, m.DeletedAt, m.CreatedAt, m.UpdatedAt,
	)

	return err
}

// GetByID retrieves a single memory by ID.
// Includes deleted memories (used for recovery scenarios).
func (s *SQLiteStore) GetByID(id string) (*core.Memory, error) {
	query := `
	SELECT id, type, scope, media_type, key, value, confidence, related_ids, tags, metadata, deleted, deleted_at, created_at, updated_at
	FROM memories
	WHERE id = ?
	`

	var m core.Memory
	var relatedIDs, tags, metadata sql.NullString

	err := s.db.QueryRow(query, id).Scan(
		&m.ID, &m.Type, &m.Scope, &m.MediaType, &m.Key, &m.Value,
		&m.Confidence, &relatedIDs, &tags, &metadata,
		&m.Deleted, &m.DeletedAt, &m.CreatedAt, &m.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Deserialize JSON fields
	if relatedIDs.Valid {
		m.RelatedIDs, _ = core.UnmarshalRelatedIDs(relatedIDs.String)
	}
	if tags.Valid {
		m.Tags, _ = core.UnmarshalTags(tags.String)
	}
	if metadata.Valid {
		m.Metadata, _ = core.UnmarshalMetadata(metadata.String)
	}

	return &m, nil
}

// List lists memories with pagination.
// Supports filtering by scope and tags, optionally includes deleted memories.
// Returns memory list and total count matching conditions.
func (s *SQLiteStore) List(req *core.ListRequest) ([]*core.Memory, int, error) {
	whereClause := "WHERE 1=1"
	args := []any{}

	// Exclude deleted memories by default
	if !req.IncludeDeleted {
		whereClause += " AND deleted = 0"
	}
	// Filter by scope
	if req.Scope != "" {
		whereClause += " AND scope = ?"
		args = append(args, req.Scope)
	}
	// Filter by tags (fuzzy match)
	if len(req.Tags) > 0 {
		for _, tag := range req.Tags {
			whereClause += " AND tags LIKE ?"
			args = append(args, "%"+tag+"%")
		}
	}

	// Query total count
	countQuery := "SELECT COUNT(*) FROM memories " + whereClause
	var total int
	if err := s.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Query list, ordered by update time descending
	query := `
	SELECT id, type, scope, media_type, key, value, confidence, related_ids, tags, metadata, deleted, deleted_at, created_at, updated_at
	FROM memories
	` + whereClause + `
	ORDER BY updated_at DESC
	LIMIT ? OFFSET ?
	`
	args = append(args, req.Limit, req.Offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var memories []*core.Memory
	for rows.Next() {
		var m core.Memory
		var relatedIDs, tags, metadata sql.NullString

		if err := rows.Scan(
			&m.ID, &m.Type, &m.Scope, &m.MediaType, &m.Key, &m.Value,
			&m.Confidence, &relatedIDs, &tags, &metadata,
			&m.Deleted, &m.DeletedAt, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		// Deserialize JSON fields
		if relatedIDs.Valid {
			m.RelatedIDs, _ = core.UnmarshalRelatedIDs(relatedIDs.String)
		}
		if tags.Valid {
			m.Tags, _ = core.UnmarshalTags(tags.String)
		}
		if metadata.Valid {
			m.Metadata, _ = core.UnmarshalMetadata(metadata.String)
		}

		memories = append(memories, &m)
	}

	return memories, total, nil
}

// Delete soft deletes a memory.
// Sets deleted=1 and deleted_at timestamp.
// Physical deletion uses HardDelete.
func (s *SQLiteStore) Delete(id string) error {
	query := `UPDATE memories SET deleted = 1, deleted_at = ? WHERE id = ?`
	_, err := s.db.Exec(query, time.Now().Unix(), id)
	return err
}

// HardDelete physically deletes a memory.
// Cannot be recovered, use with caution.
func (s *SQLiteStore) HardDelete(id string) error {
	query := `DELETE FROM memories WHERE id = ?`
	_, err := s.db.Exec(query, id)
	return err
}

// GetByKey retrieves the latest non-deleted memory by key.
// Used for Evolve mechanism: finds memory with same key for merging.
func (s *SQLiteStore) GetByKey(key string) (*core.Memory, error) {
	query := `
	SELECT id, type, scope, media_type, key, value, confidence, related_ids, tags, metadata, deleted, deleted_at, created_at, updated_at
	FROM memories
	WHERE key = ? AND deleted = 0
	ORDER BY updated_at DESC
	LIMIT 1
	`

	var m core.Memory
	var relatedIDs, tags, metadata sql.NullString

	err := s.db.QueryRow(query, key).Scan(
		&m.ID, &m.Type, &m.Scope, &m.MediaType, &m.Key, &m.Value,
		&m.Confidence, &relatedIDs, &tags, &metadata,
		&m.Deleted, &m.DeletedAt, &m.CreatedAt, &m.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Deserialize JSON fields
	if relatedIDs.Valid {
		m.RelatedIDs, _ = core.UnmarshalRelatedIDs(relatedIDs.String)
	}
	if tags.Valid {
		m.Tags, _ = core.UnmarshalTags(tags.String)
	}
	if metadata.Valid {
		m.Metadata, _ = core.UnmarshalMetadata(metadata.String)
	}

	return &m, nil
}

// GetStats retrieves memory statistics.
// Includes total count, distribution by type/scope/media type, and deleted count.
func (s *SQLiteStore) GetStats() (*core.StatsResponse, error) {
	stats := &core.StatsResponse{
		ByType:  make(map[string]int),
		ByScope: make(map[string]int),
		ByMedia: make(map[string]int),
	}

	// Query total count
	totalQuery := `SELECT COUNT(*) FROM memories WHERE deleted = 0`
	if err := s.db.QueryRow(totalQuery).Scan(&stats.Total); err != nil {
		return nil, err
	}

	// Query deleted count
	deletedQuery := `SELECT COUNT(*) FROM memories WHERE deleted = 1`
	if err := s.db.QueryRow(deletedQuery).Scan(&stats.Deleted); err != nil {
		return nil, err
	}

	// Query count grouped by type
	typeQuery := `SELECT type, COUNT(*) FROM memories WHERE deleted = 0 GROUP BY type`
	rows, err := s.db.Query(typeQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t string
		var count int
		if err := rows.Scan(&t, &count); err != nil {
			return nil, err
		}
		stats.ByType[t] = count
	}

	// Query count grouped by scope
	scopeQuery := `SELECT scope, COUNT(*) FROM memories WHERE deleted = 0 GROUP BY scope`
	rows, err = s.db.Query(scopeQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var sc string
		var count int
		if err := rows.Scan(&sc, &count); err != nil {
			return nil, err
		}
		stats.ByScope[sc] = count
	}

	// Query count grouped by media type
	mediaQuery := `SELECT media_type, COUNT(*) FROM memories WHERE deleted = 0 GROUP BY media_type`
	rows, err = s.db.Query(mediaQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var mt string
		var count int
		if err := rows.Scan(&mt, &count); err != nil {
			return nil, err
		}
		stats.ByMedia[mt] = count
	}

	return stats, nil
}

// Close closes the database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// DB returns the underlying sql.DB instance.
// Used for advanced operations or transaction management.
func (s *SQLiteStore) DB() *sql.DB {
	return s.db
}
