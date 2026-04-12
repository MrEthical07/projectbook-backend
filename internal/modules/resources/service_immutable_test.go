package resources

import "testing"

func TestResourceImmutableHelpers(t *testing.T) {
	t.Run("archived blocks non-archive content update", func(t *testing.T) {
		err := enforceArchiveOnlyForImmutableUpdate("resource", "Archived", map[string]any{"name": "doc"}, resourceImmutableStatuses)
		if err == nil {
			t.Fatal("expected immutable update error")
		}
	})

	t.Run("archived allows archive-only patch", func(t *testing.T) {
		err := enforceArchiveOnlyForImmutableUpdate("resource", "Archived", map[string]any{"status": "Archived"}, resourceImmutableStatuses)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("archived blocks non-archive status transition", func(t *testing.T) {
		err := enforceArchiveOnlyForImmutableStatusChange("resource", "Archived", "Active", resourceImmutableStatuses)
		if err == nil {
			t.Fatal("expected immutable status error")
		}
	})

	t.Run("transition matrix active to archived is allowed", func(t *testing.T) {
		ok := isAllowedTransition("Active", "Archived", map[string]map[string]struct{}{
			"Active":   {"Active": {}, "Archived": {}},
			"Archived": {"Archived": {}},
		})
		if !ok {
			t.Fatal("expected Active -> Archived to be allowed")
		}
	})
}
