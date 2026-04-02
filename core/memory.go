package core

import (
	"encoding/json"
	"time"
)

// Memory 记忆结构体
// AI Agent 的基本记忆单元，包含类型、作用域、内容、关联等信息
type Memory struct {
	ID         string     `json:"id"`                      // 唯一标识符，UUID 格式
	Type       MemoryType `json:"type"`                    // 记忆类型：preference(偏好) | fact(事实) | event(事件)
	Scope      Scope      `json:"scope"`                   // 作用域：global(全局) | session(会话级) | agent(Agent 私有)
	MediaType  MediaType  `json:"media_type"`             // 媒体类型：text(文本) | image(图像) | audio(语音) | video(视频)
	Key        string     `json:"key"`                     // 记忆的键，用于唯一标识和检索
	Value      string     `json:"value"`                  // 记忆的值，文本内容或媒体路径
	Confidence float64    `json:"confidence"`              // 置信度，0.0~1.0，表示记忆的可信程度
	RelatedIDs []string   `json:"related_ids"`            // 关联记忆 ID 列表，支持图谱式关联
	Tags       []string   `json:"tags"`                   // 标签列表，便于分类和检索
	Metadata   Metadata   `json:"metadata"`                // 扩展元数据，包含来源、语言、文件信息等
	Deleted    bool       `json:"deleted"`                // 软删除标记，true 表示已删除
	DeletedAt  int64      `json:"deleted_at"`             // 删除时间戳，Unix 时间
	Embedding  []float32 `json:"-"`                      // 向量嵌入，内存中使用，不持久化到数据库
	CreatedAt  int64     `json:"created_at"`             // 创建时间戳，Unix 时间
	UpdatedAt  int64     `json:"updated_at"`             // 更新时间戳，Unix 时间
}

// MemoryType 记忆类型枚举
// 用于区分不同性质的记忆，便于分类检索和管理
type MemoryType string

const (
	TypePreference   MemoryType = "preference"   // 用户偏好，如编程语言、设计风格等
	TypeFact         MemoryType = "fact"         // 客观事实，如项目架构、技术栈等
	TypeEvent        MemoryType = "event"        // 事件记录，如完成的功能、修复的 bug 等
	TypeSkill        MemoryType = "skill"         // 技能/能力，如使用的框架、工具等
	TypeGoal         MemoryType = "goal"          // 目标/意图，如要实现的功能等
	TypeRelationship MemoryType = "relationship"  // 关系，如与某个模块的关联等
)

// Scope 记忆作用域枚举
// 控制记忆的可见性和共享范围
type Scope string

const (
	ScopeGlobal  Scope = "global"  // 全局共享，所有 Agent 和会话可见
	ScopeSession Scope = "session" // 会话级，仅当前会话可见
	ScopeAgent   Scope = "agent"  // Agent 私有，仅当前 Agent 可见
)

// MediaType 媒体类型枚举
// MVP 阶段支持 text 和 image，语音/视频预留
type MediaType string

const (
	MediaText   MediaType = "text"   // 文本（默认）
	MediaImage  MediaType = "image"  // 图像（MVP 支持）
	MediaAudio  MediaType = "audio"  // 语音（预留）
	MediaVideo  MediaType = "video"  // 视频（预留）
)

// Metadata 扩展元数据结构
// 存储记忆的附加信息，如来源、关联文件等
type Metadata struct {
	Source     string         `json:"source,omitempty"`      // 来源：claude_code, user_input, api 等
	Language   string         `json:"language,omitempty"`    // 内容语言，如 zh, en
	FilePath   string         `json:"file_path,omitempty"`   // 关联文件路径
	FileSize   int64          `json:"file_size,omitempty"`   // 文件大小（字节）
	MimeType   string         `json:"mime_type,omitempty"`   // MIME 类型，如 image/png
	AgentID    string         `json:"agent_id,omitempty"`    // 来源 Agent 标识
	SessionID  string         `json:"session_id,omitempty"`  // 会话标识
	Extra      map[string]any `json:"extra,omitempty"`       // 扩展字段，自定义键值对
}

