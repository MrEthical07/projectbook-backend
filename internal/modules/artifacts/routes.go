package artifacts

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
		return fmt.Errorf("artifacts module requires permission resolver")
	}

	// Stories
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/stories", httpx.Adapter(m.handler.ListStories),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermStoryView),
	)
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/stories", httpx.Adapter(m.handler.CreateStory),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermStoryCreate),
	)
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/stories/{slug}", httpx.Adapter(m.handler.GetStory),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermStoryView),
	)
	r.Handle(http.MethodPut, "/api/v1/projects/{projectId}/stories/{storyId}", httpx.Adapter(m.handler.UpdateStory),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermStoryEdit),
	)

	// Journeys share story permission scopes.
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/journeys", httpx.Adapter(m.handler.ListJourneys),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermStoryView),
	)
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/journeys", httpx.Adapter(m.handler.CreateJourney),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermStoryCreate),
	)
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/journeys/{slug}", httpx.Adapter(m.handler.GetJourney),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermStoryView),
	)
	r.Handle(http.MethodPut, "/api/v1/projects/{projectId}/journeys/{journeyId}", httpx.Adapter(m.handler.UpdateJourney),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermStoryEdit),
	)

	// Problems
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/problems", httpx.Adapter(m.handler.ListProblems),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermProblemView),
	)
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/problems", httpx.Adapter(m.handler.CreateProblem),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermProblemCreate),
	)
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/problems/{slug}", httpx.Adapter(m.handler.GetProblem),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermProblemView),
	)
	r.Handle(http.MethodPut, "/api/v1/projects/{projectId}/problems/{problemId}", httpx.Adapter(m.handler.UpdateProblem),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermProblemEdit),
	)
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/problems/{problemId}/lock", httpx.Adapter(m.handler.LockProblem),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermProblemStatusChange),
	)
	r.Handle(http.MethodPut, "/api/v1/projects/{projectId}/problems/{problemId}/status", httpx.Adapter(m.handler.UpdateProblemStatus),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermProblemStatusChange),
	)

	// Ideas
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/ideas", httpx.Adapter(m.handler.ListIdeas),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermIdeaView),
	)
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/ideas", httpx.Adapter(m.handler.CreateIdea),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermIdeaCreate),
	)
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/ideas/{slug}", httpx.Adapter(m.handler.GetIdea),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermIdeaView),
	)
	r.Handle(http.MethodPut, "/api/v1/projects/{projectId}/ideas/{ideaId}", httpx.Adapter(m.handler.UpdateIdea),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermIdeaEdit),
	)
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/ideas/{ideaId}/select", httpx.Adapter(m.handler.SelectIdea),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermIdeaStatusChange),
	)
	r.Handle(http.MethodPut, "/api/v1/projects/{projectId}/ideas/{ideaId}/status", httpx.Adapter(m.handler.UpdateIdeaStatus),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermIdeaStatusChange),
	)

	// Tasks
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/tasks", httpx.Adapter(m.handler.ListTasks),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermTaskView),
	)
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/tasks", httpx.Adapter(m.handler.CreateTask),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermTaskCreate),
	)
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/tasks/{slug}", httpx.Adapter(m.handler.GetTask),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermTaskView),
	)
	r.Handle(http.MethodPut, "/api/v1/projects/{projectId}/tasks/{taskId}", httpx.Adapter(m.handler.UpdateTask),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermTaskEdit),
	)
	r.Handle(http.MethodPut, "/api/v1/projects/{projectId}/tasks/{taskId}/status", httpx.Adapter(m.handler.UpdateTaskStatus),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermTaskStatusChange),
	)

	// Feedback
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/feedback", httpx.Adapter(m.handler.ListFeedback),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermFeedbackView),
	)
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/feedback", httpx.Adapter(m.handler.CreateFeedback),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermFeedbackCreate),
	)
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/feedback/{slug}", httpx.Adapter(m.handler.GetFeedback),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermFeedbackView),
	)
	r.Handle(http.MethodPut, "/api/v1/projects/{projectId}/feedback/{feedbackId}", httpx.Adapter(m.handler.UpdateFeedback),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermFeedbackEdit),
	)

	return nil
}
