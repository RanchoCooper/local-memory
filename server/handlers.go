package server

import (
	"net/http"
	"strings"

	"localmemory/bridge"
	"localmemory/config"
	"localmemory/core"
	"localmemory/storage"

	"github.com/gin-gonic/gin"
)

// Server represents an HTTP server.
type Server struct {
	router   *gin.Engine
	cfg      *config.Config
	store    *core.Store
	recall   *core.Recall
	sqliteStore *storage.SQLiteStore
	pyBridge *bridge.PyBridge
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
	recall := core.NewRecall(sqliteStore, nil, nil, core.NewRanker(cfg.Decay.Lambda))

	srv := &Server{
		router:      router,
		cfg:         cfg,
		store:       store,
		recall:      recall,
		sqliteStore: sqliteStore,
		pyBridge:    bridge.NewPyBridge("http://" + cfg.Bridge.TCPURL),
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

		// AI extraction
		v1.POST("/extract", s.extractHandler)

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
	ProfileID  string   `json:"profile_id"`
	Type       string   `json:"type"`
	Scope      string   `json:"scope"`
	MediaType  string   `json:"media_type"`
	Key        string   `json:"key"`
	Value      string   `json:"value"`
	Confidence float64  `json:"confidence"`
	Tags       []string `json:"tags"`
}

type QueryRequest struct {
	Query     string `json:"query"`
	TopK      int    `json:"topk"`
	Scope     string `json:"scope"`
	ProfileID string `json:"profile_id"`
}

type ExtractRequest struct {
	Text string `json:"text"`
}

type ExtractResponse struct {
	Type      string  `json:"type"`
	Key       string  `json:"key"`
	Value     string  `json:"value"`
	Confidence float64 `json:"confidence"`
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

	// Validate memory type
	validTypes := map[core.MemoryType]bool{
		core.TypePreference: true, core.TypeFact: true,
		core.TypeEvent: true, core.TypeSkill: true,
		core.TypeGoal: true, core.TypeRelationship: true,
	}
	if req.Type != "" && !validTypes[core.MemoryType(req.Type)] {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid type: must be preference, fact, event, skill, goal, or relationship",
		})
		return
	}

	// Validate scope
	validScopes := map[core.Scope]bool{
		core.ScopeGlobal: true, core.ScopeSession: true, core.ScopeAgent: true,
	}
	if req.Scope != "" && !validScopes[core.Scope(req.Scope)] {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid scope: must be global, session, or agent",
		})
		return
	}

	// Validate media type
	validMedia := map[core.MediaType]bool{
		core.MediaText: true, core.MediaImage: true,
		core.MediaAudio: true, core.MediaVideo: true,
	}
	if req.MediaType != "" && !validMedia[core.MediaType(req.MediaType)] {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid media_type: must be text, image, audio, or video",
		})
		return
	}

	// Validate confidence range
	if req.Confidence < 0 || req.Confidence > 1 {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Confidence must be between 0 and 1",
		})
		return
	}

	// Validate key length
	if len(req.Key) > 500 {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Key too long: maximum 500 characters",
		})
		return
	}

	profileID := req.ProfileID
	if profileID == "" {
		profileID = s.cfg.Profile.ID
	}
	if profileID == "" {
		profileID = "default"
	}

	memory := &core.Memory{
		ProfileID:  profileID,
		Type:       core.MemoryType(req.Type),
		Scope:      core.Scope(req.Scope),
		MediaType:  core.MediaType(req.MediaType),
		Key:        req.Key,
		Value:      req.Value,
		Tags:       req.Tags,
	}

	if req.Confidence > 0 {
		memory.Confidence = req.Confidence
	}

	if err := s.store.Save(memory); err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to save memory",
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

	resp, err := s.recall.List(listReq)
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

	memory, err := s.recall.GetByID(id)
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

	forget := core.NewForget(s.sqliteStore, nil)
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

	profileID := req.ProfileID
	if profileID == "" {
		profileID = s.cfg.Profile.ID
	}
	if profileID == "" {
		profileID = "default"
	}

	// MVP: simple keyword search
	listReq := &core.ListRequest{
		Scope:     core.Scope(req.Scope),
		Limit:     100,
		ProfileID: profileID,
	}
	resp, err := s.recall.List(listReq)
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
		if strings.Contains(m.Value, req.Query) || strings.Contains(m.Key, req.Query) {
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
	stats, err := s.sqliteStore.GetStats()
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

func (s *Server) extractHandler(c *gin.Context) {
	var req ExtractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	if req.Text == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Text is required",
		})
		return
	}

	// Call Python AI service to extract memory structure
	result, err := s.pyBridge.Extract(req.Text)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to extract memory: " + err.Error(),
		})
		return
	}

	// Return extracted memory structure
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: ExtractResponse{
			Type:       result.Type,
			Key:        result.Key,
			Value:      result.Value,
			Confidence: result.Confidence,
		},
	})
}
