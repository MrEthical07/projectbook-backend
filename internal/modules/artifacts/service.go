package artifacts

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
	ListStories(ctx context.Context, projectID string, query listQuery) (StoryListResponse, error)
	CreateStory(ctx context.Context, projectID, actorUserID string, req createStoryRequest) (StoryListItem, error)
	GetStory(ctx context.Context, projectID, storyID string) (StoryPageResponse, error)
	UpdateStory(ctx context.Context, projectID, storyID, actorUserID string, req updateStoryRequest) (StoryListItem, error)

	ListJourneys(ctx context.Context, projectID string, query listQuery) (JourneyListResponse, error)
	CreateJourney(ctx context.Context, projectID, actorUserID string, req createJourneyRequest) (JourneyListItem, error)
	GetJourney(ctx context.Context, projectID, journeyID string) (JourneyPageResponse, error)
	UpdateJourney(ctx context.Context, projectID, journeyID, actorUserID string, req updateJourneyRequest) (JourneyListItem, error)

	ListProblems(ctx context.Context, projectID string, query listQuery) (ProblemListResponse, error)
	CreateProblem(ctx context.Context, projectID, actorUserID string, req createProblemRequest) (ProblemListItem, error)
	GetProblem(ctx context.Context, projectID, problemID string) (ProblemPageResponse, error)
	UpdateProblem(ctx context.Context, projectID, problemID, actorUserID string, req updateProblemRequest) (ProblemListItem, error)
	UpdateProblemStatus(ctx context.Context, projectID, problemID, actorUserID string, req updateProblemStatusRequest) (ArtifactStatusResponse, error)

	ListIdeas(ctx context.Context, projectID string, query listQuery) (IdeaListResponse, error)
	CreateIdea(ctx context.Context, projectID, actorUserID string, req createIdeaRequest) (IdeaListItem, error)
	GetIdea(ctx context.Context, projectID, ideaID string) (IdeaPageResponse, error)
	UpdateIdea(ctx context.Context, projectID, ideaID, actorUserID string, req updateIdeaRequest) (IdeaListItem, error)
	UpdateIdeaStatus(ctx context.Context, projectID, ideaID, actorUserID string, req updateIdeaStatusRequest) (ArtifactStatusResponse, error)

	ListTasks(ctx context.Context, projectID string, query listQuery) (TaskListResponse, error)
	CreateTask(ctx context.Context, projectID, actorUserID string, req createTaskRequest) (TaskListItem, error)
	GetTask(ctx context.Context, projectID, taskID string) (TaskPageResponse, error)
	UpdateTask(ctx context.Context, projectID, taskID, actorUserID string, req updateTaskRequest) (TaskListItem, error)
	UpdateTaskStatus(ctx context.Context, projectID, taskID, actorUserID string, req updateTaskStatusRequest) (ArtifactStatusResponse, error)

	ListFeedback(ctx context.Context, projectID string, query listQuery) (FeedbackListResponse, error)
	CreateFeedback(ctx context.Context, projectID, actorUserID string, req createFeedbackRequest) (FeedbackListItem, error)
	GetFeedback(ctx context.Context, projectID, feedbackID string) (FeedbackPageResponse, error)
	UpdateFeedback(ctx context.Context, projectID, feedbackID, actorUserID string, req updateFeedbackRequest) (FeedbackListItem, error)
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
		"Locked":   {},
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

func (s *service) ListStories(ctx context.Context, projectID string, query listQuery) (StoryListResponse, error) {
	items, err := s.repo.ListStories(ctx, projectID, query)
	if err != nil {
		return StoryListResponse{}, mapServiceError("list stories", err)
	}
	out := make([]StoryListItem, 0, len(items))
	for _, item := range items {
		out = append(out, decodeStoryListItem(item))
	}
	nextCursor := paginateCursor(len(out), query)
	if nextCursor != nil {
		out = out[:query.Limit]
	}
	return StoryListResponse{Items: out, NextCursor: nextCursor}, nil
}

func (s *service) CreateStory(ctx context.Context, projectID, actorUserID string, req createStoryRequest) (StoryListItem, error) {
	if err := req.Validate(); err != nil {
		return StoryListItem{}, err
	}
	if strings.TrimSpace(actorUserID) == "" {
		return StoryListItem{}, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
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
		return StoryListItem{}, mapServiceError("create story", err)
	}
	return decodeStoryListItem(created), nil
}

