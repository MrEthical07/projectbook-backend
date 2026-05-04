package artifacts

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/patchx"
	"github.com/MrEthical07/superapi/internal/core/storage"
	"github.com/jackc/pgx/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repo interface {
	ResolveProjectIdentity(ctx context.Context, projectID string) (projectIdentity, error)

	ListStories(ctx context.Context, projectID string, query listQuery) ([]map[string]any, error)
	CreateStory(ctx context.Context, projectID, actorUserID, title string, content map[string]any) (map[string]any, error)
	GetStory(ctx context.Context, projectID, slug string) (map[string]any, error)
	UpdateStory(ctx context.Context, projectID, storyID, actorUserID string, patch map[string]any) (map[string]any, error)

	ListJourneys(ctx context.Context, projectID string, query listQuery) ([]map[string]any, error)
	CreateJourney(ctx context.Context, projectID, actorUserID, title string, content map[string]any) (map[string]any, error)
	GetJourney(ctx context.Context, projectID, slug string) (map[string]any, error)
	UpdateJourney(ctx context.Context, projectID, journeyID, actorUserID string, patch map[string]any) (map[string]any, error)

	ListProblems(ctx context.Context, projectID string, query listQuery) ([]map[string]any, error)
	CreateProblem(ctx context.Context, projectID, actorUserID, statement string, content map[string]any) (map[string]any, error)
	GetProblem(ctx context.Context, projectID, slug string) (map[string]any, error)
	UpdateProblem(ctx context.Context, projectID, problemID, actorUserID string, patch map[string]any) (map[string]any, error)

	ListIdeas(ctx context.Context, projectID string, query listQuery) ([]map[string]any, error)
	CreateIdea(ctx context.Context, projectID, actorUserID, title string, content map[string]any) (map[string]any, error)
	GetIdea(ctx context.Context, projectID, slug string) (map[string]any, error)
	UpdateIdea(ctx context.Context, projectID, ideaID, actorUserID string, patch map[string]any) (map[string]any, error)

	ListTasks(ctx context.Context, projectID string, query listQuery) ([]map[string]any, error)
	CreateTask(ctx context.Context, projectID, actorUserID, title string, content map[string]any) (map[string]any, error)
	GetTask(ctx context.Context, projectID, slug string) (map[string]any, error)
	UpdateTask(ctx context.Context, projectID, taskID, actorUserID string, patch map[string]any) (map[string]any, error)

	ListFeedback(ctx context.Context, projectID string, query listQuery) ([]map[string]any, error)
	CreateFeedback(ctx context.Context, projectID, actorUserID, title string, content map[string]any) (map[string]any, error)
	GetFeedback(ctx context.Context, projectID, slug string) (map[string]any, error)
	UpdateFeedback(ctx context.Context, projectID, feedbackID, actorUserID string, patch map[string]any) (map[string]any, error)
}

type repo struct {
	store storage.RelationalStore
	docs  storage.DocumentStore
}

type projectIdentity struct {
	UUID string
	Slug string
}

func NewRepo(store storage.RelationalStore, docs storage.DocumentStore) Repo {
	return &repo{store: store, docs: docs}
}

func (r *repo) ResolveProjectIdentity(ctx context.Context, projectID string) (projectIdentity, error) {
	if err := r.requireStore(); err != nil {
		return projectIdentity{}, err
	}

	var identity projectIdentity
	err := r.store.Execute(ctx, storage.RelationalQueryOne(
		`SELECT id::text, slug FROM projects WHERE id::text = $1 LIMIT 1`,
		func(row storage.RowScanner) error {
			return row.Scan(&identity.UUID, &identity.Slug)
		},
		strings.TrimSpace(projectID),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return projectIdentity{}, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "project not found")
		}
		return projectIdentity{}, wrapRepoError("resolve project identity", err)
	}

	return identity, nil
}

func (r *repo) ListStories(ctx context.Context, projectID string, query listQuery) ([]map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}

	items := make([]map[string]any, 0, query.Limit)
	args := []any{identity.UUID}
	sql := `
SELECT
	s.id::text,
	s.slug,
	s.title,
	COALESCE(s.persona_name, ''),
	s.pain_points_count,
	s.problem_hypotheses_count,
	COALESCE(u.name, ''),
	COALESCE(to_char(s.updated_at, 'YYYY-MM-DD'), ''),
	s.status::text,
	s.is_orphan
FROM stories s
JOIN users u ON u.id = s.owner_user_id
WHERE s.project_id = $1::uuid
`
	if query.Status != "" {
		args = append(args, query.Status)
		sql += fmt.Sprintf(" AND s.status = $%d::story_status\n", len(args))
	}
	args = append(args, query.Offset, query.Limit+1)
	sql += fmt.Sprintf("ORDER BY s.updated_at DESC OFFSET $%d LIMIT $%d", len(args)-1, len(args))

	err = r.store.Execute(ctx, storage.RelationalQueryMany(sql, func(row storage.RowScanner) error {
		var id, slug, title, personaName, owner, lastUpdated, status string
		var painCount, hypothesisCount int
		var isOrphan bool
		if scanErr := row.Scan(&id, &slug, &title, &personaName, &painCount, &hypothesisCount, &owner, &lastUpdated, &status, &isOrphan); scanErr != nil {
			return scanErr
		}
		items = append(items, map[string]any{
			"id":                     id,
			"title":                  title,
			"personaName":            personaName,
			"painPointsCount":        painCount,
			"problemHypothesesCount": hypothesisCount,
			"owner":                  owner,
			"lastUpdated":            lastUpdated,
			"status":                 status,
			"isOrphan":               isOrphan,
		})
		return nil
	}, args...))
	if err != nil {
		return nil, wrapRepoError("list stories", err)
	}
	return items, nil
}

func (r *repo) CreateStory(ctx context.Context, projectID, actorUserID, title string, content map[string]any) (map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}

	slug, err := r.nextUniqueSlug(ctx, "stories", identity.UUID, title)
	if err != nil {
		return nil, err
	}

	ownerName, err := r.resolveUserName(ctx, actorUserID)
	if err != nil {
		return nil, err
	}

	var id, createdSlug, createdTitle, personaName, status, lastUpdated string
	var painCount, hypothesisCount, revision int
	var isOrphan bool
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`INSERT INTO stories (project_id, slug, title, owner_user_id, status, is_orphan)
		 VALUES ($1::uuid, $2, $3, $4::uuid, 'Draft', TRUE)
		 RETURNING id::text, slug, title, COALESCE(persona_name, ''), pain_points_count, problem_hypotheses_count, status::text, is_orphan, COALESCE(to_char(updated_at, 'YYYY-MM-DD'), ''), document_revision`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &createdSlug, &createdTitle, &personaName, &painCount, &hypothesisCount, &status, &isOrphan, &lastUpdated, &revision)
		},
		identity.UUID,
		slug,
		strings.TrimSpace(title),
		strings.TrimSpace(actorUserID),
	))
	if err != nil {
		return nil, wrapRepoError("create story", err)
	}

	if content == nil {
		content = defaultStoryContent(createdTitle)
	}
	if err := r.upsertDocument(ctx, identity.UUID, "story", id, revision, actorUserID, content); err != nil {
		return nil, err
	}

	return map[string]any{
		"id":                     id,
		"title":                  createdTitle,
		"personaName":            personaName,
		"painPointsCount":        painCount,
		"problemHypothesesCount": hypothesisCount,
		"owner":                  ownerName,
		"lastUpdated":            lastUpdated,
		"status":                 status,
		"isOrphan":               isOrphan,
	}, nil
}

func (r *repo) GetStory(ctx context.Context, projectID, slug string) (map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}

	var id, foundSlug, title, status, ownerName, createdAt, lastUpdated string
	var isOrphan bool
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`SELECT s.id::text, s.slug, s.title, s.status::text, COALESCE(u.name, ''), COALESCE(to_char(s.created_at, 'YYYY-MM-DD'), ''), COALESCE(to_char(s.updated_at, 'YYYY-MM-DD'), ''), s.is_orphan
		 FROM stories s
		 JOIN users u ON u.id = s.owner_user_id
		 WHERE s.project_id = $1::uuid AND s.id::text = $2
		 LIMIT 1`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &foundSlug, &title, &status, &ownerName, &createdAt, &lastUpdated, &isOrphan)
		},
		identity.UUID,
		strings.TrimSpace(slug),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "story not found")
		}
		return nil, wrapRepoError("get story", err)
	}

	content, err := r.loadLatestContent(ctx, identity.UUID, "story", id, defaultStoryContent(title))
	if err != nil {
		return nil, err
	}
	addOnSections := toSlice(content["addOnSections"])
	if addOnSections == nil {
		addOnSections = []any{}
	}

	return map[string]any{
		"story": map[string]any{
			"id":          id,
			"title":       title,
			"status":      status,
			"owner":       ownerName,
			"lastUpdated": lastUpdated,
		},
		"metadata": map[string]any{
			"owner":        ownerName,
			"createdBy":    ownerName,
			"createdAt":    createdAt,
			"lastEditedBy": ownerName,
			"lastEditedAt": lastUpdated,
			"lastUpdated":  lastUpdated,
		},
		"detail":        content,
		"addOnCatalog":  storyAddOnCatalog(),
		"addOnSections": addOnSections,
		"reference": map[string]any{
			"permissions": map[string]any{},
		},
		"isOrphan": isOrphan,
	}, nil
}

func (r *repo) UpdateStory(ctx context.Context, projectID, storyID, actorUserID string, patch map[string]any) (map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}

	title := toString(patch["title"])
	if title == "" {
		title = toString(patch["description"])
	}
	status := toString(patch["status"])
	personaName := ""
	if persona := toMap(patch["persona"]); persona != nil {
		personaName = toString(persona["name"])
	}

	var id, slug, outTitle, outPersonaName, outOwner, outLastUpdated, outStatus string
	var outPainCount, outHypothesisCount, revision int
	var isOrphan bool
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`UPDATE stories
		 SET title = COALESCE(NULLIF($3, ''), title),
		 	 status = CASE
		 	 	WHEN NULLIF($4, '')::story_status = 'Archived'::story_status
		 	 		AND stories.status <> 'Archived'::story_status
		 	 	THEN 'Archived'::story_status
		 	 	WHEN NULLIF($4, '')::story_status IS NOT NULL
		 	 		AND stories.status = 'Archived'::story_status
		 	 		AND NULLIF($4, '')::story_status <> 'Archived'::story_status
		 	 	THEN COALESCE(stories.archived_from_status, 'Draft'::story_status)
		 	 	ELSE COALESCE(NULLIF($4, '')::story_status, stories.status)
		 	 END,
		 	 archived_from_status = CASE
		 	 	WHEN NULLIF($4, '')::story_status = 'Archived'::story_status
		 	 		AND stories.status <> 'Archived'::story_status
		 	 	THEN stories.status
		 	 	WHEN NULLIF($4, '')::story_status IS NOT NULL
		 	 		AND stories.status = 'Archived'::story_status
		 	 		AND NULLIF($4, '')::story_status <> 'Archived'::story_status
		 	 	THEN NULL
		 	 	ELSE stories.archived_from_status
		 	 END,
		 	 persona_name = CASE WHEN $5 = '' THEN persona_name ELSE $5 END,
		 	 pain_points_count = CASE WHEN $6 < 0 THEN pain_points_count ELSE $6 END,
		 	 problem_hypotheses_count = CASE WHEN $7 < 0 THEN problem_hypotheses_count ELSE $7 END,
		 	 updated_at = NOW(),
		 	 document_revision = document_revision + 1
		 FROM users u
		 WHERE stories.project_id = $1::uuid
		   AND stories.id::text = $2
		   AND u.id = stories.owner_user_id
		   AND (
		 	 stories.status NOT IN ('Locked'::story_status, 'Archived'::story_status)
		 	 OR NULLIF($4, '')::story_status = 'Archived'::story_status
		 	 OR (
		 	 	stories.status = 'Archived'::story_status
		 	 	AND NULLIF($4, '')::story_status IS NOT NULL
		 	 	AND NULLIF($4, '')::story_status <> 'Archived'::story_status
		 	 )
		   )
		 RETURNING stories.id::text, stories.slug, stories.title, COALESCE(stories.persona_name, ''), stories.pain_points_count,
		 	 stories.problem_hypotheses_count, COALESCE(u.name, ''), COALESCE(to_char(stories.updated_at, 'YYYY-MM-DD'), ''),
		 	 stories.status::text, stories.is_orphan, stories.document_revision`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &slug, &outTitle, &outPersonaName, &outPainCount, &outHypothesisCount, &outOwner, &outLastUpdated, &outStatus, &isOrphan, &revision)
		},
		identity.UUID,
		strings.TrimSpace(storyID),
		title,
		status,
		personaName,
		painCountOrSentinel(patch),
		hypothesisCountOrSentinel(patch),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "story not found")
		}
		return nil, wrapRepoError("update story", err)
	}

	if err := r.upsertDocument(ctx, identity.UUID, "story", id, revision, actorUserID, patch); err != nil {
		return nil, err
	}

	return map[string]any{
		"id":                     id,
		"title":                  outTitle,
		"personaName":            outPersonaName,
		"painPointsCount":        outPainCount,
		"problemHypothesesCount": outHypothesisCount,
		"owner":                  outOwner,
		"lastUpdated":            outLastUpdated,
		"status":                 outStatus,
		"isOrphan":               isOrphan,
	}, nil
}

