package server

import (
	"net/http"

	"localmemory/config"
	"localmemory/core"
	"localmemory/storage"

	"github.com/gin-gonic/gin"
)

// Server represents an HTTP server.
type Server struct {
	router *gin.Engine
	cfg    *config.Config
	store  *core.Store
}

// NewServer creates an HTTP server instance.
func NewServer(cfg *config.Config) *Server {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(loggerMiddleware())
	router.Use(corsMiddleware())

	// Initialize storage
	sqliteStore, err := storage.NewSQLiteStore(cfg.Database.Path)
	if err != nil {
		panic("Failed to init storage: " + err.Error())
	}

	store := core.NewStore(sqliteStore, nil, nil)

	srv := &Server{
		router: router,
		cfg:    cfg,
		store:  store,
	}

	// Register routes
	srv.registerRoutes()

	return srv
}

// registerRoutes registers API routes.
func (s *Server) registerRoutes() {
	// Health check
	s.router.GET("/health", s.healthHandler)

	// API v1
	v1 := s.router.Group("/api/v1")
	{
		// Memory CRUD
		v1.POST("/memories", s.createMemoryHandler)
		v1.GET("/memories", s.listMemoriesHandler)
		v1.GET("/memories/:id", s.getMemoryHandler)
		v1.DELETE("/memories/:id", s.deleteMemoryHandler)

		// Semantic search
		v1.POST("/query", s.queryHandler)

		// Statistics
		v1.GET("/stats", s.statsHandler)
	}
}

// Run starts the server.
func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}

// --- Request/Response structures ---

type CreateMemoryRequest struct {
	Type       string   `json:"type"`
	Scope      string   `json:"scope"`
	MediaType  string   `json:"media_type"`
	Key        string   `json:"key"`
	Value      string   `json:"value"`
	Confidence float64  `json:"confidence"`
	Tags       []string `json:"tags"`
}

type QueryRequest struct {
	Query string `json:"query"`
	TopK  int    `json:"topk"`
	Scope string `json:"scope"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// --- Handlers ---

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]any{
			"status":  "ok",
			"service": "localmemory",
		},
	})
}

func (s *Server) createMemoryHandler(c *gin.Context) {
	var req CreateMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	memory := &core.Memory{
		Type:      core.MemoryType(req.Type),
		Scope:     core.Scope(req.Scope),
		MediaType: core.MediaType(req.MediaType),
		Key:       req.Key,
		Value:     req.Value,
		Tags:      req.Tags,
	}

	if req.Confidence > 0 {
		memory.Confidence = req.Confidence
	}

	if err := s.store.Save(memory); err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to save memory: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Data:    memory,
	})
}

func (s *Server) listMemoriesHandler(c *gin.Context) {
	scope := c.Query("scope")
	limit := 20
	offset := 0

	listReq := &core.ListRequest{
		Scope: core.Scope(scope),
		Limit: limit,
		Offset: offset,
	}

	sqliteStore, err := storage.NewSQLiteStore(s.cfg.Database.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to init storage",
		})
		return
	}
	defer sqliteStore.Close()

	recall := core.NewRecall(sqliteStore, nil, nil, core.NewRanker(s.cfg.Decay.Lambda))
	resp, err := recall.List(listReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to list memories",
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    resp,
	})
}

func (s *Server) getMemoryHandler(c *gin.Context) {
	id := c.Param("id")

	sqliteStore, err := storage.NewSQLiteStore(s.cfg.Database.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to init storage",
		})
		return
	}
	defer sqliteStore.Close()

	recall := core.NewRecall(sqliteStore, nil, nil, core.NewRanker(s.cfg.Decay.Lambda))
	memory, err := recall.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to get memory",
		})
		return
	}

	if memory == nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Error:   "Memory not found",
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    memory,
	})
}

func (s *Server) deleteMemoryHandler(c *gin.Context) {
	id := c.Param("id")

	sqliteStore, err := storage.NewSQLiteStore(s.cfg.Database.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to init storage",
		})
		return
	}
	defer sqliteStore.Close()

	forget := core.NewForget(sqliteStore, nil)
	if err := forget.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to delete memory: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"id": id},
	})
}

func (s *Server) queryHandler(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	if req.TopK <= 0 {
		req.TopK = 5
	}

	sqliteStore, err := storage.NewSQLiteStore(s.cfg.Database.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to init storage",
		})
		return
	}
	defer sqliteStore.Close()

	recall := core.NewRecall(sqliteStore, nil, nil, core.NewRanker(s.cfg.Decay.Lambda))

	// MVP: simple keyword search
	listReq := &core.ListRequest{
		Scope: core.Scope(req.Scope),
		Limit: 100,
	}
	resp, err := recall.List(listReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to query memories",
		})
		return
	}

	// Simple matching
	var results []*core.QueryResult
	for _, m := range resp.Memories {
		if containsString(m.Value, req.Query) || containsString(m.Key, req.Query) {
			results = append(results, &core.QueryResult{
				Memory: m,
				Score:  0.8,
			})
			if len(results) >= req.TopK {
				break
			}
		}
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    results,
	})
}

func (s *Server) statsHandler(c *gin.Context) {
	sqliteStore, err := storage.NewSQLiteStore(s.cfg.Database.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to init storage",
		})
		return
	}
	defer sqliteStore.Close()

	stats, err := sqliteStore.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to get stats",
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    stats,
	})
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
