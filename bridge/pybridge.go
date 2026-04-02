package bridge

import (
	"context"

	"localmemory/core"
)

// PyBridge Python 服务桥接封装
// 实现 core.EmbeddingService 接口，供核心模块使用
type PyBridge struct {
	httpBridge *HTTPBridge
}

// NewPyBridge 创建 PyBridge 实例
func NewPyBridge(baseURL string) *PyBridge {
	return &PyBridge{
		httpBridge: NewHTTPBridge(baseURL, 30),
	}
}

// Embed 实现 EmbeddingService 接口
// 调用 Python 服务生成向量嵌入
func (p *PyBridge) Embed(text string) ([]float32, error) {
	ctx := context.Background()
	return p.httpBridge.Embed(ctx, text)
}

// EmbedBatch 实现 EmbeddingService 接口
// 批量调用 Python 服务生成向量嵌入
func (p *PyBridge) EmbedBatch(texts []string) ([][]float32, error) {
	ctx := context.Background()
	return p.httpBridge.EmbedBatch(ctx, texts)
}

// Extract 从文本中提取结构化记忆
func (p *PyBridge) Extract(text string) (*ExtractResponse, error) {
	ctx := context.Background()
	return p.httpBridge.Extract(ctx, text)
}

// HealthCheck 检查 Python 服务是否可用
func (p *PyBridge) HealthCheck() error {
	ctx := context.Background()
	return p.httpBridge.HealthCheck(ctx)
}

// Ensure EmbeddingService 接口实现检查
var _ core.EmbeddingService = (*PyBridge)(nil)
