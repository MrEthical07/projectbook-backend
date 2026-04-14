package artifacts

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
)

const maxListLimit = 100

var slugSanitizer = regexp.MustCompile(`[^a-z0-9]+`)

var storyStatuses = map[string]struct{}{
	"Draft":    {},
	"Locked":   {},
	"Archived": {},
}

var journeyStatuses = map[string]struct{}{
	"Draft":    {},
	"Archived": {},
}

var problemStatuses = map[string]struct{}{
	"Draft":    {},
	"Locked":   {},
	"Archived": {},
}

var ideaStatuses = map[string]struct{}{
	"Considered": {},
	"Selected":   {},
	"Rejected":   {},
	"Archived":   {},
}

var taskStatuses = map[string]struct{}{
	"Planned":     {},
	"In Progress": {},
	"Completed":   {},
	"Abandoned":   {},
}

var feedbackOutcomes = map[string]struct{}{
	"Validated":       {},
	"Invalidated":     {},
	"Needs Iteration": {},
}

var artifactTypeMap = map[string]string{
	"story":             "story",
	"stories":           "story",
	"journey":           "journey",
	"journeys":          "journey",
	"problem":           "problem",
	"problems":          "problem",
	"problem statement": "problem",
	"idea":              "idea",
	"ideas":             "idea",
	"task":              "task",
	"tasks":             "task",
	"feedback":          "feedback",
}

type listQuery struct {
	Status  string
	Outcome string
	Offset  int
	Limit   int
}

type createStoryRequest struct {
	Title string `json:"title"`
}

func (r createStoryRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "title is required")
	}
	return nil
}

type createJourneyRequest struct {
	Title string `json:"title"`
}

func (r createJourneyRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "title is required")
	}
	return nil
}

type createProblemRequest struct {
	Statement string `json:"statement"`
}

func (r createProblemRequest) Validate() error {
	if strings.TrimSpace(r.Statement) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "statement is required")
	}
	return nil
}

type createIdeaRequest struct {
	Title string `json:"title"`
}

func (r createIdeaRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "title is required")
	}
	return nil
}

type createTaskRequest struct {
	Title string `json:"title"`
}

func (r createTaskRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "title is required")
	}
	return nil
}

type createFeedbackRequest struct {
	Title string `json:"title"`
}

func (r createFeedbackRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "title is required")
	}
	return nil
}

type updateStoryRequest struct {
	Story map[string]any `json:"story"`
}

func (r updateStoryRequest) Validate() error {
	if r.Story == nil {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "story payload is required")
	}
	if status := toString(r.Story["status"]); status != "" {
		canonicalStatus, ok := normalizeAllowedStatus(storyStatuses, status)
		if !ok {
			return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid story status")
		}
		r.Story["status"] = canonicalStatus
	}
	return nil
}

type updateJourneyRequest struct {
	Journey map[string]any `json:"journey"`
}

func (r updateJourneyRequest) Validate() error {
	if r.Journey == nil {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "journey payload is required")
	}
	if status := toString(r.Journey["status"]); status != "" {
		canonicalStatus, ok := normalizeAllowedStatus(journeyStatuses, status)
		if !ok {
			return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid journey status")
		}
		r.Journey["status"] = canonicalStatus
	}
	return nil
}

type updateProblemRequest struct {
	State map[string]any `json:"state"`
}

func (r updateProblemRequest) Validate() error {
	if r.State == nil {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "state payload is required")
	}
	if status := toString(r.State["status"]); status != "" {
		if !isAllowedStatus(problemStatuses, status) {
			return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid problem status")
		}
	}
	return nil
}

type updateIdeaRequest struct {
	State map[string]any `json:"state"`
}

func (r updateIdeaRequest) Validate() error {
	if r.State == nil {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "state payload is required")
	}
	return nil
}

type updateTaskRequest struct {
	State map[string]any `json:"state"`
}

func (r updateTaskRequest) Validate() error {
	if r.State == nil {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "state payload is required")
	}
	return nil
}

type updateFeedbackRequest struct {
	State map[string]any `json:"state"`
}

func (r updateFeedbackRequest) Validate() error {
	if r.State == nil {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "state payload is required")
	}
	if outcome := toString(r.State["outcome"]); outcome != "" {
		if !isAllowedStatus(feedbackOutcomes, outcome) {
			return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid feedback outcome")
		}
	}
	return nil
}

type updateProblemStatusRequest struct {
	Status string `json:"status"`
}

func (r updateProblemStatusRequest) Validate() error {
	if !isAllowedStatus(problemStatuses, r.Status) {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid problem status")
	}
	return nil
}

type updateIdeaStatusRequest struct {
	Status string `json:"status"`
}

func (r updateIdeaStatusRequest) Validate() error {
	if !isAllowedStatus(ideaStatuses, r.Status) {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid idea status")
	}
	return nil
}

type updateTaskStatusRequest struct {
	Status string `json:"status"`
}

func (r updateTaskStatusRequest) Validate() error {
	if !isAllowedStatus(taskStatuses, r.Status) {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid task status")
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
	_, ok := normalizeAllowedStatus(allowed, raw)
	return ok
}

func normalizeAllowedStatus(allowed map[string]struct{}, raw string) (string, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", false
	}
	if _, ok := allowed[trimmed]; ok {
		return trimmed, true
	}
	for candidate := range allowed {
		if strings.EqualFold(candidate, trimmed) {
			return candidate, true
		}
	}
	return "", false
}

func normalizeArtifactType(raw string) (string, bool) {
	key := strings.ToLower(strings.TrimSpace(raw))
	v, ok := artifactTypeMap[key]
	return v, ok
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

func countStringItems(values []any) int {
	count := 0
	for _, item := range values {
		if strings.TrimSpace(toString(item)) != "" {
			count++
		}
	}
	return count
}

func slugify(raw string) string {
	trimmed := strings.ToLower(strings.TrimSpace(raw))
	if trimmed == "" {
		return "artifact"
	}
	slug := slugSanitizer.ReplaceAllString(trimmed, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		return "artifact"
	}
	if len(slug) > 96 {
		slug = strings.Trim(slug[:96], "-")
		if slug == "" {
			return "artifact"
		}
	}
	return slug
}
