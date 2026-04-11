package resources

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
	ListResources(ctx context.Context, projectID string, query listQuery) (map[string]any, error)
	CreateResource(ctx context.Context, projectID, actorUserID, name, docType string) (map[string]any, error)
	GetResource(ctx context.Context, projectID, resourceID string) (map[string]any, error)
	UpdateResource(ctx context.Context, projectID, resourceID, actorUserID string, state map[string]any) (map[string]any, error)
	UpdateResourceStatus(ctx context.Context, projectID, resourceID, status, actorUserID string) (map[string]any, error)
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

func (r *repo) ListResources(ctx context.Context, projectID string, query listQuery) (map[string]any, error) {
	identity, err := r.resolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}

	sortExpr := "r.updated_at"
	switch query.Sort {
	case "name":
		sortExpr = "r.title"
	case "uploadDate":
		sortExpr = "r.created_at"
	}
	orderExpr := "DESC"
	if strings.EqualFold(query.Order, "asc") {
		orderExpr = "ASC"
	}

	rows := make([]map[string]any, 0, query.Limit)
	args := []any{identity.UUID}
	sql := `
SELECT
	r.id::text,
	r.slug,
	r.title,
	r.file_type,
	r.doc_type,
	COALESCE(u.name, ''),
	COALESCE((SELECT MAX(rv.version) FROM resource_versions rv WHERE rv.resource_id = r.id), 1),
	COALESCE(to_char(r.updated_at, 'YYYY-MM-DD'), ''),
	(SELECT COUNT(1) FROM artifact_links l WHERE l.project_id = r.project_id AND l.source_type = 'resource'::artifact_type AND l.source_id = r.id),
	r.status::text
FROM resources r
JOIN users u ON u.id = r.owner_user_id
WHERE r.project_id = $1::uuid
`
	if query.Status != "" {
		args = append(args, query.Status)
		sql += fmt.Sprintf(" AND r.status = $%d::resource_status\n", len(args))
	}
	if strings.TrimSpace(query.DocType) != "" {
		args = append(args, query.DocType)
		sql += fmt.Sprintf(" AND r.doc_type = $%d\n", len(args))
	}
	args = append(args, query.Offset, query.Limit)
	sql += fmt.Sprintf("ORDER BY %s %s OFFSET $%d LIMIT $%d", sortExpr, orderExpr, len(args)-1, len(args))

	err = r.store.Execute(ctx, storage.RelationalQueryMany(sql, func(row storage.RowScanner) error {
		var id, slug, name, fileType, docType, owner, lastUpdated, status string
		var version, linkedCount int
		if scanErr := row.Scan(&id, &slug, &name, &fileType, &docType, &owner, &version, &lastUpdated, &linkedCount, &status); scanErr != nil {
			return scanErr
		}
		rows = append(rows, map[string]any{
			"id":          slugOrID(slug, id),
			"name":        name,
			"fileType":    fileType,
			"docType":     docType,
			"owner":       owner,
			"version":     fmt.Sprintf("v%d", version),
			"lastUpdated": lastUpdated,
			"linkedCount": linkedCount,
			"status":      status,
		})
		return nil
	}, args...))
	if err != nil {
		return nil, wrapRepoError("list resources", err)
	}

	docTypes := make([]string, 0)
	_ = r.store.Execute(ctx, storage.RelationalQueryMany(
		`SELECT DISTINCT doc_type FROM resources WHERE project_id = $1::uuid ORDER BY doc_type ASC`,
		func(row storage.RowScanner) error {
			var value string
			if scanErr := row.Scan(&value); scanErr != nil {
				return scanErr
			}
			docTypes = append(docTypes, value)
			return nil
		},
		identity.UUID,
	))
	if len(docTypes) == 0 {
		docTypes = []string{"Pitch Deck", "Research Paper", "Specification", "Design File", "Other"}
	}

	return map[string]any{
		"rows": rows,
		"reference": map[string]any{
			"docTypes":    docTypes,
			"fileTypes":   []string{"PDF", "PPTX", "DOCX", "XLSX", "Other"},
			"owners":      []string{},
			"sortOptions": []string{"Last Updated", "Name", "Upload Date"},
		},
	}, nil
}

