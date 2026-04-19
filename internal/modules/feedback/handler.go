package feedback

import (
	"net/http"
	"strings"

	"github.com/MrEthical07/superapi/internal/core/auth"
	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/httpx"
)

// Handler contains HTTP transport handlers for feedback routes.
type Handler struct {
	svc Service
}

// NewHandler constructs feedback transport handlers.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Submit(ctx *httpx.Context, req submitFeedbackRequest) (submitFeedbackResponse, error) {
	principal, err := requireAuthenticatedPrincipal(ctx)
	if err != nil {
		return submitFeedbackResponse{}, err
	}
	return h.svc.Submit(ctx.Context(), principal.UserID, req)
}

func requireAuthenticatedPrincipal(ctx *httpx.Context) (auth.AuthContext, error) {
	if ctx == nil {
		return auth.AuthContext{}, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	}
	principal, ok := ctx.Auth()
	if !ok || strings.TrimSpace(principal.UserID) == "" {
		return auth.AuthContext{}, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	}
	return principal, nil
}
