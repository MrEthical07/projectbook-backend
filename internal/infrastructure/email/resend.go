package email

import (
	"context"
	"fmt"
	"net/mail"
	"strings"

	coreemail "github.com/MrEthical07/superapi/internal/core/email"
	"github.com/resend/resend-go/v3"
)

// SenderProfiles configures sender identities per auth flow.
type SenderProfiles struct {
	Transactional  coreemail.SenderIdentity
	Verification   coreemail.SenderIdentity
	PasswordReset  coreemail.SenderIdentity
	PasswordChange coreemail.SenderIdentity
}

// ResendSender sends transactional email through the Resend API.
type ResendSender struct {
	client   *resend.Client
	profiles SenderProfiles
}

// NewResendSender builds a sender backed by Resend.
func NewResendSender(apiKey string, profiles SenderProfiles) (*ResendSender, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("resend api key is required")
	}

	profiles = normalizeProfiles(profiles)
	if err := validateSenderIdentity(profiles.Transactional); err != nil {
		return nil, fmt.Errorf("transactional sender: %w", err)
	}
	if profiles.Verification.Email != "" {
		if err := validateSenderIdentity(profiles.Verification); err != nil {
			return nil, fmt.Errorf("verification sender: %w", err)
		}
	}
	if profiles.PasswordReset.Email != "" {
		if err := validateSenderIdentity(profiles.PasswordReset); err != nil {
			return nil, fmt.Errorf("password reset sender: %w", err)
		}
	}
	if profiles.PasswordChange.Email != "" {
		if err := validateSenderIdentity(profiles.PasswordChange); err != nil {
			return nil, fmt.Errorf("password change sender: %w", err)
		}
	}

	return &ResendSender{
		client:   resend.NewClient(strings.TrimSpace(apiKey)),
		profiles: profiles,
	}, nil
}

func (s *ResendSender) Send(ctx context.Context, message coreemail.Message) error {
	if s == nil || s.client == nil {
		return coreemail.ErrSenderUnavailable
	}

	to := coreemail.NormalizeRecipient(message.To)
	if to == "" {
		return fmt.Errorf("recipient is required")
	}

	subject := strings.TrimSpace(message.Subject)
	if subject == "" {
		return fmt.Errorf("subject is required")
	}

	htmlBody := strings.TrimSpace(message.HTMLBody)
	textBody := strings.TrimSpace(message.TextBody)
	if htmlBody == "" && textBody == "" {
		return fmt.Errorf("email body is required")
	}

	from, err := s.resolveFrom(message)
	if err != nil {
		return err
	}

	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{to},
		Subject: subject,
		Html:    htmlBody,
		Text:    textBody,
	}

	if _, err := s.client.Emails.SendWithContext(ctx, params); err != nil {
		return fmt.Errorf("resend send email: %w", err)
	}

	return nil
}

func (s *ResendSender) resolveFrom(message coreemail.Message) (string, error) {
	identity := coreemail.NormalizeSenderIdentity(message.From)
	if identity.Email == "" {
		identity = s.profileForFlow(message.Flow)
	}
	if err := validateSenderIdentity(identity); err != nil {
		return "", fmt.Errorf("sender identity invalid: %w", err)
	}

	if identity.Name == "" {
		return identity.Email, nil
	}

	return (&mail.Address{Name: identity.Name, Address: identity.Email}).String(), nil
}

func (s *ResendSender) profileForFlow(flow coreemail.Flow) coreemail.SenderIdentity {
	if s == nil {
		return coreemail.SenderIdentity{}
	}

	switch flow {
	case coreemail.FlowVerification:
		if s.profiles.Verification.Email != "" {
			return s.profiles.Verification
		}
	case coreemail.FlowPasswordReset:
		if s.profiles.PasswordReset.Email != "" {
			return s.profiles.PasswordReset
		}
	case coreemail.FlowPasswordChange:
		if s.profiles.PasswordChange.Email != "" {
			return s.profiles.PasswordChange
		}
	}

	return s.profiles.Transactional
}

func normalizeProfiles(profiles SenderProfiles) SenderProfiles {
	profiles.Transactional = coreemail.NormalizeSenderIdentity(profiles.Transactional)
	profiles.Verification = coreemail.NormalizeSenderIdentity(profiles.Verification)
	profiles.PasswordReset = coreemail.NormalizeSenderIdentity(profiles.PasswordReset)
	profiles.PasswordChange = coreemail.NormalizeSenderIdentity(profiles.PasswordChange)

	if profiles.Transactional.Name == "" && profiles.Transactional.Email != "" {
		profiles.Transactional.Name = emailLocalPart(profiles.Transactional.Email)
	}
	if profiles.Verification.Name == "" {
		profiles.Verification.Name = profiles.Transactional.Name
	}
	if profiles.PasswordReset.Name == "" {
		profiles.PasswordReset.Name = profiles.Transactional.Name
	}
	if profiles.PasswordChange.Name == "" {
		profiles.PasswordChange.Name = profiles.Transactional.Name
	}

	return profiles
}

func validateSenderIdentity(identity coreemail.SenderIdentity) error {
	if identity.Email == "" {
		return fmt.Errorf("email is required")
	}

	parsed, err := mail.ParseAddress(identity.Email)
	if err != nil {
		return fmt.Errorf("invalid email address")
	}
	if !strings.EqualFold(parsed.Address, identity.Email) {
		return fmt.Errorf("invalid email address")
	}

	return nil
}

func emailLocalPart(email string) string {
	parts := strings.SplitN(strings.TrimSpace(email), "@", 2)
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		return "sender"
	}
	return strings.TrimSpace(parts[0])
}
