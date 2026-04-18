package pages

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/patchx"
	"github.com/MrEthical07/superapi/internal/core/storage"
	"github.com/jackc/pgx/v5"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repo interface {
	ListPages(ctx context.Context, projectID string, query listQuery) ([]map[string]any, error)
	CreatePage(ctx context.Context, projectID, actorUserID, title string) (map[string]any, error)
	GetPage(ctx context.Context, projectID, slug string) (map[string]any, error)
	UpdatePage(ctx context.Context, projectID, pageID, actorUserID string, state map[string]any) (map[string]any, error)
	RenamePage(ctx context.Context, projectID, pageID, title, actorUserID string) (map[string]any, error)

	CreatePageForSidebar(ctx context.Context, projectID, actorUserID, title string) (map[string]any, error)
	RenamePageForSidebar(ctx context.Context, projectID, pageID, title, actorUserID string) (map[string]any, error)
	DeletePageForSidebar(ctx context.Context, projectID, pageID, actorUserID string) (map[string]any, error)
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

func (r *repo) ListPages(ctx context.Context, projectID string, query listQuery) ([]map[string]any, error) {
	identity, err := r.resolveProjectIdentity(ctx, projectID)
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
	(SELECT COUNT(1) FROM artifact_links l WHERE l.project_id = p.project_id AND l.source_type = 'page'::artifact_type AND l.source_id = p.id),
	p.status::text,
	p.is_orphan
FROM pages p
JOIN users u ON u.id = p.owner_user_id
WHERE p.project_id = $1::uuid
`
	if query.Status != "" {
		args = append(args, query.Status)
		sql += fmt.Sprintf(" AND p.status = $%d::page_status\n", len(args))
	}
	args = append(args, query.Offset, query.Limit+1)
	sql += fmt.Sprintf("ORDER BY p.updated_at DESC OFFSET $%d LIMIT $%d", len(args)-1, len(args))

	err = r.store.Execute(ctx, storage.RelationalQueryMany(sql, func(row storage.RowScanner) error {
		var id, slug, title, owner, lastEdited, status string
		var linkedCount int
		var isOrphan bool
		if scanErr := row.Scan(&id, &slug, &title, &owner, &lastEdited, &linkedCount, &status, &isOrphan); scanErr != nil {
			return scanErr
		}
		items = append(items, map[string]any{
			"id":                   id,
			"title":                title,
			"owner":                owner,
			"lastEdited":           lastEdited,
			"linkedArtifactsCount": linkedCount,
			"status":               status,
			"isOrphan":             isOrphan,
		})
		return nil
	}, args...))
	if err != nil {
		return nil, wrapRepoError("list pages", err)
	}
	return items, nil
}

func (r *repo) CreatePage(ctx context.Context, projectID, actorUserID, title string) (map[string]any, error) {
	identity, err := r.resolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	slug, err := r.nextUniqueSlug(ctx, "pages", identity.UUID, title)
	if err != nil {
		return nil, err
	}
	ownerName, err := r.resolveUserName(ctx, actorUserID)
	if err != nil {
		return nil, err
	}

	var id, createdSlug, createdTitle, lastEdited, status string
	var isOrphan bool
	var revision int
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`INSERT INTO pages (project_id, slug, title, owner_user_id, status, is_orphan)
		 VALUES ($1::uuid, $2, $3, $4::uuid, 'Draft', TRUE)
		 RETURNING id::text, slug, title, COALESCE(to_char(updated_at, 'YYYY-MM-DD'), ''), status::text, is_orphan, document_revision`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &createdSlug, &createdTitle, &lastEdited, &status, &isOrphan, &revision)
		},
		identity.UUID,
		slug,
		strings.TrimSpace(title),
		strings.TrimSpace(actorUserID),
	))
	if err != nil {
		return nil, wrapRepoError("create page", err)
	}
	if err := r.upsertDocument(ctx, identity.UUID, id, revision, actorUserID, defaultPageContent(createdTitle)); err != nil {
		return nil, err
	}
	if err := r.logActivity(ctx, identity.UUID, actorUserID, id, "created Page", map[string]any{
		"artifact": createdTitle,
		"href":     fmt.Sprintf("/project/%s/pages/%s", identity.UUID, id),
	}); err != nil {
		return nil, err
	}
	return map[string]any{
		"id":                   id,
		"title":                createdTitle,
		"owner":                ownerName,
		"lastEdited":           lastEdited,
		"linkedArtifactsCount": 0,
		"status":               status,
		"isOrphan":             isOrphan,
	}, nil
}

func (r *repo) GetPage(ctx context.Context, projectID, slug string) (map[string]any, error) {
	identity, err := r.resolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	var id, foundSlug, title, status, owner, lastEdited string
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`SELECT p.id::text, p.slug, p.title, p.status::text, COALESCE(u.name, ''), COALESCE(to_char(p.updated_at, 'YYYY-MM-DD'), '')
		 FROM pages p
		 JOIN users u ON u.id = p.owner_user_id
		 WHERE p.project_id = $1::uuid AND p.id::text = $2
		 LIMIT 1`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &foundSlug, &title, &status, &owner, &lastEdited)
		},
		identity.UUID,
		strings.TrimSpace(slug),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "page not found")
		}
		return nil, wrapRepoError("get page", err)
	}

	content, err := r.loadLatestContent(ctx, identity.UUID, id, defaultPageContent(title))
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"page": map[string]any{
			"id":         id,
			"title":      title,
			"status":     status,
			"owner":      owner,
			"lastEdited": lastEdited,
		},
		"detail": content,
		"reference": map[string]any{
			"tagOptions":            []string{"Research", "Alignment", "Notes", "Strategy"},
			"linkedArtifactOptions": []any{},
			"permissions":           map[string]any{},
		},
	}, nil
}

func (r *repo) UpdatePage(ctx context.Context, projectID, pageID, actorUserID string, state map[string]any) (map[string]any, error) {
	identity, err := r.resolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	status := toString(state["status"])
	linkedCount := -1
	if links, ok := state["linkedArtifacts"]; ok {
		linkedCount = len(toSlice(links))
	}

	var id, slug, title, owner, lastEdited, outStatus string
	var isOrphan bool
	var revision int
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`UPDATE pages p
		 SET status = CASE
		 	 	WHEN NULLIF($3, '')::page_status = 'Archived'::page_status
		 	 		AND p.status <> 'Archived'::page_status
		 	 	THEN 'Archived'::page_status
		 	 	WHEN NULLIF($3, '')::page_status IS NOT NULL
		 	 		AND p.status = 'Archived'::page_status
		 	 		AND NULLIF($3, '')::page_status <> 'Archived'::page_status
		 	 	THEN COALESCE(p.archived_from_status, 'Draft'::page_status)
		 	 	ELSE COALESCE(NULLIF($3, '')::page_status, p.status)
		 	 END,
		 	 archived_from_status = CASE
		 	 	WHEN NULLIF($3, '')::page_status = 'Archived'::page_status
		 	 		AND p.status <> 'Archived'::page_status
		 	 	THEN p.status
		 	 	WHEN NULLIF($3, '')::page_status IS NOT NULL
		 	 		AND p.status = 'Archived'::page_status
		 	 		AND NULLIF($3, '')::page_status <> 'Archived'::page_status
		 	 	THEN NULL
		 	 	ELSE p.archived_from_status
		 	 END,
		     is_orphan = CASE WHEN $4 < 0 THEN p.is_orphan ELSE ($4 = 0) END,
		     updated_at = NOW(),
		     document_revision = p.document_revision + 1
		 FROM users u
		 WHERE p.project_id = $1::uuid AND p.id::text = $2
		   AND u.id = p.owner_user_id
		   AND (
		 	 p.status <> 'Archived'::page_status
		 	 OR NULLIF($3, '')::page_status = 'Archived'::page_status
		 	 OR (
		 	 	p.status = 'Archived'::page_status
		 	 	AND NULLIF($3, '')::page_status IS NOT NULL
		 	 	AND NULLIF($3, '')::page_status <> 'Archived'::page_status
		 	 )
		   )
		 RETURNING p.id::text, p.slug, p.title, COALESCE(u.name, ''), COALESCE(to_char(p.updated_at, 'YYYY-MM-DD'), ''), p.status::text, p.is_orphan, p.document_revision`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &slug, &title, &owner, &lastEdited, &outStatus, &isOrphan, &revision)
		},
		identity.UUID,
		strings.TrimSpace(pageID),
		status,
		linkedCount,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "page not found")
		}
		return nil, wrapRepoError("update page", err)
	}

	if err := r.upsertDocument(ctx, identity.UUID, id, revision, actorUserID, state); err != nil {
		return nil, err
	}
	if err := r.logActivity(ctx, identity.UUID, actorUserID, id, "updated Page", map[string]any{
		"artifact": title,
		"href":     fmt.Sprintf("/project/%s/pages/%s", identity.UUID, id),
	}); err != nil {
		return nil, err
	}
	return map[string]any{
		"id":                   id,
		"title":                title,
		"owner":                owner,
		"lastEdited":           lastEdited,
		"linkedArtifactsCount": maxLinkedArtifactsFromState(state),
		"status":               outStatus,
		"isOrphan":             isOrphan,
	}, nil
}

func (r *repo) RenamePage(ctx context.Context, projectID, pageID, title, actorUserID string) (map[string]any, error) {
	identity, err := r.resolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	var id, slug, outTitle, lastEdited string
	var revision int
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`UPDATE pages
		 SET title = $3,
		     updated_at = NOW(),
		     document_revision = document_revision + 1
		 WHERE project_id = $1::uuid
		   AND id::text = $2
		   AND status <> 'Archived'::page_status
		 RETURNING id::text, slug, title, COALESCE(to_char(updated_at, 'YYYY-MM-DD'), ''), document_revision`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &slug, &outTitle, &lastEdited, &revision)
		},
		identity.UUID,
		strings.TrimSpace(pageID),
		strings.TrimSpace(title),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "page not found")
		}
		return nil, wrapRepoError("rename page", err)
	}
	if err := r.upsertDocument(ctx, identity.UUID, id, revision, actorUserID, map[string]any{"title": outTitle}); err != nil {
		return nil, err
	}
	if err := r.logActivity(ctx, identity.UUID, actorUserID, id, "renamed Page", map[string]any{
		"artifact": outTitle,
		"href":     fmt.Sprintf("/project/%s/pages/%s", identity.UUID, id),
	}); err != nil {
		return nil, err
	}
	return map[string]any{"id": id, "title": outTitle, "lastEdited": lastEdited}, nil
}

func (r *repo) CreatePageForSidebar(ctx context.Context, projectID, actorUserID, title string) (map[string]any, error) {
	return r.CreatePage(ctx, projectID, actorUserID, title)
}

func (r *repo) RenamePageForSidebar(ctx context.Context, projectID, pageID, title, actorUserID string) (map[string]any, error) {
	return r.RenamePage(ctx, projectID, pageID, title, actorUserID)
}

func (r *repo) DeletePageForSidebar(ctx context.Context, projectID, pageID, actorUserID string) (map[string]any, error) {
	identity, err := r.resolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	var id, slug, title string
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`DELETE FROM pages
		 WHERE project_id = $1::uuid
		   AND id::text = $2
		   AND status <> 'Archived'::page_status
		 RETURNING id::text, slug, title`,
		func(row storage.RowScanner) error { return row.Scan(&id, &slug, &title) },
		identity.UUID,
		strings.TrimSpace(pageID),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "page not found")
		}
		return nil, wrapRepoError("delete page", err)
	}
	if err := r.deleteDocument(ctx, identity.UUID, id, actorUserID); err != nil {
		return nil, err
	}
	if err := r.logActivity(ctx, identity.UUID, actorUserID, id, "deleted Page", map[string]any{
		"artifact": title,
		"href":     fmt.Sprintf("/project/%s/pages/%s", identity.UUID, id),
	}); err != nil {
		return nil, err
	}
	return map[string]any{"id": id}, nil
}

func (r *repo) resolveProjectIdentity(ctx context.Context, projectID string) (projectIdentity, error) {
	if err := r.requireStore(); err != nil {
		return projectIdentity{}, err
	}
	var identity projectIdentity
	err := r.store.Execute(ctx, storage.RelationalQueryOne(
		`SELECT id::text, slug FROM projects WHERE id::text = $1 LIMIT 1`,
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
		return false, wrapRepoError("check slug", err)
	}
	return exists, nil
}

func (r *repo) upsertDocument(ctx context.Context, projectUUID, pageID string, revision int, actorUserID string, content map[string]any) error {
	existingContent, err := r.loadLatestContent(ctx, projectUUID, pageID, map[string]any{})
	if err != nil {
		return err
	}
	mergedContent := patchx.MergeShallow(existingContent, content)

	documentID := fmt.Sprintf("page:%s", pageID)
	doc := map[string]any{
		"artifact_id":        pageID,
		"project_id":         projectUUID,
		"document_id":        documentID,
		"revision":           revision,
		"updated_at":         time.Now().UTC(),
		"updated_by_user_id": strings.TrimSpace(actorUserID),
		"schema_version":     1,
		"content":            mergedContent,
	}

	if err := r.docs.Execute(ctx, storage.DocumentRun(
		"page_documents:update_one",
		map[string]any{
			"filter":  map[string]any{"artifact_id": pageID},
			"update":  map[string]any{"$set": doc},
			"options": options.Update().SetUpsert(true),
		},
		nil,
	)); err != nil {
		return wrapRepoError("upsert page document", err)
	}

	return nil
}

func (r *repo) deleteDocument(ctx context.Context, projectUUID, pageID, actorUserID string) error {
	_ = actorUserID

	if err := r.docs.Execute(ctx, storage.DocumentRun(
		"page_documents:delete_one",
		map[string]any{
			"filter": map[string]any{
				"artifact_id": pageID,
				"project_id":  projectUUID,
			},
		},
		nil,
	)); err != nil {
		return wrapRepoError("delete page document", err)
	}

	return nil
}

func (r *repo) loadLatestContent(ctx context.Context, projectUUID, pageID string, fallback map[string]any) (map[string]any, error) {
	content := cloneMap(fallback)
	var doc map[string]any
	err := r.docs.Execute(ctx, storage.DocumentRun(
		"page_documents:find_one",
		map[string]any{
			"filter": map[string]any{
				"artifact_id": pageID,
				"project_id":  projectUUID,
			},
		},
		&doc,
	))
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return content, nil
		}
		return nil, wrapRepoError("load page content", err)
	}
	if payloadContent := toMap(doc["content"]); payloadContent != nil {
		for k, v := range payloadContent {
			content[k] = v
		}
	}
	return content, nil
}

func (r *repo) logActivity(ctx context.Context, projectUUID, actorUserID, pageID, action string, payload map[string]any) error {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "failed to encode activity payload"), err)
	}
	if err := r.store.Execute(ctx, storage.RelationalExec(
		`INSERT INTO activity_log (project_id, actor_user_id, artifact_type, artifact_id, action, payload)
		 VALUES ($1::uuid, $2::uuid, 'page'::artifact_type, $3::uuid, $4, $5::jsonb)`,
		projectUUID,
		strings.TrimSpace(actorUserID),
		pageID,
		action,
		string(bytes),
	)); err != nil {
		return wrapRepoError("log activity", err)
	}
	return nil
}

func defaultPageContent(title string) map[string]any {
	return map[string]any{
		"description":     "",
		"tags":            []any{},
		"linkedArtifacts": []any{},
		"docHeading":      title,
		"docBody":         "",
		"views": []any{
			map[string]any{"id": "view-doc", "name": "Document", "type": "Document"},
			map[string]any{"id": "view-table", "name": "Table", "type": "Table"},
			map[string]any{"id": "view-board", "name": "Board", "type": "Board"},
		},
		"activeViewId": "view-doc",
		"tableData":    []any{},
	}
}

func maxLinkedArtifactsFromState(state map[string]any) int {
	if state == nil {
		return 0
	}
	if links, ok := state["linkedArtifacts"]; ok {
		return len(toSlice(links))
	}
	return 0
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

func (r *repo) requireStore() error {
	if r == nil || r.store == nil || r.docs == nil {
		return apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "pages repository unavailable")
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
	return apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "failed to process pages data"), fmt.Errorf("%s: %w", action, err))
}
