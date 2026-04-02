# Local Memory

**A local-first memory layer for AI agents**

LocalMemory provides AI agents with persistent, searchable, and evolvable long-term memory capabilities. It enables AI to remember user preferences, project context, and learned knowledge across sessions.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

---

## Table of Contents

- [Why LocalMemory?](#why-localmemory)
- [Key Features](#key-features)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Development](#development)
- [Architecture Deep Dive](#architecture-deep-dive)

---

## Why LocalMemory?

Current AI systems have a fundamental flaw: **no persistent memory**. Every conversation starts from scratch. LocalMemory solves this by providing:

- **Privacy-first**: All data stays on your machine
- **Cross-session continuity**: AI remembers previous interactions
- **Structured memory**: Memories are typed, tagged, and associatively linked
- **Automatic evolution**: New information merges intelligently with existing knowledge
- **Time-aware**: Important memories persist; stale ones decay naturally

### The Problem

```
Traditional AI:
┌─────────────────────────────────────┐
│  Session 1: "I prefer Go language"  │
└─────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────┐
│  Session 2: "What language?"        │ ← AI forgot!
└─────────────────────────────────────┘
```

### With LocalMemory:

```
┌─────────────────────────────────────┐
│  Session 1: Save "prefers Go"      │
└─────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────┐
│  Memory Store (persistent)          │
│  └── preference: "Go language"      │
└─────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────┐
│  Session 2: Query "language"        │ ← AI remembers!
└─────────────────────────────────────┘
```

---

## Key Features

| Feature | Description |
|---------|-------------|
| **Memory Types** | preference, fact, event, skill, goal, relationship |
| **Scopes** | global (shared), session (temporary), agent (private) |
| **Semantic Search** | Vector-based similarity search with ranking |
| **Memory Evolution** | Automatic merging of related memories |
| **Time Decay** | Configurable decay for memory importance |
| **Soft Delete** | Recoverable deletion with restore support |
| **Memory Linking** | Graph-based associations between memories |
| **MCP Integration** | Native Claude Code integration via MCP protocol |

---

## Architecture

### System Overview

```
┌──────────────────────────────────────────────────────────────────┐
│                         LocalMemory                                │
├──────────────────────────────────────────────────────────────────┤
│  Interface Layer                                                  │
│  ┌────────────┐  ┌─────────────┐  ┌─────────────────────────┐   │
│  │    CLI     │  │  HTTP API   │  │   MCP Server (stdio)     │   │
│  └────────────┘  └─────────────┘  └─────────────────────────┘   │
├──────────────────────────────────────────────────────────────────┤
│  Core Layer                                                       │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌────────────┐ │
│  │  Store  │ │  Recall │ │  Evolve │ │  Decay  │ │   Forget   │ │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘ └────────────┘ │
├──────────────────────────────────────────────────────────────────┤
│  Supporting Layer                                                 │
│  ┌────────────────────┐  ┌────────────────────────────────────┐  │
│  │   SQLite (metadata) │  │   Vector Store (Qdrant/USearch)  │  │
│  └────────────────────┘  └────────────────────────────────────┘  │
│  ┌────────────────────┐  ┌────────────────────────────────────┐  │
│  │   Python AI Bridge  │  │   Embedding Service (optional)      │  │
│  └────────────────────┘  └────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────┘
```

### Directory Structure

```
local-memory/
├── cmd/                        # Application entry points
│   ├── cli/                    # CLI tool
│   │   ├── main.go
│   │   └── commands/           # save, query, list, forget, stats
│   ├── server/                 # HTTP API server
│   │   └── main.go
│   └── mcp/                    # MCP Server for Claude Code
│       └── main.go
│
├── core/                       # Core memory logic (no dependencies)
│   ├── memory.go              # Memory struct & types
│   ├── store.go              # Save with evolution
│   ├── recall.go             # Query & retrieval
│   ├── ranker.go            # Ranking algorithm
│   ├── decay.go              # Time-based decay
│   ├── evolve.go             # Memory merging
│   ├── forget.go             # Soft/hard delete
│   └── link.go               # Memory associations
│
├── storage/                   # Storage layer
│   ├── sqlite.go             # SQLite adapter (metadata)
│   └── vector/              # Vector storage
│       ├── interface.go      # VectorStore interface
│       ├── qdrant.go        # Qdrant adapter
│       └── usearch.go       # USearch adapter (in-memory)
│
├── bridge/                    # Go-Python communication
│   ├── http.go              # HTTP client
│   └── pybridge.go          # Python service wrapper
│
├── server/                    # HTTP API service
│   ├── handlers.go          # API handlers
│   └── middleware.go        # Logger, CORS
│
├── agent/                     # Agent integration
│   └── mcp/
│       └── server.go        # MCP JSON-RPC server
│
├── config/                    # Configuration
│   └── config.go            # JSON config loader
│
├── python/                     # Python AI services
│   ├── server.py            # FastAPI server
│   ├── ai/
│   │   ├── embedding.py     # sentence-transformers
│   │   └── extractor.py     # Rule-based memory extraction
│   └── requirements.txt
│
├── tests/                     # Test suites
│   ├── unit/                # Unit tests
│   └── integration/         # Integration tests (require CGO)
│
├── config.json               # Configuration file
├── Makefile                  # Build automation
├── Dockerfile               # Container image
└── docker-compose.yaml      # Full stack deployment
```

### Data Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                         Save Memory Flow                          │
└─────────────────────────────────────────────────────────────────┘

User Input
    │
    ▼
┌─────────────┐    ┌──────────────┐    ┌─────────────────────────┐
│   CLI/API   │───▶│  Store.Save  │───▶│  Evolve (if same key)  │
└─────────────┘    └──────┬───────┘    └───────────┬─────────────┘
                          │                        │
                          ▼                        ▼
                   ┌──────────────┐        ┌───────────────┐
                   │    SQLite     │        │ Merge memory  │
                   │  (metadata)   │        │ Update conf. │
                   └──────────────┘        └───────────────┘
                          │
                          ▼ (async)
                   ┌──────────────┐        ┌───────────────┐
                   │   Python AI  │───────▶│  Embedding    │
                   │   (bridge)   │        │  generation   │
                   └──────────────┘        └───────┬───────┘
                                                   │
                                                   ▼
                                           ┌──────────────┐
                                           │   Qdrant     │
                                           │ (vector DB)   │
                                           └──────────────┘


┌─────────────────────────────────────────────────────────────────┐
│                         Query Memory Flow                        │
└─────────────────────────────────────────────────────────────────┘

Query Input
    │
    ▼
┌─────────────┐    ┌──────────────┐    ┌─────────────────────────┐
│   CLI/API   │───▶│ Recall.Query │───▶│  Generate embedding     │
└─────────────┘    └──────┬───────┘    └───────────┬─────────────┘
                          │                        │
                          ▼ (if vector available)   ▼
                   ┌──────────────┐        ┌───────────────┐
                   │   Qdrant     │───────▶│ Vector search │
                   │              │        │ Return IDs    │
                   └──────────────┘        └───────┬───────┘
                                                   │
                                                   ▼
                                           ┌──────────────┐
                                           │    SQLite     │
                                           │  Get full    │
                                           │  memories    │
                                           └───────┬──────┘
                                                   │
                                                   ▼
                                           ┌──────────────┐
                                           │   Ranker     │
                                           │ (score by    │
                                           │ similarity + │
                                           │ recency +    │
                                           │ confidence)  │
                                           └───────┬──────┘
                                                   │
                                                   ▼
                                           ┌──────────────┐
                                           │  Top-K       │
                                           │  Results     │
                                           └──────────────┘
```

---

## Quick Start

### Prerequisites

- **Go** 1.21+
- **Python** 3.10+ (for AI features)
- **Docker** & **Docker Compose** (for full deployment)

### Option 1: CLI Only (MVP)

```bash
# Clone and build
git clone https://github.com/yourusername/local-memory.git
cd local-memory
make build

# Save a memory (value only, key defaults to first 50 chars)
./localmemory save "User prefers Go language for backend development"

# Save with explicit key-value separation
./localmemory save "language_pref" "User prefers Go language for backend"

# List memories
./localmemory list

# Query memories
./localmemory query "Go language"

# Delete a memory
./localmemory forget <memory-id>

# View statistics
./localmemory stats
```

### Option 2: Full Stack with Docker

```bash
# Start all services
docker-compose up -d

# Wait for services to be ready
docker-compose ps

# Save via HTTP API
curl -X POST http://localhost:8080/api/v1/memories \
  -H "Content-Type: application/json" \
  -d '{"type":"preference","key":"language","value":"Go","scope":"global"}'

# Query via API
curl -X POST http://localhost:8080/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{"query":"language","topk":5}'
```

### Option 3: Claude Code Integration

**Step 1: Build the MCP binary**

```bash
# From the local-memory directory
go build -o localmemory-mcp.exe ./cmd/mcp/
```

**Step 2: Add MCP server to Claude Code**

```bash
# Use claude mcp CLI to add the server
claude mcp add localmemory -- /path/to/localmemory-mcp.exe

# Windows example:
claude mcp add localmemory -- D:\code\local-memory\localmemory-mcp.exe

# Linux/macOS example:
claude mcp add localmemory -- /home/user/local-memory/localmemory-mcp
```

**Step 3: Restart Claude Code**

After adding the MCP server, **completely exit Claude Code** (ensure the process terminates) and restart.

**Step 4: Verify connection**

Run `/mcp` in Claude Code to confirm the localmemory server is connected.

**Available tools:**

| Tool | Description |
|------|-------------|
| `memory_save` | Save a new memory (key, value, type, scope, tags, confidence) |
| `memory_query` | Search memories semantically (query, topk, scope) |
| `memory_list` | List memories (limit, scope) |
| `memory_forget` | Delete a memory by ID |

**Example usage in Claude Code:**

```
User: Remember I prefer Go over Python
AI: [calls memory_save with type="preference", key="language", value="Go"]

User: What language should I use?
AI: [calls memory_query to find language preferences]

User: List all my memories
AI: [calls memory_list to show all memories]
```

---

## Installation

### From Source

```bash
# Clone repository
git clone https://github.com/yourusername/local-memory.git
cd local-memory

# Download dependencies
go mod download

# Build all binaries
make build

# Install to ~/.localmemory/
make install
```

### Using Makefile

```bash
make help                # Show all targets
make build              # Build CLI, server, and MCP binaries
make test               # Run unit tests
make test-integration   # Run integration tests (requires CGO)
make run               # Run CLI with sample data
make docker-build      # Build Docker image
make docker-up          # Start with docker-compose
make clean             # Clean build artifacts
```

### Docker Deployment

```bash
# Build image
docker build -t localmemory:latest .

# Or use docker-compose for full stack
docker-compose up -d

# View logs
docker-compose logs -f localmemory

# Stop services
docker-compose down
```

---

## Configuration

### Configuration File

Create `~/.localmemory/config.json` or use the default location `config.json`:

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

### Configuration Options

| Section | Option | Default | Description |
|---------|--------|---------|-------------|
| `database.path` | - | `./data/localmemory.db` | SQLite database file path |
| `vector_db.type` | - | `qdrant` | Vector store type: `qdrant` or `usearch` |
| `vector_db.url` | - | `http://127.0.0.1:6333` | Qdrant server URL |
| `vector_db.collection` | - | `memories` | Qdrant collection name |
| `bridge.type` | - | `tcp` | Go-Python communication type |
| `bridge.tcp_url` | - | `127.0.0.1:8081` | Python AI server address |
| `ai.embedding_model` | - | `all-MiniLM-L6-v2` | Sentence-transformers model |
| `decay.lambda` | - | `0.01` | Decay coefficient (higher = faster decay) |
| `server.port` | - | `8080` | HTTP API server port |
| `cli.default_topk` | - | `5` | Default number of results |
| `cli.default_scope` | - | `global` | Default memory scope |

### Environment Variables

```bash
# Override config file path
LOCALMEMORY_CONFIG=/path/to/config.json

# Or use inline config
LOCALMEMORY_DB_PATH=./data/localmemory.db
```

---

## Usage

### Quick Start by Platform

#### Windows

```powershell
# Download and run directly (no installation required)
.\localmemory.exe save "User prefers Go language"
.\localmemory.exe save "language" "User prefers Go language"
.\localmemory.exe list
.\localmemory.exe query Go
.\localmemory.exe stats
.\localmemory.exe forget <memory-id>

# Start HTTP server
.\localmemory-server.exe
```

#### Linux

```bash
# Download and make executable
chmod +x localmemory
./localmemory save "User prefers Go language"
./localmemory save "language" "User prefers Go language"
./localmemory list
./localmemory query Go
./localmemory stats
./localmemory forget <memory-id>

# Start HTTP server
./localmemory-server &
```

#### macOS

```bash
# Download and make executable
chmod +x localmemory
./localmemory save "User prefers Go language"
./localmemory save "language" "User prefers Go language"
./localmemory list
./localmemory query Go
./localmemory stats
./localmemory forget <memory-id>

# Start HTTP server
./localmemory-server &
```

### CLI Commands

```bash
# Save a memory (value only, key defaults to first 50 chars of value)
localmemory save "User prefers dark mode"

# Save with explicit key-value separation
localmemory save "theme" "User prefers dark mode"
localmemory save "language" "Go developer" --type fact --scope global

# Save with additional options
localmemory save "Image description" --type fact --media-type image

# Query memories
localmemory query "theme preference" --topk 10
localmemory query "programming language" --scope global

# List memories
localmemory list
localmemory list --scope global --limit 20 --offset 0
localmemory list --all  # Include deleted

# Delete memory (soft delete)
localmemory forget <memory-id>

# Permanent delete
localmemory forget <memory-id> --hard

# View statistics
localmemory stats
```

### HTTP API

```bash
# Base URL
export BASE_URL=http://localhost:8080

# Health check
curl $BASE_URL/health

# Create memory
curl -X POST $BASE_URL/api/v1/memories \
  -H "Content-Type: application/json" \
  -d '{
    "type": "preference",
    "scope": "global",
    "key": "editor",
    "value": "VS Code",
    "confidence": 0.9,
    "tags": ["editor", "development"]
  }'

# List memories
curl "$BASE_URL/api/v1/memories?scope=global&limit=10"

# Get single memory
curl "$BASE_URL/api/v1/memories/<id>"

# Query memories (semantic search)
curl -X POST $BASE_URL/api/v1/query \
  -H "Content-Type: application/json" \
  -d '{"query": "editor preference", "topk": 5, "scope": "global"}'

# Delete memory
curl -X DELETE "$BASE_URL/api/v1/memories/<id>"

# Get statistics
curl $BASE_URL/api/v1/stats
```

### MCP Tools (Claude Code)

After configuring MCP in Claude Code settings:

```
User: Remember I prefer TypeScript over Python
AI: [uses memory_save tool to store preference]

User: What language should I use for this backend?
AI: [uses memory_query to find language preferences]
```

---

## Development

### Running Tests

```bash
# Unit tests only (no CGO required)
go test ./tests/unit/... -v

# Integration tests (requires CGO-enabled Go)
CGO_ENABLED=1 go test ./tests/integration/... -v

# All tests
make test
```

### Project Structure Guidelines

```
┌─────────────────────────────────────────────────────────────┐
│                    Package Organization                       │
├─────────────────────────────────────────────────────────────┤
│  core/          │ Pure Go, no external dependencies         │
│                 │ Business logic only                        │
├─────────────────┼───────────────────────────────────────────┤
│  storage/       │ Data persistence abstractions               │
│                 │ SQLite, Qdrant, USearch adapters          │
├─────────────────┼───────────────────────────────────────────┤
│  bridge/        │ External service communication              │
│                 │ Python AI service bridge                   │
├─────────────────┼───────────────────────────────────────────┤
│  cmd/           │ Application entry points                    │
│                 │ CLI, HTTP server, MCP server              │
├─────────────────┼───────────────────────────────────────────┤
│  server/        │ HTTP API implementation                     │
│                 │ Handlers, middleware, routing             │
├─────────────────┼───────────────────────────────────────────┤
│  agent/         │ Agent protocol implementations             │
│                 │ MCP JSON-RPC server                        │
├─────────────────┼───────────────────────────────────────────┤
│  python/        │ AI services (separate process)            │
│                 │ FastAPI, sentence-transformers            │
└─────────────────┴───────────────────────────────────────────┘
```

### Adding New Memory Types

1. Define type in `core/memory.go`:
```go
const (
    TypePreference MemoryType = "preference"
    // Add new type
    TypeCustom    MemoryType = "custom"
)
```

2. Update validation and handlers as needed.

### Adding Vector Store Backends

Implement the `VectorStore` interface in `storage/vector/interface.go`:

```go
type VectorStore interface {
    Upsert(id string, vector []float32, metadata map[string]any) error
    Search(query []float32, topK int, filter *Filter) ([]Result, error)
    Delete(id string) error
    Close() error
}
```

---

## Architecture Deep Dive

### Memory Evolution (Evolve)

When saving a memory with an existing key:

```
Existing Memory          New Memory
┌─────────────────┐     ┌─────────────────┐
│ key: language   │     │ key: language   │
│ value: "Go"     │     │ value: "Python" │
│ confidence: 0.9 │     │ confidence: 0.8  │
└────────┬────────┘     └─────────────────┘
         │
         ▼ Merge
┌─────────────────┐
│ key: language   │
│ value: "Go\n    │  ← Appended
│        Python"  │
│ confidence: 1.0 │  ← Incremented
└─────────────────┘
```

### Ranking Algorithm

Final score combines multiple factors:

```
Score = (similarity × 0.7) + (recency × 0.2) + (confidence × 0.1)
```

Where:
- **similarity**: Vector distance from query (semantic match)
- **recency**: Exponential decay based on creation time
- **confidence**: User-assigned or system-computed trust level

### Time Decay

Memory importance naturally fades:

```
weight = e^(-λ × Δt)

Examples:
  λ = 0.01, Δt = 1 hour  → weight ≈ 0.97 (93% importance)
  λ = 0.01, Δt = 1 day   → weight ≈ 0.42 (42% importance)
  λ = 0.01, Δt = 7 days  → weight ≈ 0.00045 (<1% importance)
```

### Memory Linking

Create associations between memories:

```
┌──────────────────┐         ┌──────────────────┐
│ Memory: Go偏好    │◀──────▶│ Memory: 项目架构  │
│ ID: abc123       │  link   │ ID: def456       │
└──────────────────┘         └──────────────────┘
         │
         ▼ (traverse)
┌──────────────────┐         ┌──────────────────┐
│ Memory: Web服务   │◀──────▶│ Memory: 数据库    │
│ ID: ghi789       │         │ ID: jkl012       │
└──────────────────┘         └──────────────────┘
```

Use `Link()` to connect, `GetRelated()` to traverse.

---

## License

MIT License - see [LICENSE](LICENSE) for details.

---

## Contributing

Contributions welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Support

- 📖 [Documentation](docs/architecture.md)
- 🐛 [Issue Tracker](https://github.com/yourusername/local-memory/issues)
