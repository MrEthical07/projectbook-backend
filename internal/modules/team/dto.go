package team

import (
	"net/http"
	"net/mail"
	"strings"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/rbac"
)

const maxBatchInvites = 100

type teamMember struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Status   string `json:"status"`
	JoinedAt string `json:"joinedAt"`
}

type teamInvite struct {
	Email    string `json:"email"`
	Role     string `json:"role"`
	SentDate string `json:"sentDate"`
	Status   string `json:"status"`
}

type teamMembersResponse struct {
	Members []teamMember `json:"members"`
	Invites []teamInvite `json:"invites"`
}

type teamRoleMember struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	Role           string `json:"role"`
	Status         string `json:"status"`
	JoinedAt       string `json:"joinedAt"`
	IsCustom       bool   `json:"isCustom"`
	PermissionMask string `json:"permissionMask"`
}

type teamRolesResponse struct {
	RolePermissionMasks map[string]string `json:"rolePermissionMasks"`
	Members             []teamRoleMember  `json:"members"`
}

type createInviteRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (r createInviteRequest) Validate() error {
	email := normalizeEmail(r.Email)
	if email == "" || !isValidEmail(email) {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "email is invalid")
	}
	if _, ok := canonicalRole(r.Role); !ok {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "role is invalid")
	}
	return nil
}

type batchInviteRequest struct {
	Invites []createInviteRequest `json:"invites"`
}

func (r batchInviteRequest) Validate() error {
	if len(r.Invites) == 0 {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invites must contain at least one item")
	}
	if len(r.Invites) > maxBatchInvites {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invites exceeds maximum batch size")
	}
	return nil
}

type createInviteResponse struct {
	Email    string `json:"email"`
	Role     string `json:"role"`
	SentDate string `json:"sentDate"`
	Status   string `json:"status"`
}

type batchInviteFailure struct {
	Email   string `json:"email"`
	Role    string `json:"role"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type batchInviteSuccess struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type batchInviteResponse struct {
	ProjectID string               `json:"projectId"`
	Invited   []batchInviteSuccess `json:"invited"`
	Failed    []batchInviteFailure `json:"failed,omitempty"`
}

type cancelInviteResponse struct {
	Email string `json:"email"`
}

type updateMemberPermissionsRequest struct {
	Role           string `json:"role"`
	IsCustom       bool   `json:"isCustom"`
	PermissionMask uint64 `json:"permissionMask"`
}

func (r updateMemberPermissionsRequest) Validate() error {
	role, ok := canonicalRole(r.Role)
	if !ok {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "role is invalid")
	}
	if role == rbac.RoleOwner {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "owner role cannot be assigned")
	}
	return nil
}

type updateMemberPermissionsResponse struct {
	MemberID       string `json:"memberId"`
	Role           string `json:"role"`
	IsCustom       bool   `json:"isCustom"`
	PermissionMask uint64 `json:"permissionMask"`
}

type updateRolePermissionsRequest struct {
	Role           string `json:"role"`
	PermissionMask uint64 `json:"permissionMask"`
}

func (r updateRolePermissionsRequest) Validate() error {
	if _, ok := canonicalRole(r.Role); !ok {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "role is invalid")
	}
	return nil
}

type updateRolePermissionsResponse struct {
	Role                    string `json:"role"`
	PermissionMask          uint64 `json:"permissionMask"`
	CustomMembersUnaffected int    `json:"customMembersUnaffected"`
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isValidEmail(email string) bool {
	if strings.TrimSpace(email) == "" {
		return false
	}
	parsed, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(parsed.Address), strings.TrimSpace(email))
}

func canonicalRole(raw string) (string, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", false
	}

	slug := strings.ToLower(trimmed)
	switch slug {
	case "owner":
		return rbac.RoleOwner, true
	case "admin":
		return rbac.RoleAdmin, true
	case "editor":
		return rbac.RoleEditor, true
	case "member":
		return rbac.RoleMember, true
	case "viewer":
		return rbac.RoleViewer, true
	case "limited-access", "limited access":
		return rbac.RoleLimitedAccess, true
	}

	for _, role := range rbac.CanonicalRoles() {
		if strings.EqualFold(role, trimmed) {
			return role, true
		}
	}

	return "", false
}

func canonicalRoleFromSlug(raw string) (string, bool) {
	slug := strings.ToLower(strings.TrimSpace(raw))
	switch slug {
	case "owner":
		return rbac.RoleOwner, true
	case "admin":
		return rbac.RoleAdmin, true
	case "editor":
		return rbac.RoleEditor, true
	case "member":
		return rbac.RoleMember, true
	case "viewer":
		return rbac.RoleViewer, true
	case "limited-access":
		return rbac.RoleLimitedAccess, true
	default:
		return "", false
	}
}

func roleSlug(role string) string {
	canonical, ok := canonicalRole(role)
	if !ok {
		return ""
	}
	switch canonical {
	case rbac.RoleOwner:
		return "owner"
	case rbac.RoleAdmin:
		return "admin"
	case rbac.RoleEditor:
		return "editor"
	case rbac.RoleMember:
		return "member"
	case rbac.RoleViewer:
		return "viewer"
	case rbac.RoleLimitedAccess:
		return "limited-access"
	default:
		return ""
	}
}
