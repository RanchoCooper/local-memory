package unit

import (
	"os"
	"testing"

	"localmemory/config"
)

func TestConfig_Default(t *testing.T) {
	cfg := config.Default()

	if cfg.Database.Path != "./data/localmemory.db" {
		t.Errorf("Expected database path './data/localmemory.db', got '%s'", cfg.Database.Path)
	}
	if cfg.VectorDB.Type != "qdrant" {
		t.Errorf("Expected vector type 'qdrant', got '%s'", cfg.VectorDB.Type)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Expected server port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Decay.Lambda != 0.01 {
		t.Errorf("Expected decay lambda 0.01, got %f", cfg.Decay.Lambda)
	}
}

func TestConfig_Get(t *testing.T) {
	// First call should return default
	cfg1 := config.Get()
	if cfg1 == nil {
		t.Fatal("Expected non-nil config")
	}
	if cfg1.Server.Port != 8080 {
		t.Error("Expected default port 8080")
	}

	// Second call should return same instance
	cfg2 := config.Get()
	if cfg1 != cfg2 {
		t.Error("Expected same config instance")
	}
}

func TestConfig_Load(t *testing.T) {
	// Test loading non-existent file should return default
	cfg, err := config.Load("/nonexistent/path/config.json")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if cfg == nil {
		t.Fatal("Expected non-nil config")
	}

	// Test loading valid config
	tmpfile, err := os.CreateTemp("", "config_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	configContent := `{
		"database": {"path": "./test.db"},
		"server": {"port": 9090}
	}`
	if _, err := tmpfile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tmpfile.Close()

	cfg, err = config.Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if cfg.Database.Path != "./test.db" {
		t.Errorf("Expected path './test.db', got '%s'", cfg.Database.Path)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", cfg.Server.Port)
	}
}
