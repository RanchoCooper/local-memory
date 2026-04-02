# LocalMemory 技术规格文档

> 版本：v0.5.0
> 状态：规划中

---

## 1. 项目概述

**项目名称**：LocalMemory

**项目描述**：为 AI Agent 提供本地优先的持久化、可检索、可进化的长期记忆系统。

**核心能力**：
- 本地运行（隐私优先）
- 通用接入（适配任意 LLM / AI Agent）
- 低延迟（毫秒级查询）
- 可扩展（支持多 Agent）
- 多模态记忆（文本、图像；语音/视频预留）
- 记忆关联（图谱式组织）

**已支持 AI Agent**：
- Claude Code

**非目标（ MVP 阶段）**：
- 语音、视频处理
- 分布式部署
- 多用户权限系统
- 云端同步

---

## 2. 系统架构

### 2.1 分层结构

```
┌──────────────────────────────────────────────────────────────┐
│                         LocalMemory                            │
├──────────────────────────────────────────────────────────────┤
│  接口层                                                          │
│  ├── CLI       (命令行工具)                                      │
│  ├── HTTP API  (REST API)                                      │
│  └── MCP Server (Claude Code 集成)                            │
├──────────────────────────────────────────────────────────────┤
│  核心层                                                          │
│  └── core/     (store, recall, evolve, decay, forget)         │
├──────────────────────────────────────────────────────────────┤
│  支撑层                                                          │
│  ├── storage/  (sqlite, qdrant/usearch)                       │
│  ├── ai/       (embedding, extractor)                        │
│  └── bridge/   (unix socket / http)                          │
└──────────────────────────────────────────────────────────────┘
```

### 2.2 项目目录结构

```
local-memory/
├── cmd/
│   ├── cli/                # CLI 主程序
│   │   ├── main.go
│   │   └── commands/       # 命令实现
│   │       ├── save.go
│   │       ├── query.go
│   │       ├── list.go
│   │       └── forget.go
│   └── server/             # HTTP + MCP 服务入口
│       └── main.go
│
├── core/                   # 核心模块（无外部依赖）
│   ├── memory.go          # Memory 数据结构
│   ├── store.go           # 存储操作
│   ├── recall.go          # 检索操作
│   ├── evolve.go          # 进化操作
│   ├── decay.go           # 衰减操作
│   ├── forget.go          # 遗忘操作
│   ├── ranker.go          # 排序算法
│   └── link.go            # 关联操作
│
├── storage/                # 存储层
│   ├── sqlite.go          # SQLite 适配器
│   ├── vector/
│   │   ├── interface.go   # 向量存储接口
│   │   ├── qdrant.go      # Qdrant 适配器
│   │   └── usearch.go     # USearch 适配器（备选）
│   └── media.go           # 媒体存储（预留）
│
├── bridge/                 # 跨语言通信
│   ├── pybridge.go        # Python 服务封装
│   └── http.go            # HTTP 客户端
│
├── server/                 # HTTP 服务
│   ├── router.go          # 路由定义
│   ├── handlers.go        # 请求处理器
│   └── middleware.go      # 中间件
│
├── agent/                  # Agent / MCP 集成
│   ├── mcp/               # MCP Server 实现
│   │   ├── server.go      # MCP Server 主程序
│   │   ├── tools.go       # 工具定义
│   │   ├── resources.go   # 资源定义
│   │   └── handler.go     # 请求处理
│   └── sdk.go             # Agent SDK 接口
│
├── config/                 # 配置管理
│   └── config.go
│
├── python/                  # Python AI 模块
│   ├── ai/
│   │   ├── embedding.py   # 向量嵌入
│   │   └── extractor.py  # 信息提取
│   ├── server.py         # FastAPI 服务
│   └── requirements.txt
│
├── data/                    # 数据目录
│   ├── localmemory.db
│   └── qdrant/
│
├── config.json              # 配置文件
├── go.mod
├── go.sum
├── Dockerfile
├── docker-compose.yaml
└── Makefile
```

---

## 3. 数据模型

### 3.1 Memory 结构

