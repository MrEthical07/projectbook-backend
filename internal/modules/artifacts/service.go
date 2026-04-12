package artifacts

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/storage"
)

type Service interface {
	ListStories(ctx context.Context, projectID string, query listQuery) (map[string]any, error)
	CreateStory(ctx context.Context, projectID, actorUserID string, req createStoryRequest) (map[string]any, error)
	GetStory(ctx context.Context, projectID, storyID string) (map[string]any, error)
	UpdateStory(ctx context.Context, projectID, storyID, actorUserID string, req updateStoryRequest) (map[string]any, error)

	ListJourneys(ctx context.Context, projectID string, query listQuery) (map[string]any, error)
	CreateJourney(ctx context.Context, projectID, actorUserID string, req createJourneyRequest) (map[string]any, error)
	GetJourney(ctx context.Context, projectID, journeyID string) (map[string]any, error)
	UpdateJourney(ctx context.Context, projectID, journeyID, actorUserID string, req updateJourneyRequest) (map[string]any, error)

	ListProblems(ctx context.Context, projectID string, query listQuery) (map[string]any, error)
	CreateProblem(ctx context.Context, projectID, actorUserID string, req createProblemRequest) (map[string]any, error)
	GetProblem(ctx context.Context, projectID, problemID string) (map[string]any, error)
	UpdateProblem(ctx context.Context, projectID, problemID, actorUserID string, req updateProblemRequest) (map[string]any, error)
	UpdateProblemStatus(ctx context.Context, projectID, problemID, actorUserID string, req updateProblemStatusRequest) (map[string]any, error)

	ListIdeas(ctx context.Context, projectID string, query listQuery) (map[string]any, error)
	CreateIdea(ctx context.Context, projectID, actorUserID string, req createIdeaRequest) (map[string]any, error)
	GetIdea(ctx context.Context, projectID, ideaID string) (map[string]any, error)
	UpdateIdea(ctx context.Context, projectID, ideaID, actorUserID string, req updateIdeaRequest) (map[string]any, error)
	UpdateIdeaStatus(ctx context.Context, projectID, ideaID, actorUserID string, req updateIdeaStatusRequest) (map[string]any, error)

	ListTasks(ctx context.Context, projectID string, query listQuery) (map[string]any, error)
	CreateTask(ctx context.Context, projectID, actorUserID string, req createTaskRequest) (map[string]any, error)
	GetTask(ctx context.Context, projectID, taskID string) (map[string]any, error)
	UpdateTask(ctx context.Context, projectID, taskID, actorUserID string, req updateTaskRequest) (map[string]any, error)
	UpdateTaskStatus(ctx context.Context, projectID, taskID, actorUserID string, req updateTaskStatusRequest) (map[string]any, error)

	ListFeedback(ctx context.Context, projectID string, query listQuery) (map[string]any, error)
	CreateFeedback(ctx context.Context, projectID, actorUserID string, req createFeedbackRequest) (map[string]any, error)
	GetFeedback(ctx context.Context, projectID, feedbackID string) (map[string]any, error)
	UpdateFeedback(ctx context.Context, projectID, feedbackID, actorUserID string, req updateFeedbackRequest) (map[string]any, error)
}

type service struct {
	store storage.RelationalStore
	repo  Repo
}

var (
	storyImmutableStatuses = map[string]struct{}{
		"Locked":   {},
		"Archived": {},
	}
	journeyImmutableStatuses = map[string]struct{}{
		"Archived": {},
	}
	problemImmutableStatuses = map[string]struct{}{
		"Locked":   {},
		"Archived": {},
	}
	ideaImmutableStatuses = map[string]struct{}{
		"Selected": {},
		"Rejected": {},
		"Archived": {},
	}
	taskImmutableStatuses = map[string]struct{}{
		"Completed": {},
		"Abandoned": {},
	}
)

func NewService(store storage.RelationalStore, repo Repo) Service {
	return &service{store: store, repo: repo}
}

func (s *service) ListStories(ctx context.Context, projectID string, query listQuery) (map[string]any, error) {
	items, err := s.repo.ListStories(ctx, projectID, query)
	if err != nil {
		return nil, mapServiceError("list stories", err)
	}
	return listPayload(items, query), nil
}