func (r *repo) ListJourneys(ctx context.Context, projectID string, query listQuery) ([]map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}

	items := make([]map[string]any, 0, query.Limit)
	args := []any{identity.UUID}
	sql := `
SELECT
	j.id::text,
	j.slug,
	j.title,
	COALESCE(u.name, ''),
	COALESCE(to_char(j.updated_at, 'YYYY-MM-DD'), ''),
	j.status::text,
	j.is_orphan
FROM journeys j
JOIN users u ON u.id = j.owner_user_id
WHERE j.project_id = $1::uuid
`
	if query.Status != "" {
		args = append(args, query.Status)
		sql += fmt.Sprintf(" AND j.status = $%d::journey_status\n", len(args))
	}
	args = append(args, query.Offset, query.Limit+1)
	sql += fmt.Sprintf("ORDER BY j.updated_at DESC OFFSET $%d LIMIT $%d", len(args)-1, len(args))

	err = r.store.Execute(ctx, storage.RelationalQueryMany(sql, func(row storage.RowScanner) error {
		var id, slug, title, owner, lastUpdated, status string
		var isOrphan bool
		if scanErr := row.Scan(&id, &slug, &title, &owner, &lastUpdated, &status, &isOrphan); scanErr != nil {
			return scanErr
		}
		items = append(items, map[string]any{
			"id":              id,
			"title":           title,
			"linkedPersonas":  []string{},
			"stagesCount":     0,
			"painPointsCount": 0,
			"owner":           owner,
			"lastUpdated":     lastUpdated,
			"status":          status,
			"isOrphan":        isOrphan,
		})
		return nil
	}, args...))
	if err != nil {
		return nil, wrapRepoError("list journeys", err)
	}
	return items, nil
}

func (r *repo) CreateJourney(ctx context.Context, projectID, actorUserID, title string, content map[string]any) (map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}

	slug, err := r.nextUniqueSlug(ctx, "journeys", identity.UUID, title)
	if err != nil {
		return nil, err
	}
	ownerName, err := r.resolveUserName(ctx, actorUserID)
	if err != nil {
		return nil, err
	}

	var id, createdSlug, createdTitle, status, lastUpdated string
	var revision int
	var isOrphan bool
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`INSERT INTO journeys (project_id, slug, title, owner_user_id, status, is_orphan)
		 VALUES ($1::uuid, $2, $3, $4::uuid, 'Draft', TRUE)
		 RETURNING id::text, slug, title, status::text, COALESCE(to_char(updated_at, 'YYYY-MM-DD'), ''), document_revision, is_orphan`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &createdSlug, &createdTitle, &status, &lastUpdated, &revision, &isOrphan)
		},
		identity.UUID,
		slug,
		strings.TrimSpace(title),
		strings.TrimSpace(actorUserID),
	))
	if err != nil {
		return nil, wrapRepoError("create journey", err)
	}
	if content == nil {
		content = defaultJourneyContent(createdTitle)
	}
	if err := r.upsertDocument(ctx, identity.UUID, "journey", id, revision, actorUserID, content); err != nil {
		return nil, err
	}

	return map[string]any{
		"id":              id,
		"title":           createdTitle,
		"linkedPersonas":  []string{},
		"stagesCount":     0,
		"painPointsCount": 0,
		"owner":           ownerName,
		"lastUpdated":     lastUpdated,
		"status":          status,
		"isOrphan":        isOrphan,
	}, nil
}

func (r *repo) GetJourney(ctx context.Context, projectID, slug string) (map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}

	var id, foundSlug, title, status, ownerName, createdAt, lastUpdated string
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`SELECT j.id::text, j.slug, j.title, j.status::text, COALESCE(u.name, ''), COALESCE(to_char(j.created_at, 'YYYY-MM-DD'), ''), COALESCE(to_char(j.updated_at, 'YYYY-MM-DD'), '')
		 FROM journeys j
		 JOIN users u ON u.id = j.owner_user_id
		 WHERE j.project_id = $1::uuid AND j.id::text = $2
		 LIMIT 1`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &foundSlug, &title, &status, &ownerName, &createdAt, &lastUpdated)
		},
		identity.UUID,
		strings.TrimSpace(slug),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "journey not found")
		}
		return nil, wrapRepoError("get journey", err)
	}

	content, err := r.loadLatestContent(ctx, identity.UUID, "journey", id, defaultJourneyContent(title))
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"journey": map[string]any{
			"id":          id,
			"title":       title,
			"status":      status,
			"owner":       ownerName,
			"lastUpdated": lastUpdated,
		},
		"metadata": map[string]any{
			"owner":        ownerName,
			"createdBy":    ownerName,
			"createdAt":    createdAt,
			"lastEditedBy": ownerName,
			"lastEditedAt": lastUpdated,
			"lastUpdated":  lastUpdated,
		},
		"detail":         content,
		"emotionOptions": []string{"Neutral", "Frustrated", "Anxious", "Relieved"},
		"reference":      map[string]any{"permissions": map[string]any{}},
	}, nil
}

func (r *repo) UpdateJourney(ctx context.Context, projectID, journeyID, actorUserID string, patch map[string]any) (map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}

	title := toString(patch["title"])
	status := toString(patch["status"])

	var id, slug, outTitle, outOwner, outLastUpdated, outStatus string
	var revision int
	var isOrphan bool
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`UPDATE journeys
		 SET title = COALESCE(NULLIF($3, ''), title),
		 	 status = CASE
		 	 	WHEN NULLIF($4, '')::journey_status = 'Archived'::journey_status
		 	 		AND journeys.status <> 'Archived'::journey_status
		 	 	THEN 'Archived'::journey_status
		 	 	WHEN NULLIF($4, '')::journey_status IS NOT NULL
		 	 		AND journeys.status = 'Archived'::journey_status
		 	 		AND NULLIF($4, '')::journey_status <> 'Archived'::journey_status
		 	 	THEN COALESCE(journeys.archived_from_status, 'Draft'::journey_status)
		 	 	ELSE COALESCE(NULLIF($4, '')::journey_status, journeys.status)
		 	 END,
		 	 archived_from_status = CASE
		 	 	WHEN NULLIF($4, '')::journey_status = 'Archived'::journey_status
		 	 		AND journeys.status <> 'Archived'::journey_status
		 	 	THEN journeys.status
		 	 	WHEN NULLIF($4, '')::journey_status IS NOT NULL
		 	 		AND journeys.status = 'Archived'::journey_status
		 	 		AND NULLIF($4, '')::journey_status <> 'Archived'::journey_status
		 	 	THEN NULL
		 	 	ELSE journeys.archived_from_status
		 	 END,
		 	 updated_at = NOW(),
		 	 document_revision = document_revision + 1
		 FROM users u
		 WHERE journeys.project_id = $1::uuid
		   AND journeys.id::text = $2
		   AND u.id = journeys.owner_user_id
		   AND (
		 	 journeys.status NOT IN ('Locked'::journey_status, 'Archived'::journey_status)
		 	 OR NULLIF($4, '')::journey_status = 'Archived'::journey_status
		 	 OR (
		 	 	journeys.status = 'Archived'::journey_status
		 	 	AND NULLIF($4, '')::journey_status IS NOT NULL
		 	 	AND NULLIF($4, '')::journey_status <> 'Archived'::journey_status
		 	 )
		   )
		 RETURNING journeys.id::text, journeys.slug, journeys.title, COALESCE(u.name, ''), COALESCE(to_char(journeys.updated_at, 'YYYY-MM-DD'), ''), journeys.status::text, journeys.document_revision, journeys.is_orphan`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &slug, &outTitle, &outOwner, &outLastUpdated, &outStatus, &revision, &isOrphan)
		},
		identity.UUID,
		strings.TrimSpace(journeyID),
		title,
		status,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "journey not found")
		}
		return nil, wrapRepoError("update journey", err)
	}

	if err := r.upsertDocument(ctx, identity.UUID, "journey", id, revision, actorUserID, patch); err != nil {
		return nil, err
	}

	return map[string]any{
		"id":              id,
		"title":           outTitle,
		"linkedPersonas":  []string{},
		"stagesCount":     0,
		"painPointsCount": 0,
		"owner":           outOwner,
		"lastUpdated":     outLastUpdated,
		"status":          outStatus,
		"isOrphan":        isOrphan,
	}, nil
}

func (r *repo) ListProblems(ctx context.Context, projectID string, query listQuery) ([]map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}

	items := make([]map[string]any, 0, query.Limit)
	args := []any{identity.UUID}
	sql := `
SELECT
	p.id::text,
	p.slug,
	p.title,
	COALESCE(u.name, ''),
	COALESCE(to_char(p.updated_at, 'YYYY-MM-DD'), ''),
	p.status::text,
	p.is_orphan,
	(SELECT COUNT(1) FROM ideas i WHERE i.project_id = p.project_id AND i.primary_problem_id = p.id)
FROM problems p
JOIN users u ON u.id = p.owner_user_id
WHERE p.project_id = $1::uuid
`
	if query.Status != "" {
		args = append(args, query.Status)
		sql += fmt.Sprintf(" AND p.status = $%d::problem_status\n", len(args))
	}
	args = append(args, query.Offset, query.Limit+1)
	sql += fmt.Sprintf("ORDER BY p.updated_at DESC OFFSET $%d LIMIT $%d", len(args)-1, len(args))

	err = r.store.Execute(ctx, storage.RelationalQueryMany(sql, func(row storage.RowScanner) error {
		var id, slug, statement, owner, lastUpdated, status string
		var isOrphan bool
		var ideasCount int
		if scanErr := row.Scan(&id, &slug, &statement, &owner, &lastUpdated, &status, &isOrphan, &ideasCount); scanErr != nil {
			return scanErr
		}
		items = append(items, map[string]any{
			"id":              id,
			"statement":       statement,
			"linkedSources":   []string{},
			"painPointsCount": 0,
			"ideasCount":      ideasCount,
			"status":          status,
			"owner":           owner,
			"lastUpdated":     lastUpdated,
			"isOrphan":        isOrphan,
		})
		return nil
	}, args...))
	if err != nil {
		return nil, wrapRepoError("list problems", err)
	}
	return items, nil
}

func (r *repo) CreateProblem(ctx context.Context, projectID, actorUserID, statement string, content map[string]any) (map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	slug, err := r.nextUniqueSlug(ctx, "problems", identity.UUID, statement)
	if err != nil {
		return nil, err
	}
	ownerName, err := r.resolveUserName(ctx, actorUserID)
	if err != nil {
		return nil, err
	}

	var id, createdSlug, outStatement, status, lastUpdated string
	var isOrphan bool
	var revision int
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`INSERT INTO problems (project_id, slug, title, owner_user_id, status, is_locked, is_orphan)
		 VALUES ($1::uuid, $2, $3, $4::uuid, 'Draft', FALSE, TRUE)
		 RETURNING id::text, slug, title, status::text, COALESCE(to_char(updated_at, 'YYYY-MM-DD'), ''), is_orphan, document_revision`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &createdSlug, &outStatement, &status, &lastUpdated, &isOrphan, &revision)
		},
		identity.UUID,
		slug,
		strings.TrimSpace(statement),
		strings.TrimSpace(actorUserID),
	))
	if err != nil {
		return nil, wrapRepoError("create problem", err)
	}
	if content == nil {
		content = defaultProblemContent(outStatement)
	}
	if err := r.upsertDocument(ctx, identity.UUID, "problem", id, revision, actorUserID, content); err != nil {
		return nil, err
	}

	return map[string]any{
		"id":              id,
		"statement":       outStatement,
		"linkedSources":   []string{},
		"painPointsCount": 0,
		"ideasCount":      0,
		"status":          status,
		"owner":           ownerName,
		"lastUpdated":     lastUpdated,
		"isOrphan":        isOrphan,
	}, nil
}

