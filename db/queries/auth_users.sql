-- name: CreateAuthUser :one
INSERT INTO users (email, name, password_hash, is_email_verified)
VALUES ($1, $2, $3, $4)
RETURNING id, email, name, password_hash, is_email_verified, last_login_at, created_at, updated_at;

-- name: GetAuthUserByID :one
SELECT id, email, name, password_hash, is_email_verified, last_login_at, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetAuthUserByLogin :one
SELECT id, email, name, password_hash, is_email_verified, last_login_at, created_at, updated_at
FROM users
WHERE email = $1;

-- name: UpdateAuthUserPasswordHash :exec
UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1;

-- name: UpdateAuthUserEmailVerification :one
UPDATE users SET is_email_verified = $2, updated_at = NOW() WHERE id = $1 RETURNING id, email, name, password_hash, is_email_verified, last_login_at, created_at, updated_at;

