package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	goauth "github.com/MrEthical07/goAuth"
	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/golang-jwt/jwt/v5"
)

// Service defines auth business workflows.
type Service interface {
	Signup(ctx context.Context, req signupRequest) (signupResponse, error)
	Login(ctx context.Context, req loginRequest) (authTokenResponse, error)
	Refresh(ctx context.Context, req refreshRequest) (authTokenResponse, error)
	Logout(ctx context.Context, accessToken string) error
	VerifyEmail(ctx context.Context, req verifyEmailRequest) (statusResponse, error)
	ResendVerification(ctx context.Context, req resendVerificationRequest) (statusResponse, error)
	ForgotPassword(ctx context.Context, req forgotPasswordRequest) (statusResponse, error)
	ResetPassword(ctx context.Context, req resetPasswordRequest) (statusResponse, error)
}

type service struct {
	engine *goauth.Engine
	repo   Repo
}

// NewService constructs auth business workflows.
func NewService(engine *goauth.Engine, repo Repo) Service {
	return &service{engine: engine, repo: repo}
}

func (s *service) Signup(ctx context.Context, req signupRequest) (signupResponse, error) {
	engine, err := s.authEngine()
	if err != nil {
		return signupResponse{}, err
	}

	created, err := engine.CreateAccount(ctx, goauth.CreateAccountRequest{
		Identifier: normalizeEmail(req.Email),
		Password:   req.Password,
	})
	if err != nil {
		return signupResponse{}, mapSignupError(err)
	}

	if s.repo != nil {
		if err := s.repo.UpdateUserName(ctx, created.UserID, strings.TrimSpace(req.Name)); err != nil {
			return signupResponse{}, err
		}
	}

	return signupResponse{
		User: signupUserResponse{
			ID:              created.UserID,
			Name:            strings.TrimSpace(req.Name),
			Email:           normalizeEmail(req.Email),
			IsEmailVerified: false,
		},
	}, nil
}

func (s *service) Login(ctx context.Context, req loginRequest) (authTokenResponse, error) {
	engine, err := s.authEngine()
	if err != nil {
		return authTokenResponse{}, err
	}

	accessToken, refreshToken, err := engine.Login(ctx, normalizeEmail(req.Email), req.Password)
	if err != nil {
		return authTokenResponse{}, mapAuthTokenError(err, "invalid credentials")
	}

	return buildAuthTokenResponse(accessToken, refreshToken)
}

func (s *service) Refresh(ctx context.Context, req refreshRequest) (authTokenResponse, error) {
	engine, err := s.authEngine()
	if err != nil {
		return authTokenResponse{}, err
	}

	accessToken, refreshToken, err := engine.Refresh(ctx, strings.TrimSpace(req.RefreshToken))
	if err != nil {
		return authTokenResponse{}, mapAuthTokenError(err, "invalid refresh token")
	}

	return buildAuthTokenResponse(accessToken, refreshToken)
}

func (s *service) Logout(ctx context.Context, accessToken string) error {
	engine, err := s.authEngine()
	if err != nil {
		return err
	}

	if strings.TrimSpace(accessToken) == "" {
		return apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	}

	if err := engine.LogoutByAccessToken(ctx, accessToken); err != nil {
		return mapLogoutError(err)
	}

	return nil
}

func (s *service) VerifyEmail(ctx context.Context, req verifyEmailRequest) (statusResponse, error) {
	engine, err := s.authEngine()
	if err != nil {
		return statusResponse{}, err
	}

	if err := engine.ConfirmEmailVerification(ctx, strings.TrimSpace(req.Token)); err != nil {
		return statusResponse{}, mapVerifyEmailError(err)
	}

	return statusResponse{Status: "success"}, nil
}

func (s *service) ResendVerification(ctx context.Context, req resendVerificationRequest) (statusResponse, error) {
	engine, err := s.authEngine()
	if err != nil {
		return statusResponse{}, err
	}

	if _, err := engine.RequestEmailVerification(ctx, normalizeEmail(req.Email)); err != nil {
		return statusResponse{}, mapResendVerificationError(err)
	}

	return statusResponse{Status: "sent"}, nil
}

