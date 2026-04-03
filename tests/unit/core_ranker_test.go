package unit

import (
	"testing"
	"time"

	"localmemory/core"
)

func TestRanker_NewRanker(t *testing.T) {
	ranker := core.NewRanker(0.01)

	if ranker == nil {
		t.Fatal("NewRanker returned nil")
	}
}

func TestRanker_CalcScore(t *testing.T) {
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
			wantMin: 0.63,
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
		{
			name:       "zero similarity",
			similarity: 0.0,
			memory: &core.Memory{
				Confidence: 1.0,
				CreatedAt:  time.Now().Unix(),
			},
			wantMin: 0.0,
			wantMax: 0.5,
		},
		{
			name:       "perfect similarity",
			similarity: 1.0,
			memory: &core.Memory{
				Confidence: 1.0,
				CreatedAt:  time.Now().Unix(),
			},
			wantMin: 0.8,
			wantMax: 1.0,
		},
		{
			name:       "old memory with high similarity",
			similarity: 0.9,
			memory: &core.Memory{
				Confidence: 0.9,
				CreatedAt:  time.Now().Add(-7 * 24 * time.Hour).Unix(), // 7 days old
			},
			wantMin: 0.5,
			wantMax: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ranker := core.NewRanker(0.01)
			score := ranker.CalcScore(tt.similarity, tt.memory)
			if score < tt.wantMin || score > tt.wantMax {
				t.Errorf("CalcScore() = %v, want between %v and %v", score, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestRanker_ScoreSort(t *testing.T) {
	tests := []struct {
		name    string
		results []*core.QueryResult
	}{
		{
			name: "unsorted results",
			results: []*core.QueryResult{
				{Score: 0.3},
				{Score: 0.8},
				{Score: 0.5},
				{Score: 0.1},
			},
		},
		{
			name: "already sorted",
			results: []*core.QueryResult{
				{Score: 0.9},
				{Score: 0.7},
				{Score: 0.5},
				{Score: 0.3},
			},
		},
		{
			name: "reverse sorted",
			results: []*core.QueryResult{
				{Score: 0.1},
				{Score: 0.3},
				{Score: 0.5},
				{Score: 0.7},
			},
		},
		{
			name: "single element",
			results: []*core.QueryResult{
				{Score: 0.5},
			},
		},
		{
			name: "empty list",
			results: []*core.QueryResult{},
		},
		{
			name: "two elements",
			results: []*core.QueryResult{
				{Score: 0.3},
				{Score: 0.8},
			},
		},
		{
			name: "all equal scores",
			results: []*core.QueryResult{
				{Score: 0.5},
				{Score: 0.5},
				{Score: 0.5},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ranker := core.NewRanker(0.01)
			ranker.ScoreSort(tt.results)

			// Verify descending order
			for i := 0; i < len(tt.results)-1; i++ {
				if tt.results[i].Score < tt.results[i+1].Score {
					t.Errorf("Expected results[%d].Score (%v) >= results[%d].Score (%v)",
						i, tt.results[i].Score, i+1, tt.results[i+1].Score)
				}
			}
		})
	}
}

func TestRanker_DefaultConfig(t *testing.T) {
	cfg := core.DefaultRankingConfig

	if cfg == nil {
		t.Fatal("DefaultRankingConfig is nil")
	}

	if cfg.SimilarityWeight != 0.7 {
		t.Errorf("SimilarityWeight = %v, want 0.7", cfg.SimilarityWeight)
	}
	if cfg.RecencyWeight != 0.2 {
		t.Errorf("RecencyWeight = %v, want 0.2", cfg.RecencyWeight)
	}
	if cfg.ConfidenceWeight != 0.1 {
		t.Errorf("ConfidenceWeight = %v, want 0.1", cfg.ConfidenceWeight)
	}
	if cfg.MaxAgeSeconds != 86400*30 {
		t.Errorf("MaxAgeSeconds = %v, want 86400*30", cfg.MaxAgeSeconds)
	}
}

func TestRanker_RecencyWeight(t *testing.T) {
	// Test that recent memories get higher recency scores
	ranker := core.NewRanker(0.01)

	recentMemory := &core.Memory{
		Confidence: 0.5,
		CreatedAt:  time.Now().Unix(),
	}

	oldMemory := &core.Memory{
		Confidence: 0.5,
		CreatedAt:  time.Now().Add(-7 * 24 * time.Hour).Unix(),
	}

	recentScore := ranker.CalcScore(0.5, recentMemory)
	oldScore := ranker.CalcScore(0.5, oldMemory)

	// Recent should score higher due to recency weight
	if recentScore <= oldScore {
		t.Errorf("Expected recent score (%v) > old score (%v)", recentScore, oldScore)
	}
}

func TestRanker_ScoreRange(t *testing.T) {
	ranker := core.NewRanker(0.01)

	// Perfect memory
	perfect := &core.Memory{
		Confidence: 1.0,
		CreatedAt:  time.Now().Unix(),
	}
	score := ranker.CalcScore(1.0, perfect)

	// Score should be at most 1.0
	if score > 1.0 {
		t.Errorf("Score should be <= 1.0, got %v", score)
	}

	// Score should be non-negative
	if score < 0 {
		t.Errorf("Score should be >= 0, got %v", score)
	}
}

func TestRanker_ScoreComponents(t *testing.T) {
	ranker := core.NewRanker(0.01)

	memory := &core.Memory{
		Confidence: 0.5,
		CreatedAt:  time.Now().Unix(),
	}

	// With 0 similarity, score comes only from recency (20%) + confidence (10%)
	scoreZeroSim := ranker.CalcScore(0.0, memory)

	// With 1.0 similarity, score comes from similarity (70%) + recency (20%) + confidence (10%)
	scoreFullSim := ranker.CalcScore(1.0, memory)

	// The difference should be approximately 0.7 (similarity weight)
	diff := scoreFullSim - scoreZeroSim
	if diff < 0.6 || diff > 0.8 {
		t.Errorf("Expected similarity contribution around 0.7, got %v", diff)
	}
}
