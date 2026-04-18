package home

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/rbac"
	"github.com/MrEthical07/superapi/internal/core/storage"
)

const defaultProjectOrganization = "ProjectBook"

// Service defines home module business workflows.
type Service interface {
	Dashboard(ctx context.Context, userID string) (homeDashboardResponse, error)
	ListProjects(ctx context.Context, userID string, limit, offset int) ([]homeProject, error)
	CreateProject(ctx context.Context, userID string, req createProjectRequest) (projectCreationResponse, error)
	ProjectReference(ctx context.Context, userID string) (projectReferenceResponse, error)
	ListInvites(ctx context.Context, userID string) ([]homeInvite, error)
	AcceptInvite(ctx context.Context, userID, inviteID string) (inviteAcceptResponse, error)
	DeclineInvite(ctx context.Context, userID, inviteID string) (inviteDeclineResponse, error)
	ListNotifications(ctx context.Context, userID string, limit int) ([]homeNotification, error)
	ListActivity(ctx context.Context, userID string, filter activityFilter) ([]homeActivityItem, error)
	ListDashboardActivity(ctx context.Context, userID string, limit int) ([]dashboardActivityItem, error)
	GetAccountSettings(ctx context.Context, userID string) (homeAccountSettingsResponse, error)
	UpdateAccountSettings(ctx context.Context, userID string, req updateAccountRequest) (updateAccountResponse, error)
	Docs(ctx context.Context, userID string) (docsResponse, error)
}

type service struct {
	store storage.RelationalStore
	repo  Repo
}

// NewService constructs home business workflows.
func NewService(store storage.RelationalStore, repo Repo) Service {
	return &service{store: store, repo: repo}
}

func (s *service) Dashboard(ctx context.Context, userID string) (homeDashboardResponse, error) {
	if _, err := s.ensureUser(ctx, userID); err != nil {
		return homeDashboardResponse{}, err
	}

	projects, err := s.repo.ListProjects(ctx, userID, 20, 0)
	if err != nil {
		return homeDashboardResponse{}, mapHomeRepoError(err)
	}
	invites, err := s.repo.ListInvites(ctx, userID)
	if err != nil {
		return homeDashboardResponse{}, mapHomeRepoError(err)
	}
	notifications, err := s.repo.ListNotifications(ctx, userID, 10)
	if err != nil {
		return homeDashboardResponse{}, mapHomeRepoError(err)
	}
	activity, err := s.repo.ListDashboardActivity(ctx, userID, 10)
	if err != nil {
		return homeDashboardResponse{}, mapHomeRepoError(err)
	}

	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return homeDashboardResponse{}, mapHomeRepoError(err)
	}

	return homeDashboardResponse{
		User:          user,
		Projects:      projects,
		Invites:       invites,
		Notifications: notifications,
		Activity:      activity,
	}, nil
}

func (s *service) ListProjects(ctx context.Context, userID string, limit, offset int) ([]homeProject, error) {
	if _, err := s.ensureUser(ctx, userID); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	projects, err := s.repo.ListProjects(ctx, userID, limit, offset)
	if err != nil {
		return nil, mapHomeRepoError(err)
	}
	return projects, nil
}

func (s *service) CreateProject(ctx context.Context, userID string, req createProjectRequest) (projectCreationResponse, error) {
	if _, err := s.ensureUser(ctx, userID); err != nil {
		return projectCreationResponse{}, err
	}
	if err := s.requireStore(); err != nil {
		return projectCreationResponse{}, err
	}

	slug := slugify(req.Name)
	if slug == "" {
		slug = fmt.Sprintf("project-%d", time.Now().UTC().Unix())
	}
	description := ""
	if req.Description != nil {
		description = strings.TrimSpace(*req.Description)
	}

	var out homeProjectRecord
	err := s.store.WithTx(ctx, func(txCtx context.Context) error {
		created, err := s.repo.CreateProject(txCtx, createProjectInput{
			UserID:       strings.TrimSpace(userID),
			Slug:         slug,
			Name:         strings.TrimSpace(req.Name),
			Description:  description,
			Icon:         strings.TrimSpace(req.Icon),
			Organization: defaultProjectOrganization,
		})
		if err != nil {
			return err
		}

		ownerMask, ok := rbac.DefaultRoleMask(rbac.RoleOwner)
		if !ok {
			return apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "owner role mask unavailable")
		}
		if err := s.repo.UpsertProjectMember(txCtx, created.ProjectUUID, userID, rbac.RoleOwner, int64(ownerMask)); err != nil {
			return err
		}
		if err := s.repo.UpsertProjectSettings(txCtx, created.ProjectUUID, created.Project.Name, created.Project.Description, created.Project.Status); err != nil {
			return err
		}
		if err := s.repo.UpsertRolePermissions(txCtx, created.ProjectUUID, userID, rbac.DefaultRoleMasks()); err != nil {
			return err
		}

		out = created
		return nil
	})
	if err != nil {
		return projectCreationResponse{}, mapHomeRepoError(err)
	}

	return projectCreationResponse{
		ProjectID: out.Project.ID,
		Project:   out.Project,
	}, nil
}

