package auth

import (
	"fmt"
	"os"
	"strings"
	"time"

	goauth "github.com/MrEthical07/goAuth"
	"github.com/redis/go-redis/v9"
)

type providerCloser interface {
	Close()
}

const goAuthAuthOnlyPermission = "projectbook.auth_only"

var goAuthAuthOnlyRoles = map[string][]string{
	"user":  {goAuthAuthOnlyPermission},
	"admin": {goAuthAuthOnlyPermission},
}

// NewGoAuthEngine builds a goAuth engine backed by Redis and SQLC user provider.
//
// Usage:
//
//	engine, shutdown, err := auth.NewGoAuthEngine(redisClient, mode, userProvider)
//
// Notes:
// - redisClient must be non-nil
// - shutdown should be called during application shutdown
// - AUTH_TEST_SHARED_SECRET enables deterministic local signer behavior
func NewGoAuthEngine(redisClient redis.UniversalClient, mode Mode, userProvider goauth.UserProvider) (*goauth.Engine, func(), error) {
	if redisClient == nil {
		return nil, nil, fmt.Errorf("goAuth provider requires redis client")
	}

	cfg := projectBookGoAuthConfig(mode)

	// Optional deterministic signer for local perf tests across multiple processes.
	if sharedSecret := strings.TrimSpace(os.Getenv("AUTH_TEST_SHARED_SECRET")); sharedSecret != "" {
		cfg.JWT.SigningMethod = "hs256"
		cfg.JWT.PrivateKey = []byte(sharedSecret)
		cfg.JWT.PublicKey = []byte(sharedSecret)
	}

	engine, err := goauth.New().
		WithConfig(cfg).
		WithRedis(redisClient).
		WithPermissions([]string{goAuthAuthOnlyPermission}).
		WithRoles(goAuthAuthOnlyRoles).
		WithUserProvider(userProvider).
		Build()
	if err != nil {
		return nil, nil, fmt.Errorf("build goAuth engine: %w", err)
	}

	shutdown := func() {
		if closer, ok := any(engine).(providerCloser); ok {
			closer.Close()
		}
	}

	return engine, shutdown, nil
}

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
