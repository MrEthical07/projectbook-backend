package sidebar

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

func TestSidebarRegisterRequiresResolver(t *testing.T) {
	m := &Module{}
	m.BindDependencies(&app.Dependencies{})

	r := httpx.NewMux()
	if err := m.Register(r); err == nil {
		t.Fatal("expected resolver dependency error")
	}
}

func TestSidebarRoutesRequireAuth(t *testing.T) {
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
		{http.MethodPost, "/api/v1/projects/atlas-2026/sidebar/artifacts", `{"prefix":"stories","title":"Story"}`},
		{http.MethodPut, "/api/v1/projects/atlas-2026/sidebar/artifacts/st-1/rename", `{"prefix":"stories","title":"Renamed"}`},
		{http.MethodDelete, "/api/v1/projects/atlas-2026/sidebar/artifacts/st-1", `{"prefix":"stories"}`},
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