// QueryRequest 语义检索请求
// 用户通过自然语言查询记忆
type QueryRequest struct {
	Query string   `json:"query"`                      // 查询语句，自然语言描述
	TopK  int      `json:"topk"`                       // 返回结果数量上限
	Scope Scope    `json:"scope,omitempty"`            // 可选：限定作用域范围
	Tags  []string `json:"tags,omitempty"`             // 可选：限定标签范围
}

// QueryResult 单条检索结果
// 包含匹配的记忆和相关性得分
type QueryResult struct {
	Memory *Memory `json:"memory"` // 匹配到的记忆
	Score  float64 `json:"score"`  // 相关性得分，0.0~1.0
}

// QueryResponse 语义检索响应
// 返回检索结果列表
type QueryResponse struct {
	Results []*QueryResult `json:"results"` // 结果列表，按得分降序排列
}

// ListRequest 列表查询请求
// 用于分页列出记忆
type ListRequest struct {
	Scope          Scope    `json:"scope,omitempty"`            // 可选：限定作用域
	Tags           []string `json:"tags,omitempty"`             // 可选：限定标签
	Limit          int      `json:"limit"`                      // 每页数量
	Offset         int      `json:"offset"`                     // 偏移量
	IncludeDeleted bool     `json:"include_deleted"`            // 是否包含已删除的记忆
}

// ListResponse 列表查询响应
// 返回记忆列表和总数
type ListResponse struct {
	Memories []*Memory `json:"memories"` // 记忆列表
	Total    int       `json:"total"`    // 记忆总数
}

// StatsResponse 统计信息响应
// 返回记忆系统的各类统计
type StatsResponse struct {
	Total   int            `json:"total"`    // 记忆总数
	ByType  map[string]int `json:"by_type"`  // 按类型统计
	ByScope map[string]int `json:"by_scope"` // 按作用域统计
	ByMedia map[string]int `json:"by_media"` // 按媒体类型统计
	Deleted int            `json:"deleted"`  // 已删除记忆数
}

// BeforeSave 保存前的预处理
// 自动填充 ID、时间戳、默认值等字段
func (m *Memory) BeforeSave() {
	now := time.Now().Unix()
	if m.ID == "" {
		m.ID = GenerateID() // 空 ID 则自动生成 UUID
	}
	if m.CreatedAt == 0 {
		m.CreatedAt = now // 首次保存设置创建时间
	}
	if m.UpdatedAt == 0 {
		m.UpdatedAt = now // 首次保存设置更新时间
	}
	if m.Confidence == 0 {
		m.Confidence = 1.0 // 默认置信度为 1.0
	}
	if m.MediaType == "" {
		m.MediaType = MediaText // 默认媒体类型为文本
	}
}

// MarshalMetadata 将 Metadata 结构体序列化为 JSON 字符串
// 用于存储到数据库的 TEXT 字段
func (m *Memory) MarshalMetadata() (string, error) {
	if len(m.Metadata.Extra) == 0 && m.Metadata.Source == "" && m.Metadata.Language == "" {
		return "", nil // 空的 Metadata 不序列化
	}
	data, err := json.Marshal(m.Metadata)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// UnmarshalMetadata 从 JSON 字符串反序列化为 Metadata 结构体
// 用于从数据库读取 Metadata
func UnmarshalMetadata(data string) (Metadata, error) {
	if data == "" {
		return Metadata{}, nil
	}
	var m Metadata
	err := json.Unmarshal([]byte(data), &m)
	return m, err
}

// MarshalRelatedIDs 将关联记忆 ID 列表序列化为 JSON 字符串
// 用于存储到数据库的 TEXT 字段
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

// UnmarshalRelatedIDs 从 JSON 字符串反序列化为关联记忆 ID 列表
// 用于从数据库读取 RelatedIDs
func UnmarshalRelatedIDs(data string) ([]string, error) {
	if data == "" {
		return nil, nil
	}
	var ids []string
	err := json.Unmarshal([]byte(data), &ids)
	return ids, err
}

// MarshalTags 将标签列表序列化为 JSON 字符串
// 用于存储到数据库的 TEXT 字段
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

// UnmarshalTags 从 JSON 字符串反序列化为标签列表
// 用于从数据库读取 Tags
func UnmarshalTags(data string) ([]string, error) {
	if data == "" {
		return nil, nil
	}
	var tags []string
	err := json.Unmarshal([]byte(data), &tags)
	return tags, err
}
