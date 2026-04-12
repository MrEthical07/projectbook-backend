package rbac

import "strings"

const (
	RoleOwner         = "Owner"
	RoleAdmin         = "Admin"
	RoleEditor        = "Editor"
	RoleMember        = "Member"
	RoleViewer        = "Viewer"
	RoleLimitedAccess = "Limited Access"
)

var canonicalRoles = []string{
	RoleOwner,
	RoleAdmin,
	RoleEditor,
	RoleMember,
	RoleViewer,
	RoleLimitedAccess,
}

var defaultRoleMasks = map[string]uint64{
	RoleOwner:         1152921504606846975,
	RoleAdmin:         864691128455135229,
	RoleEditor:        20016033248999873,
	RoleMember:        875734824153537,
	RoleViewer:        18300341342965825,
	RoleLimitedAccess: 0,
}

// CanonicalRoles returns the stable role ordering used for role-mask lifecycle operations.
func CanonicalRoles() []string {
	roles := make([]string, len(canonicalRoles))
	copy(roles, canonicalRoles)
	return roles
}

// DefaultRoleMask returns the canonical default permission mask for one role.
func DefaultRoleMask(role string) (uint64, bool) {
	normalizedRole := strings.TrimSpace(role)
	for _, canonicalRole := range canonicalRoles {
		if strings.EqualFold(canonicalRole, normalizedRole) {
			mask, ok := defaultRoleMasks[canonicalRole]
			return mask, ok
		}
	}
	return 0, false
}

// DefaultRoleMasks returns a copy of canonical role-mask mappings.
func DefaultRoleMasks() map[string]uint64 {
	copyMasks := make(map[string]uint64, len(defaultRoleMasks))
	for role, mask := range defaultRoleMasks {
		copyMasks[role] = mask
	}
	return copyMasks
}
