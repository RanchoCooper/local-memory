package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config is the top-level configuration structure.
// Contains database, vector store, bridge, AI, decay coefficient, and other service configurations.
type Config struct {
	Database  DatabaseConfig  `json:"database"`  // SQLite database configuration
	VectorDB  VectorDBConfig `json:"vector_db"` // Vector database configuration
	Bridge    BridgeConfig   `json:"bridge"`    // Go-Python communication configuration
	AI        AIConfig       `json:"ai"`        // AI model configuration
	Decay     DecayConfig    `json:"decay"`     // Memory decay configuration
	Server    ServerConfig   `json:"server"`    // HTTP server configuration
	CLI       CLIConfig      `json:"cli"`       // CLI default parameters configuration
	Agent     AgentConfig    `json:"agent"`     // Agent identifier configuration
	Profile   ProfileConfig  `json:"profile"`   // Profile configuration
}

// DatabaseConfig holds SQLite database configuration.
type DatabaseConfig struct {
	Path string `json:"path"` // Database file path
}

// VectorDBConfig holds vector database configuration.
// Supports qdrant and usearch backends.
type VectorDBConfig struct {
	Type       string `json:"type"`       // Vector store type: qdrant | usearch
	URL        string `json:"url"`        // Qdrant service address
	Collection string `json:"collection"` // Collection name
}

// BridgeConfig holds Go-Python communication configuration.
// TCP is used in MVP, can switch to Unix Socket in production.
type BridgeConfig struct {
	Type   string `json:"type"`    // Communication type: tcp | unix | namedpipe
	TCPURL string `json:"tcp_url"` // TCP connection address
}

// AIConfig holds AI model configuration.
type AIConfig struct {
	EmbeddingModel string `json:"embedding_model"` // Embedding model name
}

// DecayConfig holds memory decay configuration.
// Memory automatically decreases weight over time.
type DecayConfig struct {
	Lambda float64 `json:"lambda"` // Decay coefficient, larger value means faster decay
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port int `json:"port"` // Listen port
}

// CLIConfig holds CLI default parameters configuration.
type CLIConfig struct {
	DefaultTopK  int    `json:"default_topk"`   // Default result count
	DefaultScope string `json:"default_scope"`  // Default scope
}

// AgentConfig holds Agent identifier configuration.
type AgentConfig struct {
	ID   string `json:"id"`   // Agent unique identifier
	Name string `json:"name"` // Agent display name
}

// ProfileConfig holds profile configuration.
type ProfileConfig struct {
	ID string `json:"id"` // Default profile ID
}

var cfg *Config

// Load loads configuration from specified path.
// If path is empty, defaults to ~/.localmemory/config.json.
// If file doesn't exist, returns default config instead of error.
func Load(path string) (*Config, error) {
	if path == "" {
		// Default config file path: ~/.localmemory/config.json
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(home, ".localmemory", "config.json")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil // Return default config if file doesn't exist
		}
		return nil, err
	}

	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	cfg = &c
	return &c, nil
}

// Default returns the default configuration.
// Used when config file doesn't exist or for testing scenarios.
func Default() *Config {
	return &Config{
		Database: DatabaseConfig{
			Path: "./data/localmemory.db",
		},
		VectorDB: VectorDBConfig{
			Type:       "qdrant",
			URL:        "http://127.0.0.1:6333",
			Collection: "memories",
		},
		Bridge: BridgeConfig{
			Type:   "tcp",
			TCPURL: "127.0.0.1:8081",
		},
		AI: AIConfig{
			EmbeddingModel: "all-MiniLM-L6-v2",
		},
		Decay: DecayConfig{
			Lambda: 0.01,
		},
		Server: ServerConfig{
			Port: 8080,
		},
		CLI: CLIConfig{
			DefaultTopK:  5,
			DefaultScope: "global",
		},
		Agent: AgentConfig{
			ID:   "localmemory",
			Name: "LocalMemory",
		},
		Profile: ProfileConfig{
			ID: "default",
		},
	}
}

// Get returns the loaded configuration.
// Returns default config if not yet loaded.
func Get() *Config {
	if cfg == nil {
		cfg = Default()
	}
	return cfg
}
