package core

import (
	"time"
)

// Evolve handles memory evolution.
// Responsible for merging and updating memories with the same key.
type Evolve struct {
	store *Store
}

// NewEvolve creates an Evolve instance.
func NewEvolve(store *Store) *Evolve {
	return &Evolve{
		store: store,
	}
}

// MergeOption contains merge options.
type MergeOption struct {
	Strategy string // Merge strategy: append | replace | max
}

// DefaultMergeOption is the default merge option.
var DefaultMergeOption = &MergeOption{
	Strategy: "append",
}

// Merge merges two memories with the same key.
// 1. Merge value (append new information)
// 2. Update confidence (take higher value + increment)
// 3. Update timestamp
// 4. Merge tags and related memories
func (e *Evolve) Merge(existing, new *Memory, opts *MergeOption) (*Memory, error) {
	if opts == nil {
		opts = DefaultMergeOption
	}

	// Merge value
	mergedValue := e.mergeValue(existing.Value, new.Value, opts.Strategy)
	existing.Value = mergedValue

	// Update confidence: take higher value + 0.1 increment
	existing.Confidence = minFloat64(1.0, existing.Confidence+0.1)

	// Update timestamp
	existing.UpdatedAt = time.Now().Unix()

	// Merge tags
	existing.Tags = mergeTags(existing.Tags, new.Tags)

	// Merge related memories (deduplicate)
	existing.RelatedIDs = mergeIDs(existing.RelatedIDs, new.RelatedIDs)

	// Keep better metadata
	if new.Metadata.Source != "" {
		existing.Metadata.Source = new.Metadata.Source
	}

	return existing, nil
}

// mergeValue merges memory values.
func (e *Evolve) mergeValue(existing, new, strategy string) string {
	switch strategy {
	case "replace":
		// New value overwrites old value
		return new
	case "max":
		// Take the longer value
		if len(new) > len(existing) {
			return new
		}
		return existing
	case "append":
		// Append new value
		if existing == "" {
			return new
		}
		if new == "" {
			return existing
		}
		return existing + "\n" + new
	default:
		return new
	}
}

// EvolveExisting checks and evolves an existing memory.
// If a memory with the same key exists, merge them.
func (e *Evolve) EvolveExisting(memory *Memory) (*Memory, bool, error) {
	// Ensure profile_id is set before lookup
	if memory.ProfileID == "" {
		memory.ProfileID = "default"
	}

	existing, err := e.store.sqliteStore.GetByKey(memory.Key, memory.ProfileID)
	if err != nil {
		return nil, false, err
	}

	if existing == nil {
		return memory, false, nil
	}

	// Merge memories
	merged, err := e.Merge(existing, memory, nil)
	if err != nil {
		return nil, false, err
	}

	return merged, true, nil
}
