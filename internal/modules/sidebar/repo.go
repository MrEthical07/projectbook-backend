package sidebar

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/storage"
	"github.com/MrEthical07/superapi/internal/modules/artifacts"
	"github.com/MrEthical07/superapi/internal/modules/pages"
	"github.com/jackc/pgx/v5"
)

type Repo interface {
	CreateSidebarArtifact(ctx context.Context, projectID, actorUserID, prefix, title string) (map[string]any, error)
	RenameSidebarArtifact(ctx context.Context, projectID, artifactID, actorUserID, prefix, title string) (map[string]any, error)
	DeleteSidebarArtifact(ctx context.Context, projectID, artifactID, actorUserID, prefix string) (map[string]any, error)
}

type repo struct {
	store         storage.RelationalStore
	docs          storage.DocumentStore
	artifactsRepo artifacts.Repo
	pagesSvc      pages.Service
}

type projectIdentity struct {
	UUID string
	Slug string
}

type artifactSpec struct {
	Prefix       string
	ArtifactType string
	Table        string
	PathSegment  string
	Label        string
}

func NewRepo(store storage.RelationalStore, docs storage.DocumentStore, artifactsRepo artifacts.Repo, pagesSvc pages.Service) Repo {
	return &repo{store: store, docs: docs, artifactsRepo: artifactsRepo, pagesSvc: pagesSvc}
}

func (r *repo) CreateSidebarArtifact(ctx context.Context, projectID, actorUserID, prefix, title string) (map[string]any, error) {
	normalized := normalizePrefix(prefix)
	spec, err := specByPrefix(normalized)
	if err != nil {
		return nil, err
	}
	identity, err := r.resolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}

	var created map[string]any
	switch normalized {
	case prefixStories:
		created, err = r.artifactsRepo.CreateStory(ctx, projectID, actorUserID, title, nil)
	case prefixJourneys:
		created, err = r.artifactsRepo.CreateJourney(ctx, projectID, actorUserID, title, nil)
	case prefixProblemStatement:
		created, err = r.artifactsRepo.CreateProblem(ctx, projectID, actorUserID, title, nil)
	case prefixIdeas:
		created, err = r.artifactsRepo.CreateIdea(ctx, projectID, actorUserID, title, nil)
	case prefixTasks:
		created, err = r.artifactsRepo.CreateTask(ctx, projectID, actorUserID, title, nil)
	case prefixFeedback:
		created, err = r.artifactsRepo.CreateFeedback(ctx, projectID, actorUserID, title, nil)
	case prefixPages:
		created, err = r.pagesSvc.CreatePageForSidebar(ctx, projectID, actorUserID, title)
	default:
		err = apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid prefix")
	}
	if err != nil {
		return nil, wrapRepoError("create sidebar artifact", err)
	}
	id := mapString(created, "id")
	if id == "" {
		return nil, apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "artifact id missing")
	}
	outTitle := mapString(created, "title", "statement", "name")
	if outTitle == "" {
		outTitle = strings.TrimSpace(title)
	}

	if normalized != prefixPages {
		if err := r.logActivity(ctx, identity.UUID, actorUserID, spec.ArtifactType, id, "created "+spec.Label, map[string]any{
			"artifact": outTitle,
			"href":     fmt.Sprintf("/project/%s/%s/%s", identity.Slug, spec.PathSegment, id),
		}); err != nil {
			return nil, err
		}
	}

	return map[string]any{"id": id, "title": outTitle}, nil
}

