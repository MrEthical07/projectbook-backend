package home

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrHomeUserNotFound     = errors.New("home user not found")
	ErrHomeProjectConflict  = errors.New("home project conflict")
	ErrHomeInviteNotFound   = errors.New("home invite not found")
	ErrHomeInviteNotPending = errors.New("home invite not pending")
)

// Repo defines home module persistence operations.
type Repo interface {
	GetUser(ctx context.Context, userID string) (homeUser, error)
	ListProjects(ctx context.Context, userID string, limit, offset int) ([]homeProject, error)
	CreateProject(ctx context.Context, input createProjectInput) (homeProjectRecord, error)
	UpsertProjectMember(ctx context.Context, projectID, userID, role string, permissionMask int64) error
	UpsertProjectSettings(ctx context.Context, projectID, projectName, projectDescription, status string) error
	UpsertRolePermissions(ctx context.Context, projectID, updatedByUserID string, roleMasks map[string]uint64) error
	ListProjectNames(ctx context.Context, userID string) ([]string, error)
	ListKnownUserEmails(ctx context.Context, userID string) ([]string, error)
	ListInvites(ctx context.Context, userID string) ([]homeInvite, error)
	GetInviteTarget(ctx context.Context, inviteID, userID string) (inviteTarget, error)
	MarkInviteAccepted(ctx context.Context, inviteID string) error
	MarkInviteDeclined(ctx context.Context, inviteID string) error
	ListNotifications(ctx context.Context, userID string, limit int) ([]homeNotification, error)
	ListActivity(ctx context.Context, userID string, filter activityFilter) ([]homeActivityItem, error)
	ListDashboardActivity(ctx context.Context, userID string, limit int) ([]dashboardActivityItem, error)
	GetAccountSettings(ctx context.Context, userID string) (homeAccountSettingsResponse, error)
	UpsertAccountSettings(ctx context.Context, userID string, settings homeAccountSettingsResponse) (time.Time, error)
}

type repo struct {
	store storage.RelationalStore
}

// NewRepo constructs a relational home repository.
func NewRepo(store storage.RelationalStore) Repo {
	return &repo{store: store}
}

type createProjectInput struct {
	UserID       string
	Slug         string
	Name         string
	Description  string
	Icon         string
	Organization string
}

type homeProjectRecord struct {
	ProjectUUID string
	Project     homeProject
}

type inviteTarget struct {
	InviteID       string
	ProjectUUID    string
	ProjectSlug    string
	RecipientEmail string
	AssignedRole   string
	PermissionMask int64
	Status         string
	ExpiresAt      time.Time
}

const queryGetUser = `
SELECT id::text, name, email
FROM users
WHERE id = $1::uuid
`

const queryListProjects = `
SELECT
	p.id::text,
	p.slug,
	p.name,
	p.organization_name,
	p.icon,
	COALESCE(p.description, ''),
	pm.role::text,
	(
		SELECT COUNT(1)
		FROM tasks t
		WHERE t.project_id = p.id
			AND t.status NOT IN ('Completed', 'Abandoned')
	),
	p.last_updated_at,
	p.status::text
FROM project_members pm
JOIN projects p ON p.id = pm.project_id
WHERE pm.user_id = $1::uuid
	AND pm.status = 'Active'
ORDER BY p.last_updated_at DESC
LIMIT $2 OFFSET $3
`

const queryCreateProject = `
INSERT INTO projects (
	slug,
	name,
	organization_name,
	icon,
	description,
	status,
	owner_user_id,
	created_by_user_id
)
VALUES ($1, $2, $3, $4, NULLIF($5, ''), 'Active', $6::uuid, $6::uuid)
RETURNING id::text, slug, name, organization_name, icon, COALESCE(description, ''), last_updated_at, status::text
`

