package resources

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/storage"
)

type Service interface {
	ListResources(ctx context.Context, projectID string, query listQuery) (ListResourcesResponse, error)
	CreateResource(ctx context.Context, projectID, actorUserID string, req createResourceRequest) (ResourceListItem, error)
	GetResource(ctx context.Context, projectID, resourceID string) (GetResourceResponse, error)
	UpdateResource(ctx context.Context, projectID, resourceID, actorUserID string, req updateResourceRequest) (ResourceListItem, error)
	UpdateResourceStatus(ctx context.Context, projectID, resourceID, actorUserID string, req updateResourceStatusRequest) (UpdateResourceStatusResponse, error)
}

type service struct {
	store storage.RelationalStore
	repo  Repo
}

var resourceImmutableStatuses = map[string]struct{}{
	"Archived": {},
}

func NewService(store storage.RelationalStore, repo Repo) Service {
	return &service{store: store, repo: repo}
}

func (s *service) ListResources(ctx context.Context, projectID string, query listQuery) (ListResourcesResponse, error) {
	rows, err := s.repo.ListResources(ctx, projectID, query)
	if err != nil {
		return ListResourcesResponse{}, mapServiceError("list resources", err)
	}
	return decodeListResourcesResponse(rows), nil
}

func (s *service) CreateResource(ctx context.Context, projectID, actorUserID string, req createResourceRequest) (ResourceListItem, error) {
	if err := req.Validate(); err != nil {
		return ResourceListItem{}, err
	}
	if strings.TrimSpace(actorUserID) == "" {
		return ResourceListItem{}, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	}
	var created map[string]any
	err := s.store.WithTx(ctx, func(txCtx context.Context) error {
		result, createErr := s.repo.CreateResource(txCtx, projectID, actorUserID, req.Name, req.DocType)
		if createErr != nil {
			return createErr
		}
		created = result
		return nil
	})
	if err != nil {
		return ResourceListItem{}, mapServiceError("create resource", err)
	}
	return decodeResourceListItem(created), nil
}

func (s *service) GetResource(ctx context.Context, projectID, resourceID string) (GetResourceResponse, error) {
	item, err := s.repo.GetResource(ctx, projectID, resourceID)
	if err != nil {
		return GetResourceResponse{}, mapServiceError("get resource", err)
	}
	return decodeGetResourceResponse(item), nil
}

func (s *service) UpdateResource(ctx context.Context, projectID, resourceID, actorUserID string, req updateResourceRequest) (ResourceListItem, error) {
	if err := req.Validate(); err != nil {
		return ResourceListItem{}, err
	}
	current, err := s.repo.GetResource(ctx, projectID, resourceID)
	if err != nil {
		return ResourceListItem{}, mapServiceError("load resource before update", err)
	}
	from := nestedString(current, "resource", "status")
	if err := enforceArchiveOnlyForImmutableUpdate("resource", from, req.State, resourceImmutableStatuses); err != nil {
		return ResourceListItem{}, err
	}
	if status := toString(req.State["status"]); status != "" {
		if !isAllowedTransition(from, status, map[string]map[string]struct{}{
			"Active":   {"Active": {}, "Archived": {}},
			"Archived": {"Archived": {}, "Active": {}},
		}) {
			return ResourceListItem{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid resource status transition")
		}
	}
	var updated map[string]any
	err = s.store.WithTx(ctx, func(txCtx context.Context) error {
		result, updateErr := s.repo.UpdateResource(txCtx, projectID, resourceID, actorUserID, req.State)
		if updateErr != nil {
			return updateErr
		}
		updated = result
		return nil
	})
	if err != nil {
		return ResourceListItem{}, mapServiceError("update resource", err)
	}
	return decodeResourceListItem(updated), nil
}

func (s *service) UpdateResourceStatus(ctx context.Context, projectID, resourceID, actorUserID string, req updateResourceStatusRequest) (UpdateResourceStatusResponse, error) {
	if err := req.Validate(); err != nil {
		return UpdateResourceStatusResponse{}, err
	}
	current, err := s.repo.GetResource(ctx, projectID, resourceID)
	if err != nil {
		return UpdateResourceStatusResponse{}, mapServiceError("load resource before status update", err)
	}
	from := nestedString(current, "resource", "status")
	if err := enforceArchiveOnlyForImmutableStatusChange("resource", from, req.Status, resourceImmutableStatuses); err != nil {
		return UpdateResourceStatusResponse{}, err
	}
	if !isAllowedTransition(from, req.Status, map[string]map[string]struct{}{
		"Active":   {"Active": {}, "Archived": {}},
		"Archived": {"Archived": {}, "Active": {}},
	}) {
		return UpdateResourceStatusResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid resource status transition")
	}
	var updated map[string]any
	err = s.store.WithTx(ctx, func(txCtx context.Context) error {
		result, updateErr := s.repo.UpdateResourceStatus(txCtx, projectID, resourceID, req.Status, actorUserID)
		if updateErr != nil {
			return updateErr
		}
		updated = result
		return nil
	})
	if err != nil {
		return UpdateResourceStatusResponse{}, mapServiceError("update resource status", err)
	}
	return UpdateResourceStatusResponse{
		ID:          toString(updated["id"]),
		Status:      toString(updated["status"]),
		LastUpdated: toString(updated["lastUpdated"]),
	}, nil
}

func mapServiceError(action string, err error) error {
	if err == nil {
		return nil
	}
	if ae, ok := apperr.AsAppError(err); ok {
		return ae
	}
	return apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "failed to process resources request"), fmt.Errorf("%s: %w", action, err))
}

