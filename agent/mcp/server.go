package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"localmemory/bridge"
	"localmemory/config"
	"localmemory/core"
	"localmemory/storage"
	"localmemory/storage/vector"
)

// MCPServer implements the MCP Server.
// Uses stdio transport protocol to communicate with Claude Code.
type MCPServer struct {
	cfg           *config.Config
	sqliteStore   *storage.SQLiteStore
	store         *core.Store
	recall        *core.Recall
	ranker        *core.Ranker
	forget        *core.Forget
	link          *core.Link
	vectorStore   core.VectorStore
	embeddingSvc  core.EmbeddingService
	useSemanticSearch bool
	mu            sync.Mutex
}

// NewMCPServer creates an MCP Server instance.
func NewMCPServer(cfg *config.Config) (*MCPServer, error) {
	sqliteStore, err := storage.NewSQLiteStore(cfg.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to init storage: %w", err)
	}

	server := &MCPServer{
		cfg:         cfg,
		sqliteStore: sqliteStore,
		store:       core.NewStore(sqliteStore, nil, nil),
		ranker:      core.NewRanker(cfg.Decay.Lambda),
		forget:      core.NewForget(sqliteStore, nil),
		link:        core.NewLink(sqliteStore),
	}

	// Try to initialize vector store and embedding service for semantic search
	server.initAIComponents()

	return server, nil
}

// initAIComponents initializes vector store and embedding service.
func (s *MCPServer) initAIComponents() {
	// Initialize vector store
	vectorCfg := vector.USearchConfig{
		Path:       "./data/usearch",
		VectorSize: 384, // Standard embedding dimension for all-MiniLM-L6-v2
		Metric:     "cosine",
	}
	vs, err := vector.NewVectorStore("usearch", vectorCfg)
	if err == nil {
		// Create adapter to convert vector.VectorStore to core.VectorStore
		s.vectorStore = &vectorStoreAdapter{store: vs}
		s.useSemanticSearch = true
	}

	// Initialize embedding service (Python AI bridge)
	if s.cfg.Bridge.TCPURL != "" {
		pb := bridge.NewPyBridge("http://" + s.cfg.Bridge.TCPURL)
		// Test if Python AI service is available
		if err := pb.HealthCheck(); err == nil {
			s.embeddingSvc = pb
		}
	}

	// Create store and recall with real components if available
	if s.vectorStore != nil && s.embeddingSvc != nil {
		s.store = core.NewStore(s.sqliteStore, s.vectorStore, s.embeddingSvc)
		s.recall = core.NewRecall(s.sqliteStore, s.vectorStore, s.embeddingSvc, s.ranker)
	}
}

// vectorStoreAdapter wraps storage/vector.VectorStore to implement core.VectorStore.
type vectorStoreAdapter struct {
	store interface {
		Upsert(id string, vector []float32, metadata map[string]any) error
		Search(query []float32, topK int, filter *vector.Filter) ([]vector.Result, error)
		Delete(id string) error
		Close() error
	}
}

func (a *vectorStoreAdapter) Upsert(id string, vector []float32, metadata map[string]any) error {
	return a.store.Upsert(id, vector, metadata)
}

func (a *vectorStoreAdapter) Search(query []float32, topK int, filter *core.VectorFilter) ([]core.VectorResult, error) {
	vf := &vector.Filter{}
	if filter != nil {
		vf = &vector.Filter{
			Scope:     filter.Scope,
			Type:      filter.Type,
			Tags:      filter.Tags,
			ProfileID: filter.ProfileID,
		}
	}
	results, err := a.store.Search(query, topK, vf)
	if err != nil {
		return nil, err
	}
	var coreResults []core.VectorResult
	for _, r := range results {
		coreResults = append(coreResults, core.VectorResult{
			ID:       r.ID,
			Score:    r.Score,
			Metadata: r.Metadata,
		})
	}
	return coreResults, nil
}

func (a *vectorStoreAdapter) Delete(id string) error {
	return a.store.Delete(id)
}

