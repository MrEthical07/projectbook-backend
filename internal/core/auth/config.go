package auth

import (
	"time"

	goauth "github.com/MrEthical07/goAuth"
)

// projectBookGoAuthConfig builds the goAuth engine configuration for ProjectBook.
func projectBookGoAuthConfig(mode Mode) goauth.Config {
	cfg := goauth.DefaultConfig()
	cfg.ValidationMode = toGoAuthValidationMode(mode)

	// JWT identity for ProjectBook API.
	cfg.JWT.AccessTTL = 5 * time.Minute
	cfg.JWT.RefreshTTL = 7 * 24 * time.Hour
	cfg.JWT.Issuer = "projectbook"
	cfg.JWT.Audience = "projectbook-api"
	cfg.JWT.KeyID = "v1"

	// goAuth handles authentication only; RBAC authorization is external.
	cfg.Result.IncludeRole = true
	cfg.Result.IncludePermissions = false
	cfg.Security.EnablePermissionVersionCheck = false
	cfg.Security.EnableRoleVersionCheck = false
	// Keep refresh tokens valid across email verification transitions so the
	// web flow can verify and continue without forcing re-login.
	cfg.Security.EnableAccountVersionCheck = false
	cfg.Security.EnforceRefreshRotation = true
	cfg.Security.EnforceRefreshReuseDetection = true
	cfg.Security.EnableLoginFailureLimiter = true
	cfg.Security.EnableIPBinding = false
	cfg.Security.EnableIPSignal = false
	cfg.Security.ProductionMode = false

	// Session behavior remains default but explicit for deterministic setup.
	cfg.Session.SlidingExpiration = true
	cfg.Session.AbsoluteSessionLifetime = 7 * 24 * time.Hour

	// Disable unsupported features for controlled pre-production setup.
	cfg.DeviceBinding.Enabled = false
	cfg.MultiTenant.Enabled = false

	// Match ProjectBook permission-mask model constraints.
	cfg.Permission.MaxBits = 64
	cfg.Permission.RootBitReserved = false

	// Enable auth-contract flows exposed through module endpoints.
	cfg.PasswordReset.Enabled = true
	cfg.PasswordReset.Strategy = goauth.ResetOTP
	cfg.PasswordReset.ResetTTL = 15 * time.Minute
	cfg.PasswordReset.MaxAttempts = 5
	cfg.PasswordReset.OTPDigits = 6
	cfg.PasswordReset.EnableRequestLimiter = true
	cfg.PasswordReset.EnableConfirmFailureLimiter = true
	cfg.EmailVerification.Enabled = true
	cfg.EmailVerification.Strategy = goauth.VerificationOTP
	// Login is allowed before verification; route access is gated by the web layer
	// until email_verified is true.
	cfg.EmailVerification.RequireForLogin = false
	cfg.EmailVerification.VerificationTTL = 15 * time.Minute
	cfg.EmailVerification.MaxAttempts = 5
	cfg.EmailVerification.OTPDigits = 6
	cfg.EmailVerification.EnableRequestLimiter = true
	cfg.EmailVerification.EnableConfirmFailureLimiter = true
	cfg.Account.Enabled = true
	cfg.Account.DefaultRole = "user"

	return cfg
}

func toGoAuthValidationMode(mode Mode) goauth.ValidationMode {
	switch mode {
	case ModeJWTOnly:
		return goauth.ModeJWTOnly
	case ModeStrict:
		return goauth.ModeStrict
	case ModeHybrid:
		return goauth.ModeHybrid
	default:
		return goauth.ModeHybrid
	}
}
