package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"localmemory/core"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStore SQLite 数据库适配器
// 负责 Memory 元数据的持久化存储
// 使用 WAL 模式提高并发性能，支持软删除
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore 创建 SQLiteStore 实例
// 自动创建数据库目录和初始化表结构
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	// 确保数据目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	// 打开数据库连接，启用 WAL 模式和外键约束
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=ON")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 验证连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	s := &SQLiteStore{db: db}
	if err := s.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return s, nil
}

// initSchema 初始化数据库表结构
// 创建 memories 表和必要的索引
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

// Save 保存或更新记忆
// 使用 ON CONFLICT 实现 upsert 语义
// 会自动调用 BeforeSave() 填充默认值
func (s *SQLiteStore) Save(m *core.Memory) error {
	m.BeforeSave()

	// 序列化复杂字段为 JSON
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

// GetByID 根据 ID 获取单条记忆
// 包含已删除的记忆（用于恢复场景）
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

	// 反序列化 JSON 字段
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

// List 分页列出记忆
// 支持按 scope、tags 过滤，可选包含已删除记忆
// 返回记忆列表和符合条件的总数
func (s *SQLiteStore) List(req *core.ListRequest) ([]*core.Memory, int, error) {
	whereClause := "WHERE 1=1"
	args := []any{}

	// 默认排除已删除记忆
	if !req.IncludeDeleted {
		whereClause += " AND deleted = 0"
	}
	// 按作用域过滤
	if req.Scope != "" {
		whereClause += " AND scope = ?"
		args = append(args, req.Scope)
	}
	// 按标签过滤（模糊匹配）
	if len(req.Tags) > 0 {
		for _, tag := range req.Tags {
			whereClause += " AND tags LIKE ?"
			args = append(args, "%"+tag+"%")
		}
	}

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM memories " + whereClause
	var total int
	if err := s.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 查询列表，按更新时间降序
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

		// 反序列化 JSON 字段
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

// Delete 软删除记忆
// 设置 deleted=1 和 deleted_at 时间戳
// 物理删除使用 HardDelete
func (s *SQLiteStore) Delete(id string) error {
	query := `UPDATE memories SET deleted = 1, deleted_at = ? WHERE id = ?`
	_, err := s.db.Exec(query, time.Now().Unix(), id)
	return err
}

// HardDelete 物理删除记忆
// 不可恢复，谨慎使用
func (s *SQLiteStore) HardDelete(id string) error {
	query := `DELETE FROM memories WHERE id = ?`
	_, err := s.db.Exec(query, id)
	return err
}

// GetByKey 根据 key 获取最新的非删除记忆
// 用于 Evolve 机制：查找同 key 记忆进行合并
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

	// 反序列化 JSON 字段
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

// GetStats 获取记忆统计信息
// 包括总数、按类型/作用域/媒体类型分布、已删除数量
func (s *SQLiteStore) GetStats() (*core.StatsResponse, error) {
	stats := &core.StatsResponse{
		ByType:  make(map[string]int),
		ByScope: make(map[string]int),
		ByMedia: make(map[string]int),
	}

	// 查询总数
	totalQuery := `SELECT COUNT(*) FROM memories WHERE deleted = 0`
	if err := s.db.QueryRow(totalQuery).Scan(&stats.Total); err != nil {
		return nil, err
	}

	// 查询已删除数量
	deletedQuery := `SELECT COUNT(*) FROM memories WHERE deleted = 1`
	if err := s.db.QueryRow(deletedQuery).Scan(&stats.Deleted); err != nil {
		return nil, err
	}

	// 按类型分组统计
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

	// 按作用域分组统计
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

	// 按媒体类型分组统计
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

// Close 关闭数据库连接
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// DB 返回底层 sql.DB 实例
// 用于高级操作或事务管理
func (s *SQLiteStore) DB() *sql.DB {
	return s.db
}
