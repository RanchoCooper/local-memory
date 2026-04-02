package core

// Link 记忆关联模块
// 负责记忆之间的关联管理
type Link struct {
	sqliteStore SQLiteStoreInterface
}

// NewLink 创建 Link 实例
func NewLink(sqliteStore SQLiteStoreInterface) *Link {
	return &Link{
		sqliteStore: sqliteStore,
	}
}

// Link 建立两个记忆之间的关联
// 会自动建立双向关联
func (l *Link) Link(id1, id2 string) error {
	if id1 == id2 {
		return ErrCannotLinkToSelf
	}

	// 获取两个记忆
	m1, err := l.sqliteStore.GetByID(id1)
	if err != nil {
		return err
	}
	if m1 == nil {
		return ErrMemoryNotFound
	}

	m2, err := l.sqliteStore.GetByID(id2)
	if err != nil {
		return err
	}
	if m2 == nil {
		return ErrMemoryNotFound
	}

	// 添加关联（去重）
	m1.RelatedIDs = addIDIfNotExists(m1.RelatedIDs, id2)
	m2.RelatedIDs = addIDIfNotExists(m2.RelatedIDs, id1)

	// 保存更新
	if err := l.sqliteStore.Save(m1); err != nil {
		return err
	}
	return l.sqliteStore.Save(m2)
}

// Unlink 解除两个记忆之间的关联
// 会自动解除双向关联
func (l *Link) Unlink(id1, id2 string) error {
	// 获取两个记忆
	m1, err := l.sqliteStore.GetByID(id1)
	if err != nil {
		return err
	}
	if m1 == nil {
		return ErrMemoryNotFound
	}

	m2, err := l.sqliteStore.GetByID(id2)
	if err != nil {
		return err
	}
	if m2 == nil {
		return ErrMemoryNotFound
	}

	// 移除关联
	m1.RelatedIDs = removeID(m1.RelatedIDs, id2)
	m2.RelatedIDs = removeID(m2.RelatedIDs, id1)

	// 保存更新
	if err := l.sqliteStore.Save(m1); err != nil {
		return err
	}
	return l.sqliteStore.Save(m2)
}

// GetRelated 获取与指定记忆关联的所有记忆
// depth 指定遍历深度，1 表示直接关联，2 表示关联的关联
func (l *Link) GetRelated(id string, depth int) ([]*Memory, error) {
	if depth < 1 {
		depth = 1
	}

	visited := make(map[string]bool)
	var result []*Memory

	// BFS 遍历
	queue := []string{id}
	currentDepth := 0

	for currentDepth < depth && len(queue) > 0 {
		size := len(queue)
		for i := 0; i < size; i++ {
			currentID := queue[0]
			queue = queue[1:]

			if visited[currentID] {
				continue
			}
			visited[currentID] = true

			memory, err := l.sqliteStore.GetByID(currentID)
			if err != nil {
				continue
			}
			if memory == nil || memory.Deleted {
				continue
			}

			// 不包含起始节点
			if currentID != id {
				result = append(result, memory)
			}

			// 将关联记忆加入队列
			for _, relatedID := range memory.RelatedIDs {
				if !visited[relatedID] {
					queue = append(queue, relatedID)
				}
			}
		}
		currentDepth++
	}

	return result, nil
}

// GetRelatedIDs 获取指定记忆的直接关联 IDs
func (l *Link) GetRelatedIDs(id string) ([]string, error) {
	memory, err := l.sqliteStore.GetByID(id)
	if err != nil {
		return nil, err
	}
	if memory == nil {
		return nil, ErrMemoryNotFound
	}
	return memory.RelatedIDs, nil
}

// addIDIfNotExists 如果 ID 不存在则添加
func addIDIfNotExists(ids []string, newID string) []string {
	for _, id := range ids {
		if id == newID {
			return ids
		}
	}
	return append(ids, newID)
}

// removeID 从列表中移除 ID
func removeID(ids []string, targetID string) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		if id != targetID {
			result = append(result, id)
		}
	}
	return result
}
