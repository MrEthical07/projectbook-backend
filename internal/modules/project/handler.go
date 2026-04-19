package project

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/MrEthical07/superapi/internal/core/auth"
	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/httpx"
	"github.com/MrEthical07/superapi/internal/core/rbac"
)

// Handler contains HTTP transport handlers for project routes.
type Handler struct {
	svc Service
}

// NewHandler constructs project transport handlers.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Dashboard(ctx *httpx.Context, _ httpx.NoBody) (projectDashboardResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return projectDashboardResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return projectDashboardResponse{}, err
	}
	return h.svc.Dashboard(ctx.Context(), principal.UserID, projectID)
}

func (h *Handler) DashboardSummary(ctx *httpx.Context, _ httpx.NoBody) (projectDashboardSummaryResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return projectDashboardSummaryResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return projectDashboardSummaryResponse{}, err
	}
	return h.svc.DashboardSummary(ctx.Context(), principal.UserID, projectID)
}

func (h *Handler) DashboardMyWork(ctx *httpx.Context, _ httpx.NoBody) (projectDashboardMyWorkResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return projectDashboardMyWorkResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return projectDashboardMyWorkResponse{}, err
	}
	return h.svc.DashboardMyWork(ctx.Context(), principal.UserID, projectID)
}

func (h *Handler) DashboardEvents(ctx *httpx.Context, _ httpx.NoBody) (projectDashboardEventsResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return projectDashboardEventsResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return projectDashboardEventsResponse{}, err
	}
	return h.svc.DashboardEvents(ctx.Context(), principal.UserID, projectID)
}

func (h *Handler) DashboardActivity(ctx *httpx.Context, _ httpx.NoBody) (projectDashboardActivityResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return projectDashboardActivityResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return projectDashboardActivityResponse{}, err
	}
	return h.svc.DashboardActivity(ctx.Context(), principal.UserID, projectID)
}

func (h *Handler) Overview(ctx *httpx.Context, _ httpx.NoBody) (projectOverviewResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return projectOverviewResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return projectOverviewResponse{}, err
	}
	return h.svc.Overview(ctx.Context(), principal.UserID, projectID)
}

func (h *Handler) Search(ctx *httpx.Context, _ httpx.NoBody) (projectSearchResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return projectSearchResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return projectSearchResponse{}, err
	}
	query := strings.TrimSpace(ctx.Query("q"))
	if query == "" {
		return projectSearchResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "q is required")
	}

	limit := 0
	limitRaw := strings.TrimSpace(ctx.Query("limit"))
	if limitRaw != "" {
		parsedLimit, parseErr := strconv.Atoi(limitRaw)
		if parseErr != nil || parsedLimit <= 0 {
			return projectSearchResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "limit must be a positive integer")
		}
		limit = parsedLimit
	}

	return h.svc.Search(ctx.Context(), principal.UserID, projectID, query, limit)
}

func (h *Handler) Access(ctx *httpx.Context, _ httpx.NoBody) (projectAccessResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return projectAccessResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return projectAccessResponse{}, err
	}
	return h.svc.Access(ctx.Context(), principal.UserID, projectID, principal.Role, principal.PermissionMask)
}

func (h *Handler) GetSettings(ctx *httpx.Context, _ httpx.NoBody) (projectSettingsResponse, error) {
	if _, err := requireAuthenticatedPrincipal(ctx); err != nil {
		return projectSettingsResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return projectSettingsResponse{}, err
	}
	return h.svc.GetSettings(ctx.Context(), projectID)
}

func (h *Handler) UpdateSettings(ctx *httpx.Context, req updateProjectSettingsRequest) (projectUpdateSettingsResponse, error) {
	if _, err := requireAuthenticatedPrincipal(ctx); err != nil {
		return projectUpdateSettingsResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return projectUpdateSettingsResponse{}, err
	}
	return h.svc.UpdateSettings(ctx.Context(), projectID, req)
}

func (h *Handler) Archive(ctx *httpx.Context, _ httpx.NoBody) (projectArchiveResponse, error) {
	if _, err := requireAuthenticatedPrincipal(ctx); err != nil {
		return projectArchiveResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return projectArchiveResponse{}, err
	}
	return h.svc.Archive(ctx.Context(), projectID)
}

func (h *Handler) Delete(ctx *httpx.Context, _ httpx.NoBody) (projectDeleteResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return projectDeleteResponse{}, err
	}
	role := strings.TrimSpace(principal.Role)
	if !strings.EqualFold(role, rbac.RoleOwner) && !strings.EqualFold(role, rbac.RoleAdmin) {
		return projectDeleteResponse{}, apperr.New(apperr.CodeForbidden, http.StatusForbidden, "only Owner or Admin can delete project")
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return projectDeleteResponse{}, err
	}
	return h.svc.Delete(ctx.Context(), projectID)
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
