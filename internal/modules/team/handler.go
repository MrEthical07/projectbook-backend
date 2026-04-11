package team

import (
	"net/http"
	"strings"

	"github.com/MrEthical07/superapi/internal/core/auth"
	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/httpx"
)

// Handler contains HTTP transport handlers for team routes.
type Handler struct {
	svc Service
}

// NewHandler constructs team transport handlers.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListMembers(ctx *httpx.Context, _ httpx.NoBody) (teamMembersResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return teamMembersResponse{}, err
	}
	pathProjectID, err := requireProjectID(ctx)
	if err != nil {
		return teamMembersResponse{}, err
	}
	return h.svc.ListMembers(ctx.Context(), resolveProjectScope(principal, pathProjectID))
}

func (h *Handler) ListRoles(ctx *httpx.Context, _ httpx.NoBody) (teamRolesResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return teamRolesResponse{}, err
	}
	pathProjectID, err := requireProjectID(ctx)
	if err != nil {
		return teamRolesResponse{}, err
	}
	return h.svc.ListRoles(ctx.Context(), resolveProjectScope(principal, pathProjectID))
}

func (h *Handler) CreateInvite(ctx *httpx.Context, req createInviteRequest) (httpx.Result[createInviteResponse], error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return httpx.Result[createInviteResponse]{}, err
	}
	pathProjectID, err := requireProjectID(ctx)
	if err != nil {
		return httpx.Result[createInviteResponse]{}, err
	}

	response, err := h.svc.CreateInvite(ctx.Context(), resolveProjectScope(principal, pathProjectID), principal.UserID, req)
	if err != nil {
		return httpx.Result[createInviteResponse]{}, err
	}

	return httpx.Result[createInviteResponse]{
		Status: http.StatusCreated,
		Data:   response,
	}, nil
}

func (h *Handler) BatchInvites(ctx *httpx.Context, req batchInviteRequest) (httpx.Result[batchInviteResponse], error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return httpx.Result[batchInviteResponse]{}, err
	}
	pathProjectID, err := requireProjectID(ctx)
	if err != nil {
		return httpx.Result[batchInviteResponse]{}, err
	}

	response, partial, err := h.svc.BatchInvites(ctx.Context(), resolveProjectScope(principal, pathProjectID), principal.UserID, req)
	if err != nil {
		return httpx.Result[batchInviteResponse]{}, err
	}

	status := http.StatusCreated
	if partial {
		status = http.StatusMultiStatus
	}

	return httpx.Result[batchInviteResponse]{
		Status: status,
		Data:   response,
	}, nil
}

func (h *Handler) CancelInvite(ctx *httpx.Context, _ httpx.NoBody) (cancelInviteResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return cancelInviteResponse{}, err
	}
	pathProjectID, err := requireProjectID(ctx)
	if err != nil {
		return cancelInviteResponse{}, err
	}

	email := normalizeEmail(ctx.Param("email"))
	if email == "" {
		return cancelInviteResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "email is required")
	}

	return h.svc.CancelInvite(ctx.Context(), resolveProjectScope(principal, pathProjectID), email)
}

func (h *Handler) UpdateMemberPermissions(ctx *httpx.Context, req updateMemberPermissionsRequest) (updateMemberPermissionsResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return updateMemberPermissionsResponse{}, err
	}
	pathProjectID, err := requireProjectID(ctx)
	if err != nil {
		return updateMemberPermissionsResponse{}, err
	}

	memberID := strings.TrimSpace(ctx.Param("memberId"))
	if memberID == "" {
		return updateMemberPermissionsResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "memberId is required")
	}

	return h.svc.UpdateMemberPermissions(ctx.Context(), resolveProjectScope(principal, pathProjectID), memberID, req)
}

func (h *Handler) UpdateRolePermissions(ctx *httpx.Context, req updateRolePermissionsRequest) (updateRolePermissionsResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return updateRolePermissionsResponse{}, err
	}
	pathProjectID, err := requireProjectID(ctx)
	if err != nil {
		return updateRolePermissionsResponse{}, err
	}

	rolePath := strings.TrimSpace(ctx.Param("role"))
	canonicalRole, ok := canonicalRoleFromSlug(rolePath)
	if !ok {
		return updateRolePermissionsResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "role path parameter is invalid")
	}

	return h.svc.UpdateRolePermissions(ctx.Context(), resolveProjectScope(principal, pathProjectID), canonicalRole, principal.UserID, req)
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