```go
type Memory struct {
    ID         string       `json:"id"`
    Type       MemoryType   `json:"type"`       // preference | fact | event
    Scope      Scope        `json:"scope"`      // global | session | agent
    MediaType  MediaType    `json:"media_type"` // text | image | audio | video
    Key        string       `json:"key"`
    Value      string       `json:"value"`       // 文本内容或媒体路径
    Confidence float64      `json:"confidence"`  // 0.0 ~ 1.0
    RelatedIDs []string     `json:"related_ids"` // 关联记忆 IDs
    Tags       []string     `json:"tags"`        // 标签，便于分类检索
    Metadata   Metadata     `json:"metadata"`    // 扩展元数据
    Deleted    bool         `json:"deleted"`     // 软删除标记
    DeletedAt  int64        `json:"deleted_at"`  // 软删除时间
    Embedding  []float32    `json:"-"`          // 向量，不持久化到 SQLite
    CreatedAt  int64        `json:"created_at"` // Unix timestamp
    UpdatedAt  int64        `json:"updated_at"`
}

type MemoryType string

const (
    TypePreference MemoryType = "preference"
    TypeFact       MemoryType = "fact"
    TypeEvent      MemoryType = "event"
    TypeSkill      MemoryType = "skill"      // 技能/能力
    TypeGoal       MemoryType = "goal"      // 目标
    TypeRelationship MemoryType = "relationship" // 关系
)

type Scope string

const (
    ScopeGlobal  Scope = "global"  // 全局共享
    ScopeSession Scope = "session" // 会话级
    ScopeAgent   Scope = "agent"  // Agent 私有
)

type MediaType string

const (
    MediaText   MediaType = "text"   // 文本（默认）
    MediaImage  MediaType = "image"  // 图像（MVP 支持）
    MediaAudio  MediaType = "audio"  // 语音（预留）
    MediaVideo  MediaType = "video"  // 视频（预留）
)

type Metadata struct {
    Source       string            `json:"source,omitempty"`       // 来源：claude_code, user_input, api
    Language     string            `json:"language,omitempty"`    // 语言
    FilePath     string            `json:"file_path,omitempty"`   // 关联文件路径
    FileSize     int64             `json:"file_size,omitempty"`  // 文件大小
    MimeType     string            `json:"mime_type,omitempty"`   // MIME 类型
    AgentID      string            `json:"agent_id,omitempty"`    // Agent 标识
    SessionID    string            `json:"session_id,omitempty"`  // 会话标识
    Extra        map[string]any    `json:"extra,omitempty"`       // 扩展字段
}
```

### 3.2 数据库 Schema

```sql
CREATE TABLE memories (
    id          TEXT PRIMARY KEY,
    type        TEXT NOT NULL,
    scope       TEXT NOT NULL,
    media_type  TEXT DEFAULT 'text',
    key         TEXT NOT NULL,
    value       TEXT NOT NULL,
    confidence  REAL DEFAULT 1.0,
    related_ids TEXT,                -- JSON 数组: ["id1", "id2", ...]
    tags        TEXT,                -- JSON 数组: ["tag1", "tag2", ...]
    metadata    TEXT,                 -- JSON 对象
    deleted     INTEGER DEFAULT 0,    -- 软删除标记 (0=未删除, 1=已删除)
    deleted_at  INTEGER,             -- 软删除时间
    created_at  INTEGER NOT NULL,
    updated_at  INTEGER NOT NULL
);

CREATE INDEX idx_memories_key ON memories(key);
CREATE INDEX idx_memories_scope ON memories(scope);
CREATE INDEX idx_memories_type ON memories(type);
CREATE INDEX idx_memories_updated ON memories(updated_at);
CREATE INDEX idx_memories_media ON memories(media_type);
CREATE INDEX idx_memories_deleted ON memories(deleted);  -- 软删除查询优化
```

### 3.3 关联存储

记忆关联通过 `related_ids` JSON 数组实现，支持：
- **双向关联**：创建关联时自动建立反向链接
- **层级关系**：支持父子、兄弟、引用等多种关系
- **图遍历**：通过 BFS/DFS 可快速获取关联记忆子图

