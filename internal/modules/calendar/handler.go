package calendar

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

func (h *Handler) ListCalendarData(ctx *httpx.Context, _ httpx.NoBody) (ListCalendarDataResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return ListCalendarDataResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return ListCalendarDataResponse{}, err
	}
	query, err := parseListQuery(ctx)
	if err != nil {
		return ListCalendarDataResponse{}, err
	}
	return h.svc.ListCalendarData(ctx.Context(), resolveProjectScope(principal, projectID), query)
}

func (h *Handler) CreateCalendarEvent(ctx *httpx.Context, req createCalendarEventRequest) (httpx.Result[CalendarListEvent], error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return httpx.Result[CalendarListEvent]{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return httpx.Result[CalendarListEvent]{}, err
	}
	created, err := h.svc.CreateCalendarEvent(ctx.Context(), resolveProjectScope(principal, projectID), principal.UserID, req)
	if err != nil {
		return httpx.Result[CalendarListEvent]{}, err
	}
	return httpx.Result[CalendarListEvent]{Status: http.StatusCreated, Data: created}, nil
}

func (h *Handler) GetCalendarEvent(ctx *httpx.Context, _ httpx.NoBody) (GetCalendarEventResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return GetCalendarEventResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return GetCalendarEventResponse{}, err
	}
	eventID := strings.TrimSpace(ctx.Param("eventId"))
	if eventID == "" {
		return GetCalendarEventResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "eventId is required")
	}
	return h.svc.GetCalendarEvent(ctx.Context(), resolveProjectScope(principal, projectID), eventID)
}

func (h *Handler) UpdateCalendarEvent(ctx *httpx.Context, req updateCalendarEventRequest) (UpdateCalendarEventResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return UpdateCalendarEventResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return UpdateCalendarEventResponse{}, err
	}
	eventID := strings.TrimSpace(ctx.Param("eventId"))
	if eventID == "" {
		return UpdateCalendarEventResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "eventId is required")
	}
	return h.svc.UpdateCalendarEvent(ctx.Context(), resolveProjectScope(principal, projectID), eventID, principal.UserID, req)
}

func (h *Handler) DeleteCalendarEvent(ctx *httpx.Context, _ httpx.NoBody) (DeleteCalendarEventResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return DeleteCalendarEventResponse{}, err
	}
	projectID, err := requireProjectID(ctx)
	if err != nil {
		return DeleteCalendarEventResponse{}, err
	}
	eventID := strings.TrimSpace(ctx.Param("eventId"))
	if eventID == "" {
		return DeleteCalendarEventResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "eventId is required")
	}
	return h.svc.DeleteCalendarEvent(ctx.Context(), resolveProjectScope(principal, projectID), eventID, principal.UserID)
}

func parseListQuery(ctx *httpx.Context) (listQuery, error) {
	offset := 0
	if cursor := queryValue(ctx, "pagination.cursor", "cursor"); cursor != "" {
		decodedOffset, decodeErr := pagination.DecodeOffsetCursor(cursor)
		if decodeErr != nil {
			return listQuery{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "cursor is invalid")
		}
		offset = decodedOffset
	}

	limit, err := parseLimit(queryValue(ctx, "pagination.limit", "limit"))
	if err != nil {
		return listQuery{}, err
	}
	return listQuery{Offset: offset, Limit: limit}, nil
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
