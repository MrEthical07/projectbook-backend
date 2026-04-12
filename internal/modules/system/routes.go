package system

import (
	"net/http"
	"strings"
	"time"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
	"github.com/MrEthical07/superapi/internal/core/httpx"
	"github.com/MrEthical07/superapi/internal/core/policy"
	"github.com/MrEthical07/superapi/internal/core/ratelimit"
)

type parseDurationRequest struct {
	Duration string `json:"duration"`
}

// Validate ensures duration string is provided for parsing.
func (r parseDurationRequest) Validate() error {
	if strings.TrimSpace(r.Duration) == "" {
		return apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "duration is required")
	}
	return nil
}

type parseDurationResponse struct {
	Duration     string `json:"duration"`
	Nanoseconds  int64  `json:"nanoseconds"`
	Milliseconds int64  `json:"milliseconds"`
}

// Register mounts system and auth demonstration routes.
//
// For policy behavior, see docs/policies.md.
func (m *Module) Register(r httpx.Router) error {
	r.Handle(http.MethodPost, "/system/parse-duration", httpx.Adapter(m.parseDuration), policy.RequireJSON())

	if limiter := m.runtime.Limiter(); limiter != nil {
		r.Handle(
			http.MethodGet,
			"/api/v1/system/whoami",
			httpx.Adapter(m.whoami),
			policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
			policy.RateLimitWithKeyer(limiter, "system.whoami", m.rateRule, ratelimit.KeyByUserOrProjectOrTokenHash(16)),
		)
		return nil
	}

	r.Handle(
		http.MethodGet,
		"/api/v1/system/whoami",
		httpx.Adapter(m.whoami),
		policy.AuthRequired(m.runtime.AuthEngine(), m.runtime.AuthMode()),
	)
	return nil
}

type whoamiResponse struct {
	UserID         string   `json:"user_id"`
	ProjectID      string   `json:"project_id,omitempty"`
	Role           string   `json:"role,omitempty"`
	PermissionMask uint64   `json:"permission_mask,omitempty"`
	Permissions    []string `json:"permissions,omitempty"`
}

func (m *Module) whoami(ctx *httpx.Context, _ httpx.NoBody) (whoamiResponse, error) {
	principal, ok := ctx.Auth()
	if !ok {
		return whoamiResponse{}, apperr.New(apperr.CodeUnauthorized, http.StatusUnauthorized, "authentication required")
	}

	return whoamiResponse{
		UserID:         principal.UserID,
		ProjectID:      principal.ProjectID,
		Role:           principal.Role,
		PermissionMask: principal.PermissionMask,
		Permissions:    append([]string(nil), principal.Permissions...),
	}, nil
}

func (m *Module) parseDuration(_ *httpx.Context, req parseDurationRequest) (parseDurationResponse, error) {
	d, err := time.ParseDuration(req.Duration)
	if err != nil {
		return parseDurationResponse{}, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "duration must be a valid Go duration string")
	}

	return parseDurationResponse{
		Duration:     d.String(),
		Nanoseconds:  d.Nanoseconds(),
		Milliseconds: d.Milliseconds(),
	}, nil
}