const queryUpsertProjectMember = `
INSERT INTO project_members (
	project_id,
	user_id,
	role,
	permission_mask,
	is_custom,
	status,
	joined_at
)
VALUES ($1::uuid, $2::uuid, $3::project_role, $4, FALSE, 'Active', CURRENT_DATE)
ON CONFLICT (project_id, user_id) DO UPDATE
SET
	role = EXCLUDED.role,
	permission_mask = EXCLUDED.permission_mask,
	is_custom = FALSE,
	status = 'Active',
	joined_at = COALESCE(project_members.joined_at, CURRENT_DATE),
	updated_at = NOW()
`

const queryUpsertProjectSettings = `
INSERT INTO project_settings (
	project_id,
	project_name,
	project_description,
	project_status,
	updated_at
)
VALUES ($1::uuid, $2, NULLIF($3, ''), $4::project_status, NOW())
ON CONFLICT (project_id) DO UPDATE
SET
	project_name = EXCLUDED.project_name,
	project_description = EXCLUDED.project_description,
	project_status = EXCLUDED.project_status,
	updated_at = NOW()
`

const queryUpsertRolePermission = `
INSERT INTO role_permissions (
	project_id,
	role,
	permission_mask,
	updated_by_user_id,
	updated_at
)
VALUES ($1::uuid, $2::project_role, $3, $4::uuid, NOW())
ON CONFLICT (project_id, role) DO UPDATE
SET
	permission_mask = EXCLUDED.permission_mask,
	updated_by_user_id = EXCLUDED.updated_by_user_id,
	updated_at = NOW()
`

const queryListProjectNames = `
SELECT DISTINCT p.name
FROM project_members pm
JOIN projects p ON p.id = pm.project_id
WHERE pm.user_id = $1::uuid
	AND pm.status = 'Active'
ORDER BY p.name ASC
`

const queryListKnownUserEmails = `
SELECT DISTINCT u.email
FROM project_members self
JOIN project_members peer ON peer.project_id = self.project_id
JOIN users u ON u.id = peer.user_id
WHERE self.user_id = $1::uuid
	AND self.status = 'Active'
	AND peer.status = 'Active'
ORDER BY u.email ASC
LIMIT 200
`

const queryListInvites = `
SELECT
	i.id::text,
	p.name,
	COALESCE(p.description, ''),
	p.status::text,
	p.slug,
	p.organization_name,
	inviter.name,
	i.inviter_role::text,
	inviter.email,
	i.assigned_role::text,
	i.sent_at,
	i.expires_at
FROM project_invites i
JOIN projects p ON p.id = i.project_id
JOIN users inviter ON inviter.id = i.invited_by_user_id
JOIN users me ON me.id = $1::uuid
WHERE i.email = me.email
	AND i.status = 'pending'
ORDER BY i.sent_at DESC
`

const queryGetInviteTarget = `
SELECT
	i.id::text,
	i.project_id::text,
	p.slug,
	i.email,
	i.assigned_role::text,
	i.permission_mask,
	i.status::text,
	i.expires_at
FROM project_invites i
JOIN projects p ON p.id = i.project_id
WHERE i.id = $1::uuid
`

const queryMarkInviteAccepted = `
UPDATE project_invites
SET
	status = 'accepted',
	accepted_at = NOW(),
	updated_at = NOW()
WHERE id = $1::uuid
	AND status = 'pending'
`

const queryMarkInviteDeclined = `
UPDATE project_invites
SET
	status = 'declined',
	declined_at = NOW(),
	updated_at = NOW()
WHERE id = $1::uuid
	AND status = 'pending'
`

const queryListNotifications = `
SELECT
	n.id::text,
	COALESCE(NULLIF(n.message, ''), n.title),
	n.created_at,
	n.is_read,
	n.source_type::text
FROM notifications n
WHERE n.user_id = $1::uuid
ORDER BY n.created_at DESC
LIMIT $2
`

