package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MrEthical07/superapi/internal/core/httpx"
	"github.com/MrEthical07/superapi/internal/core/ratelimit"
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

func TestRefreshTokenRateLimitKeyerUsesTokenHashAndPreservesBody(t *testing.T) {
	t.Parallel()

	const (
		refreshToken = "refresh-token-value"
		prefixLen    = 12
	)
	body := `{"refresh_token":"` + refreshToken + `"}`
	keyer := refreshTokenRateLimitKeyer(prefixLen)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", strings.NewReader(body))
	req.RemoteAddr = "198.51.100.7:12345"

	scope, id := keyer(req)
	if scope != ratelimit.ScopeToken {
		t.Fatalf("scope=%s want=%s", scope, ratelimit.ScopeToken)
	}

	hash := sha256.Sum256([]byte(refreshToken))
	wantID := hex.EncodeToString(hash[:])[:prefixLen]
	if id != wantID {
		t.Fatalf("identifier=%s want=%s", id, wantID)
	}

	replayed, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if got := string(replayed); got != body {
		t.Fatalf("body=%s want=%s", got, body)
	}
}

func TestRefreshTokenRateLimitKeyerFallsBackToIP(t *testing.T) {
	t.Parallel()

	keyer := refreshTokenRateLimitKeyer(16)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", strings.NewReader(`{"refresh_token":"   "}`))
	req.RemoteAddr = "203.0.113.9:4321"

	scope, id := keyer(req)
	if scope != ratelimit.ScopeIP {
		t.Fatalf("scope=%s want=%s", scope, ratelimit.ScopeIP)
	}
	if id != "203.0.113.9" {
		t.Fatalf("identifier=%s want=%s", id, "203.0.113.9")
	}
}