func (r *repo) GetProblem(ctx context.Context, projectID, slug string) (map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}

	var id, foundSlug, statement, status, ownerName, createdAt, lastUpdated string
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`SELECT p.id::text, p.slug, p.title, p.status::text, COALESCE(u.name, ''), COALESCE(to_char(p.created_at, 'YYYY-MM-DD'), ''), COALESCE(to_char(p.updated_at, 'YYYY-MM-DD'), '')
		 FROM problems p
		 JOIN users u ON u.id = p.owner_user_id
		 WHERE p.project_id = $1::uuid AND p.id::text = $2
		 LIMIT 1`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &foundSlug, &statement, &status, &ownerName, &createdAt, &lastUpdated)
		},
		identity.UUID,
		strings.TrimSpace(slug),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "problem not found")
		}
		return nil, wrapRepoError("get problem", err)
	}

	content, err := r.loadLatestContent(ctx, identity.UUID, "problem", id, defaultProblemContent(statement))
	if err != nil {
		return nil, err
	}

	// ---------- SAFE HELPERS ----------
	asSlice := func(v any) []any {
		switch val := v.(type) {
		case []any:
			return val
		case primitive.A:
			return []any(val)
		default:
			return []any{}
		}
	}

	asMap := func(v any) map[string]any {
		switch val := v.(type) {
		case map[string]any:
			return val
		case bson.M:
			return map[string]any(val)
		default:
			return nil
		}
	}

	// ---------- OPTIONS ----------
	storyOptions := make([]any, 0)
	stories, err := r.ListStories(ctx, identity.UUID, listQuery{Offset: 0, Limit: 100})
	if err != nil {
		return nil, err
	}
	for _, story := range stories {
		storyID := strings.TrimSpace(toString(story["id"]))
		storyTitle := strings.TrimSpace(toString(story["title"]))
		if storyID == "" || storyTitle == "" {
			continue
		}
		if !strings.EqualFold(strings.TrimSpace(toString(story["status"])), "Locked") {
			continue
		}
		storyOptions = append(storyOptions, map[string]any{
			"id":    storyID,
			"title": storyTitle,
			"href":  fmt.Sprintf("/project/%s/stories/%s", identity.UUID, storyID),
		})
	}

	journeyOptions := make([]any, 0)
	journeys, err := r.ListJourneys(ctx, identity.UUID, listQuery{Offset: 0, Limit: 100})
	if err != nil {
		return nil, err
	}
	for _, journey := range journeys {
		journeyID := strings.TrimSpace(toString(journey["id"]))
		journeyTitle := strings.TrimSpace(toString(journey["title"]))
		if journeyID == "" || journeyTitle == "" {
			continue
		}
		if !strings.EqualFold(strings.TrimSpace(toString(journey["status"])), "Locked") {
			continue
		}
		journeyOptions = append(journeyOptions, map[string]any{
			"id":    journeyID,
			"title": journeyTitle,
			"href":  fmt.Sprintf("/project/%s/journeys/%s", identity.UUID, journeyID),
		})
	}

	// ---------- RESOLUTION ----------
	linkedSources := asSlice(content["linkedSources"])

	linkedSourcesPayload := make([]any, 0)
	sourcePainPoints := make([]any, 0)
	sourcePersonas := make([]any, 0)
	sourcePainInsights := make([]any, 0)
	journeyPainInsights := make([]any, 0)
	sourceContextChunks := make([]string, 0)

	for _, raw := range linkedSources {
		source := asMap(raw)
		if source == nil {
			continue
		}

		sourceID := strings.TrimSpace(toString(source["id"]))
		if sourceID == "" {
			continue
		}

		sourceTypeRaw := firstNonEmpty(toString(source["artifactType"]), toString(source["type"]))
		normalizedType, ok := normalizeArtifactType(sourceTypeRaw)
		if !ok || (normalizedType != "story" && normalizedType != "journey") {
			continue
		}

		sourceTitle := strings.TrimSpace(toString(source["title"]))
		if sourceTitle == "" {
			if normalizedType == "story" {
				sourceTitle = "User Story"
			} else {
				sourceTitle = "User Journey"
			}
		}

		hrefPrefix := "stories"
		sourceLabelPrefix := "Story"
		typeLabel := "User Story"
		if normalizedType == "journey" {
			hrefPrefix = "journeys"
			sourceLabelPrefix = "Journey"
			typeLabel = "User Journey"
		}

		linkedSourcesPayload = append(linkedSourcesPayload, map[string]any{
			"id":    sourceID,
			"title": sourceTitle,
			"type":  normalizedType,
			"label": typeLabel,
			"href":  fmt.Sprintf("/project/%s/%s/%s", identity.UUID, hrefPrefix, sourceID),
		})

		resolvedID, err := r.resolveArtifactIDByIdentifier(ctx, identity.UUID, normalizedType, sourceID)
		if err != nil {
			continue
		}

		sourceContent, err := r.loadLatestContent(ctx, identity.UUID, normalizedType, resolvedID, map[string]any{})
		if err != nil {
			continue
		}

		persona := asMap(sourceContent["persona"])
		if persona != nil {
			name := strings.TrimSpace(toString(persona["name"]))
			if name != "" {
				sourcePersonas = append(sourcePersonas, map[string]any{
					"name":        name,
					"description": strings.TrimSpace(toString(persona["bio"])),
				})
			}
		}

		if normalizedType == "story" {
			for i, p := range asSlice(sourceContent["painPoints"]) {
				text := strings.TrimSpace(toString(p))
				if text == "" {
					continue
				}
				id := fmt.Sprintf("%s:story:%d", sourceID, i+1)
				sourcePainPoints = append(sourcePainPoints, map[string]any{"id": id, "text": text, "sourceLabel": sourceTitle})
				sourcePainInsights = append(sourcePainInsights, map[string]any{"id": id, "text": text, "sourceLabel": sourceTitle})
			}
		}

		if normalizedType == "journey" {
			for si, stageRaw := range asSlice(sourceContent["stages"]) {
				stage := asMap(stageRaw)
				if stage == nil {
					continue
				}
				stageName := strings.TrimSpace(toString(stage["name"]))
				if stageName == "" {
					stageName = "Stage"
				}

				for pi, p := range asSlice(stage["painPoints"]) {
					text := strings.TrimSpace(toString(p))
					if text == "" {
						continue
					}
					id := fmt.Sprintf("%s:journey:%d:%d", sourceID, si+1, pi+1)
					sourcePainPoints = append(sourcePainPoints, map[string]any{
						"id": id, "text": text,
						"sourceLabel": fmt.Sprintf("%s / %s", sourceTitle, stageName),
					})
					journeyPainInsights = append(journeyPainInsights, map[string]any{
						"id": id, "text": text,
						"journeyName": sourceTitle,
						"stageName":   stageName,
					})
				}
			}
		}

		ctxText := strings.TrimSpace(toString(sourceContent["context"]))
		if ctxText != "" {
			sourceContextChunks = append(sourceContextChunks, fmt.Sprintf("%s: %s", sourceLabelPrefix, ctxText))
		}
	}

	// ---------- RESPONSE ----------
	payload := cloneMap(content)

	// IMPORTANT: only override if we successfully processed
	if len(linkedSourcesPayload) > 0 {
		payload["linkedSources"] = linkedSourcesPayload
	}

	return map[string]any{
		"problem": map[string]any{
			"id":          id,
			"statement":   statement,
			"status":      status,
			"owner":       ownerName,
			"lastUpdated": lastUpdated,
		},
		"metadata": map[string]any{
			"owner":        ownerName,
			"createdBy":    ownerName,
			"createdAt":    createdAt,
			"lastEditedBy": ownerName,
			"lastEditedAt": lastUpdated,
			"lastUpdated":  lastUpdated,
		},
		"detail": payload,
		"reference": map[string]any{
			"storyOptions":     storyOptions,
			"journeyOptions":   journeyOptions,
			"sourcePainPoints": sourcePainPoints,
			"sourceInsights": map[string]any{
				"personas":          sourcePersonas,
				"context":           strings.Join(sourceContextChunks, "\n"),
				"painPoints":        sourcePainInsights,
				"journeyPainPoints": journeyPainInsights,
			},
			"permissions": map[string]any{},
		},
	}, nil
}

func (r *repo) UpdateProblem(ctx context.Context, projectID, problemID, actorUserID string, patch map[string]any) (map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	title := firstNonEmpty(toString(patch["finalStatement"]), toString(patch["title"]))
	status := toString(patch["status"])
	linkedSources, hasLinkedSources := patch["linkedSources"]

	var id, slug, outStatement, outOwner, outLastUpdated, outStatus string
	var revision int
	var isOrphan bool
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`UPDATE problems
		 SET title = COALESCE(NULLIF($3, ''), title),
		 	 status = CASE
		 	 	WHEN NULLIF($4, '')::problem_status = 'Archived'::problem_status
		 	 		AND problems.status <> 'Archived'::problem_status
		 	 	THEN 'Archived'::problem_status
		 	 	WHEN NULLIF($4, '')::problem_status IS NOT NULL
		 	 		AND problems.status = 'Archived'::problem_status
		 	 		AND NULLIF($4, '')::problem_status <> 'Archived'::problem_status
		 	 	THEN COALESCE(problems.archived_from_status, 'Draft'::problem_status)
		 	 	ELSE COALESCE(NULLIF($4, '')::problem_status, problems.status)
		 	 END,
		 	 archived_from_status = CASE
		 	 	WHEN NULLIF($4, '')::problem_status = 'Archived'::problem_status
		 	 		AND problems.status <> 'Archived'::problem_status
		 	 	THEN problems.status
		 	 	WHEN NULLIF($4, '')::problem_status IS NOT NULL
		 	 		AND problems.status = 'Archived'::problem_status
		 	 		AND NULLIF($4, '')::problem_status <> 'Archived'::problem_status
		 	 	THEN NULL
		 	 	ELSE problems.archived_from_status
		 	 END,
		 	 is_locked = CASE
		 	 	WHEN (
		 	 		CASE
		 	 			WHEN NULLIF($4, '')::problem_status = 'Archived'::problem_status
		 	 				AND problems.status <> 'Archived'::problem_status
		 	 			THEN 'Archived'::problem_status
		 	 			WHEN NULLIF($4, '')::problem_status IS NOT NULL
		 	 				AND problems.status = 'Archived'::problem_status
		 	 				AND NULLIF($4, '')::problem_status <> 'Archived'::problem_status
		 	 			THEN COALESCE(problems.archived_from_status, 'Draft'::problem_status)
		 	 			ELSE COALESCE(NULLIF($4, '')::problem_status, problems.status)
		 	 		END
		 	 	) = 'Locked'::problem_status
		 	 	THEN TRUE
		 	 	ELSE FALSE
		 	 END,
		 	 updated_at = NOW(),
		 	 document_revision = document_revision + 1
		 FROM users u
		 WHERE problems.project_id = $1::uuid
		   AND problems.id::text = $2
		   AND u.id = problems.owner_user_id
		   AND (
		 	 problems.status NOT IN ('Locked'::problem_status, 'Archived'::problem_status)
		 	 OR NULLIF($4, '')::problem_status = 'Archived'::problem_status
		 	 OR (
		 	 	problems.status = 'Archived'::problem_status
		 	 	AND NULLIF($4, '')::problem_status IS NOT NULL
		 	 	AND NULLIF($4, '')::problem_status <> 'Archived'::problem_status
		 	 )
		   )
		 RETURNING problems.id::text, problems.slug, problems.title, COALESCE(u.name, ''), COALESCE(to_char(problems.updated_at, 'YYYY-MM-DD'), ''), problems.status::text, problems.document_revision, problems.is_orphan`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &slug, &outStatement, &outOwner, &outLastUpdated, &outStatus, &revision, &isOrphan)
		},
		identity.UUID,
		strings.TrimSpace(problemID),
		title,
		status,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "problem not found")
		}
		return nil, wrapRepoError("update problem", err)
	}

	if hasLinkedSources {
		linkedSourceItems, ok := coerceAnySlice(linkedSources)
		if !ok {
			return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "linkedSources must be an array")
		}
		if err := r.replaceProblemSourceLinks(ctx, identity.UUID, id, actorUserID, linkedSourceItems); err != nil {
			return nil, err
		}
	}

	if err := r.upsertDocument(ctx, identity.UUID, "problem", id, revision, actorUserID, patch); err != nil {
		return nil, err
	}

	ideasCount, err := r.countIdeasForProblem(ctx, identity.UUID, id)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"id":              id,
		"statement":       outStatement,
		"linkedSources":   []string{},
		"painPointsCount": 0,
		"ideasCount":      ideasCount,
		"status":          outStatus,
		"owner":           outOwner,
		"lastUpdated":     outLastUpdated,
		"isOrphan":        isOrphan,
	}, nil
}

func (r *repo) ListIdeas(ctx context.Context, projectID string, query listQuery) ([]map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, query.Limit)
	args := []any{identity.UUID}
	sql := `
SELECT
	i.id::text,
	i.slug,
	i.title,
	COALESCE(p.title, ''),
	COALESCE(u.name, ''),
	COALESCE(to_char(i.updated_at, 'YYYY-MM-DD'), ''),
	i.status::text,
	i.is_orphan,
	COALESCE(p.status::text = 'Locked', FALSE),
	(SELECT COUNT(1) FROM tasks t WHERE t.project_id = i.project_id AND t.primary_idea_id = i.id)
FROM ideas i
JOIN users u ON u.id = i.owner_user_id
LEFT JOIN problems p ON p.id = i.primary_problem_id
WHERE i.project_id = $1::uuid
`
	if query.Status != "" {
		args = append(args, query.Status)
		sql += fmt.Sprintf(" AND i.status = $%d::idea_status\n", len(args))
	}
	args = append(args, query.Offset, query.Limit+1)
	sql += fmt.Sprintf("ORDER BY i.updated_at DESC OFFSET $%d LIMIT $%d", len(args)-1, len(args))
	err = r.store.Execute(ctx, storage.RelationalQueryMany(sql, func(row storage.RowScanner) error {
		var id, slug, title, problemStatement, owner, lastUpdated, status string
		var isOrphan, linkedProblemLocked bool
		var tasksCount int
		if scanErr := row.Scan(&id, &slug, &title, &problemStatement, &owner, &lastUpdated, &status, &isOrphan, &linkedProblemLocked, &tasksCount); scanErr != nil {
			return scanErr
		}
		items = append(items, map[string]any{
			"id":                     id,
			"title":                  title,
			"linkedProblemStatement": problemStatement,
			"persona":                "",
			"status":                 status,
			"tasksCount":             tasksCount,
			"owner":                  owner,
			"lastUpdated":            lastUpdated,
			"linkedProblemLocked":    linkedProblemLocked,
			"isOrphan":               isOrphan,
		})
		return nil
	}, args...))
	if err != nil {
		return nil, wrapRepoError("list ideas", err)
	}
	return items, nil
}

