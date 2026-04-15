package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"strings"
	"time"

	goauth "github.com/MrEthical07/goAuth"
	coreemail "github.com/MrEthical07/superapi/internal/core/email"
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
	RequestChangePasswordOTP(ctx context.Context, userID string, req changePasswordRequestOTPRequest) (statusResponse, error)
	ConfirmChangePassword(ctx context.Context, userID string, req changePasswordConfirmRequest) (statusResponse, error)
}

type service struct {
	engine        *goauth.Engine
	repo          Repo
	emailSender   coreemail.Sender
	webAppBaseURL string
}

// NewService constructs auth business workflows.
func NewService(engine *goauth.Engine, repo Repo, emailSender coreemail.Sender, webAppBaseURL string) Service {
	return &service{
		engine:        engine,
		repo:          repo,
		emailSender:   emailSender,
		webAppBaseURL: strings.TrimSpace(webAppBaseURL),
	}
}

const (
	verificationEmailSubject    = "Verify your ProjectBook account"
	passwordResetEmailSubject   = "Reset your ProjectBook password"
	passwordChangeEmailSubject  = "Confirm your ProjectBook password change"
	passwordChangedEmailSubject = "Your ProjectBook password was changed"

	defaultWebAppBaseURL = "http://localhost:5173"
	verificationPath     = "/auth/verify"
	resetPasswordPath    = "/auth/reset-password"
	accountPath          = "/account"
)

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

	// Signup should not fail if email delivery is unavailable; users can resend after login.
	_, _ = s.issueAndSendVerification(ctx, normalizeEmail(req.Email))

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

	if strings.TrimSpace(req.Token) != "" {
		if err := engine.ConfirmEmailVerification(ctx, strings.TrimSpace(req.Token)); err != nil {
			return statusResponse{}, mapVerifyEmailError(err)
		}
		return statusResponse{Status: "success"}, nil
	}

	if err := engine.ConfirmEmailVerificationCode(ctx, strings.TrimSpace(req.VerificationID), strings.TrimSpace(req.Code)); err != nil {
		return statusResponse{}, mapVerifyEmailError(err)
	}

	return statusResponse{Status: "success"}, nil
}

func (s *service) ResendVerification(ctx context.Context, req resendVerificationRequest) (statusResponse, error) {
	if s == nil || s.repo == nil {
		return statusResponse{}, apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "auth repository unavailable")
	}

	engine, err := s.authEngine()
	if err != nil {
		return statusResponse{}, err
	}

	normalizedEmail := normalizeEmail(req.Email)
	user, found, err := s.repo.LookupUserByEmail(ctx, normalizedEmail)
	if err != nil {
		return statusResponse{}, err
	}

	challenge, err := engine.RequestEmailVerification(ctx, normalizedEmail)
	if err != nil {
		if errors.Is(err, goauth.ErrUserNotFound) {
			return statusResponse{Status: "sent"}, nil
		}
		return statusResponse{}, mapResendVerificationError(err)
	}

	verificationID, code, err := splitVerificationChallenge(challenge)
	if err != nil {
		return statusResponse{}, apperr.WithCause(apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "email verification unavailable"), err)
	}

	if !found || user.IsEmailVerified {
		return statusResponse{Status: "sent"}, nil
	}

	if err := s.sendVerificationEmail(ctx, normalizedEmail, verificationID, code); err != nil {
		return statusResponse{}, mapEmailDeliveryError(err)
	}

	return statusResponse{Status: "sent", VerificationID: verificationID}, nil
}

