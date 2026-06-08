// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package blueprint

import (
	"testing"
)

func TestBuildRelationshipData(t *testing.T) {
	t.Run("multiple_ids", func(t *testing.T) {
		data := buildRelationshipData("apps", []string{"id1", "id2"})

		if len(data) != 2 {
			t.Fatalf("expected 2 items, got %d", len(data))
		}
		if data[0].Type != "apps" {
			t.Errorf("expected type apps, got %s", data[0].Type)
		}
		if data[0].ID != "id1" {
			t.Errorf("expected id id1, got %s", data[0].ID)
		}
		if data[1].ID != "id2" {
			t.Errorf("expected id id2, got %s", data[1].ID)
		}
	})

	t.Run("empty_ids", func(t *testing.T) {
		data := buildRelationshipData("apps", []string{})

		if len(data) != 0 {
			t.Errorf("expected 0 items, got %d", len(data))
		}
	})
}

func TestDiffIDs(t *testing.T) {
	t.Run("add_and_remove", func(t *testing.T) {
		toAdd, toRemove := diffIDs([]string{"a", "b"}, []string{"b", "c"})

		if len(toAdd) != 1 || toAdd[0] != "c" {
			t.Errorf("expected toAdd [c], got %v", toAdd)
		}
		if len(toRemove) != 1 || toRemove[0] != "a" {
			t.Errorf("expected toRemove [a], got %v", toRemove)
		}
	})

	t.Run("no_changes", func(t *testing.T) {
		toAdd, toRemove := diffIDs([]string{"a", "b"}, []string{"a", "b"})

		if len(toAdd) != 0 {
			t.Errorf("expected empty toAdd, got %v", toAdd)
		}
		if len(toRemove) != 0 {
			t.Errorf("expected empty toRemove, got %v", toRemove)
		}
	})

	t.Run("all_new", func(t *testing.T) {
		toAdd, toRemove := diffIDs([]string{}, []string{"a", "b"})

		if len(toAdd) != 2 {
			t.Errorf("expected 2 toAdd, got %v", toAdd)
		}
		if len(toRemove) != 0 {
			t.Errorf("expected empty toRemove, got %v", toRemove)
		}
	})

	t.Run("all_removed", func(t *testing.T) {
		toAdd, toRemove := diffIDs([]string{"a", "b"}, []string{})

		if len(toAdd) != 0 {
			t.Errorf("expected empty toAdd, got %v", toAdd)
		}
		if len(toRemove) != 2 {
			t.Errorf("expected 2 toRemove, got %v", toRemove)
		}
	})

	t.Run("duplicates_in_current", func(t *testing.T) {
		toAdd, toRemove := diffIDs([]string{"a", "a", "b"}, []string{"a"})

		if len(toAdd) != 0 {
			t.Errorf("expected empty toAdd, got %v", toAdd)
		}
		if len(toRemove) != 1 || toRemove[0] != "b" {
			t.Errorf("expected toRemove [b], got %v", toRemove)
		}
	})
}