func (s *service) ForgotPassword(ctx context.Context, req forgotPasswordRequest) (statusResponse, error) {
	engine, err := s.authEngine()
	if err != nil {
		return statusResponse{}, err
	}

	if _, err := engine.RequestPasswordReset(ctx, normalizeEmail(req.Email)); err != nil {
		return statusResponse{}, mapForgotPasswordError(err)
	}

	return statusResponse{Message: "If an account exists for this email, a reset link has been sent."}, nil
}

func (s *service) ResetPassword(ctx context.Context, req resetPasswordRequest) (statusResponse, error) {
	engine, err := s.authEngine()
	if err != nil {
		return statusResponse{}, err
	}

	if err := engine.ConfirmPasswordReset(ctx, strings.TrimSpace(req.Token), req.Password); err != nil {
		return statusResponse{}, mapResetPasswordError(err)
	}

	return statusResponse{Message: "Password has been reset successfully."}, nil
}

func (s *service) authEngine() (*goauth.Engine, error) {
	if s == nil || s.engine == nil {
		return nil, apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "auth engine unavailable")
	}
	return s.engine, nil
}

func mapAuthTokenError(err error, invalidMessage string) error {
	if err == nil {
		return nil
	}

	var authErr *goauth.AuthError
	if errors.As(err, &authErr) {
		switch authErr.Category {
		case goauth.CategoryAuthAbuse:
			return apperr.WithCause(apperr.New(apperr.CodeTooManyRequests, http.StatusTooManyRequests, "authentication temporarily limited"), err)
		case goauth.CategoryAuthState:
			return apperr.WithCause(apperr.New(apperr.CodeForbidden, http.StatusForbidden, "authentication state rejected"), err)
		case goauth.CategorySystem:
			if authErr.Code == string(goauth.CodeSystemInternalError) {
				return apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "authentication failed"), err)
			}
			return apperr.WithCause(apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "authentication unavailable"), err)
		default:
			return apperr.WithCause(apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, invalidMessage), err)
		}
	}

	if errors.Is(err, goauth.ErrLoginRateLimited) {
		return apperr.WithCause(apperr.New(apperr.CodeTooManyRequests, http.StatusTooManyRequests, "authentication temporarily limited"), err)
	}

	return apperr.WithCause(apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, invalidMessage), err)
}

func mapSignupError(err error) error {
	switch {
	case errors.Is(err, goauth.ErrAccountExists), errors.Is(err, goauth.ErrProviderDuplicateIdentifier):
		return apperr.WithCause(apperr.New(apperr.CodeConflict, http.StatusConflict, "email already registered"), err)
	case errors.Is(err, goauth.ErrAccountCreationRateLimited):
		return apperr.WithCause(apperr.New(apperr.CodeTooManyRequests, http.StatusTooManyRequests, "signup temporarily limited"), err)
	case errors.Is(err, goauth.ErrAccountCreationUnavailable):
		return apperr.WithCause(apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "signup unavailable"), err)
	case errors.Is(err, goauth.ErrAccountCreationDisabled):
		return apperr.WithCause(apperr.New(apperr.CodeForbidden, http.StatusForbidden, "signup disabled"), err)
	case errors.Is(err, goauth.ErrAccountCreationInvalid), errors.Is(err, goauth.ErrPasswordPolicy), errors.Is(err, goauth.ErrAccountRoleInvalid):
		return apperr.WithCause(apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid signup request"), err)
	default:
		return mapAuthTokenError(err, "signup failed")
	}
}

func mapLogoutError(err error) error {
	switch {
	case errors.Is(err, goauth.ErrTokenInvalid), errors.Is(err, goauth.ErrUnauthorized), errors.Is(err, goauth.ErrSessionNotFound):
		return apperr.WithCause(apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required"), err)
	default:
		return mapAuthTokenError(err, "logout failed")
	}
}

func mapVerifyEmailError(err error) error {
	switch {
	case errors.Is(err, goauth.ErrEmailVerificationInvalid):
		return apperr.WithCause(apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "verification token is invalid"), err)
	case errors.Is(err, goauth.ErrEmailVerificationRateLimited):
		return apperr.WithCause(apperr.New(apperr.CodeTooManyRequests, http.StatusTooManyRequests, "verification temporarily limited"), err)
	case errors.Is(err, goauth.ErrEmailVerificationDisabled):
		return apperr.WithCause(apperr.New(apperr.CodeForbidden, http.StatusForbidden, "email verification is disabled"), err)
	case errors.Is(err, goauth.ErrEmailVerificationUnavailable):
		return apperr.WithCause(apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "email verification unavailable"), err)
	default:
		return mapAuthTokenError(err, "verification failed")
	}
}

