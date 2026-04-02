# LocalMemory Technical Specification

> Version: v0.1.0
> Status: In Planning

---

## 1. Project Overview

**Project Name**: LocalMemory

**Project Description**: A local-first, persistent, searchable, and evolvable long-term memory system for AI Agents.

**Core Capabilities**:
- Local-first (privacy priority)
- Universal integration (adaptable to any LLM / AI Agent)
- Low latency (millisecond-level queries)
- Scalable (supports multiple Agents)
- Multi-modal memory (text, image; audio/video reserved)
- Memory associations (graph-based organization)

**Supported AI Agents**:
- Claude Code

**Non-goals (MVP Stage)**:
- Audio/Video processing
- Distributed deployment
- Multi-user permission system
- Cloud sync

---

## 2. System Architecture

### 2.1 Layered Structure

```
┌──────────────────────────────────────────────────────────────┐
│                         LocalMemory                            │
├──────────────────────────────────────────────────────────────┤
│  Interface Layer                                              │
│  ├── CLI       (Command Line Tool)                           │
│  ├── HTTP API  (REST API)                                     │
│  └── MCP Server (Claude Code Integration)                    │
├──────────────────────────────────────────────────────────────┤
│  Core Layer                                                  │
│  └── core/     (store, recall, evolve, decay, forget)        │
├──────────────────────────────────────────────────────────────┤
│  Supporting Layer                                            │
│  ├── storage/  (sqlite, qdrant/usearch)                      │
│  ├── ai/       (embedding, extractor)                        │
│  └── bridge/   (unix socket / http)                          │
└──────────────────────────────────────────────────────────────┘
```

### 2.2 Project Directory Structure

```
local-memory/
├── cmd/
│   ├── cli/                # CLI main program
│   │   ├── main.go
│   │   └── commands/       # Command implementations
│   │       ├── save.go
│   │       ├── query.go
│   │       ├── list.go
│   │       └── forget.go
│   └── server/             # HTTP + MCP service entry
│       └── main.go
│
├── core/                   # Core modules (no external dependencies)
│   ├── memory.go          # Memory data structure
│   ├── store.go           # Storage operations
│   ├── recall.go          # Retrieval operations
│   ├── evolve.go          # Evolution operations
│   ├── decay.go           # Decay operations
│   ├── forget.go          # Forget operations
│   ├── ranker.go          # Ranking algorithm
│   └── link.go            # Association operations
│
├── storage/                # Storage layer
│   ├── sqlite.go          # SQLite adapter
│   ├── vector/
│   │   ├── interface.go   # Vector store interface
│   │   ├── qdrant.go      # Qdrant adapter
│   │   └── usearch.go     # USearch adapter (alternative)
│   └── media.go           # Media storage (reserved)
│
├── bridge/                 # Cross-language communication
│   ├── pybridge.go        # Python service wrapper
│   └── http.go            # HTTP client
│
├── server/                 # HTTP service
│   ├── router.go          # Route definitions
│   ├── handlers.go        # Request handlers
│   └── middleware.go      # Middleware
│
├── agent/                  # Agent / MCP Integration
│   ├── mcp/               # MCP Server implementation
│   │   ├── server.go      # MCP Server main program
│   │   ├── tools.go       # Tool definitions
│   │   ├── resources.go   # Resource definitions
│   │   └── handler.go     # Request handling
│   └── sdk.go             # Agent SDK interface
│
├── config/                 # Configuration management
│   └── config.go
│
├── python/                  # Python AI module
│   ├── ai/
│   │   ├── embedding.py   # Vector embedding
│   │   └── extractor.py  # Information extraction
│   ├── server.py         # FastAPI service
│   └── requirements.txt
│
├── data/                    # Data directory
│   ├── localmemory.db
│   └── qdrant/
│
├── config.json              # Configuration file
├── go.mod
├── go.sum
├── Dockerfile
├── docker-compose.yaml
└── Makefile
```

---

## 3. Data Model

### 3.1 Memory Structure