func (r *repo) CreateIdea(ctx context.Context, projectID, actorUserID, title string, content map[string]any) (map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	slug, err := r.nextUniqueSlug(ctx, "ideas", identity.UUID, title)
	if err != nil {
		return nil, err
	}
	ownerName, err := r.resolveUserName(ctx, actorUserID)
	if err != nil {
		return nil, err
	}
	var id, createdSlug, outTitle, status, lastUpdated string
	var isOrphan bool
	var revision int
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`INSERT INTO ideas (project_id, slug, title, owner_user_id, status, is_orphan)
		 VALUES ($1::uuid, $2, $3, $4::uuid, 'Considered', TRUE)
		 RETURNING id::text, slug, title, status::text, COALESCE(to_char(updated_at, 'YYYY-MM-DD'), ''), is_orphan, document_revision`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &createdSlug, &outTitle, &status, &lastUpdated, &isOrphan, &revision)
		},
		identity.UUID,
		slug,
		strings.TrimSpace(title),
		strings.TrimSpace(actorUserID),
	))
	if err != nil {
		return nil, wrapRepoError("create idea", err)
	}
	if content == nil {
		content = defaultIdeaContent(outTitle)
	}
	if err := r.upsertDocument(ctx, identity.UUID, "idea", id, revision, actorUserID, content); err != nil {
		return nil, err
	}
	return map[string]any{
		"id":                     id,
		"title":                  outTitle,
		"linkedProblemStatement": "",
		"persona":                "",
		"status":                 status,
		"tasksCount":             0,
		"owner":                  ownerName,
		"lastUpdated":            lastUpdated,
		"linkedProblemLocked":    false,
		"isOrphan":               isOrphan,
	}, nil
}

func (r *repo) GetIdea(ctx context.Context, projectID, slug string) (map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	var id, foundSlug, title, status, ownerName, lastUpdated string
	var primaryProblemID *string
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`SELECT i.id::text, i.slug, i.title, i.status::text, COALESCE(u.name, ''), COALESCE(to_char(i.updated_at, 'YYYY-MM-DD'), ''), i.primary_problem_id::text
		 FROM ideas i
		 JOIN users u ON u.id = i.owner_user_id
		 WHERE i.project_id = $1::uuid AND i.id::text = $2
		 LIMIT 1`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &foundSlug, &title, &status, &ownerName, &lastUpdated, &primaryProblemID)
		},
		identity.UUID,
		strings.TrimSpace(slug),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "idea not found")
		}
		return nil, wrapRepoError("get idea", err)
	}

	content, err := r.loadLatestContent(ctx, identity.UUID, "idea", id, defaultIdeaContent(title))
	if err != nil {
		return nil, err
	}
	problemOptions, err := r.listLockedProblemOptions(ctx, identity)
	if err != nil {
		return nil, err
	}
	if primaryProblemID != nil {
		content["selectedProblemId"] = *primaryProblemID
	}
	return map[string]any{
		"idea": map[string]any{
			"id":          id,
			"title":       title,
			"status":      status,
			"owner":       ownerName,
			"lastUpdated": lastUpdated,
		},
		"detail":    content,
		"reference": map[string]any{"problemOptions": problemOptions, "linkedStories": []any{}, "derivedPersonas": []any{}, "permissions": map[string]any{}},
	}, nil
}

func (r *repo) UpdateIdea(ctx context.Context, projectID, ideaID, actorUserID string, patch map[string]any) (map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	rawProblemRef, hasProblemRef := patch["selectedProblemId"]
	problemRef := toString(rawProblemRef)
	requestedStatus := toString(patch["status"])
	var primaryProblemID *string
	if hasProblemRef {
		if problemRef == "" {
			return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "selectedProblemId must be a non-empty string")
		}
		resolved, resolveErr := r.resolveArtifactIDByIdentifier(ctx, identity.UUID, "problem", problemRef)
		if resolveErr != nil {
			return nil, resolveErr
		}
		if lockErr := r.ensureProblemLocked(ctx, identity.UUID, resolved); lockErr != nil {
			return nil, lockErr
		}
		primaryProblemID = &resolved
	}

	var id, slug, title, owner, lastUpdated, outStatus, linkedProblemStatement string
	var linkedProblemLocked, isOrphan bool
	var tasksCount, revision int
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`UPDATE ideas i
		 SET primary_problem_id = COALESCE($3::uuid, i.primary_problem_id),
		 	 status = CASE
		 	 	WHEN NULLIF($4, '')::idea_status = 'Archived'::idea_status
		 	 		AND i.status <> 'Archived'::idea_status
		 	 	THEN 'Archived'::idea_status
		 	 	WHEN NULLIF($4, '')::idea_status IS NOT NULL
		 	 		AND i.status = 'Archived'::idea_status
		 	 		AND NULLIF($4, '')::idea_status <> 'Archived'::idea_status
		 	 	THEN COALESCE(i.archived_from_status, 'Considered'::idea_status)
		 	 	WHEN NULLIF($4, '')::idea_status IS NOT NULL
		 	 	THEN NULLIF($4, '')::idea_status
		 	 	ELSE i.status
		 	 END,
		 	 archived_from_status = CASE
		 	 	WHEN NULLIF($4, '')::idea_status = 'Archived'::idea_status
		 	 		AND i.status <> 'Archived'::idea_status
		 	 	THEN i.status
		 	 	WHEN NULLIF($4, '')::idea_status IS NOT NULL
		 	 		AND i.status = 'Archived'::idea_status
		 	 		AND NULLIF($4, '')::idea_status <> 'Archived'::idea_status
		 	 	THEN NULL
		 	 	ELSE i.archived_from_status
		 	 END,
		 	 selected_at = CASE
		 	 	WHEN (
		 	 		CASE
		 	 			WHEN NULLIF($4, '')::idea_status = 'Archived'::idea_status
		 	 				AND i.status <> 'Archived'::idea_status
		 	 			THEN 'Archived'::idea_status
		 	 			WHEN NULLIF($4, '')::idea_status IS NOT NULL
		 	 				AND i.status = 'Archived'::idea_status
		 	 				AND NULLIF($4, '')::idea_status <> 'Archived'::idea_status
		 	 			THEN COALESCE(i.archived_from_status, 'Considered'::idea_status)
		 	 			WHEN NULLIF($4, '')::idea_status IS NOT NULL
		 	 			THEN NULLIF($4, '')::idea_status
		 	 			ELSE i.status
		 	 		END
		 	 	) = 'Selected'::idea_status
		 	 	THEN COALESCE(i.selected_at, NOW())
		 	 	ELSE i.selected_at
		 	 END,
		 	 updated_at = NOW(),
		 	 document_revision = document_revision + 1
		 FROM users u
		 WHERE i.project_id = $1::uuid AND i.id::text = $2
		   AND u.id = i.owner_user_id
		   AND (
		 	 	(NULLIF($4, '')::idea_status IS NULL AND i.status NOT IN ('Selected'::idea_status, 'Rejected'::idea_status, 'Archived'::idea_status))
		 	 	OR NULLIF($4, '')::idea_status IS NOT NULL
		   )
		   AND (
		 	 	NULLIF($4, '')::idea_status <> 'Selected'::idea_status
		 	 	OR (
		 	 		COALESCE($3::uuid, i.primary_problem_id) IS NOT NULL
		 	 		AND EXISTS (
		 	 			SELECT 1
		 	 			FROM problems p
		 	 			WHERE p.id = COALESCE($3::uuid, i.primary_problem_id)
		 	 			  AND p.status = 'Locked'::problem_status
		 	 		)
		 	 	)
		   )
		 RETURNING i.id::text, i.slug, i.title, COALESCE(u.name, ''), COALESCE(to_char(i.updated_at, 'YYYY-MM-DD'), ''),
		 	 i.status::text,
		 	 COALESCE((SELECT p.title FROM problems p WHERE p.id = i.primary_problem_id), ''),
		 	 COALESCE((SELECT p.status = 'Locked'::problem_status FROM problems p WHERE p.id = i.primary_problem_id), FALSE),
		 	 i.is_orphan,
		 	 (SELECT COUNT(1) FROM tasks t WHERE t.project_id = i.project_id AND t.primary_idea_id = i.id), i.document_revision`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &slug, &title, &owner, &lastUpdated, &outStatus, &linkedProblemStatement, &linkedProblemLocked, &isOrphan, &tasksCount, &revision)
		},
		identity.UUID,
		strings.TrimSpace(ideaID),
		primaryProblemID,
		requestedStatus,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "idea not found")
		}
		return nil, wrapRepoError("update idea", err)
	}

	if primaryProblemID != nil {
		if linkErr := r.replacePrimaryLink(ctx, identity.UUID, "problem", *primaryProblemID, "idea", id, actorUserID); linkErr != nil {
			return nil, linkErr
		}
	}

	if err := r.upsertDocument(ctx, identity.UUID, "idea", id, revision, actorUserID, patch); err != nil {
		return nil, err
	}

	return map[string]any{
		"id":                     id,
		"title":                  title,
		"linkedProblemStatement": linkedProblemStatement,
		"persona":                "",
		"status":                 outStatus,
		"tasksCount":             tasksCount,
		"owner":                  owner,
		"lastUpdated":            lastUpdated,
		"linkedProblemLocked":    linkedProblemLocked,
		"isOrphan":               isOrphan,
	}, nil
}

func (r *repo) ListTasks(ctx context.Context, projectID string, query listQuery) ([]map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, query.Limit)
	args := []any{identity.UUID}
	sql := `
SELECT
	t.id::text,
	t.slug,
	t.title,
	COALESCE(i.title, ''),
	COALESCE(p.title, ''),
	COALESCE(u.name, ''),
	COALESCE(to_char(t.due_at, 'YYYY-MM-DD'), ''),
	COALESCE(to_char(t.updated_at, 'YYYY-MM-DD'), ''),
	t.status::text,
	t.is_orphan,
	COALESCE(i.status::text = 'Rejected', FALSE),
	EXISTS (SELECT 1 FROM feedback f WHERE f.project_id = t.project_id AND f.primary_task_id = t.id)
FROM tasks t
JOIN users u ON u.id = t.owner_user_id
LEFT JOIN ideas i ON i.id = t.primary_idea_id
LEFT JOIN problems p ON p.id = i.primary_problem_id
WHERE t.project_id = $1::uuid
`
	if query.Status != "" {
		args = append(args, query.Status)
		sql += fmt.Sprintf(" AND t.status = $%d::task_status\n", len(args))
	}
	args = append(args, query.Offset, query.Limit+1)
	sql += fmt.Sprintf("ORDER BY t.updated_at DESC OFFSET $%d LIMIT $%d", len(args)-1, len(args))
	err = r.store.Execute(ctx, storage.RelationalQueryMany(sql, func(row storage.RowScanner) error {
		var id, slug, title, linkedIdea, linkedProblemStatement, owner, deadline, lastUpdated, status string
		var isOrphan, ideaRejected, hasFeedback bool
		if scanErr := row.Scan(&id, &slug, &title, &linkedIdea, &linkedProblemStatement, &owner, &deadline, &lastUpdated, &status, &isOrphan, &ideaRejected, &hasFeedback); scanErr != nil {
			return scanErr
		}
		items = append(items, map[string]any{
			"id":                     id,
			"title":                  title,
			"linkedIdea":             linkedIdea,
			"linkedProblemStatement": linkedProblemStatement,
			"persona":                "",
			"owner":                  owner,
			"deadline":               deadline,
			"lastUpdated":            lastUpdated,
			"status":                 status,
			"ideaRejected":           ideaRejected,
			"hasFeedback":            hasFeedback,
			"isOrphan":               isOrphan,
		})
		return nil
	}, args...))
	if err != nil {
		return nil, wrapRepoError("list tasks", err)
	}
	return items, nil
}

