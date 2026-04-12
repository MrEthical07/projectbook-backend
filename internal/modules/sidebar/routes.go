package sidebar

import (
	"fmt"
	"net/http"

	"github.com/MrEthical07/superapi/internal/core/httpx"
	"github.com/MrEthical07/superapi/internal/core/policy"
	"github.com/MrEthical07/superapi/internal/core/rbac"
)

func (m *Module) Register(r httpx.Router) error {
	if m.handler == nil {
		m.BindDependencies(m.runtime.Dependencies())
	}

	resolver := m.runtime.PermissionResolver()
	if resolver == nil {
		return fmt.Errorf("sidebar module requires permission resolver")
	}

	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/sidebar/artifacts", httpx.Adapter(m.handler.CreateSidebarArtifact),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequireAnyPermission(
			rbac.PermStoryCreate,
			rbac.PermProblemCreate,
			rbac.PermIdeaCreate,
			rbac.PermTaskCreate,
			rbac.PermFeedbackCreate,
			rbac.PermPageCreate,
		),
	)
	r.Handle(http.MethodPut, "/api/v1/projects/{projectId}/sidebar/artifacts/{artifactId}/rename", httpx.Adapter(m.handler.RenameSidebarArtifact),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequireAnyPermission(
			rbac.PermStoryEdit,
			rbac.PermProblemEdit,
			rbac.PermIdeaEdit,
			rbac.PermTaskEdit,
			rbac.PermFeedbackEdit,
			rbac.PermPageEdit,
		),
	)
	r.Handle(http.MethodDelete, "/api/v1/projects/{projectId}/sidebar/artifacts/{artifactId}", httpx.Adapter(m.handler.DeleteSidebarArtifact),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequireAnyPermission(
			rbac.PermStoryDelete,
			rbac.PermProblemDelete,
			rbac.PermIdeaDelete,
			rbac.PermTaskDelete,
			rbac.PermFeedbackDelete,
			rbac.PermPageDelete,
		),
	)

	return nil
}