```go
type Memory struct {
    ID         string       `json:"id"`
    Type       MemoryType   `json:"type"`       // preference | fact | event
    Scope      Scope        `json:"scope"`      // global | session | agent
    MediaType  MediaType    `json:"media_type"` // text | image | audio | video
    Key        string       `json:"key"`
    Value      string       `json:"value"`       // Text content or media path
    Confidence float64      `json:"confidence"`  // 0.0 ~ 1.0
    RelatedIDs []string     `json:"related_ids"` // Related memory IDs
    Tags       []string     `json:"tags"`        // Tags for categorization and retrieval
    Metadata   Metadata     `json:"metadata"`    // Extended metadata
    Deleted    bool         `json:"deleted"`     // Soft delete flag
    DeletedAt  int64        `json:"deleted_at"`  // Soft delete timestamp
    Embedding  []float32    `json:"-"`          // Vector, not persisted to SQLite
    CreatedAt  int64        `json:"created_at"` // Unix timestamp
    UpdatedAt  int64        `json:"updated_at"`
}

type MemoryType string

const (
    TypePreference MemoryType = "preference"
    TypeFact       MemoryType = "fact"
    TypeEvent      MemoryType = "event"
    TypeSkill      MemoryType = "skill"
    TypeGoal       MemoryType = "goal"
    TypeRelationship MemoryType = "relationship"
)

type Scope string

const (
    ScopeGlobal  Scope = "global"  // Globally shared
    ScopeSession Scope = "session" // Session level
    ScopeAgent   Scope = "agent"  // Agent private
)

type MediaType string

const (
    MediaText   MediaType = "text"   // Text (default)
    MediaImage  MediaType = "image"  // Image (MVP supported)
    MediaAudio  MediaType = "audio"  // Audio (reserved)
    MediaVideo  MediaType = "video"  // Video (reserved)
)

type Metadata struct {
    Source       string            `json:"source,omitempty"`       // Source: claude_code, user_input, api
    Language     string            `json:"language,omitempty"`    // Language
    FilePath     string            `json:"file_path,omitempty"`   // Associated file path
    FileSize     int64             `json:"file_size,omitempty"`   // File size
    MimeType     string            `json:"mime_type,omitempty"`   // MIME type
    AgentID      string            `json:"agent_id,omitempty"`    // Agent identifier
    SessionID    string            `json:"session_id,omitempty"`  // Session identifier
    Extra        map[string]any    `json:"extra,omitempty"`       // Extension fields
}
```

### 3.2 Database Schema

```sql
CREATE TABLE memories (
    id          TEXT PRIMARY KEY,
    type        TEXT NOT NULL,
    scope       TEXT NOT NULL,
    media_type  TEXT DEFAULT 'text',
    key         TEXT NOT NULL,
    value       TEXT NOT NULL,
    confidence  REAL DEFAULT 1.0,
    related_ids TEXT,                -- JSON array: ["id1", "id2", ...]
    tags        TEXT,                -- JSON array: ["tag1", "tag2", ...]
    metadata    TEXT,                 -- JSON object
    deleted     INTEGER DEFAULT 0,    -- Soft delete flag (0=not deleted, 1=deleted)
    deleted_at  INTEGER,             -- Soft delete timestamp
    created_at  INTEGER NOT NULL,
    updated_at  INTEGER NOT NULL
);

CREATE INDEX idx_memories_key ON memories(key);
CREATE INDEX idx_memories_scope ON memories(scope);
CREATE INDEX idx_memories_type ON memories(type);
CREATE INDEX idx_memories_updated ON memories(updated_at);
CREATE INDEX idx_memories_media ON memories(media_type);
CREATE INDEX idx_memories_deleted ON memories(deleted);  -- Soft delete query optimization
```

### 3.3 Association Storage

Memory associations are implemented via `related_ids` JSON array, supporting:
- **Bidirectional associations**: Automatically creates reverse links when creating associations
- **Hierarchical relationships**: Supports parent-child, sibling, reference and other relationships
- **Graph traversal**: Can quickly get associated memory subgraphs via BFS/DFS

```go
// Association operation examples
func (s *Store) LinkMemories(id1, id2 string) error { ... }
func (s *Store) UnlinkMemories(id1, id2 string) error { ... }
func (s *Store) GetRelated(id string, depth int) ([]*Memory, error) { ... }
```

---

## 4. Core Algorithms

### 4.1 Recall Flow

```
Query Text → Embedding → Vector Search → Ranking → TopK Results
```

### 4.2 Ranking Algorithm

```go
func CalculateScore(similarity, recency, confidence float64) float64 {
    return similarity*0.7 + recency*0.2 + confidence*0.1
}

func NormalizeRecency(createdAt int64, maxAgeSeconds int64) float64 {
    age := time.Now().Unix() - createdAt
    return math.Exp(-0.1 * float64(age) / float64(maxAgeSeconds))
}
```

