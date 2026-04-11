package pages

import (
	"fmt"
	"net/http"

	"github.com/MrEthical07/superapi/internal/core/httpx"
	"github.com/MrEthical07/superapi/internal/core/policy"
	"github.com/MrEthical07/superapi/internal/core/rbac"
)

func (m *Module) Register(r httpx.Router) error {
	if m.handler == nil {
		repo := NewRepo(m.runtime.RelationalStore())
		m.handler = NewHandler(NewService(m.runtime.RelationalStore(), repo))
	}

	resolver := m.runtime.PermissionResolver()
	if resolver == nil {
		return fmt.Errorf("pages module requires permission resolver")
	}

	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/pages", httpx.Adapter(m.handler.ListPages),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermPageView),
	)
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/pages", httpx.Adapter(m.handler.CreatePage),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermPageCreate),
	)
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/pages/{slug}", httpx.Adapter(m.handler.GetPage),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermPageView),
	)
	r.Handle(http.MethodPut, "/api/v1/projects/{projectId}/pages/{pageId}", httpx.Adapter(m.handler.UpdatePage),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermPageEdit),
	)
	r.Handle(http.MethodPut, "/api/v1/projects/{projectId}/pages/{pageId}/rename", httpx.Adapter(m.handler.RenamePage),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermPageEdit),
	)

	return nil
}