func (s *service) ForgotPassword(ctx context.Context, req forgotPasswordRequest) (statusResponse, error) {
	if s == nil || s.repo == nil {
		return statusResponse{}, apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "auth repository unavailable")
	}

	engine, err := s.authEngine()
	if err != nil {
		return statusResponse{}, err
	}

	normalizedEmail := normalizeEmail(req.Email)
	_, found, err := s.repo.LookupUserByEmail(ctx, normalizedEmail)
	if err != nil {
		return statusResponse{}, err
	}

	challenge, err := engine.RequestPasswordReset(ctx, normalizedEmail)
	if err != nil {
		return statusResponse{}, mapForgotPasswordError(err)
	}

	challengeID, code, err := splitPasswordResetOTPChallenge(challenge)
	if err != nil {
		return statusResponse{}, apperr.WithCause(apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "password reset unavailable"), err)
	}

	if found {
		if err := s.sendPasswordResetEmail(ctx, normalizedEmail, challengeID, code); err != nil {
			return statusResponse{}, mapEmailDeliveryError(err)
		}
	}

	return statusResponse{
		Status:      "sent",
		ChallengeID: challengeID,
		Message:     "If an account exists for this email, a reset code has been sent.",
	}, nil
}

func (s *service) ResetPassword(ctx context.Context, req resetPasswordRequest) (statusResponse, error) {
	engine, err := s.authEngine()
	if err != nil {
		return statusResponse{}, err
	}

	challenge := strings.TrimSpace(req.Token)
	if challenge == "" {
		challenge = joinPasswordResetOTPChallenge(req.ChallengeID, req.Code)
	}

	if err := engine.ConfirmPasswordReset(ctx, challenge, req.Password); err != nil {
		return statusResponse{}, mapResetPasswordError(err)
	}

	return statusResponse{Message: "Password has been reset successfully."}, nil
}

func (s *service) RequestChangePasswordOTP(ctx context.Context, userID string, req changePasswordRequestOTPRequest) (statusResponse, error) {
	if s == nil || s.repo == nil {
		return statusResponse{}, apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "auth repository unavailable")
	}

	engine, err := s.authEngine()
	if err != nil {
		return statusResponse{}, err
	}

	user, found, err := s.repo.LookupUserByID(ctx, strings.TrimSpace(userID))
	if err != nil {
		return statusResponse{}, err
	}
	if !found {
		return statusResponse{}, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	}

	if err := s.verifyCurrentPassword(ctx, user.Email, req.CurrentPassword); err != nil {
		return statusResponse{}, err
	}

	challenge, err := engine.RequestPasswordReset(ctx, normalizeEmail(user.Email))
	if err != nil {
		return statusResponse{}, mapForgotPasswordError(err)
	}

	challengeID, code, err := splitPasswordResetOTPChallenge(challenge)
	if err != nil {
		return statusResponse{}, apperr.WithCause(apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "password reset unavailable"), err)
	}

	if err := s.sendPasswordChangeOTPEmail(ctx, normalizeEmail(user.Email), challengeID, code); err != nil {
		return statusResponse{}, mapEmailDeliveryError(err)
	}

	return statusResponse{Status: "sent", ChallengeID: challengeID}, nil
}

func (s *service) ConfirmChangePassword(ctx context.Context, userID string, req changePasswordConfirmRequest) (statusResponse, error) {
	if s == nil || s.repo == nil {
		return statusResponse{}, apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "auth repository unavailable")
	}

	engine, err := s.authEngine()
	if err != nil {
		return statusResponse{}, err
	}

	user, found, err := s.repo.LookupUserByID(ctx, strings.TrimSpace(userID))
	if err != nil {
		return statusResponse{}, err
	}
	if !found {
		return statusResponse{}, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	}

	if err := s.verifyCurrentPassword(ctx, user.Email, req.CurrentPassword); err != nil {
		return statusResponse{}, err
	}

	if err := engine.ConfirmPasswordReset(ctx, joinPasswordResetOTPChallenge(req.ChallengeID, req.Code), req.Password); err != nil {
		return statusResponse{}, mapResetPasswordError(err)
	}

	_ = engine.LogoutAll(ctx, strings.TrimSpace(user.ID))
	_ = s.sendPasswordChangedNoticeEmail(ctx, normalizeEmail(user.Email))

	return statusResponse{Message: "Password changed successfully."}, nil
}

