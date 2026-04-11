package calendar

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

func TestCalendarRegisterRequiresResolver(t *testing.T) {
	m := &Module{}
	m.BindDependencies(&app.Dependencies{})

	r := httpx.NewMux()
	if err := m.Register(r); err == nil {
		t.Fatal("expected resolver dependency error")
	}
}

func TestCalendarRoutesRequireAuth(t *testing.T) {
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
		{http.MethodGet, "/api/v1/projects/atlas-2026/calendar", ""},
		{http.MethodPost, "/api/v1/projects/atlas-2026/calendar", `{"title":"Review","start":"2026-02-20","end":"2026-02-20","allDay":true,"owner":"Ayush","phase":"Prototype","eventKind":"Review"}`},
		{http.MethodGet, "/api/v1/projects/atlas-2026/calendar/evt-1", ""},
		{http.MethodPut, "/api/v1/projects/atlas-2026/calendar/evt-1", `{"state":{"title":"Updated"}}`},
		{http.MethodDelete, "/api/v1/projects/atlas-2026/calendar/evt-1", ""},
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