func (r *repo) CreateResource(ctx context.Context, projectID, actorUserID, name, docType string) (map[string]any, error) {
	identity, err := r.resolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	slug, err := r.nextUniqueSlug(ctx, "resources", identity.UUID, name)
	if err != nil {
		return nil, err
	}
	ownerName, err := r.resolveUserName(ctx, actorUserID)
	if err != nil {
		return nil, err
	}

	var id, createdSlug, createdName, fileType, createdDocType, status, lastUpdated string
	var revision int
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`INSERT INTO resources (project_id, slug, title, owner_user_id, status, file_type, doc_type)
		 VALUES ($1::uuid, $2, $3, $4::uuid, 'Active', 'PDF', $5)
		 RETURNING id::text, slug, title, file_type, doc_type, status::text, COALESCE(to_char(updated_at, 'YYYY-MM-DD'), ''), document_revision`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &createdSlug, &createdName, &fileType, &createdDocType, &status, &lastUpdated, &revision)
		},
		identity.UUID,
		slug,
		strings.TrimSpace(name),
		strings.TrimSpace(actorUserID),
		strings.TrimSpace(docType),
	))
	if err != nil {
		return nil, wrapRepoError("create resource", err)
	}

	if err := r.store.Execute(ctx, storage.RelationalExec(
		`INSERT INTO resource_versions (resource_id, project_id, version, title, owner_user_id, document_revision)
		 VALUES ($1::uuid, $2::uuid, 1, $3, $4::uuid, $5)`,
		id,
		identity.UUID,
		createdName,
		strings.TrimSpace(actorUserID),
		revision,
	)); err != nil {
		return nil, wrapRepoError("create resource version", err)
	}

	content := defaultResourceContent(createdName, createdDocType)
	if err := r.enqueueUpsertOutbox(ctx, identity.UUID, id, revision, actorUserID, content); err != nil {
		return nil, err
	}
	if err := r.logActivity(ctx, identity.UUID, actorUserID, id, "created Resource", map[string]any{
		"artifact": createdName,
		"href":     fmt.Sprintf("/project/%s/resources/%s", identity.Slug, slugOrID(createdSlug, id)),
	}); err != nil {
		return nil, err
	}

	return map[string]any{
		"id":          slugOrID(createdSlug, id),
		"name":        createdName,
		"fileType":    fileType,
		"docType":     createdDocType,
		"owner":       ownerName,
		"version":     "v1",
		"lastUpdated": lastUpdated,
		"linkedCount": 0,
		"status":      status,
	}, nil
}

func (r *repo) GetResource(ctx context.Context, projectID, resourceID string) (map[string]any, error) {
	identity, err := r.resolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}

	var id, slug, name, fileType, docType, status, owner, createdAt, updatedAt string
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`SELECT r.id::text, r.slug, r.title, r.file_type, r.doc_type, r.status::text, COALESCE(u.name, ''),
		        COALESCE(to_char(r.created_at, 'YYYY-MM-DD'), ''), COALESCE(to_char(r.updated_at, 'YYYY-MM-DD'), '')
		 FROM resources r
		 JOIN users u ON u.id = r.owner_user_id
		 WHERE r.project_id = $1::uuid AND (r.id::text = $2 OR r.slug = $2)
		 LIMIT 1`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &slug, &name, &fileType, &docType, &status, &owner, &createdAt, &updatedAt)
		},
		identity.UUID,
		strings.TrimSpace(resourceID),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "resource not found")
		}
		return nil, wrapRepoError("get resource", err)
	}

	content, err := r.loadLatestContent(ctx, identity.UUID, id, defaultResourceContent(name, docType))
	if err != nil {
		return nil, err
	}

	versions := make([]map[string]any, 0)
	_ = r.store.Execute(ctx, storage.RelationalQueryMany(
		`SELECT rv.version, COALESCE(u.name, ''), COALESCE(to_char(rv.created_at, 'YYYY-MM-DD'), ''), rv.title
		 FROM resource_versions rv
		 JOIN users u ON u.id = rv.owner_user_id
		 WHERE rv.resource_id = $1::uuid
		 ORDER BY rv.version DESC`,
		func(row storage.RowScanner) error {
			var version int
			var uploadedBy, uploadDate, title string
			if scanErr := row.Scan(&version, &uploadedBy, &uploadDate, &title); scanErr != nil {
				return scanErr
			}
			versions = append(versions, map[string]any{
				"version":     fmt.Sprintf("v%d", version),
				"uploadedBy":  uploadedBy,
				"uploadDate":  uploadDate,
				"description": title,
			})
			return nil
		},
		id,
	))
	if len(versions) == 0 {
		versions = []map[string]any{{
			"version":     "v1",
			"uploadedBy":  owner,
			"uploadDate":  updatedAt,
			"description": name,
		}}
	}
	content["versions"] = versions

	return map[string]any{
		"resource": map[string]any{
			"id":       slugOrID(slug, id),
			"name":     name,
			"fileType": fileType,
			"docType":  docType,
			"status":   status,
			"owner":    owner,
		},
		"detail": content,
		"reference": map[string]any{
			"storyOptions":   []any{},
			"problemOptions": []any{},
			"ideaOptions":    []any{},
			"taskOptions":    []any{},
			"permissions":    map[string]any{},
		},
		"meta": map[string]any{
			"createdAt": createdAt,
			"updatedAt": updatedAt,
		},
	}, nil
}

