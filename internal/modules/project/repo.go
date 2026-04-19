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
	Dashboard(ctx context.Context, projectID string) (projectDashboardResponse, error)
	DashboardSummary(ctx context.Context, projectID string) (projectDashboardSummaryResponse, error)
	DashboardMyWork(ctx context.Context, projectID string) (projectDashboardMyWorkResponse, error)
	DashboardEvents(ctx context.Context, projectID string) (projectDashboardEventsResponse, error)
	DashboardActivity(ctx context.Context, projectID string) (projectDashboardActivityResponse, error)
	Overview(ctx context.Context, projectID string) (projectOverviewResponse, error)
	Search(ctx context.Context, projectID, q string, limit int) (projectSearchResponse, error)
	GetUser(ctx context.Context, userID string) (accessUser, error)
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
WHERE id::text = $1
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
	SELECT s.id::text AS id, 'Story' AS artifact_type, s.title, '/project/' || $1::text || '/stories/' || s.id::text AS href, s.updated_at
	FROM stories s
	WHERE s.project_id = $1::uuid
	UNION ALL
	SELECT j.id::text AS id, 'Journey' AS artifact_type, j.title, '/project/' || $1::text || '/journeys/' || j.id::text AS href, j.updated_at
	FROM journeys j
	WHERE j.project_id = $1::uuid
	UNION ALL
	SELECT pr.id::text AS id, 'Problem' AS artifact_type, pr.title, '/project/' || $1::text || '/problem-statement/' || pr.id::text AS href, pr.updated_at
	FROM problems pr
	WHERE pr.project_id = $1::uuid
	UNION ALL
	SELECT i.id::text AS id, 'Idea' AS artifact_type, i.title, '/project/' || $1::text || '/ideas/' || i.id::text AS href, i.updated_at
	FROM ideas i
	WHERE i.project_id = $1::uuid
	UNION ALL
	SELECT t.id::text AS id, 'Task' AS artifact_type, t.title, '/project/' || $1::text || '/tasks/' || t.id::text AS href, t.updated_at
	FROM tasks t
	WHERE t.project_id = $1::uuid
	UNION ALL
	SELECT f.id::text AS id, 'Feedback' AS artifact_type, f.title, '/project/' || $1::text || '/feedback/' || f.id::text AS href, f.updated_at
	FROM feedback f
	WHERE f.project_id = $1::uuid
	UNION ALL
	SELECT r.id::text AS id, 'Resource' AS artifact_type, r.title, '/project/' || $1::text || '/resources/' || r.id::text AS href, r.updated_at
	FROM resources r
	WHERE r.project_id = $1::uuid
	UNION ALL
	SELECT pg.id::text AS id, 'Page' AS artifact_type, pg.title, '/project/' || $1::text || '/pages/' || pg.id::text AS href, pg.updated_at
	FROM pages pg
	WHERE pg.project_id = $1::uuid
) edits
ORDER BY updated_at DESC
LIMIT 10
`

const queryGetDashboardSummary = `
SELECT
	(SELECT COUNT(1) FROM stories s WHERE s.project_id = $1::uuid),
	(SELECT COUNT(1) FROM journeys j WHERE j.project_id = $1::uuid),
	(SELECT COUNT(1) FROM problems p WHERE p.project_id = $1::uuid),
	(SELECT COUNT(1) FROM ideas i WHERE i.project_id = $1::uuid),
	(SELECT COUNT(1) FROM tasks t WHERE t.project_id = $1::uuid),
	(SELECT COUNT(1) FROM feedback f WHERE f.project_id = $1::uuid),
	(SELECT COUNT(1) FROM stories s WHERE s.project_id = $1::uuid AND s.is_orphan),
	(SELECT COUNT(1) FROM journeys j WHERE j.project_id = $1::uuid AND j.is_orphan),
	(SELECT COUNT(1) FROM problems p WHERE p.project_id = $1::uuid AND p.status = 'Locked'::problem_status),
	(SELECT COUNT(1)
	 FROM problems p
	 WHERE p.project_id = $1::uuid
	   AND NOT EXISTS (
		   SELECT 1
		   FROM ideas i
		   WHERE i.project_id = p.project_id
			 AND i.primary_problem_id = p.id
	   )),
	(SELECT COUNT(1) FROM ideas i WHERE i.project_id = $1::uuid AND i.status = 'Selected'::idea_status),
	(SELECT COUNT(1)
	 FROM ideas i
	 WHERE i.project_id = $1::uuid
	   AND i.status = 'Selected'::idea_status
	   AND NOT EXISTS (
		   SELECT 1
		   FROM tasks t
		   WHERE t.project_id = i.project_id
			 AND t.primary_idea_id = i.id
	   )),
	(SELECT COUNT(1)
	 FROM tasks t
	 WHERE t.project_id = $1::uuid
	   AND t.status NOT IN ('Completed'::task_status, 'Abandoned'::task_status)),
	(SELECT COUNT(1)
	 FROM tasks t
	 WHERE t.project_id = $1::uuid
	   AND t.status NOT IN ('Completed'::task_status, 'Abandoned'::task_status)
	   AND t.due_at IS NOT NULL
	   AND t.due_at < NOW()),
	(SELECT COUNT(1) FROM tasks t WHERE t.project_id = $1::uuid AND t.status = 'Completed'::task_status),
	(SELECT COUNT(1)
	 FROM tasks t
	 WHERE t.project_id = $1::uuid
	   AND t.status IN ('Blocked'::task_status, 'Abandoned'::task_status)),
	(SELECT COUNT(1)
	 FROM tasks t
	 WHERE t.project_id = $1::uuid
	   AND t.status = 'Completed'::task_status
	   AND NOT EXISTS (
		   SELECT 1
		   FROM feedback f
		   WHERE f.project_id = t.project_id
			 AND f.primary_task_id = t.id
	   )),
	(SELECT COUNT(1)
	 FROM feedback f
	 WHERE f.project_id = $1::uuid
	   AND COALESCE(f.outcome::text, 'Needs Iteration') = 'Needs Iteration')
