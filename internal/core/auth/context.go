package auth

import "context"

type principalKey struct{}

// AuthContext represents authenticated principal data attached to request context.
type AuthContext struct {
	// UserID is the canonical authenticated user identifier.
	UserID string `json:"user_id"`
	// ProjectID is the resolved project scope used for authorization.
	ProjectID string `json:"project_id,omitempty"`
	// Role is the resolved role name for RBAC checks.
	Role string `json:"role,omitempty"`
	// PermissionMask is the resolved authorization mask used by RBAC middleware.
	PermissionMask uint64 `json:"permission_mask,omitempty"`
	// Permissions is the resolved permission set for RBAC checks.
	Permissions []string `json:"permissions,omitempty"`
}

// WithContext stores AuthContext on a request context.
func WithContext(ctx context.Context, principal AuthContext) context.Context {
	return context.WithValue(ctx, principalKey{}, principal)
}

// FromContext reads AuthContext from request context.
func FromContext(ctx context.Context) (AuthContext, bool) {
	v, ok := ctx.Value(principalKey{}).(AuthContext)
	if !ok {
		return AuthContext{}, false
	}
	return v, true
}
