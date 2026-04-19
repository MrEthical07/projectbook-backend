package pages

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

func (h *Handler) ListPages(ctx *httpx.Context, _ httpx.NoBody) (ListPagesResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return ListPagesResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return ListPagesResponse{}, err
	}
	query, err := parseListQuery(ctx)
	if err != nil {
		return ListPagesResponse{}, err
	}
	return h.svc.ListPages(ctx.Context(), resolveProjectScope(principal, projectID), query)
}

func (h *Handler) CreatePage(ctx *httpx.Context, req createPageRequest) (httpx.Result[PageListItem], error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return httpx.Result[PageListItem]{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return httpx.Result[PageListItem]{}, err
	}
	created, err := h.svc.CreatePage(ctx.Context(), resolveProjectScope(principal, projectID), principal.UserID, req)
	if err != nil {
		return httpx.Result[PageListItem]{}, err
	}
	return httpx.Result[PageListItem]{Status: http.StatusCreated, Data: created}, nil
}

func (h *Handler) GetPage(ctx *httpx.Context, _ httpx.NoBody) (GetPageResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return GetPageResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return GetPageResponse{}, err
	}
	pageID := strings.TrimSpace(ctx.Param("pageId"))
	if pageID == "" {
		return GetPageResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "pageId is required")
	}
	return h.svc.GetPage(ctx.Context(), resolveProjectScope(principal, projectID), pageID)
}

func (h *Handler) UpdatePage(ctx *httpx.Context, req updatePageRequest) (PageListItem, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return PageListItem{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return PageListItem{}, err
	}
	pageID := strings.TrimSpace(ctx.Param("pageId"))
	if pageID == "" {
		return PageListItem{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "pageId is required")
	}
	return h.svc.UpdatePage(ctx.Context(), resolveProjectScope(principal, projectID), pageID, principal.UserID, req)
}

func (h *Handler) RenamePage(ctx *httpx.Context, req renamePageRequest) (RenamePageResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return RenamePageResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return RenamePageResponse{}, err
	}
	pageID := strings.TrimSpace(ctx.Param("pageId"))
	if pageID == "" {
		return RenamePageResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "pageId is required")
	}
	return h.svc.RenamePage(ctx.Context(), resolveProjectScope(principal, projectID), pageID, principal.UserID, req)
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
	return listQuery{
		Status: queryValue(ctx, "filter.status", "status"),
		Offset: offset,
		Limit:  normalizeLimit(limit),
	}, nil
}

func queryValue(ctx *httpx.Context, keys ...string) string {
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
