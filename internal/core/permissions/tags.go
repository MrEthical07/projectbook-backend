package permissions

import (
	"context"
	"fmt"
	"strings"
)

// TagInvalidator is the minimal cache manager contract used for permission tag invalidation.
type TagInvalidator interface {
	BumpTags(ctx context.Context, tags []string) error
}

// PermissionTags returns predictable user-specific dynamic tags for permission cache invalidation.
func PermissionTags(userID, projectID string) []string {
	user := strings.TrimSpace(userID)
	project := strings.TrimSpace(projectID)
	if user == "" || project == "" {
		return nil
	}
	return []string{
		fmt.Sprintf("permissions:user=%s", user),
		fmt.Sprintf("permissions:project=%s", project),
		fmt.Sprintf("permissions:project_user=%s:%s", project, user),
	}
}

// InvalidatePermissionTags bumps dynamic permission tags for one user-project scope.
func InvalidatePermissionTags(ctx context.Context, invalidator TagInvalidator, userID, projectID string) error {
	if invalidator == nil {
		return fmt.Errorf("permission tag invalidator is nil")
	}
	tags := PermissionTags(userID, projectID)
	if len(tags) == 0 {
		return nil
	}
	return invalidator.BumpTags(ctx, tags)
}
