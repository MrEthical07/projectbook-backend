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
	var updated map[string]any
	err := s.store.WithTx(ctx, func(txCtx context.Context) error {
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
	var updated map[string]any
	err := s.store.WithTx(ctx, func(txCtx context.Context) error {
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
