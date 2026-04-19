package calendar

import (
	"fmt"
	"net/http"
	"time"

	"github.com/MrEthical07/superapi/internal/core/httpx"
	"github.com/MrEthical07/superapi/internal/core/policy"
	"github.com/MrEthical07/superapi/internal/core/ratelimit"
	"github.com/MrEthical07/superapi/internal/core/rbac"
)

func (m *Module) Register(r httpx.Router) error {
	if m.handler == nil {
		repo := NewRepo(m.runtime.RelationalStore())
		m.handler = NewHandler(NewService(m.runtime.RelationalStore(), repo))
	}

	resolver := m.runtime.PermissionResolver()
	if resolver == nil {
		return fmt.Errorf("calendar module requires permission resolver")
	}
	limiter := m.runtime.Limiter()
	writeRule := ratelimit.Rule{Limit: 40, Window: time.Minute, Scope: ratelimit.ScopeProject}

	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/calendar", httpx.Adapter(m.handler.ListCalendarData),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermCalendarView),
	)
	r.Handle(http.MethodPost, "/api/v1/projects/{projectId}/calendar", httpx.Adapter(m.handler.CreateCalendarEvent),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermCalendarCreate),
		policy.RateLimitOptional(limiter, writeRule),
	)
	r.Handle(http.MethodGet, "/api/v1/projects/{projectId}/calendar/{eventId}", httpx.Adapter(m.handler.GetCalendarEvent),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermCalendarView),
	)
	r.Handle(http.MethodPatch, "/api/v1/projects/{projectId}/calendar/{eventId}", httpx.Adapter(m.handler.UpdateCalendarEvent),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequireJSON(),
		policy.RequirePermission(rbac.PermCalendarEdit),
		policy.RateLimitOptional(limiter, writeRule),
	)
	r.Handle(http.MethodDelete, "/api/v1/projects/{projectId}/calendar/{eventId}", httpx.Adapter(m.handler.DeleteCalendarEvent),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.ProjectRequired(),
		policy.ProjectMatchFromPath("projectId"),
		policy.ResolvePermissions(resolver),
		policy.RequirePermission(rbac.PermCalendarDelete),
		policy.RateLimitOptional(limiter, writeRule),
	)

	return nil
}
