package team

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/permissions"
	"github.com/MrEthical07/superapi/internal/core/storage"
)

const inviteTTL = 7 * 24 * time.Hour

const (
	codeInviteExists      apperr.Code = "invite_exists"
	codeInviteNotFound    apperr.Code = "invite_not_found"
	codeMemberExists      apperr.Code = "member_exists"
	codeUserNotFound      apperr.Code = "user_not_found"
	codeRoleConfigMissing apperr.Code = "role_config_missing"
	codeAlreadyCancelled  apperr.Code = "already_cancelled"
)

// Service defines team module business workflows.
type Service interface {
	ListMembers(ctx context.Context, projectID string) (teamMembersResponse, error)
	ListRoles(ctx context.Context, projectID string) (teamRolesResponse, error)
	CreateInvite(ctx context.Context, projectID, actorUserID string, req createInviteRequest) (createInviteResponse, error)
	BatchInvites(ctx context.Context, projectID, actorUserID string, req batchInviteRequest) (batchInviteResponse, bool, error)
	CancelInvite(ctx context.Context, projectID, email string) (cancelInviteResponse, error)
	UpdateMemberPermissions(ctx context.Context, projectID, memberID string, req updateMemberPermissionsRequest) (updateMemberPermissionsResponse, error)
	UpdateRolePermissions(ctx context.Context, projectID, pathRole, actorUserID string, req updateRolePermissionsRequest) (updateRolePermissionsResponse, error)
}

type service struct {
	store       storage.RelationalStore
	repo        Repo
	redis       redis.UniversalClient
	invalidator permissions.TagInvalidator
}

// NewService constructs team business workflows.
func NewService(store storage.RelationalStore, repo Repo, redisClient redis.UniversalClient, invalidator permissions.TagInvalidator) Service {
	return &service{store: store, repo: repo, redis: redisClient, invalidator: invalidator}
}

func (s *service) ListMembers(ctx context.Context, projectID string) (teamMembersResponse, error) {
	if strings.TrimSpace(projectID) == "" {
		return teamMembersResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "projectId is required")
	}
	response, err := s.repo.ListMembersAndInvites(ctx, projectID)
	if err != nil {
		return teamMembersResponse{}, mapTeamRepoError(err)
	}
	return response, nil
}

func (s *service) ListRoles(ctx context.Context, projectID string) (teamRolesResponse, error) {
	if strings.TrimSpace(projectID) == "" {
		return teamRolesResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "projectId is required")
	}
	response, err := s.repo.ListRoles(ctx, projectID)
	if err != nil {
		return teamRolesResponse{}, mapTeamRepoError(err)
	}
	return response, nil
}

func (s *service) CreateInvite(ctx context.Context, projectID, actorUserID string, req createInviteRequest) (createInviteResponse, error) {
	if strings.TrimSpace(projectID) == "" {
		return createInviteResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "projectId is required")
	}
	if strings.TrimSpace(actorUserID) == "" {
		return createInviteResponse{}, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	}
	if err := s.requireStore(); err != nil {
		return createInviteResponse{}, err
	}

	role, ok := canonicalRole(req.Role)
	if !ok {
		return createInviteResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "role is invalid")
	}

	input := createInviteInput{
		ProjectID:       projectID,
		Email:           normalizeEmail(req.Email),
		Role:            role,
		InvitedByUserID: strings.TrimSpace(actorUserID),
		ExpiresAt:       time.Now().UTC().Add(inviteTTL),
	}

	var response createInviteResponse
	err := s.store.WithTx(ctx, func(txCtx context.Context) error {
		created, err := s.repo.CreateInvite(txCtx, input)
		if err != nil {
			return err
		}
		response = created
		return nil
	})
	if err != nil {
		return createInviteResponse{}, mapTeamRepoError(err)
	}

	return response, nil
}

