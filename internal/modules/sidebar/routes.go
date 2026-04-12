package sidebar

import (
	"fmt"
	"net/http"

	"github.com/MrEthical07/superapi/internal/core/cache"
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

	cacheMgr := m.runtime.CacheManager()
	invalidateTags := cache.CacheInvalidateConfig{TagSpecs: []cache.CacheTagSpec{
		{Name: "artifacts.project", PathParams: []string{"projectId"}},
		{Name: "artifacts.story", PathParams: []string{"projectId"}},
		{Name: "artifacts.journey", PathParams: []string{"projectId"}},
		{Name: "artifacts.problem", PathParams: []string{"projectId"}},
		{Name: "artifacts.idea", PathParams: []string{"projectId"}},
		{Name: "artifacts.task", PathParams: []string{"projectId"}},
		{Name: "artifacts.feedback", PathParams: []string{"projectId"}},
		{Name: "pages.project", PathParams: []string{"projectId"}},
		{Name: "pages.page", PathParams: []string{"projectId"}},
		{Name: "resources.project", PathParams: []string{"projectId"}},
		{Name: "resources.resource", PathParams: []string{"projectId"}},
	}}

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
		policy.CacheInvalidateOptional(cacheMgr, invalidateTags),
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
		policy.CacheInvalidateOptional(cacheMgr, invalidateTags),
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
		policy.CacheInvalidateOptional(cacheMgr, invalidateTags),
	)

	return nil
}