**Weight Distribution**:
- Similarity: 70%
- Recency: 20%
- Confidence: 10%

### 4.3 Decay Mechanism

```go
func CalculateDecay(createdAt int64, lambda float64) float64 {
    delta := time.Now().Unix() - createdAt
    return math.Exp(-lambda * float64(delta))
}
```

**Configuration Parameters**:
- `lambda`: Decay coefficient (default 0.01)

### 4.4 Evolve Mechanism

When a memory with the same `key` exists:
1. Merge value (preserve history)
2. Update confidence: `new_confidence = min(1.0, old_confidence + 0.1)`
3. Update timestamp

### 4.5 Forget Soft Delete Mechanism

`forget` operation performs **soft delete**, marking `deleted=true` instead of actually removing from database:

```go
func (s *Store) Forget(id string) error {
    return s.db.UpdateMemories(id, map[string]any{
        "deleted":    true,
        "deleted_at": time.Now().Unix(),
    })
}
```

**Automatic filtering on query**: Deleted memories are not returned by default queries

```go
func (s *Store) Query(...) ([]*Memory, error) {
    // Automatically adds deleted = 0 condition
    query += " AND deleted = 0"
}
```

**Recovery support**: Accidentally deleted memories can be recovered by `deleted_at`

---

## 5. Interface Design

### 5.1 CLI Commands

```bash
# Save memory
localmemory save "User prefers Go language programming"

# Semantic query
localmemory query "What language does user prefer" --topk=5 --scope=global

# List memories
localmemory list --scope=global --limit=20

# Delete memory
localmemory forget <key|id>

# Statistics
localmemory stats
```

### 5.2 HTTP API

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/memories` | Create memory |
| GET | `/api/v1/memories` | List memories |
| GET | `/api/v1/memories/:id` | Get single memory |
| DELETE | `/api/v1/memories/:id` | Delete memory |
| POST | `/api/v1/query` | Semantic search |
| POST | `/api/v1/extract` | Extract memory (AI) |
| GET | `/api/v1/stats` | Statistics |
| GET | `/health` | Health check |

**Request/Response Format**:

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
    "query": "What language does user prefer",
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

### 5.3 Claude Code Integration

#### 5.3.1 MCP Protocol Integration (Recommended)

Claude Code connects to LocalMemory via MCP (Model Context Protocol):

```json
// Claude Code configuration ~/.claude/settings.json
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

**MCP Transport**:
- **stdio** (default): Local process communication, low latency
- **HTTP + SSE**: Remote service scenarios

#### 5.3.2 Claude Code Memory Operations

| Operation | Source | Description |
|------|------|------|
| `CLAUDE.md` memory | Project `CLAUDE.md` | Auto-sync key instructions to global scope |
| Work session summary | Periodic auto | Save important operations as memories |
| User preference | User interaction | Record user's programming style, project preferences |
| Project knowledge | Code analysis | Record project architecture, tech stack, code standards |

#### 5.3.3 Agent SDK Interface

```go
// agent/sdk.go
type AgentSDK interface {
    // SaveMemory saves a memory
    SaveMemory(ctx context.Context, memory *Memory) error

    // QueryMemories semantic search
    QueryMemories(ctx context.Context, req *QueryRequest) (*QueryResponse, error)

    // ListMemories lists memories
    ListMemories(ctx context.Context, scope Scope) ([]*Memory, error)

    // GetContext gets context (for LLM context injection)
    GetContext(ctx context.Context, query string, limit int) (string, error)

    // LinkMemories associate memories
    LinkMemories(ctx context.Context, id1, id2 string) error

    // Forget deletes a memory
    Forget(ctx context.Context, id string) error
}
```

#### 5.3.2 MCP Server Tool Definitions

| Tool Name | Description | Input |
|--------|------|------|
| `memory_save` | Save memory | type, key, value, scope, confidence |
| `memory_query` | Semantic search | query, topk, scope |
| `memory_list` | List memories | scope, limit |
| `memory_forget` | Delete memory | id |
| `memory_get_context` | Get LLM context | query, limit |

**MCP Resource Definitions**:

| URI | Type | Description |
|-----|------|------|
| `memory://all` | application/json | All memories |
| `memory://stats` | application/json | Statistics |
| `memory://recent` | application/json | Recent memories |
| `memory://preference` | application/json | User preferences |

