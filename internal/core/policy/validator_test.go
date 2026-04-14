package policy

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"

	"github.com/MrEthical07/superapi/internal/core/auth"
	"github.com/MrEthical07/superapi/internal/core/cache"
	"github.com/MrEthical07/superapi/internal/core/permissions"
	"github.com/MrEthical07/superapi/internal/core/ratelimit"
	"github.com/MrEthical07/superapi/internal/core/rbac"
)

type allowLimiter struct{}

func (allowLimiter) Allow(context.Context, ratelimit.Request) (ratelimit.Decision, error) {
	return ratelimit.Decision{Allowed: true, Outcome: ratelimit.OutcomeAllowed}, nil
}

type allowPermissionResolver struct{}

func (allowPermissionResolver) Resolve(context.Context, string, string) (permissions.Resolution, error) {
	return permissions.Resolution{Mask: rbac.PermProjectView}, nil
}

func TestMustValidateRoutePanicsOnPolicyOrderViolation(t *testing.T) {
	assertRouteConfigPanic(t, "cannot appear after", func() {
		MustValidateRoute(
			http.MethodGet,
			"/api/v1/system/whoami",
			RateLimit(allowLimiter{}, ratelimit.Rule{Limit: 10, Window: time.Minute, Scope: ratelimit.ScopeAnon}),
			AuthRequired(nil, auth.ModeHybrid),
		)
	})
}

func TestMustValidateRoutePanicsOnMissingAuthDependency(t *testing.T) {
	assertRouteConfigPanic(t, string(PolicyTypeAuthRequired), func() {
		MustValidateRoute(
			http.MethodGet,
			"/api/v1/system/whoami",
			RequirePermission(rbac.PermProjectView),
		)
	})
}

func TestMustValidateRoutePanicsOnMissingResolverDependency(t *testing.T) {
	assertRouteConfigPanic(t, string(PolicyTypeResolvePermissions), func() {
		MustValidateRoute(
			http.MethodGet,
			"/api/v1/projects/{project_id}/tasks/{id}",
			AuthRequired(nil, auth.ModeHybrid),
			ProjectRequired(),
			ProjectMatchFromPath("project_id"),
			RequirePermission(rbac.PermProjectView),
		)
	})
}

func TestMustValidateRoutePanicsOnUnsafeAuthenticatedCache(t *testing.T) {
	mr := miniredis.RunT(t)
	mgr := newCacheManagerForPolicyTests(t, mr.Addr(), true)

	assertRouteConfigPanic(t, "VaryBy.UserID or VaryBy.ProjectID", func() {
		MustValidateRoute(
			http.MethodGet,
			"/api/v1/system/whoami",
			AuthRequired(nil, auth.ModeHybrid),
			CacheRead(mgr, cache.CacheReadConfig{TTL: time.Minute}),
		)
	})
}

func TestMustValidateRoutePassesOnSharedAuthenticatedCache(t *testing.T) {
	mr := miniredis.RunT(t)
	mgr := newCacheManagerForPolicyTests(t, mr.Addr(), true)

	assertRouteConfigDoesNotPanic(t, func() {
		MustValidateRoute(
			http.MethodGet,
			"/api/v1/home/docs",
			AuthRequired(nil, auth.ModeHybrid),
			CacheRead(mgr, cache.CacheReadConfig{
				TTL:                 time.Minute,
				AllowAuthenticated:  true,
				SharedAuthenticated: true,
			}),
		)
	})
}

func TestMustValidateRoutePanicsOnProjectPathWithoutMatchPolicy(t *testing.T) {
	assertRouteConfigPanic(t, string(PolicyTypeProjectMatchFromPath), func() {
		MustValidateRoute(
			http.MethodGet,
			"/api/v1/projects/{project_id}/tasks",
			AuthRequired(nil, auth.ModeHybrid),
			ProjectRequired(),
		)
	})
}

func TestMustValidateRoutePanicsWhenProjectRequiredRouteHasNoProjectPathParam(t *testing.T) {
	assertRouteConfigPanic(t, "requires path parameter {project_id} or {projectId}", func() {
		MustValidateRoute(
			http.MethodGet,
			"/api/v1/system/whoami",
			AuthRequired(nil, auth.ModeHybrid),
			ProjectRequired(),
			ProjectMatchFromPath("project_id"),
		)
	})
}

func TestMustValidateRoutePassesOnStrictValidConfiguration(t *testing.T) {
	mr := miniredis.RunT(t)
	mgr := newCacheManagerForPolicyTests(t, mr.Addr(), true)

	assertRouteConfigDoesNotPanic(t, func() {
		MustValidateRoute(
			http.MethodGet,
			"/api/v1/projects/{project_id}/tasks/{id}",
			AuthRequired(nil, auth.ModeHybrid),
			ProjectRequired(),
			ProjectMatchFromPath("project_id"),
			ResolvePermissions(allowPermissionResolver{}),
			RequirePermission(rbac.PermProjectView),
			RateLimit(allowLimiter{}, ratelimit.Rule{Limit: 10, Window: time.Minute, Scope: ratelimit.ScopeProject}),
			CacheRead(mgr, cache.CacheReadConfig{
				TTL: time.Minute,
				VaryBy: cache.CacheVaryBy{
					ProjectID: true,
				},
			}),
		)
	})
}

func assertRouteConfigPanic(t *testing.T, expectedContains string, fn func()) {
	t.Helper()
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatalf("expected panic")
		}
		message := strings.TrimSpace(toString(recovered))
		if !strings.Contains(message, "invalid route config") {
			t.Fatalf("unexpected panic message: %q", message)
		}
		if expectedContains != "" && !strings.Contains(message, expectedContains) {
			t.Fatalf("panic message %q does not contain %q", message, expectedContains)
		}
	}()

	fn()
}

func assertRouteConfigDoesNotPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("unexpected panic: %v", recovered)
		}
	}()

	fn()
}

func toString(v any) string {
	s, ok := v.(string)
	if ok {
		return s
	}
	return "panic"
}
