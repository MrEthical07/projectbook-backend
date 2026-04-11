package pages

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/storage"
	"github.com/jackc/pgx/v5"
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
}

type projectIdentity struct {
	UUID string
	Slug string
}

func NewRepo(store storage.RelationalStore) Repo {
	return &repo{store: store}
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
	COALESCE((
		SELECT jsonb_array_length(COALESCE(o.payload->'content'->'linkedArtifacts', '[]'::jsonb))
		FROM document_sync_outbox o
		WHERE o.project_id = p.project_id
		  AND o.artifact_type = 'page'::artifact_type
		  AND o.artifact_id = p.id
		  AND o.operation = 'upsert'
		ORDER BY o.document_revision DESC
		LIMIT 1
	), 0),
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
	args = append(args, query.Offset, query.Limit)
	sql += fmt.Sprintf("ORDER BY p.updated_at DESC OFFSET $%d LIMIT $%d", len(args)-1, len(args))

	err = r.store.Execute(ctx, storage.RelationalQueryMany(sql, func(row storage.RowScanner) error {
		var id, slug, title, owner, lastEdited, status string
		var linkedCount int
		var isOrphan bool
		if scanErr := row.Scan(&id, &slug, &title, &owner, &lastEdited, &linkedCount, &status, &isOrphan); scanErr != nil {
			return scanErr
		}
		items = append(items, map[string]any{
			"id":                   slugOrID(slug, id),
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
	if err := r.enqueueUpsertOutbox(ctx, identity.UUID, id, revision, actorUserID, defaultPageContent(createdTitle)); err != nil {
		return nil, err
	}
	if err := r.logActivity(ctx, identity.UUID, actorUserID, id, "created Page", map[string]any{
		"artifact": createdTitle,
		"href":     fmt.Sprintf("/project/%s/pages/%s", identity.Slug, slugOrID(createdSlug, id)),
	}); err != nil {
		return nil, err
	}
	return map[string]any{
		"id":                   slugOrID(createdSlug, id),
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
		 WHERE p.project_id = $1::uuid AND (p.slug = $2 OR p.id::text = $2)
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
			"id":         slugOrID(foundSlug, id),
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
		 SET status = COALESCE(NULLIF($3, '')::page_status, p.status),
		     is_orphan = CASE WHEN $4 < 0 THEN p.is_orphan ELSE ($4 = 0) END,
		     updated_at = NOW(),
		     document_revision = p.document_revision + 1
		 FROM users u
		 WHERE p.project_id = $1::uuid AND (p.id::text = $2 OR p.slug = $2)
		   AND u.id = p.owner_user_id
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

	if err := r.enqueueUpsertOutbox(ctx, identity.UUID, id, revision, actorUserID, state); err != nil {
		return nil, err
	}
	if err := r.logActivity(ctx, identity.UUID, actorUserID, id, "updated Page", map[string]any{
		"artifact": title,
		"href":     fmt.Sprintf("/project/%s/pages/%s", identity.Slug, slugOrID(slug, id)),
	}); err != nil {
		return nil, err
	}
	return map[string]any{
		"id":                   slugOrID(slug, id),
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
		 WHERE project_id = $1::uuid AND (id::text = $2 OR slug = $2)
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
	if err := r.enqueueUpsertOutbox(ctx, identity.UUID, id, revision, actorUserID, map[string]any{"title": outTitle}); err != nil {
		return nil, err
	}
	if err := r.logActivity(ctx, identity.UUID, actorUserID, id, "renamed Page", map[string]any{
		"artifact": outTitle,
		"href":     fmt.Sprintf("/project/%s/pages/%s", identity.Slug, slugOrID(slug, id)),
	}); err != nil {
		return nil, err
	}
	return map[string]any{"id": slugOrID(slug, id), "title": outTitle, "lastEdited": lastEdited}, nil
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
		 WHERE project_id = $1::uuid AND (id::text = $2 OR slug = $2)
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
	if err := r.enqueueDeleteOutbox(ctx, identity.UUID, id, actorUserID); err != nil {
		return nil, err
	}
	if err := r.logActivity(ctx, identity.UUID, actorUserID, id, "deleted Page", map[string]any{
		"artifact": title,
		"href":     fmt.Sprintf("/project/%s/pages/%s", identity.Slug, slugOrID(slug, id)),
	}); err != nil {
		return nil, err
	}
	return map[string]any{"id": slugOrID(slug, id)}, nil
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

func (r *repo) enqueueUpsertOutbox(ctx context.Context, projectUUID, pageID string, revision int, actorUserID string, content map[string]any) error {
	payload := map[string]any{
		"content":            content,
		"updated_by_user_id": strings.TrimSpace(actorUserID),
	}
	bytes, err := json.Marshal(payload)
	if err != nil {
		return apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "failed to encode page payload"), err)
	}
	documentID := fmt.Sprintf("page:%s", pageID)
	if err := r.store.Execute(ctx, storage.RelationalExec(
		`INSERT INTO document_sync_outbox (project_id, artifact_type, artifact_id, operation, document_id, document_revision, payload, status, next_attempt_at, updated_at)
		 VALUES ($1::uuid, 'page'::artifact_type, $2::uuid, 'upsert', $3, $4, $5::jsonb, 'pending', NOW(), NOW())
		 ON CONFLICT (project_id, artifact_type, artifact_id, document_revision, operation)
		 DO UPDATE SET payload = EXCLUDED.payload, status = 'pending', next_attempt_at = NOW(), updated_at = NOW(), last_error = NULL`,
		projectUUID,
		pageID,
		documentID,
		revision,
		string(bytes),
	)); err != nil {
		return wrapRepoError("enqueue page outbox", err)
	}
	return nil
}

func (r *repo) enqueueDeleteOutbox(ctx context.Context, projectUUID, pageID, actorUserID string) error {
	payload := map[string]any{"deleted_by_user_id": strings.TrimSpace(actorUserID)}
	bytes, err := json.Marshal(payload)
	if err != nil {
		return apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "failed to encode page delete payload"), err)
	}
	documentID := fmt.Sprintf("page:%s", pageID)
	if err := r.store.Execute(ctx, storage.RelationalExec(
		`INSERT INTO document_sync_outbox (project_id, artifact_type, artifact_id, operation, document_id, document_revision, payload, status, next_attempt_at, updated_at)
		 VALUES ($1::uuid, 'page'::artifact_type, $2::uuid, 'delete', $3, 1, $4::jsonb, 'pending', NOW(), NOW())
		 ON CONFLICT (project_id, artifact_type, artifact_id, document_revision, operation)
		 DO UPDATE SET payload = EXCLUDED.payload, status = 'pending', next_attempt_at = NOW(), updated_at = NOW(), last_error = NULL`,
		projectUUID,
		pageID,
		documentID,
		string(bytes),
	)); err != nil {
		return wrapRepoError("enqueue page delete", err)
	}
	return nil
}

func (r *repo) loadLatestContent(ctx context.Context, projectUUID, pageID string, fallback map[string]any) (map[string]any, error) {
	content := cloneMap(fallback)
	var payloadRaw []byte
	err := r.store.Execute(ctx, storage.RelationalQueryOne(
		`SELECT payload
		 FROM document_sync_outbox
		 WHERE project_id = $1::uuid AND artifact_type = 'page'::artifact_type AND artifact_id = $2::uuid AND operation = 'upsert'
		 ORDER BY document_revision DESC
		 LIMIT 1`,
		func(row storage.RowScanner) error { return row.Scan(&payloadRaw) },
		projectUUID,
		pageID,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return content, nil
		}
		return nil, wrapRepoError("load page content", err)
	}
	payload := make(map[string]any)
	if unmarshalErr := json.Unmarshal(payloadRaw, &payload); unmarshalErr != nil {
		return content, nil
	}
	if payloadContent := toMap(payload["content"]); payloadContent != nil {
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

func slugOrID(slug, id string) string {
	if strings.TrimSpace(slug) != "" {
		return strings.TrimSpace(slug)
	}
	return strings.TrimSpace(id)
}

func (r *repo) requireStore() error {
	if r == nil || r.store == nil {
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
