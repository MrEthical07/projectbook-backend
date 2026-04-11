package project

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrProjectNotFound        = errors.New("project not found")
	ErrProjectAlreadyArchived = errors.New("project already archived")
	ErrProjectConflict        = errors.New("project conflict")
)

// Repo defines project module persistence operations.
type Repo interface {
	Dashboard(ctx context.Context, projectID, userID string) (projectDashboardResponse, error)
	GetUser(ctx context.Context, userID string) (accessUser, error)
	ListUserProjects(ctx context.Context, userID string) ([]sidebarProject, error)
	ListSidebarArtifacts(ctx context.Context, projectID string) (sidebarArtifacts, error)
	GetSettings(ctx context.Context, projectID string) (projectSettingsResponse, error)
	UpdateSettings(ctx context.Context, projectID string, patch projectSettingsPatch) (projectUpdateSettingsResponse, error)
	Archive(ctx context.Context, projectID string) (projectArchiveResponse, error)
	Delete(ctx context.Context, projectID string) (projectDeleteResponse, error)
}

type repo struct {
	store storage.RelationalStore
}

type projectIdentity struct {
	UUID        string
	Slug        string
	Name        string
	Description string
	Status      string
}

// NewRepo constructs a relational project repository.
func NewRepo(store storage.RelationalStore) Repo {
	return &repo{store: store}
}

const queryResolveProjectIdentity = `
SELECT id::text, slug, name, COALESCE(description, ''), status::text
FROM projects
WHERE slug = $1 OR id::text = $1
LIMIT 1
`

const queryGetAccessUser = `
SELECT id::text, name, email
FROM users
WHERE id = $1::uuid
`

const queryListDashboardEvents = `
SELECT
	e.id::text,
	e.title,
	e.event_type::text,
	e.starts_at,
	COALESCE(u.name, 'Unknown')
FROM calendar_events e
LEFT JOIN users u ON u.id = e.owner_user_id
WHERE e.project_id = $1::uuid
ORDER BY e.starts_at ASC
LIMIT 10
`

const queryListDashboardActivity = `
SELECT
	a.id::text,
	COALESCE(u.name, 'Unknown'),
	a.action,
	COALESCE(a.payload->>'artifactName', ''),
	COALESCE(a.payload->>'href', COALESCE(a.payload->>'artifactUrl', '')),
	a.created_at
FROM activity_log a
LEFT JOIN users u ON u.id = a.actor_user_id
WHERE a.project_id = $1::uuid
ORDER BY a.created_at DESC
LIMIT 10
`

const queryListDashboardRecentEdits = `
SELECT id, artifact_type, title, href, updated_at
FROM (
	SELECT s.slug AS id, 'Story' AS artifact_type, s.title, '/project/' || p.slug || '/stories/' || s.slug AS href, s.updated_at
	FROM stories s
	JOIN projects p ON p.id = s.project_id
	WHERE s.project_id = $1::uuid
	UNION ALL
	SELECT j.slug AS id, 'Journey' AS artifact_type, j.title, '/project/' || p.slug || '/journeys/' || j.slug AS href, j.updated_at
	FROM journeys j
	JOIN projects p ON p.id = j.project_id
	WHERE j.project_id = $1::uuid
	UNION ALL
	SELECT pr.slug AS id, 'Problem' AS artifact_type, pr.title, '/project/' || p.slug || '/problem-statement/' || pr.slug AS href, pr.updated_at
	FROM problems pr
	JOIN projects p ON p.id = pr.project_id
	WHERE pr.project_id = $1::uuid
	UNION ALL
	SELECT i.slug AS id, 'Idea' AS artifact_type, i.title, '/project/' || p.slug || '/ideas/' || i.slug AS href, i.updated_at
	FROM ideas i
	JOIN projects p ON p.id = i.project_id
	WHERE i.project_id = $1::uuid
	UNION ALL
	SELECT t.slug AS id, 'Task' AS artifact_type, t.title, '/project/' || p.slug || '/tasks/' || t.slug AS href, t.updated_at
	FROM tasks t
	JOIN projects p ON p.id = t.project_id
	WHERE t.project_id = $1::uuid
	UNION ALL
	SELECT f.slug AS id, 'Feedback' AS artifact_type, f.title, '/project/' || p.slug || '/feedback/' || f.slug AS href, f.updated_at
	FROM feedback f
	JOIN projects p ON p.id = f.project_id
	WHERE f.project_id = $1::uuid
	UNION ALL
	SELECT r.slug AS id, 'Resource' AS artifact_type, r.title, '/project/' || p.slug || '/resources/' || r.slug AS href, r.updated_at
	FROM resources r
	JOIN projects p ON p.id = r.project_id
	WHERE r.project_id = $1::uuid
	UNION ALL
	SELECT pg.slug AS id, 'Page' AS artifact_type, pg.title, '/project/' || p.slug || '/pages/' || pg.slug AS href, pg.updated_at
	FROM pages pg
	JOIN projects p ON p.id = pg.project_id
	WHERE pg.project_id = $1::uuid
) edits
ORDER BY updated_at DESC
LIMIT 10
`

