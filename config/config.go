package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config 顶层配置结构
// 包含数据库、向量存储、桥接层、AI、衰减系数等服务配置
type Config struct {
	Database  DatabaseConfig  `json:"database"`  // SQLite 数据库配置
	VectorDB  VectorDBConfig `json:"vector_db"` // 向量数据库配置
	Bridge    BridgeConfig    `json:"bridge"`    // Go-Python 通信配置
	AI        AIConfig        `json:"ai"`        // AI 模型配置
	Decay     DecayConfig     `json:"decay"`     // 记忆衰减配置
	Server    ServerConfig    `json:"server"`     // HTTP 服务器配置
	CLI       CLIConfig       `json:"cli"`        // CLI 默认参数配置
	Agent     AgentConfig     `json:"agent"`      // Agent 标识配置
}

// DatabaseConfig SQLite 数据库配置
type DatabaseConfig struct {
	Path string `json:"path"` // 数据库文件路径
}

// VectorDBConfig 向量数据库配置
// 支持 qdrant 和 usearch 两种后端
type VectorDBConfig struct {
	Type       string `json:"type"`       // 向量存储类型：qdrant | usearch
	URL        string `json:"url"`        // Qdrant 服务地址
	Collection string `json:"collection"` // 集合名称
}

// BridgeConfig Go-Python 通信配置
// MVP 阶段使用 TCP，生产环境可切换到 Unix Socket
type BridgeConfig struct {
	Type   string `json:"type"`   // 通信类型：tcp | unix | namedpipe
	TCPURL string `json:"tcp_url"` // TCP 连接地址
}

// AIConfig AI 模型配置
type AIConfig struct {
	EmbeddingModel string `json:"embedding_model"` // Embedding 模型名称
}

// DecayConfig 记忆衰减配置
// 记忆随时间自动降低权重
type DecayConfig struct {
	Lambda float64 `json:"lambda"` // 衰减系数，值越大衰减越快
}

// ServerConfig HTTP 服务器配置
type ServerConfig struct {
	Port int `json:"port"` // 监听端口
}

// CLIConfig CLI 默认参数配置
type CLIConfig struct {
	DefaultTopK  int    `json:"default_topk"`   // 默认返回结果数
	DefaultScope string `json:"default_scope"`  // 默认作用域
}

// AgentConfig Agent 标识配置
type AgentConfig struct {
	ID   string `json:"id"`   // Agent 唯一标识
	Name string `json:"name"` // Agent 显示名称
}

var cfg *Config

// Load 从指定路径加载配置文件
// 如果 path 为空，则默认加载 ~/.localmemory/config.json
// 如果文件不存在，返回默认配置而非错误
func Load(path string) (*Config, error) {
	if path == "" {
		// 默认配置文件路径：~/.localmemory/config.json
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(home, ".localmemory", "config.json")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil // 文件不存在返回默认配置
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

// Default 返回默认配置
// 用于配置文件不存在时或测试场景
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
	}
}

// Get 返回已加载的配置
// 如果尚未加载，则返回默认配置
func Get() *Config {
	if cfg == nil {
		cfg = Default()
	}
	return cfg
}