func (s *service) authEngine() (*goauth.Engine, error) {
	if s == nil || s.engine == nil {
		return nil, apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "auth engine unavailable")
	}
	return s.engine, nil
}

func (s *service) issueAndSendVerification(ctx context.Context, email string) (string, error) {
	engine, err := s.authEngine()
	if err != nil {
		return "", err
	}

	challenge, err := engine.RequestEmailVerification(ctx, email)
	if err != nil {
		return "", mapResendVerificationError(err)
	}

	verificationID, code, err := splitVerificationChallenge(challenge)
	if err != nil {
		return "", apperr.WithCause(apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "email verification unavailable"), err)
	}

	if err := s.sendVerificationEmail(ctx, email, verificationID, code); err != nil {
		return "", mapEmailDeliveryError(err)
	}

	return verificationID, nil
}

func (s *service) verifyCurrentPassword(ctx context.Context, email, currentPassword string) error {
	engine, err := s.authEngine()
	if err != nil {
		return err
	}

	accessToken, _, err := engine.Login(ctx, normalizeEmail(email), strings.TrimSpace(currentPassword))
	if err != nil {
		return mapAuthTokenError(err, "current password is incorrect")
	}
	if strings.TrimSpace(accessToken) != "" {
		_ = engine.LogoutByAccessToken(ctx, accessToken)
	}

	return nil
}

func (s *service) sendVerificationEmail(ctx context.Context, recipientEmail, verificationID, code string) error {
	if s == nil || s.emailSender == nil {
		return coreemail.ErrSenderUnavailable
	}

	verificationLink := buildWebAppLink(s.webAppBaseURL, verificationPath, map[string]string{
		"verificationId": verificationID,
		"email":          normalizeEmail(recipientEmail),
	})
	textBody := buildVerificationEmailText(code, verificationLink)
	htmlBody := buildVerificationEmailHTML(code, verificationLink)

	return s.emailSender.Send(ctx, coreemail.Message{
		To:       coreemail.NormalizeRecipient(recipientEmail),
		Subject:  verificationEmailSubject,
		HTMLBody: htmlBody,
		TextBody: textBody,
		Flow:     coreemail.FlowVerification,
	})
}

func (s *service) sendPasswordResetEmail(ctx context.Context, recipientEmail, challengeID, code string) error {
	if s == nil || s.emailSender == nil {
		return coreemail.ErrSenderUnavailable
	}

	resetLink := buildWebAppLink(s.webAppBaseURL, resetPasswordPath, map[string]string{
		"challengeId": challengeID,
		"email":       normalizeEmail(recipientEmail),
	})
	textBody := buildPasswordResetEmailText(code, resetLink)
	htmlBody := buildPasswordResetEmailHTML(code, resetLink)

	return s.emailSender.Send(ctx, coreemail.Message{
		To:       coreemail.NormalizeRecipient(recipientEmail),
		Subject:  passwordResetEmailSubject,
		HTMLBody: htmlBody,
		TextBody: textBody,
		Flow:     coreemail.FlowPasswordReset,
	})
}

func (s *service) sendPasswordChangeOTPEmail(ctx context.Context, recipientEmail, challengeID, code string) error {
	if s == nil || s.emailSender == nil {
		return coreemail.ErrSenderUnavailable
	}

	changeLink := buildWebAppLink(s.webAppBaseURL, accountPath, map[string]string{
		"challengeId": challengeID,
	})
	textBody := buildPasswordChangeEmailText(code, changeLink)
	htmlBody := buildPasswordChangeEmailHTML(code, changeLink)

	return s.emailSender.Send(ctx, coreemail.Message{
		To:       coreemail.NormalizeRecipient(recipientEmail),
		Subject:  passwordChangeEmailSubject,
		HTMLBody: htmlBody,
		TextBody: textBody,
		Flow:     coreemail.FlowPasswordChange,
	})
}