func (a *vectorStoreAdapter) Close() error {
	return a.store.Close()
}

// Ensure adapter implements core.VectorStore
var _ core.VectorStore = (*vectorStoreAdapter)(nil)

// Run starts the MCP Server.
// Reads requests from stdin, outputs responses to stdout.
func (s *MCPServer) Run() error {
	decoder := json.NewDecoder(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	encoder := json.NewEncoder(writer)

	for {
		var req JSONRPCRequest
		if err := decoder.Decode(&req); err != nil {
			if err.Error() == "EOF" {
				return nil
			}
			continue
		}

		// Handle notifications (no response needed)
		if req.ID == nil {
			s.handleNotification(&req)
			continue
		}

		resp := s.handleRequest(&req)
		encoder.Encode(resp)
		writer.Flush()
	}
}

// JSONRPCRequest represents a JSON-RPC request.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC response.
type JSONRPCResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	Result interface{}       `json:"result,omitempty"`
	Error  *JSONRPCError     `json:"error,omitempty"`
	ID     interface{}       `json:"id,omitempty"`
}

// JSONRPCError represents a JSON-RPC error.
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// handleNotification handles JSON-RPC notifications (no response needed).
func (s *MCPServer) handleNotification(req *JSONRPCRequest) {
	switch req.Method {
	case "initialized":
		// Claude Code sends this after initialize, acknowledgment only
	}
}

// handleRequest handles JSON-RPC requests.
func (s *MCPServer) handleRequest(req *JSONRPCRequest) *JSONRPCResponse {
	s.mu.Lock()
	defer s.mu.Unlock()

	var result interface{}
	var err error

	switch req.Method {
	case "initialize":
		result, err = s.handleInitialize(req.Params)
	case "tools/list":
		result, err = s.handleToolsList(req.Params)
	case "tools/call":
		result, err = s.handleToolsCall(req.Params)
	case "resources/list":
		result, err = s.handleResourcesList(req.Params)
	case "resources/read":
		result, err = s.handleResourcesRead(req.Params)
	case "ping":
		result = map[string]string{"status": "ok"}
	default:
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &JSONRPCError{
				Code:    -32601,
				Message: "Method not found: " + req.Method,
			},
			ID: req.ID,
		}
	}

	if err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &JSONRPCError{
				Code:    -32603,
				Message: err.Error(),
			},
			ID: req.ID,
		}
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      req.ID,
	}
}

// handleInitialize handles initialization requests.
func (s *MCPServer) handleInitialize(params json.RawMessage) (interface{}, error) {
	return map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]any{
			"tools": map[string]any{},
		},
		"serverInfo": map[string]any{
			"name":    "localmemory",
			"version": "1.0.0",
		},
	}, nil
}

// handleToolsList handles tool list requests.
func (s *MCPServer) handleToolsList(params json.RawMessage) (interface{}, error) {
	return map[string]any{
		"tools": []map[string]any{
			{
				"name":        "memory_save",
				"description": "Save a new memory to LocalMemory",
				"inputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"type":       map[string]any{"type": "string", "enum": []string{"preference", "fact", "event", "skill", "goal"}},
						"scope":      map[string]any{"type": "string", "enum": []string{"global", "session", "agent"}},
						"profile_id": map[string]any{"type": "string"},
						"key":        map[string]any{"type": "string"},
						"value":      map[string]any{"type": "string"},
						"confidence": map[string]any{"type": "number", "minimum": 0, "maximum": 1},
						"tags":       map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					},
					"required": []string{"key", "value"},
				},
			},
			{
				"name":        "memory_query",
				"description": "Search memories semantically",
				"inputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"query":      map[string]any{"type": "string"},
						"topk":       map[string]any{"type": "integer", "default": 5},
						"scope":      map[string]any{"type": "string"},
						"profile_id": map[string]any{"type": "string"},
					},
					"required": []string{"query"},
				},
			},
			{
				"name":        "memory_list",
				"description": "List memories",
				"inputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"scope":      map[string]any{"type": "string"},
						"limit":      map[string]any{"type": "integer", "default": 20},
						"profile_id": map[string]any{"type": "string"},
					},
				},
			},
			{
				"name":        "memory_forget",
				"description": "Delete a memory",
				"inputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id": map[string]any{"type": "string"},
					},
					"required": []string{"id"},
				},
			},
		},
	}, nil
}

