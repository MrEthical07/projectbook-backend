package team

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/rbac"
	"github.com/MrEthical07/superapi/internal/core/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrTeamProjectNotFound    = errors.New("team project not found")
	ErrTeamInviteExists       = errors.New("team invite already exists")
	ErrTeamInviteNotFound     = errors.New("team invite not found")
	ErrTeamInviteNotPending   = errors.New("team invite not pending")
	ErrTeamMemberNotFound     = errors.New("team member not found")
	ErrTeamUserNotFound       = errors.New("team user not found")
	ErrTeamRoleNotFound       = errors.New("team role not found")
	ErrTeamOwnerImmutable     = errors.New("team owner immutable")
	ErrTeamMemberAlreadyExist = errors.New("team member already exists")
)

type teamProjectIdentity struct {
	UUID string
	Slug string
}

type createInviteInput struct {
	ProjectID       string
	Email           string
	Role            string
	InvitedByUserID string
	ExpiresAt       time.Time
}

type updateMemberPermissionsInput struct {
	ProjectID      string
	MemberID       string
	Role           string
	IsCustom       bool
	PermissionMask uint64
}

type updateRolePermissionsInput struct {
	ProjectID       string
	Role            string
	PermissionMask  uint64
	UpdatedByUserID string
}

// Repo defines team module persistence operations.
type Repo interface {
	ResolveProjectIdentity(ctx context.Context, projectID string) (teamProjectIdentity, error)
	ListMembersAndInvites(ctx context.Context, projectID string) (teamMembersResponse, error)
	ListRoles(ctx context.Context, projectID string) (teamRolesResponse, error)
	CreateInvite(ctx context.Context, input createInviteInput) (createInviteResponse, error)
	CancelInvite(ctx context.Context, projectID, email string) (cancelInviteResponse, error)
	UpdateMemberPermissions(ctx context.Context, input updateMemberPermissionsInput) (updateMemberPermissionsResponse, string, error)
	UpdateRolePermissions(ctx context.Context, input updateRolePermissionsInput) (updateRolePermissionsResponse, []string, error)
}

type repo struct {
	store storage.RelationalStore
}

// NewRepo constructs a relational team repository.
func NewRepo(store storage.RelationalStore) Repo {
	return &repo{store: store}
}

const queryResolveTeamProjectIdentity = `
SELECT id::text, slug
FROM projects
WHERE id::text = $1
LIMIT 1
`

const queryListTeamMembers = `
SELECT
	pm.id::text,
	u.name,
	u.email,
	pm.role::text,
	pm.status::text,
	COALESCE(to_char(pm.joined_at, 'YYYY-MM-DD'), '')
FROM project_members pm
JOIN users u ON u.id = pm.user_id
WHERE pm.project_id = $1::uuid
ORDER BY pm.created_at ASC
`

const queryListTeamInvites = `
SELECT
	i.email,
	i.assigned_role::text,
	COALESCE(to_char(i.sent_at, 'YYYY-MM-DD'), ''),
	i.status::text
FROM project_invites i
WHERE i.project_id = $1::uuid
	AND i.status = 'pending'
ORDER BY i.sent_at DESC
`

const queryListRolePermissionMasks = `
SELECT role::text, permission_mask
FROM role_permissions
WHERE project_id = $1::uuid
`

const queryListRoleMembers = `
SELECT
	pm.id::text,
	u.name,
	u.email,
	pm.role::text,
	pm.status::text,
	COALESCE(to_char(pm.joined_at, 'YYYY-MM-DD'), ''),
	pm.is_custom,
	pm.permission_mask
FROM project_members pm
JOIN users u ON u.id = pm.user_id
WHERE pm.project_id = $1::uuid
ORDER BY pm.created_at ASC
`

const queryIsMemberEmailPresent = `
SELECT EXISTS(
	SELECT 1
	FROM project_members pm
	JOIN users u ON u.id = pm.user_id
	WHERE pm.project_id = $1::uuid
		AND lower(u.email) = lower($2)
)
`

const queryIsRegisteredUserEmailPresent = `
SELECT EXISTS(
	SELECT 1
	FROM users u
	WHERE lower(u.email) = lower($1)
)
`

const queryGetPendingInviteIDByEmail = `
SELECT id::text
FROM project_invites
WHERE project_id = $1::uuid
	AND lower(email) = lower($2)
	AND status = 'pending'
LIMIT 1
`

const queryGetInviteStatusByEmail = `
SELECT status::text
FROM project_invites
WHERE project_id = $1::uuid
	AND lower(email) = lower($2)
ORDER BY sent_at DESC
LIMIT 1
`

