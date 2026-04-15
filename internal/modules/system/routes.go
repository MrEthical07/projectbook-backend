package system

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/httpx"
	"github.com/MrEthical07/superapi/internal/core/policy"
	"github.com/MrEthical07/superapi/internal/core/ratelimit"
	"github.com/MrEthical07/superapi/internal/core/storage"
)

const (
	sessionContextTokenTTLSeconds   = 30 * 60
	sessionContextTokenVersion      = 1
	sessionContextSecretEnv         = "PROJECTBOOK_PERMISSION_CONTEXT_SECRET"
	sessionContextFallbackSecret    = "projectbook-dev-permission-context-secret"
	sessionContextSecretMinLength   = 32
	queryListUserProjectPermissions = `
SELECT
	p.id::text,
	p.slug,
	pm.role::text,
	pm.is_custom,
	pm.permission_mask,
	COALESCE(rp.permission_mask, pm.permission_mask) AS role_permission_mask,
	COALESCE(EXTRACT(EPOCH FROM pm.updated_at)::bigint, 0) AS member_updated_at_unix,
	COALESCE(EXTRACT(EPOCH FROM rp.updated_at)::bigint, 0) AS role_updated_at_unix
FROM project_members pm
JOIN projects p ON p.id = pm.project_id
LEFT JOIN role_permissions rp
	ON rp.project_id = pm.project_id
	AND rp.role = pm.role
WHERE pm.user_id::text = $1
	AND pm.status = 'Active'
ORDER BY p.id::text
`
)

type parseDurationRequest struct {
	Duration string `json:"duration"`
}

// Validate ensures duration string is provided for parsing.
func (r parseDurationRequest) Validate() error {
	if strings.TrimSpace(r.Duration) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "duration is required")
	}
	return nil
}

type parseDurationResponse struct {
	Duration     string `json:"duration"`
	Nanoseconds  int64  `json:"nanoseconds"`
	Milliseconds int64  `json:"milliseconds"`
}

// Register mounts system and auth demonstration routes.
//
// For policy behavior, see docs/policies.md.
func (m *Module) Register(r httpx.Router) error {
	r.Handle(http.MethodPost, "/system/parse-duration", httpx.Adapter(m.parseDuration), policy.RequireJSON())

	if limiter := m.runtime.Limiter(); limiter != nil {
		r.Handle(
			http.MethodGet,
			"/api/v1/system/whoami",
			httpx.Adapter(m.whoami),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.RateLimitWithKeyer(limiter, "system.whoami", m.rateRule, ratelimit.KeyByUserOrProjectOrTokenHash(16)),
		)
		r.Handle(
			http.MethodGet,
			"/api/v1/system/session-context",
			httpx.Adapter(m.sessionContext),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.RateLimitWithKeyer(limiter, "system.session_context", m.rateRule, ratelimit.KeyByUserOrProjectOrTokenHash(16)),
		)
		return nil
	}

	r.Handle(
		http.MethodGet,
		"/api/v1/system/whoami",
		httpx.Adapter(m.whoami),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
	)
	r.Handle(
		http.MethodGet,
		"/api/v1/system/session-context",
		httpx.Adapter(m.sessionContext),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
	)
	return nil
}

type whoamiResponse struct {
	UserID         string   `json:"user_id"`
	ProjectID      string   `json:"project_id,omitempty"`
	Role           string   `json:"role,omitempty"`
	PermissionMask uint64   `json:"permission_mask,omitempty"`
	Permissions    []string `json:"permissions,omitempty"`
}

type sessionContextProjectPermission struct {
	ProjectID      string `json:"project_id"`
	ProjectSlug    string `json:"project_slug,omitempty"`
	Role           string `json:"role,omitempty"`
	PermissionMask uint64 `json:"permission_mask"`
	IsCustom       bool   `json:"is_custom"`
	UpdatedAtUnix  int64  `json:"updated_at_unix,omitempty"`
}