const queryGetAccountSettings = `
SELECT
	COALESCE(s.display_name, u.name),
	u.email,
	COALESCE(s.bio, ''),
	COALESCE(s.theme, 'System'),
	COALESCE(s.density, 'Comfortable'),
	COALESCE(s.landing, 'Last Project'),
	COALESCE(s.time_format, '24-hour'),
	COALESCE(s.in_app_notifications, TRUE),
	COALESCE(s.email_notifications, TRUE)
FROM users u
LEFT JOIN account_settings s ON s.user_id = u.id
WHERE u.id = $1::uuid
`

const queryUpsertAccountSettings = `
INSERT INTO account_settings (
	user_id,
	display_name,
	bio,
	theme,
	density,
	landing,
	time_format,
	in_app_notifications,
	email_notifications,
	updated_at
)
VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
ON CONFLICT (user_id) DO UPDATE
SET
	display_name = EXCLUDED.display_name,
	bio = EXCLUDED.bio,
	theme = EXCLUDED.theme,
	density = EXCLUDED.density,
	landing = EXCLUDED.landing,
	time_format = EXCLUDED.time_format,
	in_app_notifications = EXCLUDED.in_app_notifications,
	email_notifications = EXCLUDED.email_notifications,
	updated_at = NOW()
RETURNING updated_at
`

func (r *repo) GetUser(ctx context.Context, userID string) (homeUser, error) {
	if err := r.requireStore(); err != nil {
		return homeUser{}, err
	}

	var user homeUser
	err := r.store.Execute(ctx, storage.RelationalQueryOne(queryGetUser,
		func(row storage.RowScanner) error {
			return row.Scan(&user.ID, &user.Name, &user.Email)
		},
		strings.TrimSpace(userID),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return homeUser{}, ErrHomeUserNotFound
		}
		return homeUser{}, wrapRepoError("get user", err)
	}

	return user, nil
}

func (r *repo) ListProjects(ctx context.Context, userID string, limit, offset int) ([]homeProject, error) {
	if err := r.requireStore(); err != nil {
		return nil, err
	}

	projects := make([]homeProject, 0, limit)
	err := r.store.Execute(ctx, storage.RelationalQueryMany(queryListProjects,
		func(row storage.RowScanner) error {
			var projectUUID string
			var project homeProject
			var lastUpdatedAt time.Time
			if err := row.Scan(
				&projectUUID,
				&project.ID,
				&project.Name,
				&project.Organization,
				&project.Icon,
				&project.Description,
				&project.Role,
				&project.OpenTasks,
				&lastUpdatedAt,
				&project.Status,
			); err != nil {
				return err
			}
			project.LastUpdatedAt = lastUpdatedAt.UTC().Format(time.RFC3339)
			_ = projectUUID
			projects = append(projects, project)
			return nil
		},
		strings.TrimSpace(userID),
		limit,
		offset,
	))
	if err != nil {
		return nil, wrapRepoError("list projects", err)
	}

	return projects, nil
}

func (r *repo) CreateProject(ctx context.Context, input createProjectInput) (homeProjectRecord, error) {
	if err := r.requireStore(); err != nil {
		return homeProjectRecord{}, err
	}

	var projectRecord homeProjectRecord
	var lastUpdatedAt time.Time
	err := r.store.Execute(ctx, storage.RelationalQueryOne(queryCreateProject,
		func(row storage.RowScanner) error {
			return row.Scan(
				&projectRecord.ProjectUUID,
				&projectRecord.Project.ID,
				&projectRecord.Project.Name,
				&projectRecord.Project.Organization,
				&projectRecord.Project.Icon,
				&projectRecord.Project.Description,
				&lastUpdatedAt,
				&projectRecord.Project.Status,
			)
		},
		input.Slug,
		input.Name,
		input.Organization,
		input.Icon,
		input.Description,
		input.UserID,
	))
	if err != nil {
		if isUniqueViolation(err) {
			return homeProjectRecord{}, ErrHomeProjectConflict
		}
		return homeProjectRecord{}, wrapRepoError("create project", err)
	}

	projectRecord.Project.Role = "Owner"
	projectRecord.Project.OpenTasks = 0
	projectRecord.Project.LastUpdatedAt = lastUpdatedAt.UTC().Format(time.RFC3339)
	return projectRecord, nil
}