func (s *service) sendPasswordChangedNoticeEmail(ctx context.Context, recipientEmail string) error {
	if s == nil || s.emailSender == nil {
		return coreemail.ErrSenderUnavailable
	}

	accountLink := buildWebAppLink(s.webAppBaseURL, accountPath, nil)
	textBody := strings.TrimSpace("Your ProjectBook password has been changed. If this was not you, reset your password immediately and contact support.\n\nManage account security: " + accountLink)
	htmlBody := strings.TrimSpace("<p>Your ProjectBook password has been changed.</p><p>If this was not you, reset your password immediately and contact support.</p><p>Manage account security: <a href=\"" + html.EscapeString(accountLink) + "\">" + html.EscapeString(accountLink) + "</a></p>")

	return s.emailSender.Send(ctx, coreemail.Message{
		To:       coreemail.NormalizeRecipient(recipientEmail),
		Subject:  passwordChangedEmailSubject,
		HTMLBody: htmlBody,
		TextBody: textBody,
		Flow:     coreemail.FlowPasswordChange,
	})
}

func splitVerificationChallenge(challenge string) (string, string, error) {
	parts := strings.SplitN(strings.TrimSpace(challenge), ":", 3)
	if len(parts) != 3 {
		return "", "", fmt.Errorf("invalid verification challenge format")
	}

	verificationID := strings.TrimSpace(parts[1])
	code := strings.TrimSpace(parts[2])
	if verificationID == "" || code == "" {
		return "", "", fmt.Errorf("verification challenge missing fields")
	}

	return verificationID, code, nil
}

func splitPasswordResetOTPChallenge(challenge string) (string, string, error) {
	parts := strings.SplitN(strings.TrimSpace(challenge), ".", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid password reset challenge format")
	}

	challengeID := strings.TrimSpace(parts[0])
	code := strings.TrimSpace(parts[1])
	if challengeID == "" || code == "" {
		return "", "", fmt.Errorf("password reset challenge missing fields")
	}
	if !isNumericCode(code) {
		return "", "", fmt.Errorf("password reset code has invalid format")
	}

	return challengeID, code, nil
}

func joinPasswordResetOTPChallenge(challengeID, code string) string {
	return strings.TrimSpace(challengeID) + "." + strings.TrimSpace(code)
}

func buildWebAppLink(baseURL, routePath string, queryValues map[string]string) string {
	trimmedBaseURL := strings.TrimSpace(baseURL)
	if trimmedBaseURL == "" {
		trimmedBaseURL = defaultWebAppBaseURL
	}

	parsed, err := url.Parse(trimmedBaseURL)
	if err != nil {
		return trimmedBaseURL
	}

	resolvedPath := strings.TrimSpace(routePath)
	if resolvedPath == "" {
		resolvedPath = "/"
	}
	if !strings.HasPrefix(resolvedPath, "/") {
		resolvedPath = "/" + resolvedPath
	}
	resolved := parsed.ResolveReference(&url.URL{Path: resolvedPath})

	query := resolved.Query()
	for key, value := range queryValues {
		trimmedValue := strings.TrimSpace(value)
		if trimmedValue == "" {
			continue
		}
		query.Set(strings.TrimSpace(key), trimmedValue)
	}
	resolved.RawQuery = query.Encode()

	return resolved.String()
}

func buildVerificationEmailText(code, verificationLink string) string {
	return strings.TrimSpace(fmt.Sprintf("Your ProjectBook verification code is: %s\n\nThis OTP expires in 15 minutes and can only be used once.\n\nAfter signing in, open this verification page and enter the code:\n%s", code, verificationLink))
}

