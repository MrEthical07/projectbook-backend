package activity

import (
	"net/http"
	"strconv"
	"strings"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type listQuery struct {
	Offset int
	Limit  int
}

type ListProjectActivityResponse struct {
	Items      []ActivityItem `json:"items"`
	NextCursor *string        `json:"next_cursor,omitempty"`
}

type ActivityItem struct {
	ID       string `json:"id"`
	User     string `json:"user"`
	Initials string `json:"initials"`
	Action   string `json:"action"`
	Artifact string `json:"artifact"`
	Href     string `json:"href"`
	At       string `json:"at"`
}

func parseLimit(raw string) (int, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return defaultLimit, nil
	}
	parsed, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "limit must be an integer")
	}
	if parsed < 0 {
		return 0, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "limit must be non-negative")
	}
	if parsed == 0 {
		return defaultLimit, nil
	}
	if parsed > maxLimit {
		return maxLimit, nil
	}
	return parsed, nil
}

func initialsFromName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "NA"
	}
	parts := strings.Fields(trimmed)
	if len(parts) == 1 {
		runes := []rune(parts[0])
		if len(runes) == 1 {
			return strings.ToUpper(string(runes[0]))
		}
		return strings.ToUpper(string(runes[0]) + string(runes[1]))
	}
	first := []rune(parts[0])
	last := []rune(parts[len(parts)-1])
	return strings.ToUpper(string(first[0]) + string(last[0]))
}