func (r *repo) UpsertProjectMember(ctx context.Context, projectID, userID, role string, permissionMask int64) error {
	if err := r.requireStore(); err != nil {
		return err
	}

	err := r.store.Execute(ctx, storage.RelationalExec(
		queryUpsertProjectMember,
		strings.TrimSpace(projectID),
		strings.TrimSpace(userID),
		strings.TrimSpace(role),
		permissionMask,
	))
	if err != nil {
		return wrapRepoError("upsert project member", err)
	}

	return nil
}

func (r *repo) UpsertProjectSettings(ctx context.Context, projectID, projectName, projectDescription, status string) error {
	if err := r.requireStore(); err != nil {
		return err
	}

	err := r.store.Execute(ctx, storage.RelationalExec(
		queryUpsertProjectSettings,
		strings.TrimSpace(projectID),
		strings.TrimSpace(projectName),
		strings.TrimSpace(projectDescription),
		strings.TrimSpace(status),
	))
	if err != nil {
		return wrapRepoError("upsert project settings", err)
	}

	return nil
}

func (r *repo) UpsertRolePermissions(ctx context.Context, projectID, updatedByUserID string, roleMasks map[string]uint64) error {
	if err := r.requireStore(); err != nil {
		return err
	}

	roles := make([]string, 0, len(roleMasks))
	for role := range roleMasks {
		roles = append(roles, role)
	}
	sort.Strings(roles)

	for _, role := range roles {
		mask := roleMasks[role]
		err := r.store.Execute(ctx, storage.RelationalExec(
			queryUpsertRolePermission,
			strings.TrimSpace(projectID),
			role,
			int64(mask),
			strings.TrimSpace(updatedByUserID),
		))
		if err != nil {
			return wrapRepoError("upsert role permission", err)
		}
	}

	return nil
}

func (r *repo) ListProjectNames(ctx context.Context, userID string) ([]string, error) {
	if err := r.requireStore(); err != nil {
		return nil, err
	}

	names := make([]string, 0, 16)
	err := r.store.Execute(ctx, storage.RelationalQueryMany(queryListProjectNames,
		func(row storage.RowScanner) error {
			var name string
			if err := row.Scan(&name); err != nil {
				return err
			}
			names = append(names, name)
			return nil
		},
		strings.TrimSpace(userID),
	))
	if err != nil {
		return nil, wrapRepoError("list project names", err)
	}

	return names, nil
}

func (r *repo) ListKnownUserEmails(ctx context.Context, userID string) ([]string, error) {
	if err := r.requireStore(); err != nil {
		return nil, err
	}

	emails := make([]string, 0, 32)
	err := r.store.Execute(ctx, storage.RelationalQueryMany(queryListKnownUserEmails,
		func(row storage.RowScanner) error {
			var email string
			if err := row.Scan(&email); err != nil {
				return err
			}
			emails = append(emails, strings.ToLower(strings.TrimSpace(email)))
			return nil
		},
		strings.TrimSpace(userID),
	))
	if err != nil {
		return nil, wrapRepoError("list known user emails", err)
	}

	return emails, nil
}

