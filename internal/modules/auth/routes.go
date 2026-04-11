package auth

import (
	"net/http"

	"github.com/MrEthical07/superapi/internal/core/httpx"
	"github.com/MrEthical07/superapi/internal/core/policy"
)

// Register mounts ProjectBook auth routes.
func (m *Module) Register(r httpx.Router) error {
	if m.handler == nil {
		repo := NewRepo(m.runtime.RelationalStore())
		m.handler = NewHandler(NewService(m.runtime.AuthEngine(), repo))
	}

	r.Handle(http.MethodPost, "/api/v1/auth/signup", httpx.Adapter(m.handler.Signup))
	r.Handle(http.MethodPost, "/api/v1/auth/login", httpx.Adapter(m.handler.Login))
	r.Handle(http.MethodPost, "/api/v1/auth/verify-email", httpx.Adapter(m.handler.VerifyEmail))
	r.Handle(http.MethodPost, "/api/v1/auth/resend-verification", httpx.Adapter(m.handler.ResendVerification))
	r.Handle(http.MethodPost, "/api/v1/auth/forgot-password", httpx.Adapter(m.handler.ForgotPassword))
	r.Handle(http.MethodPost, "/api/v1/auth/reset-password", httpx.Adapter(m.handler.ResetPassword))

	// Compatibility endpoint kept for performance/load tooling during migration.
	r.Handle(http.MethodPost, "/api/v1/auth/refresh", httpx.Adapter(m.handler.Refresh))

	r.Handle(
		http.MethodPost,
		"/api/v1/auth/logout",
		httpx.Adapter(m.handler.Logout),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
	)

	return nil
}
