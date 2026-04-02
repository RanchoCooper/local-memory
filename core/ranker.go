package core

import (
	"math"
	"sort"
	"time"
)

// Ranker handles memory ranking.
// Responsible for calculating memory relevance scores based on multiple factors.
type Ranker struct {
	decayLambda float64 // Decay coefficient
}

// RankingConfig holds ranking configuration.
type RankingConfig struct {
	SimilarityWeight float64 // Similarity weight
	RecencyWeight    float64 // Recency weight
	ConfidenceWeight float64 // Confidence weight
	MaxAgeSeconds    int64   // Maximum age span (seconds)
}

// NewRanker creates a Ranker instance.
func NewRanker(decayLambda float64) *Ranker {
	return &Ranker{
		decayLambda: decayLambda,
	}
}

// DefaultRankingConfig is the default ranking configuration.
// Weights: similarity 70%, recency 20%, confidence 10%
var DefaultRankingConfig = &RankingConfig{
	SimilarityWeight: 0.7,
	RecencyWeight:    0.2,
	ConfidenceWeight: 0.1,
	MaxAgeSeconds:    86400 * 30, // 30 days
}

// CalcScore calculates the composite score for a memory.
// score = similarity * 0.7 + recency * 0.2 + confidence * 0.1
func (r *Ranker) CalcScore(similarity float64, memory *Memory) float64 {
	recency := r.calcRecency(memory.CreatedAt)
	confidence := memory.Confidence

	return similarity*DefaultRankingConfig.SimilarityWeight +
		recency*DefaultRankingConfig.RecencyWeight +
		confidence*DefaultRankingConfig.ConfidenceWeight
}

// calcRecency calculates the recency score.
// Uses exponential decay: weight = e^(-λ * Δt)
// Δt: time difference (seconds)
// λ: decay coefficient
func (r *Ranker) calcRecency(createdAt int64) float64 {
	delta := time.Now().Unix() - createdAt
	weight := math.Exp(-r.decayLambda * float64(delta))
	return weight
}

// ScoreSort sorts search results by score in descending order.
func (r *Ranker) ScoreSort(results []*QueryResult) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
}