```go
// 关联操作示例
func (s *Store) LinkMemories(id1, id2 string) error { ... }
func (s *Store) UnlinkMemories(id1, id2 string) error { ... }
func (s *Store) GetRelated(id string, depth int) ([]*Memory, error) { ... }
```

---

## 4. 核心算法

### 4.1 Recall 流程

```
Query Text → Embedding → Vector Search → Ranking → TopK Results
```

### 4.2 Ranking 算法

```go
func CalculateScore(similarity, recency, confidence float64) float64 {
    return similarity*0.7 + recency*0.2 + confidence*0.1
}

func NormalizeRecency(createdAt int64, maxAgeSeconds int64) float64 {
    age := time.Now().Unix() - createdAt
    return math.Exp(-0.1 * float64(age) / float64(maxAgeSeconds))
}
```

**权重分配**：
- 相似度（similarity）：70%
- 时效性（recency）：20%
- 置信度（confidence）：10%

### 4.3 Decay 衰减机制

```go
func CalculateDecay(createdAt int64, lambda float64) float64 {
    delta := time.Now().Unix() - createdAt
    return math.Exp(-lambda * float64(delta))
}
```

**配置参数**：
- `lambda`：衰减系数（默认 0.01）

### 4.4 Evolve 进化机制

当存在相同 `key` 的记忆时：
1. 合并 value（保留历史）
2. 更新 confidence：`new_confidence = min(1.0, old_confidence + 0.1)`
3. 更新时间戳

### 4.5 Forget 软删除机制

`forget` 操作执行**软删除**，标记 `deleted=true` 而非真正从数据库移除：

```go
func (s *Store) Forget(id string) error {
    return s.db.UpdateMemories(id, map[string]any{
        "deleted":    true,
        "deleted_at": time.Now().Unix(),
    })
}
```

**查询时自动过滤**：默认查询不返回已删除的记忆

```go
func (s *Store) Query(...) ([]*Memory, error) {
    // 自动添加 deleted = 0 条件
    query += " AND deleted = 0"
}
```

**恢复支持**：可按 `deleted_at` 恢复意外删除的记忆

---

## 5. 接口设计

### 5.1 CLI 命令

```bash
# 保存记忆
localmemory save "用户喜欢 Go 语言编程"

# 语义查询
localmemory query "用户偏好什么语言" --topk=5 --scope=global

# 列出记忆
localmemory list --scope=global --limit=20

# 删除记忆
localmemory forget <key|id>

# 统计信息
localmemory stats
```

### 5.2 HTTP API

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/memories` | 创建记忆 |
| GET | `/api/v1/memories` | 列出记忆 |
| GET | `/api/v1/memories/:id` | 获取单个记忆 |
| DELETE | `/api/v1/memories/:id` | 删除记忆 |
| POST | `/api/v1/query` | 语义检索 |
| POST | `/api/v1/extract` | 提取记忆（AI） |
| GET | `/api/v1/stats` | 统计信息 |
| GET | `/health` | 健康检查 |

**请求/响应格式**：

```json
// POST /api/v1/memories
// Request
{
    "type": "preference",
    "scope": "global",
    "key": "language",
    "value": "Go",
    "confidence": 0.9
}

// Response
{
    "success": true,
    "data": {
        "id": "uuid",
        "type": "preference",
        "scope": "global",
        "key": "language",
        "value": "Go",
        "confidence": 0.9,
        "created_at": 1710000000,
        "updated_at": 1710000000
    }
}

// POST /api/v1/query
// Request
{
    "query": "用户喜欢什么语言",
    "topk": 5,
    "scope": "global"
}

