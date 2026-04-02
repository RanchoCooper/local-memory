package core

// Link handles memory associations.
// Responsible for managing relationships between memories.
type Link struct {
	sqliteStore SQLiteStoreInterface
}

// NewLink creates a Link instance.
func NewLink(sqliteStore SQLiteStoreInterface) *Link {
	return &Link{
		sqliteStore: sqliteStore,
	}
}

// Link creates an association between two memories.
// Automatically creates bidirectional links.
func (l *Link) Link(id1, id2 string) error {
	if id1 == id2 {
		return ErrCannotLinkToSelf
	}

	// Get both memories
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

	// Add association (deduplicate)
	m1.RelatedIDs = addIDIfNotExists(m1.RelatedIDs, id2)
	m2.RelatedIDs = addIDIfNotExists(m2.RelatedIDs, id1)

	// Save updates
	if err := l.sqliteStore.Save(m1); err != nil {
		return err
	}
	return l.sqliteStore.Save(m2)
}

// Unlink removes the association between two memories.
// Automatically removes bidirectional links.
func (l *Link) Unlink(id1, id2 string) error {
	// Get both memories
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

	// Remove association
	m1.RelatedIDs = removeID(m1.RelatedIDs, id2)
	m2.RelatedIDs = removeID(m2.RelatedIDs, id1)

	// Save updates
	if err := l.sqliteStore.Save(m1); err != nil {
		return err
	}
	return l.sqliteStore.Save(m2)
}

// GetRelated gets all memories related to the specified memory.
// depth specifies traversal depth, 1 means direct association, 2 means association's association.
func (l *Link) GetRelated(id string, depth int) ([]*Memory, error) {
	if depth < 1 {
		depth = 1
	}

	visited := make(map[string]bool)
	var result []*Memory

	// BFS traversal
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

			// Exclude starting node
			if currentID != id {
				result = append(result, memory)
			}

			// Add related memories to queue
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

// GetRelatedIDs gets direct related IDs for the specified memory.
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

// addIDIfNotExists adds ID if it doesn't already exist.
func addIDIfNotExists(ids []string, newID string) []string {
	for _, id := range ids {
		if id == newID {
			return ids
		}
	}
	return append(ids, newID)
}

// removeID removes ID from the list.
func removeID(ids []string, targetID string) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		if id != targetID {
			result = append(result, id)
		}
	}
	return result
}