const queryListSidebarProjects = `
SELECT
	p.slug,
	p.name,
	p.icon,
	p.status::text
FROM project_members pm
JOIN projects p ON p.id = pm.project_id
WHERE pm.user_id = $1::uuid
	AND pm.status = 'Active'
ORDER BY p.last_updated_at DESC
LIMIT 50
`

const queryGetSettingsByProjectID = `
SELECT
	COALESCE(ps.project_name, p.name),
	COALESCE(ps.project_description, COALESCE(p.description, '')),
	COALESCE(ps.project_status::text, p.status::text),
	COALESCE(ps.whiteboards_enabled, TRUE),
	COALESCE(ps.advanced_databases_enabled, TRUE),
	COALESCE(ps.calendar_manual_events_enabled, TRUE),
	COALESCE(ps.resource_versioning_enabled, TRUE),
	COALESCE(ps.feedback_aggregation_enabled, TRUE),
	COALESCE(ps.notify_artifact_created, TRUE),
	COALESCE(ps.notify_artifact_locked, TRUE),
	COALESCE(ps.notify_feedback_added, TRUE),
	COALESCE(ps.notify_resource_updated, TRUE),
	COALESCE(ps.delivery_channel, 'In-app')
FROM projects p
LEFT JOIN project_settings ps ON ps.project_id = p.id
WHERE p.id = $1::uuid
LIMIT 1
`

const queryUpdateProjectCore = `
UPDATE projects
SET
	name = $2,
	description = NULLIF($3, ''),
	status = $4::project_status,
	archived_at = CASE
		WHEN $4::project_status = 'Archived' THEN COALESCE(archived_at, NOW())
		ELSE NULL
	END,
	last_updated_at = NOW(),
	updated_at = NOW()
WHERE id = $1::uuid
`

const queryUpsertProjectSettings = `
INSERT INTO project_settings (
	project_id,
	project_name,
	project_description,
	project_status,
	whiteboards_enabled,
	advanced_databases_enabled,
	calendar_manual_events_enabled,
	resource_versioning_enabled,
	feedback_aggregation_enabled,
	notify_artifact_created,
	notify_artifact_locked,
	notify_feedback_added,
	notify_resource_updated,
	delivery_channel,
	updated_at
)
VALUES (
	$1::uuid,
	$2,
	NULLIF($3, ''),
	$4::project_status,
	$5,
	$6,
	$7,
	$8,
	$9,
	$10,
	$11,
	$12,
	$13,
	$14,
	NOW()
)
ON CONFLICT (project_id) DO UPDATE
SET
	project_name = EXCLUDED.project_name,
	project_description = EXCLUDED.project_description,
	project_status = EXCLUDED.project_status,
	whiteboards_enabled = EXCLUDED.whiteboards_enabled,
	advanced_databases_enabled = EXCLUDED.advanced_databases_enabled,
	calendar_manual_events_enabled = EXCLUDED.calendar_manual_events_enabled,
	resource_versioning_enabled = EXCLUDED.resource_versioning_enabled,
	feedback_aggregation_enabled = EXCLUDED.feedback_aggregation_enabled,
	notify_artifact_created = EXCLUDED.notify_artifact_created,
	notify_artifact_locked = EXCLUDED.notify_artifact_locked,
	notify_feedback_added = EXCLUDED.notify_feedback_added,
	notify_resource_updated = EXCLUDED.notify_resource_updated,
	delivery_channel = EXCLUDED.delivery_channel,
	updated_at = NOW()
`

