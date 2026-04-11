package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/storage"
	"github.com/jackc/pgx/v5"
)

// Repo defines auth module persistence operations.
type Repo interface {
	UpdateUserName(ctx context.Context, userID, name string) error
}

type repo struct {
	store storage.RelationalStore
}

// NewRepo constructs a relational auth repository.
func NewRepo(store storage.RelationalStore) Repo {
	return &repo{store: store}
}

const queryUpdateAuthUserName = `
UPDATE users
SET name = $2, updated_at = NOW()
WHERE id = $1::uuid
RETURNING id::text
`

func (r *repo) UpdateUserName(ctx context.Context, userID, name string) error {
	if r == nil || r.store == nil {
		return apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "auth repository unavailable")
	}

	var touchedID string
	err := r.store.Execute(ctx, storage.RelationalQueryOne(queryUpdateAuthUserName,
		func(row storage.RowScanner) error {
			return row.Scan(&touchedID)
		},
		strings.TrimSpace(userID),
		strings.TrimSpace(name),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.WithCause(apperr.New(apperr.CodeNotFound, http.StatusNotFound, "user not found"), err)
		}
		return apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "failed to update user profile"), fmt.Errorf("update user name: %w", err))
	}

	return nil
}
