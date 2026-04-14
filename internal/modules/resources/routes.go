package resources

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
		return fmt.Errorf("resources module requires permission resolver")
	}

	cacheMgr := m.runtime.CacheManager()
	resourceReadTags := []cache.CacheTagSpec{
		{Name: "resources.project", PathParams: []string{"projectId"}},
		{Name: "resources.resource", PathParams: []string{"projectId"}},
	}
	resourceInvalidate := cache.CacheInvalidateConfig{TagSpecs: resourceReadTags}

	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/resources", httpx.Adapter(m.handler.ListResources),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermResourceView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                30 * time.Second,
			TagSpecs:           resourceReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:   true,
				PathParams:  []string{"projectId"},
				QueryParams: []string{"status", "offset", "limit"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, 30*time.Second),
	)
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/resources", httpx.Adapter(m.handler.CreateResource),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermResourceCreate),
		policy.CacheInvalidateOptional(cacheMgr, resourceInvalidate),
	)
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/resources/{resourceId}", httpx.Adapter(m.handler.GetResource),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermResourceView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                30 * time.Second,
			TagSpecs:           resourceReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:  true,
				PathParams: []string{"projectId", "resourceId"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, 30*time.Second),
	)
	r.Handle(http.MethodPut, "/api/v1/projects/{projectId}/resources/{resourceId}", httpx.Adapter(m.handler.UpdateResource),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermResourceEdit),
		policy.CacheInvalidateOptional(cacheMgr, resourceInvalidate),
	)
	r.Handle(http.MethodPut, "/api/v1/projects/{projectId}/resources/{resourceId}/status", httpx.Adapter(m.handler.UpdateResourceStatus),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermResourceStatusChange),
		policy.CacheInvalidateOptional(cacheMgr, resourceInvalidate),
	)

	return nil
}