func (r *repo) RenameSidebarArtifact(ctx context.Context, projectID, artifactID, actorUserID, prefix, title string) (map[string]any, error) {
	normalized := normalizePrefix(prefix)
	spec, err := specByPrefix(normalized)
	if err != nil {
		return nil, err
	}
	identity, err := r.resolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}

	var updated map[string]any
	switch normalized {
	case prefixStories:
		updated, err = r.artifactsRepo.UpdateStory(ctx, projectID, artifactID, actorUserID, map[string]any{"title": title})
	case prefixJourneys:
		updated, err = r.artifactsRepo.UpdateJourney(ctx, projectID, artifactID, actorUserID, map[string]any{"title": title})
	case prefixProblemStatement:
		updated, err = r.artifactsRepo.UpdateProblem(ctx, projectID, artifactID, actorUserID, map[string]any{"title": title})
	case prefixIdeas:
		updated, err = r.artifactsRepo.UpdateIdea(ctx, projectID, artifactID, actorUserID, map[string]any{"title": title})
	case prefixTasks:
		updated, err = r.artifactsRepo.UpdateTask(ctx, projectID, artifactID, actorUserID, map[string]any{"title": title})
	case prefixFeedback:
		updated, err = r.artifactsRepo.UpdateFeedback(ctx, projectID, artifactID, actorUserID, map[string]any{"title": title})
	case prefixPages:
		updated, err = r.pagesSvc.RenamePageForSidebar(ctx, projectID, artifactID, actorUserID, title)
	default:
		err = apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid prefix")
	}
	if err != nil {
		return nil, wrapRepoError("rename sidebar artifact", err)
	}
	id := mapString(updated, "id")
	if id == "" {
		id = strings.TrimSpace(artifactID)
	}
	outTitle := mapString(updated, "title", "statement", "name")
	if outTitle == "" {
		outTitle = strings.TrimSpace(title)
	}

	if normalized != prefixPages {
		if err := r.logActivity(ctx, identity.UUID, actorUserID, spec.ArtifactType, id, "renamed "+spec.Label, map[string]any{
			"artifact": outTitle,
			"href":     fmt.Sprintf("/project/%s/%s/%s", identity.Slug, spec.PathSegment, id),
		}); err != nil {
			return nil, err
		}
	}

	return map[string]any{"id": id, "title": outTitle}, nil
}

func (r *repo) DeleteSidebarArtifact(ctx context.Context, projectID, artifactID, actorUserID, prefix string) (map[string]any, error) {
	normalized := normalizePrefix(prefix)
	spec, err := specByPrefix(normalized)
	if err != nil {
		return nil, err
	}
	identity, err := r.resolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if normalized == prefixPages {
		deleted, deleteErr := r.pagesSvc.DeletePageForSidebar(ctx, projectID, artifactID, actorUserID)
		if deleteErr != nil {
			return nil, wrapRepoError("delete sidebar page", deleteErr)
		}
		return map[string]any{"id": mapString(deleted, "id")}, nil
	}

	id, title, err := r.deleteArtifactRow(ctx, identity.UUID, spec.Table, artifactID)
	if err != nil {
		return nil, err
	}
	if err := r.deleteDocument(ctx, identity.UUID, spec.ArtifactType, id, actorUserID); err != nil {
		return nil, err
	}
	if err := r.logActivity(ctx, identity.UUID, actorUserID, spec.ArtifactType, id, "deleted "+spec.Label, map[string]any{
		"artifact": title,
		"href":     fmt.Sprintf("/project/%s/%s/%s", identity.Slug, spec.PathSegment, id),
	}); err != nil {
		return nil, err
	}
	return map[string]any{"id": id}, nil
}

func (r *repo) deleteArtifactRow(ctx context.Context, projectUUID, table, artifactID string) (string, string, error) {
	query := fmt.Sprintf(`DELETE FROM %s WHERE project_id = $1::uuid AND (id::text = $2 OR slug = $2) RETURNING id::text, title`, table)
	var id, title string
	err := r.store.Execute(ctx, storage.RelationalQueryOne(query,
		func(row storage.RowScanner) error { return row.Scan(&id, &title) },
		projectUUID,
		strings.TrimSpace(artifactID),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", apperr.New(apperr.CodeNotFound, http.StatusNotFound, "artifact not found")
		}
		return "", "", wrapRepoError("delete artifact row", err)
	}
	return id, strings.TrimSpace(title), nil
}

