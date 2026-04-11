package activity

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

func (h *Handler) ListProjectActivity(ctx *httpx.Context, _ httpx.NoBody) ([]map[string]any, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return nil, err
	}
	limit, err := parseLimit(ctx.Query("limit"))
	if err != nil {
		return nil, err
	}
	return h.svc.ListProjectActivity(ctx.Context(), resolveProjectScope(principal, projectID), listQuery{Limit: limit})
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