func buildVerificationEmailHTML(code, verificationLink string) string {
	escapedCode := html.EscapeString(strings.TrimSpace(code))
	escapedLink := html.EscapeString(strings.TrimSpace(verificationLink))

	return strings.TrimSpace(fmt.Sprintf("<p>Your ProjectBook verification code is:</p><p style=\"font-size:24px;font-weight:700;letter-spacing:0.2em;\">%s</p><p>This OTP expires in 15 minutes and can only be used once.</p><p>After signing in, open this verification page and enter the code:</p><p><a href=\"%s\">%s</a></p>", escapedCode, escapedLink, escapedLink))
}

func buildPasswordResetEmailText(code, resetLink string) string {
	return strings.TrimSpace(fmt.Sprintf("Your ProjectBook password reset code is: %s\n\nThis OTP expires in 15 minutes and can only be used once.\n\nOpen the reset page and enter the code:\n%s", code, resetLink))
}

func buildPasswordResetEmailHTML(code, resetLink string) string {
	escapedCode := html.EscapeString(strings.TrimSpace(code))
	escapedLink := html.EscapeString(strings.TrimSpace(resetLink))

	return strings.TrimSpace(fmt.Sprintf("<p>Your ProjectBook password reset code is:</p><p style=\"font-size:24px;font-weight:700;letter-spacing:0.2em;\">%s</p><p>This OTP expires in 15 minutes and can only be used once.</p><p>Open the reset page and enter the code:</p><p><a href=\"%s\">%s</a></p>", escapedCode, escapedLink, escapedLink))
}

func buildPasswordChangeEmailText(code, changeLink string) string {
	return strings.TrimSpace(fmt.Sprintf("Your ProjectBook password change code is: %s\n\nThis OTP expires in 15 minutes and can only be used once.\n\nOpen your account settings and enter the code:\n%s", code, changeLink))
}

func buildPasswordChangeEmailHTML(code, changeLink string) string {
	escapedCode := html.EscapeString(strings.TrimSpace(code))
	escapedLink := html.EscapeString(strings.TrimSpace(changeLink))

	return strings.TrimSpace(fmt.Sprintf("<p>Your ProjectBook password change code is:</p><p style=\"font-size:24px;font-weight:700;letter-spacing:0.2em;\">%s</p><p>This OTP expires in 15 minutes and can only be used once.</p><p>Open your account settings and enter the code:</p><p><a href=\"%s\">%s</a></p>", escapedCode, escapedLink, escapedLink))
}

func mapAuthTokenError(err error, invalidMessage string) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, goauth.ErrAccountUnverified) {
		return apperr.WithCause(
			apperr.WithDetails(
				apperr.New(apperr.CodeForbidden, http.StatusForbidden, "email verification required"),
				map[string]any{"reason": "email_unverified"},
			),
			err,
		)
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
	case errors.Is(err, goauth.ErrEmailVerificationAttempts):
		return apperr.WithCause(apperr.New(apperr.CodeTooManyRequests, http.StatusTooManyRequests, "verification temporarily limited"), err)
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
	case errors.Is(err, goauth.ErrEmailVerificationAttempts), errors.Is(err, goauth.ErrEmailVerificationRateLimited):
		return apperr.WithCause(apperr.New(apperr.CodeTooManyRequests, http.StatusTooManyRequests, "verification temporarily limited"), err)
	case errors.Is(err, goauth.ErrEmailVerificationDisabled):
		return apperr.WithCause(apperr.New(apperr.CodeForbidden, http.StatusForbidden, "email verification is disabled"), err)
	case errors.Is(err, goauth.ErrEmailVerificationUnavailable):
		return apperr.WithCause(apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "email verification unavailable"), err)
	default:
		return mapVerifyEmailError(err)
	}
}

func mapEmailDeliveryError(err error) error {
	if err == nil {
		return nil
	}
	return apperr.WithCause(apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "email delivery unavailable"), err)
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
		return apperr.WithCause(apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "reset challenge is invalid"), err)
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