const queryArchiveProject = `
UPDATE projects
SET
	status = 'Archived',
	archived_at = COALESCE(archived_at, NOW()),
	last_updated_at = NOW(),
	updated_at = NOW()
WHERE id = $1::uuid
`

const queryDeleteProject = `
DELETE FROM projects
WHERE id = $1::uuid
`

func (r *repo) Dashboard(ctx context.Context, projectID, userID string) (projectDashboardResponse, error) {
	if err := r.requireStore(); err != nil {
		return projectDashboardResponse{}, err
	}

	identity, err := r.resolveProjectIdentity(ctx, projectID)
	if err != nil {
		return projectDashboardResponse{}, err
	}

	me, err := r.GetUser(ctx, userID)
	if err != nil {
		return projectDashboardResponse{}, err
	}

	response := projectDashboardResponse{
		Project:     dashboardProject{ID: identity.Slug, Name: identity.Name, Status: identity.Status},
		Me:          dashboardUser{ID: me.ID, Name: me.Name, Initials: initialsFromName(me.Name)},
		Events:      make([]dashboardEvent, 0, 10),
		Activity:    make([]dashboardActivity, 0, 10),
		RecentEdits: make([]dashboardRecentEdit, 0, 10),
	}

	err = r.store.Execute(ctx, storage.RelationalQueryMany(queryListDashboardEvents,
		func(row storage.RowScanner) error {
			var event dashboardEvent
			var startAt time.Time
			if err := row.Scan(&event.ID, &event.Title, &event.Type, &startAt, &event.Creator); err != nil {
				return err
			}
			event.StartAt = startAt.UTC().Format(time.RFC3339)
			event.Initials = initialsFromName(event.Creator)
			response.Events = append(response.Events, event)
			return nil
		},
		identity.UUID,
	))
	if err != nil {
		return projectDashboardResponse{}, wrapRepoError("list dashboard events", err)
	}

	err = r.store.Execute(ctx, storage.RelationalQueryMany(queryListDashboardActivity,
		func(row storage.RowScanner) error {
			var item dashboardActivity
			var createdAt time.Time
			if err := row.Scan(&item.ID, &item.User, &item.Action, &item.Artifact, &item.Href, &createdAt); err != nil {
				return err
			}
			item.Initials = initialsFromName(item.User)
			item.At = createdAt.UTC().Format(time.RFC3339)
			response.Activity = append(response.Activity, item)
			return nil
		},
		identity.UUID,
	))
	if err != nil {
		return projectDashboardResponse{}, wrapRepoError("list dashboard activity", err)
	}

	err = r.store.Execute(ctx, storage.RelationalQueryMany(queryListDashboardRecentEdits,
		func(row storage.RowScanner) error {
			var edit dashboardRecentEdit
			var updatedAt time.Time
			if err := row.Scan(&edit.ID, &edit.Type, &edit.Title, &edit.Href, &updatedAt); err != nil {
				return err
			}
			edit.At = updatedAt.UTC().Format(time.RFC3339)
			response.RecentEdits = append(response.RecentEdits, edit)
			return nil
		},
		identity.UUID,
	))
	if err != nil {
		return projectDashboardResponse{}, wrapRepoError("list dashboard recent edits", err)
	}

	return response, nil
}

func (r *repo) GetUser(ctx context.Context, userID string) (accessUser, error) {
	if err := r.requireStore(); err != nil {
		return accessUser{}, err
	}

	var user accessUser
	err := r.store.Execute(ctx, storage.RelationalQueryOne(queryGetAccessUser,
		func(row storage.RowScanner) error {
			return row.Scan(&user.ID, &user.Name, &user.Email)
		},
		strings.TrimSpace(userID),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return accessUser{}, ErrProjectNotFound
		}
		return accessUser{}, wrapRepoError("get user", err)
	}

	return user, nil
}

