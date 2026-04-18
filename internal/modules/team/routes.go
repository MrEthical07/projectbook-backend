package team

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

// Register mounts team routes.
func (m *Module) Register(r httpx.Router) error {
	if m.handler == nil {
		repo := NewRepo(m.runtime.RelationalStore())
		m.handler = NewHandler(NewService(m.runtime.RelationalStore(), repo, m.runtime.Redis(), m.runtime.CacheManager()))
	}

	resolver := m.runtime.PermissionResolver()
	if resolver == nil {
		return fmt.Errorf("team module requires permission resolver")
	}

	limiter := m.runtime.Limiter()
	cacheMgr := m.runtime.CacheManager()

	if cacheMgr != nil {
		r.Handle(
			http.MethodGet,
			"/api/v1/projects/{projectId}/team/members",
			httpx.Adapter(m.handler.ListMembers),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.ProjectRequired(),
			policy.ProjectMatchFromPath("projectId"),
			policy.ResolvePermissions(resolver),
			policy.RequirePermission(rbac.PermMemberView),
			policy.CacheRead(cacheMgr, cache.CacheReadConfig{
				TTL: 30 * time.Second,
				TagSpecs: []cache.CacheTagSpec{
					{Name: "team.members", ProjectID: true},
				},
				AllowAuthenticated: true,
				VaryBy:             cache.CacheVaryBy{ProjectID: true},
			}),
			policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 30 * time.Second, Vary: []string{"Authorization"}}),
		)
		r.Handle(
			http.MethodGet,
			"/api/v1/projects/{projectId}/team/roles",
			httpx.Adapter(m.handler.ListRoles),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.ProjectRequired(),
			policy.ProjectMatchFromPath("projectId"),
			policy.ResolvePermissions(resolver),
			policy.RequirePermission(rbac.PermMemberView),
			policy.CacheRead(cacheMgr, cache.CacheReadConfig{
				TTL: 30 * time.Second,
				TagSpecs: []cache.CacheTagSpec{
					{Name: "team.roles", ProjectID: true},
				},
				AllowAuthenticated: true,
				VaryBy:             cache.CacheVaryBy{ProjectID: true},
			}),
			policy.CacheControl(policy.CacheControlConfig{Private: true, MaxAge: 30 * time.Second, Vary: []string{"Authorization"}}),
		)
	} else {
		r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/team/members", httpx.Adapter(m.handler.ListMembers), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermMemberView))
		r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/team/roles", httpx.Adapter(m.handler.ListRoles), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermMemberView))
	}

	inviteRule := ratelimit.Rule{Limit: 30, Window: time.Minute, Scope: ratelimit.ScopeProject}
	batchInviteRule := ratelimit.Rule{Limit: 10, Window: time.Minute, Scope: ratelimit.ScopeProject}
	cancelInviteRule := ratelimit.Rule{Limit: 30, Window: time.Minute, Scope: ratelimit.ScopeProject}
	memberUpdateRule := ratelimit.Rule{Limit: 20, Window: time.Minute, Scope: ratelimit.ScopeProject}
	roleUpdateRule := ratelimit.Rule{Limit: 20, Window: time.Minute, Scope: ratelimit.ScopeProject}

	invalidateTeamTags := cache.CacheInvalidateConfig{TagSpecs: []cache.CacheTagSpec{
		{Name: "team.members", ProjectID: true},
		{Name: "team.roles", ProjectID: true},
	}}

	if limiter != nil && cacheMgr != nil {
		r.Handle(
			http.MethodPost,
			"/api/v1/projects/{projectId}/team/invites",
			httpx.Adapter(m.handler.CreateInvite),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.ProjectRequired(),
			policy.ProjectMatchFromPath("projectId"),
			policy.ResolvePermissions(resolver),
			policy.RequirePermission(rbac.PermMemberCreate),
			policy.RequireJSON(),
			policy.RateLimit(limiter, inviteRule),
			policy.CacheInvalidate(cacheMgr, invalidateTeamTags),
		)
		r.Handle(
			http.MethodPost,
			"/api/v1/projects/{projectId}/team/invites/batch",
			httpx.Adapter(m.handler.BatchInvites),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.ProjectRequired(),
			policy.ProjectMatchFromPath("projectId"),
			policy.ResolvePermissions(resolver),
			policy.RequirePermission(rbac.PermMemberCreate),
			policy.RequireJSON(),
			policy.RateLimit(limiter, batchInviteRule),
			policy.CacheInvalidate(cacheMgr, invalidateTeamTags),
		)
		r.Handle(
			http.MethodDelete,
			"/api/v1/projects/{projectId}/team/invites/{email}",
			httpx.Adapter(m.handler.CancelInvite),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.ProjectRequired(),
			policy.ProjectMatchFromPath("projectId"),
			policy.ResolvePermissions(resolver),
			policy.RequirePermission(rbac.PermMemberDelete),
			policy.RateLimit(limiter, cancelInviteRule),
			policy.CacheInvalidate(cacheMgr, invalidateTeamTags),
		)
		r.Handle(
			http.MethodPatch,
			"/api/v1/projects/{projectId}/team/members/{memberId}/permissions",
			httpx.Adapter(m.handler.UpdateMemberPermissions),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.ProjectRequired(),
			policy.ProjectMatchFromPath("projectId"),
			policy.ResolvePermissions(resolver),
			policy.RequirePermission(rbac.PermMemberEdit),
			policy.RequireJSON(),
			policy.RateLimit(limiter, memberUpdateRule),
			policy.CacheInvalidate(cacheMgr, invalidateTeamTags),
		)
		r.Handle(
			http.MethodPatch,
			"/api/v1/projects/{projectId}/team/roles/{role}/permissions",
			httpx.Adapter(m.handler.UpdateRolePermissions),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.ProjectRequired(),
			policy.ProjectMatchFromPath("projectId"),
			policy.ResolvePermissions(resolver),
			policy.RequirePermission(rbac.PermMemberEdit),
			policy.RequireJSON(),
			policy.RateLimit(limiter, roleUpdateRule),
			policy.CacheInvalidate(cacheMgr, invalidateTeamTags),
		)
		return nil
	}

	if limiter != nil {
		r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/team/invites", httpx.Adapter(m.handler.CreateInvite), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermMemberCreate), policy.RequireJSON(), policy.RateLimit(limiter, inviteRule))
		r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/team/invites/batch", httpx.Adapter(m.handler.BatchInvites), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermMemberCreate), policy.RequireJSON(), policy.RateLimit(limiter, batchInviteRule))
		r.Handle(http.MethodDelete, "/api/v1/projects/{projectId}/team/invites/{email}", httpx.Adapter(m.handler.CancelInvite), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermMemberDelete), policy.RateLimit(limiter, cancelInviteRule))
		r.Handle(http.MethodPatch, "/api/v1/projects/{projectId}/team/members/{memberId}/permissions", httpx.Adapter(m.handler.UpdateMemberPermissions), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermMemberEdit), policy.RequireJSON(), policy.RateLimit(limiter, memberUpdateRule))
		r.Handle(http.MethodPatch, "/api/v1/projects/{projectId}/team/roles/{role}/permissions", httpx.Adapter(m.handler.UpdateRolePermissions), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermMemberEdit), policy.RequireJSON(), policy.RateLimit(limiter, roleUpdateRule))
		return nil
	}

	if cacheMgr != nil {
		r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/team/invites", httpx.Adapter(m.handler.CreateInvite), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermMemberCreate), policy.RequireJSON(), policy.CacheInvalidate(cacheMgr, invalidateTeamTags))
		r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/team/invites/batch", httpx.Adapter(m.handler.BatchInvites), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermMemberCreate), policy.RequireJSON(), policy.CacheInvalidate(cacheMgr, invalidateTeamTags))
		r.Handle(http.MethodDelete, "/api/v1/projects/{projectId}/team/invites/{email}", httpx.Adapter(m.handler.CancelInvite), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermMemberDelete), policy.CacheInvalidate(cacheMgr, invalidateTeamTags))
		r.Handle(http.MethodPatch, "/api/v1/projects/{projectId}/team/members/{memberId}/permissions", httpx.Adapter(m.handler.UpdateMemberPermissions), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermMemberEdit), policy.RequireJSON(), policy.CacheInvalidate(cacheMgr, invalidateTeamTags))
		r.Handle(http.MethodPatch, "/api/v1/projects/{projectId}/team/roles/{role}/permissions", httpx.Adapter(m.handler.UpdateRolePermissions), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermMemberEdit), policy.RequireJSON(), policy.CacheInvalidate(cacheMgr, invalidateTeamTags))
		return nil
	}

	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/team/invites", httpx.Adapter(m.handler.CreateInvite), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermMemberCreate), policy.RequireJSON())
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/team/invites/batch", httpx.Adapter(m.handler.BatchInvites), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermMemberCreate), policy.RequireJSON())
	r.Handle(http.MethodDelete, "/api/v1/projects/{projectId}/team/invites/{email}", httpx.Adapter(m.handler.CancelInvite), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermMemberDelete))
	r.Handle(http.MethodPatch, "/api/v1/projects/{projectId}/team/members/{memberId}/permissions", httpx.Adapter(m.handler.UpdateMemberPermissions), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermMemberEdit), policy.RequireJSON())
	r.Handle(http.MethodPatch, "/api/v1/projects/{projectId}/team/roles/{role}/permissions", httpx.Adapter(m.handler.UpdateRolePermissions), policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()), policy.ProjectRequired(), policy.ProjectMatchFromPath("projectId"), policy.ResolvePermissions(resolver), policy.RequirePermission(rbac.PermMemberEdit), policy.RequireJSON())

	return nil
}