func (r *repo) deleteDocument(ctx context.Context, projectUUID, artifactType, artifactID, actorUserID string) error {
	_ = actorUserID

	collection, err := documentCollectionByArtifactType(artifactType)
	if err != nil {
		return err
	}

	if err := r.docs.Execute(ctx, storage.DocumentRun(
		collection+":delete_one",
		map[string]any{
			"filter": map[string]any{
				"artifact_id": artifactID,
				"project_id":  projectUUID,
			},
		},
		nil,
	)); err != nil {
		return wrapRepoError("delete sidebar artifact document", err)
	}

	return nil
}

func (r *repo) logActivity(ctx context.Context, projectUUID, actorUserID, artifactType, artifactID, action string, payload map[string]any) error {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "failed to encode activity payload"), err)
	}
	if err := r.store.Execute(ctx, storage.RelationalExec(
		`INSERT INTO activity_log (project_id, actor_user_id, artifact_type, artifact_id, action, payload)
		 VALUES ($1::uuid, $2::uuid, $3::artifact_type, $4::uuid, $5, $6::jsonb)`,
		projectUUID,
		strings.TrimSpace(actorUserID),
		artifactType,
		artifactID,
		action,
		string(bytes),
	)); err != nil {
		return wrapRepoError("log sidebar activity", err)
	}
	return nil
}

func (r *repo) resolveProjectIdentity(ctx context.Context, projectID string) (projectIdentity, error) {
	if err := r.requireStore(); err != nil {
		return projectIdentity{}, err
	}
	var identity projectIdentity
	err := r.store.Execute(ctx, storage.RelationalQueryOne(
		`SELECT id::text, slug FROM projects WHERE slug = $1 OR id::text = $1 LIMIT 1`,
		func(row storage.RowScanner) error { return row.Scan(&identity.UUID, &identity.Slug) },
		strings.TrimSpace(projectID),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return projectIdentity{}, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "project not found")
		}
		return projectIdentity{}, wrapRepoError("resolve project", err)
	}
	return identity, nil
}

func specByPrefix(prefix string) (artifactSpec, error) {
	switch normalizePrefix(prefix) {
	case prefixStories:
		return artifactSpec{Prefix: prefixStories, ArtifactType: "story", Table: "stories", PathSegment: "stories", Label: "Story"}, nil
	case prefixJourneys:
		return artifactSpec{Prefix: prefixJourneys, ArtifactType: "journey", Table: "journeys", PathSegment: "journeys", Label: "Journey"}, nil
	case prefixProblemStatement:
		return artifactSpec{Prefix: prefixProblemStatement, ArtifactType: "problem", Table: "problems", PathSegment: "problem-statement", Label: "Problem Statement"}, nil
	case prefixIdeas:
		return artifactSpec{Prefix: prefixIdeas, ArtifactType: "idea", Table: "ideas", PathSegment: "ideas", Label: "Idea"}, nil
	case prefixTasks:
		return artifactSpec{Prefix: prefixTasks, ArtifactType: "task", Table: "tasks", PathSegment: "tasks", Label: "Task"}, nil
	case prefixFeedback:
		return artifactSpec{Prefix: prefixFeedback, ArtifactType: "feedback", Table: "feedback", PathSegment: "feedback", Label: "Feedback"}, nil
	case prefixPages:
		return artifactSpec{Prefix: prefixPages, ArtifactType: "page", Table: "pages", PathSegment: "pages", Label: "Page"}, nil
	default:
		return artifactSpec{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid prefix")
	}
}

func documentCollectionByArtifactType(artifactType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(artifactType)) {
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
	case "page":
		return "page_documents", nil
	case "resource":
		return "resource_documents", nil
	default:
		return "", apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "unsupported artifact type")
	}
}

func (r *repo) requireStore() error {
	if r == nil || r.store == nil || r.docs == nil {
		return apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "sidebar repository unavailable")
	}
	if r.artifactsRepo == nil || r.pagesSvc == nil {
		return apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "sidebar dependencies unavailable")
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
	return apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "failed to process sidebar data"), fmt.Errorf("%s: %w", action, err))
}
