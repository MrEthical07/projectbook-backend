package project

import (
	"context"
	"errors"
	"net/http"
	"strings"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/rbac"
	"github.com/MrEthical07/superapi/internal/core/storage"
)

// Service defines project module business workflows.
type Service interface {
	Dashboard(ctx context.Context, userID, projectID string) (projectDashboardResponse, error)
	DashboardSummary(ctx context.Context, userID, projectID string) (projectDashboardSummaryResponse, error)
	DashboardMyWork(ctx context.Context, userID, projectID string) (projectDashboardMyWorkResponse, error)
	DashboardEvents(ctx context.Context, userID, projectID string) (projectDashboardEventsResponse, error)
	DashboardActivity(ctx context.Context, userID, projectID string) (projectDashboardActivityResponse, error)
	Access(ctx context.Context, userID, projectID, role string, mask uint64) (projectAccessResponse, error)
	Sidebar(ctx context.Context, userID, projectID string) (projectSidebarResponse, error)
	GetSettings(ctx context.Context, projectID string) (projectSettingsResponse, error)
	UpdateSettings(ctx context.Context, projectID string, req updateProjectSettingsRequest) (projectUpdateSettingsResponse, error)
	Archive(ctx context.Context, projectID string) (projectArchiveResponse, error)
	Delete(ctx context.Context, projectID string) (projectDeleteResponse, error)
}

type service struct {
	store storage.RelationalStore
	repo  Repo
}

// NewService constructs project business workflows.
func NewService(store storage.RelationalStore, repo Repo) Service {
	return &service{store: store, repo: repo}
}

func (s *service) Dashboard(ctx context.Context, userID, projectID string) (projectDashboardResponse, error) {
	if err := requireIdentity(userID, projectID); err != nil {
		return projectDashboardResponse{}, err
	}

	data, err := s.repo.Dashboard(ctx, projectID, userID)
	if err != nil {
		return projectDashboardResponse{}, mapProjectRepoError(err)
	}
	return data, nil
}

func (s *service) DashboardSummary(ctx context.Context, userID, projectID string) (projectDashboardSummaryResponse, error) {
	if err := requireIdentity(userID, projectID); err != nil {
		return projectDashboardSummaryResponse{}, err
	}

	data, err := s.repo.DashboardSummary(ctx, projectID, userID)
	if err != nil {
		return projectDashboardSummaryResponse{}, mapProjectRepoError(err)
	}
	return data, nil
}

func (s *service) DashboardMyWork(ctx context.Context, userID, projectID string) (projectDashboardMyWorkResponse, error) {
	if err := requireIdentity(userID, projectID); err != nil {
		return projectDashboardMyWorkResponse{}, err
	}

	data, err := s.repo.DashboardMyWork(ctx, projectID, userID)
	if err != nil {
		return projectDashboardMyWorkResponse{}, mapProjectRepoError(err)
	}
	return data, nil
}

func (s *service) DashboardEvents(ctx context.Context, userID, projectID string) (projectDashboardEventsResponse, error) {
	if err := requireIdentity(userID, projectID); err != nil {
		return projectDashboardEventsResponse{}, err
	}

	data, err := s.repo.DashboardEvents(ctx, projectID, userID)
	if err != nil {
		return projectDashboardEventsResponse{}, mapProjectRepoError(err)
	}
	return data, nil
}

func (s *service) DashboardActivity(ctx context.Context, userID, projectID string) (projectDashboardActivityResponse, error) {
	if err := requireIdentity(userID, projectID); err != nil {
		return projectDashboardActivityResponse{}, err
	}

	data, err := s.repo.DashboardActivity(ctx, projectID, userID)
	if err != nil {
		return projectDashboardActivityResponse{}, mapProjectRepoError(err)
	}
	return data, nil
}

func (s *service) Access(ctx context.Context, userID, projectID, role string, mask uint64) (projectAccessResponse, error) {
	if err := requireIdentity(userID, projectID); err != nil {
		return projectAccessResponse{}, err
	}

	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return projectAccessResponse{}, mapProjectRepoError(err)
	}

	role = strings.TrimSpace(role)
	if role == "" {
		role = rbac.RoleMember
	}

	return projectAccessResponse{
		User:        user,
		Role:        role,
		Permissions: buildPermissionMatrix(mask),
	}, nil
}

func (s *service) Sidebar(ctx context.Context, userID, projectID string) (projectSidebarResponse, error) {
	if err := requireIdentity(userID, projectID); err != nil {
		return projectSidebarResponse{}, err
	}

	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return projectSidebarResponse{}, mapProjectRepoError(err)
	}

	projects, err := s.repo.ListUserProjects(ctx, userID)
	if err != nil {
		return projectSidebarResponse{}, mapProjectRepoError(err)
	}

	artifacts, err := s.repo.ListSidebarArtifacts(ctx, projectID)
	if err != nil {
		return projectSidebarResponse{}, mapProjectRepoError(err)
	}

	return projectSidebarResponse{
		User:      user,
		Projects:  projects,
		Artifacts: artifacts,
	}, nil
}

func (s *service) GetSettings(ctx context.Context, projectID string) (projectSettingsResponse, error) {
	if strings.TrimSpace(projectID) == "" {
		return projectSettingsResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "projectId is required")
	}

	settings, err := s.repo.GetSettings(ctx, projectID)
	if err != nil {
		return projectSettingsResponse{}, mapProjectRepoError(err)
	}
	return settings, nil
}