`

const queryListDashboardMyTasks = `
SELECT
	t.id::text,
	t.title,
	t.status::text,
	COALESCE(to_char(t.due_at, 'YYYY-MM-DD'), '')
FROM tasks t
WHERE t.project_id = $1::uuid
ORDER BY
	CASE WHEN t.due_at IS NULL THEN 1 ELSE 0 END,
	t.due_at ASC,
	t.updated_at DESC
LIMIT 5
`

const queryListDashboardMyFeedback = `
SELECT
	f.id::text,
	f.title,
	COALESCE(f.outcome::text, 'Needs Iteration')
FROM feedback f
WHERE f.project_id = $1::uuid
ORDER BY f.updated_at DESC
LIMIT 5
`

const querySearchProjectArtifacts = `
WITH normalized AS (
	SELECT
		NULLIF(trim($2::text), '') AS q,
		NULLIF(websearch_to_tsquery('english', trim($2::text))::text, '') AS web_text,
		NULLIF(plainto_tsquery('english', trim($2::text))::text, '') AS plain_text
), picked AS (
	SELECT
		CASE
			WHEN q IS NULL THEN NULL::tsquery
			WHEN web_text IS NOT NULL THEN websearch_to_tsquery('english', q)
			WHEN plain_text IS NOT NULL THEN plainto_tsquery('english', q)
			ELSE NULL::tsquery
		END AS tsq
	FROM normalized
)
SELECT
	si.artifact_id::text AS id,
	si.artifact_type,
	si.title,
	COALESCE(NULLIF(si.description, ''), '') AS description,
	COALESCE(si.status, '') AS status,
	si.href,
	si.updated_at
FROM search_index si
JOIN picked p ON p.tsq IS NOT NULL
WHERE si.project_id = $1::uuid
	AND si.search_vector @@ p.tsq
ORDER BY ts_rank(si.search_vector, p.tsq) DESC, si.updated_at DESC
LIMIT $3
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

