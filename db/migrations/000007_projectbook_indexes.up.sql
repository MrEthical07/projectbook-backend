CREATE INDEX IF NOT EXISTS users_created_at_idx ON users (created_at);

CREATE INDEX IF NOT EXISTS auth_sessions_user_expires_idx ON auth_sessions (user_id, expires_at);
CREATE INDEX IF NOT EXISTS auth_sessions_active_idx ON auth_sessions (user_id, expires_at) WHERE revoked_at IS NULL;

CREATE INDEX IF NOT EXISTS email_verification_tokens_user_expires_idx ON email_verification_tokens (user_id, expires_at DESC);
CREATE INDEX IF NOT EXISTS password_reset_tokens_user_expires_idx ON password_reset_tokens (user_id, expires_at DESC);
CREATE INDEX IF NOT EXISTS auth_email_log_user_sent_idx ON auth_email_log (user_id, sent_at DESC);
CREATE INDEX IF NOT EXISTS auth_email_log_recipient_sent_idx ON auth_email_log (recipient_email, sent_at DESC);

CREATE INDEX IF NOT EXISTS projects_owner_user_id_idx ON projects (owner_user_id);
CREATE INDEX IF NOT EXISTS projects_status_idx ON projects (status);
CREATE INDEX IF NOT EXISTS projects_last_updated_at_idx ON projects (last_updated_at DESC);

CREATE INDEX IF NOT EXISTS project_members_project_role_idx ON project_members (project_id, role);
CREATE INDEX IF NOT EXISTS project_members_user_id_idx ON project_members (user_id);
CREATE INDEX IF NOT EXISTS project_members_project_custom_idx ON project_members (project_id, is_custom);

CREATE UNIQUE INDEX IF NOT EXISTS project_invites_pending_email_unique_idx
ON project_invites (project_id, email)
WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS project_invites_project_status_idx ON project_invites (project_id, status);
CREATE INDEX IF NOT EXISTS project_invites_expires_at_idx ON project_invites (expires_at);

CREATE INDEX IF NOT EXISTS role_permissions_project_mask_idx ON role_permissions (project_id, permission_mask);

CREATE INDEX IF NOT EXISTS stories_project_updated_idx ON stories (project_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS journeys_project_updated_idx ON journeys (project_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS problems_project_updated_idx ON problems (project_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS ideas_project_updated_idx ON ideas (project_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS tasks_project_updated_idx ON tasks (project_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS feedback_project_updated_idx ON feedback (project_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS resources_project_updated_idx ON resources (project_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS pages_project_updated_idx ON pages (project_id, updated_at DESC);

CREATE INDEX IF NOT EXISTS resource_versions_resource_created_idx ON resource_versions (resource_id, created_at DESC);

CREATE INDEX IF NOT EXISTS calendar_events_project_starts_idx ON calendar_events (project_id, starts_at);

CREATE INDEX IF NOT EXISTS artifact_links_source_idx ON artifact_links (project_id, source_type, source_id);
CREATE INDEX IF NOT EXISTS artifact_links_target_idx ON artifact_links (project_id, target_type, target_id);

CREATE INDEX IF NOT EXISTS activity_log_project_created_idx ON activity_log (project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS activity_log_actor_created_idx ON activity_log (actor_user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS notifications_user_read_created_idx ON notifications (user_id, is_read, created_at DESC);
CREATE INDEX IF NOT EXISTS notifications_project_created_idx ON notifications (project_id, created_at DESC);

CREATE INDEX IF NOT EXISTS document_sync_outbox_status_next_attempt_idx ON document_sync_outbox (status, next_attempt_at);
CREATE INDEX IF NOT EXISTS document_sync_outbox_project_artifact_idx ON document_sync_outbox (project_id, artifact_type, artifact_id);