const queryResolveInviterRole = `
SELECT role::text
FROM project_members
WHERE project_id = $1::uuid
	AND user_id = $2::uuid
	AND status = 'Active'
LIMIT 1
`

const queryGetRoleMask = `
SELECT permission_mask
FROM role_permissions
WHERE project_id = $1::uuid
	AND role = $2::project_role
LIMIT 1
`

const queryCreateInvite = `
INSERT INTO project_invites (
	project_id,
	email,
	assigned_role,
	permission_mask,
	invited_by_user_id,
	inviter_role,
	status,
	sent_at,
	expires_at
)
VALUES (
	$1::uuid,
	$2,
	$3::project_role,
	$4,
	$5::uuid,
	$6::project_role,
	'pending',
	NOW(),
	$7
)
RETURNING email, assigned_role::text, COALESCE(to_char(sent_at, 'YYYY-MM-DD'), ''), status::text
`

const queryCancelInvite = `
UPDATE project_invites
SET
	status = 'cancelled',
	cancelled_at = NOW(),
	updated_at = NOW()
WHERE project_id = $1::uuid
	AND lower(email) = lower($2)
	AND status = 'pending'
RETURNING email
`

const queryGetMemberForPermissionsUpdate = `
SELECT id::text, user_id::text, role::text
FROM project_members
WHERE project_id = $1::uuid
	AND id = $2::uuid
LIMIT 1
`

const queryUpdateMemberPermissions = `
UPDATE project_members
SET
	role = $3::project_role,
	is_custom = $4,
	permission_mask = $5,
	updated_at = NOW()
WHERE project_id = $1::uuid
	AND id = $2::uuid
RETURNING id::text, role::text, is_custom, permission_mask, user_id::text
`

const queryUpdateRolePermissions = `
UPDATE role_permissions
SET
	permission_mask = $3,
	updated_by_user_id = NULLIF($4, '')::uuid,
	updated_at = NOW()
WHERE project_id = $1::uuid
	AND role = $2::project_role
RETURNING role::text, permission_mask
`

const querySyncNonCustomMembersForRole = `
UPDATE project_members
SET
	permission_mask = $3,
	updated_at = NOW()
WHERE project_id = $1::uuid
	AND role = $2::project_role
	AND is_custom = FALSE
	AND permission_mask <> $3
RETURNING user_id::text
`

const queryCountCustomMembersForRole = `
SELECT COUNT(1)
FROM project_members
WHERE project_id = $1::uuid
	AND role = $2::project_role
	AND is_custom = TRUE
`

func (r *repo) ResolveProjectIdentity(ctx context.Context, projectID string) (teamProjectIdentity, error) {
	if err := r.requireStore(); err != nil {
		return teamProjectIdentity{}, err
	}

	var identity teamProjectIdentity
	err := r.store.Execute(ctx, storage.RelationalQueryOne(
		queryResolveTeamProjectIdentity,
		func(row storage.RowScanner) error {
			return row.Scan(&identity.UUID, &identity.Slug)
		},
		strings.TrimSpace(projectID),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return teamProjectIdentity{}, ErrTeamProjectNotFound
		}
		return teamProjectIdentity{}, wrapRepoError("resolve project identity", err)
	}

	return identity, nil
}

func (r *repo) ListMembersAndInvites(ctx context.Context, projectID string) (teamMembersResponse, error) {
	if err := r.requireStore(); err != nil {
		return teamMembersResponse{}, err
	}

	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return teamMembersResponse{}, err
	}

	response := teamMembersResponse{
		Members: make([]teamMember, 0, 32),
		Invites: make([]teamInvite, 0, 32),
	}

	err = r.store.Execute(ctx, storage.RelationalQueryMany(
		queryListTeamMembers,
		func(row storage.RowScanner) error {
			var member teamMember
			if err := row.Scan(&member.ID, &member.Name, &member.Email, &member.Role, &member.Status, &member.JoinedAt); err != nil {
				return err
			}
			response.Members = append(response.Members, member)
			return nil
		},
		identity.UUID,
	))
	if err != nil {
		return teamMembersResponse{}, wrapRepoError("list team members", err)
	}

	err = r.store.Execute(ctx, storage.RelationalQueryMany(
		queryListTeamInvites,
		func(row storage.RowScanner) error {
			var invite teamInvite
			if err := row.Scan(&invite.Email, &invite.Role, &invite.SentDate, &invite.Status); err != nil {
				return err
			}
			response.Invites = append(response.Invites, invite)
			return nil
		},
		identity.UUID,
	))
	if err != nil {
		return teamMembersResponse{}, wrapRepoError("list team invites", err)
	}

	return response, nil
}

