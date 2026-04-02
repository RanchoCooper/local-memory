package bridge

import (
	"context"

	"localmemory/core"
)

// PyBridge is a Python service bridge wrapper.
// Implements core.EmbeddingService interface for use by core modules.
type PyBridge struct {
	httpBridge *HTTPBridge
}

// NewPyBridge creates a PyBridge instance.
func NewPyBridge(baseURL string) *PyBridge {
	return &PyBridge{
		httpBridge: NewHTTPBridge(baseURL, 30),
	}
}

// Embed implements EmbeddingService interface.
// Calls Python service to generate vector embeddings.
func (p *PyBridge) Embed(text string) ([]float32, error) {
	ctx := context.Background()
	return p.httpBridge.Embed(ctx, text)
}

// EmbedBatch implements EmbeddingService interface.
// Batch calls Python service to generate vector embeddings.
func (p *PyBridge) EmbedBatch(texts []string) ([][]float32, error) {
	ctx := context.Background()
	return p.httpBridge.EmbedBatch(ctx, texts)
}

// Extract extracts structured memory from text.
func (p *PyBridge) Extract(text string) (*ExtractResponse, error) {
	ctx := context.Background()
	return p.httpBridge.Extract(ctx, text)
}

// HealthCheck checks if Python service is available.
func (p *PyBridge) HealthCheck() error {
	ctx := context.Background()
	return p.httpBridge.HealthCheck(ctx)
}

// Ensure EmbeddingService interface implementation check.
var _ core.EmbeddingService = (*PyBridge)(nil)
