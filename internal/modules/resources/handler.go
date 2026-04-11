package resources

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

func (h *Handler) ListResources(ctx *httpx.Context, _ httpx.NoBody) (map[string]any, error) {
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
	return h.svc.ListResources(ctx.Context(), resolveProjectScope(principal, projectID), query)
}

func (h *Handler) CreateResource(ctx *httpx.Context, req createResourceRequest) (httpx.Result[map[string]any], error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return httpx.Result[map[string]any]{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return httpx.Result[map[string]any]{}, err
	}
	created, err := h.svc.CreateResource(ctx.Context(), resolveProjectScope(principal, projectID), principal.UserID, req)
	if err != nil {
		return httpx.Result[map[string]any]{}, err
	}
	return httpx.Result[map[string]any]{Status: http.StatusCreated, Data: created}, nil
}

func (h *Handler) GetResource(ctx *httpx.Context, _ httpx.NoBody) (map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	resourceID := strings.TrimSpace(ctx.Param("resourceId"))
	if resourceID == "" {
		return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "resourceId is required")
	}
	return h.svc.GetResource(ctx.Context(), resolveProjectScope(principal, projectID), resourceID)
}

func (h *Handler) UpdateResource(ctx *httpx.Context, req updateResourceRequest) (map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	resourceID := strings.TrimSpace(ctx.Param("resourceId"))
	if resourceID == "" {
		return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "resourceId is required")
	}
	return h.svc.UpdateResource(ctx.Context(), resolveProjectScope(principal, projectID), resourceID, principal.UserID, req)
}

func (h *Handler) UpdateResourceStatus(ctx *httpx.Context, req updateResourceStatusRequest) (map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	resourceID := strings.TrimSpace(ctx.Param("resourceId"))
	if resourceID == "" {
		return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "resourceId is required")
	}
	return h.svc.UpdateResourceStatus(ctx.Context(), resolveProjectScope(principal, projectID), resourceID, principal.UserID, req)
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
	return listQuery{
		Status:  strings.TrimSpace(ctx.Query("status")),
		DocType: strings.TrimSpace(ctx.Query("docType")),
		Sort:    normalizeSort(ctx.Query("sort")),
		Order:   normalizeOrder(ctx.Query("order")),
		Offset:  offset,
		Limit:   normalizeLimit(limit),
	}, nil
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
