package sidebar

import (
	"net/http"
	"strings"

	"github.com/MrEthical07/superapi/internal/core/auth"
	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/httpx"
	"github.com/MrEthical07/superapi/internal/core/rbac"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateSidebarArtifact(ctx *httpx.Context, req createSidebarArtifactRequest) (httpx.Result[SidebarArtifactResponse], error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return httpx.Result[SidebarArtifactResponse]{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return httpx.Result[SidebarArtifactResponse]{}, err
	}
	prefix := normalizePrefix(req.Prefix)
	if !hasPrefixPermission(principal.PermissionMask, prefix, operationCreate) {
		return httpx.Result[SidebarArtifactResponse]{}, apperr.New(apperr.CodeForbidden, http.StatusForbidden, "forbidden")
	}
	created, err := h.svc.CreateSidebarArtifact(ctx.Context(), resolveProjectScope(principal, projectID), principal.UserID, req)
	if err != nil {
		return httpx.Result[SidebarArtifactResponse]{}, err
	}
	return httpx.Result[SidebarArtifactResponse]{Status: http.StatusCreated, Data: created}, nil
}

func (h *Handler) RenameSidebarArtifact(ctx *httpx.Context, req renameSidebarArtifactRequest) (SidebarArtifactResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return SidebarArtifactResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return SidebarArtifactResponse{}, err
	}
	artifactID := strings.TrimSpace(ctx.Param("artifactId"))
	if artifactID == "" {
		return SidebarArtifactResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "artifactId is required")
	}
	prefix := normalizePrefix(req.Prefix)
	if !hasPrefixPermission(principal.PermissionMask, prefix, operationEdit) {
		return SidebarArtifactResponse{}, apperr.New(apperr.CodeForbidden, http.StatusForbidden, "forbidden")
	}
	return h.svc.RenameSidebarArtifact(ctx.Context(), resolveProjectScope(principal, projectID), artifactID, principal.UserID, req)
}

func (h *Handler) DeleteSidebarArtifact(ctx *httpx.Context, req deleteSidebarArtifactRequest) (SidebarDeleteResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return SidebarDeleteResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return SidebarDeleteResponse{}, err
	}
	artifactID := strings.TrimSpace(ctx.Param("artifactId"))
	if artifactID == "" {
		return SidebarDeleteResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "artifactId is required")
	}
	req.Prefix = strings.TrimSpace(ctx.Query("prefix"))
	prefix := normalizePrefix(req.Prefix)
	if !hasPrefixPermission(principal.PermissionMask, prefix, operationDelete) {
		return SidebarDeleteResponse{}, apperr.New(apperr.CodeForbidden, http.StatusForbidden, "forbidden")
	}
	return h.svc.DeleteSidebarArtifact(ctx.Context(), resolveProjectScope(principal, projectID), artifactID, principal.UserID, req)
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

type permissionOperation string

const (
	operationCreate permissionOperation = "create"
	operationEdit   permissionOperation = "edit"
	operationDelete permissionOperation = "delete"
)

func hasPrefixPermission(mask uint64, prefix string, op permissionOperation) bool {
	var perm uint64
	switch normalizePrefix(prefix) {
	case prefixStories, prefixJourneys:
		perm = permissionForOperation(op, rbac.PermStoryCreate, rbac.PermStoryEdit, rbac.PermStoryDelete)
	case prefixProblemStatement:
		perm = permissionForOperation(op, rbac.PermProblemCreate, rbac.PermProblemEdit, rbac.PermProblemDelete)
	case prefixIdeas:
		perm = permissionForOperation(op, rbac.PermIdeaCreate, rbac.PermIdeaEdit, rbac.PermIdeaDelete)
	case prefixTasks:
		perm = permissionForOperation(op, rbac.PermTaskCreate, rbac.PermTaskEdit, rbac.PermTaskDelete)
	case prefixFeedback:
		perm = permissionForOperation(op, rbac.PermFeedbackCreate, rbac.PermFeedbackEdit, rbac.PermFeedbackDelete)
	case prefixPages:
		perm = permissionForOperation(op, rbac.PermPageCreate, rbac.PermPageEdit, rbac.PermPageDelete)
	default:
		return false
	}
	return mask&perm != 0
}

func permissionForOperation(op permissionOperation, createPerm, editPerm, deletePerm uint64) uint64 {
	switch op {
	case operationCreate:
		return createPerm
	case operationEdit:
		return editPerm
	case operationDelete:
		return deletePerm
	default:
		return 0
	}
}