func (r *repo) Dashboard(ctx context.Context, projectID string) (projectDashboardResponse, error) {
	summary, err := r.DashboardSummary(ctx, projectID)
	if err != nil {
		return projectDashboardResponse{}, err
	}

	myWork, err := r.DashboardMyWork(ctx, projectID)
	if err != nil {
		return projectDashboardResponse{}, err
	}

	events, err := r.DashboardEvents(ctx, projectID)
	if err != nil {
		return projectDashboardResponse{}, err
	}

	activity, err := r.DashboardActivity(ctx, projectID)
	if err != nil {
		return projectDashboardResponse{}, err
	}

	return projectDashboardResponse{
		Project:  summary.Project,
		Summary:  summary.Summary,
		Events:   events.Events,
		Activity: activity.Activity,
		MyWork: projectDashboardMyWorkBlock{
			FocusUser:   myWork.Me,
			Tasks:       myWork.MyTasks,
			Feedback:    myWork.MyFeedback,
			RecentEdits: myWork.RecentEdits,
		},
	}, nil
}

func (r *repo) DashboardSummary(ctx context.Context, projectID string) (projectDashboardSummaryResponse, error) {
	if err := r.requireStore(); err != nil {
		return projectDashboardSummaryResponse{}, err
	}

	identity, err := r.resolveProjectIdentity(ctx, projectID)
	if err != nil {
		return projectDashboardSummaryResponse{}, err
	}

	summary, err := r.loadDashboardSummary(ctx, identity.UUID)
	if err != nil {
		return projectDashboardSummaryResponse{}, err
	}

	return projectDashboardSummaryResponse{
		Project: dashboardProject{ID: identity.UUID, Name: identity.Name, Status: identity.Status},
		Summary: summary,
	}, nil
}

func (r *repo) DashboardMyWork(ctx context.Context, projectID string) (projectDashboardMyWorkResponse, error) {
	if err := r.requireStore(); err != nil {
		return projectDashboardMyWorkResponse{}, err
	}

	identity, err := r.resolveProjectIdentity(ctx, projectID)
	if err != nil {
		return projectDashboardMyWorkResponse{}, err
	}

	myTasks, err := r.loadDashboardMyTasks(ctx, identity.UUID)
	if err != nil {
		return projectDashboardMyWorkResponse{}, err
	}

	myFeedback, err := r.loadDashboardMyFeedback(ctx, identity.UUID)
	if err != nil {
		return projectDashboardMyWorkResponse{}, err
	}

	recentEdits, err := r.loadDashboardRecentEdits(ctx, identity.UUID)
	if err != nil {
		return projectDashboardMyWorkResponse{}, err
	}

	return projectDashboardMyWorkResponse{
		Me:          dashboardUser{ID: "project", Name: "Project Team", Initials: "PT"},
		MyTasks:     myTasks,
		MyFeedback:  myFeedback,
		RecentEdits: recentEdits,
	}, nil
}

func (r *repo) DashboardEvents(ctx context.Context, projectID string) (projectDashboardEventsResponse, error) {
	if err := r.requireStore(); err != nil {
		return projectDashboardEventsResponse{}, err
	}

	identity, err := r.resolveProjectIdentity(ctx, projectID)
	if err != nil {
		return projectDashboardEventsResponse{}, err
	}

	events, err := r.loadDashboardEvents(ctx, identity.UUID)
	if err != nil {
		return projectDashboardEventsResponse{}, err
	}

	return projectDashboardEventsResponse{Events: events}, nil
}

func (r *repo) DashboardActivity(ctx context.Context, projectID string) (projectDashboardActivityResponse, error) {
	if err := r.requireStore(); err != nil {
		return projectDashboardActivityResponse{}, err
	}

	identity, err := r.resolveProjectIdentity(ctx, projectID)
	if err != nil {
		return projectDashboardActivityResponse{}, err
	}

	activity, err := r.loadDashboardActivity(ctx, identity.UUID)
	if err != nil {
		return projectDashboardActivityResponse{}, err
	}

	return projectDashboardActivityResponse{
		Activity: activity,
	}, nil
}

func (r *repo) Overview(ctx context.Context, projectID string) (projectOverviewResponse, error) {
	summary, err := r.DashboardSummary(ctx, projectID)
	if err != nil {
		return projectOverviewResponse{}, err
	}

	return projectOverviewResponse{
		Project: summary.Project,
		Summary: summary.Summary,
	}, nil
}

