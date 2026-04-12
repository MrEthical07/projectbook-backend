package home

import (
	"net/http"
	"time"

	"github.com/MrEthical07/superapi/internal/core/cache"
	"github.com/MrEthical07/superapi/internal/core/httpx"
	"github.com/MrEthical07/superapi/internal/core/policy"
	"github.com/MrEthical07/superapi/internal/core/ratelimit"
)

// Register mounts home routes.
func (m *Module) Register(r httpx.Router) error {
	if m.handler == nil {
		repo := NewRepo(m.runtime.RelationalStore())
		m.handler = NewHandler(NewService(m.runtime.RelationalStore(), repo))
	}

	limiter := m.runtime.Limiter()
	cacheMgr := m.runtime.CacheManager()

	if cacheMgr != nil {
		r.Handle(
			http.MethodGet,
			"/api/v1/home/dashboard",
			httpx.Adapter(m.handler.Dashboard),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.CacheRead(cacheMgr, cache.CacheReadConfig{
				TTL: 30 * time.Second,
				TagSpecs: []cache.CacheTagSpec{
					{Name: "home.dashboard", UserID: true},
				},
				AllowAuthenticated: true,
				VaryBy:             cache.CacheVaryBy{UserID: true},
			}),
			policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 30 * time.Second, Vary: []string{"Authorization"}}),
		)
		r.Handle(
			http.MethodGet,
			"/api/v1/home/projects",
			httpx.Adapter(m.handler.ListProjects),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.CacheRead(cacheMgr, cache.CacheReadConfig{
				TTL: 30 * time.Second,
				TagSpecs: []cache.CacheTagSpec{
					{Name: "home.projects", UserID: true},
				},
				AllowAuthenticated: true,
				VaryBy:             cache.CacheVaryBy{UserID: true, QueryParams: []string{"limit", "offset"}},
			}),
			policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 30 * time.Second, Vary: []string{"Authorization"}}),
		)
		r.Handle(
			http.MethodGet,
			"/api/v1/home/projects/reference",
			httpx.Adapter(m.handler.ProjectReference),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.CacheRead(cacheMgr, cache.CacheReadConfig{
				TTL: 60 * time.Second,
				TagSpecs: []cache.CacheTagSpec{
					{Name: "home.reference", UserID: true},
				},
				AllowAuthenticated: true,
				VaryBy:             cache.CacheVaryBy{UserID: true},
			}),
			policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 60 * time.Second, Vary: []string{"Authorization"}}),
		)
		r.Handle(
			http.MethodGet,
			"/api/v1/home/invites",
			httpx.Adapter(m.handler.ListInvites),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.CacheRead(cacheMgr, cache.CacheReadConfig{
				TTL: 20 * time.Second,
				TagSpecs: []cache.CacheTagSpec{
					{Name: "home.invites", UserID: true},
				},
				AllowAuthenticated: true,
				VaryBy:             cache.CacheVaryBy{UserID: true},
			}),
			policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 20 * time.Second, Vary: []string{"Authorization"}}),
		)
		r.Handle(
			http.MethodGet,
			"/api/v1/home/notifications",
			httpx.Adapter(m.handler.ListNotifications),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.CacheRead(cacheMgr, cache.CacheReadConfig{
				TTL: 15 * time.Second,
				TagSpecs: []cache.CacheTagSpec{
					{Name: "home.notifications", UserID: true},
				},
				AllowAuthenticated: true,
				VaryBy:             cache.CacheVaryBy{UserID: true, QueryParams: []string{"limit"}},
			}),
			policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 15 * time.Second, Vary: []string{"Authorization"}}),
		)
		r.Handle(
			http.MethodGet,
			"/api/v1/home/activity",
			httpx.Adapter(m.handler.ListActivity),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.CacheRead(cacheMgr, cache.CacheReadConfig{
				TTL: 15 * time.Second,
				TagSpecs: []cache.CacheTagSpec{
					{Name: "home.activity", UserID: true},
				},
				AllowAuthenticated: true,
				VaryBy:             cache.CacheVaryBy{UserID: true, QueryParams: []string{"limit", "type", "projectId"}},
			}),
			policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 15 * time.Second, Vary: []string{"Authorization"}}),
		)
		r.Handle(
			http.MethodGet,
			"/api/v1/home/dashboard-activity",
			httpx.Adapter(m.handler.DashboardActivity),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.CacheRead(cacheMgr, cache.CacheReadConfig{
				TTL: 15 * time.Second,
				TagSpecs: []cache.CacheTagSpec{
					{Name: "home.dashboard_activity", UserID: true},
				},
				AllowAuthenticated: true,
				VaryBy:             cache.CacheVaryBy{UserID: true, QueryParams: []string{"limit"}},
			}),
			policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 15 * time.Second, Vary: []string{"Authorization"}}),
		)
		r.Handle(
			http.MethodGet,
			"/api/v1/home/account",
			httpx.Adapter(m.handler.GetAccountSettings),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.CacheRead(cacheMgr, cache.CacheReadConfig{
				TTL: 60 * time.Second,
				TagSpecs: []cache.CacheTagSpec{
					{Name: "home.account", UserID: true},
				},
				AllowAuthenticated: true,
				VaryBy:             cache.CacheVaryBy{UserID: true},
			}),
			policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 60 * time.Second, Vary: []string{"Authorization"}}),
		)
		r.Handle(
			http.MethodGet,
			"/api/v1/home/docs",
			httpx.Adapter(m.handler.Docs),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.CacheRead(cacheMgr, cache.CacheReadConfig{
				TTL: 5 * time.Minute,
				TagSpecs: []cache.CacheTagSpec{
					{Name: "home.docs", UserID: true},
				},
				AllowAuthenticated: true,
				VaryBy:             cache.CacheVaryBy{UserID: true},
			}),
			policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 5 * time.Minute, Vary: []string{"Authorization"}}),
		)
	} else {
		r.Handle(http.MethodGet, "/api/v1/home/dashboard", httpx.Adapter(m.handler.Dashboard), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()))
		r.Handle(http.MethodGet, "/api/v1/home/projects", httpx.Adapter(m.handler.ListProjects), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()))
		r.Handle(http.MethodGet, "/api/v1/home/projects/reference", httpx.Adapter(m.handler.ProjectReference), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()))
		r.Handle(http.MethodGet, "/api/v1/home/invites", httpx.Adapter(m.handler.ListInvites), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()))
		r.Handle(http.MethodGet, "/api/v1/home/notifications", httpx.Adapter(m.handler.ListNotifications), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()))
		r.Handle(http.MethodGet, "/api/v1/home/activity", httpx.Adapter(m.handler.ListActivity), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()))
		r.Handle(http.MethodGet, "/api/v1/home/dashboard-activity", httpx.Adapter(m.handler.DashboardActivity), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()))
		r.Handle(http.MethodGet, "/api/v1/home/account", httpx.Adapter(m.handler.GetAccountSettings), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()))
		r.Handle(http.MethodGet, "/api/v1/home/docs", httpx.Adapter(m.handler.Docs), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()))
	}

	createProjectRule := ratelimit.Rule{Limit: 10, Window: time.Minute, Scope: ratelimit.ScopeUser}
	inviteActionRule := ratelimit.Rule{Limit: 30, Window: time.Minute, Scope: ratelimit.ScopeUser}
	accountUpdateRule := ratelimit.Rule{Limit: 20, Window: time.Minute, Scope: ratelimit.ScopeUser}

	if limiter != nil && cacheMgr != nil {
		r.Handle(
			http.MethodPost,
			"/api/v1/home/projects",
			httpx.Adapter(m.handler.CreateProject),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.RequireJSON(),
			policy.RateLimit(limiter, createProjectRule),
			policy.CacheInvalidate(cacheMgr, cache.CacheInvalidateConfig{TagSpecs: []cache.CacheTagSpec{
				{Name: "home.dashboard", UserID: true},
				{Name: "home.projects", UserID: true},
				{Name: "home.reference", UserID: true},
			}}),
		)
		r.Handle(
			http.MethodPost,
			"/api/v1/home/invites/{inviteId}/accept",
			httpx.Adapter(m.handler.AcceptInvite),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.RateLimit(limiter, inviteActionRule),
			policy.CacheInvalidate(cacheMgr, cache.CacheInvalidateConfig{TagSpecs: []cache.CacheTagSpec{
				{Name: "home.dashboard", UserID: true},
				{Name: "home.invites", UserID: true},
				{Name: "home.projects", UserID: true},
			}}),
		)
		r.Handle(
			http.MethodPost,
			"/api/v1/home/invites/{inviteId}/decline",
			httpx.Adapter(m.handler.DeclineInvite),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.RateLimit(limiter, inviteActionRule),
			policy.CacheInvalidate(cacheMgr, cache.CacheInvalidateConfig{TagSpecs: []cache.CacheTagSpec{
				{Name: "home.invites", UserID: true},
				{Name: "home.dashboard", UserID: true},
			}}),
		)
		r.Handle(
			http.MethodPut,
			"/api/v1/home/account",
			httpx.Adapter(m.handler.UpdateAccountSettings),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.RequireJSON(),
			policy.RateLimit(limiter, accountUpdateRule),
			policy.CacheInvalidate(cacheMgr, cache.CacheInvalidateConfig{TagSpecs: []cache.CacheTagSpec{
				{Name: "home.account", UserID: true},
				{Name: "home.dashboard", UserID: true},
			}}),
		)
		return nil
	}

	if limiter != nil {
		r.Handle(http.MethodPost, "/api/v1/home/projects", httpx.Adapter(m.handler.CreateProject), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.RequireJSON(), policy.RateLimit(limiter, createProjectRule))
		r.Handle(http.MethodPost, "/api/v1/home/invites/{inviteId}/accept", httpx.Adapter(m.handler.AcceptInvite), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.RateLimit(limiter, inviteActionRule))
		r.Handle(http.MethodPost, "/api/v1/home/invites/{inviteId}/decline", httpx.Adapter(m.handler.DeclineInvite), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.RateLimit(limiter, inviteActionRule))
		r.Handle(http.MethodPut, "/api/v1/home/account", httpx.Adapter(m.handler.UpdateAccountSettings), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.RequireJSON(), policy.RateLimit(limiter, accountUpdateRule))
		return nil
	}

	if cacheMgr != nil {
		r.Handle(http.MethodPost, "/api/v1/home/projects", httpx.Adapter(m.handler.CreateProject), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.RequireJSON(), policy.CacheInvalidate(cacheMgr, cache.CacheInvalidateConfig{TagSpecs: []cache.CacheTagSpec{{Name: "home.dashboard", UserID: true}, {Name: "home.projects", UserID: true}, {Name: "home.reference", UserID: true}}}))
		r.Handle(http.MethodPost, "/api/v1/home/invites/{inviteId}/accept", httpx.Adapter(m.handler.AcceptInvite), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.CacheInvalidate(cacheMgr, cache.CacheInvalidateConfig{TagSpecs: []cache.CacheTagSpec{{Name: "home.dashboard", UserID: true}, {Name: "home.invites", UserID: true}, {Name: "home.projects", UserID: true}}}))
		r.Handle(http.MethodPost, "/api/v1/home/invites/{inviteId}/decline", httpx.Adapter(m.handler.DeclineInvite), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.CacheInvalidate(cacheMgr, cache.CacheInvalidateConfig{TagSpecs: []cache.CacheTagSpec{{Name: "home.invites", UserID: true}, {Name: "home.dashboard", UserID: true}}}))
		r.Handle(http.MethodPut, "/api/v1/home/account", httpx.Adapter(m.handler.UpdateAccountSettings), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.RequireJSON(), policy.CacheInvalidate(cacheMgr, cache.CacheInvalidateConfig{TagSpecs: []cache.CacheTagSpec{{Name: "home.account", UserID: true}, {Name: "home.dashboard", UserID: true}}}))
		return nil
	}

	r.Handle(http.MethodPost, "/api/v1/home/projects", httpx.Adapter(m.handler.CreateProject), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.RequireJSON())
	r.Handle(http.MethodPost, "/api/v1/home/invites/{inviteId}/accept", httpx.Adapter(m.handler.AcceptInvite), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()))
	r.Handle(http.MethodPost, "/api/v1/home/invites/{inviteId}/decline", httpx.Adapter(m.handler.DeclineInvite), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()))
	r.Handle(http.MethodPut, "/api/v1/home/account", httpx.Adapter(m.handler.UpdateAccountSettings), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.RequireJSON())

	return nil
}