func (r *repo) CreateTask(ctx context.Context, projectID, actorUserID, title string, content map[string]any) (map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	slug, err := r.nextUniqueSlug(ctx, "tasks", identity.UUID, title)
	if err != nil {
		return nil, err
	}
	ownerName, err := r.resolveUserName(ctx, actorUserID)
	if err != nil {
		return nil, err
	}
	var id, createdSlug, outTitle, status, deadline, lastUpdated string
	var isOrphan bool
	var revision int
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`INSERT INTO tasks (project_id, slug, title, owner_user_id, status, is_orphan)
		 VALUES ($1::uuid, $2, $3, $4::uuid, 'Planned', TRUE)
		 RETURNING id::text, slug, title, status::text, COALESCE(to_char(due_at, 'YYYY-MM-DD'), ''), COALESCE(to_char(updated_at, 'YYYY-MM-DD'), ''), is_orphan, document_revision`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &createdSlug, &outTitle, &status, &deadline, &lastUpdated, &isOrphan, &revision)
		},
		identity.UUID,
		slug,
		strings.TrimSpace(title),
		strings.TrimSpace(actorUserID),
	))
	if err != nil {
		return nil, wrapRepoError("create task", err)
	}
	ownerAssignees := []string{strings.TrimSpace(actorUserID)}
	if err := r.replaceTaskAssignees(ctx, identity.UUID, id, actorUserID, ownerAssignees); err != nil {
		return nil, err
	}
	if content == nil {
		content = defaultTaskContent(outTitle)
	}
	content["assignedToIds"] = stringSliceToAny(ownerAssignees)
	content["assignedToId"] = firstNonEmpty(ownerAssignees...)
	if err := r.upsertDocument(ctx, identity.UUID, "task", id, revision, actorUserID, content); err != nil {
		return nil, err
	}
	return map[string]any{
		"id":                     id,
		"title":                  outTitle,
		"linkedIdea":             "",
		"linkedProblemStatement": "",
		"persona":                "",
		"owner":                  ownerName,
		"deadline":               deadline,
		"lastUpdated":            lastUpdated,
		"status":                 status,
		"ideaRejected":           false,
		"hasFeedback":            false,
		"isOrphan":               isOrphan,
	}, nil
}

func (r *repo) GetTask(ctx context.Context, projectID, slug string) (map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	var id, foundSlug, title, status, ownerName, ownerUserID, deadline string
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`SELECT t.id::text, t.slug, t.title, t.status::text, COALESCE(u.name, ''), t.owner_user_id::text, COALESCE(to_char(t.due_at, 'YYYY-MM-DD'), '')
		 FROM tasks t
		 JOIN users u ON u.id = t.owner_user_id
		 WHERE t.project_id = $1::uuid AND t.id::text = $2
		 LIMIT 1`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &foundSlug, &title, &status, &ownerName, &ownerUserID, &deadline)
		},
		identity.UUID,
		strings.TrimSpace(slug),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "task not found")
		}
		return nil, wrapRepoError("get task", err)
	}
	content, err := r.loadLatestContent(ctx, identity.UUID, "task", id, defaultTaskContent(title))
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(deadline) != "" {
		content["deadline"] = deadline
	}
	assignedToIDs, err := r.listTaskAssigneeIDs(ctx, identity.UUID, id)
	if err != nil {
		return nil, err
	}
	if len(assignedToIDs) == 0 {
		fallbackOwner := strings.TrimSpace(ownerUserID)
		if fallbackOwner != "" {
			assignedToIDs = []string{fallbackOwner}
		}
	}
	content["assignedToIds"] = stringSliceToAny(assignedToIDs)
	if len(assignedToIDs) > 0 {
		content["assignedToId"] = assignedToIDs[0]
	} else {
		content["assignedToId"] = ""
	}

	assigneeOptions, err := r.listTaskAssigneeOptions(ctx, identity.UUID)
	if err != nil {
		return nil, err
	}
	ideaOptions, err := r.listTaskIdeaOptions(ctx, identity)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"task": map[string]any{
			"id":       id,
			"title":    title,
			"status":   status,
			"owner":    ownerName,
			"deadline": deadline,
		},
		"detail": content,
		"reference": map[string]any{
			"assigneeOptions": assigneeOptions,
			"ideaOptions":     ideaOptions,
			"permissions":     map[string]any{},
		},
	}, nil
}

func (r *repo) UpdateTask(ctx context.Context, projectID, taskID, actorUserID string, patch map[string]any) (map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	rawIdeaRef, hasIdeaRef := patch["selectedIdeaId"]
	ideaRef := toString(rawIdeaRef)
	deadlineRaw := toString(patch["deadline"])
	requestedStatus := toString(patch["status"])
	var primaryIdeaID *string
	if hasIdeaRef {
		if ideaRef == "" {
			return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "selectedIdeaId must be a non-empty string")
		}
		resolved, resolveErr := r.resolveArtifactIDByIdentifier(ctx, identity.UUID, "idea", ideaRef)
		if resolveErr != nil {
			return nil, resolveErr
		}
		primaryIdeaID = &resolved
		if selectedErr := r.ensureIdeaSelectedForTask(ctx, identity.UUID, resolved); selectedErr != nil {
			return nil, selectedErr
		}
	}

	assigneeIDs, hasAssigneePatch := extractTaskAssigneeIDs(patch)
	if hasAssigneePatch {
		resolvedAssignees, resolveErr := r.resolveActiveProjectMemberIDs(ctx, identity.UUID, assigneeIDs)
		if resolveErr != nil {
			return nil, resolveErr
		}
		assigneeIDs = resolvedAssignees
	}

	var dueAt any
	if deadlineRaw != "" {
		parsed, parseErr := time.Parse("2006-01-02", deadlineRaw)
		if parseErr != nil {
			return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "deadline must be ISO date")
		}
		dueAt = parsed
	}

	var id, slug, title, linkedIdea, linkedProblemStatement, owner, deadline, lastUpdated, outStatus string
	var isOrphan, ideaRejected, hasFeedback bool
	var revision int
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`UPDATE tasks t
		 SET primary_idea_id = COALESCE($3::uuid, t.primary_idea_id),
		 	 due_at = COALESCE($4::date, t.due_at),
		 	 status = COALESCE(NULLIF($5, '')::task_status, t.status),
		 	 started_at = CASE
		 	 	WHEN COALESCE(NULLIF($5, '')::task_status, t.status) = 'In Progress'::task_status AND t.started_at IS NULL
		 	 	THEN NOW()
		 	 	ELSE t.started_at
		 	 END,
		 	 completed_at = CASE
		 	 	WHEN COALESCE(NULLIF($5, '')::task_status, t.status) = 'Completed'::task_status
		 	 	THEN NOW()
		 	 	ELSE t.completed_at
		 	 END,
		 	 updated_at = NOW(),
		 	 document_revision = document_revision + 1
		 FROM users u
		 WHERE t.project_id = $1::uuid AND t.id::text = $2
		   AND u.id = t.owner_user_id
		   AND (
		 	 	(NULLIF($5, '')::task_status IS NULL AND t.status NOT IN ('Completed'::task_status, 'Abandoned'::task_status))
		 	 	OR NULLIF($5, '')::task_status IS NOT NULL
		   )
		   AND (
		 	 	COALESCE(NULLIF($5, '')::task_status, t.status) <> 'In Progress'::task_status
		 	 	OR NOT EXISTS (
		 	 		SELECT 1
		 	 		FROM ideas i
		 	 		WHERE i.id = COALESCE($3::uuid, t.primary_idea_id)
		 	 		  AND i.status = 'Rejected'::idea_status
		 	 	)
		   )
		 RETURNING t.id::text, t.slug, t.title,
		 	 COALESCE((SELECT i.title FROM ideas i WHERE i.id = t.primary_idea_id), ''),
		 	 COALESCE((SELECT p.title FROM ideas i JOIN problems p ON p.id = i.primary_problem_id WHERE i.id = t.primary_idea_id), ''),
		 	 COALESCE(u.name, ''),
		 	 COALESCE(to_char(t.due_at, 'YYYY-MM-DD'), ''), COALESCE(to_char(t.updated_at, 'YYYY-MM-DD'), ''),
		 	 t.status::text, t.is_orphan,
		 	 COALESCE((SELECT i.status = 'Rejected'::idea_status FROM ideas i WHERE i.id = t.primary_idea_id), FALSE),
		 	 EXISTS (SELECT 1 FROM feedback f WHERE f.project_id = t.project_id AND f.primary_task_id = t.id),
		 	 t.document_revision`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &slug, &title, &linkedIdea, &linkedProblemStatement, &owner, &deadline, &lastUpdated, &outStatus, &isOrphan, &ideaRejected, &hasFeedback, &revision)
		},
		identity.UUID,
		strings.TrimSpace(taskID),
		primaryIdeaID,
		dueAt,
		requestedStatus,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "task not found")
		}
		return nil, wrapRepoError("update task", err)
	}

	if primaryIdeaID != nil {
		if linkErr := r.replacePrimaryLink(ctx, identity.UUID, "idea", *primaryIdeaID, "task", id, actorUserID); linkErr != nil {
			return nil, linkErr
		}
	}
	if hasAssigneePatch {
		if assigneeErr := r.replaceTaskAssignees(ctx, identity.UUID, id, actorUserID, assigneeIDs); assigneeErr != nil {
			return nil, assigneeErr
		}
		patch["assignedToIds"] = stringSliceToAny(assigneeIDs)
		if len(assigneeIDs) > 0 {
			patch["assignedToId"] = assigneeIDs[0]
		} else {
			patch["assignedToId"] = ""
		}
	}
	if err := r.upsertDocument(ctx, identity.UUID, "task", id, revision, actorUserID, patch); err != nil {
		return nil, err
	}
	return map[string]any{
		"id":                     id,
		"title":                  title,
		"linkedIdea":             linkedIdea,
		"linkedProblemStatement": linkedProblemStatement,
		"persona":                "",
		"owner":                  owner,
		"deadline":               deadline,
		"lastUpdated":            lastUpdated,
		"status":                 outStatus,
		"ideaRejected":           ideaRejected,
		"hasFeedback":            hasFeedback,
		"isOrphan":               isOrphan,
	}, nil
}

func (r *repo) ListFeedback(ctx context.Context, projectID string, query listQuery) ([]map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, query.Limit)
	args := []any{identity.UUID}
	sql := `
SELECT
	f.id::text,
	f.slug,
	f.title,
	COALESCE(f.outcome::text, 'Needs Iteration'),
	f.status::text,
	COALESCE(u.name, ''),
	COALESCE(to_char(f.created_at, 'YYYY-MM-DD'), ''),
	f.is_orphan,
	f.primary_task_id IS NOT NULL,
	COALESCE(t.title, '')
FROM feedback f
JOIN users u ON u.id = f.owner_user_id
LEFT JOIN tasks t ON t.id = f.primary_task_id
WHERE f.project_id = $1::uuid
`
	if query.Status != "" {
		args = append(args, query.Status)
		sql += fmt.Sprintf(" AND f.status = $%d::feedback_status\n", len(args))
	}
	if query.Outcome != "" {
		args = append(args, query.Outcome)
		sql += fmt.Sprintf(" AND COALESCE(f.outcome::text, 'Needs Iteration') = $%d\n", len(args))
	}
	args = append(args, query.Offset, query.Limit+1)
	sql += fmt.Sprintf("ORDER BY f.updated_at DESC OFFSET $%d LIMIT $%d", len(args)-1, len(args))
	err = r.store.Execute(ctx, storage.RelationalQueryMany(sql, func(row storage.RowScanner) error {
		var id, slug, title, outcome, status, owner, createdDate, linkedTask string
		var isOrphan, hasTaskLink bool
		if scanErr := row.Scan(&id, &slug, &title, &outcome, &status, &owner, &createdDate, &isOrphan, &hasTaskLink, &linkedTask); scanErr != nil {
			return scanErr
		}
		linkedTaskOrIdea := ""
		if strings.TrimSpace(linkedTask) != "" {
			linkedTaskOrIdea = "Task: " + linkedTask
		}
		items = append(items, map[string]any{
			"id":               id,
			"title":            title,
			"linkedArtifacts":  []string{},
			"outcome":          outcome,
			"status":           status,
			"linkedTaskOrIdea": linkedTaskOrIdea,
			"owner":            owner,
			"createdDate":      createdDate,
			"hasTaskLink":      hasTaskLink,
			"isOrphan":         isOrphan,
		})
		return nil
	}, args...))
	if err != nil {
		return nil, wrapRepoError("list feedback", err)
	}
	return items, nil
}

