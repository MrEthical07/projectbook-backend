package artifacts

import (
	"net/http"
	"strings"

	"github.com/MrEthical07/superapi/internal/core/auth"
	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/httpx"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListStories(ctx *httpx.Context, _ httpx.NoBody) (map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	query, err := parseListQuery(ctx)
	if err != nil {
		return nil, err
	}
	query.Status = strings.TrimSpace(ctx.Query("status"))
	return h.svc.ListStories(ctx.Context(), resolveProjectScope(principal, projectID), query)
}

func (h *Handler) CreateStory(ctx *httpx.Context, req createStoryRequest) (httpx.Result[map[string]any], error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return httpx.Result[map[string]any]{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return httpx.Result[map[string]any]{}, err
	}
	created, err := h.svc.CreateStory(ctx.Context(), resolveProjectScope(principal, projectID), principal.UserID, req)
	if err != nil {
		return httpx.Result[map[string]any]{}, err
	}
	return httpx.Result[map[string]any]{Status: http.StatusCreated, Data: created}, nil
}

func (h *Handler) GetStory(ctx *httpx.Context, _ httpx.NoBody) (map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	storyID := strings.TrimSpace(ctx.Param("slug"))
	if storyID == "" {
		return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "slug is required")
	}
	return h.svc.GetStory(ctx.Context(), resolveProjectScope(principal, projectID), storyID)
}

func (h *Handler) UpdateStory(ctx *httpx.Context, req updateStoryRequest) (map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	storyID := strings.TrimSpace(ctx.Param("storyId"))
	if storyID == "" {
		return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "storyId is required")
	}
	return h.svc.UpdateStory(ctx.Context(), resolveProjectScope(principal, projectID), storyID, principal.UserID, req)
}

func (h *Handler) ListJourneys(ctx *httpx.Context, _ httpx.NoBody) (map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	query, err := parseListQuery(ctx)
	if err != nil {
		return nil, err
	}
	query.Status = strings.TrimSpace(ctx.Query("status"))
	return h.svc.ListJourneys(ctx.Context(), resolveProjectScope(principal, projectID), query)
}

func (h *Handler) CreateJourney(ctx *httpx.Context, req createJourneyRequest) (httpx.Result[map[string]any], error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return httpx.Result[map[string]any]{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return httpx.Result[map[string]any]{}, err
	}
	created, err := h.svc.CreateJourney(ctx.Context(), resolveProjectScope(principal, projectID), principal.UserID, req)
	if err != nil {
		return httpx.Result[map[string]any]{}, err
	}
	return httpx.Result[map[string]any]{Status: http.StatusCreated, Data: created}, nil
}

func (h *Handler) GetJourney(ctx *httpx.Context, _ httpx.NoBody) (map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	journeyID := strings.TrimSpace(ctx.Param("slug"))
	if journeyID == "" {
		return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "slug is required")
	}
	return h.svc.GetJourney(ctx.Context(), resolveProjectScope(principal, projectID), journeyID)
}

func (h *Handler) UpdateJourney(ctx *httpx.Context, req updateJourneyRequest) (map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	journeyID := strings.TrimSpace(ctx.Param("journeyId"))
	if journeyID == "" {
		return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "journeyId is required")
	}
	return h.svc.UpdateJourney(ctx.Context(), resolveProjectScope(principal, projectID), journeyID, principal.UserID, req)
}

func (h *Handler) ListProblems(ctx *httpx.Context, _ httpx.NoBody) (map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	query, err := parseListQuery(ctx)
	if err != nil {
		return nil, err
	}
	query.Status = strings.TrimSpace(ctx.Query("status"))
	return h.svc.ListProblems(ctx.Context(), resolveProjectScope(principal, projectID), query)
}

func (h *Handler) CreateProblem(ctx *httpx.Context, req createProblemRequest) (httpx.Result[map[string]any], error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return httpx.Result[map[string]any]{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return httpx.Result[map[string]any]{}, err
	}
	created, err := h.svc.CreateProblem(ctx.Context(), resolveProjectScope(principal, projectID), principal.UserID, req)
	if err != nil {
		return httpx.Result[map[string]any]{}, err
	}
	return httpx.Result[map[string]any]{Status: http.StatusCreated, Data: created}, nil
}

