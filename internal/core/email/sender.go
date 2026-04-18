package email

import (
	"context"
	"errors"
	"strings"
)

var ErrSenderUnavailable = errors.New("email sender unavailable")
var ErrProviderUnavailable = errors.New("email provider unavailable")
var ErrRateLimited = errors.New("email provider rate limited")
var ErrInvalidRecipient = errors.New("email recipient invalid")
var ErrSenderIdentityRejected = errors.New("email sender identity rejected")

// Flow identifies the transactional email flow type for sender selection.
type Flow string

const (
	FlowTransactional  Flow = "transactional"
	FlowVerification   Flow = "verification"
	FlowPasswordReset  Flow = "password_reset"
	FlowPasswordChange Flow = "password_change"
)

// SenderIdentity describes an email sender display name and address.
type SenderIdentity struct {
	Name  string
	Email string
}

// Message describes a generic transactional email payload.
type Message struct {
	To       string
	Subject  string
	HTMLBody string
	TextBody string
	Flow     Flow
	From     SenderIdentity
}

// Sender sends transactional email messages.
type Sender interface {
	Send(ctx context.Context, message Message) error
}

// NoopSender keeps startup flexible when email is disabled.
type NoopSender struct{}

func (NoopSender) Send(context.Context, Message) error {
	return ErrSenderUnavailable
}

// NormalizeSenderIdentity trims sender display metadata.
func NormalizeSenderIdentity(value SenderIdentity) SenderIdentity {
	return SenderIdentity{
		Name:  strings.TrimSpace(value.Name),
		Email: NormalizeRecipient(value.Email),
	}
}

// NormalizeRecipient trims and lowercases recipient email values.
func NormalizeRecipient(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
