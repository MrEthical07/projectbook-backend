package system

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MrEthical07/superapi/internal/core/httpx"
)

func TestWhoamiRequiresAuth(t *testing.T) {
	m := &Module{}
	m.BindDependencies(nil)

	r := httpx.NewMux()
	if err := m.Register(r); err != nil {
		t.Fatalf("register: %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/whoami", nil)
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusUnauthorized)
	}
}

func TestSessionContextRequiresAuth(t *testing.T) {
	m := &Module{}
	m.BindDependencies(nil)

	r := httpx.NewMux()
	if err := m.Register(r); err != nil {
		t.Fatalf("register: %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/session-context", nil)
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusUnauthorized)
	}
}

func TestParseDurationRequiresJSON(t *testing.T) {
	m := &Module{}
	m.BindDependencies(nil)

	r := httpx.NewMux()
	if err := m.Register(r); err != nil {
		t.Fatalf("register: %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/system/parse-duration", strings.NewReader(`{"duration":"1s"}`))
	req.Header.Set("Content-Type", "text/plain")
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("status=%d want=%d body=%s", rr.Code, http.StatusUnsupportedMediaType, rr.Body.String())
	}
}

func TestBuildSessionContextTokenEmbedsExpectedClaims(t *testing.T) {
	t.Setenv(sessionContextSecretEnv, strings.Repeat("a", 40))

	response := sessionContextResponse{
		UserID:      "user-123",
		BackendRole: "Admin",
		ProjectPermissions: []sessionContextProjectPermission{
			{
				ProjectID:      "project-1",
				ProjectSlug:    "my-project",
				Role:           "Editor",
				PermissionMask: 31,
				IsCustom:       false,
				UpdatedAtUnix:  1700000000,
			},
		},
		SnapshotHash:     "snapshot-abc",
		ExpiresInSeconds: 600,
	}

	now := time.Unix(1700000100, 0).UTC()
	token, expUnix, err := buildSessionContextToken(response, now)
	if err != nil {
		t.Fatalf("buildSessionContextToken() error = %v", err)
	}
	if got, want := expUnix, now.Unix()+600; got != want {
		t.Fatalf("expiresUnix=%d want=%d", got, want)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("token parts=%d want=3", len(parts))
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	var claims sessionContextTokenClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		t.Fatalf("unmarshal claims: %v", err)
	}

	if got, want := claims.UserID, response.UserID; got != want {
		t.Fatalf("claims.user_id=%q want=%q", got, want)
	}
	if got, want := claims.SnapshotHash, response.SnapshotHash; got != want {
		t.Fatalf("claims.snapshot_hash=%q want=%q", got, want)
	}
	if got, want := claims.IssuedAtUnix, now.Unix(); got != want {
		t.Fatalf("claims.iat=%d want=%d", got, want)
	}
	if got, want := claims.ExpiresAtUnix, expUnix; got != want {
		t.Fatalf("claims.exp=%d want=%d", got, want)
	}
	if got, want := claims.Version, sessionContextTokenVersion; got != want {
		t.Fatalf("claims.v=%d want=%d", got, want)
	}

	unsigned := parts[0] + "." + parts[1]
	if got, want := parts[2], signSessionContextToken(unsigned); got != want {
		t.Fatalf("token signature=%q want=%q", got, want)
	}
}

func TestResolveSessionContextSigningSecret(t *testing.T) {
	t.Setenv(sessionContextSecretEnv, "short")
	if got, want := resolveSessionContextSigningSecret(), sessionContextFallbackSecret; got != want {
		t.Fatalf("fallback secret=%q want=%q", got, want)
	}

	configured := strings.Repeat("z", sessionContextSecretMinLength)
	t.Setenv(sessionContextSecretEnv, configured)
	if got, want := resolveSessionContextSigningSecret(), configured; got != want {
		t.Fatalf("configured secret=%q want=%q", got, want)
	}
}