func (r *repo) UpdateResource(ctx context.Context, projectID, resourceID, actorUserID string, state map[string]any) (map[string]any, error) {
	identity, err := r.resolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	name := firstNonEmpty(toString(state["name"]), toString(state["title"]))
	docType := toString(state["docType"])

	var id, slug, outName, fileType, outDocType, owner, status, lastUpdated string
	var revision int
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`UPDATE resources r
		 SET title = COALESCE(NULLIF($3, ''), r.title),
		     doc_type = COALESCE(NULLIF($4, ''), r.doc_type),
		     updated_at = NOW(),
		     document_revision = r.document_revision + 1
		 FROM users u
		 WHERE r.project_id = $1::uuid AND (r.id::text = $2 OR r.slug = $2)
		   AND u.id = r.owner_user_id
		 RETURNING r.id::text, r.slug, r.title, r.file_type, r.doc_type, COALESCE(u.name, ''),
		           r.status::text, COALESCE(to_char(r.updated_at, 'YYYY-MM-DD'), ''), r.document_revision`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &slug, &outName, &fileType, &outDocType, &owner, &status, &lastUpdated, &revision)
		},
		identity.UUID,
		strings.TrimSpace(resourceID),
		name,
		docType,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "resource not found")
		}
		return nil, wrapRepoError("update resource", err)
	}

	if err := r.insertResourceVersion(ctx, id, identity.UUID, outName, actorUserID, revision); err != nil {
		return nil, err
	}
	if err := r.replaceResourceLinks(ctx, identity.UUID, id, actorUserID, toSlice(state["linkedArtifacts"])); err != nil {
		return nil, err
	}
	if err := r.enqueueUpsertOutbox(ctx, identity.UUID, id, revision, actorUserID, state); err != nil {
		return nil, err
	}
	if err := r.logActivity(ctx, identity.UUID, actorUserID, id, "updated Resource", map[string]any{
		"artifact": outName,
		"href":     fmt.Sprintf("/project/%s/resources/%s", identity.Slug, slugOrID(slug, id)),
	}); err != nil {
		return nil, err
	}

	linkedCount, _ := r.countResourceLinks(ctx, identity.UUID, id)
	version, _ := r.latestResourceVersion(ctx, id)

	return map[string]any{
		"id":          slugOrID(slug, id),
		"name":        outName,
		"fileType":    fileType,
		"docType":     outDocType,
		"owner":       owner,
		"version":     fmt.Sprintf("v%d", version),
		"lastUpdated": lastUpdated,
		"linkedCount": linkedCount,
		"status":      status,
	}, nil
}

func (r *repo) UpdateResourceStatus(ctx context.Context, projectID, resourceID, status, actorUserID string) (map[string]any, error) {
	identity, err := r.resolveProjectIdentity(ctx, projectID)
	if err != nil {
		return nil, err
	}
	var id, slug, name, outStatus, lastUpdated string
	var revision int
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		`UPDATE resources
		 SET status = $3::resource_status,
		     updated_at = NOW(),
		     document_revision = document_revision + 1
		 WHERE project_id = $1::uuid AND (id::text = $2 OR slug = $2)
		 RETURNING id::text, slug, title, status::text, COALESCE(to_char(updated_at, 'YYYY-MM-DD'), ''), document_revision`,
		func(row storage.RowScanner) error {
			return row.Scan(&id, &slug, &name, &outStatus, &lastUpdated, &revision)
		},
		identity.UUID,
		strings.TrimSpace(resourceID),
		status,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, http.StatusNotFound, "resource not found")
		}
		return nil, wrapRepoError("update resource status", err)
	}
	if err := r.enqueueUpsertOutbox(ctx, identity.UUID, id, revision, actorUserID, map[string]any{"status": outStatus}); err != nil {
		return nil, err
	}
	if err := r.logActivity(ctx, identity.UUID, actorUserID, id, "changed Resource status", map[string]any{
		"artifact": name,
		"href":     fmt.Sprintf("/project/%s/resources/%s", identity.Slug, slugOrID(slug, id)),
		"status":   outStatus,
	}); err != nil {
		return nil, err
	}
	return map[string]any{"id": slugOrID(slug, id), "status": outStatus, "lastUpdated": lastUpdated}, nil
}

