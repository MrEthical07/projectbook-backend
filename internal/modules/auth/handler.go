package auth

import (
	"net/http"
	"strings"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/httpx"
)

// Handler contains HTTP transport handlers for auth routes.
type Handler struct {
	svc Service
}

// NewHandler constructs auth transport handlers.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Signup(ctx *httpx.Context, req signupRequest) (httpx.Result[signupResponse], error) {
	resp, err := h.svc.Signup(ctx.Context(), req)
	if err != nil {
		return httpx.Result[signupResponse]{}, err
	}

	return httpx.Result[signupResponse]{
		Status: http.StatusCreated,
		Data:   resp,
	}, nil
}

func (h *Handler) Login(ctx *httpx.Context, req loginRequest) (authTokenResponse, error) {
	return h.svc.Login(ctx.Context(), req)
}

func (h *Handler) Refresh(ctx *httpx.Context, req refreshRequest) (authTokenResponse, error) {
	return h.svc.Refresh(ctx.Context(), req)
}

func (h *Handler) Logout(ctx *httpx.Context, _ httpx.NoBody) (httpx.Result[any], error) {
	accessToken, err := bearerTokenFromHeader(ctx.Header("Authorization"))
	if err != nil {
		return httpx.Result[any]{}, err
	}

	if err := h.svc.Logout(ctx.Context(), accessToken); err != nil {
		return httpx.Result[any]{}, err
	}

	return httpx.Result[any]{
		Status: http.StatusOK,
		Data:   nil,
	}, nil
}

func (h *Handler) VerifyEmail(ctx *httpx.Context, req verifyEmailRequest) (statusResponse, error) {
	return h.svc.VerifyEmail(ctx.Context(), req)
}

func (h *Handler) ResendVerification(ctx *httpx.Context, req resendVerificationRequest) (statusResponse, error) {
	return h.svc.ResendVerification(ctx.Context(), req)
}

func (h *Handler) ForgotPassword(ctx *httpx.Context, req forgotPasswordRequest) (statusResponse, error) {
	return h.svc.ForgotPassword(ctx.Context(), req)
}

func (h *Handler) ResetPassword(ctx *httpx.Context, req resetPasswordRequest) (statusResponse, error) {
	return h.svc.ResetPassword(ctx.Context(), req)
}

func (h *Handler) RequestChangePasswordOTP(ctx *httpx.Context, req changePasswordRequestOTPRequest) (statusResponse, error) {
	principal, ok := ctx.Auth()
	if !ok || strings.TrimSpace(principal.UserID) == "" {
		return statusResponse{}, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	}

	return h.svc.RequestChangePasswordOTP(ctx.Context(), strings.TrimSpace(principal.UserID), req)
}

func (h *Handler) ConfirmChangePassword(ctx *httpx.Context, req changePasswordConfirmRequest) (statusResponse, error) {
	principal, ok := ctx.Auth()
	if !ok || strings.TrimSpace(principal.UserID) == "" {
		return statusResponse{}, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	}

	return h.svc.ConfirmChangePassword(ctx.Context(), strings.TrimSpace(principal.UserID), req)
}

func bearerTokenFromHeader(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	}

	parts := strings.SplitN(trimmed, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(strings.TrimSpace(parts[0]), "Bearer") {
		return "", apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	}

	return token, nil
}