func (r *repo) ListUserProjects(ctx context.Context, userID string) ([]sidebarProject, error) {
	if err := r.requireStore(); err != nil {
		return nil, err
	}

	projects := make([]sidebarProject, 0, 16)
	err := r.store.Execute(ctx, storage.RelationalQueryMany(queryListSidebarProjects,
		func(row storage.RowScanner) error {
			var item sidebarProject
			if err := row.Scan(&item.ID, &item.Name, &item.Icon, &item.Status); err != nil {
				return err
			}
			projects = append(projects, item)
			return nil
		},
		strings.TrimSpace(userID),
	))
	if err != nil {
		return nil, wrapRepoError("list user projects", err)
	}

	return projects, nil
}

func (r *repo) ListSidebarArtifacts(ctx context.Context, projectID string) (sidebarArtifacts, error) {
	if err := r.requireStore(); err != nil {
		return sidebarArtifacts{}, err
	}

	identity, err := r.resolveProjectIdentity(ctx, projectID)
	if err != nil {
		return sidebarArtifacts{}, err
	}

	artifacts := sidebarArtifacts{
		Stories:  make([]sidebarArtifact, 0),
		Journeys: make([]sidebarArtifact, 0),
		Problems: make([]sidebarArtifact, 0),
		Ideas:    make([]sidebarArtifact, 0),
		Tasks:    make([]sidebarArtifact, 0),
		Feedback: make([]sidebarArtifact, 0),
		Pages:    make([]sidebarArtifact, 0),
	}

	stories, err := r.listArtifactsByTable(ctx, identity.UUID, "stories")
	if err != nil {
		return sidebarArtifacts{}, err
	}
	artifacts.Stories = stories

	journeys, err := r.listArtifactsByTable(ctx, identity.UUID, "journeys")
	if err != nil {
		return sidebarArtifacts{}, err
	}
	artifacts.Journeys = journeys

	problems, err := r.listArtifactsByTable(ctx, identity.UUID, "problems")
	if err != nil {
		return sidebarArtifacts{}, err
	}
	artifacts.Problems = problems

	ideas, err := r.listArtifactsByTable(ctx, identity.UUID, "ideas")
	if err != nil {
		return sidebarArtifacts{}, err
	}
	artifacts.Ideas = ideas

	tasks, err := r.listArtifactsByTable(ctx, identity.UUID, "tasks")
	if err != nil {
		return sidebarArtifacts{}, err
	}
	artifacts.Tasks = tasks

	feedback, err := r.listArtifactsByTable(ctx, identity.UUID, "feedback")
	if err != nil {
		return sidebarArtifacts{}, err
	}
	artifacts.Feedback = feedback

	pages, err := r.listArtifactsByTable(ctx, identity.UUID, "pages")
	if err != nil {
		return sidebarArtifacts{}, err
	}
	artifacts.Pages = pages

	return artifacts, nil
}

func (r *repo) GetSettings(ctx context.Context, projectID string) (projectSettingsResponse, error) {
	if err := r.requireStore(); err != nil {
		return projectSettingsResponse{}, err
	}

	identity, err := r.resolveProjectIdentity(ctx, projectID)
	if err != nil {
		return projectSettingsResponse{}, err
	}

	return r.getSettingsByProjectUUID(ctx, identity.UUID)
}

