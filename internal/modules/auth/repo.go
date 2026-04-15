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
	LookupUserByEmail(ctx context.Context, email string) (UserLookup, bool, error)
	LookupUserByID(ctx context.Context, userID string) (UserLookup, bool, error)
}

// UserLookup is the minimal auth user projection needed by resend flows.
type UserLookup struct {
	ID              string
	Email           string
	IsEmailVerified bool
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

const queryLookupAuthUserByEmail = `
SELECT
	id::text,
	email,
	is_email_verified
FROM users
WHERE email = $1
`

const queryLookupAuthUserByID = `
SELECT
	id::text,
	email,
	is_email_verified
FROM users
WHERE id = $1::uuid
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

func (r *repo) LookupUserByEmail(ctx context.Context, email string) (UserLookup, bool, error) {
	if r == nil || r.store == nil {
		return UserLookup{}, false, apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "auth repository unavailable")
	}

	var user UserLookup
	err := r.store.Execute(ctx, storage.RelationalQueryOne(queryLookupAuthUserByEmail,
		func(row storage.RowScanner) error {
			return row.Scan(&user.ID, &user.Email, &user.IsEmailVerified)
		},
		normalizeEmail(email),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserLookup{}, false, nil
		}
		return UserLookup{}, false, apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "failed to read user status"), fmt.Errorf("lookup user by email: %w", err))
	}

	return user, true, nil
}

func (r *repo) LookupUserByID(ctx context.Context, userID string) (UserLookup, bool, error) {
	if r == nil || r.store == nil {
		return UserLookup{}, false, apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "auth repository unavailable")
	}

	var user UserLookup
	err := r.store.Execute(ctx, storage.RelationalQueryOne(queryLookupAuthUserByID,
		func(row storage.RowScanner) error {
			return row.Scan(&user.ID, &user.Email, &user.IsEmailVerified)
		},
		strings.TrimSpace(userID),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserLookup{}, false, nil
		}
		return UserLookup{}, false, apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "failed to read user status"), fmt.Errorf("lookup user by id: %w", err))
	}

	return user, true, nil
}