func (r *repo) ListRoles(ctx context.Context, projectID string) (teamRolesResponse, error) {
	if err := r.requireStore(); err != nil {
		return teamRolesResponse{}, err
	}

	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return teamRolesResponse{}, err
	}

	response := teamRolesResponse{
		RolePermissionMasks: make(map[string]string, len(rbac.CanonicalRoles())),
		Members:             make([]teamRoleMember, 0, 32),
	}

	for _, role := range rbac.CanonicalRoles() {
		mask, ok := rbac.DefaultRoleMask(role)
		if ok {
			response.RolePermissionMasks[role] = fmt.Sprintf("%d", mask)
		}
	}

	err = r.store.Execute(ctx, storage.RelationalQueryMany(
		queryListRolePermissionMasks,
		func(row storage.RowScanner) error {
			var role string
			var mask int64
			if err := row.Scan(&role, &mask); err != nil {
				return err
			}
			if mask < 0 {
				return apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "role permission mask is invalid")
			}
			response.RolePermissionMasks[role] = fmt.Sprintf("%d", mask)
			return nil
		},
		identity.UUID,
	))
	if err != nil {
		return teamRolesResponse{}, wrapRepoError("list role permission masks", err)
	}

	err = r.store.Execute(ctx, storage.RelationalQueryMany(
		queryListRoleMembers,
		func(row storage.RowScanner) error {
			var member teamRoleMember
			var permissionMask int64
			if err := row.Scan(
				&member.ID,
				&member.Name,
				&member.Email,
				&member.Role,
				&member.Status,
				&member.JoinedAt,
				&member.IsCustom,
				&permissionMask,
			); err != nil {
				return err
			}
			if permissionMask < 0 {
				return apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "member permission mask is invalid")
			}
			member.PermissionMask = fmt.Sprintf("%d", permissionMask)
			response.Members = append(response.Members, member)
			return nil
		},
		identity.UUID,
	))
	if err != nil {
		return teamRolesResponse{}, wrapRepoError("list role members", err)
	}

	return response, nil
}

func (r *repo) CreateInvite(ctx context.Context, input createInviteInput) (createInviteResponse, error) {
	if err := r.requireStore(); err != nil {
		return createInviteResponse{}, err
	}

	identity, err := r.ResolveProjectIdentity(ctx, input.ProjectID)
	if err != nil {
		return createInviteResponse{}, err
	}

	userExists, err := r.userEmailExists(ctx, input.Email)
	if err != nil {
		return createInviteResponse{}, err
	}
	if !userExists {
		return createInviteResponse{}, ErrTeamUserNotFound
	}

	exists, err := r.memberEmailExists(ctx, identity.UUID, input.Email)
	if err != nil {
		return createInviteResponse{}, err
	}
	if exists {
		return createInviteResponse{}, ErrTeamMemberAlreadyExist
	}

	pendingInviteID, err := r.pendingInviteID(ctx, identity.UUID, input.Email)
	if err != nil {
		return createInviteResponse{}, err
	}
	if pendingInviteID != "" {
		return createInviteResponse{}, ErrTeamInviteExists
	}

	mask, err := r.roleMask(ctx, identity.UUID, input.Role)
	if err != nil {
		return createInviteResponse{}, err
	}

	inviterRole, err := r.inviterRole(ctx, identity.UUID, input.InvitedByUserID)
	if err != nil {
		return createInviteResponse{}, err
	}

	var response createInviteResponse
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		queryCreateInvite,
		func(row storage.RowScanner) error {
			return row.Scan(&response.Email, &response.Role, &response.SentDate, &response.Status)
		},
		identity.UUID,
		normalizeEmail(input.Email),
		input.Role,
		int64(mask),
		strings.TrimSpace(input.InvitedByUserID),
		inviterRole,
		input.ExpiresAt.UTC(),
	))
	if err != nil {
		if isUniqueViolation(err) {
			return createInviteResponse{}, ErrTeamInviteExists
		}
		return createInviteResponse{}, wrapRepoError("create invite", err)
	}

	return response, nil
}

