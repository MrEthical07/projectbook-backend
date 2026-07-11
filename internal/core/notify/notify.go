// Package notify fans project events out to the per-user notifications inbox.
//
// A single Event is expanded into notifications rows for the relevant users,
// applying two gating layers that already exist in the schema:
//
//   - project_settings.notify_*  — per-project toggle for a class of event
//     (passed as GateColumn; empty means "no project gate", e.g. invitations).
//   - account_settings.in_app_notifications — the recipient's personal master
//     switch (defaults TRUE when the user has no settings row yet).
//
// The actor (the user who caused the event) is never notified about their own
// action. Delivery is best-effort from the caller's perspective: callers should
// treat a Publish error as non-fatal so a notification failure never blocks the
// primary action that triggered it.
package notify

import (
	"context"
	"fmt"
	"strings"

	"github.com/MrEthical07/projectbook-backend/internal/core/storage"
)

// Notification source buckets (notification_source_type enum values).
const (
	SourceProjectActivity   = "Project Activity"
	SourceProjectInvitation = "Project Invitation"
	SourceSystem            = "System Notification"
)

// project_settings gate columns.
const (
	GateArtifactCreated = "notify_artifact_created"
	GateArtifactLocked  = "notify_artifact_locked"
	GateFeedbackAdded   = "notify_feedback_added"
	GateResourceUpdated = "notify_resource_updated"
	GateTaskAssigned    = "notify_task_assigned"
)

// allowedGates whitelists the column names that may be interpolated into SQL.
// GateColumn always originates from the constants above (never user input), but
// the whitelist keeps that guarantee explicit and local.
var allowedGates = map[string]bool{
	GateArtifactCreated: true,
	GateArtifactLocked:  true,
	GateFeedbackAdded:   true,
	GateResourceUpdated: true,
	GateTaskAssigned:    true,
}

// Fanout writes notifications derived from project events.
type Fanout struct {
	store storage.RelationalStore
}

// NewFanout builds a Fanout over the given relational store.
func NewFanout(store storage.RelationalStore) *Fanout {
	return &Fanout{store: store}
}

// Event describes something that happened in a project.
type Event struct {
	ProjectUUID string
	ActorUserID string
	SourceType  string // one of the Source* constants
	GateColumn  string // one of the Gate* constants, or "" for no project gate
	Title       string
	Message     string
	SourceID    string // originating artifact/invite id, or "" if none

	// Recipients, when non-empty, directs the notification at those users only
	// (e.g. a task assignee). When empty the event is broadcast to every active
	// project member except the actor.
	Recipients []string
}

// Publish expands e into notifications rows, applying the gating layers above.
func (f *Fanout) Publish(ctx context.Context, e Event) error {
	if f == nil || f.store == nil {
		return nil
	}
	if strings.TrimSpace(e.ProjectUUID) == "" || strings.TrimSpace(e.SourceType) == "" {
		return fmt.Errorf("notify: project and source type are required")
	}

	gate := "TRUE"
	if e.GateColumn != "" {
		if !allowedGates[e.GateColumn] {
			return fmt.Errorf("notify: unknown gate column %q", e.GateColumn)
		}
		gate = fmt.Sprintf(
			"COALESCE((SELECT ps.%s FROM project_settings ps WHERE ps.project_id = $1::uuid), TRUE)",
			e.GateColumn,
		)
	}
	actor := strings.TrimSpace(e.ActorUserID)

	if len(e.Recipients) == 0 {
		query := fmt.Sprintf(`
INSERT INTO notifications (user_id, project_id, source_type, source_id, title, message)
SELECT pm.user_id, $1::uuid, $2::notification_source_type, NULLIF($3, '')::uuid, $4, $5
FROM project_members pm
LEFT JOIN account_settings a ON a.user_id = pm.user_id
WHERE pm.project_id = $1::uuid
  AND pm.status = 'Active'::member_status
  AND ($6 = '' OR pm.user_id::text <> $6)
  AND COALESCE(a.in_app_notifications, TRUE)
  AND %s`, gate)
		return f.store.Execute(ctx, storage.RelationalExec(
			query,
			e.ProjectUUID, e.SourceType, e.SourceID, e.Title, e.Message, actor,
		))
	}

	// Directed delivery: one insert per recipient (recipient sets are small).
	query := fmt.Sprintf(`
INSERT INTO notifications (user_id, project_id, source_type, source_id, title, message)
SELECT $7::uuid, $1::uuid, $2::notification_source_type, NULLIF($3, '')::uuid, $4, $5
	WHERE ($6 = '' OR $7 <> NULLIF($6, '')::uuid)
  AND EXISTS (SELECT 1 FROM users u WHERE u.id = $7::uuid)
  AND COALESCE((SELECT a.in_app_notifications FROM account_settings a WHERE a.user_id = $7::uuid), TRUE)
  AND %s`, gate)

	seen := make(map[string]struct{}, len(e.Recipients))
	for _, raw := range e.Recipients {
		rcpt := strings.TrimSpace(raw)
		if rcpt == "" {
			continue
		}
		if _, dup := seen[rcpt]; dup {
			continue
		}
		seen[rcpt] = struct{}{}
		if err := f.store.Execute(ctx, storage.RelationalExec(
			query,
			e.ProjectUUID, e.SourceType, e.SourceID, e.Title, e.Message, actor, rcpt,
		)); err != nil {
			return err
		}
	}
	return nil
}