func (s *service) ProjectReference(ctx context.Context, userID string) (projectReferenceResponse, error) {
	if _, err := s.ensureUser(ctx, userID); err != nil {
		return projectReferenceResponse{}, err
	}

	names, err := s.repo.ListProjectNames(ctx, userID)
	if err != nil {
		return projectReferenceResponse{}, mapHomeRepoError(err)
	}
	emails, err := s.repo.ListKnownUserEmails(ctx, userID)
	if err != nil {
		return projectReferenceResponse{}, mapHomeRepoError(err)
	}

	return projectReferenceResponse{
		ExistingProjects: names,
		ExistingUsers:    emails,
	}, nil
}

func (s *service) ListInvites(ctx context.Context, userID string) ([]homeInvite, error) {
	if _, err := s.ensureUser(ctx, userID); err != nil {
		return nil, err
	}

	invites, err := s.repo.ListInvites(ctx, userID)
	if err != nil {
		return nil, mapHomeRepoError(err)
	}
	return invites, nil
}

func (s *service) AcceptInvite(ctx context.Context, userID, inviteID string) (inviteAcceptResponse, error) {
	if _, err := s.ensureUser(ctx, userID); err != nil {
		return inviteAcceptResponse{}, err
	}
	if err := s.requireStore(); err != nil {
		return inviteAcceptResponse{}, err
	}

	target, err := s.repo.GetInviteTarget(ctx, inviteID, userID)
	if err != nil {
		return inviteAcceptResponse{}, mapHomeRepoError(err)
	}

	if !strings.EqualFold(target.Status, "pending") {
		return inviteAcceptResponse{}, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "invite not found")
	}
	if target.ExpiresAt.UTC().Before(time.Now().UTC()) {
		return inviteAcceptResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invite expired")
	}

	role := strings.TrimSpace(target.AssignedRole)
	if role == "" {
		role = rbac.RoleMember
	}

	err = s.store.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.MarkInviteAccepted(txCtx, target.InviteID); err != nil {
			return err
		}
		if err := s.repo.UpsertProjectMember(txCtx, target.ProjectUUID, userID, role, target.PermissionMask); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return inviteAcceptResponse{}, mapHomeRepoError(err)
	}

	return inviteAcceptResponse{
		InviteID:  target.InviteID,
		ProjectID: target.ProjectUUID,
	}, nil
}

func (s *service) DeclineInvite(ctx context.Context, userID, inviteID string) (inviteDeclineResponse, error) {
	if _, err := s.ensureUser(ctx, userID); err != nil {
		return inviteDeclineResponse{}, err
	}
	if err := s.requireStore(); err != nil {
		return inviteDeclineResponse{}, err
	}

	target, err := s.repo.GetInviteTarget(ctx, inviteID, userID)
	if err != nil {
		return inviteDeclineResponse{}, mapHomeRepoError(err)
	}
	if !strings.EqualFold(target.Status, "pending") {
		return inviteDeclineResponse{}, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "invite not found")
	}

	err = s.store.WithTx(ctx, func(txCtx context.Context) error {
		return s.repo.MarkInviteDeclined(txCtx, target.InviteID)
	})
	if err != nil {
		return inviteDeclineResponse{}, mapHomeRepoError(err)
	}

	return inviteDeclineResponse{InviteID: target.InviteID}, nil
}

func (s *service) ListNotifications(ctx context.Context, userID string, limit int) ([]homeNotification, error) {
	if _, err := s.ensureUser(ctx, userID); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	items, err := s.repo.ListNotifications(ctx, userID, limit)
	if err != nil {
		return nil, mapHomeRepoError(err)
	}
	return items, nil
}

func (s *service) ListActivity(ctx context.Context, userID string, filter activityFilter) ([]homeActivityItem, error) {
	if _, err := s.ensureUser(ctx, userID); err != nil {
		return nil, err
	}
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if err := filter.Validate(); err != nil {
		return nil, err
	}

	items, err := s.repo.ListActivity(ctx, userID, filter)
	if err != nil {
		return nil, mapHomeRepoError(err)
	}
	return items, nil
}

