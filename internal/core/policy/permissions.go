package policy

import (
	"net/http"
	"strings"

	"github.com/MrEthical07/superapi/internal/core/auth"
	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/permissions"
	"github.com/MrEthical07/superapi/internal/core/requestid"
	"github.com/MrEthical07/superapi/internal/core/response"
)

// ResolvePermissions resolves the effective project permission mask for downstream RBAC policies.
//
// Behavior:
// - Returns 401 when authentication context is absent
// - Returns 403 when user is authenticated but not a member of the project
// - Returns 500 when membership exists but permission mask is inconsistent
// - Returns 503 when resolver dependencies fail
func ResolvePermissions(resolver permissions.Resolver) Policy {
	if resolver == nil {
		panicInvalidRouteConfigf("%s requires a non-nil permissions resolver", PolicyTypeResolvePermissions)
	}

	p := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rid := requestid.FromContext(r.Context())

			principal, ok := auth.FromContext(r.Context())
			if !ok || strings.TrimSpace(principal.UserID) == "" {
				response.Error(w, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required"), rid)
				return
			}

			projectID := strings.TrimSpace(principal.ProjectID)
			if projectID == "" {
				response.Error(w, apperr.New(apperr.CodeForbidden, http.StatusForbidden, "project scope required"), rid)
				return
			}

			resolved, err := resolver.Resolve(r.Context(), principal.UserID, projectID)
			if err != nil {
				switch {
				case permissions.IsMembershipNotFound(err):
					response.Error(w, apperr.New(apperr.CodeForbidden, http.StatusForbidden, "project membership required"), rid)
				case permissions.IsMaskInconsistent(err):
					response.Error(w, apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "permission mask unavailable"), rid)
				case permissions.IsDependencyFailure(err):
					response.Error(w, apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "permission resolver unavailable"), rid)
				default:
					response.Error(w, apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "permission resolution failed"), rid)
				}
				return
			}

			principal.PermissionMask = resolved.Mask
			if role := strings.TrimSpace(resolved.Role); role != "" {
				principal.Role = role
			}
			principal.ProjectID = projectID

			next.ServeHTTP(w, r.WithContext(auth.WithContext(r.Context(), principal)))
		})
	}

	return annotatePolicy(p, Metadata{Type: PolicyTypeResolvePermissions, Name: "ResolvePermissions"})
}
