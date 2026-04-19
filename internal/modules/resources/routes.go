package resources

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
		return fmt.Errorf("resources module requires permission resolver")
	}

	cacheMgr := m.runtime.CacheManager()
	limiter := m.runtime.Limiter()
	resourceReadTags := []cache.CacheTagSpec{
		{Name: "resources.project", PathParams: []string{"projectId"}},
		{Name: "resources.resource", PathParams: []string{"projectId"}},
	}
	resourceInvalidate := cache.CacheInvalidateConfig{TagSpecs: resourceReadTags}
	writeRule := ratelimit.Rule{Limit: 40, Window: time.Minute, Scope: ratelimit.ScopeProject}
	listTTL := 30 * time.Second
	detailTTL := 90 * time.Second
	payloadMaxBytes := 100 * 1024

	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/resources", httpx.Adapter(m.handler.ListResources),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermResourceView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                listTTL,
			MaxBytes:           payloadMaxBytes,
			TagSpecs:           resourceReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:   true,
				PathParams:  []string{"projectId"},
				QueryParams: []string{"filter.status", "status", "filter.docType", "docType", "sorting.sort", "sort", "sorting.order", "order", "pagination.cursor", "cursor", "pagination.limit", "limit"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, listTTL),
	)
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/resources", httpx.Adapter(m.handler.CreateResource),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermResourceCreate),
		policy.RateLimitOptional(limiter, writeRule),
		policy.CacheInvalidateOptional(cacheMgr, resourceInvalidate),
	)
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/resources/{resourceId}", httpx.Adapter(m.handler.GetResource),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermResourceView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                detailTTL,
			MaxBytes:           payloadMaxBytes,
			TagSpecs:           resourceReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:  true,
				PathParams: []string{"projectId", "resourceId"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, detailTTL),
	)
	r.Handle(http.MethodPatch, "/api/v1/projects/{projectId}/resources/{resourceId}", httpx.Adapter(m.handler.UpdateResource),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermResourceEdit),
		policy.RateLimitOptional(limiter, writeRule),
		policy.CacheInvalidateOptional(cacheMgr, resourceInvalidate),
	)

	return nil
}
