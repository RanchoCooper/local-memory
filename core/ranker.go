package core

import (
	"math"
	"sort"
	"time"
)

// Ranker 记忆排序模块
// 负责根据多维度因素计算记忆相关性得分
type Ranker struct {
	decayLambda float64 // 衰减系数
}

// RankingConfig 排序配置
type RankingConfig struct {
	SimilarityWeight float64 // 相似度权重
	RecencyWeight    float64 // 时效性权重
	ConfidenceWeight float64 // 置信度权重
	MaxAgeSeconds    int64   // 最大时间跨度（秒）
}

// NewRanker 创建 Ranker 实例
func NewRanker(decayLambda float64) *Ranker {
	return &Ranker{
		decayLambda: decayLambda,
	}
}

// DefaultRankingConfig 默认排序配置
// 权重分配：相似度 70%，时效性 20%，置信度 10%
var DefaultRankingConfig = &RankingConfig{
	SimilarityWeight: 0.7,
	RecencyWeight:    0.2,
	ConfidenceWeight: 0.1,
	MaxAgeSeconds:    86400 * 30, // 30 天
}

// CalcScore 计算记忆的综合得分
// score = similarity * 0.7 + recency * 0.2 + confidence * 0.1
func (r *Ranker) CalcScore(similarity float64, memory *Memory) float64 {
	recency := r.calcRecency(memory.CreatedAt)
	confidence := memory.Confidence

	return similarity*DefaultRankingConfig.SimilarityWeight +
		recency*DefaultRankingConfig.RecencyWeight +
		confidence*DefaultRankingConfig.ConfidenceWeight
}

// calcRecency 计算时效性得分
// 使用指数衰减：weight = e^(-λ * Δt)
// Δt：时间差（秒）
// λ：衰减系数
func (r *Ranker) calcRecency(createdAt int64) float64 {
	delta := time.Now().Unix() - createdAt
	weight := math.Exp(-r.decayLambda * float64(delta))
	return weight
}

// ScoreSort 对检索结果按得分降序排序
func (r *Ranker) ScoreSort(results []*QueryResult) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
}