func (r *repo) CreateFeedback(ctx context.Context, projectID, actorUserID, title string, content map[string]any) (map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	slug, err := r.nextUniqueSlug(ctx, "feedback", identity.UUID, title)
	if err != nil {
		return nil, err
	}
	ownerName, err := r.resolveUserName(ctx, actorUserID)
	if err != nil {
		return nil, err
	}
	var id, createdSlug, outTitle, outcome, status, createdDate string
	var hasTaskLink, isOrphan bool
	var revision int
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`INSERT INTO feedback (project_id, slug, title, owner_user_id, outcome, status, is_orphan)
		 VALUES ($1::uuid, $2, $3, $4::uuid, 'Needs Iteration', 'Active', TRUE)
		 RETURNING id::text, slug, title, COALESCE(outcome::text, 'Needs Iteration'), status::text,
		           COALESCE(to_char(created_at, 'YYYY-MM-DD'), ''), primary_task_id IS NOT NULL, is_orphan, document_revision`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &createdSlug, &outTitle, &outcome, &status, &createdDate, &hasTaskLink, &isOrphan, &revision)
		},
		identity.UUID,
		slug,
		strings.TrimSpace(title),
		strings.TrimSpace(actorUserID),
	))
	if err != nil {
		return nil, wrapRepoError("create feedback", err)
	}
	if content == nil {
		content = defaultFeedbackContent(outTitle)
	}
	if err := r.upsertDocument(ctx, identity.UUID, "feedback", id, revision, actorUserID, content); err != nil {
		return nil, err
	}
	return map[string]any{
		"id":               id,
		"title":            outTitle,
		"linkedArtifacts":  []string{},
		"outcome":          outcome,
		"status":           status,
		"linkedTaskOrIdea": "",
		"owner":            ownerName,
		"createdDate":      createdDate,
		"hasTaskLink":      hasTaskLink,
		"isOrphan":         isOrphan,
	}, nil
}

func (r *repo) GetFeedback(ctx context.Context, projectID, slug string) (map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	var id, foundSlug, title, outcome, status, ownerName, createdDate, updatedDate string
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`SELECT f.id::text, f.slug, f.title, COALESCE(f.outcome::text, 'Needs Iteration'), f.status::text,
		        COALESCE(u.name, ''), COALESCE(to_char(f.created_at, 'YYYY-MM-DD'), ''), COALESCE(to_char(f.updated_at, 'YYYY-MM-DD'), '')
		 FROM feedback f
		 JOIN users u ON u.id = f.owner_user_id
		 WHERE f.project_id = $1::uuid AND f.id::text = $2
		 LIMIT 1`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &foundSlug, &title, &outcome, &status, &ownerName, &createdDate, &updatedDate)
		},
		identity.UUID,
		strings.TrimSpace(slug),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "feedback not found")
		}
		return nil, wrapRepoError("get feedback", err)
	}
	content, err := r.loadLatestContent(ctx, identity.UUID, "feedback", id, defaultFeedbackContent(title))
	if err != nil {
		return nil, err
	}

	// ----------------------------
	// TASK OPTIONS
	// ----------------------------
	var taskOptions []map[string]any

	err = r.store.Execute(ctx, storage.RelationalQueryMany(
		`SELECT t.id::text, t.title, t.status::text
		 FROM tasks t
		 WHERE t.project_id = $1::uuid
		 AND t.status = 'Completed'::task_status`,
		func(row storage.RowScanner) error {
			var tid, ttitle, tstatus string
			if err := row.Scan(&tid, &ttitle, &tstatus); err != nil {
				return err
			}

			taskOptions = append(taskOptions, map[string]any{
				"id":     tid,
				"title":  ttitle,
				"type":   "task",
				"phase":  "Prototype",
				"href":   fmt.Sprintf("/project/%s/tasks/%s", projectID, tid),
				"status": tstatus,
			})
			return nil
		},
		identity.UUID,
	))
	if err != nil {
		return nil, err
	}

	// ----------------------------
	// IDEA OPTIONS
	// ----------------------------
	var ideaOptions []map[string]any

	err = r.store.Execute(ctx, storage.RelationalQueryMany(
		`SELECT i.id::text, i.title, i.status::text
		 FROM ideas i
		 WHERE i.project_id = $1::uuid
		AND i.status = 'Selected'::idea_status`,
		func(row storage.RowScanner) error {
			var iid, ititle, istatus string
			if err := row.Scan(&iid, &ititle, &istatus); err != nil {
				return err
			}

			ideaOptions = append(ideaOptions, map[string]any{
				"id":     iid,
				"title":  ititle,
				"type":   "idea",
				"phase":  "Ideate",
				"href":   fmt.Sprintf("/project/%s/ideas/%s", projectID, iid),
				"status": istatus,
			})
			return nil
		},
		identity.UUID,
	))
	if err != nil {
		return nil, err
	}

	// ----------------------------
	// PROBLEM OPTIONS
	// ----------------------------
	problemOptions := make([]map[string]any, 0)

	err = r.store.Execute(ctx, storage.RelationalQueryMany(
		`SELECT p.id::text, p.title, p.status::text
		 FROM problems p
		 WHERE p.project_id = $1::uuid
		 AND p.status = 'Locked'::problem_status`,
		func(row storage.RowScanner) error {
			var pid, ptitle, pstatus string
			if err := row.Scan(&pid, &ptitle, &pstatus); err != nil {
				return err
			}

			problemOptions = append(problemOptions, map[string]any{
				"id":     pid,
				"title":  ptitle,
				"type":   "problem",
				"phase":  "Define",
				"href":   fmt.Sprintf("/project/%s/problem-statement/%s", projectID, pid),
				"status": pstatus,
			})
			return nil
		},
		identity.UUID,
	))
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"feedback": map[string]any{
			"id":          id,
			"title":       title,
			"outcome":     outcome,
			"status":      status,
			"owner":       ownerName,
			"createdDate": createdDate,
			"lastUpdated": updatedDate,
		},
		"metadata": map[string]any{
			"owner":        ownerName,
			"createdBy":    ownerName,
			"createdAt":    createdDate,
			"lastEditedBy": ownerName,
			"lastEditedAt": updatedDate,
			"lastUpdated":  updatedDate,
		},
		"detail": content,
		"reference": map[string]any{
			"taskOptions":    taskOptions,
			"ideaOptions":    ideaOptions,
			"problemOptions": problemOptions,
			"permissions":    map[string]any{},
		},
	}, nil
}

func (r *repo) UpdateFeedback(ctx context.Context, projectID, feedbackID, actorUserID string, patch map[string]any) (map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	outcome := toString(patch["outcome"])
	status := toString(patch["status"])
	linkedArtifacts := toSlice(patch["linkedArtifacts"])

	primaryTaskID, linkIDs, err := r.resolveFeedbackLinks(ctx, identity.UUID, linkedArtifacts)
	if err != nil {
		return nil, err
	}

	var id, slug, title, outOutcome, outStatus, owner, createdDate string
	var hasTaskLink, isOrphan bool
	var revision int
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`UPDATE feedback f
		 SET outcome = COALESCE(NULLIF($3, '')::feedback_outcome, f.outcome),
		 	 status = COALESCE(NULLIF($4, '')::feedback_status, f.status),
		 	 archived_from_status = CASE
		 	 	WHEN COALESCE(NULLIF($4, ''), f.status::text) = 'Archived' AND f.status <> 'Archived'::feedback_status THEN f.status
		 	 	WHEN COALESCE(NULLIF($4, ''), f.status::text) <> 'Archived' AND f.status = 'Archived'::feedback_status THEN NULL
		 	 	ELSE f.archived_from_status
		 	 END,
		 	 primary_task_id = $5::uuid,
		 	 updated_at = NOW(),
		 	 document_revision = document_revision + 1
		 FROM users u
		 WHERE f.project_id = $1::uuid
		   AND f.id::text = $2
		   AND u.id = f.owner_user_id
		 RETURNING f.id::text, f.slug, f.title, COALESCE(f.outcome::text, 'Needs Iteration'), f.status::text, COALESCE(u.name, ''),
		 	 COALESCE(to_char(f.created_at, 'YYYY-MM-DD'), ''), f.primary_task_id IS NOT NULL, f.is_orphan, f.document_revision`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &slug, &title, &outOutcome, &outStatus, &owner, &createdDate, &hasTaskLink, &isOrphan, &revision)
		},
		identity.UUID,
		strings.TrimSpace(feedbackID),
		outcome,
		status,
		primaryTaskID,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "feedback not found")
		}
		return nil, wrapRepoError("update feedback", err)
	}

	if err := r.replaceFeedbackLinks(ctx, identity.UUID, id, actorUserID, linkIDs); err != nil {
		return nil, err
	}
	if err := r.upsertDocument(ctx, identity.UUID, "feedback", id, revision, actorUserID, patch); err != nil {
		return nil, err
	}

	linkedTaskOrIdea := ""
	if hasTaskLink {
		linkedTaskOrIdea = "Task"
	}

	return map[string]any{
		"id":               id,
		"title":            title,
		"linkedArtifacts":  []string{},
		"outcome":          outOutcome,
		"status":           outStatus,
		"linkedTaskOrIdea": linkedTaskOrIdea,
		"owner":            owner,
		"createdDate":      createdDate,
		"hasTaskLink":      hasTaskLink,
		"isOrphan":         isOrphan,
	}, nil
}

// shared helpers

func (r *repo) resolveUserName(ctx context.Context, userID string) (string, error) {
	var name string
	err := r.store.Execute(ctx, storage.RelationalQueryOne(
		`SELECT name FROM users WHERE id = $1::uuid LIMIT 1`,
		func(row storage.RowScanner) error { return row.Scan(&name) },
		strings.TrimSpace(userID),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperr.New(apperr.CodeNotFound, http.StatusNotFound, "user not found")
		}
		return "", wrapRepoError("resolve user", err)
	}
	return name, nil
}

func (r *repo) nextUniqueSlug(ctx context.Context, table, projectUUID, seed string) (string, error) {
	base := slugify(seed)
	slug := base
	for i := 1; i <= 200; i++ {
		exists, err := r.slugExists(ctx, table, projectUUID, slug)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i+1)
	}
	return "", apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "failed to allocate slug")
}

func (r *repo) slugExists(ctx context.Context, table, projectUUID, slug string) (bool, error) {
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE project_id = $1::uuid AND slug = $2)", table)
	err := r.store.Execute(ctx, storage.RelationalQueryOne(query, func(row storage.RowScanner) error {
		return row.Scan(&exists)
	}, projectUUID, slug))
	if err != nil {
		return false, wrapRepoError("check slug exists", err)
	}
	return exists, nil
}

func (r *repo) upsertDocument(ctx context.Context, projectUUID, artifactType, artifactID string, revision int, actorUserID string, content map[string]any) error {
	collection, err := documentCollectionByArtifactType(artifactType)
	if err != nil {
		return err
	}

	existingContent, err := r.loadLatestContent(ctx, projectUUID, artifactType, artifactID, map[string]any{})
	if err != nil {
		return err
	}
	mergedContent := patchx.MergeShallow(existingContent, content)

	documentID := fmt.Sprintf("%s:%s", artifactType, artifactID)
	doc := map[string]any{
		"artifact_id":        artifactID,
		"project_id":         projectUUID,
		"document_id":        documentID,
		"revision":           revision,
		"updated_at":         time.Now().UTC(),
		"updated_by_user_id": strings.TrimSpace(actorUserID),
		"schema_version":     1,
		"content":            mergedContent,
	}

	if err := r.docs.Execute(ctx, storage.DocumentRun(
		collection+":update_one",
		map[string]any{
			"filter":  map[string]any{"artifact_id": artifactID},
			"update":  map[string]any{"$set": doc},
			"options": options.Update().SetUpsert(true),
		},
		nil,
	)); err != nil {
		return wrapRepoError("upsert artifact document", err)
	}

	return nil
}

func (r *repo) loadLatestContent(ctx context.Context, projectUUID, artifactType, artifactID string, fallback map[string]any) (map[string]any, error) {
	content := cloneMap(fallback)
	collection, err := documentCollectionByArtifactType(artifactType)
	if err != nil {
		return nil, err
	}

	var doc map[string]any
	err = r.docs.Execute(ctx, storage.DocumentRun(
		collection+":find_one",
		map[string]any{
			"filter": map[string]any{
				"artifact_id": artifactID,
				"project_id":  projectUUID,
			},
		},
		&doc,
	))
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return content, nil
		}
		return nil, wrapRepoError("load latest content", err)
	}
	if payloadContent := toMap(doc["content"]); payloadContent != nil {
		for k, v := range payloadContent {
			content[k] = v
		}
	}
	return content, nil
}

func (r *repo) resolveArtifactIDByIdentifier(ctx context.Context, projectUUID, artifactType, identifier string) (string, error) {
	table, err := tableByArtifactType(artifactType)
	if err != nil {
		return "", err
	}
	query := fmt.Sprintf("SELECT id::text FROM %s WHERE project_id = $1::uuid AND id::text = $2 LIMIT 1", table)
	var id string
	err = r.store.Execute(ctx, storage.RelationalQueryOne(query, func(row storage.RowScanner) error {
		return row.Scan(&id)
	}, projectUUID, strings.TrimSpace(identifier)))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperr.New(apperr.CodeNotFound, http.StatusNotFound, artifactType+" not found")
		}
		return "", wrapRepoError("resolve artifact identifier", err)
	}
	return id, nil
}

func tableByArtifactType(artifactType string) (string, error) {
	switch artifactType {
	case "story":
		return "stories", nil
	case "journey":
		return "journeys", nil
	case "problem":
		return "problems", nil
	case "idea":
		return "ideas", nil
	case "task":
		return "tasks", nil
	case "feedback":
		return "feedback", nil
	default:
		return "", apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "unsupported artifact type")
	}
}