func (s *service) CreateStory(ctx context.Context, projectID, actorUserID string, req createStoryRequest) (map[string]any, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(actorUserID) == "" {
		return nil, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	}
	var created map[string]any
	err := s.store.WithTx(ctx, func(txCtx context.Context) error {
		result, createErr := s.repo.CreateStory(txCtx, projectID, actorUserID, req.Title, nil)
		if createErr != nil {
			return createErr
		}
		created = result
		return nil
	})
	if err != nil {
		return nil, mapServiceError("create story", err)
	}
	return created, nil
}

func (s *service) GetStory(ctx context.Context, projectID, storyID string) (map[string]any, error) {
	item, err := s.repo.GetStory(ctx, projectID, storyID)
	if err != nil {
		return nil, mapServiceError("get story", err)
	}
	return item, nil
}

func (s *service) UpdateStory(ctx context.Context, projectID, storyID, actorUserID string, req updateStoryRequest) (map[string]any, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	current, err := s.repo.GetStory(ctx, projectID, storyID)
	if err != nil {
		return nil, mapServiceError("load story before update", err)
	}
	from := nestedString(current, "story", "status")
	if err := enforceArchiveOnlyForImmutableUpdate("story", from, req.Story, storyImmutableStatuses); err != nil {
		return nil, err
	}
	if status := toString(req.Story["status"]); status != "" {
		if !isAllowedTransition(from, status, map[string]map[string]struct{}{
			"Draft":    {"Draft": {}, "Locked": {}, "Archived": {}},
			"Locked":   {"Locked": {}, "Archived": {}},
			"Archived": {"Archived": {}},
		}) {
			return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid story status transition")
		}
	}

	var updated map[string]any
	err = s.store.WithTx(ctx, func(txCtx context.Context) error {
		result, updateErr := s.repo.UpdateStory(txCtx, projectID, storyID, actorUserID, req.Story)
		if updateErr != nil {
			return updateErr
		}
		updated = result
		return nil
	})
	if err != nil {
		return nil, mapServiceError("update story", err)
	}
	return updated, nil
}

func (s *service) ListJourneys(ctx context.Context, projectID string, query listQuery) (map[string]any, error) {
	items, err := s.repo.ListJourneys(ctx, projectID, query)
	if err != nil {
		return nil, mapServiceError("list journeys", err)
	}
	return listPayload(items, query), nil
}

func (s *service) CreateJourney(ctx context.Context, projectID, actorUserID string, req createJourneyRequest) (map[string]any, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(actorUserID) == "" {
		return nil, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	}
	var created map[string]any
	err := s.store.WithTx(ctx, func(txCtx context.Context) error {
		result, createErr := s.repo.CreateJourney(txCtx, projectID, actorUserID, req.Title, nil)
		if createErr != nil {
			return createErr
		}
		created = result
		return nil
	})
	if err != nil {
		return nil, mapServiceError("create journey", err)
	}
	return created, nil
}

func (s *service) GetJourney(ctx context.Context, projectID, journeyID string) (map[string]any, error) {
	item, err := s.repo.GetJourney(ctx, projectID, journeyID)
	if err != nil {
		return nil, mapServiceError("get journey", err)
	}
	return item, nil
}

func (s *service) UpdateJourney(ctx context.Context, projectID, journeyID, actorUserID string, req updateJourneyRequest) (map[string]any, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	current, err := s.repo.GetJourney(ctx, projectID, journeyID)
	if err != nil {
		return nil, mapServiceError("load journey before update", err)
	}
	from := nestedString(current, "journey", "status")
	if err := enforceArchiveOnlyForImmutableUpdate("journey", from, req.Journey, journeyImmutableStatuses); err != nil {
		return nil, err
	}
	if status := toString(req.Journey["status"]); status != "" {
		if !isAllowedTransition(from, status, map[string]map[string]struct{}{
			"Draft":    {"Draft": {}, "Archived": {}},
			"Archived": {"Archived": {}},
		}) {
			return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid journey status transition")
		}
	}
	var updated map[string]any
	err = s.store.WithTx(ctx, func(txCtx context.Context) error {
		result, updateErr := s.repo.UpdateJourney(txCtx, projectID, journeyID, actorUserID, req.Journey)
		if updateErr != nil {
			return updateErr
		}
		updated = result
		return nil
	})
	if err != nil {
		return nil, mapServiceError("update journey", err)
	}
	return updated, nil
}

