package feedback

import (
	"net/http"
	"time"

	"github.com/MrEthical07/superapi/internal/core/httpx"
	"github.com/MrEthical07/superapi/internal/core/policy"
	"github.com/MrEthical07/superapi/internal/core/ratelimit"
)

// Register mounts global feedback submission routes.
func (m *Module) Register(r httpx.Router) error {
	if m.handler == nil {
		repo := NewRepo(m.runtime.RelationalStore())
		m.handler = NewHandler(NewService(m.runtime.RelationalStore(), repo, m.runtime.EmailSender()))
	}

	feedbackRule := ratelimit.Rule{Limit: 5, Window: time.Minute, Scope: ratelimit.ScopeUser}

	if limiter := m.runtime.Limiter(); limiter != nil {
		r.Handle(
			http.MethodPost,
			"/api/v1/feedback",
			httpx.Adapter(m.handler.Submit),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.RequireJSON(),
			policy.RateLimit(limiter, feedbackRule),
		)
		return nil
	}

	r.Handle(
		http.MethodPost,
		"/api/v1/feedback",
		httpx.Adapter(m.handler.Submit),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
		policy.RequireJSON(),
	)
	return nil
}