func (r *repo) ListInvites(ctx context.Context, userID string) ([]homeInvite, error) {
	if err := r.requireStore(); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	invites := make([]homeInvite, 0, 16)
	err := r.store.Execute(ctx, storage.RelationalQueryMany(queryListInvites,
		func(row storage.RowScanner) error {
			var invite homeInvite
			var sentAt time.Time
			var expiresAt time.Time
			if err := row.Scan(
				&invite.ID,
				&invite.ProjectName,
				&invite.ProjectDescription,
				&invite.ProjectStatus,
				&invite.ProjectID,
				&invite.OrganizationName,
				&invite.InviterName,
				&invite.InviterRole,
				&invite.InviterEmail,
				&invite.AssignedRole,
				&sentAt,
				&expiresAt,
			); err != nil {
				return err
			}
			invite.SentAt = sentAt.UTC().Format("Jan 2, 2006")
			invite.ExpiresAt = expiresAt.UTC().Format("Jan 2, 2006")
			invite.Expired = expiresAt.UTC().Before(now)
			if !invite.Expired {
				invite.ExpiresSoon = expiresAt.UTC().Before(now.Add(72 * time.Hour))
			}
			invites = append(invites, invite)
			return nil
		},
		strings.TrimSpace(userID),
	))
	if err != nil {
		return nil, wrapRepoError("list invites", err)
	}

	return invites, nil
}

func (r *repo) GetInviteTarget(ctx context.Context, inviteID, userID string) (inviteTarget, error) {
	if err := r.requireStore(); err != nil {
		return inviteTarget{}, err
	}

	var target inviteTarget
	err := r.store.Execute(ctx, storage.RelationalQueryOne(queryGetInviteTarget,
		func(row storage.RowScanner) error {
			return row.Scan(
				&target.InviteID,
				&target.ProjectUUID,
				&target.ProjectSlug,
				&target.RecipientEmail,
				&target.AssignedRole,
				&target.PermissionMask,
				&target.Status,
				&target.ExpiresAt,
			)
		},
		strings.TrimSpace(inviteID),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return inviteTarget{}, ErrHomeInviteNotFound
		}
		return inviteTarget{}, wrapRepoError("get invite target", err)
	}

	user, err := r.GetUser(ctx, userID)
	if err != nil {
		return inviteTarget{}, err
	}
	if !strings.EqualFold(strings.TrimSpace(target.RecipientEmail), strings.TrimSpace(user.Email)) {
		return inviteTarget{}, apperr.New(apperr.CodeForbidden, http.StatusForbidden, "not invite recipient")
	}

	return target, nil
}

func (r *repo) MarkInviteAccepted(ctx context.Context, inviteID string) error {
	if err := r.requireStore(); err != nil {
		return err
	}

	err := r.store.Execute(ctx, storage.RelationalExec(queryMarkInviteAccepted, strings.TrimSpace(inviteID)))
	if err != nil {
		return wrapRepoError("mark invite accepted", err)
	}
	return nil
}

func (r *repo) MarkInviteDeclined(ctx context.Context, inviteID string) error {
	if err := r.requireStore(); err != nil {
		return err
	}

	err := r.store.Execute(ctx, storage.RelationalExec(queryMarkInviteDeclined, strings.TrimSpace(inviteID)))
	if err != nil {
		return wrapRepoError("mark invite declined", err)
	}
	return nil
}

func (r *repo) ListNotifications(ctx context.Context, userID string, limit int) ([]homeNotification, error) {
	if err := r.requireStore(); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	items := make([]homeNotification, 0, limit)
	err := r.store.Execute(ctx, storage.RelationalQueryMany(queryListNotifications,
		func(row storage.RowScanner) error {
			var item homeNotification
			var createdAt time.Time
			if err := row.Scan(
				&item.ID,
				&item.Text,
				&createdAt,
				&item.Read,
				&item.SourceType,
			); err != nil {
				return err
			}
			item.Timestamp = relativeTime(createdAt.UTC(), now)
			item.URL = "/notifications"
			item.Dismissed = false
			items = append(items, item)
			return nil
		},
		strings.TrimSpace(userID),
		limit,
	))
	if err != nil {
		return nil, wrapRepoError("list notifications", err)
	}

	return items, nil
}