func documentCollectionByArtifactType(artifactType string) (string, error) {
	switch artifactType {
	case "story":
		return "story_documents", nil
	case "journey":
		return "journey_documents", nil
	case "problem":
		return "problem_documents", nil
	case "idea":
		return "idea_documents", nil
	case "task":
		return "task_documents", nil
	case "feedback":
		return "feedback_documents", nil
	default:
		return "", apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "unsupported artifact type")
	}
}

func (r *repo) replaceProblemSourceLinks(ctx context.Context, projectUUID, problemID, actorUserID string, linkedSources []any) error {
	if err := r.store.Execute(ctx, storage.RelationalExec(
		`DELETE FROM artifact_links
		 WHERE project_id = $1::uuid
		   AND target_type = 'problem'::artifact_type
		   AND target_id = $2::uuid
		   AND source_type IN ('story'::artifact_type, 'journey'::artifact_type)`,
		projectUUID,
		problemID,
	)); err != nil {
		return wrapRepoError("clear problem source links", err)
	}

	for _, raw := range linkedSources {
		item := toMap(raw)
		if item == nil {
			return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "linked source entry must be an object")
		}
		typeRaw := firstNonEmpty(toString(item["artifactType"]), toString(item["type"]))
		normalized, ok := normalizeArtifactType(typeRaw)
		if !ok || (normalized != "story" && normalized != "journey") {
			return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "linked source type must be story or journey")
		}
		identifier := toString(item["id"])
		if identifier == "" {
			return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "linked source id is required")
		}
		sourceID, err := r.resolveArtifactIDByIdentifier(ctx, projectUUID, normalized, identifier)
		if err != nil {
			return err
		}
		if lockErr := r.ensureProblemSourceLocked(ctx, projectUUID, normalized, sourceID); lockErr != nil {
			return lockErr
		}
		if err := r.store.Execute(ctx, storage.RelationalExec(
			`INSERT INTO artifact_links (project_id, source_type, source_id, target_type, target_id, link_kind, created_by_user_id)
			 VALUES ($1::uuid, $2::artifact_type, $3::uuid, 'problem'::artifact_type, $4::uuid, 'reference', $5::uuid)
			 ON CONFLICT (project_id, source_type, source_id, target_type, target_id, link_kind)
			 DO NOTHING`,
			projectUUID,
			normalized,
			sourceID,
			problemID,
			strings.TrimSpace(actorUserID),
		)); err != nil {
			return wrapRepoError("insert problem source link", err)
		}
	}

	return nil
}

func (r *repo) countIdeasForProblem(ctx context.Context, projectUUID, problemID string) (int, error) {
	var count int
	err := r.store.Execute(ctx, storage.RelationalQueryOne(
		`SELECT COUNT(1) FROM ideas WHERE project_id = $1::uuid AND primary_problem_id = $2::uuid`,
		func(row storage.RowScanner) error { return row.Scan(&count) },
		projectUUID,
		problemID,
	))
	if err != nil {
		return 0, wrapRepoError("count ideas", err)
	}
	return count, nil
}

func (r *repo) replacePrimaryLink(ctx context.Context, projectUUID, sourceType, sourceID, targetType, targetID, actorUserID string) error {
	if err := r.store.Execute(ctx, storage.RelationalExec(
		`DELETE FROM artifact_links
		 WHERE project_id = $1::uuid
		   AND target_type = $2::artifact_type
		   AND target_id = $3::uuid
		   AND source_type = $4::artifact_type`,
		projectUUID,
		targetType,
		targetID,
		sourceType,
	)); err != nil {
		return wrapRepoError("clear primary link", err)
	}

	if err := r.store.Execute(ctx, storage.RelationalExec(
		`INSERT INTO artifact_links (project_id, source_type, source_id, target_type, target_id, link_kind, created_by_user_id)
		 VALUES ($1::uuid, $2::artifact_type, $3::uuid, $4::artifact_type, $5::uuid, 'primary', $6::uuid)
		 ON CONFLICT (project_id, source_type, source_id, target_type, target_id, link_kind)
		 DO NOTHING`,
		projectUUID,
		sourceType,
		sourceID,
		targetType,
		targetID,
		strings.TrimSpace(actorUserID),
	)); err != nil {
		return wrapRepoError("insert primary link", err)
	}
	return nil
}

func (r *repo) resolveFeedbackLinks(ctx context.Context, projectUUID string, linkedArtifacts []any) (*string, []resolvedArtifactLink, error) {
	links := make([]resolvedArtifactLink, 0, len(linkedArtifacts))
	var primaryTaskID *string
	for _, raw := range linkedArtifacts {
		item := toMap(raw)
		if item == nil {
			continue
		}
		typeRaw := firstNonEmpty(toString(item["type"]), toString(item["artifactType"]))
		normalized, ok := normalizeArtifactType(typeRaw)
		if !ok || (normalized != "task" && normalized != "idea" && normalized != "problem") {
			continue
		}
		identifier := toString(item["id"])
		if identifier == "" {
			continue
		}
		resolvedID, err := r.resolveArtifactIDByIdentifier(ctx, projectUUID, normalized, identifier)
		if err != nil {
			return nil, nil, err
		}
		links = append(links, resolvedArtifactLink{ArtifactType: normalized, ArtifactID: resolvedID})
		if normalized == "task" && primaryTaskID == nil {
			id := resolvedID
			primaryTaskID = &id
		}
	}
	return primaryTaskID, links, nil
}

type resolvedArtifactLink struct {
	ArtifactType string
	ArtifactID   string
}

func (r *repo) replaceFeedbackLinks(ctx context.Context, projectUUID, feedbackID, actorUserID string, links []resolvedArtifactLink) error {
	if err := r.store.Execute(ctx, storage.RelationalExec(
		`DELETE FROM artifact_links
		 WHERE project_id = $1::uuid
		   AND target_type = 'feedback'::artifact_type
		   AND target_id = $2::uuid
		   AND source_type IN ('task'::artifact_type, 'idea'::artifact_type, 'problem'::artifact_type)`,
		projectUUID,
		feedbackID,
	)); err != nil {
		return wrapRepoError("clear feedback links", err)
	}

	for _, link := range links {
		if err := r.store.Execute(ctx, storage.RelationalExec(
			`INSERT INTO artifact_links (project_id, source_type, source_id, target_type, target_id, link_kind, created_by_user_id)
			 VALUES ($1::uuid, $2::artifact_type, $3::uuid, 'feedback'::artifact_type, $4::uuid, 'reference', $5::uuid)
			 ON CONFLICT (project_id, source_type, source_id, target_type, target_id, link_kind)
			 DO NOTHING`,
			projectUUID,
			link.ArtifactType,
			link.ArtifactID,
			feedbackID,
			strings.TrimSpace(actorUserID),
		)); err != nil {
			return wrapRepoError("insert feedback link", err)
		}
	}

	return nil
}

func stringSliceToAny(values []string) []any {
	trimmed := normalizeUniqueStrings(values)
	out := make([]any, 0, len(trimmed))
	for _, value := range trimmed {
		out = append(out, value)
	}
	return out
}

func extractTaskAssigneeIDs(patch map[string]any) ([]string, bool) {
	if patch == nil {
		return nil, false
	}
	if raw, ok := patch["assignedToIds"]; ok {
		return normalizeUniqueStrings(asStringSlice(raw)), true
	}
	if raw, ok := patch["assignedToId"]; ok {
		value := strings.TrimSpace(toString(raw))
		if value == "" {
			return []string{}, true
		}
		return []string{value}, true
	}
	return nil, false
}

func asStringSlice(value any) []string {
	if value == nil {
		return []string{}
	}
	switch typed := value.(type) {
	case []string:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			trimmed := strings.TrimSpace(item)
			if trimmed != "" {
				out = append(out, trimmed)
			}
		}
		return out
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			trimmed := strings.TrimSpace(toString(item))
			if trimmed != "" {
				out = append(out, trimmed)
			}
		}
		return out
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return []string{}
		}
		return []string{trimmed}
	default:
		return []string{}
	}
}

func coerceAnySlice(value any) ([]any, bool) {
	switch typed := value.(type) {
	case []any:
		return typed, true
	case []map[string]any:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, item)
		}
		return out, true
	default:
		return nil, false
	}
}

func normalizeUniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	unique := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		unique = append(unique, trimmed)
	}
	return unique
}

func (r *repo) listTaskAssigneeIDs(ctx context.Context, projectUUID, taskID string) ([]string, error) {
	ids := make([]string, 0, 4)
	err := r.store.Execute(ctx, storage.RelationalQueryMany(
		`SELECT ta.user_id::text
		 FROM task_assignees ta
		 WHERE ta.project_id = $1::uuid
		   AND ta.task_id = $2::uuid
		 ORDER BY ta.assigned_at ASC, ta.user_id ASC`,
		func(row storage.RowScanner) error {
			var userID string
			if scanErr := row.Scan(&userID); scanErr != nil {
				return scanErr
			}
			trimmed := strings.TrimSpace(userID)
			if trimmed != "" {
				ids = append(ids, trimmed)
			}
			return nil
		},
		projectUUID,
		taskID,
	))
	if err != nil {
		return nil, wrapRepoError("list task assignees", err)
	}
	return normalizeUniqueStrings(ids), nil
}

func (r *repo) listTaskAssigneeOptions(ctx context.Context, projectUUID string) ([]any, error) {
	options := make([]any, 0, 8)
	err := r.store.Execute(ctx, storage.RelationalQueryMany(
		`WITH assignable AS (
		 	SELECT pm.user_id, pm.role::text AS role
		 	FROM project_members pm
		 	WHERE pm.project_id = $1::uuid
		 	  AND pm.status = 'Active'::member_status
		 	UNION ALL
		 	SELECT p.owner_user_id, 'Owner'
		 	FROM projects p
		 	WHERE p.id = $1::uuid
		 ), dedup AS (
		 	SELECT
		 		a.user_id,
		 		CASE
		 			WHEN BOOL_OR(a.role = 'Owner') THEN 'Owner'
		 			ELSE MIN(a.role)
		 		END AS role
		 	FROM assignable a
		 	GROUP BY a.user_id
		 )
		 SELECT d.user_id::text, COALESCE(u.name, ''), d.role
		 FROM dedup d
		 JOIN users u ON u.id = d.user_id
		 ORDER BY CASE WHEN d.role = 'Owner' THEN 0 ELSE 1 END, u.name ASC`,
		func(row storage.RowScanner) error {
			var userID, name, role string
			if scanErr := row.Scan(&userID, &name, &role); scanErr != nil {
				return scanErr
			}
			trimmedID := strings.TrimSpace(userID)
			trimmedName := strings.TrimSpace(name)
			if trimmedID == "" || trimmedName == "" {
				return nil
			}
			trimmedRole := strings.TrimSpace(role)
			if trimmedRole == "" {
				trimmedRole = "Member"
			}
			options = append(options, map[string]any{
				"id":   trimmedID,
				"name": trimmedName,
				"role": trimmedRole,
			})
			return nil
		},
		projectUUID,
	))
	if err != nil {
		return nil, wrapRepoError("list task assignee options", err)
	}
	return options, nil
}

func (r *repo) resolveActiveProjectMemberIDs(ctx context.Context, projectUUID string, userIDs []string) ([]string, error) {
	uniqueIDs := normalizeUniqueStrings(userIDs)
	if len(uniqueIDs) == 0 {
		return []string{}, nil
	}

	active := make(map[string]struct{}, len(uniqueIDs))
	err := r.store.Execute(ctx, storage.RelationalQueryMany(
		`SELECT candidate.user_id::text
		 FROM (
		 	SELECT pm.user_id
		 	FROM project_members pm
		 	WHERE pm.project_id = $1::uuid
		 	  AND pm.status = 'Active'::member_status
		 	  AND pm.user_id::text = ANY($2::text[])
		 	UNION
		 	SELECT p.owner_user_id
		 	FROM projects p
		 	WHERE p.id = $1::uuid
		 	  AND p.owner_user_id::text = ANY($2::text[])
		 ) AS candidate`,
		func(row storage.RowScanner) error {
			var userID string
			if scanErr := row.Scan(&userID); scanErr != nil {
				return scanErr
			}
			active[strings.TrimSpace(userID)] = struct{}{}
			return nil
		},
		projectUUID,
		uniqueIDs,
	))
	if err != nil {
		return nil, wrapRepoError("resolve active project members", err)
	}

	for _, userID := range uniqueIDs {
		if _, ok := active[userID]; !ok {
			return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "assignee must be an active project member")
		}
	}

	return uniqueIDs, nil
}