func (s *service) ListDashboardActivity(ctx context.Context, userID string, limit int) ([]dashboardActivityItem, error) {
	if _, err := s.ensureUser(ctx, userID); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	items, err := s.repo.ListDashboardActivity(ctx, userID, limit)
	if err != nil {
		return nil, mapHomeRepoError(err)
	}
	return items, nil
}

func (s *service) GetAccountSettings(ctx context.Context, userID string) (homeAccountSettingsResponse, error) {
	if _, err := s.ensureUser(ctx, userID); err != nil {
		return homeAccountSettingsResponse{}, err
	}

	settings, err := s.repo.GetAccountSettings(ctx, userID)
	if err != nil {
		return homeAccountSettingsResponse{}, mapHomeRepoError(err)
	}
	return settings, nil
}

func (s *service) UpdateAccountSettings(ctx context.Context, userID string, req updateAccountRequest) (updateAccountResponse, error) {
	if _, err := s.ensureUser(ctx, userID); err != nil {
		return updateAccountResponse{}, err
	}
	if err := s.requireStore(); err != nil {
		return updateAccountResponse{}, err
	}

	current, err := s.repo.GetAccountSettings(ctx, userID)
	if err != nil {
		return updateAccountResponse{}, mapHomeRepoError(err)
	}

	next := current
	next.DisplayName = strings.TrimSpace(req.Settings.DisplayName)
	if req.Settings.Bio != nil {
		next.Bio = strings.TrimSpace(*req.Settings.Bio)
	}
	if req.Settings.Theme != nil {
		next.Theme = strings.TrimSpace(*req.Settings.Theme)
	}
	if req.Settings.Density != nil {
		next.Density = strings.TrimSpace(*req.Settings.Density)
	}
	if req.Settings.Landing != nil {
		next.Landing = strings.TrimSpace(*req.Settings.Landing)
	}
	if req.Settings.TimeFormat != nil {
		next.TimeFormat = strings.TrimSpace(*req.Settings.TimeFormat)
	}
	if req.Settings.InAppNotifications != nil {
		next.InAppNotifications = *req.Settings.InAppNotifications
	}
	if req.Settings.EmailNotifications != nil {
		next.EmailNotifications = *req.Settings.EmailNotifications
	}

	var updatedAt time.Time
	err = s.store.WithTx(ctx, func(txCtx context.Context) error {
		stamp, err := s.repo.UpsertAccountSettings(txCtx, userID, next)
		if err != nil {
			return err
		}
		updatedAt = stamp
		return nil
	})
	if err != nil {
		return updateAccountResponse{}, mapHomeRepoError(err)
	}

	return updateAccountResponse{UpdatedAt: updatedAt.UTC().Format(time.RFC3339)}, nil
}

func (s *service) Docs(_ context.Context, _ string) (docsResponse, error) {
	return docsResponse{Sections: []string{}}, nil
}

func (s *service) ensureUser(ctx context.Context, userID string) (homeUser, error) {
	trimmed := strings.TrimSpace(userID)
	if trimmed == "" {
		return homeUser{}, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	}

	user, err := s.repo.GetUser(ctx, trimmed)
	if err != nil {
		return homeUser{}, mapHomeRepoError(err)
	}
	return user, nil
}

func (s *service) requireStore() error {
	if s == nil || s.store == nil {
		return apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "home service unavailable")
	}
	return nil
}

func mapHomeRepoError(err error) error {
	if err == nil {
		return nil
	}
	if ae, ok := apperr.AsAppError(err); ok {
		return ae
	}

	switch {
	case errors.Is(err, ErrHomeUserNotFound):
		return apperr.New(apperr.CodeForbidden, http.StatusForbidden, "permission denied")
	case errors.Is(err, ErrHomeProjectConflict):
		return apperr.New(apperr.CodeConflict, http.StatusConflict, "project with this slug already exists")
	case errors.Is(err, ErrHomeInviteNotFound):
		return apperr.New(apperr.CodeNotFound, http.StatusNotFound, "invite not found")
	default:
		return apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "home operation failed"), err)
	}
}

func slugify(value string) string {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	if trimmed == "" {
		return ""
	}

	var b strings.Builder
	lastDash := false
	for _, ch := range trimmed {
		switch {
		case unicode.IsLetter(ch), unicode.IsDigit(ch):
			b.WriteRune(ch)
			lastDash = false
		case ch == '-', ch == '_', unicode.IsSpace(ch):
			if !lastDash {
				b.WriteRune('-')
				lastDash = true
			}
		default:
			if !lastDash {
				b.WriteRune('-')
				lastDash = true
			}
		}
	}

	return strings.Trim(b.String(), "-")
}