func (r *repo) UpdateSettings(ctx context.Context, projectID string, patch projectSettingsPatch) (projectUpdateSettingsResponse, error) {
	if err := r.requireStore(); err != nil {
		return projectUpdateSettingsResponse{}, err
	}

	identity, err := r.resolveProjectIdentity(ctx, projectID)
	if err != nil {
		return projectUpdateSettingsResponse{}, err
	}

	current, err := r.getSettingsByProjectUUID(ctx, identity.UUID)
	if err != nil {
		return projectUpdateSettingsResponse{}, err
	}

	next := current
	next.ProjectName = strings.TrimSpace(patch.ProjectName)
	next.ProjectStatus = strings.TrimSpace(patch.ProjectStatus)
	if patch.ProjectDescription != nil {
		next.ProjectDescription = strings.TrimSpace(*patch.ProjectDescription)
	}
	if patch.WhiteboardsEnabled != nil {
		next.WhiteboardsEnabled = *patch.WhiteboardsEnabled
	}
	if patch.AdvancedDatabasesEnabled != nil {
		next.AdvancedDatabasesEnabled = *patch.AdvancedDatabasesEnabled
	}
	if patch.CalendarManualEventsEnabled != nil {
		next.CalendarManualEventsEnabled = *patch.CalendarManualEventsEnabled
	}
	if patch.ResourceVersioningEnabled != nil {
		next.ResourceVersioningEnabled = *patch.ResourceVersioningEnabled
	}
	if patch.FeedbackAggregationEnabled != nil {
		next.FeedbackAggregationEnabled = *patch.FeedbackAggregationEnabled
	}
	if patch.NotifyArtifactCreated != nil {
		next.NotifyArtifactCreated = *patch.NotifyArtifactCreated
	}
	if patch.NotifyArtifactLocked != nil {
		next.NotifyArtifactLocked = *patch.NotifyArtifactLocked
	}
	if patch.NotifyFeedbackAdded != nil {
		next.NotifyFeedbackAdded = *patch.NotifyFeedbackAdded
	}
	if patch.NotifyResourceUpdated != nil {
		next.NotifyResourceUpdated = *patch.NotifyResourceUpdated
	}
	if patch.DeliveryChannel != nil {
		next.DeliveryChannel = strings.TrimSpace(*patch.DeliveryChannel)
	}

	err = r.store.Execute(ctx, storage.RelationalExec(
		queryUpdateProjectCore,
		identity.UUID,
		next.ProjectName,
		next.ProjectDescription,
		next.ProjectStatus,
	))
	if err != nil {
		if isUniqueViolation(err) {
			return projectUpdateSettingsResponse{}, ErrProjectConflict
		}
		return projectUpdateSettingsResponse{}, wrapRepoError("update project", err)
	}

	err = r.store.Execute(ctx, storage.RelationalExec(
		queryUpsertProjectSettings,
		identity.UUID,
		next.ProjectName,
		next.ProjectDescription,
		next.ProjectStatus,
		next.WhiteboardsEnabled,
		next.AdvancedDatabasesEnabled,
		next.CalendarManualEventsEnabled,
		next.ResourceVersioningEnabled,
		next.FeedbackAggregationEnabled,
		next.NotifyArtifactCreated,
		next.NotifyArtifactLocked,
		next.NotifyFeedbackAdded,
		next.NotifyResourceUpdated,
		next.DeliveryChannel,
	))
	if err != nil {
		return projectUpdateSettingsResponse{}, wrapRepoError("upsert project settings", err)
	}

	return projectUpdateSettingsResponse{ProjectID: identity.Slug}, nil
}

func (r *repo) Archive(ctx context.Context, projectID string) (projectArchiveResponse, error) {
	if err := r.requireStore(); err != nil {
		return projectArchiveResponse{}, err
	}

	identity, err := r.resolveProjectIdentity(ctx, projectID)
	if err != nil {
		return projectArchiveResponse{}, err
	}
	if strings.EqualFold(identity.Status, "Archived") {
		return projectArchiveResponse{}, ErrProjectAlreadyArchived
	}

	err = r.store.Execute(ctx, storage.RelationalExec(queryArchiveProject, identity.UUID))
	if err != nil {
		return projectArchiveResponse{}, wrapRepoError("archive project", err)
	}

	err = r.store.Execute(ctx, storage.RelationalExec(
		queryUpsertProjectSettings,
		identity.UUID,
		identity.Name,
		identity.Description,
		"Archived",
		true,
		true,
		true,
		true,
		true,
		true,
		true,
		true,
		true,
		"In-app",
	))
	if err != nil {
		return projectArchiveResponse{}, wrapRepoError("sync archived settings", err)
	}

	return projectArchiveResponse{ProjectID: identity.Slug, Status: "Archived"}, nil
}