func (r *repo) Search(ctx context.Context, projectID, q string, limit int) (projectSearchResponse, error) {
	if err := r.requireStore(); err != nil {
		return projectSearchResponse{}, err
	}

	identity, err := r.resolveProjectIdentity(ctx, projectID)
	if err != nil {
		return projectSearchResponse{}, err
	}

	items := make([]projectSearchResultItem, 0, limit)
	err = r.store.Execute(ctx, storage.RelationalQueryMany(querySearchProjectArtifacts,
		func(row storage.RowScanner) error {
			var item projectSearchResultItem
			var updatedAt time.Time
			if scanErr := row.Scan(&item.ID, &item.Type, &item.Title, &item.Description, &item.Status, &item.Href, &updatedAt); scanErr != nil {
				return scanErr
			}
			item.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
			item.Description = strings.TrimSpace(item.Description)
			item.Status = strings.TrimSpace(item.Status)
			items = append(items, item)
			return nil
		},
		identity.UUID,
		strings.TrimSpace(q),
		limit,
	))
	if err != nil {
		return projectSearchResponse{}, wrapRepoError("search project", err)
	}

	return projectSearchResponse{Query: strings.TrimSpace(q), Items: items}, nil
}

func (r *repo) loadDashboardSummary(ctx context.Context, projectUUID string) (dashboardSummary, error) {
	var summaryStories int64
	var summaryJourneys int64
	var summaryProblems int64
	var summaryIdeas int64
	var summaryTasks int64
	var summaryFeedback int64
	var summaryOrphanStories int64
	var summaryOrphanJourneys int64
	var summaryLockedProblems int64
	var summaryProblemsWithoutIdeas int64
	var summarySelectedIdeas int64
	var summarySelectedIdeasWithoutTasks int64
	var summaryOpenTasks int64
	var summaryOverdueTasks int64
	var summaryCompletedTasks int64
	var summaryBlockedOrAbandonedTasks int64
	var summaryCompletedTasksNoFeedback int64
	var summaryFeedbackNeedsIteration int64

	err := r.store.Execute(ctx, storage.RelationalQueryOne(queryGetDashboardSummary,
		func(row storage.RowScanner) error {
			return row.Scan(
				&summaryStories,
				&summaryJourneys,
				&summaryProblems,
				&summaryIdeas,
				&summaryTasks,
				&summaryFeedback,
				&summaryOrphanStories,
				&summaryOrphanJourneys,
				&summaryLockedProblems,
				&summaryProblemsWithoutIdeas,
				&summarySelectedIdeas,
				&summarySelectedIdeasWithoutTasks,
				&summaryOpenTasks,
				&summaryOverdueTasks,
				&summaryCompletedTasks,
				&summaryBlockedOrAbandonedTasks,
				&summaryCompletedTasksNoFeedback,
				&summaryFeedbackNeedsIteration,
			)
		},
		projectUUID,
	))
	if err != nil {
		return dashboardSummary{}, wrapRepoError("load dashboard summary", err)
	}

	return dashboardSummary{
		Stories:                   int(summaryStories),
		Journeys:                  int(summaryJourneys),
		Problems:                  int(summaryProblems),
		Ideas:                     int(summaryIdeas),
		Tasks:                     int(summaryTasks),
		Feedback:                  int(summaryFeedback),
		OrphanStories:             int(summaryOrphanStories),
		OrphanJourneys:            int(summaryOrphanJourneys),
		LockedProblems:            int(summaryLockedProblems),
		ProblemsWithoutIdeas:      int(summaryProblemsWithoutIdeas),
		SelectedIdeas:             int(summarySelectedIdeas),
		SelectedIdeasWithoutTasks: int(summarySelectedIdeasWithoutTasks),
		OpenTasks:                 int(summaryOpenTasks),
		OverdueTasks:              int(summaryOverdueTasks),
		CompletedTasks:            int(summaryCompletedTasks),
		BlockedOrAbandonedTasks:   int(summaryBlockedOrAbandonedTasks),
		CompletedTasksNoFeedback:  int(summaryCompletedTasksNoFeedback),
		FeedbackNeedsIteration:    int(summaryFeedbackNeedsIteration),
	}, nil
}