func (r *repo) ListActivity(ctx context.Context, userID string, filter activityFilter) ([]homeActivityItem, error) {
	if err := r.requireStore(); err != nil {
		return nil, err
	}

	query, args := buildListActivityQuery(userID, filter)
	now := time.Now().UTC()
	items := make([]homeActivityItem, 0, filter.Limit)
	err := r.store.Execute(ctx, storage.RelationalQueryMany(query,
		func(row storage.RowScanner) error {
			var item homeActivityItem
			var occurredAt time.Time
			if err := row.Scan(
				&item.ID,
				&item.UserName,
				&item.Action,
				&item.ArtifactName,
				&item.ArtifactURL,
				&item.ProjectID,
				&item.ProjectName,
				&item.Type,
				&occurredAt,
			); err != nil {
				return err
			}
			item.UserInitials = initialsFromName(item.UserName)
			item.OccurredAt = occurredAt.UTC().Format(time.RFC3339)
			item.Timestamp = relativeTime(occurredAt.UTC(), now)
			items = append(items, item)
			return nil
		},
		args...,
	))
	if err != nil {
		return nil, wrapRepoError("list activity", err)
	}

	return items, nil
}

func (r *repo) ListDashboardActivity(ctx context.Context, userID string, limit int) ([]dashboardActivityItem, error) {
	if err := r.requireStore(); err != nil {
		return nil, err
	}

	const query = `
SELECT
	a.id::text,
	COALESCE(actor.name, 'Unknown'),
	CASE
		WHEN a.action ILIKE '%comment%' THEN a.action
		ELSE a.action
	END,
	p.name,
	a.created_at,
	(a.actor_user_id = $1::uuid) AS involved
FROM activity_log a
JOIN projects p ON p.id = a.project_id
JOIN project_members me ON me.project_id = a.project_id
	AND me.user_id = $1::uuid
	AND me.status = 'Active'
LEFT JOIN users actor ON actor.id = a.actor_user_id
ORDER BY a.created_at DESC
LIMIT $2
`

	now := time.Now().UTC()
	items := make([]dashboardActivityItem, 0, limit)
	err := r.store.Execute(ctx, storage.RelationalQueryMany(query,
		func(row storage.RowScanner) error {
			var item dashboardActivityItem
			var occurredAt time.Time
			if err := row.Scan(
				&item.ID,
				&item.UserName,
				&item.Action,
				&item.ProjectName,
				&occurredAt,
				&item.Involved,
			); err != nil {
				return err
			}
			item.UserInitials = initialsFromName(item.UserName)
			item.OccurredAt = occurredAt.UTC().Format(time.RFC3339)
			item.Timestamp = relativeTime(occurredAt.UTC(), now)
			items = append(items, item)
			return nil
		},
		strings.TrimSpace(userID),
		limit,
	))
	if err != nil {
		return nil, wrapRepoError("list dashboard activity", err)
	}

	return items, nil
}

func (r *repo) GetAccountSettings(ctx context.Context, userID string) (homeAccountSettingsResponse, error) {
	if err := r.requireStore(); err != nil {
		return homeAccountSettingsResponse{}, err
	}

	var settings homeAccountSettingsResponse
	err := r.store.Execute(ctx, storage.RelationalQueryOne(queryGetAccountSettings,
		func(row storage.RowScanner) error {
			return row.Scan(
				&settings.DisplayName,
				&settings.Email,
				&settings.Bio,
				&settings.Theme,
				&settings.Density,
				&settings.Landing,
				&settings.TimeFormat,
				&settings.InAppNotifications,
				&settings.EmailNotifications,
			)
		},
		strings.TrimSpace(userID),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return homeAccountSettingsResponse{}, ErrHomeUserNotFound
		}
		return homeAccountSettingsResponse{}, wrapRepoError("get account settings", err)
	}

	return settings, nil
}