func (s *service) BatchInvites(ctx context.Context, projectID, actorUserID string, req batchInviteRequest) (batchInviteResponse, bool, error) {
	if strings.TrimSpace(projectID) == "" {
		return batchInviteResponse{}, false, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "projectId is required")
	}
	if strings.TrimSpace(actorUserID) == "" {
		return batchInviteResponse{}, false, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	}
	if err := s.requireStore(); err != nil {
		return batchInviteResponse{}, false, err
	}

	projectIdentity, err := s.repo.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return batchInviteResponse{}, false, mapTeamRepoError(err)
	}

	response := batchInviteResponse{
		ProjectID: projectIdentity.UUID,
		Invited:   make([]batchInviteSuccess, 0, len(req.Invites)),
		Failed:    make([]batchInviteFailure, 0),
	}

	for _, invite := range req.Invites {
		normalizedEmail := normalizeEmail(invite.Email)
		role, roleOK := canonicalRole(invite.Role)
		if normalizedEmail == "" || !isValidEmail(normalizedEmail) {
			response.Failed = append(response.Failed, batchInviteFailure{Email: normalizedEmail, Role: invite.Role, Code: string(apperr.CodeBadRequest), Message: "email is invalid"})
			continue
		}
		if !roleOK {
			response.Failed = append(response.Failed, batchInviteFailure{Email: normalizedEmail, Role: invite.Role, Code: string(apperr.CodeBadRequest), Message: "role is invalid"})
			continue
		}

		input := createInviteInput{
			ProjectID:       projectIdentity.UUID,
			Email:           normalizedEmail,
			Role:            role,
			InvitedByUserID: strings.TrimSpace(actorUserID),
			ExpiresAt:       time.Now().UTC().Add(inviteTTL),
		}

		err := s.store.WithTx(ctx, func(txCtx context.Context) error {
			_, createErr := s.repo.CreateInvite(txCtx, input)
			return createErr
		})
		if err != nil {
			mapped := mapTeamRepoError(err)
			ae, ok := apperr.AsAppError(mapped)
			if ok {
				if ae.Code == apperr.CodeInternal || ae.Code == apperr.CodeDependencyFailure {
					return batchInviteResponse{}, false, mapped
				}
				response.Failed = append(response.Failed, batchInviteFailure{Email: normalizedEmail, Role: role, Code: string(ae.Code), Message: ae.Message})
				continue
			}
			return batchInviteResponse{}, false, mapped
		}

		response.Invited = append(response.Invited, batchInviteSuccess{Email: normalizedEmail, Role: role})
	}

	return response, len(response.Failed) > 0, nil
}

func (s *service) CancelInvite(ctx context.Context, projectID, email string) (cancelInviteResponse, error) {
	if strings.TrimSpace(projectID) == "" {
		return cancelInviteResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "projectId is required")
	}
	if normalizeEmail(email) == "" || !isValidEmail(normalizeEmail(email)) {
		return cancelInviteResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "email is invalid")
	}
	if err := s.requireStore(); err != nil {
		return cancelInviteResponse{}, err
	}

	var response cancelInviteResponse
	err := s.store.WithTx(ctx, func(txCtx context.Context) error {
		cancelled, cancelErr := s.repo.CancelInvite(txCtx, projectID, email)
		if cancelErr != nil {
			return cancelErr
		}
		response = cancelled
		return nil
	})
	if err != nil {
		return cancelInviteResponse{}, mapTeamRepoError(err)
	}

	return response, nil
}

func (s *service) UpdateMemberPermissions(ctx context.Context, projectID, memberID string, req updateMemberPermissionsRequest) (updateMemberPermissionsResponse, error) {
	if strings.TrimSpace(projectID) == "" {
		return updateMemberPermissionsResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "projectId is required")
	}
	if strings.TrimSpace(memberID) == "" {
		return updateMemberPermissionsResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "memberId is required")
	}
	if err := s.requireStore(); err != nil {
		return updateMemberPermissionsResponse{}, err
	}

	role, ok := canonicalRole(req.Role)
	if !ok {
		return updateMemberPermissionsResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "role is invalid")
	}
	if role == rbacRoleOwner() {
		return updateMemberPermissionsResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "owner role cannot be assigned")
	}
	if err := validatePermissionMaskDependencies(req.PermissionMask); err != nil {
		return updateMemberPermissionsResponse{}, err
	}

	roleMask, err := s.resolveRoleMask(ctx, projectID, role)
	if err != nil {
		return updateMemberPermissionsResponse{}, err
	}

	isCustom := req.IsCustom
	mask := req.PermissionMask
	if !req.IsCustom {
		if mask != roleMask {
			return updateMemberPermissionsResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "permissionMask must match role mask when isCustom is false")
		}
	}
	if req.IsCustom && mask == roleMask {
		isCustom = false
	}

	input := updateMemberPermissionsInput{
		ProjectID:      projectID,
		MemberID:       memberID,
		Role:           role,
		IsCustom:       isCustom,
		PermissionMask: mask,
	}

	var response updateMemberPermissionsResponse
	var affectedUserID string
	err = s.store.WithTx(ctx, func(txCtx context.Context) error {
		updated, userID, updateErr := s.repo.UpdateMemberPermissions(txCtx, input)
		if updateErr != nil {
			return updateErr
		}
		response = updated
		affectedUserID = strings.TrimSpace(userID)
		return nil
	})
	if err != nil {
		return updateMemberPermissionsResponse{}, mapTeamRepoError(err)
	}

	s.invalidatePermissionScopes(ctx, projectID, []string{affectedUserID})
	return response, nil
}