func (r *repo) CancelInvite(ctx context.Context, projectID, email string) (cancelInviteResponse, error) {
	if err := r.requireStore(); err != nil {
		return cancelInviteResponse{}, err
	}

	identity, err := r.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return cancelInviteResponse{}, err
	}

	status, err := r.inviteStatus(ctx, identity.UUID, email)
	if err != nil {
		return cancelInviteResponse{}, err
	}
	if status == "" {
		return cancelInviteResponse{}, ErrTeamInviteNotFound
	}
	if !strings.EqualFold(status, "pending") {
		return cancelInviteResponse{}, ErrTeamInviteNotPending
	}

	var response cancelInviteResponse
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		queryCancelInvite,
		func(row storage.RowScanner) error {
			return row.Scan(&response.Email)
		},
		identity.UUID,
		normalizeEmail(email),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return cancelInviteResponse{}, ErrTeamInviteNotFound
		}
		return cancelInviteResponse{}, wrapRepoError("cancel invite", err)
	}

	return response, nil
}

func (r *repo) UpdateMemberPermissions(ctx context.Context, input updateMemberPermissionsInput) (updateMemberPermissionsResponse, string, error) {
	if err := r.requireStore(); err != nil {
		return updateMemberPermissionsResponse{}, "", err
	}

	identity, err := r.ResolveProjectIdentity(ctx, input.ProjectID)
	if err != nil {
		return updateMemberPermissionsResponse{}, "", err
	}

	current, err := r.memberByID(ctx, identity.UUID, input.MemberID)
	if err != nil {
		return updateMemberPermissionsResponse{}, "", err
	}
	if current.Role == rbac.RoleOwner {
		return updateMemberPermissionsResponse{}, "", ErrTeamOwnerImmutable
	}

	var response updateMemberPermissionsResponse
	var affectedUserID string
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		queryUpdateMemberPermissions,
		func(row storage.RowScanner) error {
			var mask int64
			if err := row.Scan(&response.MemberID, &response.Role, &response.IsCustom, &mask, &affectedUserID); err != nil {
				return err
			}
			if mask < 0 {
				return apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "member permission mask is invalid")
			}
			response.PermissionMask = uint64(mask)
			return nil
		},
		identity.UUID,
		strings.TrimSpace(input.MemberID),
		input.Role,
		input.IsCustom,
		int64(input.PermissionMask),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return updateMemberPermissionsResponse{}, "", ErrTeamMemberNotFound
		}
		return updateMemberPermissionsResponse{}, "", wrapRepoError("update member permissions", err)
	}

	return response, affectedUserID, nil
}

func (r *repo) UpdateRolePermissions(ctx context.Context, input updateRolePermissionsInput) (updateRolePermissionsResponse, []string, error) {
	if err := r.requireStore(); err != nil {
		return updateRolePermissionsResponse{}, nil, err
	}

	identity, err := r.ResolveProjectIdentity(ctx, input.ProjectID)
	if err != nil {
		return updateRolePermissionsResponse{}, nil, err
	}

	if strings.EqualFold(input.Role, rbac.RoleOwner) {
		return updateRolePermissionsResponse{}, nil, ErrTeamOwnerImmutable
	}

	if _, err := r.roleMask(ctx, identity.UUID, input.Role); err != nil {
		return updateRolePermissionsResponse{}, nil, err
	}

	var response updateRolePermissionsResponse
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		queryUpdateRolePermissions,
		func(row storage.RowScanner) error {
			var mask int64
			if err := row.Scan(&response.Role, &mask); err != nil {
				return err
			}
			if mask < 0 {
				return apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "role permission mask is invalid")
			}
			response.PermissionMask = uint64(mask)
			return nil
		},
		identity.UUID,
		input.Role,
		int64(input.PermissionMask),
		strings.TrimSpace(input.UpdatedByUserID),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return updateRolePermissionsResponse{}, nil, ErrTeamRoleNotFound
		}
		return updateRolePermissionsResponse{}, nil, wrapRepoError("update role permissions", err)
	}

	affectedUserIDs := make([]string, 0, 32)
	err = r.store.Execute(ctx, storage.RelationalQueryMany(
		querySyncNonCustomMembersForRole,
		func(row storage.RowScanner) error {
			var userID string
			if err := row.Scan(&userID); err != nil {
				return err
			}
			affectedUserIDs = append(affectedUserIDs, strings.TrimSpace(userID))
			return nil
		},
		identity.UUID,
		input.Role,
		int64(input.PermissionMask),
	))
	if err != nil {
		return updateRolePermissionsResponse{}, nil, wrapRepoError("sync non-custom members", err)
	}

	customUnaffected, err := r.customMembersCount(ctx, identity.UUID, input.Role)
	if err != nil {
		return updateRolePermissionsResponse{}, nil, err
	}
	response.CustomMembersUnaffected = customUnaffected

	return response, uniqueStrings(affectedUserIDs), nil
}

