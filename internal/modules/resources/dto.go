package resources

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
)

const maxListLimit = 100

var slugSanitizer = regexp.MustCompile(`[^a-z0-9]+`)

var resourceStatuses = map[string]struct{}{
	"Active":   {},
	"Archived": {},
}

type listQuery struct {
	Status  string
	DocType string
	Sort    string
	Order   string
	Offset  int
	Limit   int
}

type createResourceRequest struct {
	Name    string `json:"name"`
	DocType string `json:"docType"`
}

func (r createResourceRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "name is required")
	}
	if strings.TrimSpace(r.DocType) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "docType is required")
	}
	return nil
}

type updateResourceRequest struct {
	State map[string]any `json:"state"`
}

func (r updateResourceRequest) Validate() error {
	if r.State == nil {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "state payload is required")
	}
	return nil
}

type updateResourceStatusRequest struct {
	Status string `json:"status"`
}

func (r updateResourceStatusRequest) Validate() error {
	if !isAllowedStatus(resourceStatuses, r.Status) {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid resource status")
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

func normalizeSort(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "name":
		return "name"
	case "uploaddate":
		return "uploadDate"
	default:
		return "lastUpdated"
	}
}

func normalizeOrder(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "asc") {
		return "asc"
	}
	return "desc"
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
		return "resource"
	}
	slug := slugSanitizer.ReplaceAllString(trimmed, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		return "resource"
	}
	if len(slug) > 96 {
		slug = strings.Trim(slug[:96], "-")
		if slug == "" {
			return "resource"
		}
	}
	return slug
}
