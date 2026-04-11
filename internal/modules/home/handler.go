package home

import (
	"net/http"
	"strconv"
	"strings"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/httpx"
)

// Handler contains HTTP transport handlers for home routes.
type Handler struct {
	svc Service
}

// NewHandler constructs home transport handlers.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Dashboard(ctx *httpx.Context, _ httpx.NoBody) (homeDashboardResponse, error) {
	userID, err := requireAuthenticatedUser(ctx)
	if err != nil {
		return homeDashboardResponse{}, err
	}
	return h.svc.Dashboard(ctx.Context(), userID)
}

func (h *Handler) ListProjects(ctx *httpx.Context, _ httpx.NoBody) ([]homeProject, error) {
	userID, err := requireAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	limit, err := parseOptionalIntQuery(ctx, "limit", 100)
	if err != nil {
		return nil, err
	}
	offset, err := parseOptionalIntQuery(ctx, "offset", 0)
	if err != nil {
		return nil, err
	}

	return h.svc.ListProjects(ctx.Context(), userID, limit, offset)
}

func (h *Handler) CreateProject(ctx *httpx.Context, req createProjectRequest) (httpx.Result[projectCreationResponse], error) {
	userID, err := requireAuthenticatedUser(ctx)
	if err != nil {
		return httpx.Result[projectCreationResponse]{}, err
	}

	out, err := h.svc.CreateProject(ctx.Context(), userID, req)
	if err != nil {
		return httpx.Result[projectCreationResponse]{}, err
	}

	return httpx.Result[projectCreationResponse]{
		Status: http.StatusCreated,
		Data:   out,
	}, nil
}

func (h *Handler) ProjectReference(ctx *httpx.Context, _ httpx.NoBody) (projectReferenceResponse, error) {
	userID, err := requireAuthenticatedUser(ctx)
	if err != nil {
		return projectReferenceResponse{}, err
	}

	return h.svc.ProjectReference(ctx.Context(), userID)
}

func (h *Handler) ListInvites(ctx *httpx.Context, _ httpx.NoBody) ([]homeInvite, error) {
	userID, err := requireAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	return h.svc.ListInvites(ctx.Context(), userID)
}

func (h *Handler) AcceptInvite(ctx *httpx.Context, _ httpx.NoBody) (inviteAcceptResponse, error) {
	userID, err := requireAuthenticatedUser(ctx)
	if err != nil {
		return inviteAcceptResponse{}, err
	}

	inviteID := strings.TrimSpace(ctx.Param("inviteId"))
	if inviteID == "" {
		return inviteAcceptResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "inviteId is required")
	}

	return h.svc.AcceptInvite(ctx.Context(), userID, inviteID)
}

func (h *Handler) DeclineInvite(ctx *httpx.Context, _ httpx.NoBody) (inviteDeclineResponse, error) {
	userID, err := requireAuthenticatedUser(ctx)
	if err != nil {
		return inviteDeclineResponse{}, err
	}

	inviteID := strings.TrimSpace(ctx.Param("inviteId"))
	if inviteID == "" {
		return inviteDeclineResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "inviteId is required")
	}

	return h.svc.DeclineInvite(ctx.Context(), userID, inviteID)
}

func (h *Handler) ListNotifications(ctx *httpx.Context, _ httpx.NoBody) ([]homeNotification, error) {
	userID, err := requireAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	limit, err := parseOptionalIntQuery(ctx, "limit", 50)
	if err != nil {
		return nil, err
	}

	return h.svc.ListNotifications(ctx.Context(), userID, limit)
}

func (h *Handler) ListActivity(ctx *httpx.Context, _ httpx.NoBody) ([]homeActivityItem, error) {
	userID, err := requireAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	limit, err := parseOptionalIntQuery(ctx, "limit", 50)
	if err != nil {
		return nil, err
	}

	filter := activityFilter{
		Limit:     limit,
		Type:      strings.TrimSpace(ctx.Query("type")),
		ProjectID: strings.TrimSpace(ctx.Query("projectId")),
	}

	return h.svc.ListActivity(ctx.Context(), userID, filter)
}

func (h *Handler) DashboardActivity(ctx *httpx.Context, _ httpx.NoBody) ([]dashboardActivityItem, error) {
	userID, err := requireAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	limit, err := parseOptionalIntQuery(ctx, "limit", 10)
	if err != nil {
		return nil, err
	}

	return h.svc.ListDashboardActivity(ctx.Context(), userID, limit)
}

func (h *Handler) GetAccountSettings(ctx *httpx.Context, _ httpx.NoBody) (homeAccountSettingsResponse, error) {
	userID, err := requireAuthenticatedUser(ctx)
	if err != nil {
		return homeAccountSettingsResponse{}, err
	}

	return h.svc.GetAccountSettings(ctx.Context(), userID)
}

func (h *Handler) UpdateAccountSettings(ctx *httpx.Context, req updateAccountRequest) (updateAccountResponse, error) {
	userID, err := requireAuthenticatedUser(ctx)
	if err != nil {
		return updateAccountResponse{}, err
	}

	return h.svc.UpdateAccountSettings(ctx.Context(), userID, req)
}

func (h *Handler) Docs(ctx *httpx.Context, _ httpx.NoBody) (docsResponse, error) {
	userID, err := requireAuthenticatedUser(ctx)
	if err != nil {
		return docsResponse{}, err
	}

	return h.svc.Docs(ctx.Context(), userID)
}

func requireAuthenticatedUser(ctx *httpx.Context) (string, error) {
	if ctx == nil {
		return "", apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	}
	principal, ok := ctx.Auth()
	if !ok || strings.TrimSpace(principal.UserID) == "" {
		return "", apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	}
	return strings.TrimSpace(principal.UserID), nil
}

func parseOptionalIntQuery(ctx *httpx.Context, name string, fallback int) (int, error) {
	if ctx == nil {
		return fallback, nil
	}
	value := strings.TrimSpace(ctx.Query(name))
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, name+" must be an integer")
	}
	if parsed < 0 {
		return 0, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, name+" must be non-negative")
	}
	return parsed, nil
}
