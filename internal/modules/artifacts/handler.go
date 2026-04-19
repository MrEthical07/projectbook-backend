package artifacts

import (
	"net/http"
	"strings"

	"github.com/MrEthical07/superapi/internal/core/auth"
	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/httpx"
	"github.com/MrEthical07/superapi/internal/core/pagination"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListStories(ctx *httpx.Context, _ httpx.NoBody) (StoryListResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return StoryListResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return StoryListResponse{}, err
	}
	query, err := parseListQuery(ctx)
	if err != nil {
		return StoryListResponse{}, err
	}
	query.Status = queryValue(ctx, "filter.status", "status")
	return h.svc.ListStories(ctx.Context(), resolveProjectScope(principal, projectID), query)
}

func (h *Handler) CreateStory(ctx *httpx.Context, req createStoryRequest) (httpx.Result[StoryListItem], error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return httpx.Result[StoryListItem]{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return httpx.Result[StoryListItem]{}, err
	}
	created, err := h.svc.CreateStory(ctx.Context(), resolveProjectScope(principal, projectID), principal.UserID, req)
	if err != nil {
		return httpx.Result[StoryListItem]{}, err
	}
	return httpx.Result[StoryListItem]{Status: http.StatusCreated, Data: created}, nil
}

func (h *Handler) GetStory(ctx *httpx.Context, _ httpx.NoBody) (StoryPageResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return StoryPageResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return StoryPageResponse{}, err
	}
	storyID := strings.TrimSpace(ctx.Param("storyId"))
	if storyID == "" {
		return StoryPageResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "storyId is required")
	}
	return h.svc.GetStory(ctx.Context(), resolveProjectScope(principal, projectID), storyID)
}

func (h *Handler) UpdateStory(ctx *httpx.Context, req updateStoryRequest) (StoryListItem, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return StoryListItem{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return StoryListItem{}, err
	}
	storyID := strings.TrimSpace(ctx.Param("storyId"))
	if storyID == "" {
		return StoryListItem{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "storyId is required")
	}
	return h.svc.UpdateStory(ctx.Context(), resolveProjectScope(principal, projectID), storyID, principal.UserID, req)
}

func (h *Handler) ListJourneys(ctx *httpx.Context, _ httpx.NoBody) (JourneyListResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return JourneyListResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return JourneyListResponse{}, err
	}
	query, err := parseListQuery(ctx)
	if err != nil {
		return JourneyListResponse{}, err
	}
	query.Status = queryValue(ctx, "filter.status", "status")
	return h.svc.ListJourneys(ctx.Context(), resolveProjectScope(principal, projectID), query)
}

func (h *Handler) CreateJourney(ctx *httpx.Context, req createJourneyRequest) (httpx.Result[JourneyListItem], error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return httpx.Result[JourneyListItem]{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return httpx.Result[JourneyListItem]{}, err
	}
	created, err := h.svc.CreateJourney(ctx.Context(), resolveProjectScope(principal, projectID), principal.UserID, req)
	if err != nil {
		return httpx.Result[JourneyListItem]{}, err
	}
	return httpx.Result[JourneyListItem]{Status: http.StatusCreated, Data: created}, nil
}

func (h *Handler) GetJourney(ctx *httpx.Context, _ httpx.NoBody) (JourneyPageResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return JourneyPageResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return JourneyPageResponse{}, err
	}
	journeyID := strings.TrimSpace(ctx.Param("journeyId"))
	if journeyID == "" {
		return JourneyPageResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "journeyId is required")
	}
	return h.svc.GetJourney(ctx.Context(), resolveProjectScope(principal, projectID), journeyID)
}

func (h *Handler) UpdateJourney(ctx *httpx.Context, req updateJourneyRequest) (JourneyListItem, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return JourneyListItem{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return JourneyListItem{}, err
	}
	journeyID := strings.TrimSpace(ctx.Param("journeyId"))
	if journeyID == "" {
		return JourneyListItem{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "journeyId is required")
	}
	return h.svc.UpdateJourney(ctx.Context(), resolveProjectScope(principal, projectID), journeyID, principal.UserID, req)
}