func (h *Handler) GetProblem(ctx *httpx.Context, _ httpx.NoBody) (map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	problemID := strings.TrimSpace(ctx.Param("slug"))
	if problemID == "" {
		return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "slug is required")
	}
	return h.svc.GetProblem(ctx.Context(), resolveProjectScope(principal, projectID), problemID)
}

func (h *Handler) UpdateProblem(ctx *httpx.Context, req updateProblemRequest) (map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	problemID := strings.TrimSpace(ctx.Param("problemId"))
	if problemID == "" {
		return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "problemId is required")
	}
	return h.svc.UpdateProblem(ctx.Context(), resolveProjectScope(principal, projectID), problemID, principal.UserID, req)
}

func (h *Handler) LockProblem(ctx *httpx.Context, _ httpx.NoBody) (map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	problemID := strings.TrimSpace(ctx.Param("problemId"))
	if problemID == "" {
		return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "problemId is required")
	}
	return h.svc.UpdateProblemStatus(ctx.Context(), resolveProjectScope(principal, projectID), problemID, principal.UserID, updateProblemStatusRequest{Status: "Locked"})
}

func (h *Handler) UpdateProblemStatus(ctx *httpx.Context, req updateProblemStatusRequest) (map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	problemID := strings.TrimSpace(ctx.Param("problemId"))
	if problemID == "" {
		return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "problemId is required")
	}
	return h.svc.UpdateProblemStatus(ctx.Context(), resolveProjectScope(principal, projectID), problemID, principal.UserID, req)
}

func (h *Handler) ListIdeas(ctx *httpx.Context, _ httpx.NoBody) (map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	query, err := parseListQuery(ctx)
	if err != nil {
		return nil, err
	}
	query.Status = strings.TrimSpace(ctx.Query("status"))
	return h.svc.ListIdeas(ctx.Context(), resolveProjectScope(principal, projectID), query)
}

func (h *Handler) CreateIdea(ctx *httpx.Context, req createIdeaRequest) (httpx.Result[map[string]any], error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return httpx.Result[map[string]any]{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return httpx.Result[map[string]any]{}, err
	}
	created, err := h.svc.CreateIdea(ctx.Context(), resolveProjectScope(principal, projectID), principal.UserID, req)
	if err != nil {
		return httpx.Result[map[string]any]{}, err
	}
	return httpx.Result[map[string]any]{Status: http.StatusCreated, Data: created}, nil
}

func (h *Handler) GetIdea(ctx *httpx.Context, _ httpx.NoBody) (map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	ideaID := strings.TrimSpace(ctx.Param("slug"))
	if ideaID == "" {
		return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "slug is required")
	}
	return h.svc.GetIdea(ctx.Context(), resolveProjectScope(principal, projectID), ideaID)
}

func (h *Handler) SelectIdea(ctx *httpx.Context, _ httpx.NoBody) (map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	ideaID := strings.TrimSpace(ctx.Param("ideaId"))
	if ideaID == "" {
		return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "ideaId is required")
	}
	return h.svc.UpdateIdeaStatus(ctx.Context(), resolveProjectScope(principal, projectID), ideaID, principal.UserID, updateIdeaStatusRequest{Status: "Selected"})
}

func (h *Handler) UpdateIdea(ctx *httpx.Context, req updateIdeaRequest) (map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	ideaID := strings.TrimSpace(ctx.Param("ideaId"))
	if ideaID == "" {
		return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "ideaId is required")
	}
	return h.svc.UpdateIdea(ctx.Context(), resolveProjectScope(principal, projectID), ideaID, principal.UserID, req)
}

func (h *Handler) UpdateIdeaStatus(ctx *httpx.Context, req updateIdeaStatusRequest) (map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	ideaID := strings.TrimSpace(ctx.Param("ideaId"))
	if ideaID == "" {
		return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "ideaId is required")
	}
	return h.svc.UpdateIdeaStatus(ctx.Context(), resolveProjectScope(principal, projectID), ideaID, principal.UserID, req)
}

func (h *Handler) ListTasks(ctx *httpx.Context, _ httpx.NoBody) (map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	query, err := parseListQuery(ctx)
	if err != nil {
		return nil, err
	}
	query.Status = strings.TrimSpace(ctx.Query("status"))
	return h.svc.ListTasks(ctx.Context(), resolveProjectScope(principal, projectID), query)
}

