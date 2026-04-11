package policy

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/MrEthical07/superapi/internal/core/auth"
	"github.com/MrEthical07/superapi/internal/core/rbac"
)

func TestAuthRequiredMissingTokenUnauthorized(t *testing.T) {
	engine, _ := newPolicyTestAuthEngine(t)

	handlerCalled := false
	h := Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		}),
		AuthRequired(engine, auth.ModeHybrid),
	)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusUnauthorized)
	}
	if handlerCalled {
		t.Fatalf("expected handler not called")
	}
	if !strings.Contains(rr.Body.String(), `"code":"unauthorized"`) {
		t.Fatalf("expected unauthorized error code, got body=%s", rr.Body.String())
	}
}

func TestAuthRequiredValidTokenInjectsContext(t *testing.T) {
	engine, token := newPolicyTestAuthEngine(t)

	h := Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := auth.FromContext(r.Context())
			if !ok {
				t.Fatalf("expected principal in context")
			}
			if principal.UserID != "u1" {
				t.Fatalf("principal.user_id=%q want=%q", principal.UserID, "u1")
			}
			w.WriteHeader(http.StatusOK)
		}),
		AuthRequired(engine, auth.ModeHybrid),
	)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusOK)
	}
}

func TestRequirePermissionForbiddenWhenMissing(t *testing.T) {
	h := Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		RequirePermission(rbac.PermProjectEdit),
	)

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req = req.WithContext(auth.WithContext(req.Context(), auth.AuthContext{
		UserID:         "u1",
		PermissionMask: rbac.PermProjectView,
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusForbidden)
	}
	if !strings.Contains(rr.Body.String(), `"code":"forbidden"`) {
		t.Fatalf("expected forbidden code, got body=%s", rr.Body.String())
	}
}

func TestAuthRequiredNoSecretLeakOnFailure(t *testing.T) {
	engine, _ := newPolicyTestAuthEngine(t)

	h := Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		AuthRequired(engine, auth.ModeHybrid),
	)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req.Header.Set("Authorization", "Bearer token-signature-mismatch")
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusUnauthorized)
	}
	if strings.Contains(strings.ToLower(rr.Body.String()), "secret") {
		t.Fatalf("response leaked secret: %s", rr.Body.String())
	}
}

func TestRequireAnyPermissionForbiddenWhenMissingAll(t *testing.T) {
	h := Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		RequireAnyPermission(rbac.PermProjectEdit, rbac.PermTaskDelete),
	)

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req = req.WithContext(auth.WithContext(req.Context(), auth.AuthContext{
		UserID:         "u1",
		PermissionMask: rbac.PermProjectView,
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusForbidden)
	}
}

func TestRequireAnyPermissionAllowsWhenAnyPresent(t *testing.T) {
	h := Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		RequireAnyPermission(rbac.PermProjectEdit, rbac.PermProjectView),
	)

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req = req.WithContext(auth.WithContext(req.Context(), auth.AuthContext{
		UserID:         "u1",
		PermissionMask: rbac.PermProjectView,
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusOK)
	}
}

func TestRequireAllPermissionsRequiresEveryBit(t *testing.T) {
	h := Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		RequireAllPermissions(rbac.PermProjectView, rbac.PermProjectEdit),
	)

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req = req.WithContext(auth.WithContext(req.Context(), auth.AuthContext{
		UserID:         "u1",
		PermissionMask: rbac.PermProjectView,
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusForbidden)
	}
}

func TestRequirePermissionTreatsMissingMaskAsZero(t *testing.T) {
	h := Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		RequirePermission(rbac.PermProjectView),
	)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/secure", nil))

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusForbidden)
	}
}

func TestProjectRequiredUnauthorizedWhenMissingAuth(t *testing.T) {
	h := Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		ProjectRequired(),
	)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/secure", nil))

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusUnauthorized)
	}
}

func TestProjectRequiredForbiddenWhenProjectMissing(t *testing.T) {
	h := Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		ProjectRequired(),
	)

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req = req.WithContext(auth.WithContext(req.Context(), auth.AuthContext{UserID: "u1"}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusForbidden)
	}
}

func TestProjectRequiredRejectsAuthContextFallbackWithoutPathProject(t *testing.T) {
	h := Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		ProjectRequired(),
	)

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req = req.WithContext(auth.WithContext(req.Context(), auth.AuthContext{UserID: "u1", ProjectID: "p1"}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusForbidden)
	}
}

func TestProjectMatchFromPathPassesOnMatch(t *testing.T) {
	r := chi.NewRouter()
	r.With(ProjectMatchFromPath("id")).Get("/api/v1/projects/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/p1", nil)
	req = req.WithContext(auth.WithContext(req.Context(), auth.AuthContext{UserID: "u1", ProjectID: "p1"}))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusOK)
	}
}

func TestProjectMatchFromPathReturnsForbiddenOnMismatch(t *testing.T) {
	r := chi.NewRouter()
	r.With(ProjectMatchFromPath("id")).Get("/api/v1/projects/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/p2", nil)
	req = req.WithContext(auth.WithContext(req.Context(), auth.AuthContext{UserID: "u1", ProjectID: "p1"}))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusForbidden)
	}
}

func TestProjectMatchFromPathReturnsBadRequestOnMissingParam(t *testing.T) {
	h := Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		ProjectMatchFromPath("id"),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	req = req.WithContext(auth.WithContext(req.Context(), auth.AuthContext{UserID: "u1", ProjectID: "p1"}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusBadRequest)
	}
}