// handleToolsCall handles tool call requests.
func (s *MCPServer) handleToolsCall(params json.RawMessage) (interface{}, error) {
	var callParams struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(params, &callParams); err != nil {
		return nil, err
	}

	switch callParams.Name {
	case "memory_save":
		return s.toolMemorySave(callParams.Arguments)
	case "memory_query":
		return s.toolMemoryQuery(callParams.Arguments)
	case "memory_list":
		return s.toolMemoryList(callParams.Arguments)
	case "memory_forget":
		return s.toolMemoryForget(callParams.Arguments)
	default:
		return nil, fmt.Errorf("unknown tool: %s", callParams.Name)
	}
}

// toolMemorySave saves a memory.
func (s *MCPServer) toolMemorySave(args json.RawMessage) (interface{}, error) {
	var params struct {
		Type       string   `json:"type"`
		Scope      string   `json:"scope"`
		ProfileID  string   `json:"profile_id"`
		Key        string   `json:"key"`
		Value      string   `json:"value"`
		Confidence float64  `json:"confidence"`
		Tags       []string `json:"tags"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}

	profileID := params.ProfileID
	if profileID == "" {
		profileID = s.cfg.Profile.ID
	}
	if profileID == "" {
		profileID = "default"
	}

	memory := &core.Memory{
		ProfileID:  profileID,
		Type:       core.MemoryType(params.Type),
		Scope:      core.Scope(params.Scope),
		Key:        params.Key,
		Value:      params.Value,
		Tags:       params.Tags,
		Confidence: params.Confidence,
		Metadata: core.Metadata{
			Source: "claude_code",
		},
	}

	if err := s.store.Save(memory); err != nil {
		return nil, err
	}

	return map[string]any{
		"content": []map[string]any{
			{
				"type": "text",
				"text": fmt.Sprintf("Memory saved: ID=%s, Key=%s", memory.ID, memory.Key),
			},
		},
	}, nil
}

// toolMemoryQuery searches memories.
func (s *MCPServer) toolMemoryQuery(args json.RawMessage) (interface{}, error) {
	var params struct {
		Query     string `json:"query"`
		TopK      int    `json:"topk"`
		Scope     string `json:"scope"`
		ProfileID string `json:"profile_id"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}

	if params.TopK <= 0 {
		params.TopK = 5
	}

	profileID := params.ProfileID
	if profileID == "" {
		profileID = s.cfg.Profile.ID
	}
	if profileID == "" {
		profileID = "default"
	}

	var results []*core.QueryResult

	// Try semantic search if AI components are available
	if s.useSemanticSearch && s.recall != nil {
		queryReq := &core.QueryRequest{
			Query:     params.Query,
			TopK:      params.TopK,
			Scope:     core.Scope(params.Scope),
			ProfileID: profileID,
		}
		queryResp, err := s.recall.Query(queryReq)
		if err == nil && len(queryResp.Results) > 0 {
			results = queryResp.Results
		}
	}

	// Fallback to keyword matching if semantic search didn't return results
	if len(results) == 0 {
		recall := core.NewRecall(s.sqliteStore, nil, nil, s.ranker)
		listReq := &core.ListRequest{
			Scope:     core.Scope(params.Scope),
			Limit:     100,
			ProfileID: profileID,
		}

		resp, err := recall.List(listReq)
		if err != nil {
			return nil, err
		}

		for _, m := range resp.Memories {
			if containsString(m.Value, params.Query) || containsString(m.Key, params.Query) {
				results = append(results, &core.QueryResult{
					Memory: m,
					Score:  0.8,
				})
				if len(results) >= params.TopK {
					break
				}
			}
		}
	}

	// Build output
	var text string
	if len(results) == 0 {
		text = "No related memories found"
	} else {
		text = "Found related memories:\n"
		for i, r := range results {
			text += fmt.Sprintf("%d. [%s] %s: %s (score: %.2f)\n", i+1, r.Memory.Type, r.Memory.Key, truncate(r.Memory.Value, 100), r.Score)
		}
	}

	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": text},
		},
	}, nil
}