type memberRecord struct {
	ID     string
	UserID string
	Role   string
}

func (r *repo) memberByID(ctx context.Context, projectUUID, memberID string) (memberRecord, error) {
	var record memberRecord
	err := r.store.Execute(ctx, storage.RelationalQueryOne(
		queryGetMemberForPermissionsUpdate,
		func(row storage.RowScanner) error {
			return row.Scan(&record.ID, &record.UserID, &record.Role)
		},
		projectUUID,
		strings.TrimSpace(memberID),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return memberRecord{}, ErrTeamMemberNotFound
		}
		return memberRecord{}, wrapRepoError("get member by id", err)
	}
	return record, nil
}

func (r *repo) memberEmailExists(ctx context.Context, projectUUID, email string) (bool, error) {
	var exists bool
	err := r.store.Execute(ctx, storage.RelationalQueryOne(
		queryIsMemberEmailPresent,
		func(row storage.RowScanner) error {
			return row.Scan(&exists)
		},
		projectUUID,
		normalizeEmail(email),
	))
	if err != nil {
		return false, wrapRepoError("check member email", err)
	}
	return exists, nil
}

func (r *repo) userEmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.store.Execute(ctx, storage.RelationalQueryOne(
		queryIsRegisteredUserEmailPresent,
		func(row storage.RowScanner) error {
			return row.Scan(&exists)
		},
		normalizeEmail(email),
	))
	if err != nil {
		return false, wrapRepoError("check registered user email", err)
	}
	return exists, nil
}

func (r *repo) pendingInviteID(ctx context.Context, projectUUID, email string) (string, error) {
	var inviteID string
	err := r.store.Execute(ctx, storage.RelationalQueryOne(
		queryGetPendingInviteIDByEmail,
		func(row storage.RowScanner) error {
			return row.Scan(&inviteID)
		},
		projectUUID,
		normalizeEmail(email),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", wrapRepoError("get pending invite id", err)
	}
	return inviteID, nil
}

func (r *repo) inviteStatus(ctx context.Context, projectUUID, email string) (string, error) {
	var status string
	err := r.store.Execute(ctx, storage.RelationalQueryOne(
		queryGetInviteStatusByEmail,
		func(row storage.RowScanner) error {
			return row.Scan(&status)
		},
		projectUUID,
		normalizeEmail(email),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", wrapRepoError("get invite status", err)
	}
	return strings.TrimSpace(status), nil
}

func (r *repo) inviterRole(ctx context.Context, projectUUID, userID string) (string, error) {
	if strings.TrimSpace(userID) == "" {
		return rbac.RoleMember, nil
	}

	var role string
	err := r.store.Execute(ctx, storage.RelationalQueryOne(
		queryResolveInviterRole,
		func(row storage.RowScanner) error {
			return row.Scan(&role)
		},
		projectUUID,
		strings.TrimSpace(userID),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return rbac.RoleMember, nil
		}
		return "", wrapRepoError("resolve inviter role", err)
	}

	canonical, ok := canonicalRole(role)
	if !ok {
		return rbac.RoleMember, nil
	}
	return canonical, nil
}

func (r *repo) roleMask(ctx context.Context, projectUUID, role string) (uint64, error) {
	var mask int64
	err := r.store.Execute(ctx, storage.RelationalQueryOne(
		queryGetRoleMask,
		func(row storage.RowScanner) error {
			return row.Scan(&mask)
		},
		projectUUID,
		role,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrTeamRoleNotFound
		}
		return 0, wrapRepoError("get role mask", err)
	}
	if mask < 0 {
		return 0, apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "role permission mask is invalid")
	}
	return uint64(mask), nil
}

func (r *repo) customMembersCount(ctx context.Context, projectUUID, role string) (int, error) {
	var count int
	err := r.store.Execute(ctx, storage.RelationalQueryOne(
		queryCountCustomMembersForRole,
		func(row storage.RowScanner) error {
			return row.Scan(&count)
		},
		projectUUID,
		role,
	))
	if err != nil {
		return 0, wrapRepoError("count custom members", err)
	}
	return count, nil
}

func (r *repo) requireStore() error {
	if r == nil || r.store == nil {
		return apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "team repository unavailable")
	}
	return nil
}

func wrapRepoError(action string, err error) error {
	if err == nil {
		return nil
	}
	return apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "failed to process team data"), fmt.Errorf("%s: %w", action, err))
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, v := range values {
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	sort.Strings(out)
	return out
}
