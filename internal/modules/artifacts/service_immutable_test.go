package artifacts

import "testing"

func TestEnforceArchiveOnlyForImmutableUpdate(t *testing.T) {
	tests := []struct {
		name      string
		from      string
		patch     map[string]any
		immutable map[string]struct{}
		wantErr   bool
	}{
		{
			name:      "locked allows archive-only patch",
			from:      "Locked",
			patch:     map[string]any{"status": "Archived"},
			immutable: storyImmutableStatuses,
			wantErr:   false,
		},
		{
			name:      "locked blocks content update",
			from:      "Locked",
			patch:     map[string]any{"status": "Archived", "title": "new title"},
			immutable: storyImmutableStatuses,
			wantErr:   true,
		},
		{
			name:      "archived allows archived idempotent patch",
			from:      "Archived",
			patch:     map[string]any{"status": "Archived"},
			immutable: storyImmutableStatuses,
			wantErr:   false,
		},
		{
			name:      "non-immutable status allows update",
			from:      "Draft",
			patch:     map[string]any{"title": "new title"},
			immutable: storyImmutableStatuses,
			wantErr:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := enforceArchiveOnlyForImmutableUpdate("story", tc.from, tc.patch, tc.immutable)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tc.wantErr)
			}
		})
	}
}

func TestEnforceArchiveOnlyForImmutableStatusChange(t *testing.T) {
	tests := []struct {
		name      string
		from      string
		to        string
		immutable map[string]struct{}
		wantErr   bool
	}{
		{
			name:      "locked allows archive transition",
			from:      "Locked",
			to:        "Archived",
			immutable: problemImmutableStatuses,
			wantErr:   false,
		},
		{
			name:      "locked blocks non-archive transition",
			from:      "Locked",
			to:        "Locked",
			immutable: problemImmutableStatuses,
			wantErr:   true,
		},
		{
			name:      "completed task blocks status changes",
			from:      "Completed",
			to:        "Completed",
			immutable: taskImmutableStatuses,
			wantErr:   true,
		},
		{
			name:      "non-immutable status allows transition",
			from:      "Draft",
			to:        "Locked",
			immutable: problemImmutableStatuses,
			wantErr:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := enforceArchiveOnlyForImmutableStatusChange("artifact", tc.from, tc.to, tc.immutable)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tc.wantErr)
			}
		})
	}
}
