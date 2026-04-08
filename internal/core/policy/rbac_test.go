package policy

import (
	"context"
	"testing"

	"github.com/MrEthical07/superapi/internal/core/auth"
	"github.com/MrEthical07/superapi/internal/core/rbac"
)

func TestGetUserMaskReturnsZeroWhenMissing(t *testing.T) {
	if got := GetUserMask(context.Background()); got != 0 {
		t.Fatalf("GetUserMask()=%d want=0", got)
	}
}

func TestGetUserMaskFromAuthContext(t *testing.T) {
	ctx := auth.WithContext(context.Background(), auth.AuthContext{PermissionMask: rbac.PermTaskEdit})
	if got := GetUserMask(ctx); got != rbac.PermTaskEdit {
		t.Fatalf("GetUserMask()=%d want=%d", got, rbac.PermTaskEdit)
	}
}

func TestHasPermUsesBitMask(t *testing.T) {
	mask := rbac.PermProjectView | rbac.PermTaskDelete
	if !HasPerm(mask, rbac.PermTaskDelete) {
		t.Fatalf("HasPerm()=false want=true")
	}
	if HasPerm(mask, rbac.PermTaskCreate) {
		t.Fatalf("HasPerm()=true want=false")
	}
}

func TestHasRoleCaseInsensitive(t *testing.T) {
	ctx := auth.WithContext(context.Background(), auth.AuthContext{Role: "Owner"})
	if !HasRole(ctx, "owner") {
		t.Fatalf("HasRole()=false want=true")
	}
	if HasRole(ctx, "viewer") {
		t.Fatalf("HasRole()=true want=false")
	}
}

func TestRequirePermissionPanicsOnZeroPerm(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatalf("expected panic")
		}
	}()
	_ = RequirePermission(0)
}

func TestRequireAnyPermissionPanicsOnZeroOnly(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatalf("expected panic")
		}
	}()
	_ = RequireAnyPermission(0)
}

func TestRequireAllPermissionsPanicsOnZeroOnly(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatalf("expected panic")
		}
	}()
	_ = RequireAllPermissions(0)
}