func (s *service) UpdateRolePermissions(ctx context.Context, projectID, pathRole, actorUserID string, req updateRolePermissionsRequest) (updateRolePermissionsResponse, error) {
	if strings.TrimSpace(projectID) == "" {
		return updateRolePermissionsResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "projectId is required")
	}
	if strings.TrimSpace(pathRole) == "" {
		return updateRolePermissionsResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "role path parameter is required")
	}
	if strings.TrimSpace(actorUserID) == "" {
		return updateRolePermissionsResponse{}, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	}
	if err := s.requireStore(); err != nil {
		return updateRolePermissionsResponse{}, err
	}

	bodyRole, ok := canonicalRole(req.Role)
	if !ok {
		return updateRolePermissionsResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "role is invalid")
	}
	if !strings.EqualFold(bodyRole, pathRole) {
		return updateRolePermissionsResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "role must match path role")
	}
	if strings.EqualFold(pathRole, rbacRoleOwner()) {
		return updateRolePermissionsResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "owner role cannot be modified")
	}
	if err := validatePermissionMaskDependencies(req.PermissionMask); err != nil {
		return updateRolePermissionsResponse{}, err
	}

	input := updateRolePermissionsInput{
		ProjectID:       projectID,
		Role:            pathRole,
		PermissionMask:  req.PermissionMask,
		UpdatedByUserID: actorUserID,
	}

	var response updateRolePermissionsResponse
	var affectedUserIDs []string
	err := s.store.WithTx(ctx, func(txCtx context.Context) error {
		updated, userIDs, updateErr := s.repo.UpdateRolePermissions(txCtx, input)
		if updateErr != nil {
			return updateErr
		}
		response = updated
		affectedUserIDs = append([]string(nil), userIDs...)
		return nil
	})
	if err != nil {
		return updateRolePermissionsResponse{}, mapTeamRepoError(err)
	}

	s.invalidatePermissionScopes(ctx, projectID, affectedUserIDs)
	return response, nil
}

func (s *service) resolveRoleMask(ctx context.Context, projectID, role string) (uint64, error) {
	identity, err := s.repo.ResolveProjectIdentity(ctx, projectID)
	if err != nil {
		return 0, mapTeamRepoError(err)
	}
	response, err := s.repo.ListRoles(ctx, identity.UUID)
	if err != nil {
		return 0, mapTeamRepoError(err)
	}
	maskRaw, ok := response.RolePermissionMasks[role]
	if !ok {
		return 0, apperr.New(codeRoleConfigMissing, http.StatusNotFound, "role permissions not configured for project")
	}
	var mask uint64
	_, parseErr := fmt.Sscanf(maskRaw, "%d", &mask)
	if parseErr != nil {
		return 0, apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "role permission mask is invalid")
	}
	return mask, nil
}

func (s *service) invalidatePermissionScopes(ctx context.Context, projectID string, userIDs []string) {
	normalizedUsers := uniqueStrings(userIDs)
	if len(normalizedUsers) == 0 {
		return
	}

	_ = permissions.InvalidateResolverUserCache(ctx, s.redis, normalizedUsers)
	if s.invalidator == nil {
		return
	}

	tags := make([]string, 0, len(normalizedUsers)*3)
	for _, userID := range normalizedUsers {
		tags = append(tags, permissions.PermissionTags(userID, projectID)...)
	}
	_ = s.invalidator.BumpTags(ctx, tags)
}

func (s *service) requireStore() error {
	if s == nil || s.store == nil {
		return apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "team service unavailable")
	}
	return nil
}

func mapTeamRepoError(err error) error {
	if err == nil {
		return nil
	}
	if ae, ok := apperr.AsAppError(err); ok {
		return ae
	}

	switch {
	case errors.Is(err, ErrTeamProjectNotFound):
		return apperr.New(apperr.CodeNotFound, http.StatusNotFound, "project not found")
	case errors.Is(err, ErrTeamInviteExists):
		return apperr.New(codeInviteExists, http.StatusConflict, "pending invite already exists for this email")
	case errors.Is(err, ErrTeamInviteNotFound):
		return apperr.New(codeInviteNotFound, http.StatusNotFound, "pending invite not found")
	case errors.Is(err, ErrTeamInviteNotPending):
		return apperr.New(codeAlreadyCancelled, http.StatusBadRequest, "only pending invites can be cancelled")
	case errors.Is(err, ErrTeamMemberNotFound):
		return apperr.New(apperr.CodeNotFound, http.StatusNotFound, "member not found")
	case errors.Is(err, ErrTeamUserNotFound):
		return apperr.New(codeUserNotFound, http.StatusNotFound, "user not found for invite email")
	case errors.Is(err, ErrTeamRoleNotFound):
		return apperr.New(codeRoleConfigMissing, http.StatusNotFound, "role permissions not configured for project")
	case errors.Is(err, ErrTeamOwnerImmutable):
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "owner role cannot be modified")
	case errors.Is(err, ErrTeamMemberAlreadyExist):
		return apperr.New(codeMemberExists, http.StatusConflict, "user is already a project member")
	default:
		return apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "team operation failed"), err)
	}
}

func validatePermissionMaskDependencies(mask uint64) error {
	for domain := 0; domain < 10; domain++ {
		viewBit := uint64(1) << uint(domain*6)
		var elevatedBits uint64
		for action := 1; action <= 5; action++ {
			elevatedBits |= uint64(1) << uint(domain*6+action)
		}
		if mask&elevatedBits != 0 && mask&viewBit == 0 {
			return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "permission mask violates view dependency rule")
		}
	}
	return nil
}

func rbacRoleOwner() string {
	return "Owner"
}
