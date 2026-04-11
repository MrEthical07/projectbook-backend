package policy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MrEthical07/superapi/internal/core/auth"
	"github.com/MrEthical07/superapi/internal/core/permissions"
	"github.com/MrEthical07/superapi/internal/core/rbac"
)

type stubPermissionResolver struct {
	result permissions.Resolution
	err    error

	called     bool
	lastUserID string
	lastProjID string
}

func (s *stubPermissionResolver) Resolve(_ context.Context, userID, projectID string) (permissions.Resolution, error) {
	s.called = true
	s.lastUserID = userID
	s.lastProjID = projectID
	return s.result, s.err
}

func TestResolvePermissionsUnauthorizedWithoutAuthContext(t *testing.T) {
	resolver := &stubPermissionResolver{}

	h := Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		ResolvePermissions(resolver),
	)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/secure", nil))

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusUnauthorized)
	}
	if resolver.called {
		t.Fatalf("resolver should not be called without auth context")
	}
}

func TestResolvePermissionsForbiddenWithoutProjectScope(t *testing.T) {
	resolver := &stubPermissionResolver{}

	h := Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		ResolvePermissions(resolver),
	)

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req = req.WithContext(auth.WithContext(req.Context(), auth.AuthContext{UserID: "u1"}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusForbidden)
	}
	if resolver.called {
		t.Fatalf("resolver should not be called without project scope")
	}
}

func TestResolvePermissionsErrorMappings(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{
			name:       "membership missing",
			err:        fmt.Errorf("%w: no active membership", permissions.ErrMembershipNotFound),
			wantStatus: http.StatusForbidden,
			wantCode:   "forbidden",
		},
		{
			name:       "mask inconsistent",
			err:        fmt.Errorf("%w: role mask missing", permissions.ErrMaskInconsistent),
			wantStatus: http.StatusInternalServerError,
			wantCode:   "internal_error",
		},
		{
			name:       "dependency failure",
			err:        fmt.Errorf("%w: redis timeout", permissions.ErrDependencyFailure),
			wantStatus: http.StatusServiceUnavailable,
			wantCode:   "dependency_unavailable",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resolver := &stubPermissionResolver{err: tc.err}
			h := Chain(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
				ResolvePermissions(resolver),
			)

			req := httptest.NewRequest(http.MethodGet, "/secure", nil)
			req = req.WithContext(auth.WithContext(req.Context(), auth.AuthContext{UserID: "u1", ProjectID: "p1"}))

			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Fatalf("status=%d want=%d", rr.Code, tc.wantStatus)
			}
			if !strings.Contains(rr.Body.String(), `"code":"`+tc.wantCode+`"`) {
				t.Fatalf("expected error code %q in body=%s", tc.wantCode, rr.Body.String())
			}
		})
	}
}

func TestResolvePermissionsInjectsResolvedMaskAndRole(t *testing.T) {
	resolver := &stubPermissionResolver{
		result: permissions.Resolution{Mask: rbac.PermProjectEdit, Role: "editor"},
	}

	h := Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := auth.FromContext(r.Context())
			if !ok {
				t.Fatalf("expected auth context")
			}
			if principal.PermissionMask != rbac.PermProjectEdit {
				t.Fatalf("permission mask=%d want=%d", principal.PermissionMask, rbac.PermProjectEdit)
			}
			if principal.Role != "editor" {
				t.Fatalf("role=%q want=%q", principal.Role, "editor")
			}
			w.WriteHeader(http.StatusOK)
		}),
		ResolvePermissions(resolver),
	)

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req = req.WithContext(auth.WithContext(req.Context(), auth.AuthContext{UserID: "u1", ProjectID: "p1"}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusOK)
	}
	if !resolver.called {
		t.Fatalf("expected resolver to be called")
	}
	if resolver.lastUserID != "u1" || resolver.lastProjID != "p1" {
		t.Fatalf("resolver args user=%q project=%q", resolver.lastUserID, resolver.lastProjID)
	}
}