#### 5.3.3 Claude Code Use Cases

| Scenario | Description |
|------|------|
| `CLAUDE.md` memory | Auto-sync key instructions from project `CLAUDE.md` |
| Work session summary | Periodically save important operations as memories |
| User preference | Record user's programming style, project preferences |
| Project knowledge | Record project architecture, tech stack, code standards |

---

## 6. Technology Selection

| Module | Technology | Description |
|------|------|------|
| CLI | Go + Cobra | Command line tool |
| HTTP API | Go + Gin | REST API |
| MCP Server | Go + json-rpc | Claude Code integration |
| Metadata storage | SQLite | Lightweight local database |
| Vector storage | Qdrant | Local vector database |
| Go-Python communication | TCP localhost | Cross-platform compatibility |
| Embedding | sentence-transformers | Python local model |
| AI service | Python + FastAPI | AI processing layer |
| Configuration | JSON | Configuration file |

---

## 6.1 Go-Python Communication

**Recommended: TCP localhost** (cross-platform compatible)

```
┌─────────────┐      TCP localhost      ┌─────────────┐
│  Go (bridge)│ ←────────────────────→ │ Python (AI) │
└─────────────┘                        └─────────────┘
```

| Option | Windows | Linux | Mac | Performance | Recommendation |
|------|---------|-------|-----|------|--------|
| TCP localhost | ✅ | ✅ | ✅ | Medium | ⭐⭐⭐⭐⭐ |
| Unix Domain Socket | ❌ | ✅ | ✅ | High | ⭐⭐⭐⭐ |
| Named Pipe | ✅ | ✅ | ✅ | High | ⭐⭐⭐⭐ |

**Configuration**:
```json
{
  "bridge": {
    "type": "tcp",
    "tcp_url": "127.0.0.1:8081"
  }
}
```

**Future optional upgrade**: Unix Socket for Linux/Mac, Named Pipe for Windows

---

## 6.2 Vector Database

**Qdrant** (primary, high performance, full-featured)

```yaml
# docker-compose.yaml
qdrant:
  image: qdrant/qdrant
  ports:
    - "6333:6333"
  volumes:
    - ./data/qdrant:/qdrant/storage
```

**Design: Swappable vector storage interface**

```go
// storage/vector/interface.go
type VectorStore interface {
    Upsert(id string, vector []float32, metadata map[string]any) error
    Search(query []float32, topK int, filter *Filter) ([]Result, error)
    Delete(id string) error
    Close() error
}

// Implementations
// - QdrantStore  (primary)
// - USearchStore (lightweight alternative)
```

**Switch configuration**:
```json
{
  "vector_db": {
    "type": "qdrant",  // "qdrant" | "usearch"
    "url": "http://127.0.0.1:6333"
  }
}
```

| Option | Use Case | Deployment |
|------|---------|------|
| Qdrant | Production, large data | Docker |
| USearch | MVP, lightweight, no Docker | Pure Go library |

---

## 7. Configuration Format

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

## 8. Performance Goals

| Metric | Goal |
|------|------|
| Query latency | < 50ms |
| Storage latency | < 100ms |
| Throughput | 1000 QPS (single machine) |

---

## 9. Implementation Phases

### Phase 1: Project Foundation
**Goal**: Establish project structure, configuration system, core data structures

**Deliverables**:
- Go module initialization
- Directory structure creation
- `config.json` and configuration loading module
- `Memory` data structure definition

**Steps**:
1. Initialize Go module
2. Create directory structure
3. Implement configuration loading
4. Define Memory struct

**Dependencies**: None

---

### Phase 2: Storage Layer
**Goal**: Implement SQLite metadata storage and vector store interface

**Deliverables**:
- SQLite adapter (CRUD + indexes)
- Vector store interface abstraction
- Qdrant adapter
- USearch adapter (alternative)

**Steps**:
1. Implement SQLite adapter
2. Define vector store interface
3. Implement Qdrant adapter
4. Implement USearch adapter

**Dependencies**: Phase 1

---

### Phase 3: Core Business Logic
**Goal**: Implement memory storage, retrieval, ranking, decay, evolution, forgetting, associations

**Deliverables**:
- Store module (storage)
- Recall module (retrieval)
- Ranker module (ranking)
- Decay module (decay)
- Evolve module (evolution)
- Forget module (soft delete)
- Link module (associations)

