package auth

import (
	"net/http"
	"time"

	"github.com/MrEthical07/superapi/internal/core/httpx"
	"github.com/MrEthical07/superapi/internal/core/policy"
	"github.com/MrEthical07/superapi/internal/core/ratelimit"
)

// Register mounts ProjectBook auth routes.
func (m *Module) Register(r httpx.Router) error {
	if m.handler == nil {
		repo := NewRepo(m.runtime.RelationalStore())
		m.handler = NewHandler(NewService(m.runtime.AuthEngine(), repo))
	}

	limiter := m.runtime.Limiter()
	signupRule := ratelimit.Rule{Limit: 20, Window: time.Minute, Scope: ratelimit.ScopeIP}
	loginRule := ratelimit.Rule{Limit: 30, Window: time.Minute, Scope: ratelimit.ScopeIP}
	verifyRule := ratelimit.Rule{Limit: 20, Window: time.Minute, Scope: ratelimit.ScopeIP}
	passwordRule := ratelimit.Rule{Limit: 15, Window: time.Minute, Scope: ratelimit.ScopeIP}
	refreshRule := ratelimit.Rule{Limit: 45, Window: time.Minute, Scope: ratelimit.ScopeIP}
	logoutRule := ratelimit.Rule{Limit: 60, Window: time.Minute, Scope: ratelimit.ScopeUser}

	if limiter != nil {
		r.Handle(http.MethodPost, "/api/v1/auth/signup", httpx.Adapter(m.handler.Signup),
			policy.RequireJSON(),
			policy.RateLimitWithKeyer(limiter, "auth.signup", signupRule, ratelimit.KeyByIP()),
		)
		r.Handle(http.MethodPost, "/api/v1/auth/login", httpx.Adapter(m.handler.Login),
			policy.RequireJSON(),
			policy.RateLimitWithKeyer(limiter, "auth.login", loginRule, ratelimit.KeyByIP()),
		)
		r.Handle(http.MethodPost, "/api/v1/auth/verify-email", httpx.Adapter(m.handler.VerifyEmail),
			policy.RequireJSON(),
			policy.RateLimitWithKeyer(limiter, "auth.verify_email", verifyRule, ratelimit.KeyByIP()),
		)
		r.Handle(http.MethodPost, "/api/v1/auth/resend-verification", httpx.Adapter(m.handler.ResendVerification),
			policy.RequireJSON(),
			policy.RateLimitWithKeyer(limiter, "auth.resend_verification", verifyRule, ratelimit.KeyByIP()),
		)
		r.Handle(http.MethodPost, "/api/v1/auth/forgot-password", httpx.Adapter(m.handler.ForgotPassword),
			policy.RequireJSON(),
			policy.RateLimitWithKeyer(limiter, "auth.forgot_password", passwordRule, ratelimit.KeyByIP()),
		)
		r.Handle(http.MethodPost, "/api/v1/auth/reset-password", httpx.Adapter(m.handler.ResetPassword),
			policy.RequireJSON(),
			policy.RateLimitWithKeyer(limiter, "auth.reset_password", passwordRule, ratelimit.KeyByIP()),
		)

		// Compatibility endpoint kept for performance/load tooling during migration.
		r.Handle(http.MethodPost, "/api/v1/auth/refresh", httpx.Adapter(m.handler.Refresh),
			policy.RequireJSON(),
			policy.RateLimitWithKeyer(limiter, "auth.refresh", refreshRule, ratelimit.KeyByIP()),
		)

		r.Handle(
			http.MethodPost,
			"/api/v1/auth/logout",
			httpx.Adapter(m.handler.Logout),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.RateLimitWithKeyer(limiter, "auth.logout", logoutRule, ratelimit.KeyByUserOrProjectOrTokenHash(16)),
		)
		return nil
	}

	r.Handle(http.MethodPost, "/api/v1/auth/signup", httpx.Adapter(m.handler.Signup), policy.RequireJSON())
	r.Handle(http.MethodPost, "/api/v1/auth/login", httpx.Adapter(m.handler.Login), policy.RequireJSON())
	r.Handle(http.MethodPost, "/api/v1/auth/verify-email", httpx.Adapter(m.handler.VerifyEmail), policy.RequireJSON())
	r.Handle(http.MethodPost, "/api/v1/auth/resend-verification", httpx.Adapter(m.handler.ResendVerification), policy.RequireJSON())
	r.Handle(http.MethodPost, "/api/v1/auth/forgot-password", httpx.Adapter(m.handler.ForgotPassword), policy.RequireJSON())
	r.Handle(http.MethodPost, "/api/v1/auth/reset-password", httpx.Adapter(m.handler.ResetPassword), policy.RequireJSON())

	// Compatibility endpoint kept for performance/load tooling during migration.
	r.Handle(http.MethodPost, "/api/v1/auth/refresh", httpx.Adapter(m.handler.Refresh), policy.RequireJSON())

	r.Handle(
		http.MethodPost,
		"/api/v1/auth/logout",
		httpx.Adapter(m.handler.Logout),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
	)

	return nil
}
