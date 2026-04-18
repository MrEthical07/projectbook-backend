package patchx

import "testing"

func TestValidatePatchRejectsUnknownTopLevelField(t *testing.T) {
	rules := map[string]FieldRule{
		"title": {},
	}

	err := ValidatePatch(map[string]any{"unknown": "x"}, rules)
	if err == nil {
		t.Fatal("expected unknown field validation error")
	}
}

func TestValidatePatchRejectsUnknownNestedField(t *testing.T) {
	rules := map[string]FieldRule{
		"persona": {
			Nested: map[string]FieldRule{
				"name": {},
			},
		},
	}

	err := ValidatePatch(map[string]any{
		"persona": map[string]any{"nickname": "ali"},
	}, rules)
	if err == nil {
		t.Fatal("expected unknown nested field validation error")
	}
}

func TestMergeShallowSemantics(t *testing.T) {
	base := map[string]any{
		"title": "A",
		"tags":  []any{"x", "y"},
		"persona": map[string]any{
			"name": "Ana",
			"role": "PM",
		},
		"notes": "keep",
	}

	patch := map[string]any{
		"tags": []any{"z"},
		"persona": map[string]any{
			"role": nil,
			"bio":  "hello",
		},
		"notes": nil,
	}

	merged := MergeShallow(base, patch)

	if merged["title"] != "A" {
		t.Fatalf("missing field should remain unchanged")
	}

	tags, ok := merged["tags"].([]any)
	if !ok || len(tags) != 1 || tags[0] != "z" {
		t.Fatalf("arrays must be fully replaced")
	}

	persona, ok := merged["persona"].(map[string]any)
	if !ok {
		t.Fatalf("expected persona map")
	}
	if _, hasRole := persona["role"]; hasRole {
		t.Fatalf("null nested value should remove nested field")
	}
	if persona["name"] != "Ana" || persona["bio"] != "hello" {
		t.Fatalf("nested object should shallow merge")
	}

	if _, hasNotes := merged["notes"]; hasNotes {
		t.Fatalf("null should remove top-level field")
	}
}
