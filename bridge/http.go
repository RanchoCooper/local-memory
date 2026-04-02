package bridge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HTTPBridge HTTP 客户端
// 用于 Go 与 Python AI 服务的通信
type HTTPBridge struct {
	baseURL    string
	client     *http.Client
	timeout    time.Duration
}

// NewHTTPBridge 创建 HTTPBridge 实例
func NewHTTPBridge(baseURL string, timeout time.Duration) *HTTPBridge {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &HTTPBridge{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

// EmbedRequest 嵌入请求
type EmbedRequest struct {
	Text string `json:"text"`
}

// EmbedResponse 嵌入响应
type EmbedResponse struct {
	Embedding []float32 `json:"embedding"`
	Error     string    `json:"error,omitempty"`
}

// EmbedBatchRequest 批量嵌入请求
type EmbedBatchRequest struct {
	Texts []string `json:"texts"`
}

// EmbedBatchResponse 批量嵌入响应
type EmbedBatchResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
	Error      string      `json:"error,omitempty"`
}

// ExtractRequest 提取请求
type ExtractRequest struct {
	Text string `json:"text"`
}

// ExtractResponse 提取响应
type ExtractResponse struct {
	Type       string `json:"type"`
	Key        string `json:"key"`
	Value      string `json:"value"`
	Confidence float64 `json:"confidence"`
	Error      string `json:"error,omitempty"`
}

// Embed 生成单个文本的向量嵌入
func (h *HTTPBridge) Embed(ctx context.Context, text string) ([]float32, error) {
	reqBody := EmbedRequest{Text: text}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", h.baseURL+"/embed", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var respBody EmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if respBody.Error != "" {
		return nil, fmt.Errorf("embedding error: %s", respBody.Error)
	}

	return respBody.Embedding, nil
}

// EmbedBatch 批量生成向量嵌入
func (h *HTTPBridge) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	reqBody := EmbedBatchRequest{Texts: texts}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", h.baseURL+"/embed/batch", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var respBody EmbedBatchResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if respBody.Error != "" {
		return nil, fmt.Errorf("embedding error: %s", respBody.Error)
	}

	return respBody.Embeddings, nil
}

// Extract 从文本中提取结构化记忆
func (h *HTTPBridge) Extract(ctx context.Context, text string) (*ExtractResponse, error) {
	reqBody := ExtractRequest{Text: text}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", h.baseURL+"/extract", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var respBody ExtractResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if respBody.Error != "" {
		return nil, fmt.Errorf("extract error: %s", respBody.Error)
	}

	return &respBody, nil
}

// HealthCheck 健康检查
func (h *HTTPBridge) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", h.baseURL+"/health", nil)
	if err != nil {
		return err
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: %d", resp.StatusCode)
	}

	return nil
}
