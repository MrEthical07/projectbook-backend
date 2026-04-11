package pages

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
)

const maxListLimit = 100

var slugSanitizer = regexp.MustCompile(`[^a-z0-9]+`)

var pageStatuses = map[string]struct{}{
	"Draft":    {},
	"Archived": {},
}

type listQuery struct {
	Status string
	Offset int
	Limit  int
}

type createPageRequest struct {
	Title string `json:"title"`
}

func (r createPageRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "title is required")
	}
	return nil
}

type updatePageRequest struct {
	State map[string]any `json:"state"`
}

func (r updatePageRequest) Validate() error {
	if r.State == nil {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "state payload is required")
	}
	if status := toString(r.State["status"]); status != "" {
		if !isAllowedStatus(pageStatuses, status) {
			return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid page status")
		}
	}
	return nil
}

type renamePageRequest struct {
	Title string `json:"title"`
}

func (r renamePageRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "title is required")
	}
	return nil
}

func parseOptionalIntQuery(raw string, fallback int, name string) (int, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, name+" must be an integer")
	}
	if parsed < 0 {
		return 0, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, name+" must be non-negative")
	}
	return parsed, nil
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 25
	}
	if limit > maxListLimit {
		return maxListLimit
	}
	return limit
}

func isAllowedStatus(allowed map[string]struct{}, raw string) bool {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return false
	}
	_, ok := allowed[trimmed]
	return ok
}

func toString(v any) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func toSlice(v any) []any {
	s, _ := v.([]any)
	return s
}

func toMap(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

func slugify(raw string) string {
	trimmed := strings.ToLower(strings.TrimSpace(raw))
	if trimmed == "" {
		return "page"
	}
	slug := slugSanitizer.ReplaceAllString(trimmed, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		return "page"
	}
	if len(slug) > 96 {
		slug = strings.Trim(slug[:96], "-")
		if slug == "" {
			return "page"
		}
	}
	return slug
}
