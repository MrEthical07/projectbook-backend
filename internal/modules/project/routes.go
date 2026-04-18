package project

import (
	"fmt"
	"net/http"
	"time"

	"github.com/MrEthical07/superapi/internal/core/cache"
	"github.com/MrEthical07/superapi/internal/core/httpx"
	"github.com/MrEthical07/superapi/internal/core/policy"
	"github.com/MrEthical07/superapi/internal/core/ratelimit"
	"github.com/MrEthical07/superapi/internal/core/rbac"
)

// Register mounts project routes.
func (m *Module) Register(r httpx.Router) error {
	if m.handler == nil {
		repo := NewRepo(m.runtime.RelationalStore())
		m.handler = NewHandler(NewService(m.runtime.RelationalStore(), repo))
	}

	resolver := m.runtime.PermissionResolver()
	if resolver == nil {
		return fmt.Errorf("project module requires permission resolver")
	}

	limiter := m.runtime.Limiter()
	cacheMgr := m.runtime.CacheManager()

	if cacheMgr != nil {
		r.Handle(
			http.MethodGet,
			"/api/v1/projects/{projectId}/dashboard/summary",
			httpx.Adapter(m.handler.DashboardSummary),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.ProjectRequired(),
			policy.ProjectMatchFromPath("projectId"),
			policy.ResolvePermissions(resolver),
			policy.RequirePermission(rbac.PermProjectView),
			policy.CacheRead(cacheMgr, cache.CacheReadConfig{
				TTL: 20 * time.Second,
				TagSpecs: []cache.CacheTagSpec{
					{Name: "project.dashboard.summary", ProjectID: true},
				},
				AllowAuthenticated: true,
				VaryBy:             cache.CacheVaryBy{ProjectID: true, UserID: true},
			}),
			policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 20 * time.Second, Vary: []string{"Authorization"}}),
		)
		r.Handle(
			http.MethodGet,
			"/api/v1/projects/{projectId}/dashboard/my-work",
			httpx.Adapter(m.handler.DashboardMyWork),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.ProjectRequired(),
			policy.ProjectMatchFromPath("projectId"),
			policy.ResolvePermissions(resolver),
			policy.RequirePermission(rbac.PermProjectView),
			policy.CacheRead(cacheMgr, cache.CacheReadConfig{
				TTL: 15 * time.Second,
				TagSpecs: []cache.CacheTagSpec{
					{Name: "project.dashboard.my_work", ProjectID: true},
				},
				AllowAuthenticated: true,
				VaryBy:             cache.CacheVaryBy{ProjectID: true, UserID: true},
			}),
			policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 15 * time.Second, Vary: []string{"Authorization"}}),
		)
		r.Handle(
			http.MethodGet,
			"/api/v1/projects/{projectId}/dashboard/events",
			httpx.Adapter(m.handler.DashboardEvents),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.ProjectRequired(),
			policy.ProjectMatchFromPath("projectId"),
			policy.ResolvePermissions(resolver),
			policy.RequirePermission(rbac.PermProjectView),
			policy.CacheRead(cacheMgr, cache.CacheReadConfig{
				TTL: 15 * time.Second,
				TagSpecs: []cache.CacheTagSpec{
					{Name: "project.dashboard.events", ProjectID: true},
				},
				AllowAuthenticated: true,
				VaryBy:             cache.CacheVaryBy{ProjectID: true, UserID: true},
			}),
			policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 15 * time.Second, Vary: []string{"Authorization"}}),
		)
		r.Handle(
			http.MethodGet,
			"/api/v1/projects/{projectId}/dashboard/activity",
			httpx.Adapter(m.handler.DashboardActivity),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.ProjectRequired(),
			policy.ProjectMatchFromPath("projectId"),
			policy.ResolvePermissions(resolver),
			policy.RequirePermission(rbac.PermProjectView),
			policy.CacheRead(cacheMgr, cache.CacheReadConfig{
				TTL: 15 * time.Second,
				TagSpecs: []cache.CacheTagSpec{
					{Name: "project.dashboard.activity", ProjectID: true},
				},
				AllowAuthenticated: true,
				VaryBy:             cache.CacheVaryBy{ProjectID: true, UserID: true},
			}),
			policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 15 * time.Second, Vary: []string{"Authorization"}}),
		)
		r.Handle(
			http.MethodGet,
			"/api/v1/projects/{projectId}/dashboard",
			httpx.Adapter(m.handler.Dashboard),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.ProjectRequired(),
			policy.ProjectMatchFromPath("projectId"),
			policy.ResolvePermissions(resolver),
			policy.RequirePermission(rbac.PermProjectView),
			policy.CacheRead(cacheMgr, cache.CacheReadConfig{
				TTL: 30 * time.Second,
				TagSpecs: []cache.CacheTagSpec{
					{Name: "project.dashboard", ProjectID: true},
				},
				AllowAuthenticated: true,
				VaryBy:             cache.CacheVaryBy{ProjectID: true, UserID: true},
			}),
			policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 30 * time.Second, Vary: []string{"Authorization"}}),
		)
		r.Handle(
			http.MethodGet,
			"/api/v1/projects/{projectId}/access",
			httpx.Adapter(m.handler.Access),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.ProjectRequired(),
			policy.ProjectMatchFromPath("projectId"),
			policy.ResolvePermissions(resolver),
			policy.CacheRead(cacheMgr, cache.CacheReadConfig{
				TTL: 30 * time.Second,
				TagSpecs: []cache.CacheTagSpec{
					{Name: "project.access", ProjectID: true},
				},
				AllowAuthenticated: true,
				VaryBy:             cache.CacheVaryBy{ProjectID: true, UserID: true},
			}),
			policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 30 * time.Second, Vary: []string{"Authorization"}}),
		)
		r.Handle(
			http.MethodGet,
			"/api/v1/projects/{projectId}/sidebar",
			httpx.Adapter(m.handler.Sidebar),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.ProjectRequired(),
			policy.ProjectMatchFromPath("projectId"),
			policy.ResolvePermissions(resolver),
			policy.RequirePermission(rbac.PermProjectView),
			policy.CacheRead(cacheMgr, cache.CacheReadConfig{
				TTL: 30 * time.Second,
				TagSpecs: []cache.CacheTagSpec{
					{Name: "project.sidebar", ProjectID: true},
				},
				AllowAuthenticated: true,
				VaryBy:             cache.CacheVaryBy{ProjectID: true, UserID: true},
			}),
			policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 30 * time.Second, Vary: []string{"Authorization"}}),
		)
		r.Handle(
			http.MethodGet,
			"/api/v1/projects/{projectId}/settings",
			httpx.Adapter(m.handler.GetSettings),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.ProjectRequired(),
			policy.ProjectMatchFromPath("projectId"),
			policy.ResolvePermissions(resolver),
			policy.RequirePermission(rbac.PermProjectView),
			policy.CacheRead(cacheMgr, cache.CacheReadConfig{
				TTL: 60 * time.Second,
				TagSpecs: []cache.CacheTagSpec{
					{Name: "project.settings", ProjectID: true},
				},
				AllowAuthenticated: true,
				VaryBy:             cache.CacheVaryBy{ProjectID: true},
			}),
			policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 60 * time.Second, Vary: []string{"Authorization"}}),
		)
	} else {
		r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/dashboard/summary", httpx.Adapter(m.handler.DashboardSummary), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermProjectView))
		r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/dashboard/my-work", httpx.Adapter(m.handler.DashboardMyWork), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermProjectView))
		r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/dashboard/events", httpx.Adapter(m.handler.DashboardEvents), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermProjectView))
		r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/dashboard/activity", httpx.Adapter(m.handler.DashboardActivity), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermProjectView))
		r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/dashboard", httpx.Adapter(m.handler.Dashboard), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermProjectView))
		r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/access", httpx.Adapter(m.handler.Access), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver))
		r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/sidebar", httpx.Adapter(m.handler.Sidebar), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermProjectView))
		r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/settings", httpx.Adapter(m.handler.GetSettings), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermProjectView))
	}

	settingsRule := ratelimit.Rule{Limit: 20, Window: time.Minute, Scope: ratelimit.ScopeProject}
	archiveRule := ratelimit.Rule{Limit: 10, Window: time.Minute, Scope: ratelimit.ScopeProject}
	deleteRule := ratelimit.Rule{Limit: 5, Window: time.Minute, Scope: ratelimit.ScopeProject}
	invalidateProjectTags := cache.CacheInvalidateConfig{TagSpecs: []cache.CacheTagSpec{
		{Name: "project.dashboard", ProjectID: true},
		{Name: "project.dashboard.summary", ProjectID: true},
		{Name: "project.dashboard.my_work", ProjectID: true},
		{Name: "project.dashboard.events", ProjectID: true},
		{Name: "project.dashboard.activity", ProjectID: true},
		{Name: "project.access", ProjectID: true},
		{Name: "project.sidebar", ProjectID: true},
		{Name: "project.settings", ProjectID: true},
	}}

	if limiter != nil && cacheMgr != nil {
		r.Handle(
			http.MethodPatch,
			"/api/v1/projects/{projectId}/settings",
			httpx.Adapter(m.handler.UpdateSettings),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.ProjectRequired(),
			policy.ProjectMatchFromPath("projectId"),
			policy.ResolvePermissions(resolver),
			policy.RequirePermission(rbac.PermProjectEdit),
			policy.RequireJSON(),
			policy.RateLimit(limiter, settingsRule),
			policy.CacheInvalidate(cacheMgr, invalidateProjectTags),
		)
		r.Handle(
			http.MethodPost,
			"/api/v1/projects/{projectId}/archive",
			httpx.Adapter(m.handler.Archive),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.ProjectRequired(),
			policy.ProjectMatchFromPath("projectId"),
			policy.ResolvePermissions(resolver),
			policy.RequirePermission(rbac.PermProjectArchive),
			policy.RateLimit(limiter, archiveRule),
			policy.CacheInvalidate(cacheMgr, invalidateProjectTags),
		)
		r.Handle(
			http.MethodDelete,
			"/api/v1/projects/{projectId}",
			httpx.Adapter(m.handler.Delete),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.ProjectRequired(),
			policy.ProjectMatchFromPath("projectId"),
			policy.ResolvePermissions(resolver),
			policy.RequirePermission(rbac.PermProjectDelete),
			policy.RateLimit(limiter, deleteRule),
			policy.CacheInvalidate(cacheMgr, invalidateProjectTags),
		)
		return nil
	}

	if limiter != nil {
		r.Handle(http.MethodPatch, "/api/v1/projects/{projectId}/settings", httpx.Adapter(m.handler.UpdateSettings), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermProjectEdit), policy.RequireJSON(), policy.RateLimit(limiter, settingsRule))
		r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/archive", httpx.Adapter(m.handler.Archive), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermProjectArchive), policy.RateLimit(limiter, archiveRule))
		r.Handle(http.MethodDelete, "/api/v1/projects/{projectId}", httpx.Adapter(m.handler.Delete), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermProjectDelete), policy.RateLimit(limiter, deleteRule))
		return nil
	}

	if cacheMgr != nil {
		r.Handle(http.MethodPatch, "/api/v1/projects/{projectId}/settings", httpx.Adapter(m.handler.UpdateSettings), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermProjectEdit), policy.RequireJSON(), policy.CacheInvalidate(cacheMgr, invalidateProjectTags))
		r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/archive", httpx.Adapter(m.handler.Archive), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermProjectArchive), policy.CacheInvalidate(cacheMgr, invalidateProjectTags))
		r.Handle(http.MethodDelete, "/api/v1/projects/{projectId}", httpx.Adapter(m.handler.Delete), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermProjectDelete), policy.CacheInvalidate(cacheMgr, invalidateProjectTags))
		return nil
	}

	r.Handle(http.MethodPatch, "/api/v1/projects/{projectId}/settings", httpx.Adapter(m.handler.UpdateSettings), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermProjectEdit), policy.RequireJSON())
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/archive", httpx.Adapter(m.handler.Archive), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermProjectArchive))
	r.Handle(http.MethodDelete, "/api/v1/projects/{projectId}", httpx.Adapter(m.handler.Delete), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermProjectDelete))

	return nil
}
