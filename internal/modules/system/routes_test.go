package system

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