func (s *service) ListProblems(ctx context.Context, projectID string, query listQuery) (map[string]any, error) {
	items, err := s.repo.ListProblems(ctx, projectID, query)
	if err != nil {
		return nil, mapServiceError("list problems", err)
	}
	return listPayload(items, query), nil
}

func (s *service) CreateProblem(ctx context.Context, projectID, actorUserID string, req createProblemRequest) (map[string]any, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var created map[string]any
	err := s.store.WithTx(ctx, func(txCtx context.Context) error {
		result, createErr := s.repo.CreateProblem(txCtx, projectID, actorUserID, req.Statement, nil)
		if createErr != nil {
			return createErr
		}
		created = result
		return nil
	})
	if err != nil {
		return nil, mapServiceError("create problem", err)
	}
	return created, nil
}

func (s *service) GetProblem(ctx context.Context, projectID, problemID string) (map[string]any, error) {
	item, err := s.repo.GetProblem(ctx, projectID, problemID)
	if err != nil {
		return nil, mapServiceError("get problem", err)
	}
	return item, nil
}

func (s *service) UpdateProblem(ctx context.Context, projectID, problemID, actorUserID string, req updateProblemRequest) (map[string]any, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	current, err := s.repo.GetProblem(ctx, projectID, problemID)
	if err != nil {
		return nil, mapServiceError("load problem before update", err)
	}
	from := nestedString(current, "problem", "status")
	if err := enforceArchiveOnlyForImmutableUpdate("problem", from, req.State, problemImmutableStatuses); err != nil {
		return nil, err
	}
	status := toString(req.State["status"])
	if status != "" {
		if !isAllowedTransition(from, status, map[string]map[string]struct{}{
			"Draft":    {"Draft": {}, "Locked": {}, "Archived": {}},
			"Locked":   {"Locked": {}, "Archived": {}},
			"Archived": {"Archived": {}},
		}) {
			return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid problem status transition")
		}
	}

	var updated map[string]any
	err = s.store.WithTx(ctx, func(txCtx context.Context) error {
		result, updateErr := s.repo.UpdateProblem(txCtx, projectID, problemID, actorUserID, req.State)
		if updateErr != nil {
			return updateErr
		}
		updated = result
		return nil
	})
	if err != nil {
		return nil, mapServiceError("update problem", err)
	}
	return updated, nil
}

func (s *service) UpdateProblemStatus(ctx context.Context, projectID, problemID, actorUserID string, req updateProblemStatusRequest) (map[string]any, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	current, err := s.repo.GetProblem(ctx, projectID, problemID)
	if err != nil {
		return nil, mapServiceError("load problem before status update", err)
	}
	from := nestedString(current, "problem", "status")
	if err := enforceArchiveOnlyForImmutableStatusChange("problem", from, req.Status, problemImmutableStatuses); err != nil {
		return nil, err
	}
	if !isAllowedTransition(from, req.Status, map[string]map[string]struct{}{
		"Draft":    {"Draft": {}, "Locked": {}, "Archived": {}},
		"Locked":   {"Locked": {}, "Archived": {}},
		"Archived": {"Archived": {}},
	}) {
		return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid problem status transition")
	}

	var updated map[string]any
	err = s.store.WithTx(ctx, func(txCtx context.Context) error {
		if req.Status == "Locked" {
			result, lockErr := s.repo.LockProblem(txCtx, projectID, problemID, actorUserID)
			if lockErr != nil {
				return lockErr
			}
			updated = result
			return nil
		}
		result, updateErr := s.repo.UpdateProblemStatus(txCtx, projectID, problemID, req.Status, actorUserID)
		if updateErr != nil {
			return updateErr
		}
		updated = result
		return nil
	})
	if err != nil {
		return nil, mapServiceError("update problem status", err)
	}
	return updated, nil
}

func (s *service) ListIdeas(ctx context.Context, projectID string, query listQuery) (map[string]any, error) {
	items, err := s.repo.ListIdeas(ctx, projectID, query)
	if err != nil {
		return nil, mapServiceError("list ideas", err)
	}
	return listPayload(items, query), nil
}