func (s *service) GetStory(ctx context.Context, projectID, storyID string) (StoryPageResponse, error) {
	item, err := s.repo.GetStory(ctx, projectID, storyID)
	if err != nil {
		return StoryPageResponse{}, mapServiceError("get story", err)
	}
	storyRaw := toMap(item["story"])
	metadataRaw := toMap(item["metadata"])
	if storyRaw == nil {
		storyRaw = map[string]any{}
	}
	if metadataRaw == nil {
		metadataRaw = map[string]any{}
	}
	return StoryPageResponse{
		Story: StoryPage{
			ID:          toString(storyRaw["id"]),
			Title:       toString(storyRaw["title"]),
			Status:      toString(storyRaw["status"]),
			Owner:       toString(storyRaw["owner"]),
			LastUpdated: toString(storyRaw["lastUpdated"]),
		},
		Metadata:      decodeArtifactMetadata(metadataRaw),
		Detail:        toRawJSON(item["detail"], "{}"),
		AddOnCatalog:  toRawJSON(item["addOnCatalog"], "[]"),
		AddOnSections: toRawJSON(item["addOnSections"], "[]"),
		Reference:     toRawJSON(item["reference"], "{}"),
	}, nil
}

func (s *service) UpdateStory(ctx context.Context, projectID, storyID, actorUserID string, req updateStoryRequest) (StoryListItem, error) {
	if err := req.Validate(); err != nil {
		return StoryListItem{}, err
	}
	current, err := s.repo.GetStory(ctx, projectID, storyID)
	if err != nil {
		return StoryListItem{}, mapServiceError("load story before update", err)
	}
	from := nestedString(current, "story", "status")
	if err := enforceArchiveOnlyForImmutableUpdate("story", from, req.Story, storyImmutableStatuses); err != nil {
		return StoryListItem{}, err
	}
	if status := toString(req.Story["status"]); status != "" {
		if !isAllowedTransition(from, status, map[string]map[string]struct{}{
			"Draft":    {"Draft": {}, "Locked": {}, "Archived": {}},
			"Locked":   {"Locked": {}, "Archived": {}},
			"Archived": {"Archived": {}, "Draft": {}, "Locked": {}},
		}) {
			return StoryListItem{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid story status transition")
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
		return StoryListItem{}, mapServiceError("update story", err)
	}
	return decodeStoryListItem(updated), nil
}

func (s *service) ListJourneys(ctx context.Context, projectID string, query listQuery) (JourneyListResponse, error) {
	items, err := s.repo.ListJourneys(ctx, projectID, query)
	if err != nil {
		return JourneyListResponse{}, mapServiceError("list journeys", err)
	}
	out := make([]JourneyListItem, 0, len(items))
	for _, item := range items {
		out = append(out, decodeJourneyListItem(item))
	}
	nextCursor := paginateCursor(len(out), query)
	if nextCursor != nil {
		out = out[:query.Limit]
	}
	return JourneyListResponse{Items: out, NextCursor: nextCursor}, nil
}

func (s *service) CreateJourney(ctx context.Context, projectID, actorUserID string, req createJourneyRequest) (JourneyListItem, error) {
	if err := req.Validate(); err != nil {
		return JourneyListItem{}, err
	}
	if strings.TrimSpace(actorUserID) == "" {
		return JourneyListItem{}, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
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
		return JourneyListItem{}, mapServiceError("create journey", err)
	}
	return decodeJourneyListItem(created), nil
}

func (s *service) GetJourney(ctx context.Context, projectID, journeyID string) (JourneyPageResponse, error) {
	item, err := s.repo.GetJourney(ctx, projectID, journeyID)
	if err != nil {
		return JourneyPageResponse{}, mapServiceError("get journey", err)
	}
	journeyRaw := toMap(item["journey"])
	metadataRaw := toMap(item["metadata"])
	if journeyRaw == nil {
		journeyRaw = map[string]any{}
	}
	if metadataRaw == nil {
		metadataRaw = map[string]any{}
	}
	return JourneyPageResponse{
		Journey: JourneyPage{
			ID:          toString(journeyRaw["id"]),
			Title:       toString(journeyRaw["title"]),
			Status:      toString(journeyRaw["status"]),
			Owner:       toString(journeyRaw["owner"]),
			LastUpdated: toString(journeyRaw["lastUpdated"]),
		},
		Metadata:       decodeArtifactMetadata(metadataRaw),
		Detail:         toRawJSON(item["detail"], "{}"),
		EmotionOptions: toStringSlice(item["emotionOptions"]),
		Reference:      toRawJSON(item["reference"], "{}"),
	}, nil
}

func (s *service) UpdateJourney(ctx context.Context, projectID, journeyID, actorUserID string, req updateJourneyRequest) (JourneyListItem, error) {
	if err := req.Validate(); err != nil {
		return JourneyListItem{}, err
	}
	current, err := s.repo.GetJourney(ctx, projectID, journeyID)
	if err != nil {
		return JourneyListItem{}, mapServiceError("load journey before update", err)
	}
	from := nestedString(current, "journey", "status")
	if err := enforceArchiveOnlyForImmutableUpdate("journey", from, req.Journey, journeyImmutableStatuses); err != nil {
		return JourneyListItem{}, err
	}
	if status := toString(req.Journey["status"]); status != "" {
		if !isAllowedTransition(from, status, map[string]map[string]struct{}{
			"Draft":    {"Draft": {}, "Locked": {}, "Archived": {}},
			"Locked":   {"Locked": {}, "Archived": {}},
			"Archived": {"Archived": {}, "Draft": {}, "Locked": {}},
		}) {
			return JourneyListItem{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid journey status transition")
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
		return JourneyListItem{}, mapServiceError("update journey", err)
	}
	return decodeJourneyListItem(updated), nil
}

func (s *service) ListProblems(ctx context.Context, projectID string, query listQuery) (ProblemListResponse, error) {
	items, err := s.repo.ListProblems(ctx, projectID, query)
	if err != nil {
		return ProblemListResponse{}, mapServiceError("list problems", err)
	}
	out := make([]ProblemListItem, 0, len(items))
	for _, item := range items {
		out = append(out, decodeProblemListItem(item))
	}
	nextCursor := paginateCursor(len(out), query)
	if nextCursor != nil {
		out = out[:query.Limit]
	}
	return ProblemListResponse{Items: out, NextCursor: nextCursor}, nil
}

func (s *service) CreateProblem(ctx context.Context, projectID, actorUserID string, req createProblemRequest) (ProblemListItem, error) {
	if err := req.Validate(); err != nil {
		return ProblemListItem{}, err
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
		return ProblemListItem{}, mapServiceError("create problem", err)
	}
	return decodeProblemListItem(created), nil
}

func (s *service) GetProblem(ctx context.Context, projectID, problemID string) (ProblemPageResponse, error) {
	item, err := s.repo.GetProblem(ctx, projectID, problemID)
	if err != nil {
		return ProblemPageResponse{}, mapServiceError("get problem", err)
	}
	problemRaw := toMap(item["problem"])
	metadataRaw := toMap(item["metadata"])
	if problemRaw == nil {
		problemRaw = map[string]any{}
	}
	if metadataRaw == nil {
		metadataRaw = map[string]any{}
	}
	return ProblemPageResponse{
		Problem: ProblemPage{
			ID:          toString(problemRaw["id"]),
			Statement:   toString(problemRaw["statement"]),
			Title:       toString(problemRaw["title"]),
			Status:      toString(problemRaw["status"]),
			Owner:       toString(problemRaw["owner"]),
			LastUpdated: toString(problemRaw["lastUpdated"]),
		},
		Metadata:  decodeArtifactMetadata(metadataRaw),
		Detail:    toRawJSON(item["detail"], "{}"),
		Reference: toRawJSON(item["reference"], "{}"),
	}, nil
}

func (s *service) UpdateProblem(ctx context.Context, projectID, problemID, actorUserID string, req updateProblemRequest) (ProblemListItem, error) {
	if err := req.Validate(); err != nil {
		return ProblemListItem{}, err
	}
	current, err := s.repo.GetProblem(ctx, projectID, problemID)
	if err != nil {
		return ProblemListItem{}, mapServiceError("load problem before update", err)
	}
	from := nestedString(current, "problem", "status")
	if err := enforceArchiveOnlyForImmutableUpdate("problem", from, req.State, problemImmutableStatuses); err != nil {
		return ProblemListItem{}, err
	}
	status := toString(req.State["status"])
	if status != "" {
		if !isAllowedTransition(from, status, map[string]map[string]struct{}{
			"Draft":    {"Draft": {}, "Locked": {}, "Archived": {}},
			"Locked":   {"Locked": {}, "Archived": {}},
			"Archived": {"Archived": {}, "Draft": {}, "Locked": {}},
		}) {
			return ProblemListItem{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid problem status transition")
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
		return ProblemListItem{}, mapServiceError("update problem", err)
	}
	return decodeProblemListItem(updated), nil
}

func (s *service) UpdateProblemStatus(ctx context.Context, projectID, problemID, actorUserID string, req updateProblemStatusRequest) (ArtifactStatusResponse, error) {
	if err := req.Validate(); err != nil {
		return ArtifactStatusResponse{}, err
	}
	current, err := s.repo.GetProblem(ctx, projectID, problemID)
	if err != nil {
		return ArtifactStatusResponse{}, mapServiceError("load problem before status update", err)
	}
	from := nestedString(current, "problem", "status")
	if err := enforceArchiveOnlyForImmutableStatusChange("problem", from, req.Status, problemImmutableStatuses); err != nil {
		return ArtifactStatusResponse{}, err
	}
	if !isAllowedTransition(from, req.Status, map[string]map[string]struct{}{
		"Draft":    {"Draft": {}, "Locked": {}, "Archived": {}},
		"Locked":   {"Locked": {}, "Archived": {}},
		"Archived": {"Archived": {}, "Draft": {}, "Locked": {}},
	}) {
		return ArtifactStatusResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid problem status transition")
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
		return ArtifactStatusResponse{}, mapServiceError("update problem status", err)
	}
	return ArtifactStatusResponse{ID: toString(updated["id"]), Status: toString(updated["status"]), LastUpdated: toString(updated["lastUpdated"])}, nil
}

func (s *service) ListIdeas(ctx context.Context, projectID string, query listQuery) (IdeaListResponse, error) {
	items, err := s.repo.ListIdeas(ctx, projectID, query)
	if err != nil {
		return IdeaListResponse{}, mapServiceError("list ideas", err)
	}
	out := make([]IdeaListItem, 0, len(items))
	for _, item := range items {
		out = append(out, decodeIdeaListItem(item))
	}
	nextCursor := paginateCursor(len(out), query)
	if nextCursor != nil {
		out = out[:query.Limit]
	}
	return IdeaListResponse{Items: out, NextCursor: nextCursor}, nil
}

func (s *service) CreateIdea(ctx context.Context, projectID, actorUserID string, req createIdeaRequest) (IdeaListItem, error) {
	if err := req.Validate(); err != nil {
		return IdeaListItem{}, err
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
		return IdeaListItem{}, mapServiceError("create idea", err)
	}
	return decodeIdeaListItem(created), nil
}

func (s *service) GetIdea(ctx context.Context, projectID, ideaID string) (IdeaPageResponse, error) {
	item, err := s.repo.GetIdea(ctx, projectID, ideaID)
	if err != nil {
		return IdeaPageResponse{}, mapServiceError("get idea", err)
	}
	ideaRaw := toMap(item["idea"])
	if ideaRaw == nil {
		ideaRaw = map[string]any{}
	}
	return IdeaPageResponse{
		Idea: IdeaPage{
			ID:          toString(ideaRaw["id"]),
			Title:       toString(ideaRaw["title"]),
			Status:      toString(ideaRaw["status"]),
			Owner:       toString(ideaRaw["owner"]),
			LastUpdated: toString(ideaRaw["lastUpdated"]),
		},
		Detail:    toRawJSON(item["detail"], "{}"),
		Reference: toRawJSON(item["reference"], "{}"),
	}, nil
}

func (s *service) UpdateIdea(ctx context.Context, projectID, ideaID, actorUserID string, req updateIdeaRequest) (IdeaListItem, error) {
	if err := req.Validate(); err != nil {
		return IdeaListItem{}, err
	}
	current, err := s.repo.GetIdea(ctx, projectID, ideaID)
	if err != nil {
		return IdeaListItem{}, mapServiceError("load idea before update", err)
	}
	from := nestedString(current, "idea", "status")
	if err := enforceArchiveOnlyForImmutableUpdate("idea", from, req.State, ideaImmutableStatuses); err != nil {
		return IdeaListItem{}, err
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
		return IdeaListItem{}, mapServiceError("update idea", err)
	}
	return decodeIdeaListItem(updated), nil
}

func (s *service) UpdateIdeaStatus(ctx context.Context, projectID, ideaID, actorUserID string, req updateIdeaStatusRequest) (ArtifactStatusResponse, error) {
	if err := req.Validate(); err != nil {
		return ArtifactStatusResponse{}, err
	}
	current, err := s.repo.GetIdea(ctx, projectID, ideaID)
	if err != nil {
		return ArtifactStatusResponse{}, mapServiceError("load idea before status update", err)
	}
	from := nestedString(current, "idea", "status")
	if err := enforceArchiveOnlyForImmutableStatusChange("idea", from, req.Status, ideaImmutableStatuses); err != nil {
		return ArtifactStatusResponse{}, err
	}
	if !isAllowedTransition(from, req.Status, map[string]map[string]struct{}{
		"Considered": {"Considered": {}, "Selected": {}, "Rejected": {}, "Archived": {}},
		"Selected":   {"Selected": {}, "Rejected": {}, "Archived": {}},
		"Rejected":   {"Rejected": {}, "Archived": {}},
		"Archived":   {"Archived": {}, "Considered": {}, "Selected": {}, "Rejected": {}},
	}) {
		return ArtifactStatusResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid idea status transition")
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
		return ArtifactStatusResponse{}, mapServiceError("update idea status", err)
	}
	return ArtifactStatusResponse{ID: toString(updated["id"]), Status: toString(updated["status"]), LastUpdated: toString(updated["lastUpdated"])}, nil
}

func (s *service) ListTasks(ctx context.Context, projectID string, query listQuery) (TaskListResponse, error) {
	items, err := s.repo.ListTasks(ctx, projectID, query)
	if err != nil {
		return TaskListResponse{}, mapServiceError("list tasks", err)
	}
	out := make([]TaskListItem, 0, len(items))
	for _, item := range items {
		out = append(out, decodeTaskListItem(item))
	}
	nextCursor := paginateCursor(len(out), query)
	if nextCursor != nil {
		out = out[:query.Limit]
	}
	return TaskListResponse{Items: out, NextCursor: nextCursor}, nil
}

func (s *service) CreateTask(ctx context.Context, projectID, actorUserID string, req createTaskRequest) (TaskListItem, error) {
	if err := req.Validate(); err != nil {
		return TaskListItem{}, err
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
		return TaskListItem{}, mapServiceError("create task", err)
	}
	return decodeTaskListItem(created), nil
}

func (s *service) GetTask(ctx context.Context, projectID, taskID string) (TaskPageResponse, error) {
	item, err := s.repo.GetTask(ctx, projectID, taskID)
	if err != nil {
		return TaskPageResponse{}, mapServiceError("get task", err)
	}
	taskRaw := toMap(item["task"])
	if taskRaw == nil {
		taskRaw = map[string]any{}
	}
	return TaskPageResponse{
		Task: TaskPage{
			ID:          toString(taskRaw["id"]),
			Title:       toString(taskRaw["title"]),
			Status:      toString(taskRaw["status"]),
			Owner:       toString(taskRaw["owner"]),
			Deadline:    toString(taskRaw["deadline"]),
			LastUpdated: toString(taskRaw["lastUpdated"]),
		},
		Detail:    toRawJSON(item["detail"], "{}"),
		Reference: toRawJSON(item["reference"], "{}"),
	}, nil
}

func (s *service) UpdateTask(ctx context.Context, projectID, taskID, actorUserID string, req updateTaskRequest) (TaskListItem, error) {
	if err := req.Validate(); err != nil {
		return TaskListItem{}, err
	}
	current, err := s.repo.GetTask(ctx, projectID, taskID)
	if err != nil {
		return TaskListItem{}, mapServiceError("load task before update", err)
	}
	from := nestedString(current, "task", "status")
	if err := enforceArchiveOnlyForImmutableUpdate("task", from, req.State, taskImmutableStatuses); err != nil {
		return TaskListItem{}, err
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
		return TaskListItem{}, mapServiceError("update task", err)
	}
	return decodeTaskListItem(updated), nil
}

func (s *service) UpdateTaskStatus(ctx context.Context, projectID, taskID, actorUserID string, req updateTaskStatusRequest) (ArtifactStatusResponse, error) {
	if err := req.Validate(); err != nil {
		return ArtifactStatusResponse{}, err
	}
	current, err := s.repo.GetTask(ctx, projectID, taskID)
	if err != nil {
		return ArtifactStatusResponse{}, mapServiceError("load task before status update", err)
	}
	from := nestedString(current, "task", "status")
	if err := enforceArchiveOnlyForImmutableStatusChange("task", from, req.Status, taskImmutableStatuses); err != nil {
		return ArtifactStatusResponse{}, err
	}
	if !isAllowedTransition(from, req.Status, map[string]map[string]struct{}{
		"Planned":     {"Planned": {}, "In Progress": {}, "Abandoned": {}},
		"In Progress": {"In Progress": {}, "Completed": {}, "Abandoned": {}},
		"Completed":   {"Completed": {}},
		"Abandoned":   {"Abandoned": {}},
	}) {
		return ArtifactStatusResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid task status transition")
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
		return ArtifactStatusResponse{}, mapServiceError("update task status", err)
	}
	return ArtifactStatusResponse{ID: toString(updated["id"]), Status: toString(updated["status"]), LastUpdated: toString(updated["lastUpdated"])}, nil
}

func (s *service) ListFeedback(ctx context.Context, projectID string, query listQuery) (FeedbackListResponse, error) {
	items, err := s.repo.ListFeedback(ctx, projectID, query)
	if err != nil {
		return FeedbackListResponse{}, mapServiceError("list feedback", err)
	}
	out := make([]FeedbackListItem, 0, len(items))
	for _, item := range items {
		out = append(out, decodeFeedbackListItem(item))
	}
	nextCursor := paginateCursor(len(out), query)
	if nextCursor != nil {
		out = out[:query.Limit]
	}
	return FeedbackListResponse{Items: out, NextCursor: nextCursor}, nil
}

func (s *service) CreateFeedback(ctx context.Context, projectID, actorUserID string, req createFeedbackRequest) (FeedbackListItem, error) {
	if err := req.Validate(); err != nil {
		return FeedbackListItem{}, err
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
		return FeedbackListItem{}, mapServiceError("create feedback", err)
	}
	return decodeFeedbackListItem(created), nil
}

func (s *service) GetFeedback(ctx context.Context, projectID, feedbackID string) (FeedbackPageResponse, error) {
	item, err := s.repo.GetFeedback(ctx, projectID, feedbackID)
	if err != nil {
		return FeedbackPageResponse{}, mapServiceError("get feedback", err)
	}
	feedbackRaw := toMap(item["feedback"])
	metadataRaw := toMap(item["metadata"])
	if feedbackRaw == nil {
		feedbackRaw = map[string]any{}
	}
	if metadataRaw == nil {
		metadataRaw = map[string]any{}
	}
	return FeedbackPageResponse{
		Feedback: FeedbackPage{
			ID:          toString(feedbackRaw["id"]),
			Title:       toString(feedbackRaw["title"]),
			Outcome:     toString(feedbackRaw["outcome"]),
			Status:      toString(feedbackRaw["status"]),
			Owner:       toString(feedbackRaw["owner"]),
			CreatedDate: toString(feedbackRaw["createdDate"]),
			LastUpdated: toString(feedbackRaw["lastUpdated"]),
		},
		Metadata:  decodeArtifactMetadata(metadataRaw),
		Detail:    toRawJSON(item["detail"], "{}"),
		Reference: toRawJSON(item["reference"], "{}"),
	}, nil
}

func (s *service) UpdateFeedback(ctx context.Context, projectID, feedbackID, actorUserID string, req updateFeedbackRequest) (FeedbackListItem, error) {
	if err := req.Validate(); err != nil {
		return FeedbackListItem{}, err
	}
	current, err := s.repo.GetFeedback(ctx, projectID, feedbackID)
	if err != nil {
		return FeedbackListItem{}, mapServiceError("load feedback before update", err)
	}
	from := nestedString(current, "feedback", "status")
	if from == "" {
		from = "Active"
	}
	if status := toString(req.State["status"]); status != "" {
		if !isAllowedTransition(from, status, map[string]map[string]struct{}{
			"Active":   {"Active": {}, "Archived": {}},
			"Archived": {"Archived": {}, "Active": {}},
		}) {
			return FeedbackListItem{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid feedback status transition")
		}
	}
	var updated map[string]any
	err = s.store.WithTx(ctx, func(txCtx context.Context) error {
		result, updateErr := s.repo.UpdateFeedback(txCtx, projectID, feedbackID, actorUserID, req.State)
		if updateErr != nil {
			return updateErr
		}
		updated = result
		return nil
	})
	if err != nil {
		return FeedbackListItem{}, mapServiceError("update feedback", err)
	}
	return decodeFeedbackListItem(updated), nil
}

func paginateCursor(itemCount int, query listQuery) *string {
	if itemCount <= query.Limit {
		return nil
	}
	cursor := pagination.EncodeOffsetCursor(query.Offset + query.Limit)
	return &cursor
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
