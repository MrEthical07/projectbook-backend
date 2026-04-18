package sidebar

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MrEthical07/superapi/internal/core/app"
	"github.com/MrEthical07/superapi/internal/core/httpx"
)

func TestSidebarRegisterNoLongerRequiresResolver(t *testing.T) {
	m := &Module{}
	m.BindDependencies(&app.Dependencies{})

	r := httpx.NewMux()
	if err := m.Register(r); err != nil {
		t.Fatalf("register: %v", err)
	}
}

func TestSidebarMutationRoutesAreRemoved(t *testing.T) {
	m := &Module{}
	m.BindDependencies(&app.Dependencies{})

	r := httpx.NewMux()
	if err := m.Register(r); err != nil {
		t.Fatalf("register: %v", err)
	}

	tests := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodPost, "/api/v1/projects/atlas-2026/sidebar/artifacts", `{"prefix":"stories","title":"Story"}`},
		{http.MethodPatch, "/api/v1/projects/atlas-2026/sidebar/artifacts/st-1/rename", `{"prefix":"stories","title":"Renamed"}`},
		{http.MethodDelete, "/api/v1/projects/atlas-2026/sidebar/artifacts/st-1", `{"prefix":"stories"}`},
	}

	for _, tc := range tests {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		if tc.body != "" {
			req.Header.Set("Content-Type", "application/json")
		}

		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("path=%s status=%d want=%d body=%s", tc.path, rr.Code, http.StatusNotFound, rr.Body.String())
		}
	}
}
