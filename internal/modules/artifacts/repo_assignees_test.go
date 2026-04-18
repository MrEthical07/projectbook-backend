package artifacts

import (
	"reflect"
	"testing"
)

func TestExtractTaskAssigneeIDs(t *testing.T) {
	tests := []struct {
		name      string
		patch     map[string]any
		wantIDs   []string
		wantFound bool
	}{
		{
			name:      "nil patch",
			patch:     nil,
			wantIDs:   nil,
			wantFound: false,
		},
		{
			name: "assignedToIds normalized and deduplicated",
			patch: map[string]any{
				"assignedToIds": []any{" user-a ", "user-b", "user-a", ""},
			},
			wantIDs:   []string{"user-a", "user-b"},
			wantFound: true,
		},
		{
			name: "legacy assignedToId fallback",
			patch: map[string]any{
				"assignedToId": " user-legacy ",
			},
			wantIDs:   []string{"user-legacy"},
			wantFound: true,
		},
		{
			name: "legacy assignedToId clear assignment",
			patch: map[string]any{
				"assignedToId": "",
			},
			wantIDs:   []string{},
			wantFound: true,
		},
		{
			name: "assignedToIds takes precedence over assignedToId",
			patch: map[string]any{
				"assignedToIds": []string{"user-primary", "user-secondary"},
				"assignedToId":  "user-legacy",
			},
			wantIDs:   []string{"user-primary", "user-secondary"},
			wantFound: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotIDs, gotFound := extractTaskAssigneeIDs(tc.patch)
			if gotFound != tc.wantFound {
				t.Fatalf("found=%v wantFound=%v", gotFound, tc.wantFound)
			}
			if !reflect.DeepEqual(gotIDs, tc.wantIDs) {
				t.Fatalf("ids=%v want=%v", gotIDs, tc.wantIDs)
			}
		})
	}
}

func TestNormalizeUniqueStringsPreservesOrder(t *testing.T) {
	got := normalizeUniqueStrings([]string{" a ", "b", "a", "", "c", "b"})
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%v want=%v", got, want)
	}
}
