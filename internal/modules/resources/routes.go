package resources

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
		return fmt.Errorf("resources module requires permission resolver")
	}

	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/resources", httpx.Adapter(m.handler.ListResources),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermResourceView),
	)
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/resources", httpx.Adapter(m.handler.CreateResource),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermResourceCreate),
	)
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/resources/{resourceId}", httpx.Adapter(m.handler.GetResource),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermResourceView),
	)
	r.Handle(http.MethodPut, "/api/v1/projects/{projectId}/resources/{resourceId}", httpx.Adapter(m.handler.UpdateResource),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermResourceEdit),
	)
	r.Handle(http.MethodPut, "/api/v1/projects/{projectId}/resources/{resourceId}/status", httpx.Adapter(m.handler.UpdateResourceStatus),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermResourceStatusChange),
	)

	return nil
}
