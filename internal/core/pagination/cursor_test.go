package pagination

import "testing"

func TestEncodeDecodeOffsetCursor(t *testing.T) {
	offset := 42
	cursor := EncodeOffsetCursor(offset)
	decoded, err := DecodeOffsetCursor(cursor)
	if err != nil {
		t.Fatalf("DecodeOffsetCursor returned error: %v", err)
	}
	if decoded != offset {
		t.Fatalf("expected offset %d, got %d", offset, decoded)
	}
}

func TestDecodeOffsetCursorRejectsInvalid(t *testing.T) {
	cases := []string{"", "abc", "djI6MTA="}
	for _, cursor := range cases {
		if _, err := DecodeOffsetCursor(cursor); err == nil {
			t.Fatalf("expected error for cursor %q", cursor)
		}
	}
}