func (h *Handler) CreateTask(ctx *httpx.Context, req createTaskRequest) (httpx.Result[map[string]any], error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return httpx.Result[map[string]any]{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return httpx.Result[map[string]any]{}, err
	}
	created, err := h.svc.CreateTask(ctx.Context(), resolveProjectScope(principal, projectID), principal.UserID, req)
	if err != nil {
		return httpx.Result[map[string]any]{}, err
	}
	return httpx.Result[map[string]any]{Status: http.StatusCreated, Data: created}, nil
}

func (h *Handler) GetTask(ctx *httpx.Context, _ httpx.NoBody) (map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	taskID := strings.TrimSpace(ctx.Param("slug"))
	if taskID == "" {
		return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "slug is required")
	}
	return h.svc.GetTask(ctx.Context(), resolveProjectScope(principal, projectID), taskID)
}

func (h *Handler) UpdateTask(ctx *httpx.Context, req updateTaskRequest) (map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	taskID := strings.TrimSpace(ctx.Param("taskId"))
	if taskID == "" {
		return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "taskId is required")
	}
	return h.svc.UpdateTask(ctx.Context(), resolveProjectScope(principal, projectID), taskID, principal.UserID, req)
}

func (h *Handler) UpdateTaskStatus(ctx *httpx.Context, req updateTaskStatusRequest) (map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	taskID := strings.TrimSpace(ctx.Param("taskId"))
	if taskID == "" {
		return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "taskId is required")
	}
	return h.svc.UpdateTaskStatus(ctx.Context(), resolveProjectScope(principal, projectID), taskID, principal.UserID, req)
}

func (h *Handler) ListFeedback(ctx *httpx.Context, _ httpx.NoBody) (map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	query, err := parseListQuery(ctx)
	if err != nil {
		return nil, err
	}
	query.Outcome = strings.TrimSpace(ctx.Query("outcome"))
	return h.svc.ListFeedback(ctx.Context(), resolveProjectScope(principal, projectID), query)
}

func (h *Handler) CreateFeedback(ctx *httpx.Context, req createFeedbackRequest) (httpx.Result[map[string]any], error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return httpx.Result[map[string]any]{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return httpx.Result[map[string]any]{}, err
	}
	created, err := h.svc.CreateFeedback(ctx.Context(), resolveProjectScope(principal, projectID), principal.UserID, req)
	if err != nil {
		return httpx.Result[map[string]any]{}, err
	}
	return httpx.Result[map[string]any]{Status: http.StatusCreated, Data: created}, nil
}

func (h *Handler) GetFeedback(ctx *httpx.Context, _ httpx.NoBody) (map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	feedbackID := strings.TrimSpace(ctx.Param("slug"))
	if feedbackID == "" {
		return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "slug is required")
	}
	return h.svc.GetFeedback(ctx.Context(), resolveProjectScope(principal, projectID), feedbackID)
}

func (h *Handler) UpdateFeedback(ctx *httpx.Context, req updateFeedbackRequest) (map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	feedbackID := strings.TrimSpace(ctx.Param("feedbackId"))
	if feedbackID == "" {
		return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "feedbackId is required")
	}
	return h.svc.UpdateFeedback(ctx.Context(), resolveProjectScope(principal, projectID), feedbackID, principal.UserID, req)
}

func parseListQuery(ctx *httpx.Context) (listQuery, error) {
	offset, err := parseOptionalIntQuery(ctx.Query("offset"), 0, "offset")
	if err != nil {
		return listQuery{}, err
	}
	limit, err := parseOptionalIntQuery(ctx.Query("limit"), 25, "limit")
	if err != nil {
		return listQuery{}, err
	}
	return listQuery{Offset: offset, Limit: normalizeLimit(limit)}, nil
}

func requireAuthenticatedPrincipal(ctx *httpx.Context) (auth.AuthContext, error) {
	if ctx == nil {
		return auth.AuthContext{}, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	}
	principal, ok := ctx.Auth()
	if !ok || strings.TrimSpace(principal.UserID) == "" {
		return auth.AuthContext{}, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	}
	return principal, nil
}

func requireProjectID(ctx *httpx.Context) (string, error) {
	if ctx == nil {
		return "", apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "projectId is required")
	}
	projectID := strings.TrimSpace(ctx.Param("projectId"))
	if projectID == "" {
		return "", apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "projectId is required")
	}
	return projectID, nil
}

func resolveProjectScope(principal auth.AuthContext, pathProjectID string) string {
	projectID := strings.TrimSpace(principal.ProjectID)
	if projectID != "" {
		return projectID
	}
	return strings.TrimSpace(pathProjectID)
}
