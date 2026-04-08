package policy

import (
	"context"
	"net/http"
	"strings"

	"github.com/MrEthical07/superapi/internal/core/auth"
	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/requestid"
	"github.com/MrEthical07/superapi/internal/core/response"
)

// GetUserMask extracts user permission mask from request context.
//
// Missing auth context or missing mask resolves to a zero mask.
func GetUserMask(ctx context.Context) uint64 {
	principal, ok := auth.FromContext(ctx)
	if !ok {
		return 0
	}
	return principal.PermissionMask
}

// HasPerm evaluates whether mask contains the requested permission bit.
func HasPerm(mask uint64, perm uint64) bool {
	return mask&perm != 0
}

// HasRole checks whether request context carries the provided role value.
//
// Role is a convenience helper and is not the source of truth for authorization.
func HasRole(ctx context.Context, role string) bool {
	principal, ok := auth.FromContext(ctx)
	if !ok {
		return false
	}

	expected := strings.TrimSpace(role)
	if expected == "" {
		return false
	}

	return strings.EqualFold(strings.TrimSpace(principal.Role), expected)
}

// RequirePermission enforces one permission bit for the current request.
func RequirePermission(perm uint64) Policy {
	if perm == 0 {
		panicInvalidRouteConfigf("%s requires a non-zero permission", PolicyTypeRequirePermission)
	}

	p := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rid := requestid.FromContext(r.Context())
			mask := GetUserMask(r.Context())
			if !HasPerm(mask, perm) {
				response.Error(w, apperr.New(apperr.CodeForbidden, http.StatusForbidden, "forbidden"), rid)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	return annotatePolicy(p, Metadata{Type: PolicyTypeRequirePermission, Name: "RequirePermission"})
}

// RequireAnyPermission enforces any-of permission checks.
func RequireAnyPermission(perms ...uint64) Policy {
	required := normalizeRequiredPermissions(perms)
	if len(required) == 0 {
		panicInvalidRouteConfigf("%s requires at least one non-zero permission", PolicyTypeRequireAnyPermission)
	}

	p := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rid := requestid.FromContext(r.Context())
			mask := GetUserMask(r.Context())
			for _, perm := range required {
				if HasPerm(mask, perm) {
					next.ServeHTTP(w, r)
					return
				}
			}
			response.Error(w, apperr.New(apperr.CodeForbidden, http.StatusForbidden, "forbidden"), rid)
		})
	}

	return annotatePolicy(p, Metadata{Type: PolicyTypeRequireAnyPermission, Name: "RequireAnyPermission"})
}

// RequireAllPermissions enforces all-of permission checks.
func RequireAllPermissions(perms ...uint64) Policy {
	required := normalizeRequiredPermissions(perms)
	if len(required) == 0 {
		panicInvalidRouteConfigf("%s requires at least one non-zero permission", PolicyTypeRequireAllPermissions)
	}

	p := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rid := requestid.FromContext(r.Context())
			mask := GetUserMask(r.Context())
			for _, perm := range required {
				if !HasPerm(mask, perm) {
					response.Error(w, apperr.New(apperr.CodeForbidden, http.StatusForbidden, "forbidden"), rid)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}

	return annotatePolicy(p, Metadata{Type: PolicyTypeRequireAllPermissions, Name: "RequireAllPermissions"})
}

func normalizeRequiredPermissions(perms []uint64) []uint64 {
	required := make([]uint64, 0, len(perms))
	for _, perm := range perms {
		if perm != 0 {
			required = append(required, perm)
		}
	}
	return required
}
