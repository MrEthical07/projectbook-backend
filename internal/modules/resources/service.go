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
	ListResources(ctx context.Context, projectID string, query listQuery) (map[string]any, error)
	CreateResource(ctx context.Context, projectID, actorUserID string, req createResourceRequest) (map[string]any, error)
	GetResource(ctx context.Context, projectID, resourceID string) (map[string]any, error)
	UpdateResource(ctx context.Context, projectID, resourceID, actorUserID string, req updateResourceRequest) (map[string]any, error)
	UpdateResourceStatus(ctx context.Context, projectID, resourceID, actorUserID string, req updateResourceStatusRequest) (map[string]any, error)
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

func (s *service) ListResources(ctx context.Context, projectID string, query listQuery) (map[string]any, error) {
	rows, err := s.repo.ListResources(ctx, projectID, query)
	if err != nil {
		return nil, mapServiceError("list resources", err)
	}
	return rows, nil
}

func (s *service) CreateResource(ctx context.Context, projectID, actorUserID string, req createResourceRequest) (map[string]any, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(actorUserID) == "" {
		return nil, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
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
		return nil, mapServiceError("create resource", err)
	}
	return created, nil
}

func (s *service) GetResource(ctx context.Context, projectID, resourceID string) (map[string]any, error) {
	item, err := s.repo.GetResource(ctx, projectID, resourceID)
	if err != nil {
		return nil, mapServiceError("get resource", err)
	}
	return item, nil
}

func (s *service) UpdateResource(ctx context.Context, projectID, resourceID, actorUserID string, req updateResourceRequest) (map[string]any, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	current, err := s.repo.GetResource(ctx, projectID, resourceID)
	if err != nil {
		return nil, mapServiceError("load resource before update", err)
	}
	from := nestedString(current, "resource", "status")
	if err := enforceArchiveOnlyForImmutableUpdate("resource", from, req.State, resourceImmutableStatuses); err != nil {
		return nil, err
	}
	if status := toString(req.State["status"]); status != "" {
		if !isAllowedTransition(from, status, map[string]map[string]struct{}{
			"Active":   {"Active": {}, "Archived": {}},
			"Archived": {"Archived": {}},
		}) {
			return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid resource status transition")
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
		return nil, mapServiceError("update resource", err)
	}
	return updated, nil
}

func (s *service) UpdateResourceStatus(ctx context.Context, projectID, resourceID, actorUserID string, req updateResourceStatusRequest) (map[string]any, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	current, err := s.repo.GetResource(ctx, projectID, resourceID)
	if err != nil {
		return nil, mapServiceError("load resource before status update", err)
	}
	from := nestedString(current, "resource", "status")
	if err := enforceArchiveOnlyForImmutableStatusChange("resource", from, req.Status, resourceImmutableStatuses); err != nil {
		return nil, err
	}
	if !isAllowedTransition(from, req.Status, map[string]map[string]struct{}{
		"Active":   {"Active": {}, "Archived": {}},
		"Archived": {"Archived": {}},
	}) {
		return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid resource status transition")
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
		return nil, mapServiceError("update resource status", err)
	}
	return updated, nil
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
	return immutableStateError(entity, from)
}

func enforceArchiveOnlyForImmutableStatusChange(entity, from, to string, immutableStatuses map[string]struct{}) error {
	if !isImmutableStatus(from, immutableStatuses) {
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

func isImmutableStatus(status string, immutableStatuses map[string]struct{}) bool {
	trimmed := strings.TrimSpace(status)
	if trimmed == "" {
		return false
	}
	_, ok := immutableStatuses[trimmed]
	return ok
}

func immutableStateError(entity, status string) error {
	message := fmt.Sprintf("%s in status %q is immutable; only archive operation is allowed", strings.TrimSpace(entity), strings.TrimSpace(status))
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