// toolMemoryList lists memories.
func (s *MCPServer) toolMemoryList(args json.RawMessage) (interface{}, error) {
	var params struct {
		Scope     string `json:"scope"`
		Limit     int    `json:"limit"`
		ProfileID string `json:"profile_id"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}

	if params.Limit <= 0 {
		params.Limit = 20
	}

	profileID := params.ProfileID
	if profileID == "" {
		profileID = s.cfg.Profile.ID
	}
	if profileID == "" {
		profileID = "default"
	}

	var memories []*core.Memory
	var total int

	// Use recall if available, otherwise create one with nil components
	if s.recall != nil {
		resp, err := s.recall.List(&core.ListRequest{
			Scope:     core.Scope(params.Scope),
			Limit:     params.Limit,
			ProfileID: profileID,
		})
		if err != nil {
			return nil, err
		}
		memories = resp.Memories
		total = resp.Total
	} else {
		recall := core.NewRecall(s.sqliteStore, nil, nil, s.ranker)
		resp, err := recall.List(&core.ListRequest{
			Scope:     core.Scope(params.Scope),
			Limit:     params.Limit,
			ProfileID: profileID,
		})
		if err != nil {
			return nil, err
		}
		memories = resp.Memories
		total = resp.Total
	}

	text := fmt.Sprintf("Memory list (%d total):\n", total)
	for i, m := range memories {
		text += fmt.Sprintf("%d. [%s] %s: %s\n", i+1, m.Type, m.Key, truncate(m.Value, 80))
	}

	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": text},
		},
	}, nil
}

// toolMemoryForget deletes a memory.
func (s *MCPServer) toolMemoryForget(args json.RawMessage) (interface{}, error) {
	var params struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}

	if err := s.forget.Delete(params.ID); err != nil {
		return nil, err
	}

	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": fmt.Sprintf("Memory deleted: %s", params.ID)},
		},
	}, nil
}

// handleResourcesList handles resource list requests.
func (s *MCPServer) handleResourcesList(params json.RawMessage) (interface{}, error) {
	return map[string]any{
		"resources": []map[string]any{
			{"uri": "memory://all", "name": "all_memories", "description": "All memories", "mimeType": "application/json"},
			{"uri": "memory://stats", "name": "memory_stats", "description": "Statistics", "mimeType": "application/json"},
		},
	}, nil
}

// handleResourcesRead handles resource read requests.
func (s *MCPServer) handleResourcesRead(params json.RawMessage) (interface{}, error) {
	var reqParams struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal(params, &reqParams); err != nil {
		return nil, err
	}

	switch reqParams.URI {
	case "memory://all":
		recall := core.NewRecall(s.sqliteStore, nil, nil, s.ranker)
		resp, err := recall.List(&core.ListRequest{Limit: 100})
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"contents": []map[string]any{
				{"uri": "memory://all", "mimeType": "application/json", "text": mustMarshalJSON(resp)},
			},
		}, nil
	case "memory://stats":
		stats, err := s.sqliteStore.GetStats()
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"contents": []map[string]any{
				{"uri": "memory://stats", "mimeType": "application/json", "text": mustMarshalJSON(stats)},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown resource: %s", reqParams.URI)
	}
}

// containsString checks if string contains substring.
func containsString(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// truncate truncates a string.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// mustMarshalJSON forces JSON serialization.
func mustMarshalJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
