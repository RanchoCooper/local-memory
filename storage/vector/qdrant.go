package vector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// QdrantConfig holds Qdrant configuration.
type QdrantConfig struct {
	URL        string
	Collection string
	VectorSize int
}

// QdrantStore is the Qdrant vector store implementation (HTTP API client).
// Requires Qdrant service running at the configured URL.
type QdrantStore struct {
	url        string
	collection string
	vectorSize int
	client     *http.Client
	mu         sync.RWMutex
	vectors    map[string][]float32
	metadata   map[string]map[string]any
}

func NewQdrantStore(cfg any) (*QdrantStore, error) {
	c, ok := cfg.(QdrantConfig)
	if !ok {
		return nil, fmt.Errorf("invalid qdrant config")
	}

	store := &QdrantStore{
		url:        strings.TrimSuffix(c.URL, "/"),
		collection: c.Collection,
		vectorSize: c.VectorSize,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		vectors:  make(map[string][]float32),
		metadata: make(map[string]map[string]any),
	}

	ctx := context.Background()
	if err := store.ensureCollection(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure collection: %w", err)
	}

	return store, nil
}

func (s *QdrantStore) ensureCollection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", s.url+"/collections/"+s.collection, nil)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	createReq := map[string]any{
		"vectors": map[string]any{
			"size":     s.vectorSize,
			"distance": "Cosine",
		},
	}

	body, err := json.Marshal(createReq)
	if err != nil {
		return err
	}

	req, err = http.NewRequestWithContext(ctx, "PUT", s.url+"/collections/"+s.collection, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err = s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create collection: %s", resp.Status)
	}

	return nil
}

func (s *QdrantStore) Upsert(id string, vector []float32, metadata map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vecCopy := make([]float32, len(vector))
	copy(vecCopy, vector)
	s.vectors[id] = vecCopy

	metaCopy := make(map[string]any)
	for k, v := range metadata {
		metaCopy[k] = v
	}
	s.metadata[id] = metaCopy

	ctx := context.Background()
	payload := map[string]any{
		"points": []map[string]any{
			{
				"id":       id,
				"vector":   vector,
				"payload":  metadata,
			},
		},
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "PUT", s.url+"/collections/"+s.collection+"/points", strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("upsert failed: %s", resp.Status)
	}

	return nil
}

func (s *QdrantStore) Search(query []float32, topK int, filter *Filter) ([]Result, error) {
	ctx := context.Background()

	searchReq := map[string]any{
		"vector":       query,
		"limit":        topK,
		"with_vector":  false,
		"with_payload": true,
	}

	if filter != nil {
		conditions := []map[string]any{}
		if filter.Scope != "" {
			conditions = append(conditions, map[string]any{
				"key": "scope",
				"match": map[string]any{
					"keyword": filter.Scope,
				},
			})
		}
		if filter.Type != "" {
			conditions = append(conditions, map[string]any{
				"key": "type",
				"match": map[string]any{
					"keyword": filter.Type,
				},
			})
		}
		if filter.ProfileID != "" {
			conditions = append(conditions, map[string]any{
				"key": "profile_id",
				"match": map[string]any{
					"keyword": filter.ProfileID,
				},
			})
		}
		if len(conditions) > 0 {
			searchReq["filter"] = map[string]any{
				"must": conditions,
			}
		}
	}

	body, _ := json.Marshal(searchReq)
	req, err := http.NewRequestWithContext(ctx, "POST", s.url+"/collections/"+s.collection+"/points/search", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed: %s", resp.Status)
	}

	var results []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}

	var searchResults []Result
	for _, r := range results {
		id, _ := r["id"].(string)
		score, _ := r["score"].(float64)
		payload, _ := r["payload"].(map[string]any)

		searchResults = append(searchResults, Result{
			ID:       id,
			Score:    score,
			Metadata: payload,
		})
	}

	return searchResults, nil
}

func (s *QdrantStore) Delete(id string) error {
	ctx := context.Background()

	s.mu.Lock()
	delete(s.vectors, id)
	delete(s.metadata, id)
	s.mu.Unlock()

	deleteReq := map[string]any{
		"points": []string{id},
	}

	body, _ := json.Marshal(deleteReq)
	req, err := http.NewRequestWithContext(ctx, "POST", s.url+"/collections/"+s.collection+"/points/delete", strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete failed: %s", resp.Status)
	}

	return nil
}

func (s *QdrantStore) Close() error {
	return nil
}
