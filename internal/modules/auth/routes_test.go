package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MrEthical07/superapi/internal/core/httpx"
)

func TestLogoutRequiresAuth(t *testing.T) {
	m := &Module{}
	m.BindDependencies(nil)

	r := httpx.NewMux()
	if err := m.Register(r); err != nil {
		t.Fatalf("register: %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusUnauthorized)
	}
}

func TestLoginReturnsDependencyFailureWithoutEngine(t *testing.T) {
	m := &Module{}
	m.BindDependencies(nil)

	r := httpx.NewMux()
	if err := m.Register(r); err != nil {
		t.Fatalf("register: %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"test@example.com","password":"Passw0rd!"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusServiceUnavailable)
	}
	if !strings.Contains(rr.Body.String(), `"code":"dependency_unavailable"`) {
		t.Fatalf("expected dependency_unavailable body=%s", rr.Body.String())
	}
}
