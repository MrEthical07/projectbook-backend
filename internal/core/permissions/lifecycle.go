package permissions

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/redis/go-redis/v9"

	"github.com/MrEthical07/superapi/internal/core/rbac"
	"github.com/MrEthical07/superapi/internal/core/storage"
)

const querySeedRolePermissionMask = `
INSERT INTO role_permissions (project_id, role, permission_mask, updated_at)
SELECT p.id, $1::project_role, $2::bigint, NOW()
FROM projects p
ON CONFLICT (project_id, role) DO NOTHING
`

const queryResyncNonCustomMemberMasks = `
UPDATE project_members pm
SET permission_mask = rp.permission_mask,
	updated_at = NOW()
FROM role_permissions rp
WHERE pm.project_id = rp.project_id
	AND pm.role = rp.role
	AND pm.is_custom = FALSE
	AND pm.permission_mask <> rp.permission_mask
RETURNING pm.user_id::text, pm.project_id::text
`

// AffectedScope identifies one user-project authorization scope changed by lifecycle updates.
type AffectedScope struct {
	UserID    string
	ProjectID string
}

// Lifecycle coordinates role-mask seeding and member-mask synchronization.
type Lifecycle struct {
	store       storage.RelationalStore
	redis       redis.UniversalClient
	invalidator TagInvalidator
}

// NewLifecycle constructs lifecycle operations for permission consistency.
func NewLifecycle(store storage.RelationalStore, redisClient redis.UniversalClient, invalidator TagInvalidator) (*Lifecycle, error) {
	if store == nil {
		return nil, fmt.Errorf("permissions lifecycle requires relational store")
	}
	return &Lifecycle{store: store, redis: redisClient, invalidator: invalidator}, nil
}

// RunStartup seeds role masks, resyncs non-custom members, and invalidates affected caches.
func (l *Lifecycle) RunStartup(ctx context.Context) error {
	if err := l.SeedRoleMasks(ctx); err != nil {
		return err
	}

	affectedScopes, err := l.ResyncNonCustomMemberMasks(ctx)
	if err != nil {
		return err
	}

	if len(affectedScopes) == 0 {
		return nil
	}

	if err := InvalidateResolverUserCache(ctx, l.redis, uniqueUserIDs(affectedScopes)); err != nil {
		return err
	}

	tags := uniquePermissionTags(affectedScopes)
	if len(tags) > 0 && l.invalidator != nil {
		if err := l.invalidator.BumpTags(ctx, tags); err != nil {
			return fmt.Errorf("invalidate permission tags: %w", err)
		}
	}

	return nil
}

// SeedRoleMasks ensures every project has canonical role mask rows.
func (l *Lifecycle) SeedRoleMasks(ctx context.Context) error {
	for _, role := range rbac.CanonicalRoles() {
		mask, ok := rbac.DefaultRoleMask(role)
		if !ok {
			return fmt.Errorf("default role mask missing for role %q", role)
		}
		if mask > math.MaxInt64 {
			return fmt.Errorf("default role mask overflow for role %q", role)
		}

		if err := l.store.Execute(ctx, storage.RelationalExec(querySeedRolePermissionMask, role, int64(mask))); err != nil {
			return fmt.Errorf("seed role mask for %q: %w", role, err)
		}
	}

	return nil
}

// ResyncNonCustomMemberMasks aligns non-custom member masks with role masks.
func (l *Lifecycle) ResyncNonCustomMemberMasks(ctx context.Context) ([]AffectedScope, error) {
	affectedScopes := make([]AffectedScope, 0, 32)
	err := l.store.Execute(ctx, storage.RelationalQueryMany(
		queryResyncNonCustomMemberMasks,
		func(row storage.RowScanner) error {
			var userID string
			var projectID string
			if err := row.Scan(&userID, &projectID); err != nil {
				return err
			}
			affectedScopes = append(affectedScopes, AffectedScope{UserID: userID, ProjectID: projectID})
			return nil
		},
	))
	if err != nil {
		return nil, fmt.Errorf("resync non-custom member masks: %w", err)
	}

	return uniqueAffectedScopes(affectedScopes), nil
}

func uniqueAffectedScopes(affectedScopes []AffectedScope) []AffectedScope {
	if len(affectedScopes) == 0 {
		return nil
	}

	unique := make([]AffectedScope, 0, len(affectedScopes))
	seen := make(map[string]struct{}, len(affectedScopes))
	for _, affectedScope := range affectedScopes {
		userID := strings.TrimSpace(affectedScope.UserID)
		projectID := strings.TrimSpace(affectedScope.ProjectID)
		if userID == "" || projectID == "" {
			continue
		}

		key := projectID + ":" + userID
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, AffectedScope{UserID: userID, ProjectID: projectID})
	}

	return unique
}

func uniqueUserIDs(affectedScopes []AffectedScope) []string {
	if len(affectedScopes) == 0 {
		return nil
	}

	unique := make([]string, 0, len(affectedScopes))
	seen := make(map[string]struct{}, len(affectedScopes))
	for _, affectedScope := range affectedScopes {
		userID := strings.TrimSpace(affectedScope.UserID)
		if userID == "" {
			continue
		}
		if _, exists := seen[userID]; exists {
			continue
		}
		seen[userID] = struct{}{}
		unique = append(unique, userID)
	}

	return unique
}

func uniquePermissionTags(affectedScopes []AffectedScope) []string {
	if len(affectedScopes) == 0 {
		return nil
	}

	tags := make([]string, 0, len(affectedScopes)*3)
	seen := make(map[string]struct{}, len(affectedScopes)*3)
	for _, affectedScope := range affectedScopes {
		for _, tag := range PermissionTags(affectedScope.UserID, affectedScope.ProjectID) {
			normalizedTag := strings.TrimSpace(tag)
			if normalizedTag == "" {
				continue
			}
			if _, exists := seen[normalizedTag]; exists {
				continue
			}
			seen[normalizedTag] = struct{}{}
			tags = append(tags, normalizedTag)
		}
	}

	return tags
}