// Response
{
    "success": true,
    "data": [
        {
            "memory": {...},
            "score": 0.85
        }
    ]
}
```

### 5.3 Claude Code 集成

#### 5.3.1 MCP 协议集成（推荐）

Claude Code 通过 MCP (Model Context Protocol) 接入 LocalMemory：

```json
// Claude Code 配置 ~/.claude/settings.json
{
  "mcpServers": {
    "localmemory": {
      "command": "localmemory",
      "args": ["mcp"],
      "env": {
        "LOCALMEMORY_DB_PATH": "./data/localmemory.db"
      }
    }
  }
}
```

**MCP 传输方式**：
- **stdio**（默认）：本地进程通信，低延迟
- **HTTP + SSE**：远程服务场景

#### 5.3.2 Claude Code 记忆操作

| 操作 | 来源 | 说明 |
|------|------|------|
| `CLAUDE.md` 记忆 | 项目 `CLAUDE.md` | 自动同步关键指令到 global scope |
| 工作会话总结 | 定期自动 | 将重要操作保存为记忆 |
| 用户偏好 | 用户交互 | 记录用户编程风格、项目偏好 |
| 项目知识 | 代码分析 | 记录项目架构、技术栈、代码规范 |

#### 5.3.3 Agent SDK 接口

```go
// agent/sdk.go
type AgentSDK interface {
    // SaveMemory 保存记忆
    SaveMemory(ctx context.Context, memory *Memory) error

    // QueryMemories 语义检索
    QueryMemories(ctx context.Context, req *QueryRequest) (*QueryResponse, error)

    // ListMemories 列出记忆
    ListMemories(ctx context.Context, scope Scope) ([]*Memory, error)

    // GetContext 获取上下文（用于 LLM 上下文注入）
    GetContext(ctx context.Context, query string, limit int) (string, error)

    // LinkMemories 关联记忆
    LinkMemories(ctx context.Context, id1, id2 string) error

    // Forget 删除记忆
    Forget(ctx context.Context, id string) error
}
```

#### 5.3.2 MCP Server 工具定义

| 工具名 | 描述 | 输入 |
|--------|------|------|
| `memory_save` | 保存记忆 | type, key, value, scope, confidence |
| `memory_query` | 语义检索 | query, topk, scope |
| `memory_list` | 列出记忆 | scope, limit |
| `memory_forget` | 删除记忆 | id |
| `memory_get_context` | 获取 LLM 上下文 | query, limit |

**MCP 资源定义**：

| URI | 类型 | 描述 |
|-----|------|------|
| `memory://all` | application/json | 所有记忆 |
| `memory://stats` | application/json | 统计信息 |
| `memory://recent` | application/json | 最近记忆 |
| `memory://preference` | application/json | 用户偏好 |

#### 5.3.3 Claude Code 使用场景

| 场景 | 说明 |
|------|------|
| `CLAUDE.md` 记忆 | 从项目 `CLAUDE.md` 自动同步关键指令 |
| 工作会话总结 | 定期将重要操作保存为记忆 |
| 用户偏好 | 记录用户编程风格、项目偏好 |
| 项目知识 | 记录项目架构、技术栈、代码规范 |

---

## 6. 技术选型

| 模块 | 技术 | 说明 |
|------|------|------|
| CLI | Go + Cobra | 命令行工具 |
| HTTP API | Go + Gin | REST API |
| MCP Server | Go + json-rpc | Claude Code 集成 |
| 元数据存储 | SQLite | 轻量级本地数据库 |
| 向量存储 | Qdrant | 本地向量数据库 |
| Go-Python 通信 | TCP localhost | 跨平台兼容 |
| Embedding | sentence-transformers | Python 本地模型 |
| AI 服务 | Python + FastAPI | AI 处理层 |
| 配置 | JSON | 配置文件 |

---

## 6.1 Go-Python 通信方案

**推荐：TCP localhost**（跨平台兼容）

```
┌─────────────┐      TCP localhost      ┌─────────────┐
│  Go (bridge)│ ←────────────────────→ │ Python (AI) │
└─────────────┘                        └─────────────┘
```

| 方案 | Windows | Linux | Mac | 性能 | 推荐度 |
|------|---------|-------|-----|------|--------|
| TCP localhost | ✅ | ✅ | ✅ | 中 | ⭐⭐⭐⭐⭐ |
| Unix Domain Socket | ❌ | ✅ | ✅ | 高 | ⭐⭐⭐⭐ |
| Named Pipe | ✅ | ✅ | ✅ | 高 | ⭐⭐⭐⭐ |

