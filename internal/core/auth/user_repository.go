package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/MrEthical07/superapi/internal/core/storage"
	"github.com/jackc/pgx/v5"
)

var ErrAuthUserNotFound = errors.New("auth user not found")

// StoredUser is the storage-layer projection used by the auth repository.
type StoredUser struct {
	ID             string
	Email          string
	Name           string
	PasswordHash   string
	AccountVersion uint32
	EmailVerified  bool
}

// CreateStoredUserInput is the repository input model for creating auth users.
type CreateStoredUserInput struct {
	Identifier     string
	Name           string
	PasswordHash   string
	AccountVersion uint32
	EmailVerified  bool
}

// UserRepository defines domain-level auth user persistence operations.
type UserRepository interface {
	GetByIdentifier(ctx context.Context, identifier string) (StoredUser, error)
	GetByID(ctx context.Context, userID string) (StoredUser, error)
	UpdatePasswordHash(ctx context.Context, userID, newHash string) error
	Create(ctx context.Context, input CreateStoredUserInput) (StoredUser, error)
	UpdateStatus(ctx context.Context, userID string, status string) (StoredUser, error)
}

type relationalUserRepository struct {
	store storage.RelationalStore
}

// NewRelationalUserRepository creates an auth repository over a relational store.
func NewRelationalUserRepository(store storage.RelationalStore) UserRepository {
	if store == nil {
		return nil
	}
	return &relationalUserRepository{store: store}
}

const (
	queryAuthUserByIdentifier = `
SELECT
	id::text,
	email,
	name,
	password_hash,
	account_version,
	is_email_verified
FROM users
WHERE email = $1
`

	queryAuthUserByID = `
SELECT
	id::text,
	email,
	name,
	password_hash,
	account_version,
	is_email_verified
FROM users
WHERE id = $1::uuid
`

	queryUpdatePasswordHash = `
UPDATE users
SET password_hash = $2,
	account_version = account_version + 1,
	updated_at = NOW()
WHERE id = $1::uuid
RETURNING id::text
`

	queryCreateAuthUser = `
INSERT INTO users (email, name, password_hash, is_email_verified, account_version)
VALUES ($1, $2, $3, $4, $5)
RETURNING
	id::text,
	email,
	name,
	password_hash,
	account_version,
	is_email_verified
`

	queryUpdateEmailVerification = `
UPDATE users
SET is_email_verified = $2,
	account_version = account_version + 1,
	updated_at = NOW()
WHERE id = $1::uuid
RETURNING
	id::text,
	email,
	name,
	password_hash,
	account_version,
	is_email_verified
`
)

func (r *relationalUserRepository) GetByIdentifier(ctx context.Context, identifier string) (StoredUser, error) {
	var user StoredUser
	err := r.store.Execute(ctx, storage.RelationalQueryOne(queryAuthUserByIdentifier,
		func(row storage.RowScanner) error {
			return row.Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.AccountVersion, &user.EmailVerified)
		},
		strings.TrimSpace(identifier),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return StoredUser{}, ErrAuthUserNotFound
		}
		return StoredUser{}, fmt.Errorf("get user by identifier: %w", err)
	}
	return user, nil
}

func (r *relationalUserRepository) GetByID(ctx context.Context, userID string) (StoredUser, error) {
	var user StoredUser
	err := r.store.Execute(ctx, storage.RelationalQueryOne(queryAuthUserByID,
		func(row storage.RowScanner) error {
			return row.Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.AccountVersion, &user.EmailVerified)
		},
		strings.TrimSpace(userID),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return StoredUser{}, ErrAuthUserNotFound
		}
		return StoredUser{}, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func (r *relationalUserRepository) UpdatePasswordHash(ctx context.Context, userID, newHash string) error {
	var touchedID string
	err := r.store.Execute(ctx, storage.RelationalQueryOne(queryUpdatePasswordHash,
		func(row storage.RowScanner) error {
			return row.Scan(&touchedID)
		},
		strings.TrimSpace(userID),
		newHash,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrAuthUserNotFound
		}
		return fmt.Errorf("update password hash: %w", err)
	}
	return nil
}

func (r *relationalUserRepository) Create(ctx context.Context, input CreateStoredUserInput) (StoredUser, error) {
	var user StoredUser
	err := r.store.Execute(ctx, storage.RelationalQueryOne(queryCreateAuthUser,
		func(row storage.RowScanner) error {
			return row.Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.AccountVersion, &user.EmailVerified)
		},
		strings.TrimSpace(input.Identifier),
		strings.TrimSpace(input.Name),
		input.PasswordHash,
		input.EmailVerified,
		coerceAccountVersion(input.AccountVersion),
	))
	if err != nil {
		return StoredUser{}, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

func (r *relationalUserRepository) UpdateStatus(ctx context.Context, userID string, status string) (StoredUser, error) {
	var user StoredUser
	err := r.store.Execute(ctx, storage.RelationalQueryOne(queryUpdateEmailVerification,
		func(row storage.RowScanner) error {
			return row.Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.AccountVersion, &user.EmailVerified)
		},
		strings.TrimSpace(userID),
		statusToVerifiedFlag(status),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return StoredUser{}, ErrAuthUserNotFound
		}
		return StoredUser{}, fmt.Errorf("update account status: %w", err)
	}
	return user, nil
}

func statusToVerifiedFlag(status string) bool {
	return strings.EqualFold(strings.TrimSpace(status), "active")
}

func coerceAccountVersion(version uint32) int64 {
	if version == 0 {
		return 1
	}
	return int64(version)
}
