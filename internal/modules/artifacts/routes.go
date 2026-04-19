package artifacts

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
		return fmt.Errorf("artifacts module requires permission resolver")
	}

	cacheMgr := m.runtime.CacheManager()
	limiter := m.runtime.Limiter()
	invalidateStoryTags := cache.CacheInvalidateConfig{TagSpecs: []cache.CacheTagSpec{
		{Name: "artifacts.project", PathParams: []string{"projectId"}},
		{Name: "artifacts.story", PathParams: []string{"projectId"}},
	}}
	invalidateJourneyTags := cache.CacheInvalidateConfig{TagSpecs: []cache.CacheTagSpec{
		{Name: "artifacts.project", PathParams: []string{"projectId"}},
		{Name: "artifacts.journey", PathParams: []string{"projectId"}},
	}}
	invalidateProblemTags := cache.CacheInvalidateConfig{TagSpecs: []cache.CacheTagSpec{
		{Name: "artifacts.project", PathParams: []string{"projectId"}},
		{Name: "artifacts.problem", PathParams: []string{"projectId"}},
	}}
	invalidateIdeaTags := cache.CacheInvalidateConfig{TagSpecs: []cache.CacheTagSpec{
		{Name: "artifacts.project", PathParams: []string{"projectId"}},
		{Name: "artifacts.idea", PathParams: []string{"projectId"}},
	}}
	invalidateTaskTags := cache.CacheInvalidateConfig{TagSpecs: []cache.CacheTagSpec{
		{Name: "artifacts.project", PathParams: []string{"projectId"}},
		{Name: "artifacts.task", PathParams: []string{"projectId"}},
	}}
	invalidateFeedbackTags := cache.CacheInvalidateConfig{TagSpecs: []cache.CacheTagSpec{
		{Name: "artifacts.project", PathParams: []string{"projectId"}},
		{Name: "artifacts.feedback", PathParams: []string{"projectId"}},
	}}
	writeRule := ratelimit.Rule{Limit: 40, Window: time.Minute, Scope: ratelimit.ScopeProject}
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
	listTTL := 30 * time.Second
	detailTTL := 90 * time.Second
	payloadMaxBytes := 100 * 1024

	// Stories
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/stories", httpx.Adapter(m.handler.ListStories),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermStoryView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                listTTL,
			MaxBytes:           payloadMaxBytes,
			TagSpecs:           storyReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:   true,
				PathParams:  []string{"projectId"},
				QueryParams: []string{"filter.status", "status", "pagination.cursor", "cursor", "pagination.limit", "limit"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, listTTL),
	)
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/stories", httpx.Adapter(m.handler.CreateStory),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermStoryCreate),
		policy.RateLimitOptional(limiter, writeRule),
		policy.CacheInvalidateOptional(cacheMgr, invalidateStoryTags),
	)
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/stories/{storyId}", httpx.Adapter(m.handler.GetStory),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermStoryView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                detailTTL,
			MaxBytes:           payloadMaxBytes,
			TagSpecs:           storyReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:  true,
				PathParams: []string{"projectId", "storyId"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, detailTTL),
	)
	r.Handle(http.MethodPatch, "/api/v1/projects/{projectId}/stories/{storyId}", httpx.Adapter(m.handler.UpdateStory),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermStoryEdit),
		policy.RateLimitOptional(limiter, writeRule),
		policy.CacheInvalidateOptional(cacheMgr, invalidateStoryTags),
	)

	// Journeys share story permission scopes.
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/journeys", httpx.Adapter(m.handler.ListJourneys),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermStoryView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                listTTL,
			MaxBytes:           payloadMaxBytes,
			TagSpecs:           journeyReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:   true,
				PathParams:  []string{"projectId"},
				QueryParams: []string{"filter.status", "status", "pagination.cursor", "cursor", "pagination.limit", "limit"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, listTTL),
	)
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/journeys", httpx.Adapter(m.handler.CreateJourney),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermStoryCreate),
		policy.RateLimitOptional(limiter, writeRule),
		policy.CacheInvalidateOptional(cacheMgr, invalidateJourneyTags),
	)
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/journeys/{journeyId}", httpx.Adapter(m.handler.GetJourney),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermStoryView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                detailTTL,
			MaxBytes:           payloadMaxBytes,
			TagSpecs:           journeyReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:  true,
				PathParams: []string{"projectId", "journeyId"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, detailTTL),
	)
	r.Handle(http.MethodPatch, "/api/v1/projects/{projectId}/journeys/{journeyId}", httpx.Adapter(m.handler.UpdateJourney),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermStoryEdit),
		policy.RateLimitOptional(limiter, writeRule),
		policy.CacheInvalidateOptional(cacheMgr, invalidateJourneyTags),
	)

	// Problems
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/problems", httpx.Adapter(m.handler.ListProblems),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermProblemView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                listTTL,
			MaxBytes:           payloadMaxBytes,
			TagSpecs:           problemReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:   true,
				PathParams:  []string{"projectId"},
				QueryParams: []string{"filter.status", "status", "pagination.cursor", "cursor", "pagination.limit", "limit"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, listTTL),
	)
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/problems", httpx.Adapter(m.handler.CreateProblem),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermProblemCreate),
		policy.RateLimitOptional(limiter, writeRule),
		policy.CacheInvalidateOptional(cacheMgr, invalidateProblemTags),
	)
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/problems/{problemId}", httpx.Adapter(m.handler.GetProblem),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermProblemView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                detailTTL,
			MaxBytes:           payloadMaxBytes,
			TagSpecs:           problemReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:  true,
				PathParams: []string{"projectId", "problemId"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, detailTTL),
	)
	r.Handle(http.MethodPatch, "/api/v1/projects/{projectId}/problems/{problemId}", httpx.Adapter(m.handler.UpdateProblem),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermProblemEdit),
		policy.RateLimitOptional(limiter, writeRule),
		policy.CacheInvalidateOptional(cacheMgr, invalidateProblemTags),
	)

	// Ideas
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/ideas", httpx.Adapter(m.handler.ListIdeas),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermIdeaView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                listTTL,
			MaxBytes:           payloadMaxBytes,
			TagSpecs:           ideaReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:   true,
				PathParams:  []string{"projectId"},
				QueryParams: []string{"filter.status", "status", "pagination.cursor", "cursor", "pagination.limit", "limit"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, listTTL),
	)
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/ideas", httpx.Adapter(m.handler.CreateIdea),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermIdeaCreate),
		policy.RateLimitOptional(limiter, writeRule),
		policy.CacheInvalidateOptional(cacheMgr, invalidateIdeaTags),
	)
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/ideas/{ideaId}", httpx.Adapter(m.handler.GetIdea),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermIdeaView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                detailTTL,
			MaxBytes:           payloadMaxBytes,
			TagSpecs:           ideaReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:  true,
				PathParams: []string{"projectId", "ideaId"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, detailTTL),
	)
	r.Handle(http.MethodPatch, "/api/v1/projects/{projectId}/ideas/{ideaId}", httpx.Adapter(m.handler.UpdateIdea),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermIdeaEdit),
		policy.RateLimitOptional(limiter, writeRule),
		policy.CacheInvalidateOptional(cacheMgr, invalidateIdeaTags),
	)

	// Tasks
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/tasks", httpx.Adapter(m.handler.ListTasks),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermTaskView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                listTTL,
			MaxBytes:           payloadMaxBytes,
			TagSpecs:           taskReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:   true,
				PathParams:  []string{"projectId"},
				QueryParams: []string{"filter.status", "status", "pagination.cursor", "cursor", "pagination.limit", "limit"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, listTTL),
	)
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/tasks", httpx.Adapter(m.handler.CreateTask),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermTaskCreate),
		policy.RateLimitOptional(limiter, writeRule),
		policy.CacheInvalidateOptional(cacheMgr, invalidateTaskTags),
	)
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/tasks/{taskId}", httpx.Adapter(m.handler.GetTask),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermTaskView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                detailTTL,
			MaxBytes:           payloadMaxBytes,
			TagSpecs:           taskReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:  true,
				PathParams: []string{"projectId", "taskId"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, detailTTL),
	)
	r.Handle(http.MethodPatch, "/api/v1/projects/{projectId}/tasks/{taskId}", httpx.Adapter(m.handler.UpdateTask),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermTaskEdit),
		policy.RateLimitOptional(limiter, writeRule),
		policy.CacheInvalidateOptional(cacheMgr, invalidateTaskTags),
	)

	// Feedback
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/feedback", httpx.Adapter(m.handler.ListFeedback),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermFeedbackView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                listTTL,
			MaxBytes:           payloadMaxBytes,
			TagSpecs:           feedbackReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:   true,
				PathParams:  []string{"projectId"},
				QueryParams: []string{"filter.status", "status", "filter.outcome", "outcome", "pagination.cursor", "cursor", "pagination.limit", "limit"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, listTTL),
	)
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/feedback", httpx.Adapter(m.handler.CreateFeedback),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermFeedbackCreate),
		policy.RateLimitOptional(limiter, writeRule),
		policy.CacheInvalidateOptional(cacheMgr, invalidateFeedbackTags),
	)
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/feedback/{feedbackId}", httpx.Adapter(m.handler.GetFeedback),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermFeedbackView),
		policy.CacheReadOptional(cacheMgr, cache.CacheReadConfig{
			TTL:                detailTTL,
			MaxBytes:           payloadMaxBytes,
			TagSpecs:           feedbackReadTags,
			AllowAuthenticated: true,
			VaryBy: cache.CacheVaryBy{
				ProjectID:  true,
				PathParams: []string{"projectId", "feedbackId"},
			},
		}),
		policy.CacheControlOptional(cacheMgr, detailTTL),
	)
	r.Handle(http.MethodPatch, "/api/v1/projects/{projectId}/feedback/{feedbackId}", httpx.Adapter(m.handler.UpdateFeedback),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermFeedbackEdit),
		policy.RateLimitOptional(limiter, writeRule),
		policy.CacheInvalidateOptional(cacheMgr, invalidateFeedbackTags),
	)

	return nil
}