type sessionContextResponse struct {
	UserID              string                            `json:"user_id"`
	BackendRole         string                            `json:"backend_role,omitempty"`
	ProjectPermissions  []sessionContextProjectPermission `json:"project_permissions"`
	SnapshotHash        string                            `json:"snapshot_hash"`
	ExpiresInSeconds    int                               `json:"expires_in_seconds"`
	ContextToken        string                            `json:"context_token"`
	ContextTokenExpUTC  string                            `json:"context_token_expires_utc,omitempty"`
	ContextTokenExpUnix int64                             `json:"context_token_expires_unix,omitempty"`
	ContextTokenVer     int                               `json:"context_token_version"`
}

type sessionContextTokenClaims struct {
	UserID             string                            `json:"user_id"`
	BackendRole        string                            `json:"backend_role,omitempty"`
	ProjectPermissions []sessionContextProjectPermission `json:"project_permissions"`
	SnapshotHash       string                            `json:"snapshot_hash,omitempty"`
	IssuedAtUnix       int64                             `json:"iat"`
	ExpiresAtUnix      int64                             `json:"exp"`
	Version            int                               `json:"v"`
}

func (m *Module) whoami(ctx *httpx.Context, _ httpx.NoBody) (whoamiResponse, error) {
	principal, ok := ctx.Auth()
	if !ok {
		return whoamiResponse{}, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	}

	return whoamiResponse{
		UserID:         principal.UserID,
		ProjectID:      principal.ProjectID,
		Role:           principal.Role,
		PermissionMask: principal.PermissionMask,
		Permissions:    append([]string(nil), principal.Permissions...),
	}, nil
}

func (m *Module) sessionContext(ctx *httpx.Context, _ httpx.NoBody) (sessionContextResponse, error) {
	principal, ok := ctx.Auth()
	if !ok {
		return sessionContextResponse{}, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	}

	projectPermissions, err := m.listUserProjectPermissions(ctx.Context(), principal.UserID)
	if err != nil {
		return sessionContextResponse{}, err
	}

	response := sessionContextResponse{
		UserID:             principal.UserID,
		BackendRole:        principal.Role,
		ProjectPermissions: projectPermissions,
		ExpiresInSeconds:   sessionContextTokenTTLSeconds,
	}

	hash, err := buildSessionContextHash(response)
	if err != nil {
		return sessionContextResponse{}, apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "session context hash failed"), err)
	}
	response.SnapshotHash = hash

	contextToken, contextTokenExpUnix, err := buildSessionContextToken(response, time.Now().UTC())
	if err != nil {
		return sessionContextResponse{}, apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "session context token failed"), err)
	}
	response.ContextToken = contextToken
	response.ContextTokenExpUnix = contextTokenExpUnix
	response.ContextTokenExpUTC = time.Unix(contextTokenExpUnix, 0).UTC().Format(time.RFC3339)
	response.ContextTokenVer = sessionContextTokenVersion

	return response, nil
}

func buildSessionContextToken(response sessionContextResponse, now time.Time) (string, int64, error) {
	issuedAtUnix := now.UTC().Unix()
	ttlSeconds := response.ExpiresInSeconds
	if ttlSeconds <= 0 {
		ttlSeconds = sessionContextTokenTTLSeconds
	}
	expiresAtUnix := issuedAtUnix + int64(ttlSeconds)

	claims := sessionContextTokenClaims{
		UserID:             strings.TrimSpace(response.UserID),
		BackendRole:        strings.TrimSpace(response.BackendRole),
		ProjectPermissions: response.ProjectPermissions,
		SnapshotHash:       strings.TrimSpace(response.SnapshotHash),
		IssuedAtUnix:       issuedAtUnix,
		ExpiresAtUnix:      expiresAtUnix,
		Version:            sessionContextTokenVersion,
	}

	headerBytes, err := json.Marshal(map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	})
	if err != nil {
		return "", 0, err
	}
	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return "", 0, err
	}

	encodedHeader := base64.RawURLEncoding.EncodeToString(headerBytes)
	encodedClaims := base64.RawURLEncoding.EncodeToString(claimsBytes)
	unsigned := encodedHeader + "." + encodedClaims
	signature, err := signSessionContextToken(unsigned)
	if err != nil {
		return "", 0, err
	}

	return unsigned + "." + signature, expiresAtUnix, nil
}

