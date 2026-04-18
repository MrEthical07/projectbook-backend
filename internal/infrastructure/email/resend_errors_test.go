package email

import (
	"errors"
	"strings"
	"testing"

	coreemail "github.com/MrEthical07/superapi/internal/core/email"
	"github.com/resend/resend-go/v3"
)

func TestClassifyResendSendError(t *testing.T) {
	t.Run("rate limited", func(t *testing.T) {
		err := classifyResendSendError(resend.ErrRateLimit)
		if !errors.Is(err, coreemail.ErrRateLimited) {
			t.Fatalf("expected ErrRateLimited, got %v", err)
		}
	})

	t.Run("invalid recipient", func(t *testing.T) {
		err := classifyResendSendError(errors.New("[ERROR]: Invalid `to` field"))
		if !errors.Is(err, coreemail.ErrInvalidRecipient) {
			t.Fatalf("expected ErrInvalidRecipient, got %v", err)
		}
	})

	t.Run("sender identity rejected", func(t *testing.T) {
		err := classifyResendSendError(errors.New("[ERROR]: The from address does not match a verified Sender Identity"))
		if !errors.Is(err, coreemail.ErrSenderIdentityRejected) {
			t.Fatalf("expected ErrSenderIdentityRejected, got %v", err)
		}
	})

	t.Run("provider unavailable", func(t *testing.T) {
		err := classifyResendSendError(errors.New("dial tcp timeout"))
		if !errors.Is(err, coreemail.ErrProviderUnavailable) {
			t.Fatalf("expected ErrProviderUnavailable, got %v", err)
		}
	})
}

func TestResolveTransactionalFromFallback(t *testing.T) {
	sender := &ResendSender{
		profiles: normalizeProfiles(SenderProfiles{
			Transactional: coreemail.SenderIdentity{Name: "ProjectBook", Email: "no-reply@projectbook.dev"},
			Verification:  coreemail.SenderIdentity{Name: "ProjectBook Verify", Email: "verify@projectbook.dev"},
		}),
	}

	message := coreemail.Message{Flow: coreemail.FlowVerification}
	fallbackFrom, ok := sender.resolveTransactionalFrom(
		message,
		"ProjectBook Verify <verify@projectbook.dev>",
		errors.New("[ERROR]: from address is not verified"),
	)
	if !ok {
		t.Fatal("expected fallback sender")
	}
	if !strings.Contains(strings.ToLower(fallbackFrom), "no-reply@projectbook.dev") {
		t.Fatalf("expected transactional sender email in fallback, got %q", fallbackFrom)
	}
}

func TestResolveTransactionalFromNoFallbackForExplicitFrom(t *testing.T) {
	sender := &ResendSender{
		profiles: normalizeProfiles(SenderProfiles{
			Transactional: coreemail.SenderIdentity{Name: "ProjectBook", Email: "no-reply@projectbook.dev"},
			Verification:  coreemail.SenderIdentity{Name: "ProjectBook Verify", Email: "verify@projectbook.dev"},
		}),
	}

	message := coreemail.Message{
		Flow: coreemail.FlowVerification,
		From: coreemail.SenderIdentity{Email: "custom@projectbook.dev", Name: "Custom"},
	}
	if _, ok := sender.resolveTransactionalFrom(message, "Custom <custom@projectbook.dev>", errors.New("[ERROR]: from address is not verified")); ok {
		t.Fatal("did not expect fallback when message has explicit from")
	}
}
