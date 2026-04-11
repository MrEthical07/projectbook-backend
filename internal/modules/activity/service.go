package activity

import (
	"context"
	"fmt"
	"net/http"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
)

type Service interface {
	ListProjectActivity(ctx context.Context, projectID string, query listQuery) ([]map[string]any, error)
}

type service struct {
	repo Repo
}

func NewService(repo Repo) Service {
	return &service{repo: repo}
}

func (s *service) ListProjectActivity(ctx context.Context, projectID string, query listQuery) ([]map[string]any, error) {
	items, err := s.repo.ListProjectActivity(ctx, projectID, query)
	if err != nil {
		return nil, mapServiceError("list project activity", err)
	}
	return items, nil
}

func mapServiceError(action string, err error) error {
	if err == nil {
		return nil
	}
	if ae, ok := apperr.AsAppError(err); ok {
		return ae
	}
	return apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "failed to process activity request"), fmt.Errorf("%s: %w", action, err))
}
