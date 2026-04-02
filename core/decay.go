package core

import (
	"math"
	"time"
)

// Decay handles memory decay over time.
// Responsible for calculating and applying temporal decay to memories.
type Decay struct {
	lambda float64 // Decay coefficient
}

// NewDecay creates a Decay instance.
func NewDecay(lambda float64) *Decay {
	return &Decay{
		lambda: lambda,
	}
}

// Calculate calculates the decay weight of a memory.
// Formula: weight = e^(-λ * Δt)
// Δt: time difference between now and creation (seconds)
// λ: decay coefficient, larger value means faster decay
//
// Examples:
//   - λ = 0.01, Δt = 1 hour (3600s) → weight ≈ 0.97
//   - λ = 0.01, Δt = 1 day (86400s) → weight ≈ 0.42
//   - λ = 0.01, Δt = 7 days → weight ≈ 0.00045
func (d *Decay) Calculate(createdAt int64) float64 {
	delta := time.Now().Unix() - createdAt
	return math.Exp(-d.lambda * float64(delta))
}

// CalculateWithFactor calculates decay with a custom factor.
func (d *Decay) CalculateWithFactor(createdAt int64, factor float64) float64 {
	delta := time.Now().Unix() - createdAt
	return math.Exp(-d.lambda * factor * float64(delta))
}

// ApplyDecay applies decay to a memory's confidence.
// Returns the new confidence value.
func (d *Decay) ApplyDecay(memory *Memory) float64 {
	decay := d.Calculate(memory.CreatedAt)
	return memory.Confidence * decay
}

// IsExpired checks if a memory has expired.
// Memory is considered expired when decay weight is below threshold.
func (d *Decay) IsExpired(createdAt int64, threshold float64) bool {
	weight := d.Calculate(createdAt)
	return weight < threshold
}
