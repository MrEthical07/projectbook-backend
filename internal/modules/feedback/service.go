package feedback

import (
	"context"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"strings"
	"time"

	coreemail "github.com/MrEthical07/superapi/internal/core/email"
	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/requestid"
	"github.com/MrEthical07/superapi/internal/core/storage"
)

const (
	feedbackMailbox     = "feedback@projectbook.dev"
	feedbackEmailWindow = 20 * time.Second
)

// Service defines feedback module business workflows.
type Service interface {
	Submit(ctx context.Context, userID string, req submitFeedbackRequest) (submitFeedbackResponse, error)
}

type service struct {
	store       storage.RelationalStore
	repo        Repo
	emailSender coreemail.Sender
}

// NewService constructs feedback business workflows.
func NewService(store storage.RelationalStore, repo Repo, emailSender coreemail.Sender) Service {
	return &service{store: store, repo: repo, emailSender: emailSender}
}

func (s *service) Submit(ctx context.Context, userID string, req submitFeedbackRequest) (submitFeedbackResponse, error) {
	if strings.TrimSpace(userID) == "" {
		return submitFeedbackResponse{}, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	}
	if err := req.Validate(); err != nil {
		return submitFeedbackResponse{}, err
	}
	if err := s.requireStore(); err != nil {
		return submitFeedbackResponse{}, err
	}

	input := createFeedbackSubmissionInput{
		UserID:            strings.TrimSpace(userID),
		ProjectID:         strings.TrimSpace(req.Context.ProjectID),
		Subject:           strings.TrimSpace(req.Subject),
		Message:           strings.TrimSpace(req.Message),
		PagePath:          strings.TrimSpace(req.Context.Path),
		UserAgent:         strings.TrimSpace(req.Context.UserAgent),
		Mode:              strings.TrimSpace(req.Context.Mode),
		ClientSubmittedAt: strings.TrimSpace(req.Context.SubmittedAt),
		RequestID:         strings.TrimSpace(requestid.FromContext(ctx)),
	}

	var created feedbackSubmissionRecord
	err := s.store.WithTx(ctx, func(txCtx context.Context) error {
		createdRecord, createErr := s.repo.CreateSubmission(txCtx, input)
		if createErr != nil {
			return createErr
		}
		created = createdRecord
		return nil
	})
	if err != nil {
		return submitFeedbackResponse{}, mapFeedbackRepoError(err)
	}

	slog.Info("feedback submission accepted",
		"feedback_id", created.ID,
		"user_id", created.UserID,
		"project_id", created.ProjectID,
		"path", created.PagePath,
	)

	s.dispatchFeedbackEmail(created)

	return submitFeedbackResponse{
		FeedbackID:  created.ID,
		Status:      "accepted",
		SubmittedAt: created.SubmittedAt.UTC().Format(time.RFC3339),
	}, nil
}

func (s *service) dispatchFeedbackEmail(record feedbackSubmissionRecord) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), feedbackEmailWindow)
		defer cancel()

		if err := s.markEmailQueued(ctx, record.ID); err != nil {
			slog.Warn("feedback email queue status update failed",
				"feedback_id", record.ID,
				"error", err.Error(),
			)
		}

		if s.emailSender == nil {
			err := coreemail.ErrSenderUnavailable
			s.markEmailFailed(ctx, record.ID, err.Error())
			slog.Error("feedback email sender unavailable",
				"feedback_id", record.ID,
				"error", err.Error(),
			)
			return
		}

		emailMessage := buildFeedbackEmailMessage(record)
		if err := s.emailSender.Send(ctx, emailMessage); err != nil {
			s.markEmailFailed(ctx, record.ID, err.Error())
			slog.Error("feedback email dispatch failed",
				"feedback_id", record.ID,
				"error", err.Error(),
			)
			return
		}

		if err := s.markEmailSent(ctx, record.ID); err != nil {
			slog.Warn("feedback email sent status update failed",
				"feedback_id", record.ID,
				"error", err.Error(),
			)
		}

		slog.Info("feedback email dispatched",
			"feedback_id", record.ID,
			"user_id", record.UserID,
		)
	}()
}

func (s *service) markEmailQueued(ctx context.Context, feedbackID string) error {
	if err := s.requireStore(); err != nil {
		return err
	}
	return s.store.WithTx(ctx, func(txCtx context.Context) error {
		return s.repo.MarkEmailQueued(txCtx, feedbackID)
	})
}

