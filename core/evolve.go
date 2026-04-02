package core

import (
	"time"
)

// Evolve 记忆进化模块
// 负责同 key 记忆的合并和更新
type Evolve struct {
	store *Store
}

// NewEvolve 创建 Evolve 实例
func NewEvolve(store *Store) *Evolve {
	return &Evolve{
		store: store,
	}
}

// MergeOption 合并选项
type MergeOption struct {
	Strategy string // 合并策略：append | replace | max
}

// DefaultMergeOption 默认合并选项
var DefaultMergeOption = &MergeOption{
	Strategy: "append",
}

// Merge 合并两个同 key 记忆
// 1. 合并 value（追加新信息）
// 2. 更新 confidence（取较高值 + 增量）
// 3. 更新时间戳
// 4. 合并标签和关联记忆
func (e *Evolve) Merge(existing, new *Memory, opts *MergeOption) (*Memory, error) {
	if opts == nil {
		opts = DefaultMergeOption
	}

	// 合并 value
	mergedValue := e.mergeValue(existing.Value, new.Value, opts.Strategy)
	existing.Value = mergedValue

	// 更新置信度：取较高值 + 0.1 增量
	existing.Confidence = minFloat64(1.0, existing.Confidence+0.1)

	// 更新时间戳
	existing.UpdatedAt = time.Now().Unix()

	// 合并标签
	existing.Tags = mergeTags(existing.Tags, new.Tags)

	// 合并关联记忆（去重）
	existing.RelatedIDs = mergeIDs(existing.RelatedIDs, new.RelatedIDs)

	// 保留更好的 metadata
	if new.Metadata.Source != "" {
		existing.Metadata.Source = new.Metadata.Source
	}

	return existing, nil
}

// mergeValue 合并记忆值
func (e *Evolve) mergeValue(existing, new, strategy string) string {
	switch strategy {
	case "replace":
		// 新值覆盖旧值
		return new
	case "max":
		// 取较长的值
		if len(new) > len(existing) {
			return new
		}
		return existing
	case "append":
		// 追加新值
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

// EvolveExisting 检查并进化已有记忆
// 如果存在同 key 记忆，则合并
func (e *Evolve) EvolveExisting(memory *Memory) (*Memory, bool, error) {
	existing, err := e.store.sqliteStore.GetByKey(memory.Key)
	if err != nil {
		return nil, false, err
	}

	if existing == nil {
		return memory, false, nil
	}

	// 合并记忆
	merged, err := e.Merge(existing, memory, nil)
	if err != nil {
		return nil, false, err
	}

	return merged, true, nil
}
