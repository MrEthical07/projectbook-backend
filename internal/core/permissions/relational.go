package permissions

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MrEthical07/superapi/internal/core/storage"
	"github.com/jackc/pgx/v5"
)

const queryResolveProjectMembership = `
SELECT
	p.id::text,
	pm.role::text,
	pm.is_custom,
	pm.permission_mask,
	rp.permission_mask,
	(rp.permission_mask IS NOT NULL) AS has_role_mask,
	COALESCE(EXTRACT(EPOCH FROM pm.updated_at)::bigint, 0),
	COALESCE(EXTRACT(EPOCH FROM rp.updated_at)::bigint, 0)
FROM projects p
JOIN project_members pm
	ON pm.project_id = p.id
LEFT JOIN role_permissions rp
	ON rp.project_id = pm.project_id
	AND rp.role = pm.role
WHERE (p.id::text = $1 OR p.slug = $1)
	AND pm.user_id = $2::uuid
	AND pm.status = 'Active'
LIMIT 1
`

// RelationalResolver resolves effective permission masks from relational storage.
type RelationalResolver struct {
	store        storage.RelationalStore
	queryTimeout time.Duration
}

// NewRelationalResolver constructs a relational fallback resolver.
func NewRelationalResolver(store storage.RelationalStore, queryTimeout time.Duration) (*RelationalResolver, error) {
	if store == nil {
		return nil, fmt.Errorf("relational resolver requires relational store")
	}
	if queryTimeout <= 0 {
		queryTimeout = 750 * time.Millisecond
	}
	return &RelationalResolver{store: store, queryTimeout: queryTimeout}, nil
}

// Resolve fetches project membership and computes the effective permission mask.
func (r *RelationalResolver) Resolve(ctx context.Context, userID, projectID string) (Resolution, error) {
	userID = strings.TrimSpace(userID)
	projectID = strings.TrimSpace(projectID)
	if userID == "" || projectID == "" {
		return Resolution{}, fmt.Errorf("%w: missing user or project id", ErrMaskInconsistent)
	}

	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	var role string
	var resolvedProjectID string
	var isCustom bool
	var memberMask int64
	var roleMask int64
	var hasRoleMask bool
	var memberUpdated int64
	var roleUpdated int64

	err := r.store.Execute(queryCtx, storage.RelationalQueryOne(
		queryResolveProjectMembership,
		func(row storage.RowScanner) error {
			return row.Scan(&resolvedProjectID, &role, &isCustom, &memberMask, &roleMask, &hasRoleMask, &memberUpdated, &roleUpdated)
		},
		projectID,
		userID,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Resolution{}, fmt.Errorf("%w: project membership not found", ErrMembershipNotFound)
		}
		return Resolution{}, fmt.Errorf("%w: resolve project membership: %w", ErrDependencyFailure, err)
	}

	if memberMask < 0 {
		return Resolution{}, fmt.Errorf("%w: member permission mask is negative", ErrMaskInconsistent)
	}
	if !isCustom {
		if !hasRoleMask {
			return Resolution{}, fmt.Errorf("%w: role permission mask missing", ErrMaskInconsistent)
		}
		if roleMask < 0 {
			return Resolution{}, fmt.Errorf("%w: role permission mask is negative", ErrMaskInconsistent)
		}
	}

	if strings.TrimSpace(resolvedProjectID) == "" {
		return Resolution{}, fmt.Errorf("%w: resolved project id is empty", ErrMaskInconsistent)
	}

	effectiveMask := uint64(memberMask)
	updatedAtUnix := memberUpdated
	if !isCustom {
		effectiveMask = uint64(roleMask)
		if roleUpdated > updatedAtUnix {
			updatedAtUnix = roleUpdated
		}
	}
	if updatedAtUnix <= 0 {
		updatedAtUnix = time.Now().Unix()
	}

	return Resolution{
		UserID:        userID,
		ProjectID:     strings.TrimSpace(resolvedProjectID),
		Role:          role,
		Mask:          effectiveMask,
		IsCustom:      isCustom,
		UpdatedAtUnix: updatedAtUnix,
	}, nil
}
