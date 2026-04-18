package pages

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/pagination"
	"github.com/MrEthical07/superapi/internal/core/storage"
)

type Service interface {
	ListPages(ctx context.Context, projectID string, query listQuery) (ListPagesResponse, error)
	CreatePage(ctx context.Context, projectID, actorUserID string, req createPageRequest) (PageListItem, error)
	GetPage(ctx context.Context, projectID, pageID string) (GetPageResponse, error)
	UpdatePage(ctx context.Context, projectID, pageID, actorUserID string, req updatePageRequest) (PageListItem, error)
	RenamePage(ctx context.Context, projectID, pageID, actorUserID string, req renamePageRequest) (RenamePageResponse, error)

	CreatePageForSidebar(ctx context.Context, projectID, actorUserID, title string) (PageListItem, error)
	RenamePageForSidebar(ctx context.Context, projectID, pageID, actorUserID, title string) (RenamePageResponse, error)
	DeletePageForSidebar(ctx context.Context, projectID, pageID, actorUserID string) (DeletePageResponse, error)
}

type service struct {
	store storage.RelationalStore
	repo  Repo
}

var pageImmutableStatuses = map[string]struct{}{
	"Archived": {},
}

func NewService(store storage.RelationalStore, repo Repo) Service {
	return &service{store: store, repo: repo}
}

func (s *service) ListPages(ctx context.Context, projectID string, query listQuery) (ListPagesResponse, error) {
	rows, err := s.repo.ListPages(ctx, projectID, query)
	if err != nil {
		return ListPagesResponse{}, mapServiceError("list pages", err)
	}

	items := make([]PageListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, decodePageListItem(row))
	}

	var nextCursor *string
	if len(rows) > query.Limit {
		items = items[:query.Limit]
		cursor := pagination.EncodeOffsetCursor(query.Offset + len(items))
		nextCursor = &cursor
	}

	return ListPagesResponse{Items: items, NextCursor: nextCursor}, nil
}

func (s *service) CreatePage(ctx context.Context, projectID, actorUserID string, req createPageRequest) (PageListItem, error) {
	if err := req.Validate(); err != nil {
		return PageListItem{}, err
	}
	if strings.TrimSpace(actorUserID) == "" {
		return PageListItem{}, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	}
	var created map[string]any
	err := s.store.WithTx(ctx, func(txCtx context.Context) error {
		result, createErr := s.repo.CreatePage(txCtx, projectID, actorUserID, req.Title)
		if createErr != nil {
			return createErr
		}
		created = result
		return nil
	})
	if err != nil {
		return PageListItem{}, mapServiceError("create page", err)
	}
	return decodePageListItem(created), nil
}

func (s *service) GetPage(ctx context.Context, projectID, pageID string) (GetPageResponse, error) {
	item, err := s.repo.GetPage(ctx, projectID, pageID)
	if err != nil {
		return GetPageResponse{}, mapServiceError("get page", err)
	}
	return decodeGetPageResponse(item), nil
}