func (h *Handler) ListProblems(ctx *httpx.Context, _ httpx.NoBody) (ProblemListResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return ProblemListResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return ProblemListResponse{}, err
	}
	query, err := parseListQuery(ctx)
	if err != nil {
		return ProblemListResponse{}, err
	}
	query.Status = queryValue(ctx, "filter.status", "status")
	return h.svc.ListProblems(ctx.Context(), resolveProjectScope(principal, projectID), query)
}

func (h *Handler) CreateProblem(ctx *httpx.Context, req createProblemRequest) (httpx.Result[ProblemListItem], error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return httpx.Result[ProblemListItem]{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return httpx.Result[ProblemListItem]{}, err
	}
	created, err := h.svc.CreateProblem(ctx.Context(), resolveProjectScope(principal, projectID), principal.UserID, req)
	if err != nil {
		return httpx.Result[ProblemListItem]{}, err
	}
	return httpx.Result[ProblemListItem]{Status: http.StatusCreated, Data: created}, nil
}

func (h *Handler) GetProblem(ctx *httpx.Context, _ httpx.NoBody) (ProblemPageResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return ProblemPageResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return ProblemPageResponse{}, err
	}
	problemID := strings.TrimSpace(ctx.Param("problemId"))
	if problemID == "" {
		return ProblemPageResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "problemId is required")
	}
	return h.svc.GetProblem(ctx.Context(), resolveProjectScope(principal, projectID), problemID)
}

func (h *Handler) UpdateProblem(ctx *httpx.Context, req updateProblemRequest) (ProblemListItem, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return ProblemListItem{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return ProblemListItem{}, err
	}
	problemID := strings.TrimSpace(ctx.Param("problemId"))
	if problemID == "" {
		return ProblemListItem{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "problemId is required")
	}
	return h.svc.UpdateProblem(ctx.Context(), resolveProjectScope(principal, projectID), problemID, principal.UserID, req)
}

func (h *Handler) ListIdeas(ctx *httpx.Context, _ httpx.NoBody) (IdeaListResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return IdeaListResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return IdeaListResponse{}, err
	}
	query, err := parseListQuery(ctx)
	if err != nil {
		return IdeaListResponse{}, err
	}
	query.Status = queryValue(ctx, "filter.status", "status")
	return h.svc.ListIdeas(ctx.Context(), resolveProjectScope(principal, projectID), query)
}

func (h *Handler) CreateIdea(ctx *httpx.Context, req createIdeaRequest) (httpx.Result[IdeaListItem], error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return httpx.Result[IdeaListItem]{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return httpx.Result[IdeaListItem]{}, err
	}
	created, err := h.svc.CreateIdea(ctx.Context(), resolveProjectScope(principal, projectID), principal.UserID, req)
	if err != nil {
		return httpx.Result[IdeaListItem]{}, err
	}
	return httpx.Result[IdeaListItem]{Status: http.StatusCreated, Data: created}, nil
}

func (h *Handler) GetIdea(ctx *httpx.Context, _ httpx.NoBody) (IdeaPageResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return IdeaPageResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return IdeaPageResponse{}, err
	}
	ideaID := strings.TrimSpace(ctx.Param("ideaId"))
	if ideaID == "" {
		return IdeaPageResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "ideaId is required")
	}
	return h.svc.GetIdea(ctx.Context(), resolveProjectScope(principal, projectID), ideaID)
}

func (h *Handler) UpdateIdea(ctx *httpx.Context, req updateIdeaRequest) (IdeaListItem, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return IdeaListItem{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return IdeaListItem{}, err
	}
	ideaID := strings.TrimSpace(ctx.Param("ideaId"))
	if ideaID == "" {
		return IdeaListItem{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "ideaId is required")
	}
	return h.svc.UpdateIdea(ctx.Context(), resolveProjectScope(principal, projectID), ideaID, principal.UserID, req)
}
func (h *Handler) ListTasks(ctx *httpx.Context, _ httpx.NoBody) (TaskListResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return TaskListResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return TaskListResponse{}, err
	}
	query, err := parseListQuery(ctx)
	if err != nil {
		return TaskListResponse{}, err
	}
	query.Status = queryValue(ctx, "filter.status", "status")
	return h.svc.ListTasks(ctx.Context(), resolveProjectScope(principal, projectID), query)
}

