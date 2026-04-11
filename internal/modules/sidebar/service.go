package sidebar

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/storage"
)

type Service interface {
	CreateSidebarArtifact(ctx context.Context, projectID, actorUserID string, req createSidebarArtifactRequest) (map[string]any, error)
	RenameSidebarArtifact(ctx context.Context, projectID, artifactID, actorUserID string, req renameSidebarArtifactRequest) (map[string]any, error)
	DeleteSidebarArtifact(ctx context.Context, projectID, artifactID, actorUserID string, req deleteSidebarArtifactRequest) (map[string]any, error)
}

type service struct {
	store storage.RelationalStore
	repo  Repo
}

func NewService(store storage.RelationalStore, repo Repo) Service {
	return &service{store: store, repo: repo}
}

func (s *service) CreateSidebarArtifact(ctx context.Context, projectID, actorUserID string, req createSidebarArtifactRequest) (map[string]any, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(actorUserID) == "" {
		return nil, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	}
	var out map[string]any
	err := s.store.WithTx(ctx, func(txCtx context.Context) error {
		result, createErr := s.repo.CreateSidebarArtifact(txCtx, projectID, actorUserID, normalizePrefix(req.Prefix), req.Title)
		if createErr != nil {
			return createErr
		}
		out = result
		return nil
	})
	if err != nil {
		return nil, mapServiceError("create sidebar artifact", err)
	}
	return out, nil
}

func (s *service) RenameSidebarArtifact(ctx context.Context, projectID, artifactID, actorUserID string, req renameSidebarArtifactRequest) (map[string]any, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var out map[string]any
	err := s.store.WithTx(ctx, func(txCtx context.Context) error {
		result, renameErr := s.repo.RenameSidebarArtifact(txCtx, projectID, artifactID, actorUserID, normalizePrefix(req.Prefix), req.Title)
		if renameErr != nil {
			return renameErr
		}
		out = result
		return nil
	})
	if err != nil {
		return nil, mapServiceError("rename sidebar artifact", err)
	}
	return out, nil
}

func (s *service) DeleteSidebarArtifact(ctx context.Context, projectID, artifactID, actorUserID string, req deleteSidebarArtifactRequest) (map[string]any, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var out map[string]any
	err := s.store.WithTx(ctx, func(txCtx context.Context) error {
		result, deleteErr := s.repo.DeleteSidebarArtifact(txCtx, projectID, artifactID, actorUserID, normalizePrefix(req.Prefix))
		if deleteErr != nil {
			return deleteErr
		}
		out = result
		return nil
	})
	if err != nil {
		return nil, mapServiceError("delete sidebar artifact", err)
	}
	return out, nil
}

func mapServiceError(action string, err error) error {
	if err == nil {
		return nil
	}
	if ae, ok := apperr.AsAppError(err); ok {
		return ae
	}
	return apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "failed to process sidebar request"), fmt.Errorf("%s: %w", action, err))
}
