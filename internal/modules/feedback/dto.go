package feedback

import (
	"net/http"
	"strings"
	"time"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
)

const (
	maxFeedbackSubjectLength     = 160
	maxFeedbackMessageLength     = 5000
	maxFeedbackProjectIDLength   = 64
	maxFeedbackPathLength        = 512
	maxFeedbackUserAgentLength   = 512
	maxFeedbackModeLength        = 32
	maxFeedbackSubmittedAtLength = 64
)

type submitFeedbackRequest struct {
	Subject string                `json:"subject"`
	Message string                `json:"message"`
	Context submitFeedbackContext `json:"context"`
}

type submitFeedbackContext struct {
	ProjectID   string `json:"projectId,omitempty"`
	Path        string `json:"path,omitempty"`
	UserAgent   string `json:"userAgent,omitempty"`
	Mode        string `json:"mode,omitempty"`
	SubmittedAt string `json:"submittedAt,omitempty"`
}

func (r submitFeedbackRequest) Validate() error {
	subject := strings.TrimSpace(r.Subject)
	if subject == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "subject is required")
	}
	if len(subject) > maxFeedbackSubjectLength {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "subject exceeds maximum length")
	}

	message := strings.TrimSpace(r.Message)
	if message == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "message is required")
	}
	if len(message) > maxFeedbackMessageLength {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "message exceeds maximum length")
	}

	if len(strings.TrimSpace(r.Context.ProjectID)) > maxFeedbackProjectIDLength {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "context.projectId exceeds maximum length")
	}
	if len(strings.TrimSpace(r.Context.Path)) > maxFeedbackPathLength {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "context.path exceeds maximum length")
	}
	if len(strings.TrimSpace(r.Context.UserAgent)) > maxFeedbackUserAgentLength {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "context.userAgent exceeds maximum length")
	}
	if len(strings.TrimSpace(r.Context.Mode)) > maxFeedbackModeLength {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "context.mode exceeds maximum length")
	}
	if len(strings.TrimSpace(r.Context.SubmittedAt)) > maxFeedbackSubmittedAtLength {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "context.submittedAt exceeds maximum length")
	}

	return nil
}

type submitFeedbackResponse struct {
	FeedbackID  string `json:"feedbackId"`
	Status      string `json:"status"`
	SubmittedAt string `json:"submittedAt"`
}

type createFeedbackSubmissionInput struct {
	UserID            string
	ProjectID         string
	Subject           string
	Message           string
	PagePath          string
	UserAgent         string
	Mode              string
	ClientSubmittedAt string
	RequestID         string
}

type feedbackSubmissionRecord struct {
	ID          string
	UserID      string
	UserName    string
	UserEmail   string
	ProjectID   string
	Subject     string
	Message     string
	PagePath    string
	ContextRaw  string
	SubmittedAt time.Time
}
