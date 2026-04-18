package artifacts

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
		return fmt.Errorf("artifacts module requires permission resolver")
	}

	cacheMgr := m.runtime.CacheManager()
	invalidateArtifactTags := cache.CacheInvalidateConfig{TagSpecs: []cache.CacheTagSpec{
		{Name: "artifacts.project", PathParams: []string{"projectId"}},
		{Name: "artifacts.story", PathParams: []string{"projectId"}},
		{Name: "artifacts.journey", PathParams: []string{"projectId"}},
		{Name: "artifacts.problem", PathParams: []string{"projectId"}},
		{Name: "artifacts.idea", PathParams: []string{"projectId"}},
		{Name: "artifacts.task", PathParams: []string{"projectId"}},
		{Name: "artifacts.feedback", PathParams: []string{"projectId"}},
		{Name: "project.navigation", PathParams: []string{"projectId"}},
	}}
	storyReadTags := []cache.CacheTagSpec{
		{Name: "artifacts.project", PathParams: []string{"projectId"}},
		{Name: "artifacts.story", PathParams: []string{"projectId"}},
	}
	journeyReadTags := []cache.CacheTagSpec{
		{Name: "artifacts.project", PathParams: []string{"projectId"}},
		{Name: "artifacts.journey", PathParams: []string{"projectId"}},
	}
	problemReadTags := []cache.CacheTagSpec{
		{Name: "artifacts.project", PathParams: []string{"projectId"}},
		{Name: "artifacts.problem", PathParams: []string{"projectId"}},
	}
	ideaReadTags := []cache.CacheTagSpec{
		{Name: "artifacts.project", PathParams: []string{"projectId"}},
		{Name: "artifacts.idea", PathParams: []string{"projectId"}},
	}
	taskReadTags := []cache.CacheTagSpec{
		{Name: "artifacts.project", PathParams: []string{"projectId"}},
		{Name: "artifacts.task", PathParams: []string{"projectId"}},
	}
	feedbackReadTags := []cache.CacheTagSpec{
		{Name: "artifacts.project", PathParams: []string{"projectId"}},
		{Name: "artifacts.feedback", PathParams: []string{"projectId"}},
	}
	ttl := 30 * time.Second

	// Stories
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/stories", httpx.Adapter(m.handler.ListStories),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermStoryView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                ttl,
			TagSpecs:           storyReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:   true,
				PathParams:  []string{"projectId"},
				QueryParams: []string{"status", "cursor", "limit"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, ttl),
	)
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/stories", httpx.Adapter(m.handler.CreateStory),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermStoryCreate),
		policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags),
	)
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/stories/{storyId}", httpx.Adapter(m.handler.GetStory),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermStoryView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                ttl,
			TagSpecs:           storyReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:  true,
				PathParams: []string{"projectId", "storyId"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, ttl),
	)
	r.Handle(http.MethodPatch, "/api/v1/projects/{projectId}/stories/{storyId}", httpx.Adapter(m.handler.UpdateStory),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermStoryEdit),
		policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags),
	)

	// Journeys share story permission scopes.
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/journeys", httpx.Adapter(m.handler.ListJourneys),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermStoryView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                ttl,
			TagSpecs:           journeyReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:   true,
				PathParams:  []string{"projectId"},
				QueryParams: []string{"status", "cursor", "limit"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, ttl),
	)
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/journeys", httpx.Adapter(m.handler.CreateJourney),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermStoryCreate),
		policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags),
	)
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/journeys/{journeyId}", httpx.Adapter(m.handler.GetJourney),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermStoryView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                ttl,
			TagSpecs:           journeyReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:  true,
				PathParams: []string{"projectId", "journeyId"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, ttl),
	)
	r.Handle(http.MethodPatch, "/api/v1/projects/{projectId}/journeys/{journeyId}", httpx.Adapter(m.handler.UpdateJourney),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermStoryEdit),
		policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags),
	)

	// Problems
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/problems", httpx.Adapter(m.handler.ListProblems),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermProblemView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                ttl,
			TagSpecs:           problemReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:   true,
				PathParams:  []string{"projectId"},
				QueryParams: []string{"status", "cursor", "limit"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, ttl),
	)
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/problems", httpx.Adapter(m.handler.CreateProblem),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermProblemCreate),
		policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags),
	)
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/problems/{problemId}", httpx.Adapter(m.handler.GetProblem),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermProblemView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                ttl,
			TagSpecs:           problemReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:  true,
				PathParams: []string{"projectId", "problemId"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, ttl),
	)
	r.Handle(http.MethodPatch, "/api/v1/projects/{projectId}/problems/{problemId}", httpx.Adapter(m.handler.UpdateProblem),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermProblemEdit),
		policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags),
	)
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/problems/{problemId}/lock", httpx.Adapter(m.handler.LockProblem),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermProblemStatusChange),
		policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags),
	)
	r.Handle(http.MethodPatch, "/api/v1/projects/{projectId}/problems/{problemId}/status", httpx.Adapter(m.handler.UpdateProblemStatus),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermProblemStatusChange),
		policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags),
	)

	// Ideas
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/ideas", httpx.Adapter(m.handler.ListIdeas),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermIdeaView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                ttl,
			TagSpecs:           ideaReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:   true,
				PathParams:  []string{"projectId"},
				QueryParams: []string{"status", "cursor", "limit"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, ttl),
	)
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/ideas", httpx.Adapter(m.handler.CreateIdea),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermIdeaCreate),
		policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags),
	)
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/ideas/{ideaId}", httpx.Adapter(m.handler.GetIdea),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermIdeaView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                ttl,
			TagSpecs:           ideaReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:  true,
				PathParams: []string{"projectId", "ideaId"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, ttl),
	)
	r.Handle(http.MethodPatch, "/api/v1/projects/{projectId}/ideas/{ideaId}", httpx.Adapter(m.handler.UpdateIdea),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermIdeaEdit),
		policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags),
	)
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/ideas/{ideaId}/select", httpx.Adapter(m.handler.SelectIdea),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermIdeaStatusChange),
		policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags),
	)
	r.Handle(http.MethodPatch, "/api/v1/projects/{projectId}/ideas/{ideaId}/status", httpx.Adapter(m.handler.UpdateIdeaStatus),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermIdeaStatusChange),
		policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags),
	)

	// Tasks
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/tasks", httpx.Adapter(m.handler.ListTasks),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermTaskView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                ttl,
			TagSpecs:           taskReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:   true,
				PathParams:  []string{"projectId"},
				QueryParams: []string{"status", "cursor", "limit"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, ttl),
	)
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/tasks", httpx.Adapter(m.handler.CreateTask),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermTaskCreate),
		policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags),
	)
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/tasks/{taskId}", httpx.Adapter(m.handler.GetTask),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermTaskView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                ttl,
			TagSpecs:           taskReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:  true,
				PathParams: []string{"projectId", "taskId"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, ttl),
	)
	r.Handle(http.MethodPatch, "/api/v1/projects/{projectId}/tasks/{taskId}", httpx.Adapter(m.handler.UpdateTask),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermTaskEdit),
		policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags),
	)
	r.Handle(http.MethodPatch, "/api/v1/projects/{projectId}/tasks/{taskId}/status", httpx.Adapter(m.handler.UpdateTaskStatus),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermTaskStatusChange),
		policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags),
	)

	// Feedback
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/feedback", httpx.Adapter(m.handler.ListFeedback),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermFeedbackView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                ttl,
			TagSpecs:           feedbackReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:   true,
				PathParams:  []string{"projectId"},
				QueryParams: []string{"status", "outcome", "cursor", "limit"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, ttl),
	)
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/feedback", httpx.Adapter(m.handler.CreateFeedback),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermFeedbackCreate),
		policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags),
	)
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/feedback/{feedbackId}", httpx.Adapter(m.handler.GetFeedback),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermFeedbackView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                ttl,
			TagSpecs:           feedbackReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:  true,
				PathParams: []string{"projectId", "feedbackId"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, ttl),
	)
	r.Handle(http.MethodPatch, "/api/v1/projects/{projectId}/feedback/{feedbackId}", httpx.Adapter(m.handler.UpdateFeedback),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermFeedbackEdit),
		policy.CacheInvalidateOptional(cacheMgr, invalidateArtifactTags),
	)

	return nil
}