func nestedString(payload map[string]any, key, field string) string {
	obj := toMap(payload[key])
	if obj == nil {
		return ""
	}
	return toString(obj[field])
}

func enforceArchiveOnlyForImmutableUpdate(entity, from string, patch map[string]any, immutableStatuses map[string]struct{}) error {
	if !isImmutableStatus(from, immutableStatuses) {
		return nil
	}
	if isArchiveOnlyPatch(patch) {
		return nil
	}
	if isRestoreOnlyPatch(from, patch) {
		return nil
	}
	return immutableStateError(entity, from)
}

func enforceArchiveOnlyForImmutableStatusChange(entity, from, to string, immutableStatuses map[string]struct{}) error {
	if !isImmutableStatus(from, immutableStatuses) {
		return nil
	}
	if strings.EqualFold(strings.TrimSpace(from), "Archived") && !strings.EqualFold(strings.TrimSpace(to), "Archived") {
		return nil
	}
	if strings.EqualFold(strings.TrimSpace(to), "Archived") {
		return nil
	}
	return immutableStateError(entity, from)
}

func isArchiveOnlyPatch(patch map[string]any) bool {
	if len(patch) != 1 {
		return false
	}
	status := strings.TrimSpace(toString(patch["status"]))
	return strings.EqualFold(status, "Archived")
}

func isRestoreOnlyPatch(from string, patch map[string]any) bool {
	if !strings.EqualFold(strings.TrimSpace(from), "Archived") {
		return false
	}
	if len(patch) != 1 {
		return false
	}
	status := strings.TrimSpace(toString(patch["status"]))
	return status != "" && !strings.EqualFold(status, "Archived")
}

func isImmutableStatus(status string, immutableStatuses map[string]struct{}) bool {
	trimmed := strings.TrimSpace(status)
	if trimmed == "" {
		return false
	}
	_, ok := immutableStatuses[trimmed]
	return ok
}

func immutableStateError(entity, status string) error {
	message := fmt.Sprintf("%s in status %q is immutable; only archive or restore status operation is allowed", strings.TrimSpace(entity), strings.TrimSpace(status))
	return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, message)
}

func isAllowedTransition(from, to string, matrix map[string]map[string]struct{}) bool {
	trimmedFrom := strings.TrimSpace(from)
	trimmedTo := strings.TrimSpace(to)
	if trimmedFrom == "" || trimmedTo == "" {
		return false
	}
	allowed, ok := matrix[trimmedFrom]
	if !ok {
		return false
	}
	_, ok = allowed[trimmedTo]
	return ok
}