func signSessionContextToken(unsignedToken string) (string, error) {
	secret, err := resolveSessionContextSigningSecret()
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256([]byte(unsignedToken + "." + secret))
	return hex.EncodeToString(hash[:]), nil
}

func resolveSessionContextSigningSecret() (string, error) {
	configured := strings.TrimSpace(os.Getenv(sessionContextSecretEnv))
	if len(configured) >= sessionContextSecretMinLength && configured != sessionContextFallbackSecret {
		return configured, nil
	}
	if isProductionSessionContextRuntime() {
		return "", fmt.Errorf("%s must be configured with a non-default secret in production", sessionContextSecretEnv)
	}
	return sessionContextFallbackSecret, nil
}

func isProductionSessionContextRuntime() bool {
	env := strings.TrimSpace(os.Getenv("APP_ENV"))
	if strings.EqualFold(env, "prod") || strings.EqualFold(env, "production") {
		return true
	}
	profile := strings.TrimSpace(os.Getenv("APP_PROFILE"))
	return strings.EqualFold(profile, "prod")
}

func (m *Module) listUserProjectPermissions(ctx context.Context, userID string) ([]sessionContextProjectPermission, error) {
	store := m.runtime.RelationalStore()
	if store == nil {
		return []sessionContextProjectPermission{}, nil
	}

	projectPermissions := make([]sessionContextProjectPermission, 0, 8)
	err := store.Execute(ctx, storage.RelationalQueryMany(
		queryListUserProjectPermissions,
		func(row storage.RowScanner) error {
			var projectID string
			var projectSlug string
			var role string
			var isCustom bool
			var memberMask int64
			var roleMask int64
			var memberUpdatedAtUnix int64
			var roleUpdatedAtUnix int64
			if err := row.Scan(
				&projectID,
				&projectSlug,
				&role,
				&isCustom,
				&memberMask,
				&roleMask,
				&memberUpdatedAtUnix,
				&roleUpdatedAtUnix,
			); err != nil {
				return err
			}

			if memberMask < 0 || roleMask < 0 {
				return apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "permission mask unavailable")
			}

			effectiveMask := roleMask
			if isCustom {
				effectiveMask = memberMask
			}

			updatedAtUnix := memberUpdatedAtUnix
			if roleUpdatedAtUnix > updatedAtUnix {
				updatedAtUnix = roleUpdatedAtUnix
			}

			projectPermissions = append(projectPermissions, sessionContextProjectPermission{
				ProjectID:      projectID,
				ProjectSlug:    projectSlug,
				Role:           role,
				PermissionMask: uint64(effectiveMask),
				IsCustom:       isCustom,
				UpdatedAtUnix:  updatedAtUnix,
			})
			return nil
		},
		userID,
	))
	if err != nil {
		if apiErr, ok := apperr.AsAppError(err); ok {
			return nil, apiErr
		}
		return nil, apperr.WithCause(apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "permission matrix unavailable"), err)
	}

	return projectPermissions, nil
}

func buildSessionContextHash(response sessionContextResponse) (string, error) {
	hashInput := struct {
		UserID             string                            `json:"user_id"`
		BackendRole        string                            `json:"backend_role,omitempty"`
		ProjectPermissions []sessionContextProjectPermission `json:"project_permissions"`
	}{
		UserID:             response.UserID,
		BackendRole:        response.BackendRole,
		ProjectPermissions: response.ProjectPermissions,
	}

	payload, err := json.Marshal(hashInput)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(payload)
	return hex.EncodeToString(hash[:]), nil
}

func (m *Module) parseDuration(_ *httpx.Context, req parseDurationRequest) (parseDurationResponse, error) {
	d, err := time.ParseDuration(req.Duration)
	if err != nil {
		return parseDurationResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "duration must be a valid Go duration string")
	}

	return parseDurationResponse{
		Duration:     d.String(),
		Nanoseconds:  d.Nanoseconds(),
		Milliseconds: d.Milliseconds(),
	}, nil
}