func (s *service) CreateIdea(ctx context.Context, projectID, actorUserID string, req createIdeaRequest) (map[string]any, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var created map[string]any
	err := s.store.WithTx(ctx, func(txCtx context.Context) error {
		result, createErr := s.repo.CreateIdea(txCtx, projectID, actorUserID, req.Title, nil)
		if createErr != nil {
			return createErr
		}
		created = result
		return nil
	})
	if err != nil {
		return nil, mapServiceError("create idea", err)
	}
	return created, nil
}

func (s *service) GetIdea(ctx context.Context, projectID, ideaID string) (map[string]any, error) {
	item, err := s.repo.GetIdea(ctx, projectID, ideaID)
	if err != nil {
		return nil, mapServiceError("get idea", err)
	}
	return item, nil
}

func (s *service) UpdateIdea(ctx context.Context, projectID, ideaID, actorUserID string, req updateIdeaRequest) (map[string]any, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	current, err := s.repo.GetIdea(ctx, projectID, ideaID)
	if err != nil {
		return nil, mapServiceError("load idea before update", err)
	}
	from := nestedString(current, "idea", "status")
	if err := enforceArchiveOnlyForImmutableUpdate("idea", from, req.State, ideaImmutableStatuses); err != nil {
		return nil, err
	}
	var updated map[string]any
	err = s.store.WithTx(ctx, func(txCtx context.Context) error {
		result, updateErr := s.repo.UpdateIdea(txCtx, projectID, ideaID, actorUserID, req.State)
		if updateErr != nil {
			return updateErr
		}
		updated = result
		return nil
	})
	if err != nil {
		return nil, mapServiceError("update idea", err)
	}
	return updated, nil
}

func (s *service) UpdateIdeaStatus(ctx context.Context, projectID, ideaID, actorUserID string, req updateIdeaStatusRequest) (map[string]any, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	current, err := s.repo.GetIdea(ctx, projectID, ideaID)
	if err != nil {
		return nil, mapServiceError("load idea before status update", err)
	}
	from := nestedString(current, "idea", "status")
	if err := enforceArchiveOnlyForImmutableStatusChange("idea", from, req.Status, ideaImmutableStatuses); err != nil {
		return nil, err
	}
	if !isAllowedTransition(from, req.Status, map[string]map[string]struct{}{
		"Considered": {"Considered": {}, "Selected": {}, "Rejected": {}, "Archived": {}},
		"Selected":   {"Selected": {}, "Rejected": {}, "Archived": {}},
		"Rejected":   {"Rejected": {}, "Archived": {}},
		"Archived":   {"Archived": {}},
	}) {
		return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid idea status transition")
	}

	var updated map[string]any
	err = s.store.WithTx(ctx, func(txCtx context.Context) error {
		if req.Status == "Selected" {
			result, selectErr := s.repo.SelectIdea(txCtx, projectID, ideaID, actorUserID)
			if selectErr != nil {
				return selectErr
			}
			updated = result
			return nil
		}
		result, updateErr := s.repo.UpdateIdeaStatus(txCtx, projectID, ideaID, req.Status, actorUserID)
		if updateErr != nil {
			return updateErr
		}
		updated = result
		return nil
	})
	if err != nil {
		return nil, mapServiceError("update idea status", err)
	}
	return updated, nil
}

func (s *service) ListTasks(ctx context.Context, projectID string, query listQuery) (map[string]any, error) {
	items, err := s.repo.ListTasks(ctx, projectID, query)
	if err != nil {
		return nil, mapServiceError("list tasks", err)
	}
	return listPayload(items, query), nil
}

func (s *service) CreateTask(ctx context.Context, projectID, actorUserID string, req createTaskRequest) (map[string]any, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var created map[string]any
	err := s.store.WithTx(ctx, func(txCtx context.Context) error {
		result, createErr := s.repo.CreateTask(txCtx, projectID, actorUserID, req.Title, nil)
		if createErr != nil {
			return createErr
		}
		created = result
		return nil
	})
	if err != nil {
		return nil, mapServiceError("create task", err)
	}
	return created, nil
}

func (s *service) GetTask(ctx context.Context, projectID, taskID string) (map[string]any, error) {
	item, err := s.repo.GetTask(ctx, projectID, taskID)
	if err != nil {
		return nil, mapServiceError("get task", err)
	}
	return item, nil
}

