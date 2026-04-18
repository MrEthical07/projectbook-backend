package sidebar

import (
	"net/http"
	"strings"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
)

const (
	prefixStories          = "stories"
	prefixJourneys         = "journeys"
	prefixProblemStatement = "problem-statement"
	prefixIdeas            = "ideas"
	prefixTasks            = "tasks"
	prefixFeedback         = "feedback"
	prefixPages            = "pages"
)

var allowedPrefixes = map[string]struct{}{
	prefixStories:          {},
	prefixJourneys:         {},
	prefixProblemStatement: {},
	prefixIdeas:            {},
	prefixTasks:            {},
	prefixFeedback:         {},
	prefixPages:            {},
}

type createSidebarArtifactRequest struct {
	Prefix string `json:"prefix"`
	Title  string `json:"title"`
}

type SidebarArtifactResponse struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type SidebarDeleteResponse struct {
	ID string `json:"id"`
}

func (r createSidebarArtifactRequest) Validate() error {
	if _, ok := allowedPrefixes[normalizePrefix(r.Prefix)]; !ok {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid prefix")
	}
	if strings.TrimSpace(r.Title) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "title is required")
	}
	return nil
}

type renameSidebarArtifactRequest struct {
	Prefix string `json:"prefix"`
	Title  string `json:"title"`
}

func (r renameSidebarArtifactRequest) Validate() error {
	if _, ok := allowedPrefixes[normalizePrefix(r.Prefix)]; !ok {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid prefix")
	}
	if strings.TrimSpace(r.Title) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "title is required")
	}
	return nil
}

type deleteSidebarArtifactRequest struct {
	Prefix string `json:"prefix"`
}

func (r deleteSidebarArtifactRequest) Validate() error {
	if _, ok := allowedPrefixes[normalizePrefix(r.Prefix)]; !ok {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid prefix")
	}
	return nil
}

func normalizePrefix(raw string) string {
	trimmed := strings.ToLower(strings.TrimSpace(raw))
	switch trimmed {
	case "problem", "problems", "problem-statement":
		return prefixProblemStatement
	default:
		return trimmed
	}
}

func mapString(item map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := item[key]; ok {
			if text, ok := value.(string); ok && strings.TrimSpace(text) != "" {
				return strings.TrimSpace(text)
			}
		}
	}
	return ""
}

func decodeSidebarArtifactResponse(payload map[string]any) SidebarArtifactResponse {
	return SidebarArtifactResponse{
		ID:    mapString(payload, "id"),
		Title: mapString(payload, "title"),
	}
}

func decodeSidebarDeleteResponse(payload map[string]any) SidebarDeleteResponse {
	return SidebarDeleteResponse{ID: mapString(payload, "id")}
}
