package unit

import (
	"math"
	"testing"
	"time"

	"localmemory/core"
)

func TestDecay_NewDecay(t *testing.T) {
	decay := core.NewDecay(0.01)

	if decay == nil {
		t.Fatal("NewDecay returned nil")
	}
}

func TestDecay_Calculate(t *testing.T) {
	tests := []struct {
		name      string
		lambda    float64
		createdAt int64
		wantMin   float64
		wantMax   float64
	}{
		{
			name:      "current memory high weight",
			lambda:    0.01,
			createdAt: time.Now().Unix(),
			wantMin:   0.9,
			wantMax:   1.0,
		},
		{
			name:      "1 hour ago very low weight",
			lambda:    0.01,
			createdAt: time.Now().Add(-1 * time.Hour).Unix(),
			wantMin:   0.0,
			wantMax:   0.1,
		},
		{
			name:      "1 day ago essentially zero",
			lambda:    0.01,
			createdAt: time.Now().Add(-24 * time.Hour).Unix(),
			wantMin:   0.0,
			wantMax:   0.01,
		},
		{
			name:      "1 week ago zero",
			lambda:    0.01,
			createdAt: time.Now().Add(-7 * 24 * time.Hour).Unix(),
			wantMin:   0.0,
			wantMax:   0.001,
		},
		{
			name:      "zero lambda no decay",
			lambda:    0.0,
			createdAt: time.Now().Add(-100 * 24 * time.Hour).Unix(),
			wantMin:   0.99,
			wantMax:   1.0,
		},
		{
			name:      "high lambda fast decay",
			lambda:    0.1,
			createdAt: time.Now().Add(-1 * time.Hour).Unix(),
			wantMin:   0.0,
			wantMax:   0.01,
		},
		{
			name:      "small lambda slow decay",
			lambda:    0.001,
			createdAt: time.Now().Add(-1 * time.Hour).Unix(),
			wantMin:   0.0,
			wantMax:   1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decay := core.NewDecay(tt.lambda)
			got := decay.Calculate(tt.createdAt)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("Calculate() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestDecay_CalculateWithFactor(t *testing.T) {
	decay := core.NewDecay(0.01)

	// Use a memory created 1 hour ago so delta is significant
	createdAt := time.Now().Add(-1 * time.Hour).Unix()

	// Without factor (factor = 1.0)
	weight1 := decay.CalculateWithFactor(createdAt, 1.0)

	// With factor 2.0 (should decay faster, lower weight)
	weight2 := decay.CalculateWithFactor(createdAt, 2.0)

	// With factor 0.5 (should decay slower, higher weight)
	weight3 := decay.CalculateWithFactor(createdAt, 0.5)

	// weight2 should be less than weight1 (faster decay)
	if weight2 >= weight1 {
		t.Errorf("Expected weight with factor 2.0 (%v) < weight with factor 1.0 (%v)", weight2, weight1)
	}

	// weight3 should be greater than weight1 (slower decay)
	if weight3 <= weight1 {
		t.Errorf("Expected weight with factor 0.5 (%v) > weight with factor 1.0 (%v)", weight3, weight1)
	}
}

func TestDecay_ApplyDecay(t *testing.T) {
	decay := core.NewDecay(0.01)

	memory := &core.Memory{
		Confidence: 1.0,
		CreatedAt:  time.Now().Unix(),
	}

	newConfidence := decay.ApplyDecay(memory)

	// New confidence should be less than or equal to original
	if newConfidence > memory.Confidence {
		t.Error("Decayed confidence should be less than or equal to original")
	}

	// New confidence should be non-negative
	if newConfidence < 0 {
		t.Error("Confidence should not be negative")
	}
}

func TestDecay_ApplyDecay_ZeroConfidence(t *testing.T) {
	decay := core.NewDecay(0.01)

	memory := &core.Memory{
		Confidence: 0.0,
		CreatedAt:  time.Now().Unix(),
	}

	newConfidence := decay.ApplyDecay(memory)

	if newConfidence != 0.0 {
		t.Errorf("Expected 0.0, got %f", newConfidence)
	}
}

func TestDecay_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		lambda    float64
		createdAt int64
		threshold float64
		want      bool
	}{
		{
			name:      "recent memory not expired",
			lambda:    0.01,
			createdAt: time.Now().Unix(),
			threshold: 0.1,
			want:      false,
		},
		{
			name:      "very old memory expired",
			lambda:    0.01,
			createdAt: time.Now().Add(-30 * 24 * time.Hour).Unix(),
			threshold: 0.1,
			want:      true,
		},
		{
			name:      "at threshold or above not expired",
			lambda:    0.0,
			createdAt: time.Now().Unix(),
			threshold: 1.0,
			want:      false,
		},
		{
			name:      "below threshold expired",
			lambda:    0.01,
			createdAt: time.Now().Add(-24 * time.Hour).Unix(),
			threshold: 0.001,
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decay := core.NewDecay(tt.lambda)
			got := decay.IsExpired(tt.createdAt, tt.threshold)
			if got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecay_ExponentialDecayFormula(t *testing.T) {
	// Test that the decay formula is correct: weight = e^(-lambda * delta)
	decay := core.NewDecay(0.01)

	now := time.Now().Unix()
	oneHourAgo := now - 3600 // 1 hour in seconds

	calculated := decay.Calculate(oneHourAgo)
	expected := math.Exp(-0.01 * 3600)

	// Allow small floating point difference
	diff := math.Abs(calculated - expected)
	if diff > 0.0001 {
		t.Errorf("Calculate() = %v, expected formula result %v (diff %v)", calculated, expected, diff)
	}
}

func TestDecay_NormalizeAge(t *testing.T) {
	// This tests the decay based on age normalization
	decay := core.NewDecay(0.01)

	// Very recent memory should have high normalized age (close to 1)
	recent := time.Now().Unix()
	weight := decay.Calculate(recent)
	if weight < 0.9 {
		t.Errorf("Expected recent memory to have weight > 0.9, got %f", weight)
	}

	// Very old memory should have low normalized age (close to 0)
	old := time.Now().Add(-30 * 24 * time.Hour).Unix()
	weightOld := decay.Calculate(old)
	if weightOld > 0.1 {
		t.Errorf("Expected old memory to have weight < 0.1, got %f", weightOld)
	}
}
