package activity

import (
	"context"
	"fmt"
	"net/http"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/pagination"
)

type Service interface {
	ListProjectActivity(ctx context.Context, projectID string, query listQuery) (ListProjectActivityResponse, error)
}

type service struct {
	repo Repo
}

func NewService(repo Repo) Service {
	return &service{repo: repo}
}

func (s *service) ListProjectActivity(ctx context.Context, projectID string, query listQuery) (ListProjectActivityResponse, error) {
	items, err := s.repo.ListProjectActivity(ctx, projectID, query)
	if err != nil {
		return ListProjectActivityResponse{}, mapServiceError("list project activity", err)
	}

	hasMore := len(items) > query.Limit
	if hasMore {
		items = items[:query.Limit]
	}

	var nextCursor *string
	if hasMore {
		cursor := pagination.EncodeOffsetCursor(query.Offset + query.Limit)
		nextCursor = &cursor
	}

	return ListProjectActivityResponse{Items: items, NextCursor: nextCursor}, nil
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
