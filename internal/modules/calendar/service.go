package calendar

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/storage"
)

type Service interface {
	ListCalendarData(ctx context.Context, projectID string, query listQuery) (ListCalendarDataResponse, error)
	CreateCalendarEvent(ctx context.Context, projectID, actorUserID string, req createCalendarEventRequest) (CalendarListEvent, error)
	GetCalendarEvent(ctx context.Context, projectID, eventID string) (GetCalendarEventResponse, error)
	UpdateCalendarEvent(ctx context.Context, projectID, eventID, actorUserID string, req updateCalendarEventRequest) (UpdateCalendarEventResponse, error)
	DeleteCalendarEvent(ctx context.Context, projectID, eventID, actorUserID string) (DeleteCalendarEventResponse, error)
}

type service struct {
	store storage.RelationalStore
	repo  Repo
}

func NewService(store storage.RelationalStore, repo Repo) Service {
	return &service{store: store, repo: repo}
}

func (s *service) ListCalendarData(ctx context.Context, projectID string, query listQuery) (ListCalendarDataResponse, error) {
	data, err := s.repo.ListCalendarData(ctx, projectID, query)
	if err != nil {
		return ListCalendarDataResponse{}, mapServiceError("list calendar", err)
	}
	return data, nil
}

func (s *service) CreateCalendarEvent(ctx context.Context, projectID, actorUserID string, req createCalendarEventRequest) (CalendarListEvent, error) {
	if err := req.Validate(); err != nil {
		return CalendarListEvent{}, err
	}
	if strings.TrimSpace(actorUserID) == "" {
		return CalendarListEvent{}, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	}
	var created CalendarListEvent
	err := s.store.WithTx(ctx, func(txCtx context.Context) error {
		result, createErr := s.repo.CreateCalendarEvent(txCtx, projectID, actorUserID, req)
		if createErr != nil {
			return createErr
		}
		created = result
		return nil
	})
	if err != nil {
		return CalendarListEvent{}, mapServiceError("create calendar event", err)
	}
	return created, nil
}

func (s *service) GetCalendarEvent(ctx context.Context, projectID, eventID string) (GetCalendarEventResponse, error) {
	item, err := s.repo.GetCalendarEvent(ctx, projectID, eventID)
	if err != nil {
		return GetCalendarEventResponse{}, mapServiceError("get calendar event", err)
	}
	return item, nil
}

func (s *service) UpdateCalendarEvent(ctx context.Context, projectID, eventID, actorUserID string, req updateCalendarEventRequest) (UpdateCalendarEventResponse, error) {
	if err := req.Validate(); err != nil {
		return UpdateCalendarEventResponse{}, err
	}
	var updated UpdateCalendarEventResponse
	err := s.store.WithTx(ctx, func(txCtx context.Context) error {
		result, updateErr := s.repo.UpdateCalendarEvent(txCtx, projectID, eventID, actorUserID, req.State)
		if updateErr != nil {
			return updateErr
		}
		updated = result
		return nil
	})
	if err != nil {
		return UpdateCalendarEventResponse{}, mapServiceError("update calendar event", err)
	}
	return updated, nil
}

func (s *service) DeleteCalendarEvent(ctx context.Context, projectID, eventID, actorUserID string) (DeleteCalendarEventResponse, error) {
	var deleted DeleteCalendarEventResponse
	err := s.store.WithTx(ctx, func(txCtx context.Context) error {
		result, deleteErr := s.repo.DeleteCalendarEvent(txCtx, projectID, eventID, actorUserID)
		if deleteErr != nil {
			return deleteErr
		}
		deleted = result
		return nil
	})
	if err != nil {
		return DeleteCalendarEventResponse{}, mapServiceError("delete calendar event", err)
	}
	return deleted, nil
}

func mapServiceError(action string, err error) error {
	if err == nil {
		return nil
	}
	if ae, ok := apperr.AsAppError(err); ok {
		return ae
	}
	return apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "failed to process calendar request"), fmt.Errorf("%s: %w", action, err))
}