**Key Algorithms**:
```
Ranking Score = similarity×0.7 + recency×0.2 + confidence×0.1
Decay = e^(-λ × Δt)
```

**Dependencies**: Phase 1, Phase 2

---

### Phase 4: Go-Python Bridge Layer
**Goal**: Implement TCP communication between Go and Python AI modules

**Deliverables**:
- TCP client
- JSON-RPC protocol definition
- Python FastAPI service
- Embedding/Extractor modules

**Steps**:
1. Define communication protocol
2. Implement Go TCP client
3. Implement Python FastAPI service
4. Integrate embedding/extractor

**Dependencies**: Phase 3

---

### Phase 5: CLI Tool
**Goal**: Provide command line interface

**Deliverables**:
- CLI main program
- save command
- query command
- list command
- forget command
- stats command

**Dependencies**: Phase 3 (query requires Phase 4)

---

### Phase 6: HTTP API
**Goal**: Provide REST API interface

**Deliverables**:
- HTTP service
- Route definitions
- Handlers
- Middleware

**API Endpoints**: CRUD + query + extract + stats + health

**Dependencies**: Phase 3, Phase 4

---

### Phase 7: MCP Server
**Goal**: Implement Claude Code MCP integration

**Deliverables**:
- MCP Server main program
- Tool definitions (5 tools)
- Resource definitions (4 resources)
- stdio transport

**Dependencies**: Phase 3, Phase 4

---

### Phase 8: Testing and Deployment
**Goal**: Ensure code quality and maintainability

**Deliverables**:
- Unit tests (> 80% coverage)
- Integration tests
- Makefile
- Dockerfile
- docker-compose.yaml

**Dependencies**: Phase 5, Phase 6, Phase 7

---

### MVP vs Production Division

| Phase | MVP | Production |
|------|-----|------------|
| Phase 1-3 | ✅ | ✅ |
| Phase 4 (Mock) | ✅ | ✅ (real) |
| Phase 5 | ✅ (basic commands) | ✅ |
| Phase 6 | ❌ | ✅ |
| Phase 7 | ❌ | ✅ |
| Phase 8 | ❌ | ✅ |

**MVP Acceptance Criteria**:
- `localmemory save` successfully saves memory
- `localmemory list` returns memory list
- `localmemory forget <id>` soft delete succeeds
- Same key memories auto-merge (Evolve)

**Production Acceptance Criteria**:
- Semantic search returns relevant results
- HTTP API CRUD works normally
- MCP Server stdio transport works normally
- Memory Decay decays according to configuration
- Vector store switching works normally

---

## 10. Risks and Mitigation

| Risk | Probability | Impact | Mitigation |
|------|------|------|----------|
| Qdrant Docker dependency | High | Medium | USearch as MVP alternative |
| Python service startup failure | Medium | High | Provide one-click startup script |
| Slow embedding model loading | Medium | Low | Model cache + warmup |
| Memory pollution (low quality) | Low | High | confidence threshold filtering |
| Data bloat | Medium | Medium | Periodic decay + cleanup tasks |
| MCP protocol compatibility | Low | High | Reference official implementation + strict testing |
| Go-Python communication latency | Medium | Low | Unix Socket (Linux/Mac) / Named Pipe (Windows) |

---

---

## 11. Future Evolution

| Version | Plan |
|------|------|
| V2 | Web UI, visual memory graph |
| V3 | Multi-Agent shared memory, distributed sync |
| V4 | Memory Marketplace |

---

## 12. Acceptance Criteria

- [ ] `localmemory save "User prefers Go"` successfully saves memory
- [ ] `localmemory query "User preference"` returns related memories
- [ ] `localmemory forget <id>` successfully deletes memory
- [ ] HTTP API CRUD works normally
- [ ] Semantic search returns relevant results
- [ ] Memory Decay decays according to configuration
- [ ] Memory Evolve merges same key memories
- [ ] Memory Link associates memories normally
- [ ] Multi-modal MediaType field correctly handled (text + image MVP supported)
- [ ] MCP Server stdio transport works normally
- [ ] Claude Code connects via MCP normally
- [ ] TCP localhost communication works normally
- [ ] Qdrant vector storage works normally
- [ ] Vector store switching works (Qdrant ↔ USearch)
- [ ] Unit test coverage > 80%
