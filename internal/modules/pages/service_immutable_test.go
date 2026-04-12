package pages

import "testing"

func TestPageImmutableHelpers(t *testing.T) {
	t.Run("archived blocks rename/delete style operations", func(t *testing.T) {
		err := enforceMutableOperation("page", "Archived", pageImmutableStatuses)
		if err == nil {
			t.Fatal("expected immutable operation error")
		}
	})

	t.Run("archived allows archive-only patch", func(t *testing.T) {
		err := enforceArchiveOnlyForImmutableUpdate("page", "Archived", map[string]any{"status": "Archived"}, pageImmutableStatuses)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("archived blocks mixed patch", func(t *testing.T) {
		err := enforceArchiveOnlyForImmutableUpdate("page", "Archived", map[string]any{"status": "Archived", "tags": []any{"x"}}, pageImmutableStatuses)
		if err == nil {
			t.Fatal("expected immutable mixed patch rejection")
		}
	})

	t.Run("transition matrix draft to archived is allowed", func(t *testing.T) {
		ok := isAllowedTransition("Draft", "Archived", map[string]map[string]struct{}{
			"Draft":    {"Draft": {}, "Archived": {}},
			"Archived": {"Archived": {}},
		})
		if !ok {
			t.Fatal("expected Draft -> Archived to be allowed")
		}
	})
}
