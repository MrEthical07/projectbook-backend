package pages

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MrEthical07/superapi/internal/core/app"
	"github.com/MrEthical07/superapi/internal/core/httpx"
	"github.com/MrEthical07/superapi/internal/core/permissions"
)

type allowResolver struct{}

func (allowResolver) Resolve(_ context.Context, userID, projectID string) (permissions.Resolution, error) {
	return permissions.Resolution{UserID: userID, ProjectID: projectID, Role: "Member", Mask: 1, UpdatedAtUnix: 1}, nil
}

func TestPagesRegisterRequiresResolver(t *testing.T) {
	m := &Module{}
	m.BindDependencies(&app.Dependencies{})

	r := httpx.NewMux()
	if err := m.Register(r); err == nil {
		t.Fatal("expected resolver dependency error")
	}
}

func TestPagesRoutesRequireAuth(t *testing.T) {
	m := &Module{}
	m.BindDependencies(&app.Dependencies{PermissionsResolver: allowResolver{}})

	r := httpx.NewMux()
	if err := m.Register(r); err != nil {
		t.Fatalf("register: %v", err)
	}

	tests := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/api/v1/projects/atlas-2026/pages", ""},
		{http.MethodPost, "/api/v1/projects/atlas-2026/pages", `{"title":"Research"}`},
		{http.MethodGet, "/api/v1/projects/atlas-2026/pages/research-notes", ""},
		{http.MethodPatch, "/api/v1/projects/atlas-2026/pages/pg-1", `{"state":{"status":"Draft"}}`},
		{http.MethodPatch, "/api/v1/projects/atlas-2026/pages/pg-1/rename", `{"title":"Updated"}`},
	}

	for _, tc := range tests {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		if tc.body != "" {
			req.Header.Set("Content-Type", "application/json")
		}

		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("path=%s status=%d want=%d body=%s", tc.path, rr.Code, http.StatusUnauthorized, rr.Body.String())
		}
	}
}
