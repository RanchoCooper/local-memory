package core

import (
	"math"
	"time"
)

// Decay 记忆衰减模块
// 负责计算和应用记忆的时间衰减
type Decay struct {
	lambda float64 // 衰减系数
}

// NewDecay 创建 Decay 实例
func NewDecay(lambda float64) *Decay {
	return &Decay{
		lambda: lambda,
	}
}

// Calculate 计算记忆的衰减权重
// 公式：weight = e^(-λ * Δt)
// Δt：当前时间与创建时间的差（秒）
// λ：衰减系数，值越大衰减越快
//
// 示例：
//   - λ = 0.01, Δt = 1小时(3600s) → weight ≈ 0.97
//   - λ = 0.01, Δt = 1天(86400s) → weight ≈ 0.42
//   - λ = 0.01, Δt = 7天 → weight ≈ 0.00045
func (d *Decay) Calculate(createdAt int64) float64 {
	delta := time.Now().Unix() - createdAt
	return math.Exp(-d.lambda * float64(delta))
}

// CalculateWithFactor 使用自定义因子计算衰减
// factor 用于调整衰减速度
func (d *Decay) CalculateWithFactor(createdAt int64, factor float64) float64 {
	delta := time.Now().Unix() - createdAt
	return math.Exp(-d.lambda * factor * float64(delta))
}

// ApplyDecay 将衰减应用到记忆的置信度
// 返回新的置信度
func (d *Decay) ApplyDecay(memory *Memory) float64 {
	decay := d.Calculate(memory.CreatedAt)
	return memory.Confidence * decay
}

// IsExpired 判断记忆是否已过期
// 阈值：衰减权重低于 0.1 时认为过期
func (d *Decay) IsExpired(createdAt int64, threshold float64) bool {
	weight := d.Calculate(createdAt)
	return weight < threshold
}
