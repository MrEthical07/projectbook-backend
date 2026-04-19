package feedback

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrFeedbackUserNotFound     = errors.New("feedback user not found")
	ErrFeedbackInvalidProjectID = errors.New("feedback project id invalid")
)

// Repo defines feedback module persistence operations.
type Repo interface {
	CreateSubmission(ctx context.Context, input createFeedbackSubmissionInput) (feedbackSubmissionRecord, error)
	MarkEmailQueued(ctx context.Context, submissionID string) error
	MarkEmailSent(ctx context.Context, submissionID string) error
	MarkEmailFailed(ctx context.Context, submissionID, reason string) error
}

type repo struct {
	store storage.RelationalStore
}

// NewRepo constructs a relational feedback repository.
func NewRepo(store storage.RelationalStore) Repo {
	return &repo{store: store}
}

const queryCreateFeedbackSubmission = `
WITH principal AS (
	SELECT id, name, email
	FROM users
	WHERE id = $1::uuid
), inserted AS (
	INSERT INTO global_feedback_submissions (
		user_id,
		project_id,
		subject,
		message,
		page_path,
		context
	)
	SELECT
		principal.id,
		NULLIF($2, '')::uuid,
		$3,
		$4,
		$5,
		$6::jsonb
	FROM principal
	RETURNING id, user_id, project_id, subject, message, page_path, context, submitted_at
)
SELECT
	inserted.id::text,
	inserted.user_id::text,
	COALESCE(principal.name, ''),
	COALESCE(principal.email, ''),
	COALESCE(inserted.project_id::text, ''),
	inserted.subject,
	inserted.message,
	inserted.page_path,
	inserted.context::text,
	inserted.submitted_at
FROM inserted
JOIN principal ON principal.id = inserted.user_id
LIMIT 1
`

const queryMarkFeedbackEmailQueued = `
UPDATE global_feedback_submissions
SET
	email_queued_at = COALESCE(email_queued_at, NOW()),
	email_error = NULL
WHERE id = $1::uuid
`

const queryMarkFeedbackEmailSent = `
UPDATE global_feedback_submissions
SET
	email_sent_at = NOW(),
	email_error = NULL
WHERE id = $1::uuid
`

const queryMarkFeedbackEmailFailed = `
UPDATE global_feedback_submissions
SET
	email_error = $2,
	email_sent_at = NULL
WHERE id = $1::uuid
`

func (r *repo) CreateSubmission(ctx context.Context, input createFeedbackSubmissionInput) (feedbackSubmissionRecord, error) {
	if err := r.requireStore(); err != nil {
		return feedbackSubmissionRecord{}, err
	}

	contextPayload, err := json.Marshal(map[string]string{
		"path":              strings.TrimSpace(input.PagePath),
		"userAgent":         strings.TrimSpace(input.UserAgent),
		"mode":              strings.TrimSpace(input.Mode),
		"submittedAt":       strings.TrimSpace(input.ClientSubmittedAt),
		"requestId":         strings.TrimSpace(input.RequestID),
		"projectId":         strings.TrimSpace(input.ProjectID),
		"authenticatedUser": strings.TrimSpace(input.UserID),
	})
	if err != nil {
		return feedbackSubmissionRecord{}, wrapRepoError("marshal feedback context", err)
	}

	record := feedbackSubmissionRecord{}
	err = r.store.Execute(ctx, storage.RelationalQueryOne(
		queryCreateFeedbackSubmission,
		func(row storage.RowScanner) error {
			return row.Scan(
				&record.ID,
				&record.UserID,
				&record.UserName,
				&record.UserEmail,
				&record.ProjectID,
				&record.Subject,
				&record.Message,
				&record.PagePath,
				&record.ContextRaw,
				&record.SubmittedAt,
			)
		},
		strings.TrimSpace(input.UserID),
		strings.TrimSpace(input.ProjectID),
		strings.TrimSpace(input.Subject),
		strings.TrimSpace(input.Message),
		strings.TrimSpace(input.PagePath),
		string(contextPayload),
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return feedbackSubmissionRecord{}, ErrFeedbackUserNotFound
		}
		if isInvalidUUIDTextError(err) {
			return feedbackSubmissionRecord{}, ErrFeedbackInvalidProjectID
		}
		return feedbackSubmissionRecord{}, wrapRepoError("create feedback submission", err)
	}

	return record, nil
}

func (r *repo) MarkEmailQueued(ctx context.Context, submissionID string) error {
	if err := r.requireStore(); err != nil {
		return err
	}
	return r.store.Execute(ctx, storage.RelationalExec(queryMarkFeedbackEmailQueued, strings.TrimSpace(submissionID)))
}

func (r *repo) MarkEmailSent(ctx context.Context, submissionID string) error {
	if err := r.requireStore(); err != nil {
		return err
	}
	return r.store.Execute(ctx, storage.RelationalExec(queryMarkFeedbackEmailSent, strings.TrimSpace(submissionID)))
}

func (r *repo) MarkEmailFailed(ctx context.Context, submissionID, reason string) error {
	if err := r.requireStore(); err != nil {
		return err
	}
	trimmedReason := strings.TrimSpace(reason)
	if len(trimmedReason) > 1024 {
		trimmedReason = trimmedReason[:1024]
	}
	return r.store.Execute(ctx, storage.RelationalExec(queryMarkFeedbackEmailFailed, strings.TrimSpace(submissionID), trimmedReason))
}

func (r *repo) requireStore() error {
	if r == nil || r.store == nil {
		return apperr.New(apperr.CodeDependencyFailure, http.StatusServiceUnavailable, "feedback repository unavailable")
	}
	return nil
}

func wrapRepoError(action string, err error) error {
	if err == nil {
		return nil
	}
	return apperr.WithCause(
		apperr.New(apperr.CodeInternal, http.StatusInternalServerError, "failed to process feedback data"),
		fmt.Errorf("%s: %w", action, err),
	)
}

func isInvalidUUIDTextError(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "22P02"
	}
	return false
}
