package unit

import (
	"testing"
	"time"

	"localmemory/core"
)

func TestRanker_CalcScore(t *testing.T) {
	ranker := core.NewRanker(0.01)

	tests := []struct {
		name       string
		similarity float64
		memory     *core.Memory
		wantMin    float64
		wantMax    float64
	}{
		{
			name:       "high similarity and confidence",
			similarity: 0.9,
			memory: &core.Memory{
				Confidence: 0.9,
				CreatedAt:  time.Now().Unix(),
			},
			wantMin: 0.63, // 0.9*0.7 + recency*0.2 + 0.9*0.1
			wantMax: 1.0,
		},
		{
			name:       "low similarity",
			similarity: 0.1,
			memory: &core.Memory{
				Confidence: 0.5,
				CreatedAt:  time.Now().Unix(),
			},
			wantMin: 0.07,
			wantMax: 0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := ranker.CalcScore(tt.similarity, tt.memory)
			if score < tt.wantMin || score > tt.wantMax {
				t.Errorf("CalcScore() = %v, want between %v and %v", score, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestRanker_ScoreSort(t *testing.T) {
	ranker := core.NewRanker(0.01)

	results := []*core.QueryResult{
		{Memory: &core.Memory{}, Score: 0.3},
		{Memory: &core.Memory{}, Score: 0.8},
		{Memory: &core.Memory{}, Score: 0.5},
		{Memory: &core.Memory{}, Score: 0.1},
	}

	ranker.ScoreSort(results)

	// After sorting, scores should be descending
	for i := 0; i < len(results)-1; i++ {
		if results[i].Score < results[i+1].Score {
			t.Errorf("Expected results[%d].Score (%v) >= results[%d].Score (%v)",
				i, results[i].Score, i+1, results[i+1].Score)
		}
	}
}

func TestRanker_DefaultConfig(t *testing.T) {
	cfg := core.DefaultRankingConfig

	if cfg.SimilarityWeight != 0.7 {
		t.Errorf("Expected SimilarityWeight 0.7, got %f", cfg.SimilarityWeight)
	}
	if cfg.RecencyWeight != 0.2 {
		t.Errorf("Expected RecencyWeight 0.2, got %f", cfg.RecencyWeight)
	}
	if cfg.ConfidenceWeight != 0.1 {
		t.Errorf("Expected ConfidenceWeight 0.1, got %f", cfg.ConfidenceWeight)
	}
}
