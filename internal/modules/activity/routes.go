package activity

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
		m.handler = NewHandler(NewService(repo))
	}

	resolver := m.runtime.PermissionResolver()
	if resolver == nil {
		return fmt.Errorf("activity module requires permission resolver")
	}

	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/activity", httpx.Adapter(m.handler.ListProjectActivity),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermProjectView),
	)

	return nil
}
