package activity

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MrEthical07/superapi/internal/core/app"
	"github.com/MrEthical07/superapi/internal/core/httpx"
	"github.com/MrEthical07/superapi/internal/core/permissions"
)

type allowResolver struct{}

func (allowResolver) Resolve(_ context.Context, userID, projectID string) (permissions.Resolution, error) {
	return permissions.Resolution{UserID: userID, ProjectID: projectID, Role: "Member", Mask: 1, UpdatedAtUnix: 1}, nil
}

func TestActivityRegisterRequiresResolver(t *testing.T) {
	m := &Module{}
	m.BindDependencies(&app.Dependencies{})

	r := httpx.NewMux()
	if err := m.Register(r); err == nil {
		t.Fatal("expected resolver dependency error")
	}
}

func TestActivityRoutesRequireAuth(t *testing.T) {
	m := &Module{}
	m.BindDependencies(&app.Dependencies{PermissionsResolver: allowResolver{}})

	r := httpx.NewMux()
	if err := m.Register(r); err != nil {
		t.Fatalf("register: %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/atlas-2026/activity", nil)
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d want=%d body=%s", rr.Code, http.StatusUnauthorized, rr.Body.String())
	}
}
