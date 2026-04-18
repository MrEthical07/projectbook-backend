package patchx

import (
	"fmt"
)

// FieldRule defines validation constraints for a patch field.
type FieldRule struct {
	AllowNull      bool
	Nested         map[string]FieldRule
	AllowAnyNested bool
}

// ValidatePatch rejects unknown fields and validates nested object constraints.
func ValidatePatch(patch map[string]any, rules map[string]FieldRule) error {
	if patch == nil {
		return fmt.Errorf("patch payload is required")
	}
	return validateMap(patch, rules, "")
}

func validateMap(patch map[string]any, rules map[string]FieldRule, parent string) error {
	for key, value := range patch {
		rule, ok := rules[key]
		if !ok {
			if parent == "" {
				return fmt.Errorf("invalid field %q", key)
			}
			return fmt.Errorf("invalid field %q", parent+"."+key)
		}

		if value == nil {
			if !rule.AllowNull {
				if parent == "" {
					return fmt.Errorf("field %q cannot be null", key)
				}
				return fmt.Errorf("field %q cannot be null", parent+"."+key)
			}
			continue
		}

		if rule.Nested == nil && !rule.AllowAnyNested {
			continue
		}

		nestedPatch, ok := asMap(value)
		if !ok {
			if parent == "" {
				return fmt.Errorf("field %q must be an object", key)
			}
			return fmt.Errorf("field %q must be an object", parent+"."+key)
		}

		if rule.AllowAnyNested {
			continue
		}

		nextParent := key
		if parent != "" {
			nextParent = parent + "." + key
		}
		if err := validateMap(nestedPatch, rule.Nested, nextParent); err != nil {
			return err
		}
	}

	return nil
}

// MergeShallow applies patch semantics:
// - missing fields: unchanged
// - null values: remove field
// - arrays: full replace
// - nested objects: shallow merge
func MergeShallow(base map[string]any, patch map[string]any) map[string]any {
	output := cloneMap(base)
	if patch == nil {
		return output
	}

	for key, patchValue := range patch {
		if patchValue == nil {
			delete(output, key)
			continue
		}

		patchMap, isPatchMap := asMap(patchValue)
		if !isPatchMap {
			output[key] = cloneValue(patchValue)
			continue
		}

		existingMap, hasExistingMap := asMap(output[key])
		if !hasExistingMap {
			output[key] = cloneMap(patchMap)
			continue
		}

		mergedNested := cloneMap(existingMap)
		for nestedKey, nestedValue := range patchMap {
			if nestedValue == nil {
				delete(mergedNested, nestedKey)
				continue
			}
			mergedNested[nestedKey] = cloneValue(nestedValue)
		}
		output[key] = mergedNested
	}

	return output
}

func asMap(value any) (map[string]any, bool) {
	mapped, ok := value.(map[string]any)
	return mapped, ok
}

func cloneMap(in map[string]any) map[string]any {
	if in == nil {
		return map[string]any{}
	}

	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = cloneValue(value)
	}
	return out
}

func cloneSlice(in []any) []any {
	if in == nil {
		return []any{}
	}

	out := make([]any, 0, len(in))
	for _, value := range in {
		out = append(out, cloneValue(value))
	}
	return out
}

func cloneValue(value any) any {
	if value == nil {
		return nil
	}
	if mapped, ok := value.(map[string]any); ok {
		return cloneMap(mapped)
	}
	if arr, ok := value.([]any); ok {
		return cloneSlice(arr)
	}
	return value
}
