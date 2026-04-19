package pages

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

func (m *Module) Register(r httpx.Router) error {
	if m.handler == nil {
		repo := NewRepo(m.runtime.RelationalStore(), m.runtime.DocumentStore())
		m.handler = NewHandler(NewService(m.runtime.RelationalStore(), repo))
	}

	resolver := m.runtime.PermissionResolver()
	if resolver == nil {
		return fmt.Errorf("pages module requires permission resolver")
	}

	cacheMgr := m.runtime.CacheManager()
	limiter := m.runtime.Limiter()
	pageReadTags := []cache.CacheTagSpec{
		{Name: "pages.project", PathParams: []string{"projectId"}},
		{Name: "pages.page", PathParams: []string{"projectId"}},
	}
	pageInvalidate := cache.CacheInvalidateConfig{TagSpecs: pageReadTags}
	writeRule := ratelimit.Rule{Limit: 40, Window: time.Minute, Scope: ratelimit.ScopeProject}
	listTTL := 30 * time.Second
	detailTTL := 90 * time.Second
	payloadMaxBytes := 100 * 1024

	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/pages", httpx.Adapter(m.handler.ListPages),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermPageView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                listTTL,
			MaxBytes:           payloadMaxBytes,
			TagSpecs:           pageReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:   true,
				PathParams:  []string{"projectId"},
				QueryParams: []string{"filter.status", "status", "pagination.cursor", "cursor", "pagination.limit", "limit"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, listTTL),
	)
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/pages", httpx.Adapter(m.handler.CreatePage),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermPageCreate),
		policy.RateLimitOptional(limiter, writeRule),
		policy.CacheInvalidateOptional(cacheMgr, pageInvalidate),
	)
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/pages/{pageId}", httpx.Adapter(m.handler.GetPage),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermPageView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                detailTTL,
			MaxBytes:           payloadMaxBytes,
			TagSpecs:           pageReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:  true,
				PathParams: []string{"projectId", "pageId"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, detailTTL),
	)
	r.Handle(http.MethodPatch, "/api/v1/projects/{projectId}/pages/{pageId}", httpx.Adapter(m.handler.UpdatePage),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermPageEdit),
		policy.RateLimitOptional(limiter, writeRule),
		policy.CacheInvalidateOptional(cacheMgr, pageInvalidate),
	)
	r.Handle(http.MethodPatch, "/api/v1/projects/{projectId}/pages/{pageId}/rename", httpx.Adapter(m.handler.RenamePage),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermPageEdit),
		policy.RateLimitOptional(limiter, writeRule),
		policy.CacheInvalidateOptional(cacheMgr, pageInvalidate),
	)

	return nil
}
