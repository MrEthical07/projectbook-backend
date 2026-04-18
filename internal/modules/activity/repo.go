package activity

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
)

type Repo interface {
	ListProjectActivity(ctx context.Context, projectID string, query listQuery) ([]ActivityItem, error)
}

type repo struct {
	store storage.RelationalStore
}

func NewRepo(store storage.RelationalStore) Repo {
	return &repo{store: store}
}

func (r *repo) ListProjectActivity(ctx context.Context, projectID string, query listQuery) ([]ActivityItem, error) {
	if err := r.requireStore(); err != nil {
		return nil, err
	}
	projectUUID, err := r.resolveProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	items := make([]ActivityItem, 0, query.Limit+1)
	err = r.store.Execute(ctx, storage.RelationalQueryMany(
		`SELECT
			a.id::text,
			COALESCE(u.name, 'Unknown'),
			a.action,
			COALESCE(a.payload->>'artifact', COALESCE(a.payload->>'artifactName', '')),
			COALESCE(a.payload->>'href', COALESCE(a.payload->>'artifactUrl', '')),
			a.created_at
		 FROM activity_log a
		 LEFT JOIN users u ON u.id = a.actor_user_id
		 WHERE a.project_id = $1::uuid
		 ORDER BY a.created_at DESC
		 OFFSET $2 LIMIT $3`,
		func(row storage.RowScanner) error {
			var id, userName, action, artifact, href string
			var createdAt time.Time
			if scanErr := row.Scan(&id, &userName, &action, &artifact, &href, &createdAt); scanErr != nil {
				return scanErr
			}
			items = append(items, ActivityItem{
				ID:       id,
				User:     userName,
				Initials: initialsFromName(userName),
				Action:   action,
				Artifact: artifact,
				Href:     href,
				At:       createdAt.UTC().Format(time.RFC3339),
			})
			return nil
		},
		projectUUID,
		query.Offset,
		query.Limit+1,
	))
	if err != nil {
		return nil, wrapRepoError("list project activity", err)
	}

	return items, nil
}

func (r *repo) resolveProjectID(ctx context.Context, projectID string) (string, error) {
	var id string
	err := r.store.Execute(ctx, storage.RelationalQueryOne(
		`SELECT id::text FROM projects WHERE id::text = $1 LIMIT 1`,
		func(row storage.RowScanner) error { return row.Scan(&id) },
		strings.TrimSpace(projectID),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperr.New(apperr.CodeNotFound, http.StatusNotFound, "project not found")
		}
		return "", wrapRepoError("resolve project", err)
	}
	return id, nil
}

func (r *repo) requireStore() error {
	if r == nil || r.store == nil {
		return apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "activity repository unavailable")
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
	return apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "failed to process activity data"), fmt.Errorf("%s: %w", action, err))
}