func (s *service) UpdateTask(ctx context.Context, projectID, taskID, actorUserID string, req updateTaskRequest) (map[string]any, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	current, err := s.repo.GetTask(ctx, projectID, taskID)
	if err != nil {
		return nil, mapServiceError("load task before update", err)
	}
	from := nestedString(current, "task", "status")
	if err := enforceArchiveOnlyForImmutableUpdate("task", from, req.State, taskImmutableStatuses); err != nil {
		return nil, err
	}
	var updated map[string]any
	err = s.store.WithTx(ctx, func(txCtx context.Context) error {
		result, updateErr := s.repo.UpdateTask(txCtx, projectID, taskID, actorUserID, req.State)
		if updateErr != nil {
			return updateErr
		}
		updated = result
		return nil
	})
	if err != nil {
		return nil, mapServiceError("update task", err)
	}
	return updated, nil
}

func (s *service) UpdateTaskStatus(ctx context.Context, projectID, taskID, actorUserID string, req updateTaskStatusRequest) (map[string]any, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	current, err := s.repo.GetTask(ctx, projectID, taskID)
	if err != nil {
		return nil, mapServiceError("load task before status update", err)
	}
	from := nestedString(current, "task", "status")
	if err := enforceArchiveOnlyForImmutableStatusChange("task", from, req.Status, taskImmutableStatuses); err != nil {
		return nil, err
	}
	if !isAllowedTransition(from, req.Status, map[string]map[string]struct{}{
		"Planned":     {"Planned": {}, "In Progress": {}, "Abandoned": {}},
		"In Progress": {"In Progress": {}, "Completed": {}, "Abandoned": {}},
		"Completed":   {"Completed": {}},
		"Abandoned":   {"Abandoned": {}},
	}) {
		return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid task status transition")
	}
	var updated map[string]any
	err = s.store.WithTx(ctx, func(txCtx context.Context) error {
		result, updateErr := s.repo.UpdateTaskStatus(txCtx, projectID, taskID, req.Status, actorUserID)
		if updateErr != nil {
			return updateErr
		}
		updated = result
		return nil
	})
	if err != nil {
		return nil, mapServiceError("update task status", err)
	}
	return updated, nil
}

func (s *service) ListFeedback(ctx context.Context, projectID string, query listQuery) (map[string]any, error) {
	items, err := s.repo.ListFeedback(ctx, projectID, query)
	if err != nil {
		return nil, mapServiceError("list feedback", err)
	}
	return listPayload(items, query), nil
}

func (s *service) CreateFeedback(ctx context.Context, projectID, actorUserID string, req createFeedbackRequest) (map[string]any, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var created map[string]any
	err := s.store.WithTx(ctx, func(txCtx context.Context) error {
		result, createErr := s.repo.CreateFeedback(txCtx, projectID, actorUserID, req.Title, nil)
		if createErr != nil {
			return createErr
		}
		created = result
		return nil
	})
	if err != nil {
		return nil, mapServiceError("create feedback", err)
	}
	return created, nil
}

func (s *service) GetFeedback(ctx context.Context, projectID, feedbackID string) (map[string]any, error) {
	item, err := s.repo.GetFeedback(ctx, projectID, feedbackID)
	if err != nil {
		return nil, mapServiceError("get feedback", err)
	}
	return item, nil
}

func (s *service) UpdateFeedback(ctx context.Context, projectID, feedbackID, actorUserID string, req updateFeedbackRequest) (map[string]any, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var updated map[string]any
	err := s.store.WithTx(ctx, func(txCtx context.Context) error {
		result, updateErr := s.repo.UpdateFeedback(txCtx, projectID, feedbackID, actorUserID, req.State)
		if updateErr != nil {
			return updateErr
		}
		updated = result
		return nil
	})
	if err != nil {
		return nil, mapServiceError("update feedback", err)
	}
	return updated, nil
}

func listPayload(items []map[string]any, query listQuery) map[string]any {
	return map[string]any{
		"items":  items,
		"offset": query.Offset,
		"limit":  query.Limit,
	}
}

func nestedString(payload map[string]any, key, field string) string {
	obj := toMap(payload[key])
	if obj == nil {
		return ""
	}
	return toString(obj[field])
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

func mapServiceError(action string, err error) error {
	if err == nil {
		return nil
	}
	if ae, ok := apperr.AsAppError(err); ok {
		return ae
	}
	return apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "failed to process artifacts request"), fmt.Errorf("%s: %w", action, err))
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
