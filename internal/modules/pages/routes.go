package pages

import (
	"fmt"
	"net/http"
	"time"

	"github.com/MrEthical07/superapi/internal/core/cache"
	"github.com/MrEthical07/superapi/internal/core/httpx"
	"github.com/MrEthical07/superapi/internal/core/policy"
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
	pageReadTags := []cache.CacheTagSpec{
		{Name: "pages.project", PathParams: []string{"projectId"}},
		{Name: "pages.page", PathParams: []string{"projectId"}},
	}
	pageInvalidate := cache.CacheInvalidateConfig{TagSpecs: pageReadTags}

	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/pages", httpx.Adapter(m.handler.ListPages),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermPageView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                30 * time.Second,
			TagSpecs:           pageReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:   true,
				PathParams:  []string{"projectId"},
				QueryParams: []string{"status", "offset", "limit"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, 30*time.Second),
	)
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/pages", httpx.Adapter(m.handler.CreatePage),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermPageCreate),
		policy.CacheInvalidateOptional(cacheMgr, pageInvalidate),
	)
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/pages/{slug}", httpx.Adapter(m.handler.GetPage),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermPageView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                30 * time.Second,
			TagSpecs:           pageReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:  true,
				PathParams: []string{"projectId", "slug"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, 30*time.Second),
	)
	r.Handle(http.MethodPut, "/api/v1/projects/{projectId}/pages/{pageId}", httpx.Adapter(m.handler.UpdatePage),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermPageEdit),
		policy.CacheInvalidateOptional(cacheMgr, pageInvalidate),
	)
	r.Handle(http.MethodPut, "/api/v1/projects/{projectId}/pages/{pageId}/rename", httpx.Adapter(m.handler.RenamePage),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermPageEdit),
		policy.CacheInvalidateOptional(cacheMgr, pageInvalidate),
	)

	return nil
}
