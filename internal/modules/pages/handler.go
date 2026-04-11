package pages

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

func (h *Handler) ListPages(ctx *httpx.Context, _ httpx.NoBody) ([]map[string]any, error) {
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
	return h.svc.ListPages(ctx.Context(), resolveProjectScope(principal, projectID), query)
}

func (h *Handler) CreatePage(ctx *httpx.Context, req createPageRequest) (httpx.Result[map[string]any], error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return httpx.Result[map[string]any]{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return httpx.Result[map[string]any]{}, err
	}
	created, err := h.svc.CreatePage(ctx.Context(), resolveProjectScope(principal, projectID), principal.UserID, req)
	if err != nil {
		return httpx.Result[map[string]any]{}, err
	}
	return httpx.Result[map[string]any]{Status: http.StatusCreated, Data: created}, nil
}

func (h *Handler) GetPage(ctx *httpx.Context, _ httpx.NoBody) (map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	pageID := strings.TrimSpace(ctx.Param("slug"))
	if pageID == "" {
		return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "slug is required")
	}
	return h.svc.GetPage(ctx.Context(), resolveProjectScope(principal, projectID), pageID)
}

func (h *Handler) UpdatePage(ctx *httpx.Context, req updatePageRequest) (map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	pageID := strings.TrimSpace(ctx.Param("pageId"))
	if pageID == "" {
		return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "pageId is required")
	}
	return h.svc.UpdatePage(ctx.Context(), resolveProjectScope(principal, projectID), pageID, principal.UserID, req)
}

func (h *Handler) RenamePage(ctx *httpx.Context, req renamePageRequest) (map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	pageID := strings.TrimSpace(ctx.Param("pageId"))
	if pageID == "" {
		return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "pageId is required")
	}
	return h.svc.RenamePage(ctx.Context(), resolveProjectScope(principal, projectID), pageID, principal.UserID, req)
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
		Status: strings.TrimSpace(ctx.Query("status")),
		Offset: offset,
		Limit:  normalizeLimit(limit),
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