func (s *service) UpdateSettings(ctx context.Context, projectID string, req updateProjectSettingsRequest) (projectUpdateSettingsResponse, error) {
	if strings.TrimSpace(projectID) == "" {
		return projectUpdateSettingsResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "projectId is required")
	}
	if err := s.requireStore(); err != nil {
		return projectUpdateSettingsResponse{}, err
	}

	var response projectUpdateSettingsResponse
	err := s.store.WithTx(ctx, func(txCtx context.Context) error {
		updated, err := s.repo.UpdateSettings(txCtx, projectID, req.Settings)
		if err != nil {
			return err
		}
		response = updated
		return nil
	})
	if err != nil {
		return projectUpdateSettingsResponse{}, mapProjectRepoError(err)
	}

	return response, nil
}

func (s *service) Archive(ctx context.Context, projectID string) (projectArchiveResponse, error) {
	if strings.TrimSpace(projectID) == "" {
		return projectArchiveResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "projectId is required")
	}
	if err := s.requireStore(); err != nil {
		return projectArchiveResponse{}, err
	}

	var response projectArchiveResponse
	err := s.store.WithTx(ctx, func(txCtx context.Context) error {
		archived, err := s.repo.Archive(txCtx, projectID)
		if err != nil {
			return err
		}
		response = archived
		return nil
	})
	if err != nil {
		return projectArchiveResponse{}, mapProjectRepoError(err)
	}

	return response, nil
}

func (s *service) Delete(ctx context.Context, projectID string) (projectDeleteResponse, error) {
	if strings.TrimSpace(projectID) == "" {
		return projectDeleteResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "projectId is required")
	}
	if err := s.requireStore(); err != nil {
		return projectDeleteResponse{}, err
	}

	var response projectDeleteResponse
	err := s.store.WithTx(ctx, func(txCtx context.Context) error {
		deleted, err := s.repo.Delete(txCtx, projectID)
		if err != nil {
			return err
		}
		response = deleted
		return nil
	})
	if err != nil {
		return projectDeleteResponse{}, mapProjectRepoError(err)
	}

	return response, nil
}

func (s *service) requireStore() error {
	if s == nil || s.store == nil {
		return apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "project service unavailable")
	}
	return nil
}

func requireIdentity(userID, projectID string) error {
	if strings.TrimSpace(userID) == "" {
		return apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	}
	if strings.TrimSpace(projectID) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "projectId is required")
	}
	return nil
}

func mapProjectRepoError(err error) error {
	if err == nil {
		return nil
	}
	if ae, ok := apperr.AsAppError(err); ok {
		return ae
	}

	switch {
	case errors.Is(err, ErrProjectNotFound):
		return apperr.New(apperr.CodeNotFound, http.StatusNotFound, "project not found")
	case errors.Is(err, ErrProjectAlreadyArchived):
		return apperr.New(apperr.CodeConflict, http.StatusConflict, "project already archived")
	case errors.Is(err, ErrProjectConflict):
		return apperr.New(apperr.CodeConflict, http.StatusConflict, "project settings conflict")
	default:
		return apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "project operation failed"), err)
	}
}

func buildPermissionMatrix(mask uint64) permissionMatrix {
	return permissionMatrix{
		Project:  permSet(mask, rbac.PermProjectView, rbac.PermProjectCreate, rbac.PermProjectEdit, rbac.PermProjectDelete, rbac.PermProjectArchive, rbac.PermProjectStatusChange),
		Story:    permSet(mask, rbac.PermStoryView, rbac.PermStoryCreate, rbac.PermStoryEdit, rbac.PermStoryDelete, rbac.PermStoryArchive, rbac.PermStoryStatusChange),
		Problem:  permSet(mask, rbac.PermProblemView, rbac.PermProblemCreate, rbac.PermProblemEdit, rbac.PermProblemDelete, rbac.PermProblemArchive, rbac.PermProblemStatusChange),
		Idea:     permSet(mask, rbac.PermIdeaView, rbac.PermIdeaCreate, rbac.PermIdeaEdit, rbac.PermIdeaDelete, rbac.PermIdeaArchive, rbac.PermIdeaStatusChange),
		Task:     permSet(mask, rbac.PermTaskView, rbac.PermTaskCreate, rbac.PermTaskEdit, rbac.PermTaskDelete, rbac.PermTaskArchive, rbac.PermTaskStatusChange),
		Feedback: permSet(mask, rbac.PermFeedbackView, rbac.PermFeedbackCreate, rbac.PermFeedbackEdit, rbac.PermFeedbackDelete, rbac.PermFeedbackArchive, rbac.PermFeedbackStatusChange),
		Resource: permSet(mask, rbac.PermResourceView, rbac.PermResourceCreate, rbac.PermResourceEdit, rbac.PermResourceDelete, rbac.PermResourceArchive, rbac.PermResourceStatusChange),
		Page:     permSet(mask, rbac.PermPageView, rbac.PermPageCreate, rbac.PermPageEdit, rbac.PermPageDelete, rbac.PermPageArchive, rbac.PermPageStatusChange),
		Calendar: permSet(mask, rbac.PermCalendarView, rbac.PermCalendarCreate, rbac.PermCalendarEdit, rbac.PermCalendarDelete, rbac.PermCalendarArchive, rbac.PermCalendarStatusChange),
		Member:   permSet(mask, rbac.PermMemberView, rbac.PermMemberCreate, rbac.PermMemberEdit, rbac.PermMemberDelete, rbac.PermMemberArchive, rbac.PermMemberStatusChange),
	}
}

func permSet(mask, view, create, edit, deletePerm, archive, status uint64) permissionSet {
	return permissionSet{
		View:         mask&view != 0,
		Create:       mask&create != 0,
		Edit:         mask&edit != 0,
		Delete:       mask&deletePerm != 0,
		Archive:      mask&archive != 0,
		StatusChange: mask&status != 0,
	}
}