func (h *Handler) CreateTask(ctx *httpx.Context, req createTaskRequest) (httpx.Result[TaskListItem], error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return httpx.Result[TaskListItem]{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return httpx.Result[TaskListItem]{}, err
	}
	created, err := h.svc.CreateTask(ctx.Context(), resolveProjectScope(principal, projectID), principal.UserID, req)
	if err != nil {
		return httpx.Result[TaskListItem]{}, err
	}
	return httpx.Result[TaskListItem]{Status: http.StatusCreated, Data: created}, nil
}

func (h *Handler) GetTask(ctx *httpx.Context, _ httpx.NoBody) (TaskPageResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return TaskPageResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return TaskPageResponse{}, err
	}
	taskID := strings.TrimSpace(ctx.Param("taskId"))
	if taskID == "" {
		return TaskPageResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "taskId is required")
	}
	return h.svc.GetTask(ctx.Context(), resolveProjectScope(principal, projectID), taskID)
}

func (h *Handler) UpdateTask(ctx *httpx.Context, req updateTaskRequest) (TaskListItem, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return TaskListItem{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return TaskListItem{}, err
	}
	taskID := strings.TrimSpace(ctx.Param("taskId"))
	if taskID == "" {
		return TaskListItem{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "taskId is required")
	}
	return h.svc.UpdateTask(ctx.Context(), resolveProjectScope(principal, projectID), taskID, principal.UserID, req)
}

func (h *Handler) ListFeedback(ctx *httpx.Context, _ httpx.NoBody) (FeedbackListResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return FeedbackListResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return FeedbackListResponse{}, err
	}
	query, err := parseListQuery(ctx)
	if err != nil {
		return FeedbackListResponse{}, err
	}
	query.Status = queryValue(ctx, "filter.status", "status")
	query.Outcome = queryValue(ctx, "filter.outcome", "outcome")
	return h.svc.ListFeedback(ctx.Context(), resolveProjectScope(principal, projectID), query)
}

func (h *Handler) CreateFeedback(ctx *httpx.Context, req createFeedbackRequest) (httpx.Result[FeedbackListItem], error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return httpx.Result[FeedbackListItem]{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return httpx.Result[FeedbackListItem]{}, err
	}
	created, err := h.svc.CreateFeedback(ctx.Context(), resolveProjectScope(principal, projectID), principal.UserID, req)
	if err != nil {
		return httpx.Result[FeedbackListItem]{}, err
	}
	return httpx.Result[FeedbackListItem]{Status: http.StatusCreated, Data: created}, nil
}

func (h *Handler) GetFeedback(ctx *httpx.Context, _ httpx.NoBody) (FeedbackPageResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return FeedbackPageResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return FeedbackPageResponse{}, err
	}
	feedbackID := strings.TrimSpace(ctx.Param("feedbackId"))
	if feedbackID == "" {
		return FeedbackPageResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "feedbackId is required")
	}
	return h.svc.GetFeedback(ctx.Context(), resolveProjectScope(principal, projectID), feedbackID)
}

func (h *Handler) UpdateFeedback(ctx *httpx.Context, req updateFeedbackRequest) (FeedbackListItem, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return FeedbackListItem{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return FeedbackListItem{}, err
	}
	feedbackID := strings.TrimSpace(ctx.Param("feedbackId"))
	if feedbackID == "" {
		return FeedbackListItem{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "feedbackId is required")
	}
	return h.svc.UpdateFeedback(ctx.Context(), resolveProjectScope(principal, projectID), feedbackID, principal.UserID, req)
}

func parseListQuery(ctx *httpx.Context) (listQuery, error) {
	offset := 0
	if cursor := queryValue(ctx, "pagination.cursor", "cursor"); cursor != "" {
		decodedOffset, decodeErr := pagination.DecodeOffsetCursor(cursor)
		if decodeErr != nil {
			return listQuery{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "cursor is invalid")
		}
		offset = decodedOffset
	}
	limit, err := parseOptionalIntQuery(queryValue(ctx, "pagination.limit", "limit"), 20, "limit")
	if err != nil {
		return listQuery{}, err
	}
	return listQuery{Offset: offset, Limit: normalizeLimit(limit)}, nil
}

func queryValue(ctx *httpx.Context, keys ...string) string {
	if ctx == nil {
		return ""
	}
	for _, key := range keys {
		if value := strings.TrimSpace(ctx.Query(key)); value != "" {
			return value
		}
	}
	return ""
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