func (r *repo) Delete(ctx context.Context, projectID string) (projectDeleteResponse, error) {
	if err := r.requireStore(); err != nil {
		return projectDeleteResponse{}, err
	}

	identity, err := r.resolveProjectIdentity(ctx, projectID)
	if err != nil {
		return projectDeleteResponse{}, err
	}

	err = r.store.Execute(ctx, storage.RelationalExec(queryDeleteProject, identity.UUID))
	if err != nil {
		return projectDeleteResponse{}, wrapRepoError("delete project", err)
	}

	return projectDeleteResponse{ProjectID: identity.Slug, Status: "Deleted"}, nil
}

func (r *repo) resolveProjectIdentity(ctx context.Context, projectID string) (projectIdentity, error) {
	if err := r.requireStore(); err != nil {
		return projectIdentity{}, err
	}

	var identity projectIdentity
	err := r.store.Execute(ctx, storage.RelationalQueryOne(queryResolveProjectIdentity,
		func(row storage.RowScanner) error {
			return row.Scan(&identity.UUID, &identity.Slug, &identity.Name, &identity.Description, &identity.Status)
		},
		strings.TrimSpace(projectID),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return projectIdentity{}, ErrProjectNotFound
		}
		return projectIdentity{}, wrapRepoError("resolve project identity", err)
	}

	return identity, nil
}

func (r *repo) getSettingsByProjectUUID(ctx context.Context, projectUUID string) (projectSettingsResponse, error) {
	settings := projectSettingsResponse{}
	err := r.store.Execute(ctx, storage.RelationalQueryOne(queryGetSettingsByProjectID,
		func(row storage.RowScanner) error {
			return row.Scan(
				&settings.ProjectName,
				&settings.ProjectDescription,
				&settings.ProjectStatus,
				&settings.WhiteboardsEnabled,
				&settings.AdvancedDatabasesEnabled,
				&settings.CalendarManualEventsEnabled,
				&settings.ResourceVersioningEnabled,
				&settings.FeedbackAggregationEnabled,
				&settings.NotifyArtifactCreated,
				&settings.NotifyArtifactLocked,
				&settings.NotifyFeedbackAdded,
				&settings.NotifyResourceUpdated,
				&settings.DeliveryChannel,
			)
		},
		strings.TrimSpace(projectUUID),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return projectSettingsResponse{}, ErrProjectNotFound
		}
		return projectSettingsResponse{}, wrapRepoError("get project settings", err)
	}

	return settings, nil
}

func (r *repo) listArtifactsByTable(ctx context.Context, projectUUID, table string) ([]sidebarArtifact, error) {
	query, err := sidebarArtifactsQuery(table)
	if err != nil {
		return nil, err
	}

	items := make([]sidebarArtifact, 0, 32)
	err = r.store.Execute(ctx, storage.RelationalQueryMany(query,
		func(row storage.RowScanner) error {
			var item sidebarArtifact
			if err := row.Scan(&item.ID, &item.Title); err != nil {
				return err
			}
			items = append(items, item)
			return nil
		},
		strings.TrimSpace(projectUUID),
	))
	if err != nil {
		return nil, wrapRepoError("list "+table+" artifacts", err)
	}

	return items, nil
}

func sidebarArtifactsQuery(table string) (string, error) {
	switch table {
	case "stories", "journeys", "problems", "ideas", "tasks", "feedback", "pages":
		return fmt.Sprintf(`
SELECT slug, title
FROM %s
WHERE project_id = $1::uuid
ORDER BY updated_at DESC
LIMIT 100
`, table), nil
	default:
		return "", apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "invalid artifact table")
	}
}

func (r *repo) requireStore() error {
	if r == nil || r.store == nil {
		return apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "project repository unavailable")
	}
	return nil
}

func wrapRepoError(action string, err error) error {
	if err == nil {
		return nil
	}
	return apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "failed to process project data"), fmt.Errorf("%s: %w", action, err))
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