**配置**：
```json
{
  "bridge": {
    "type": "tcp",
    "tcp_url": "127.0.0.1:8081"
  }
}
```

**未来可选升级**：Linux/Mac 用 Unix Socket，Windows 用 Named Pipe

---

## 6.2 向量数据库

**Qdrant**（主选，高性能、功能完整）

```yaml
# docker-compose.yaml
qdrant:
  image: qdrant/qdrant
  ports:
    - "6333:6333"
  volumes:
    - ./data/qdrant:/qdrant/storage
```

**设计：可切换向量存储接口**

```go
// storage/vector/interface.go
type VectorStore interface {
    Upsert(id string, vector []float32, metadata map[string]any) error
    Search(query []float32, topK int, filter *Filter) ([]Result, error)
    Delete(id string) error
    Close() error
}

// 实现
// - QdrantStore  (主选)
// - USearchStore (轻量备选)
```

**切换配置**：
```json
{
  "vector_db": {
    "type": "qdrant",  // "qdrant" | "usearch"
    "url": "http://127.0.0.1:6333"
  }
}
```

| 方案 | 适用场景 | 部署 |
|------|---------|------|
| Qdrant | 生产环境、数据量大 | Docker |
| USearch | MVP、轻量级、无 Docker | 纯 Go 库 |

---

## 7. 配置格式

```json
{
  "database": {
    "path": "./data/localmemory.db"
  },
  "vector_db": {
    "type": "qdrant",
    "url": "http://127.0.0.1:6333",
    "collection": "memories"
  },
  "bridge": {
    "type": "tcp",
    "tcp_url": "127.0.0.1:8081"
  },
  "ai": {
    "embedding_model": "all-MiniLM-L6-v2"
  },
  "decay": {
    "lambda": 0.01
  },
  "server": {
    "port": 8080
  },
  "cli": {
    "default_topk": 5,
    "default_scope": "global"
  },
  "agent": {
    "id": "localmemory",
    "name": "LocalMemory"
  }
}
```

---

## 8. 性能目标

| 指标 | 目标 |
|------|------|
| 查询延迟 | < 50ms |
| 存储延迟 | < 100ms |
| 吞吐量 | 1000 QPS（单机） |

---

## 9. 实现阶段

### 阶段 1：项目骨架（Foundation）
**目标**：建立项目结构、配置系统、核心数据结构

**交付物**：
- Go module 初始化
- 目录结构创建
- `config.json` 和配置加载模块
- `Memory` 数据结构定义

**步骤**：
1. 初始化 Go module
2. 创建目录结构
3. 实现配置加载
4. 定义 Memory 结构体

**依赖**：无

---

### 阶段 2：存储层（Storage Layer）
**目标**：实现 SQLite 元数据存储和向量存储接口

**交付物**：
- SQLite 适配器（CRUD + 索引）
- 向量存储接口抽象
- Qdrant 适配器
- USearch 适配器（备选）

**步骤**：
1. 实现 SQLite 适配器
2. 定义向量存储接口
3. 实现 Qdrant 适配器
4. 实现 USearch 适配器

**依赖**：Phase 1

---

### 阶段 3：核心业务逻辑（Core Logic）
**目标**：实现记忆的存储、检索、排序、衰减、进化、遗忘、关联

**交付物**：
- Store 模块（存储）
- Recall 模块（检索）
- Ranker 模块（排序）
- Decay 模块（衰减）
- Evolve 模块（进化）
- Forget 模块（软删除）
- Link 模块（关联）

**关键算法**：
```
Ranking Score = similarity×0.7 + recency×0.2 + confidence×0.1
Decay = e^(-λ × Δt)
```

**依赖**：Phase 1, Phase 2

---

### 阶段 4：Go-Python 桥接层（Bridge Layer）
**目标**：实现 Go 与 Python AI 模块的 TCP 通信

**交付物**：
- TCP 客户端
- JSON-RPC 协议定义
- Python FastAPI 服务
- Embedding/Extractor 模块

