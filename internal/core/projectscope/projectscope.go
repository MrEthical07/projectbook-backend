package projectscope

import (
	"context"
	"net/http"
	"strings"

	"github.com/MrEthical07/superapi/internal/core/auth"
	apperr "github.com/MrEthical07/superapi/internal/core/errors"
)

// ProjectIDFromContext extracts normalized project id from auth context.
func ProjectIDFromContext(ctx context.Context) (string, bool) {
	principal, ok := auth.FromContext(ctx)
	if !ok {
		return "", false
	}
	projectID := strings.TrimSpace(principal.ProjectID)
	if projectID == "" {
		return "", false
	}
	return projectID, true
}

// RequireProject returns forbidden error when request has no project scope.
func RequireProject(ctx context.Context) error {
	if _, ok := ProjectIDFromContext(ctx); ok {
		return nil
	}
	return apperr.New(apperr.CodeForbidden, http.StatusForbidden, "project scope required")
}

// IsSameProject compares principal and resource project identifiers.
func IsSameProject(principalProjectID, resourceProjectID string) bool {
	principal := strings.TrimSpace(principalProjectID)
	resource := strings.TrimSpace(resourceProjectID)
	return principal != "" && principal == resource
}
