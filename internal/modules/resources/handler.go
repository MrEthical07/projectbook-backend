package resources

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

func (h *Handler) ListResources(ctx *httpx.Context, _ httpx.NoBody) (ListResourcesResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return ListResourcesResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return ListResourcesResponse{}, err
	}
	query, err := parseListQuery(ctx)
	if err != nil {
		return ListResourcesResponse{}, err
	}
	return h.svc.ListResources(ctx.Context(), resolveProjectScope(principal, projectID), query)
}

func (h *Handler) CreateResource(ctx *httpx.Context, req createResourceRequest) (httpx.Result[ResourceListItem], error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return httpx.Result[ResourceListItem]{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return httpx.Result[ResourceListItem]{}, err
	}
	created, err := h.svc.CreateResource(ctx.Context(), resolveProjectScope(principal, projectID), principal.UserID, req)
	if err != nil {
		return httpx.Result[ResourceListItem]{}, err
	}
	return httpx.Result[ResourceListItem]{Status: http.StatusCreated, Data: created}, nil
}

func (h *Handler) GetResource(ctx *httpx.Context, _ httpx.NoBody) (GetResourceResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return GetResourceResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return GetResourceResponse{}, err
	}
	resourceID := strings.TrimSpace(ctx.Param("resourceId"))
	if resourceID == "" {
		return GetResourceResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "resourceId is required")
	}
	return h.svc.GetResource(ctx.Context(), resolveProjectScope(principal, projectID), resourceID)
}

func (h *Handler) UpdateResource(ctx *httpx.Context, req updateResourceRequest) (ResourceListItem, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return ResourceListItem{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return ResourceListItem{}, err
	}
	resourceID := strings.TrimSpace(ctx.Param("resourceId"))
	if resourceID == "" {
		return ResourceListItem{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "resourceId is required")
	}
	return h.svc.UpdateResource(ctx.Context(), resolveProjectScope(principal, projectID), resourceID, principal.UserID, req)
}

func parseListQuery(ctx *httpx.Context) (listQuery, error) {
	offset := 0
	if cursor := queryValue(ctx, "pagination.cursor", "cursor"); cursor != "" {
		decodedOffset, err := pagination.DecodeOffsetCursor(cursor)
		if err != nil {
			return listQuery{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "cursor is invalid")
		}
		offset = decodedOffset
	}

	limit, err := parseOptionalIntQuery(queryValue(ctx, "pagination.limit", "limit"), 20, "limit")
	if err != nil {
		return listQuery{}, err
	}
	return listQuery{
		Status:  queryValue(ctx, "filter.status", "status"),
		DocType: queryValue(ctx, "filter.docType", "docType"),
		Sort:    normalizeSort(queryValue(ctx, "sorting.sort", "sort")),
		Order:   normalizeOrder(queryValue(ctx, "sorting.order", "order")),
		Offset:  offset,
		Limit:   normalizeLimit(limit),
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
