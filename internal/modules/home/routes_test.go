package home

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MrEthical07/superapi/internal/core/httpx"
)

func TestHomeRoutesRequireAuth(t *testing.T) {
	m := &Module{}
	m.BindDependencies(nil)

	r := httpx.NewMux()
	if err := m.Register(r); err != nil {
		t.Fatalf("register: %v", err)
	}

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{name: "dashboard", method: http.MethodGet, path: "/api/v1/home/dashboard"},
		{name: "create project", method: http.MethodPost, path: "/api/v1/home/projects", body: `{"name":"Project A","icon":"rocket"}`},
		{name: "account", method: http.MethodGet, path: "/api/v1/home/account"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			r.ServeHTTP(rr, req)
			if rr.Code != http.StatusUnauthorized {
				t.Fatalf("status=%d want=%d body=%s", rr.Code, http.StatusUnauthorized, rr.Body.String())
			}
		})
	}
}