func (r *repo) UpsertAccountSettings(ctx context.Context, userID string, settings homeAccountSettingsResponse) (time.Time, error) {
	if err := r.requireStore(); err != nil {
		return time.Time{}, err
	}

	var updatedAt time.Time
	err := r.store.Execute(ctx, storage.RelationalQueryOne(queryUpsertAccountSettings,
		func(row storage.RowScanner) error {
			return row.Scan(&updatedAt)
		},
		strings.TrimSpace(userID),
		strings.TrimSpace(settings.DisplayName),
		strings.TrimSpace(settings.Bio),
		strings.TrimSpace(settings.Theme),
		strings.TrimSpace(settings.Density),
		strings.TrimSpace(settings.Landing),
		strings.TrimSpace(settings.TimeFormat),
		settings.InAppNotifications,
		settings.EmailNotifications,
	))
	if err != nil {
		return time.Time{}, wrapRepoError("upsert account settings", err)
	}

	return updatedAt.UTC(), nil
}

func (r *repo) requireStore() error {
	if r == nil || r.store == nil {
		return apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "home repository unavailable")
	}
	return nil
}

func wrapRepoError(action string, err error) error {
	if err == nil {
		return nil
	}
	return apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "failed to process home data"), fmt.Errorf("%s: %w", action, err))
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func buildListActivityQuery(userID string, filter activityFilter) (string, []any) {
	base := `
SELECT
	a.id::text,
	COALESCE(actor.name, 'Unknown'),
	a.action,
	COALESCE(a.payload->>'artifactName', ''),
	COALESCE(a.payload->>'artifactUrl', ''),
	p.slug,
	p.name,
	CASE
		WHEN a.action ILIKE '%comment%' THEN 'Comments'
		WHEN a.artifact_type = 'task' THEN 'Tasks'
		WHEN a.artifact_type = 'feedback' THEN 'Feedback'
		ELSE 'Artifacts'
	END,
	a.created_at
FROM activity_log a
JOIN projects p ON p.id = a.project_id
JOIN project_members me ON me.project_id = a.project_id
	AND me.user_id = $1::uuid
	AND me.status = 'Active'
LEFT JOIN users actor ON actor.id = a.actor_user_id
WHERE 1 = 1
`

	args := []any{strings.TrimSpace(userID)}
	argPos := 2

	if filter.ProjectID != "" {
		base += fmt.Sprintf("\nAND p.slug = $%d", argPos)
		args = append(args, strings.TrimSpace(filter.ProjectID))
		argPos++
	}

	switch filter.Type {
	case "Comments":
		base += "\nAND a.action ILIKE '%comment%'"
	case "Tasks":
		base += "\nAND a.artifact_type = 'task'"
	case "Feedback":
		base += "\nAND a.artifact_type = 'feedback'"
	case "Artifacts":
		base += "\nAND a.action NOT ILIKE '%comment%'"
		base += "\nAND COALESCE(a.artifact_type::text, '') NOT IN ('task', 'feedback')"
	}

	base += fmt.Sprintf("\nORDER BY a.created_at DESC\nLIMIT $%d", argPos)
	args = append(args, filter.Limit)
	return base, args
}

func initialsFromName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "NA"
	}
	parts := strings.Fields(trimmed)
	if len(parts) == 1 {
		runes := []rune(parts[0])
		if len(runes) == 1 {
			return strings.ToUpper(string(runes[0]))
		}
		return strings.ToUpper(string(runes[0]) + string(runes[1]))
	}
	first := []rune(parts[0])
	last := []rune(parts[len(parts)-1])
	return strings.ToUpper(string(first[0]) + string(last[0]))
}

func relativeTime(eventTime, now time.Time) string {
	if eventTime.After(now) {
		return "just now"
	}
	delta := now.Sub(eventTime)
	switch {
	case delta < time.Minute:
		return "just now"
	case delta < time.Hour:
		minutes := int(delta / time.Minute)
		return fmt.Sprintf("%dm ago", minutes)
	case delta < 24*time.Hour:
		hours := int(delta / time.Hour)
		return fmt.Sprintf("%dh ago", hours)
	case delta < 7*24*time.Hour:
		days := int(delta / (24 * time.Hour))
		return fmt.Sprintf("%dd ago", days)
	default:
		return eventTime.UTC().Format("Jan 2")
	}
}