func (r *repo) loadDashboardMyTasks(ctx context.Context, projectUUID string) ([]dashboardTask, error) {
	items := make([]dashboardTask, 0, 5)
	err := r.store.Execute(ctx, storage.RelationalQueryMany(queryListDashboardMyTasks,
		func(row storage.RowScanner) error {
			var item dashboardTask
			if err := row.Scan(&item.ID, &item.Title, &item.Status, &item.Deadline); err != nil {
				return err
			}
			item.ID = strings.TrimSpace(item.ID)
			items = append(items, item)
			return nil
		},
		projectUUID,
	))
	if err != nil {
		return nil, wrapRepoError("list dashboard my tasks", err)
	}
	return items, nil
}

func (r *repo) loadDashboardMyFeedback(ctx context.Context, projectUUID string) ([]dashboardFeedback, error) {
	items := make([]dashboardFeedback, 0, 5)
	err := r.store.Execute(ctx, storage.RelationalQueryMany(queryListDashboardMyFeedback,
		func(row storage.RowScanner) error {
			var item dashboardFeedback
			if err := row.Scan(&item.ID, &item.Title, &item.Outcome); err != nil {
				return err
			}
			item.ID = strings.TrimSpace(item.ID)
			items = append(items, item)
			return nil
		},
		projectUUID,
	))
	if err != nil {
		return nil, wrapRepoError("list dashboard my feedback", err)
	}
	return items, nil
}

func (r *repo) loadDashboardEvents(ctx context.Context, projectUUID string) ([]dashboardEvent, error) {
	items := make([]dashboardEvent, 0, 10)
	err := r.store.Execute(ctx, storage.RelationalQueryMany(queryListDashboardEvents,
		func(row storage.RowScanner) error {
			var event dashboardEvent
			var startAt time.Time
			if err := row.Scan(&event.ID, &event.Title, &event.Type, &startAt, &event.Creator); err != nil {
				return err
			}
			event.StartAt = startAt.UTC().Format(time.RFC3339)
			event.Initials = initialsFromName(event.Creator)
			items = append(items, event)
			return nil
		},
		projectUUID,
	))
	if err != nil {
		return nil, wrapRepoError("list dashboard events", err)
	}
	return items, nil
}

func (r *repo) loadDashboardActivity(ctx context.Context, projectUUID string) ([]dashboardActivity, error) {
	items := make([]dashboardActivity, 0, 10)
	err := r.store.Execute(ctx, storage.RelationalQueryMany(queryListDashboardActivity,
		func(row storage.RowScanner) error {
			var item dashboardActivity
			var createdAt time.Time
			if err := row.Scan(&item.ID, &item.User, &item.Action, &item.Artifact, &item.Href, &createdAt); err != nil {
				return err
			}
			item.Initials = initialsFromName(item.User)
			item.At = createdAt.UTC().Format(time.RFC3339)
			items = append(items, item)
			return nil
		},
		projectUUID,
	))
	if err != nil {
		return nil, wrapRepoError("list dashboard activity", err)
	}
	return items, nil
}

func (r *repo) loadDashboardRecentEdits(ctx context.Context, projectUUID string) ([]dashboardRecentEdit, error) {
	items := make([]dashboardRecentEdit, 0, 10)
	err := r.store.Execute(ctx, storage.RelationalQueryMany(queryListDashboardRecentEdits,
		func(row storage.RowScanner) error {
			var edit dashboardRecentEdit
			var updatedAt time.Time
			if err := row.Scan(&edit.ID, &edit.Type, &edit.Title, &edit.Href, &updatedAt); err != nil {
				return err
			}
			edit.At = updatedAt.UTC().Format(time.RFC3339)
			items = append(items, edit)
			return nil
		},
		projectUUID,
	))
	if err != nil {
		return nil, wrapRepoError("list dashboard recent edits", err)
	}
	return items, nil
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

	return projectUpdateSettingsResponse{ProjectID: identity.UUID}, nil
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

	return projectArchiveResponse{ProjectID: identity.UUID, Status: "Archived"}, nil
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

	return projectDeleteResponse{ProjectID: identity.UUID, Status: "Deleted"}, nil
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