func (s *service) UpdatePage(ctx context.Context, projectID, pageID, actorUserID string, req updatePageRequest) (PageListItem, error) {
	if err := req.Validate(); err != nil {
		return PageListItem{}, err
	}
	current, err := s.repo.GetPage(ctx, projectID, pageID)
	if err != nil {
		return PageListItem{}, mapServiceError("load page before update", err)
	}
	from := nestedString(current, "page", "status")
	if err := enforceArchiveOnlyForImmutableUpdate("page", from, req.State, pageImmutableStatuses); err != nil {
		return PageListItem{}, err
	}
	if status := toString(req.State["status"]); status != "" {
		if !isAllowedTransition(from, status, map[string]map[string]struct{}{
			"Draft":    {"Draft": {}, "Archived": {}},
			"Archived": {"Archived": {}, "Draft": {}},
		}) {
			return PageListItem{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid page status transition")
		}
	}
	var updated map[string]any
	err = s.store.WithTx(ctx, func(txCtx context.Context) error {
		result, updateErr := s.repo.UpdatePage(txCtx, projectID, pageID, actorUserID, req.State)
		if updateErr != nil {
			return updateErr
		}
		updated = result
		return nil
	})
	if err != nil {
		return PageListItem{}, mapServiceError("update page", err)
	}
	return decodePageListItem(updated), nil
}

func (s *service) RenamePage(ctx context.Context, projectID, pageID, actorUserID string, req renamePageRequest) (RenamePageResponse, error) {
	if err := req.Validate(); err != nil {
		return RenamePageResponse{}, err
	}
	current, err := s.repo.GetPage(ctx, projectID, pageID)
	if err != nil {
		return RenamePageResponse{}, mapServiceError("load page before rename", err)
	}
	if err := enforceMutableOperation("page", nestedString(current, "page", "status"), pageImmutableStatuses); err != nil {
		return RenamePageResponse{}, err
	}
	var out map[string]any
	err = s.store.WithTx(ctx, func(txCtx context.Context) error {
		result, renameErr := s.repo.RenamePage(txCtx, projectID, pageID, req.Title, actorUserID)
		if renameErr != nil {
			return renameErr
		}
		out = result
		return nil
	})
	if err != nil {
		return RenamePageResponse{}, mapServiceError("rename page", err)
	}
	return decodeRenamePageResponse(out), nil
}

func (s *service) CreatePageForSidebar(ctx context.Context, projectID, actorUserID, title string) (PageListItem, error) {
	if strings.TrimSpace(title) == "" {
		return PageListItem{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "title is required")
	}
	var out map[string]any
	err := s.store.WithTx(ctx, func(txCtx context.Context) error {
		result, createErr := s.repo.CreatePageForSidebar(txCtx, projectID, actorUserID, title)
		if createErr != nil {
			return createErr
		}
		out = result
		return nil
	})
	if err != nil {
		return PageListItem{}, mapServiceError("create page sidebar", err)
	}
	return decodePageListItem(out), nil
}

func (s *service) RenamePageForSidebar(ctx context.Context, projectID, pageID, actorUserID, title string) (RenamePageResponse, error) {
	if strings.TrimSpace(title) == "" {
		return RenamePageResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "title is required")
	}
	current, err := s.repo.GetPage(ctx, projectID, pageID)
	if err != nil {
		return RenamePageResponse{}, mapServiceError("load page before sidebar rename", err)
	}
	if err := enforceMutableOperation("page", nestedString(current, "page", "status"), pageImmutableStatuses); err != nil {
		return RenamePageResponse{}, err
	}
	var out map[string]any
	err = s.store.WithTx(ctx, func(txCtx context.Context) error {
		result, renameErr := s.repo.RenamePageForSidebar(txCtx, projectID, pageID, title, actorUserID)
		if renameErr != nil {
			return renameErr
		}
		out = result
		return nil
	})
	if err != nil {
		return RenamePageResponse{}, mapServiceError("rename page sidebar", err)
	}
	return decodeRenamePageResponse(out), nil
}

func (s *service) DeletePageForSidebar(ctx context.Context, projectID, pageID, actorUserID string) (DeletePageResponse, error) {
	current, err := s.repo.GetPage(ctx, projectID, pageID)
	if err != nil {
		return DeletePageResponse{}, mapServiceError("load page before sidebar delete", err)
	}
	if err := enforceMutableOperation("page", nestedString(current, "page", "status"), pageImmutableStatuses); err != nil {
		return DeletePageResponse{}, err
	}
	var out map[string]any
	err = s.store.WithTx(ctx, func(txCtx context.Context) error {
		result, deleteErr := s.repo.DeletePageForSidebar(txCtx, projectID, pageID, actorUserID)
		if deleteErr != nil {
			return deleteErr
		}
		out = result
		return nil
	})
	if err != nil {
		return DeletePageResponse{}, mapServiceError("delete page sidebar", err)
	}
	return decodeDeletePageResponse(out), nil
}

func mapServiceError(action string, err error) error {
	if err == nil {
		return nil
	}
	if ae, ok := apperr.AsAppError(err); ok {
		return ae
	}
	return apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "failed to process pages request"), fmt.Errorf("%s: %w", action, err))
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

func enforceMutableOperation(entity, status string, immutableStatuses map[string]struct{}) error {
	if isImmutableStatus(status, immutableStatuses) {
		return immutableStateError(entity, status)
	}
	return nil
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