**步骤**：
1. 定义通信协议
2. 实现 Go TCP 客户端
3. 实现 Python FastAPI 服务
4. 集成 embedding/extractor

**依赖**：Phase 3

---

### 阶段 5：CLI 工具
**目标**：提供命令行界面

**交付物**：
- CLI 主程序
- save 命令
- query 命令
- list 命令
- forget 命令
- stats 命令

**依赖**：Phase 3（query 需要 Phase 4）

---

### 阶段 6：HTTP API
**目标**：提供 REST API 接口

**交付物**：
- HTTP 服务
- 路由定义
- 处理器
- 中间件

**API 端点**：CRUD + query + extract + stats + health

**依赖**：Phase 3, Phase 4

---

### 阶段 7：MCP Server
**目标**：实现 Claude Code MCP 集成

**交付物**：
- MCP Server 主程序
- 工具定义（5个工具）
- 资源定义（4个资源）
- stdio 传输

**依赖**：Phase 3, Phase 4

---

### 阶段 8：测试与部署
**目标**：确保代码质量和可维护性

**交付物**：
- 单元测试（覆盖率 > 80%）
- 集成测试
- Makefile
- Dockerfile
- docker-compose.yaml

**依赖**：Phase 5, Phase 6, Phase 7

---

### MVP vs Production 划分

| 阶段 | MVP | Production |
|------|-----|------------|
| Phase 1-3 | ✅ | ✅ |
| Phase 4 (Mock) | ✅ | ✅ (真实) |
| Phase 5 | ✅ (基本命令) | ✅ |
| Phase 6 | ❌ | ✅ |
| Phase 7 | ❌ | ✅ |
| Phase 8 | ❌ | ✅ |

**MVP 验收标准**：
- `localmemory save` 成功保存记忆
- `localmemory list` 返回记忆列表
- `localmemory forget <id>` 软删除成功
- 同 key 记忆自动合并 (Evolve)

**Production 验收标准**：
- 语义检索返回相关结果
- HTTP API CRUD 正常
- MCP Server stdio 传输正常
- Memory Decay 按配置衰减
- 向量存储切换正常
- 单元测试覆盖率 > 80%

---

## 10. 风险与应对

| 风险 | 概率 | 影响 | 应对策略 |
|------|------|------|----------|
| Qdrant Docker 依赖 | 高 | 中 | USearch 作为 MVP 备选 |
| Python 服务启动失败 | 中 | 高 | 提供一键启动脚本 |
| embedding 模型加载慢 | 中 | 低 | 模型缓存 + 预热 |
| 记忆污染（低质量） | 低 | 高 | confidence 阈值过滤 |
| 数据膨胀 | 中 | 中 | 定期 decay + 清理任务 |
| MCP 协议兼容性 | 低 | 高 | 参考官方实现 + 严格测试 |
| Go-Python 通信延迟 | 中 | 低 | Unix Socket（Linux/Mac）/ Named Pipe（Windows） |

---

---

## 11. 未来演进

| 版本 | 规划 |
|------|------|
| V2 | Web UI，可视化记忆图谱 |
| V3 | 多 Agent 共享记忆，分布式同步 |
| V4 | Memory Marketplace |

---

## 12. 验收标准

- [ ] `localmemory save "用户喜欢 Go"` 成功保存记忆
- [ ] `localmemory query "用户偏好"` 返回相关记忆
- [ ] `localmemory forget <id>` 成功删除记忆
- [ ] HTTP API CRUD 正常
- [ ] 语义检索正常返回结果
- [ ] Memory Decay 按配置衰减
- [ ] Memory Evolve 同 key 记忆合并
- [ ] Memory Link 关联记忆正常
- [ ] 多模态 MediaType 字段正确处理（text + image MVP 支持）
- [ ] MCP Server stdio 传输正常
- [ ] Claude Code 通过 MCP 接入正常
- [ ] TCP localhost 通信正常
- [ ] Qdrant 向量存储正常
- [ ] 向量存储支持切换（Qdrant ↔ USearch）
- [ ] 单元测试覆盖率 > 80%