func (r *repo) insertResourceVersion(ctx context.Context, resourceID, projectUUID, title, actorUserID string, revision int) error {
	if err := r.store.Execute(ctx, storage.RelationalExec(
		`INSERT INTO resource_versions (resource_id, project_id, version, title, owner_user_id, document_revision)
		 SELECT $1::uuid, $2::uuid, COALESCE(MAX(version), 0) + 1, $3, $4::uuid, $5
		 FROM resource_versions
		 WHERE resource_id = $1::uuid`,
		resourceID,
		projectUUID,
		title,
		strings.TrimSpace(actorUserID),
		revision,
	)); err != nil {
		return wrapRepoError("insert resource version", err)
	}
	return nil
}

func (r *repo) replaceResourceLinks(ctx context.Context, projectUUID, resourceID, actorUserID string, linkedArtifacts []any) error {
	if err := r.store.Execute(ctx, storage.RelationalExec(
		`DELETE FROM artifact_links
		 WHERE project_id = $1::uuid
		   AND source_type = 'resource'::artifact_type
		   AND source_id = $2::uuid`,
		projectUUID,
		resourceID,
	)); err != nil {
		return wrapRepoError("clear resource links", err)
	}

	for _, raw := range linkedArtifacts {
		item := toMap(raw)
		if item == nil {
			continue
		}
		typeName := toString(item["type"])
		artifactType, ok := mapLinkedArtifactType(typeName)
		if !ok {
			continue
		}
		identifier := toString(item["id"])
		if identifier == "" {
			continue
		}
		targetID, err := r.resolveArtifactIDByIdentifier(ctx, projectUUID, artifactType, identifier)
		if err != nil {
			continue
		}
		if err := r.store.Execute(ctx, storage.RelationalExec(
			`INSERT INTO artifact_links (project_id, source_type, source_id, target_type, target_id, link_kind, created_by_user_id)
			 VALUES ($1::uuid, 'resource'::artifact_type, $2::uuid, $3::artifact_type, $4::uuid, 'reference', $5::uuid)
			 ON CONFLICT (project_id, source_type, source_id, target_type, target_id, link_kind)
			 DO NOTHING`,
			projectUUID,
			resourceID,
			artifactType,
			targetID,
			strings.TrimSpace(actorUserID),
		)); err != nil {
			return wrapRepoError("insert resource link", err)
		}
	}
	return nil
}

func (r *repo) countResourceLinks(ctx context.Context, projectUUID, resourceID string) (int, error) {
	var count int
	err := r.store.Execute(ctx, storage.RelationalQueryOne(
		`SELECT COUNT(1)
		 FROM artifact_links
		 WHERE project_id = $1::uuid
		   AND source_type = 'resource'::artifact_type
		   AND source_id = $2::uuid`,
		func(row storage.RowScanner) error { return row.Scan(&count) },
		projectUUID,
		resourceID,
	))
	if err != nil {
		return 0, wrapRepoError("count resource links", err)
	}
	return count, nil
}

func (r *repo) latestResourceVersion(ctx context.Context, resourceID string) (int, error) {
	var version int
	err := r.store.Execute(ctx, storage.RelationalQueryOne(
		`SELECT COALESCE(MAX(version), 1) FROM resource_versions WHERE resource_id = $1::uuid`,
		func(row storage.RowScanner) error { return row.Scan(&version) },
		resourceID,
	))
	if err != nil {
		return 1, wrapRepoError("latest resource version", err)
	}
	return version, nil
}

