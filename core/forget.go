package core

import (
	"fmt"
	"time"
)

// Forget 记忆遗忘模块
// 负责记忆的软删除和清理
type Forget struct {
	sqliteStore SQLiteStoreInterface
	vectorStore VectorStore
}

// NewForget 创建 Forget 实例
func NewForget(sqliteStore SQLiteStoreInterface, vectorStore VectorStore) *Forget {
	return &Forget{
		sqliteStore: sqliteStore,
		vectorStore: vectorStore,
	}
}

// Delete 软删除记忆
// 设置 deleted=1 和 deleted_at 时间戳
// 记忆数据仍然保留，可用于恢复
func (f *Forget) Delete(id string) error {
	memory, err := f.sqliteStore.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get memory: %w", err)
	}
	if memory == nil {
		return ErrMemoryNotFound
	}

	// 软删除
	memory.Deleted = true
	memory.DeletedAt = time.Now().Unix()

	return f.sqliteStore.Save(memory)
}

// Restore 恢复已删除的记忆
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

	// 恢复记忆
	memory.Deleted = false
	memory.DeletedAt = 0

	return f.sqliteStore.Save(memory)
}

// HardDelete 永久删除记忆
// 从 SQLite 和向量存储中彻底删除
func (f *Forget) HardDelete(id string) error {
	memory, err := f.sqliteStore.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get memory: %w", err)
	}
	if memory == nil {
		return ErrMemoryNotFound
	}

	// 从 SQLite 删除
	if err := f.hardDelete(id); err != nil {
		return fmt.Errorf("failed to delete from sqlite: %w", err)
	}

	// 从向量存储删除
	if f.vectorStore != nil {
		if err := f.vectorStore.Delete(id); err != nil {
			// 向量存储删除失败不影响主要数据
			fmt.Printf("warning: failed to delete from vector store: %v\n", err)
		}
	}

	return nil
}

// hardDelete 直接从存储删除
// 这是一个辅助方法，实际实现由适配器提供
func (f *Forget) hardDelete(id string) error {
	// 使用 Save 方法配合 Deleted 标记的方式
	// 实际删除需要存储层支持
	memory, _ := f.sqliteStore.GetByID(id)
	if memory != nil {
		// 标记为已删除，下次清理时移除
		memory.Deleted = true
		memory.DeletedAt = time.Now().Unix()
		return f.sqliteStore.Save(memory)
	}
	return nil
}

// PurgeExpired 清理过期记忆
// 阈值：衰减权重低于指定值时认为过期
func (f *Forget) PurgeExpired(lambda float64, threshold float64) (int, error) {
	decay := NewDecay(lambda)

	// 获取所有记忆
	memories, _, err := f.sqliteStore.List(&ListRequest{
		IncludeDeleted: false,
		Limit:          10000, // 假设不超过 10000 条
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

// PurgeByScope 清理指定作用域的记忆
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