func (r *repo) replaceTaskAssignees(ctx context.Context, projectUUID, taskID, actorUserID string, assigneeUserIDs []string) error {
	if err := r.store.Execute(ctx, storage.RelationalExec(
		`DELETE FROM task_assignees
		 WHERE project_id = $1::uuid
		   AND task_id = $2::uuid`,
		projectUUID,
		taskID,
	)); err != nil {
		return wrapRepoError("clear task assignees", err)
	}

	for _, userID := range normalizeUniqueStrings(assigneeUserIDs) {
		if err := r.store.Execute(ctx, storage.RelationalExec(
			`INSERT INTO task_assignees (task_id, project_id, user_id, assigned_by_user_id)
			 VALUES ($1::uuid, $2::uuid, $3::uuid, NULLIF($4, '')::uuid)
			 ON CONFLICT (task_id, user_id)
			 DO UPDATE SET assigned_at = NOW(), assigned_by_user_id = EXCLUDED.assigned_by_user_id`,
			taskID,
			projectUUID,
			userID,
			strings.TrimSpace(actorUserID),
		)); err != nil {
			return wrapRepoError("insert task assignee", err)
		}
	}

	return nil
}

func (r *repo) listLockedProblemOptions(ctx context.Context, identity projectIdentity) ([]any, error) {
	options := make([]any, 0, 16)
	err := r.store.Execute(ctx, storage.RelationalQueryMany(
		`SELECT p.id::text, p.slug, p.title, p.status::text
		 FROM problems p
		 WHERE p.project_id = $1::uuid
		   AND p.status = 'Locked'::problem_status
		 ORDER BY p.updated_at DESC
		 LIMIT 200`,
		func(row storage.RowScanner) error {
			var id, slug, title, status string
			if scanErr := row.Scan(&id, &slug, &title, &status); scanErr != nil {
				return scanErr
			}
			problemID := id
			if strings.TrimSpace(problemID) == "" || strings.TrimSpace(title) == "" {
				return nil
			}
			options = append(options, map[string]any{
				"id":     problemID,
				"title":  strings.TrimSpace(title),
				"phase":  "Define",
				"href":   fmt.Sprintf("/project/%s/problems/%s", identity.UUID, problemID),
				"status": strings.TrimSpace(status),
			})
			return nil
		},
		identity.UUID,
	))
	if err != nil {
		return nil, wrapRepoError("list locked problem options", err)
	}
	return options, nil
}

func (r *repo) ensureProblemLocked(ctx context.Context, projectUUID, problemID string) error {
	var status string
	err := r.store.Execute(ctx, storage.RelationalQueryOne(
		`SELECT status::text
		 FROM problems
		 WHERE project_id = $1::uuid
		   AND id = $2::uuid
		 LIMIT 1`,
		func(row storage.RowScanner) error { return row.Scan(&status) },
		projectUUID,
		problemID,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.New(apperr.CodeNotFound, http.StatusNotFound, "problem not found")
		}
		return wrapRepoError("check problem status", err)
	}
	if !strings.EqualFold(strings.TrimSpace(status), "Locked") {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "idea must link to a locked problem")
	}
	return nil
}

func (r *repo) ensureIdeaSelectedForTask(ctx context.Context, projectUUID, ideaID string) error {
	var status string
	err := r.store.Execute(ctx, storage.RelationalQueryOne(
		`SELECT status::text
		 FROM ideas
		 WHERE project_id = $1::uuid
		   AND id = $2::uuid
		 LIMIT 1`,
		func(row storage.RowScanner) error { return row.Scan(&status) },
		projectUUID,
		ideaID,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.New(apperr.CodeNotFound, http.StatusNotFound, "idea not found")
		}
		return wrapRepoError("check idea status", err)
	}
	if !strings.EqualFold(strings.TrimSpace(status), "Selected") {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "task must link to a selected idea")
	}
	return nil
}

func (r *repo) ensureProblemSourceLocked(ctx context.Context, projectUUID, artifactType, artifactID string) error {
	var table string
	switch strings.TrimSpace(artifactType) {
	case "story":
		table = "stories"
	case "journey":
		table = "journeys"
	default:
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "unsupported linked source type")
	}

	query := fmt.Sprintf(`SELECT status::text FROM %s WHERE project_id = $1::uuid AND id = $2::uuid LIMIT 1`, table)
	var status string
	err := r.store.Execute(ctx, storage.RelationalQueryOne(
		query,
		func(row storage.RowScanner) error { return row.Scan(&status) },
		projectUUID,
		artifactID,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.New(apperr.CodeNotFound, http.StatusNotFound, "linked source not found")
		}
		return wrapRepoError("check linked source status", err)
	}
	if !strings.EqualFold(strings.TrimSpace(status), "Locked") {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "problem sources must be locked")
	}
	return nil
}

func (r *repo) listTaskIdeaOptions(ctx context.Context, identity projectIdentity) ([]any, error) {
	options := make([]any, 0, 16)
	err := r.store.Execute(ctx, storage.RelationalQueryMany(
		`SELECT
			i.id::text,
			i.slug,
			i.title,
			p.id::text,
			p.slug,
			p.title,
			p.status::text,
			COALESCE(src.source_type, ''),
			COALESCE(src.source_id, ''),
			COALESCE(src.source_slug, ''),
			COALESCE(src.source_title, ''),
			COALESCE(src.source_status, '')
		 FROM ideas i
		 JOIN problems p ON p.id = i.primary_problem_id
		 LEFT JOIN LATERAL (
		 	SELECT
		 		l.source_type::text AS source_type,
		 		COALESCE(s.id::text, j.id::text) AS source_id,
		 		COALESCE(s.slug, j.slug) AS source_slug,
		 		COALESCE(s.title, j.title) AS source_title,
		 		COALESCE(s.status::text, j.status::text) AS source_status
		 	FROM artifact_links l
		 	LEFT JOIN stories s ON l.source_type = 'story'::artifact_type AND s.id = l.source_id
		 	LEFT JOIN journeys j ON l.source_type = 'journey'::artifact_type AND j.id = l.source_id
		 	WHERE l.project_id = i.project_id
		 	  AND l.target_type = 'problem'::artifact_type
		 	  AND l.target_id = p.id
		 	  AND l.source_type IN ('story'::artifact_type, 'journey'::artifact_type)
		 	ORDER BY l.created_at ASC
		 	LIMIT 1
		 ) src ON TRUE
		 WHERE i.project_id = $1::uuid
		   AND i.status = 'Selected'::idea_status
		 ORDER BY i.updated_at DESC
		 LIMIT 200`,
		func(row storage.RowScanner) error {
			var ideaID, ideaSlug, ideaTitle string
			var problemID, problemSlug, problemTitle, problemStatus string
			var sourceType, sourceID, sourceSlug, sourceTitle, sourceStatus string
			if scanErr := row.Scan(
				&ideaID,
				&ideaSlug,
				&ideaTitle,
				&problemID,
				&problemSlug,
				&problemTitle,
				&problemStatus,
				&sourceType,
				&sourceID,
				&sourceSlug,
				&sourceTitle,
				&sourceStatus,
			); scanErr != nil {
				return scanErr
			}

			ideaIdentifier := strings.TrimSpace(ideaID)
			problemIdentifier := strings.TrimSpace(problemID)
			if strings.TrimSpace(ideaIdentifier) == "" || strings.TrimSpace(problemIdentifier) == "" {
				return nil
			}

			contextType := "Persona"
			contextPath := "stories"
			if strings.EqualFold(strings.TrimSpace(sourceType), "journey") {
				contextType = "User Journey"
				contextPath = "journeys"
			}

			contextIdentifier := strings.TrimSpace(sourceID)
			contextTitle := strings.TrimSpace(sourceTitle)
			if contextTitle == "" {
				contextTitle = strings.TrimSpace(problemTitle)
			}

			contextHref := fmt.Sprintf("/project/%s/problems/%s", identity.UUID, problemIdentifier)
			if strings.TrimSpace(contextIdentifier) != "" {
				contextHref = fmt.Sprintf("/project/%s/%s/%s", identity.UUID, contextPath, contextIdentifier)
			}

			contextStatus := "Active"
			if strings.EqualFold(strings.TrimSpace(sourceStatus), "Archived") {
				contextStatus = "Archived"
			}

			options = append(options, map[string]any{
				"id":     ideaIdentifier,
				"title":  strings.TrimSpace(ideaTitle),
				"phase":  "Ideate",
				"href":   fmt.Sprintf("/project/%s/ideas/%s", identity.UUID, ideaIdentifier),
				"status": "Active",
				"problem": map[string]any{
					"id":     problemIdentifier,
					"title":  strings.TrimSpace(problemTitle),
					"phase":  "Define",
					"href":   fmt.Sprintf("/project/%s/problems/%s", identity.UUID, problemIdentifier),
					"status": strings.TrimSpace(problemStatus),
				},
				"context": map[string]any{
					"type":   contextType,
					"title":  contextTitle,
					"detail": "",
					"phase":  "Empathize",
					"href":   contextHref,
					"status": contextStatus,
				},
			})
			return nil
		},
		identity.UUID,
	))
	if err != nil {
		return nil, wrapRepoError("list task idea options", err)
	}
	return options, nil
}

func painCountOrSentinel(patch map[string]any) int {
	if _, ok := patch["painPoints"]; !ok {
		return -1
	}
	return countStringItems(toSlice(patch["painPoints"]))
}

func hypothesisCountOrSentinel(patch map[string]any) int {
	if _, ok := patch["hypothesis"]; !ok {
		return -1
	}
	return countStringItems(toSlice(patch["hypothesis"]))
}

func cloneMap(in map[string]any) map[string]any {
	if in == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func (r *repo) requireStore() error {
	if r == nil || r.store == nil || r.docs == nil {
		return apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "artifacts repository unavailable")
	}
	return nil
}

func wrapRepoError(action string, err error) error {
	if err == nil {
		return nil
	}
	if ae, ok := apperr.AsAppError(err); ok {
		return ae
	}
	return apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "failed to process artifacts data"), fmt.Errorf("%s: %w", action, err))
}

func defaultStoryContent(title string) map[string]any {
	return map[string]any{
		"title":       title,
		"description": "",
		"status":      "draft",
		"persona": map[string]any{
			"name": "", "bio": "", "role": "", "age": 0, "job": "", "edu": "",
		},
		"context":       "",
		"empathyMap":    map[string]any{"says": "", "thinks": "", "does": "", "feels": ""},
		"painPoints":    []any{},
		"hypothesis":    []any{},
		"addOnSections": []any{},
		"notes":         "",
	}
}

func storyAddOnCatalog() []any {
	return []any{
		map[string]any{
			"type":        "goals_success",
			"name":        "Goals & Success Criteria",
			"description": "Define what success looks like from the user's perspective.",
			"tag":         "Recommended",
		},
		map[string]any{
			"type":        "jtbd",
			"name":        "Jobs To Be Done (JTBD)",
			"description": "Capture functional, supporting, and emotional jobs.",
			"tag":         "Recommended",
		},
		map[string]any{
			"type":        "assumptions",
			"name":        "Assumptions",
			"description": "Make hidden assumptions explicit.",
			"tag":         "Optional",
		},
		map[string]any{
			"type":        "constraints",
			"name":        "Constraints",
			"description": "Capture environmental, technical, or behavioral limits.",
			"tag":         "Optional",
		},
		map[string]any{
			"type":        "risks_unknowns",
			"name":        "Risks & Unknowns",
			"description": "Identify uncertainty early.",
			"tag":         "Recommended",
		},
		map[string]any{
			"type":        "evidence",
			"name":        "Evidence / Research References",
			"description": "Ground the story in data and references.",
			"tag":         "Optional",
		},
		map[string]any{
			"type":        "scenarios",
			"name":        "Scenarios / Edge Cases",
			"description": "Capture non-happy paths and expectations.",
			"tag":         "Optional",
		},
	}
}

func defaultJourneyContent(title string) map[string]any {
	return map[string]any{
		"title":       title,
		"description": "",
		"status":      "draft",
		"persona": map[string]any{
			"name": "", "bio": "", "role": "", "age": 0, "job": "", "edu": "",
		},
		"context": "",
		"stages":  []any{},
		"notes":   "",
	}
}

func defaultProblemContent(statement string) map[string]any {
	return map[string]any{
		"title":              statement,
		"finalStatement":     "",
		"selectedPainPoints": []any{},
		"linkedSources":      []any{},
		"activeModules":      []any{},
		"moduleContent":      map[string]any{},
		"notes":              "",
	}
}

func defaultIdeaContent(title string) map[string]any {
	return map[string]any{
		"description":       "",
		"status":            "Considered",
		"summary":           "",
		"notes":             "",
		"selectedProblemId": "",
		"activeModules":     []any{},
		"moduleContent":     map[string]any{},
		"title":             title,
	}
}

func defaultTaskContent(title string) map[string]any {
	return map[string]any{
		"assignedToId":   "",
		"assignedToIds":  []any{},
		"selectedIdeaId": "",
		"deadline":       "",
		"hypothesis":     "",
		"planItems":      []any{},
		"executionLinks": []any{},
		"notes":          "",
		"activeModules":  []any{},
		"abandonReason":  "",
		"title":          title,
	}
}

func defaultFeedbackContent(title string) map[string]any {
	return map[string]any{
		"description":     "",
		"status":          "Active",
		"outcome":         "Needs Iteration",
		"linkedArtifacts": []any{},
		"activeModules":   []any{},
		"moduleContent":   map[string]any{},
		"notes":           "",
		"title":           title,
	}
}
