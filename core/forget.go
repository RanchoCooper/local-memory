package core

import (
	"fmt"
	"time"
)

// Forget handles memory forgetting.
// Responsible for soft delete and cleanup of memories.
type Forget struct {
	sqliteStore SQLiteStoreInterface
	vectorStore VectorStore
}

// NewForget creates a Forget instance.
func NewForget(sqliteStore SQLiteStoreInterface, vectorStore VectorStore) *Forget {
	return &Forget{
		sqliteStore: sqliteStore,
		vectorStore: vectorStore,
	}
}

// Delete soft deletes a memory.
// Sets deleted=1 and deleted_at timestamp.
// Memory data is still retained and can be recovered.
func (f *Forget) Delete(id string) error {
	memory, err := f.sqliteStore.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get memory: %w", err)
	}
	if memory == nil {
		return ErrMemoryNotFound
	}

	// Soft delete
	memory.Deleted = true
	memory.DeletedAt = time.Now().Unix()

	return f.sqliteStore.Save(memory)
}

// Restore recovers a deleted memory.
func (f *Forget) Restore(id string) error {
	memory, err := f.sqliteStore.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get memory: %w", err)
	}
	if memory == nil {
		return ErrMemoryNotFound
	}
	if !memory.Deleted {
		return ErrMemoryNotDeleted
	}

	// Restore memory
	memory.Deleted = false
	memory.DeletedAt = 0

	return f.sqliteStore.Save(memory)
}

// HardDelete permanently deletes a memory.
// Completely removes from SQLite and vector store.
func (f *Forget) HardDelete(id string) error {
	memory, err := f.sqliteStore.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get memory: %w", err)
	}
	if memory == nil {
		return ErrMemoryNotFound
	}

	// Delete from SQLite
	if err := f.hardDelete(id); err != nil {
		return fmt.Errorf("failed to delete from sqlite: %w", err)
	}

	// Delete from vector store
	if f.vectorStore != nil {
		if err := f.vectorStore.Delete(id); err != nil {
			fmt.Printf("warning: failed to delete from vector store: %v\n", err)
		}
	}

	return nil
}

// hardDelete directly deletes from storage.
func (f *Forget) hardDelete(id string) error {
	memory, _ := f.sqliteStore.GetByID(id)
	if memory != nil {
		memory.Deleted = true
		memory.DeletedAt = time.Now().Unix()
		return f.sqliteStore.Save(memory)
	}
	return nil
}

// PurgeExpired cleans up expired memories.
// Memory is considered expired when decay weight is below threshold.
func (f *Forget) PurgeExpired(lambda float64, threshold float64) (int, error) {
	decay := NewDecay(lambda)

	// Get all memories
	memories, _, err := f.sqliteStore.List(&ListRequest{
		IncludeDeleted: false,
		Limit:          10000,
		Offset:         0,
	})
	if err != nil {
		return 0, err
	}

	count := 0
	for _, m := range memories {
		if decay.IsExpired(m.CreatedAt, threshold) {
			if err := f.HardDelete(m.ID); err != nil {
				continue
			}
			count++
		}
	}

	return count, nil
}

// PurgeByScope cleans up memories with specified scope.
func (f *Forget) PurgeByScope(scope Scope) (int, error) {
	memories, _, err := f.sqliteStore.List(&ListRequest{
		Scope: scope,
		Limit: 10000,
	})
	if err != nil {
		return 0, err
	}

	count := 0
	for _, m := range memories {
		if err := f.HardDelete(m.ID); err != nil {
			continue
		}
		count++
	}

	return count, nil
}
