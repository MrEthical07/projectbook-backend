package projectscope

import (
	"net/http"
	"testing"

	"github.com/MrEthical07/superapi/internal/core/auth"
)

func TestProjectIDFromContext(t *testing.T) {
	ctx := auth.WithContext(t.Context(), auth.AuthContext{UserID: "u1", ProjectID: "p1"})
	projectID, ok := ProjectIDFromContext(ctx)
	if !ok {
		t.Fatalf("expected project id from context")
	}
	if projectID != "p1" {
		t.Fatalf("projectID=%q want=%q", projectID, "p1")
	}
}

func TestRequireProjectMissing(t *testing.T) {
	err := RequireProject(t.Context())
	if err == nil {
		t.Fatalf("expected project required error")
	}
	ae := err.Error()
	if ae == "" {
		t.Fatalf("expected non-empty error")
	}
}

func TestRequireProjectPresent(t *testing.T) {
	ctx := auth.WithContext(t.Context(), auth.AuthContext{ProjectID: "p1"})
	if err := RequireProject(ctx); err != nil {
		t.Fatalf("RequireProject() error = %v", err)
	}
}

func TestIsSameProject(t *testing.T) {
	if !IsSameProject("p1", "p1") {
		t.Fatalf("expected same project")
	}
	if IsSameProject("p1", "p2") {
		t.Fatalf("expected mismatch")
	}
	if IsSameProject("", "p1") {
		t.Fatalf("expected empty principal project to fail")
	}
}

func TestRequireProjectErrorShape(t *testing.T) {
	err := RequireProject(t.Context())
	ae, ok := err.(interface{ Error() string })
	if !ok || ae.Error() == "" {
		t.Fatalf("expected app error compatible error")
	}
	_ = http.StatusForbidden
}
