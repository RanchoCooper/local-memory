package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"localmemory/config"
	"localmemory/core"
	"localmemory/storage"
)

// MCPServer MCP Server 实现
// 使用 stdio 传输协议与 Claude Code 通信
type MCPServer struct {
	cfg         *config.Config
	sqliteStore *storage.SQLiteStore
	store      *core.Store
	ranker     *core.Ranker
	forget     *core.Forget
	link       *core.Link
	mu         sync.Mutex
}

// NewMCPServer 创建 MCP Server 实例
func NewMCPServer(cfg *config.Config) (*MCPServer, error) {
	sqliteStore, err := storage.NewSQLiteStore(cfg.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to init storage: %w", err)
	}

	server := &MCPServer{
		cfg:         cfg,
		sqliteStore: sqliteStore,
		store:      core.NewStore(sqliteStore, nil, nil),
		ranker:     core.NewRanker(cfg.Decay.Lambda),
		forget:     core.NewForget(sqliteStore, nil),
		link:       core.NewLink(sqliteStore),
	}

	return server, nil
}

// Run 启动 MCP Server
// 从 stdin 读取请求，输出响应到 stdout
func (s *MCPServer) Run() error {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		var req JSONRPCRequest
		if err := decoder.Decode(&req); err != nil {
			if err.Error() == "EOF" {
				return nil
			}
			continue
		}

		resp := s.handleRequest(&req)
		encoder.Encode(resp)
	}
}

// JSONRPCRequest JSON-RPC 请求
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage  `json:"params,omitempty"`
	ID      interface{}      `json:"id,omitempty"`
}

// JSONRPCResponse JSON-RPC 响应
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result interface{}  `json:"result,omitempty"`
	Error  *JSONRPCError `json:"error,omitempty"`
	ID     interface{}  `json:"id,omitempty"`
}

// JSONRPCError JSON-RPC 错误
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// handleRequest 处理 JSON-RPC 请求
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

// handleInitialize 处理初始化请求
func (s *MCPServer) handleInitialize(params json.RawMessage) (interface{}, error) {
	return map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]any{
			"tools":    map[string]any{},
			"resources": map[string]any{},
		},
		"serverInfo": map[string]any{
			"name":    "localmemory",
			"version": "1.0.0",
		},
	}, nil
}

// handleToolsList 处理工具列表请求
func (s *MCPServer) handleToolsList(params json.RawMessage) (interface{}, error) {
	return map[string]any{
		"tools": []map[string]any{
			{
				"name":        "memory_save",
				"description": "保存新的记忆到 LocalMemory",
				"inputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"type":       map[string]any{"type": "string", "enum": []string{"preference", "fact", "event", "skill", "goal"}},
						"scope":      map[string]any{"type": "string", "enum": []string{"global", "session", "agent"}},
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
				"description": "语义检索记忆",
				"inputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"query": map[string]any{"type": "string"},
						"topk":  map[string]any{"type": "integer", "default": 5},
						"scope":  map[string]any{"type": "string"},
					},
					"required": []string{"query"},
				},
			},
			{
				"name":        "memory_list",
				"description": "列出记忆",
				"inputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"scope": map[string]any{"type": "string"},
						"limit": map[string]any{"type": "integer", "default": 20},
					},
				},
			},
			{
				"name":        "memory_forget",
				"description": "删除记忆",
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

// handleToolsCall 处理工具调用请求
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

// toolMemorySave 保存记忆
func (s *MCPServer) toolMemorySave(args json.RawMessage) (interface{}, error) {
	var params struct {
		Type      string   `json:"type"`
		Scope     string   `json:"scope"`
		Key       string   `json:"key"`
		Value     string   `json:"value"`
		Confidence float64 `json:"confidence"`
		Tags      []string `json:"tags"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}

	memory := &core.Memory{
		Type:      core.MemoryType(params.Type),
		Scope:     core.Scope(params.Scope),
		Key:       params.Key,
		Value:     params.Value,
		Tags:      params.Tags,
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
				"text": fmt.Sprintf("记忆已保存: ID=%s, Key=%s", memory.ID, memory.Key),
			},
		},
	}, nil
}

// toolMemoryQuery 检索记忆
func (s *MCPServer) toolMemoryQuery(args json.RawMessage) (interface{}, error) {
	var params struct {
		Query string `json:"query"`
		TopK  int    `json:"topk"`
		Scope string `json:"scope"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}

	if params.TopK <= 0 {
		params.TopK = 5
	}

	recall := core.NewRecall(s.sqliteStore, nil, nil, s.ranker)
	listReq := &core.ListRequest{
		Scope: core.Scope(params.Scope),
		Limit: 100,
	}

	resp, err := recall.List(listReq)
	if err != nil {
		return nil, err
	}

	// 简单关键词匹配
	var results []*core.QueryResult
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

	// 构建输出
	var text string
	if len(results) == 0 {
		text = "未找到相关记忆"
	} else {
		text = "找到以下相关记忆:\n"
		for i, r := range results {
			text += fmt.Sprintf("%d. [%s] %s: %s\n", i+1, r.Memory.Type, r.Memory.Key, truncate(r.Memory.Value, 100))
		}
	}

	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": text},
		},
	}, nil
}

// toolMemoryList 列出记忆
func (s *MCPServer) toolMemoryList(args json.RawMessage) (interface{}, error) {
	var params struct {
		Scope string `json:"scope"`
		Limit int    `json:"limit"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}

	if params.Limit <= 0 {
		params.Limit = 20
	}

	recall := core.NewRecall(s.sqliteStore, nil, nil, s.ranker)
	resp, err := recall.List(&core.ListRequest{
		Scope: core.Scope(params.Scope),
		Limit: params.Limit,
	})
	if err != nil {
		return nil, err
	}

	text := fmt.Sprintf("记忆列表（共 %d 条）:\n", resp.Total)
	for i, m := range resp.Memories {
		text += fmt.Sprintf("%d. [%s] %s: %s\n", i+1, m.Type, m.Key, truncate(m.Value, 80))
	}

	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": text},
		},
	}, nil
}

// toolMemoryForget 删除记忆
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
			{"type": "text", "text": fmt.Sprintf("记忆已删除: %s", params.ID)},
		},
	}, nil
}

// handleResourcesList 处理资源列表请求
func (s *MCPServer) handleResourcesList(params json.RawMessage) (interface{}, error) {
	return map[string]any{
		"resources": []map[string]any{
			{"uri": "memory://all", "name": "all_memories", "description": "所有记忆", "mimeType": "application/json"},
			{"uri": "memory://stats", "name": "memory_stats", "description": "统计信息", "mimeType": "application/json"},
		},
	}, nil
}

// handleResourcesRead 处理资源读取请求
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

// containsString 简单字符串包含判断
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

// truncate 截断字符串
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// mustMarshalJSON 强制 JSON 序列化
func mustMarshalJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
