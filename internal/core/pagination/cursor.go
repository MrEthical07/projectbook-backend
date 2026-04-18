package pagination

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

const offsetCursorPrefix = "v1:"

// EncodeOffsetCursor creates an opaque cursor token for a non-negative offset.
func EncodeOffsetCursor(offset int) string {
	if offset < 0 {
		offset = 0
	}
	payload := fmt.Sprintf("%s%d", offsetCursorPrefix, offset)
	return base64.RawURLEncoding.EncodeToString([]byte(payload))
}

// DecodeOffsetCursor parses an opaque cursor token into an offset.
func DecodeOffsetCursor(cursor string) (int, error) {
	trimmed := strings.TrimSpace(cursor)
	if trimmed == "" {
		return 0, fmt.Errorf("cursor is required")
	}

	raw, err := base64.RawURLEncoding.DecodeString(trimmed)
	if err != nil {
		return 0, fmt.Errorf("cursor is invalid")
	}

	decoded := string(raw)
	if !strings.HasPrefix(decoded, offsetCursorPrefix) {
		return 0, fmt.Errorf("cursor is invalid")
	}

	offsetRaw := strings.TrimPrefix(decoded, offsetCursorPrefix)
	offset, err := strconv.Atoi(offsetRaw)
	if err != nil || offset < 0 {
		return 0, fmt.Errorf("cursor is invalid")
	}

	return offset, nil
}
