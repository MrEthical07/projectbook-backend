package artifacts

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
	LockProblem(ctx context.Context, projectID, problemID, actorUserID string) (map[string]any, error)
	UpdateProblemStatus(ctx context.Context, projectID, problemID, status, actorUserID string) (map[string]any, error)

	ListIdeas(ctx context.Context, projectID string, query listQuery) ([]map[string]any, error)
	CreateIdea(ctx context.Context, projectID, actorUserID, title string, content map[string]any) (map[string]any, error)
	GetIdea(ctx context.Context, projectID, slug string) (map[string]any, error)
	UpdateIdea(ctx context.Context, projectID, ideaID, actorUserID string, patch map[string]any) (map[string]any, error)
	SelectIdea(ctx context.Context, projectID, ideaID, actorUserID string) (map[string]any, error)
	UpdateIdeaStatus(ctx context.Context, projectID, ideaID, status, actorUserID string) (map[string]any, error)

	ListTasks(ctx context.Context, projectID string, query listQuery) ([]map[string]any, error)
	CreateTask(ctx context.Context, projectID, actorUserID, title string, content map[string]any) (map[string]any, error)
	GetTask(ctx context.Context, projectID, slug string) (map[string]any, error)
	UpdateTask(ctx context.Context, projectID, taskID, actorUserID string, patch map[string]any) (map[string]any, error)
	UpdateTaskStatus(ctx context.Context, projectID, taskID, status, actorUserID string) (map[string]any, error)

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
		`SELECT id::text, slug FROM projects WHERE slug = $1 OR id::text = $1 LIMIT 1`,
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
	args = append(args, query.Offset, query.Limit)
	sql += fmt.Sprintf("ORDER BY s.updated_at DESC OFFSET $%d LIMIT $%d", len(args)-1, len(args))

	err = r.store.Execute(ctx, storage.RelationalQueryMany(sql, func(row storage.RowScanner) error {
		var id, slug, title, personaName, owner, lastUpdated, status string
		var painCount, hypothesisCount int
		var isOrphan bool
		if scanErr := row.Scan(&id, &slug, &title, &personaName, &painCount, &hypothesisCount, &owner, &lastUpdated, &status, &isOrphan); scanErr != nil {
			return scanErr
		}
		items = append(items, map[string]any{
			"id":                     slugOrID(slug, id),
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
		"id":                     slugOrID(createdSlug, id),
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

	var id, foundSlug, title, status, ownerName, lastUpdated string
	var isOrphan bool
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`SELECT s.id::text, s.slug, s.title, s.status::text, COALESCE(u.name, ''), COALESCE(to_char(s.updated_at, 'YYYY-MM-DD'), ''), s.is_orphan
		 FROM stories s
		 JOIN users u ON u.id = s.owner_user_id
		 WHERE s.project_id = $1::uuid AND (s.slug = $2 OR s.id::text = $2)
		 LIMIT 1`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &foundSlug, &title, &status, &ownerName, &lastUpdated, &isOrphan)
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

	return map[string]any{
		"story": map[string]any{
			"id":          slugOrID(foundSlug, id),
			"title":       title,
			"status":      status,
			"owner":       ownerName,
			"lastUpdated": lastUpdated,
		},
		"detail":        content,
		"addOnCatalog":  []any{},
		"addOnSections": []any{},
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
		 	 status = COALESCE(NULLIF($4, '')::story_status, status),
		 	 persona_name = CASE WHEN $5 = '' THEN persona_name ELSE $5 END,
		 	 pain_points_count = CASE WHEN $6 < 0 THEN pain_points_count ELSE $6 END,
		 	 problem_hypotheses_count = CASE WHEN $7 < 0 THEN problem_hypotheses_count ELSE $7 END,
		 	 updated_at = NOW(),
		 	 document_revision = document_revision + 1
		 FROM users u
		 WHERE stories.project_id = $1::uuid
		   AND (stories.id::text = $2 OR stories.slug = $2)
		   AND u.id = stories.owner_user_id
		   AND (
		 	 stories.status NOT IN ('Locked'::story_status, 'Archived'::story_status)
		 	 OR NULLIF($4, '')::story_status = 'Archived'::story_status
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
		"id":                     slugOrID(slug, id),
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
	args = append(args, query.Offset, query.Limit)
	sql += fmt.Sprintf("ORDER BY j.updated_at DESC OFFSET $%d LIMIT $%d", len(args)-1, len(args))

	err = r.store.Execute(ctx, storage.RelationalQueryMany(sql, func(row storage.RowScanner) error {
		var id, slug, title, owner, lastUpdated, status string
		var isOrphan bool
		if scanErr := row.Scan(&id, &slug, &title, &owner, &lastUpdated, &status, &isOrphan); scanErr != nil {
			return scanErr
		}
		items = append(items, map[string]any{
			"id":              slugOrID(slug, id),
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
		"id":              slugOrID(createdSlug, id),
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

	var id, foundSlug, title, status, ownerName, lastUpdated string
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`SELECT j.id::text, j.slug, j.title, j.status::text, COALESCE(u.name, ''), COALESCE(to_char(j.updated_at, 'YYYY-MM-DD'), '')
		 FROM journeys j
		 JOIN users u ON u.id = j.owner_user_id
		 WHERE j.project_id = $1::uuid AND (j.slug = $2 OR j.id::text = $2)
		 LIMIT 1`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &foundSlug, &title, &status, &ownerName, &lastUpdated)
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
			"id":          slugOrID(foundSlug, id),
			"title":       title,
			"status":      status,
			"owner":       ownerName,
			"lastUpdated": lastUpdated,
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
		 	 status = COALESCE(NULLIF($4, '')::journey_status, status),
		 	 updated_at = NOW(),
		 	 document_revision = document_revision + 1
		 FROM users u
		 WHERE journeys.project_id = $1::uuid
		   AND (journeys.id::text = $2 OR journeys.slug = $2)
		   AND u.id = journeys.owner_user_id
		   AND (
		 	 journeys.status <> 'Archived'::journey_status
		 	 OR NULLIF($4, '')::journey_status = 'Archived'::journey_status
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
		"id":              slugOrID(slug, id),
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
	args = append(args, query.Offset, query.Limit)
	sql += fmt.Sprintf("ORDER BY p.updated_at DESC OFFSET $%d LIMIT $%d", len(args)-1, len(args))

	err = r.store.Execute(ctx, storage.RelationalQueryMany(sql, func(row storage.RowScanner) error {
		var id, slug, statement, owner, lastUpdated, status string
		var isOrphan bool
		var ideasCount int
		if scanErr := row.Scan(&id, &slug, &statement, &owner, &lastUpdated, &status, &isOrphan, &ideasCount); scanErr != nil {
			return scanErr
		}
		items = append(items, map[string]any{
			"id":              slugOrID(slug, id),
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
		"id":              slugOrID(createdSlug, id),
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
	var id, foundSlug, statement, status, ownerName, lastUpdated string
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`SELECT p.id::text, p.slug, p.title, p.status::text, COALESCE(u.name, ''), COALESCE(to_char(p.updated_at, 'YYYY-MM-DD'), '')
		 FROM problems p
		 JOIN users u ON u.id = p.owner_user_id
		 WHERE p.project_id = $1::uuid AND (p.slug = $2 OR p.id::text = $2)
		 LIMIT 1`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &foundSlug, &statement, &status, &ownerName, &lastUpdated)
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

	return map[string]any{
		"problem": map[string]any{
			"id":          slugOrID(foundSlug, id),
			"statement":   statement,
			"status":      status,
			"owner":       ownerName,
			"lastUpdated": lastUpdated,
		},
		"detail": content,
		"reference": map[string]any{
			"storyOptions":     []any{},
			"journeyOptions":   []any{},
			"sourcePainPoints": []any{},
			"permissions":      map[string]any{},
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

	var id, slug, outStatement, outOwner, outLastUpdated, outStatus string
	var revision int
	var isOrphan bool
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`UPDATE problems
		 SET title = COALESCE(NULLIF($3, ''), title),
		 	 status = COALESCE(NULLIF($4, '')::problem_status, status),
		 	 is_locked = CASE WHEN COALESCE(NULLIF($4, ''), status::text) = 'Locked' THEN TRUE ELSE is_locked END,
		 	 updated_at = NOW(),
		 	 document_revision = document_revision + 1
		 FROM users u
		 WHERE problems.project_id = $1::uuid
		   AND (problems.id::text = $2 OR problems.slug = $2)
		   AND u.id = problems.owner_user_id
		   AND (
		 	 problems.status NOT IN ('Locked'::problem_status, 'Archived'::problem_status)
		 	 OR NULLIF($4, '')::problem_status = 'Archived'::problem_status
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

	if err := r.replaceProblemSourceLinks(ctx, identity.UUID, id, actorUserID, toSlice(patch["linkedSources"])); err != nil {
		return nil, err
	}

	if err := r.upsertDocument(ctx, identity.UUID, "problem", id, revision, actorUserID, patch); err != nil {
		return nil, err
	}

	ideasCount, err := r.countIdeasForProblem(ctx, identity.UUID, id)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"id":              slugOrID(slug, id),
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

func (r *repo) LockProblem(ctx context.Context, projectID, problemID, actorUserID string) (map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	var id, status, lastUpdated string
	var revision int
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`UPDATE problems
		 SET status = 'Locked', is_locked = TRUE, updated_at = NOW(), document_revision = document_revision + 1
		 WHERE project_id = $1::uuid AND (id::text = $2 OR slug = $2)
		 RETURNING id::text, status::text, COALESCE(to_char(updated_at, 'YYYY-MM-DD'), ''), document_revision`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &status, &lastUpdated, &revision)
		},
		identity.UUID,
		strings.TrimSpace(problemID),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "problem not found")
		}
		return nil, wrapRepoError("lock problem", err)
	}
	if err := r.upsertDocument(ctx, identity.UUID, "problem", id, revision, actorUserID, map[string]any{"status": "Locked"}); err != nil {
		return nil, err
	}
	return map[string]any{"id": id, "status": status, "lastUpdated": lastUpdated}, nil
}

func (r *repo) UpdateProblemStatus(ctx context.Context, projectID, problemID, status, actorUserID string) (map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	var id, outStatus, lastUpdated string
	var revision int
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`UPDATE problems
		 SET status = $3::problem_status,
		 	 is_locked = CASE WHEN $3::problem_status = 'Locked' THEN TRUE ELSE FALSE END,
		 	 updated_at = NOW(),
		 	 document_revision = document_revision + 1
		 WHERE project_id = $1::uuid
		   AND (id::text = $2 OR slug = $2)
		   AND (
		 	 status NOT IN ('Locked'::problem_status, 'Archived'::problem_status)
		 	 OR $3::problem_status = 'Archived'::problem_status
		   )
		 RETURNING id::text, status::text, COALESCE(to_char(updated_at, 'YYYY-MM-DD'), ''), document_revision`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &outStatus, &lastUpdated, &revision)
		},
		identity.UUID,
		strings.TrimSpace(problemID),
		status,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "problem not found")
		}
		return nil, wrapRepoError("update problem status", err)
	}
	if err := r.upsertDocument(ctx, identity.UUID, "problem", id, revision, actorUserID, map[string]any{"status": outStatus}); err != nil {
		return nil, err
	}
	return map[string]any{"id": id, "status": outStatus, "lastUpdated": lastUpdated}, nil
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
	args = append(args, query.Offset, query.Limit)
	sql += fmt.Sprintf("ORDER BY i.updated_at DESC OFFSET $%d LIMIT $%d", len(args)-1, len(args))
	err = r.store.Execute(ctx, storage.RelationalQueryMany(sql, func(row storage.RowScanner) error {
		var id, slug, title, problemStatement, owner, lastUpdated, status string
		var isOrphan, linkedProblemLocked bool
		var tasksCount int
		if scanErr := row.Scan(&id, &slug, &title, &problemStatement, &owner, &lastUpdated, &status, &isOrphan, &linkedProblemLocked, &tasksCount); scanErr != nil {
			return scanErr
		}
		items = append(items, map[string]any{
			"id":                     slugOrID(slug, id),
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
		"id":                     slugOrID(createdSlug, id),
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
		 WHERE i.project_id = $1::uuid AND (i.slug = $2 OR i.id::text = $2)
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
	if primaryProblemID != nil {
		content["selectedProblemId"] = *primaryProblemID
	}
	return map[string]any{
		"idea": map[string]any{
			"id":          slugOrID(foundSlug, id),
			"title":       title,
			"status":      status,
			"owner":       ownerName,
			"lastUpdated": lastUpdated,
		},
		"detail":    content,
		"reference": map[string]any{"problemOptions": []any{}, "linkedStories": []any{}, "derivedPersonas": []any{}, "permissions": map[string]any{}},
	}, nil
}

func (r *repo) UpdateIdea(ctx context.Context, projectID, ideaID, actorUserID string, patch map[string]any) (map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	problemRef := toString(patch["selectedProblemId"])
	var primaryProblemID *string
	if problemRef != "" {
		resolved, resolveErr := r.resolveArtifactIDByIdentifier(ctx, identity.UUID, "problem", problemRef)
		if resolveErr != nil {
			return nil, resolveErr
		}
		primaryProblemID = &resolved
	}

	var id, slug, title, owner, lastUpdated, status, linkedProblemStatement string
	var linkedProblemLocked, isOrphan bool
	var tasksCount, revision int
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`UPDATE ideas i
		 SET primary_problem_id = COALESCE($3::uuid, i.primary_problem_id),
		 	 updated_at = NOW(),
		 	 document_revision = document_revision + 1
		 FROM users u
		 WHERE i.project_id = $1::uuid AND (i.id::text = $2 OR i.slug = $2)
		   AND u.id = i.owner_user_id
		   AND i.status NOT IN ('Selected'::idea_status, 'Rejected'::idea_status, 'Archived'::idea_status)
		 RETURNING i.id::text, i.slug, i.title, COALESCE(u.name, ''), COALESCE(to_char(i.updated_at, 'YYYY-MM-DD'), ''),
		 	 i.status::text,
		 	 COALESCE((SELECT p.title FROM problems p WHERE p.id = i.primary_problem_id), ''),
		 	 COALESCE((SELECT p.status = 'Locked'::problem_status FROM problems p WHERE p.id = i.primary_problem_id), FALSE),
		 	 i.is_orphan,
		 	 (SELECT COUNT(1) FROM tasks t WHERE t.project_id = i.project_id AND t.primary_idea_id = i.id), i.document_revision`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &slug, &title, &owner, &lastUpdated, &status, &linkedProblemStatement, &linkedProblemLocked, &isOrphan, &tasksCount, &revision)
		},
		identity.UUID,
		strings.TrimSpace(ideaID),
		primaryProblemID,
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
		"id":                     slugOrID(slug, id),
		"title":                  title,
		"linkedProblemStatement": linkedProblemStatement,
		"persona":                "",
		"status":                 status,
		"tasksCount":             tasksCount,
		"owner":                  owner,
		"lastUpdated":            lastUpdated,
		"linkedProblemLocked":    linkedProblemLocked,
		"isOrphan":               isOrphan,
	}, nil
}

func (r *repo) SelectIdea(ctx context.Context, projectID, ideaID, actorUserID string) (map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	var id, status, lastUpdated string
	var revision int
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`UPDATE ideas i
		 SET status = 'Selected', selected_at = NOW(), updated_at = NOW(), document_revision = document_revision + 1
		 WHERE i.project_id = $1::uuid
		   AND (i.id::text = $2 OR i.slug = $2)
		   AND i.primary_problem_id IS NOT NULL
		   AND EXISTS (SELECT 1 FROM problems p WHERE p.id = i.primary_problem_id AND p.status = 'Locked')
		 RETURNING i.id::text, i.status::text, COALESCE(to_char(i.updated_at, 'YYYY-MM-DD'), ''), i.document_revision`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &status, &lastUpdated, &revision)
		},
		identity.UUID,
		strings.TrimSpace(ideaID),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "linked locked problem is required")
		}
		return nil, wrapRepoError("select idea", err)
	}
	if err := r.upsertDocument(ctx, identity.UUID, "idea", id, revision, actorUserID, map[string]any{"status": "Selected"}); err != nil {
		return nil, err
	}
	return map[string]any{"id": id, "status": status, "lastUpdated": lastUpdated}, nil
}

func (r *repo) UpdateIdeaStatus(ctx context.Context, projectID, ideaID, status, actorUserID string) (map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	var id, outStatus, lastUpdated string
	var revision int
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`UPDATE ideas
		 SET status = $3::idea_status, updated_at = NOW(), document_revision = document_revision + 1
		 WHERE project_id = $1::uuid
		   AND (id::text = $2 OR slug = $2)
		   AND (
		 	 status NOT IN ('Selected'::idea_status, 'Rejected'::idea_status, 'Archived'::idea_status)
		 	 OR $3::idea_status = 'Archived'::idea_status
		   )
		 RETURNING id::text, status::text, COALESCE(to_char(updated_at, 'YYYY-MM-DD'), ''), document_revision`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &outStatus, &lastUpdated, &revision)
		},
		identity.UUID,
		strings.TrimSpace(ideaID),
		status,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "idea not found")
		}
		return nil, wrapRepoError("update idea status", err)
	}
	if err := r.upsertDocument(ctx, identity.UUID, "idea", id, revision, actorUserID, map[string]any{"status": outStatus}); err != nil {
		return nil, err
	}
	return map[string]any{"id": id, "status": outStatus, "lastUpdated": lastUpdated}, nil
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
	args = append(args, query.Offset, query.Limit)
	sql += fmt.Sprintf("ORDER BY t.updated_at DESC OFFSET $%d LIMIT $%d", len(args)-1, len(args))
	err = r.store.Execute(ctx, storage.RelationalQueryMany(sql, func(row storage.RowScanner) error {
		var id, slug, title, linkedIdea, linkedProblemStatement, owner, deadline, lastUpdated, status string
		var isOrphan, ideaRejected, hasFeedback bool
		if scanErr := row.Scan(&id, &slug, &title, &linkedIdea, &linkedProblemStatement, &owner, &deadline, &lastUpdated, &status, &isOrphan, &ideaRejected, &hasFeedback); scanErr != nil {
			return scanErr
		}
		items = append(items, map[string]any{
			"id":                     slugOrID(slug, id),
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
	if content == nil {
		content = defaultTaskContent(outTitle)
	}
	if err := r.upsertDocument(ctx, identity.UUID, "task", id, revision, actorUserID, content); err != nil {
		return nil, err
	}
	return map[string]any{
		"id":                     slugOrID(createdSlug, id),
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
	var id, foundSlug, title, status, ownerName, deadline string
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`SELECT t.id::text, t.slug, t.title, t.status::text, COALESCE(u.name, ''), COALESCE(to_char(t.due_at, 'YYYY-MM-DD'), '')
		 FROM tasks t
		 JOIN users u ON u.id = t.owner_user_id
		 WHERE t.project_id = $1::uuid AND (t.slug = $2 OR t.id::text = $2)
		 LIMIT 1`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &foundSlug, &title, &status, &ownerName, &deadline)
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
	return map[string]any{
		"task": map[string]any{
			"id":       slugOrID(foundSlug, id),
			"title":    title,
			"status":   status,
			"owner":    ownerName,
			"deadline": deadline,
		},
		"detail":    content,
		"reference": map[string]any{"assigneeOptions": []any{}, "ideaOptions": []any{}, "permissions": map[string]any{}},
	}, nil
}

func (r *repo) UpdateTask(ctx context.Context, projectID, taskID, actorUserID string, patch map[string]any) (map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	ideaRef := toString(patch["selectedIdeaId"])
	deadlineRaw := toString(patch["deadline"])
	var primaryIdeaID *string
	if ideaRef != "" {
		resolved, resolveErr := r.resolveArtifactIDByIdentifier(ctx, identity.UUID, "idea", ideaRef)
		if resolveErr != nil {
			return nil, resolveErr
		}
		primaryIdeaID = &resolved
	}
	var dueAt any
	if deadlineRaw != "" {
		parsed, parseErr := time.Parse("2006-01-02", deadlineRaw)
		if parseErr != nil {
			return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "deadline must be ISO date")
		}
		dueAt = parsed
	}

	var id, slug, title, linkedIdea, linkedProblemStatement, owner, deadline, lastUpdated, status string
	var isOrphan, ideaRejected, hasFeedback bool
	var revision int
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`UPDATE tasks t
		 SET primary_idea_id = COALESCE($3::uuid, t.primary_idea_id),
		 	 due_at = COALESCE($4::date, t.due_at),
		 	 updated_at = NOW(),
		 	 document_revision = document_revision + 1
		 FROM users u
		 WHERE t.project_id = $1::uuid AND (t.id::text = $2 OR t.slug = $2)
		   AND u.id = t.owner_user_id
		   AND t.status NOT IN ('Completed'::task_status, 'Abandoned'::task_status)
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
			return row.Scan(&id, &slug, &title, &linkedIdea, &linkedProblemStatement, &owner, &deadline, &lastUpdated, &status, &isOrphan, &ideaRejected, &hasFeedback, &revision)
		},
		identity.UUID,
		strings.TrimSpace(taskID),
		primaryIdeaID,
		dueAt,
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
	if err := r.upsertDocument(ctx, identity.UUID, "task", id, revision, actorUserID, patch); err != nil {
		return nil, err
	}
	return map[string]any{
		"id":                     slugOrID(slug, id),
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
	}, nil
}

func (r *repo) UpdateTaskStatus(ctx context.Context, projectID, taskID, status, actorUserID string) (map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	var id, outStatus, lastUpdated string
	var revision int
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`UPDATE tasks
		 SET status = $3::task_status,
		 	 started_at = CASE WHEN $3::task_status = 'In Progress' AND started_at IS NULL THEN NOW() ELSE started_at END,
		 	 completed_at = CASE WHEN $3::task_status = 'Completed' THEN NOW() ELSE completed_at END,
		 	 updated_at = NOW(),
		 	 document_revision = document_revision + 1
		 WHERE project_id = $1::uuid
		   AND (id::text = $2 OR slug = $2)
		   AND status NOT IN ('Completed'::task_status, 'Abandoned'::task_status)
		   AND (
			 $3::task_status <> 'In Progress'
			 OR NOT EXISTS (
			 	SELECT 1
			 	FROM ideas i
			 	WHERE i.id = tasks.primary_idea_id
			 	  AND i.status = 'Rejected'::idea_status
			 )
		   )
		 RETURNING id::text, status::text, COALESCE(to_char(updated_at, 'YYYY-MM-DD'), ''), document_revision`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &outStatus, &lastUpdated, &revision)
		},
		identity.UUID,
		strings.TrimSpace(taskID),
		status,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			if strings.EqualFold(strings.TrimSpace(status), "In Progress") {
				blocked, checkErr := r.taskBlockedForStart(ctx, identity.UUID, taskID)
				if checkErr != nil {
					return nil, checkErr
				}
				if blocked {
					return nil, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "cannot move task to In Progress when linked idea is Rejected")
				}
			}
			return nil, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "task not found")
		}
		return nil, wrapRepoError("update task status", err)
	}
	if err := r.upsertDocument(ctx, identity.UUID, "task", id, revision, actorUserID, map[string]any{"status": outStatus}); err != nil {
		return nil, err
	}
	return map[string]any{"id": id, "status": outStatus, "lastUpdated": lastUpdated}, nil
}

func (r *repo) taskBlockedForStart(ctx context.Context, projectUUID, taskID string) (bool, error) {
	var blocked bool
	err := r.store.Execute(ctx, storage.RelationalQueryOne(
		`SELECT EXISTS(
			SELECT 1
			FROM tasks t
			JOIN ideas i ON i.id = t.primary_idea_id
			WHERE t.project_id = $1::uuid
			  AND (t.id::text = $2 OR t.slug = $2)
			  AND i.status = 'Rejected'::idea_status
		)`,
		func(row storage.RowScanner) error { return row.Scan(&blocked) },
		projectUUID,
		strings.TrimSpace(taskID),
	))
	if err != nil {
		return false, wrapRepoError("check task start eligibility", err)
	}
	return blocked, nil
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
	if query.Outcome != "" {
		args = append(args, query.Outcome)
		sql += fmt.Sprintf(" AND COALESCE(f.outcome::text, 'Needs Iteration') = $%d\n", len(args))
	}
	args = append(args, query.Offset, query.Limit)
	sql += fmt.Sprintf("ORDER BY f.updated_at DESC OFFSET $%d LIMIT $%d", len(args)-1, len(args))
	err = r.store.Execute(ctx, storage.RelationalQueryMany(sql, func(row storage.RowScanner) error {
		var id, slug, title, outcome, owner, createdDate, linkedTask string
		var isOrphan, hasTaskLink bool
		if scanErr := row.Scan(&id, &slug, &title, &outcome, &owner, &createdDate, &isOrphan, &hasTaskLink, &linkedTask); scanErr != nil {
			return scanErr
		}
		linkedTaskOrIdea := ""
		if strings.TrimSpace(linkedTask) != "" {
			linkedTaskOrIdea = "Task: " + linkedTask
		}
		items = append(items, map[string]any{
			"id":               slugOrID(slug, id),
			"title":            title,
			"linkedArtifacts":  []string{},
			"outcome":          outcome,
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
	var id, createdSlug, outTitle, outcome, createdDate string
	var hasTaskLink, isOrphan bool
	var revision int
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`INSERT INTO feedback (project_id, slug, title, owner_user_id, outcome, is_orphan)
		 VALUES ($1::uuid, $2, $3, $4::uuid, 'Needs Iteration', TRUE)
		 RETURNING id::text, slug, title, COALESCE(outcome::text, 'Needs Iteration'), COALESCE(to_char(created_at, 'YYYY-MM-DD'), ''), primary_task_id IS NOT NULL, is_orphan, document_revision`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &createdSlug, &outTitle, &outcome, &createdDate, &hasTaskLink, &isOrphan, &revision)
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
		"id":               slugOrID(createdSlug, id),
		"title":            outTitle,
		"linkedArtifacts":  []string{},
		"outcome":          outcome,
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
	var id, foundSlug, title, outcome, ownerName, createdDate string
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`SELECT f.id::text, f.slug, f.title, COALESCE(f.outcome::text, 'Needs Iteration'), COALESCE(u.name, ''), COALESCE(to_char(f.created_at, 'YYYY-MM-DD'), '')
		 FROM feedback f
		 JOIN users u ON u.id = f.owner_user_id
		 WHERE f.project_id = $1::uuid AND (f.slug = $2 OR f.id::text = $2)
		 LIMIT 1`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &foundSlug, &title, &outcome, &ownerName, &createdDate)
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
	return map[string]any{
		"feedback": map[string]any{
			"id":          slugOrID(foundSlug, id),
			"title":       title,
			"outcome":     outcome,
			"owner":       ownerName,
			"createdDate": createdDate,
		},
		"detail":    content,
		"reference": map[string]any{"taskOptions": []any{}, "ideaOptions": []any{}, "problemOptions": []any{}, "permissions": map[string]any{}},
	}, nil
}

func (r *repo) UpdateFeedback(ctx context.Context, projectID, feedbackID, actorUserID string, patch map[string]any) (map[string]any, error) {
	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	outcome := toString(patch["outcome"])
	linkedArtifacts := toSlice(patch["linkedArtifacts"])

	primaryTaskID, linkIDs, err := r.resolveFeedbackLinks(ctx, identity.UUID, linkedArtifacts)
	if err != nil {
		return nil, err
	}

	var id, slug, title, outOutcome, owner, createdDate string
	var hasTaskLink, isOrphan bool
	var revision int
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`UPDATE feedback f
		 SET outcome = COALESCE(NULLIF($3, '')::feedback_outcome, f.outcome),
		 	 primary_task_id = $4::uuid,
		 	 updated_at = NOW(),
		 	 document_revision = document_revision + 1
		 FROM users u
		 WHERE f.project_id = $1::uuid
		   AND (f.id::text = $2 OR f.slug = $2)
		   AND u.id = f.owner_user_id
		 RETURNING f.id::text, f.slug, f.title, COALESCE(f.outcome::text, 'Needs Iteration'), COALESCE(u.name, ''),
		 	 COALESCE(to_char(f.created_at, 'YYYY-MM-DD'), ''), f.primary_task_id IS NOT NULL, f.is_orphan, f.document_revision`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &slug, &title, &outOutcome, &owner, &createdDate, &hasTaskLink, &isOrphan, &revision)
		},
		identity.UUID,
		strings.TrimSpace(feedbackID),
		outcome,
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
		"id":               slugOrID(slug, id),
		"title":            title,
		"linkedArtifacts":  []string{},
		"outcome":          outOutcome,
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

	documentID := fmt.Sprintf("%s:%s", artifactType, artifactID)
	doc := map[string]any{
		"artifact_id":        artifactID,
		"project_id":         projectUUID,
		"document_id":        documentID,
		"revision":           revision,
		"updated_at":         time.Now().UTC(),
		"updated_by_user_id": strings.TrimSpace(actorUserID),
		"schema_version":     1,
		"content":            cloneMap(content),
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
	query := fmt.Sprintf("SELECT id::text FROM %s WHERE project_id = $1::uuid AND (id::text = $2 OR slug = $2) LIMIT 1", table)
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
			continue
		}
		typeRaw := firstNonEmpty(toString(item["type"]), toString(item["artifactType"]))
		normalized, ok := normalizeArtifactType(typeRaw)
		if !ok || (normalized != "story" && normalized != "journey") {
			continue
		}
		identifier := toString(item["id"])
		if identifier == "" {
			continue
		}
		sourceID, err := r.resolveArtifactIDByIdentifier(ctx, projectUUID, normalized, identifier)
		if err != nil {
			return err
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

func slugOrID(slug, id string) string {
	if strings.TrimSpace(slug) != "" {
		return strings.TrimSpace(slug)
	}
	return strings.TrimSpace(id)
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
		"context":    "",
		"empathyMap": map[string]any{"says": "", "thinks": "", "does": "", "feels": ""},
		"painPoints": []any{},
		"hypothesis": []any{},
		"notes":      "",
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
		"outcome":         "Needs Iteration",
		"linkedArtifacts": []any{},
		"activeModules":   []any{},
		"moduleContent":   map[string]any{},
		"notes":           "",
		"title":           title,
	}
}