func mapResendVerificationError(err error) error {
	switch {
	case errors.Is(err, goauth.ErrUserNotFound):
		return apperr.WithCause(apperr.New(apperr.CodeNotFound, http.StatusNotFound, "account not found"), err)
	case errors.Is(err, goauth.ErrEmailVerificationRateLimited):
		return apperr.WithCause(apperr.New(apperr.CodeTooManyRequests, http.StatusTooManyRequests, "verification temporarily limited"), err)
	case errors.Is(err, goauth.ErrEmailVerificationUnavailable):
		return apperr.WithCause(apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "email verification unavailable"), err)
	default:
		return mapVerifyEmailError(err)
	}
}

func mapForgotPasswordError(err error) error {
	switch {
	case errors.Is(err, goauth.ErrPasswordResetRateLimited):
		return apperr.WithCause(apperr.New(apperr.CodeTooManyRequests, http.StatusTooManyRequests, "password reset temporarily limited"), err)
	case errors.Is(err, goauth.ErrPasswordResetDisabled):
		return apperr.WithCause(apperr.New(apperr.CodeForbidden, http.StatusForbidden, "password reset is disabled"), err)
	case errors.Is(err, goauth.ErrPasswordResetUnavailable):
		return apperr.WithCause(apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "password reset unavailable"), err)
	default:
		return mapAuthTokenError(err, "password reset request failed")
	}
}

func mapResetPasswordError(err error) error {
	switch {
	case errors.Is(err, goauth.ErrPasswordResetInvalid):
		return apperr.WithCause(apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "reset token is invalid"), err)
	case errors.Is(err, goauth.ErrPasswordPolicy), errors.Is(err, goauth.ErrPasswordReuse):
		return apperr.WithCause(apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "password does not satisfy policy"), err)
	case errors.Is(err, goauth.ErrPasswordResetAttempts), errors.Is(err, goauth.ErrPasswordResetRateLimited):
		return apperr.WithCause(apperr.New(apperr.CodeTooManyRequests, http.StatusTooManyRequests, "password reset temporarily limited"), err)
	case errors.Is(err, goauth.ErrPasswordResetUnavailable):
		return apperr.WithCause(apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "password reset unavailable"), err)
	default:
		return mapAuthTokenError(err, "password reset failed")
	}
}

func buildAuthTokenResponse(accessToken, refreshToken string) (authTokenResponse, error) {
	expiresUnix, err := parseJWTExpiryUnix(accessToken)
	if err != nil {
		return authTokenResponse{}, apperr.WithCause(apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "invalid access token payload"), err)
	}

	expiresAt := time.Unix(expiresUnix, 0).UTC()
	return authTokenResponse{
		AccessToken:       accessToken,
		RefreshToken:      refreshToken,
		AccessExpiresUTC:  expiresAt.Format(time.RFC3339),
		AccessExpiresUnix: expiresUnix,
	}, nil
}

func parseJWTExpiryUnix(accessToken string) (int64, error) {
	claims := jwt.MapClaims{}
	if _, _, err := jwt.NewParser().ParseUnverified(accessToken, claims); err != nil {
		return 0, err
	}

	expRaw, ok := claims["exp"]
	if !ok {
		return 0, errors.New("jwt exp claim missing")
	}

	switch exp := expRaw.(type) {
	case float64:
		return int64(exp), nil
	case int64:
		return exp, nil
	case int:
		return int64(exp), nil
	case json.Number:
		value, err := exp.Int64()
		if err != nil {
			return 0, err
		}
		return value, nil
	default:
		return 0, errors.New("jwt exp claim has invalid type")
	}
}
