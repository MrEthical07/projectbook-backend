package project

import (
	"net/http"
	"strings"

	"github.com/MrEthical07/superapi/internal/core/auth"
	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/httpx"
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

func (h *Handler) Sidebar(ctx *httpx.Context, _ httpx.NoBody) (projectSidebarResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return projectSidebarResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return projectSidebarResponse{}, err
	}
	return h.svc.Sidebar(ctx.Context(), principal.UserID, projectID)
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
	if _, err := requireAuthenticatedPrincipal(ctx); err != nil {
		return projectDeleteResponse{}, err
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
