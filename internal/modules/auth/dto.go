package auth

import (
	"net/http"
	"net/mail"
	"strings"
	"unicode"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
)

type signupRequest struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

func (r signupRequest) Validate() error {
	if len(strings.TrimSpace(r.Name)) < 2 {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "name must be at least 2 characters")
	}
	if !isValidEmail(r.Email) {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "email must be a valid email address")
	}
	if err := validateStrongPassword(r.Password); err != nil {
		return err
	}
	if r.Password != r.ConfirmPassword {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "confirmPassword must match password")
	}
	return nil
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

func (r loginRequest) Validate() error {
	if !isValidEmail(r.Email) {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "email must be a valid email address")
	}
	if strings.TrimSpace(r.Password) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "password is required")
	}
	return nil
}

type verifyEmailRequest struct {
	Token string `json:"token"`
}

func (r verifyEmailRequest) Validate() error {
	if strings.TrimSpace(r.Token) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "token is required")
	}
	return nil
}

type resendVerificationRequest struct {
	Email string `json:"email"`
}

func (r resendVerificationRequest) Validate() error {
	if !isValidEmail(r.Email) {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "email must be a valid email address")
	}
	return nil
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

func (r forgotPasswordRequest) Validate() error {
	if !isValidEmail(r.Email) {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "email must be a valid email address")
	}
	return nil
}

type resetPasswordRequest struct {
	Token           string `json:"token"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

func (r resetPasswordRequest) Validate() error {
	if strings.TrimSpace(r.Token) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "token is required")
	}
	if err := validateStrongPassword(r.Password); err != nil {
		return err
	}
	if r.Password != r.ConfirmPassword {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "confirmPassword must match password")
	}
	return nil
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (r refreshRequest) Validate() error {
	if strings.TrimSpace(r.RefreshToken) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "refresh_token is required")
	}
	return nil
}

type signupUserResponse struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	IsEmailVerified bool   `json:"isEmailVerified"`
}

type signupResponse struct {
	User signupUserResponse `json:"user"`
}

type authTokenResponse struct {
	AccessToken       string `json:"access_token"`
	RefreshToken      string `json:"refresh_token"`
	AccessExpiresUTC  string `json:"access_expires_utc"`
	AccessExpiresUnix int64  `json:"access_expires_unix"`
}

type statusResponse struct {
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}

func normalizeEmail(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func isValidEmail(email string) bool {
	normalized := normalizeEmail(email)
	if normalized == "" {
		return false
	}
	parsed, err := mail.ParseAddress(normalized)
	if err != nil {
		return false
	}
	return strings.EqualFold(parsed.Address, normalized)
}

func validateStrongPassword(password string) error {
	if len(password) < 10 {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "password must be at least 10 characters")
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "password must include uppercase, lowercase, number, and special character")
	}

	return nil
}
