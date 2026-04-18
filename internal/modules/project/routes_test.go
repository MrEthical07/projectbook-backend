package project

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

func TestProjectRegisterRequiresResolver(t *testing.T) {
	m := &Module{}
	m.BindDependencies(&app.Dependencies{})

	r := httpx.NewMux()
	err := m.Register(r)
	if err == nil {
		t.Fatal("expected resolver dependency error")
	}
}

func TestProjectRoutesRequireAuth(t *testing.T) {
	m := &Module{}
	m.BindDependencies(&app.Dependencies{PermissionsResolver: allowResolver{}})

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
		{name: "dashboard", method: http.MethodGet, path: "/api/v1/projects/atlas-2026/dashboard"},
		{name: "dashboard events", method: http.MethodGet, path: "/api/v1/projects/atlas-2026/dashboard/events"},
		{name: "access", method: http.MethodGet, path: "/api/v1/projects/atlas-2026/access"},
		{name: "update settings", method: http.MethodPatch, path: "/api/v1/projects/atlas-2026/settings", body: `{"settings":{"projectName":"Atlas","projectStatus":"Active"}}`},
		{name: "delete", method: http.MethodDelete, path: "/api/v1/projects/atlas-2026"},
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
