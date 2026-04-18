package auth

import (
	"net/http"
	"testing"

	goauth "github.com/MrEthical07/goAuth"
	apperr "github.com/MrEthical07/superapi/internal/core/errors"
)

func TestMapAuthTokenErrorUnverifiedIncludesReason(t *testing.T) {
	err := mapAuthTokenError(goauth.ErrAccountUnverified, "invalid credentials")
	ae, ok := apperr.AsAppError(err)
	if !ok {
		t.Fatal("expected app error")
	}
	if ae.Code != apperr.CodeForbidden {
		t.Fatalf("expected code %s, got %s", apperr.CodeForbidden, ae.Code)
	}
	if ae.StatusCode != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, ae.StatusCode)
	}
	details, ok := ae.Details.(map[string]any)
	if !ok {
		t.Fatalf("expected details map, got %T", ae.Details)
	}
	if details["reason"] != "email_unverified" {
		t.Fatalf("expected reason email_unverified, got %v", details["reason"])
	}
}

func TestMapAuthTokenErrorUnverifiedIncludesVerificationID(t *testing.T) {
	err := mapAuthTokenError(
		goauth.ErrAccountUnverified,
		"invalid credentials",
		map[string]any{"verificationId": "verify-123"},
	)
	ae, ok := apperr.AsAppError(err)
	if !ok {
		t.Fatal("expected app error")
	}
	details, ok := ae.Details.(map[string]any)
	if !ok {
		t.Fatalf("expected details map, got %T", ae.Details)
	}
	if details["reason"] != "email_unverified" {
		t.Fatalf("expected reason email_unverified, got %v", details["reason"])
	}
	if details["verificationId"] != "verify-123" {
		t.Fatalf("expected verificationId verify-123, got %v", details["verificationId"])
	}
}