func (s *service) markEmailSent(ctx context.Context, feedbackID string) error {
	if err := s.requireStore(); err != nil {
		return err
	}
	return s.store.WithTx(ctx, func(txCtx context.Context) error {
		return s.repo.MarkEmailSent(txCtx, feedbackID)
	})
}

func (s *service) markEmailFailed(ctx context.Context, feedbackID, reason string) {
	if err := s.requireStore(); err != nil {
		slog.Warn("feedback email failure status update skipped",
			"feedback_id", feedbackID,
			"error", err.Error(),
		)
		return
	}
	if err := s.store.WithTx(ctx, func(txCtx context.Context) error {
		return s.repo.MarkEmailFailed(txCtx, feedbackID, reason)
	}); err != nil {
		slog.Warn("feedback email failure status update failed",
			"feedback_id", feedbackID,
			"error", err.Error(),
		)
	}
}

func buildFeedbackEmailMessage(record feedbackSubmissionRecord) coreemail.Message {
	projectID := strings.TrimSpace(record.ProjectID)
	if projectID == "" {
		projectID = "n/a"
	}
	userName := strings.TrimSpace(record.UserName)
	if userName == "" {
		userName = "Unknown"
	}
	userEmail := strings.TrimSpace(record.UserEmail)
	if userEmail == "" {
		userEmail = "unknown"
	}
	path := strings.TrimSpace(record.PagePath)
	if path == "" {
		path = "n/a"
	}

	textBody := strings.Join([]string{
		"New ProjectBook feedback submission received.",
		"",
		"Feedback ID: " + record.ID,
		"Submitted At: " + record.SubmittedAt.UTC().Format(time.RFC3339),
		"User ID: " + record.UserID,
		"User Name: " + userName,
		"User Email: " + userEmail,
		"Project ID: " + projectID,
		"Path: " + path,
		"",
		"Subject:",
		record.Subject,
		"",
		"Message:",
		record.Message,
		"",
		"Context JSON:",
		record.ContextRaw,
	}, "\n")

	htmlBody := strings.Join([]string{
		"<p><strong>New ProjectBook feedback submission received.</strong></p>",
		"<p><strong>Feedback ID:</strong> " + html.EscapeString(record.ID) + "</p>",
		"<p><strong>Submitted At:</strong> " + html.EscapeString(record.SubmittedAt.UTC().Format(time.RFC3339)) + "</p>",
		"<p><strong>User ID:</strong> " + html.EscapeString(record.UserID) + "</p>",
		"<p><strong>User Name:</strong> " + html.EscapeString(userName) + "</p>",
		"<p><strong>User Email:</strong> " + html.EscapeString(userEmail) + "</p>",
		"<p><strong>Project ID:</strong> " + html.EscapeString(projectID) + "</p>",
		"<p><strong>Path:</strong> " + html.EscapeString(path) + "</p>",
		"<h3>Subject</h3>",
		"<p>" + html.EscapeString(record.Subject) + "</p>",
		"<h3>Message</h3>",
		"<p>" + strings.ReplaceAll(html.EscapeString(record.Message), "\n", "<br>") + "</p>",
		"<h3>Context</h3>",
		"<pre>" + html.EscapeString(record.ContextRaw) + "</pre>",
	}, "")

	trimmedSubject := strings.TrimSpace(record.Subject)
	if trimmedSubject == "" {
		trimmedSubject = "Untitled feedback"
	}

	return coreemail.Message{
		To:       coreemail.NormalizeRecipient(feedbackMailbox),
		Subject:  "[ProjectBook Feedback] " + trimmedSubject,
		TextBody: textBody,
		HTMLBody: htmlBody,
		Flow:     coreemail.FlowTransactional,
	}
}

func (s *service) requireStore() error {
	if s == nil || s.store == nil || s.repo == nil {
		return apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "feedback service unavailable")
	}
	return nil
}

func mapFeedbackRepoError(err error) error {
	if err == nil {
		return nil
	}
	if ae, ok := apperr.AsAppError(err); ok {
		return ae
	}

	switch {
	case errors.Is(err, ErrFeedbackUserNotFound):
		return apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	case errors.Is(err, ErrFeedbackInvalidProjectID):
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "context.projectId is invalid")
	default:
		return apperr.WithCause(
			apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "feedback submission failed"),
			fmt.Errorf("submit feedback: %w", err),
		)
	}
}
