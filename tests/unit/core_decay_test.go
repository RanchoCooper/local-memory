package unit

import (
	"testing"
	"time"

	"localmemory/core"
)

func TestDecay_Calculate(t *testing.T) {
	decay := core.NewDecay(0.01)

	// Test case 1: Current memory should have high weight
	currentMemory := time.Now().Unix()
	weight := decay.Calculate(currentMemory)
	if weight < 0.9 {
		t.Errorf("Expected weight > 0.9 for recent memory, got %f", weight)
	}
	if weight > 1.0 {
		t.Errorf("Expected weight <= 1.0, got %f", weight)
	}

	// Test case 2: Old memory should have low weight
	oldMemory := time.Now().Add(-24 * time.Hour).Unix()
	weightOld := decay.Calculate(oldMemory)
	if weightOld > weight {
		t.Errorf("Expected older memory to have lower weight")
	}
}

func TestDecay_IsExpired(t *testing.T) {
	decay := core.NewDecay(0.01)

	// Test case 1: Recent memory should not be expired
	recent := time.Now().Unix()
	if decay.IsExpired(recent, 0.1) {
		t.Error("Recent memory should not be expired")
	}

	// Test case 2: Very old memory should be expired
	veryOld := time.Now().Add(-30 * 24 * time.Hour).Unix()
	if !decay.IsExpired(veryOld, 0.1) {
		t.Error("Old memory should be expired")
	}
}

func TestDecay_ApplyDecay(t *testing.T) {
	decay := core.NewDecay(0.01)

	memory := &core.Memory{
		Confidence: 1.0,
		CreatedAt:  time.Now().Unix(),
	}

	newConfidence := decay.ApplyDecay(memory)
	if newConfidence > memory.Confidence {
		t.Error("Decayed confidence should be less than or equal to original")
	}
	if newConfidence < 0 {
		t.Error("Confidence should not be negative")
	}
}
