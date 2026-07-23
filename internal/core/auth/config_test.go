package auth

import (
	"context"
	"testing"
	"time"

	goauth "github.com/MrEthical07/goAuth"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

type noopUserProvider struct{}

func (noopUserProvider) GetUserByIdentifier(string) (goauth.UserRecord, error) {
	return goauth.UserRecord{}, goauth.ErrUserNotFound
}

func (noopUserProvider) GetUserByID(string) (goauth.UserRecord, error) {
	return goauth.UserRecord{}, goauth.ErrUserNotFound
}

func (noopUserProvider) UpdatePasswordHash(string, string) error {
	return goauth.ErrUserNotFound
}

func (noopUserProvider) CreateUser(context.Context, goauth.CreateUserInput) (goauth.UserRecord, error) {
	return goauth.UserRecord{}, goauth.ErrUnauthorized
}

func (noopUserProvider) UpdateAccountStatus(context.Context, string, goauth.AccountStatus) (goauth.UserRecord, error) {
	return goauth.UserRecord{}, goauth.ErrUnauthorized
}

func (noopUserProvider) GetTOTPSecret(context.Context, string) (*goauth.TOTPRecord, error) {
	return nil, goauth.ErrUnauthorized
}

func (noopUserProvider) EnableTOTP(context.Context, string, []byte) error {
	return goauth.ErrUnauthorized
}

func (noopUserProvider) DisableTOTP(context.Context, string) error {
	return goauth.ErrUnauthorized
}

func (noopUserProvider) MarkTOTPVerified(context.Context, string) error {
	return goauth.ErrUnauthorized
}

func (noopUserProvider) UpdateTOTPLastUsedCounter(context.Context, string, int64) error {
	return goauth.ErrUnauthorized
}

func (noopUserProvider) GetBackupCodes(context.Context, string) ([]goauth.BackupCodeRecord, error) {
	return nil, goauth.ErrUnauthorized
}

func (noopUserProvider) ReplaceBackupCodes(context.Context, string, []goauth.BackupCodeRecord) error {
	return goauth.ErrUnauthorized
}

func (noopUserProvider) ConsumeBackupCode(context.Context, string, [32]byte) (bool, error) {
	return false, goauth.ErrUnauthorized
}

func TestProjectBookGoAuthConfigControlledOverrides(t *testing.T) {
	cfg := projectBookGoAuthConfig(ModeHybrid)

	if cfg.JWT.AccessTTL != 5*time.Minute {
		t.Fatalf("AccessTTL=%s want=%s", cfg.JWT.AccessTTL, 5*time.Minute)
	}
	if cfg.JWT.RefreshTTL != 7*24*time.Hour {
		t.Fatalf("RefreshTTL=%s want=%s", cfg.JWT.RefreshTTL, 7*24*time.Hour)
	}
	if cfg.JWT.Issuer != "projectbook" {
		t.Fatalf("Issuer=%q want=%q", cfg.JWT.Issuer, "projectbook")
	}
	if cfg.JWT.Audience != "projectbook-api" {
		t.Fatalf("Audience=%q want=%q", cfg.JWT.Audience, "projectbook-api")
	}
	if cfg.JWT.KeyID != "v1" {
		t.Fatalf("KeyID=%q want=%q", cfg.JWT.KeyID, "v1")
	}

	if cfg.Security.EnablePermissionVersionCheck {
		t.Fatalf("EnablePermissionVersionCheck=true want=false")
	}
	if cfg.Security.EnableRoleVersionCheck {
		t.Fatalf("EnableRoleVersionCheck=true want=false")
	}
	if cfg.Security.EnableAccountVersionCheck {
		t.Fatalf("EnableAccountVersionCheck=true want=false")
	}
	if !cfg.Security.EnforceRefreshRotation {
		t.Fatalf("EnforceRefreshRotation=false want=true")
	}
	if !cfg.Security.EnforceRefreshReuseDetection {
		t.Fatalf("EnforceRefreshReuseDetection=false want=true")
	}
	if !cfg.Security.EnableLoginFailureLimiter {
		t.Fatalf("EnableLoginFailureLimiter=false want=true")
	}
	if cfg.Security.EnableIPBinding {
		t.Fatalf("EnableIPBinding=true want=false")
	}
	if cfg.Security.EnableIPSignal {
		t.Fatalf("EnableIPSignal=true want=false")
	}
	if cfg.Security.ProductionMode {
		t.Fatalf("ProductionMode=true want=false")
	}

	if !cfg.Session.SlidingExpiration {
		t.Fatalf("SlidingExpiration=false want=true")
	}
	if cfg.Session.AbsoluteSessionLifetime != 7*24*time.Hour {
		t.Fatalf("AbsoluteSessionLifetime=%s want=%s", cfg.Session.AbsoluteSessionLifetime, 7*24*time.Hour)
	}
	if cfg.DeviceBinding.Enabled {
		t.Fatalf("DeviceBinding.Enabled=true want=false")
	}
	if cfg.MultiTenant.Enabled {
		t.Fatalf("MultiTenant.Enabled=true want=false")
	}
	if cfg.Permission.MaxBits != 64 {
		t.Fatalf("Permission.MaxBits=%d want=64", cfg.Permission.MaxBits)
	}
	if cfg.Permission.RootBitReserved {
		t.Fatalf("Permission.RootBitReserved=true want=false")
	}
	if !cfg.PasswordReset.Enabled {
		t.Fatalf("PasswordReset.Enabled=false want=true")
	}
	if cfg.PasswordReset.Strategy != goauth.ResetOTP {
		t.Fatalf("PasswordReset.Strategy=%v want=%v", cfg.PasswordReset.Strategy, goauth.ResetOTP)
	}
	if cfg.PasswordReset.ResetTTL != 15*time.Minute {
		t.Fatalf("PasswordReset.ResetTTL=%s want=%s", cfg.PasswordReset.ResetTTL, 15*time.Minute)
	}
	if cfg.PasswordReset.MaxAttempts != 5 {
		t.Fatalf("PasswordReset.MaxAttempts=%d want=5", cfg.PasswordReset.MaxAttempts)
	}
	if cfg.PasswordReset.OTPDigits != 6 {
		t.Fatalf("PasswordReset.OTPDigits=%d want=6", cfg.PasswordReset.OTPDigits)
	}
	if !cfg.PasswordReset.EnableRequestLimiter {
		t.Fatalf("PasswordReset.EnableRequestLimiter=false want=true")
	}
	if !cfg.PasswordReset.EnableConfirmFailureLimiter {
		t.Fatalf("PasswordReset.EnableConfirmFailureLimiter=false want=true")
	}
	if !cfg.EmailVerification.Enabled {
		t.Fatalf("EmailVerification.Enabled=false want=true")
	}
	if cfg.EmailVerification.RequireForLogin {
		t.Fatalf("EmailVerification.RequireForLogin=true want=false")
	}
	if cfg.EmailVerification.Strategy != goauth.VerificationOTP {
		t.Fatalf("EmailVerification.Strategy=%v want=%v", cfg.EmailVerification.Strategy, goauth.VerificationOTP)
	}
	if cfg.EmailVerification.VerificationTTL != 15*time.Minute {
		t.Fatalf("EmailVerification.VerificationTTL=%s want=%s", cfg.EmailVerification.VerificationTTL, 15*time.Minute)
	}
	if cfg.EmailVerification.MaxAttempts != 5 {
		t.Fatalf("EmailVerification.MaxAttempts=%d want=5", cfg.EmailVerification.MaxAttempts)
	}
	if cfg.EmailVerification.OTPDigits != 6 {
		t.Fatalf("EmailVerification.OTPDigits=%d want=6", cfg.EmailVerification.OTPDigits)
	}
	if !cfg.Account.Enabled {
		t.Fatalf("Account.Enabled=false want=true")
	}
	if cfg.Account.DefaultRole != "user" {
		t.Fatalf("Account.DefaultRole=%q want=%q", cfg.Account.DefaultRole, "user")
	}
}

func TestNewGoAuthEngineBuildsWithControlledConfig(t *testing.T) {
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
		mr.Close()
	})

	engine, shutdown, err := NewGoAuthEngine(redisClient, ModeHybrid, noopUserProvider{})
	if err != nil {
		t.Fatalf("NewGoAuthEngine() error = %v", err)
	}
	if engine == nil {
		t.Fatalf("engine=nil want non-nil")
	}
	if shutdown == nil {
		t.Fatalf("shutdown=nil want non-nil")
	}

	shutdown()
}