func mapLinkedArtifactType(raw string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "user story", "story", "stories":
		return "story", true
	case "problem statement", "problem", "problems":
		return "problem", true
	case "idea", "ideas":
		return "idea", true
	case "task", "tasks":
		return "task", true
	case "page", "pages":
		return "page", true
	default:
		return "", false
	}
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
			return "", apperr.New(apperr.CodeNotFound, http.StatusNotFound, "artifact not found")
		}
		return "", wrapRepoError("resolve linked artifact", err)
	}
	return id, nil
}

func tableByArtifactType(artifactType string) (string, error) {
	switch artifactType {
	case "story":
		return "stories", nil
	case "problem":
		return "problems", nil
	case "idea":
		return "ideas", nil
	case "task":
		return "tasks", nil
	case "page":
		return "pages", nil
	default:
		return "", apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "unsupported artifact type")
	}
}

func (r *repo) resolveProjectIdentity(ctx context.Context, projectID string) (projectIdentity, error) {
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

func (r *repo) enqueueUpsertOutbox(ctx context.Context, projectUUID, resourceID string, revision int, actorUserID string, content map[string]any) error {
	payload := map[string]any{
		"content":            content,
		"updated_by_user_id": strings.TrimSpace(actorUserID),
	}
	bytes, err := json.Marshal(payload)
	if err != nil {
		return apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "failed to encode resource payload"), err)
	}
	documentID := fmt.Sprintf("resource:%s", resourceID)
	if err := r.store.Execute(ctx, storage.RelationalExec(
		`INSERT INTO document_sync_outbox (project_id, artifact_type, artifact_id, operation, document_id, document_revision, payload, status, next_attempt_at, updated_at)
		 VALUES ($1::uuid, 'resource'::artifact_type, $2::uuid, 'upsert', $3, $4, $5::jsonb, 'pending', NOW(), NOW())
		 ON CONFLICT (project_id, artifact_type, artifact_id, document_revision, operation)
		 DO UPDATE SET payload = EXCLUDED.payload, status = 'pending', next_attempt_at = NOW(), updated_at = NOW(), last_error = NULL`,
		projectUUID,
		resourceID,
		documentID,
		revision,
		string(bytes),
	)); err != nil {
		return wrapRepoError("enqueue resource outbox", err)
	}
	return nil
}

func (r *repo) loadLatestContent(ctx context.Context, projectUUID, resourceID string, fallback map[string]any) (map[string]any, error) {
	content := cloneMap(fallback)
	var payloadRaw []byte
	err := r.store.Execute(ctx, storage.RelationalQueryOne(
		`SELECT payload
		 FROM document_sync_outbox
		 WHERE project_id = $1::uuid AND artifact_type = 'resource'::artifact_type AND artifact_id = $2::uuid AND operation = 'upsert'
		 ORDER BY document_revision DESC
		 LIMIT 1`,
		func(row storage.RowScanner) error { return row.Scan(&payloadRaw) },
		projectUUID,
		resourceID,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return content, nil
		}
		return nil, wrapRepoError("load resource content", err)
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

func (r *repo) logActivity(ctx context.Context, projectUUID, actorUserID, artifactID, action string, payload map[string]any) error {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "failed to encode activity payload"), err)
	}
	if err := r.store.Execute(ctx, storage.RelationalExec(
		`INSERT INTO activity_log (project_id, actor_user_id, artifact_type, artifact_id, action, payload)
		 VALUES ($1::uuid, $2::uuid, 'resource'::artifact_type, $3::uuid, $4, $5::jsonb)`,
		projectUUID,
		strings.TrimSpace(actorUserID),
		artifactID,
		action,
		string(bytes),
	)); err != nil {
		return wrapRepoError("log activity", err)
	}
	return nil
}

func defaultResourceContent(name, docType string) map[string]any {
	return map[string]any{
		"name":            name,
		"docType":         docType,
		"description":     "",
		"linkedArtifacts": []any{},
		"versions":        []any{},
		"notes":           "",
	}
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

func slugOrID(slug, id string) string {
	if strings.TrimSpace(slug) != "" {
		return strings.TrimSpace(slug)
	}
	return strings.TrimSpace(id)
}

func (r *repo) requireStore() error {
	if r == nil || r.store == nil {
		return apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "resources repository unavailable")
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
	return apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "failed to process resources data"), fmt.Errorf("%s: %w", action, err))
}
