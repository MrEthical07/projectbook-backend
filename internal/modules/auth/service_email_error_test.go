package auth

import (
	"errors"
	"net/http"
	"testing"

	coreemail "github.com/MrEthical07/superapi/internal/core/email"
	apperr "github.com/MrEthical07/superapi/internal/core/errors"
)

func TestMapEmailDeliveryErrorRateLimited(t *testing.T) {
	err := mapEmailDeliveryError(errors.Join(coreemail.ErrRateLimited, errors.New("provider returned 429")))
	ae, ok := apperr.AsAppError(err)
	if !ok {
		t.Fatal("expected app error")
	}
	if ae.Code != apperr.CodeTooManyRequests {
		t.Fatalf("expected code %s, got %s", apperr.CodeTooManyRequests, ae.Code)
	}
	if ae.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, ae.StatusCode)
	}
}

func TestMapEmailDeliveryErrorInvalidRecipient(t *testing.T) {
	err := mapEmailDeliveryError(errors.Join(coreemail.ErrInvalidRecipient, errors.New("invalid to field")))
	ae, ok := apperr.AsAppError(err)
	if !ok {
		t.Fatal("expected app error")
	}
	if ae.Code != apperr.CodeBadRequest {
		t.Fatalf("expected code %s, got %s", apperr.CodeBadRequest, ae.Code)
	}
	if ae.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, ae.StatusCode)
	}
}

func TestMapEmailDeliveryErrorSenderIdentityRejected(t *testing.T) {
	err := mapEmailDeliveryError(errors.Join(coreemail.ErrSenderIdentityRejected, errors.New("from not verified")))
	ae, ok := apperr.AsAppError(err)
	if !ok {
		t.Fatal("expected app error")
	}
	if ae.Code != apperr.CodeDependencyFailure {
		t.Fatalf("expected code %s, got %s", apperr.CodeDependencyFailure, ae.Code)
	}
	if ae.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, ae.StatusCode)
	}
	details, ok := ae.Details.(map[string]any)
	if !ok {
		t.Fatalf("expected details map, got %T", ae.Details)
	}
	if details["reason"] != "sender_identity_unverified" {
		t.Fatalf("expected details reason sender_identity_unverified, got %v", details["reason"])
	}
}

func TestMapEmailDeliveryErrorDefaultDependencyFailure(t *testing.T) {
	err := mapEmailDeliveryError(errors.New("network timeout"))
	ae, ok := apperr.AsAppError(err)
	if !ok {
		t.Fatal("expected app error")
	}
	if ae.Code != apperr.CodeDependencyFailure {
		t.Fatalf("expected code %s, got %s", apperr.CodeDependencyFailure, ae.Code)
	}
	if ae.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, ae.StatusCode)
	}
}
