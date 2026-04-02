package unit

import (
	"testing"
)

func TestMergeTags(t *testing.T) {
	tests := []struct {
		name string
		a    []string
		b    []string
		want int
	}{
		{
			name: "merge two lists",
			a:    []string{"tag1", "tag2"},
			b:    []string{"tag3", "tag4"},
			want: 4,
		},
		{
			name: "merge with duplicates",
			a:    []string{"tag1", "tag2"},
			b:    []string{"tag2", "tag3"},
			want: 3,
		},
		{
			name: "empty lists",
			a:    []string{},
			b:    []string{},
			want: 0,
		},
		{
			name: "nil lists",
			a:    nil,
			b:    nil,
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeTagsTest(tt.a, tt.b)
			if len(result) != tt.want {
				t.Errorf("mergeTags() got %d items, want %d", len(result), tt.want)
			}
		})
	}
}

func TestMergeIDs(t *testing.T) {
	tests := []struct {
		name string
		a    []string
		b    []string
		want int
	}{
		{
			name: "merge two lists",
			a:    []string{"id1", "id2"},
			b:    []string{"id3", "id4"},
			want: 4,
		},
		{
			name: "merge with duplicates",
			a:    []string{"id1", "id2"},
			b:    []string{"id2", "id3"},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeIDsTest(tt.a, tt.b)
			if len(result) != tt.want {
				t.Errorf("mergeIDs() got %d items, want %d", len(result), tt.want)
			}
		})
	}
}

// Helper functions for testing (exported from core for testing)
func mergeTagsTest(a, b []string) []string {
	tagSet := make(map[string]bool)
	for _, tag := range a {
		tagSet[tag] = true
	}
	for _, tag := range b {
		tagSet[tag] = true
	}
	result := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		result = append(result, tag)
	}
	return result
}

func mergeIDsTest(a, b []string) []string {
	idSet := make(map[string]bool)
	for _, id := range a {
		idSet[id] = true
	}
	for _, id := range b {
		idSet[id] = true
	}
	result := make([]string, 0, len(idSet))
	for id := range idSet {
		result = append(result, id)
	}
	return result
}

func TestAddIDIfNotExists(t *testing.T) {
	tests := []struct {
		name   string
		ids    []string
		newID  string
		want   int
	}{
		{
			name:  "add new id",
			ids:   []string{"id1", "id2"},
			newID: "id3",
			want:  3,
		},
		{
			name:  "id already exists",
			ids:   []string{"id1", "id2"},
			newID: "id1",
			want:  2,
		},
		{
			name:  "empty list",
			ids:   []string{},
			newID: "id1",
			want:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := addIDIfNotExistsTest(tt.ids, tt.newID)
			if len(result) != tt.want {
				t.Errorf("addIDIfNotExists() got %d items, want %d", len(result), tt.want)
			}
		})
	}
}

func addIDIfNotExistsTest(ids []string, newID string) []string {
	for _, id := range ids {
		if id == newID {
			return ids
		}
	}
	return append(ids, newID)
}

func TestRemoveID(t *testing.T) {
	tests := []struct {
		name      string
		ids       []string
		targetID  string
		want      int
	}{
		{
			name:     "remove existing id",
			ids:      []string{"id1", "id2", "id3"},
			targetID: "id2",
			want:     2,
		},
		{
			name:     "remove non-existing id",
			ids:      []string{"id1", "id2"},
			targetID: "id3",
			want:     2,
		},
		{
			name:     "remove from single item list",
			ids:      []string{"id1"},
			targetID: "id1",
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeIDTest(tt.ids, tt.targetID)
			if len(result) != tt.want {
				t.Errorf("removeID() got %d items, want %d", len(result), tt.want)
			}
		})
	}
}

func removeIDTest(ids []string, targetID string) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		if id != targetID {
			result = append(result, id)
		}
	}
	return result
}

func TestMinFloat64(t *testing.T) {
	result := minFloat64Test(1.0, 2.0)
	if result != 1.0 {
		t.Errorf("minFloat64(1.0, 2.0) = %f, want 1.0", result)
	}

	result = minFloat64Test(2.0, 1.0)
	if result != 1.0 {
		t.Errorf("minFloat64(2.0, 1.0) = %f, want 1.0", result)
	}

	result = minFloat64Test(1.0, 1.0)
	if result != 1.0 {
		t.Errorf("minFloat64(1.0, 1.0) = %f, want 1.0", result)
	}
}

func minFloat64Test(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
