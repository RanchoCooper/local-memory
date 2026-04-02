package core

import (
	"github.com/google/uuid"
)

// GenerateID generates a new UUID v4 string.
// Used as the unique identifier for Memory.
// UUID v4 uses random numbers with sufficiently low collision probability.
func GenerateID() string {
	return uuid.New().String()
}
