DROP INDEX IF EXISTS document_sync_outbox_project_artifact_idx;
DROP INDEX IF EXISTS document_sync_outbox_status_next_attempt_idx;

DROP INDEX IF EXISTS notifications_project_created_idx;
DROP INDEX IF EXISTS notifications_user_read_created_idx;

DROP INDEX IF EXISTS activity_log_actor_created_idx;
DROP INDEX IF EXISTS activity_log_project_created_idx;

DROP INDEX IF EXISTS artifact_links_target_idx;
DROP INDEX IF EXISTS artifact_links_source_idx;

DROP INDEX IF EXISTS calendar_events_project_starts_idx;

DROP INDEX IF EXISTS resource_versions_resource_created_idx;

DROP INDEX IF EXISTS pages_project_updated_idx;
DROP INDEX IF EXISTS resources_project_updated_idx;
DROP INDEX IF EXISTS feedback_project_updated_idx;
DROP INDEX IF EXISTS tasks_project_updated_idx;
DROP INDEX IF EXISTS ideas_project_updated_idx;
DROP INDEX IF EXISTS problems_project_updated_idx;
DROP INDEX IF EXISTS journeys_project_updated_idx;
DROP INDEX IF EXISTS stories_project_updated_idx;

DROP INDEX IF EXISTS role_permissions_project_mask_idx;

DROP INDEX IF EXISTS project_invites_expires_at_idx;
DROP INDEX IF EXISTS project_invites_project_status_idx;
DROP INDEX IF EXISTS project_invites_pending_email_unique_idx;

DROP INDEX IF EXISTS project_members_project_custom_idx;
DROP INDEX IF EXISTS project_members_user_id_idx;
DROP INDEX IF EXISTS project_members_project_role_idx;

DROP INDEX IF EXISTS projects_last_updated_at_idx;
DROP INDEX IF EXISTS projects_status_idx;
DROP INDEX IF EXISTS projects_owner_user_id_idx;

DROP INDEX IF EXISTS auth_sessions_active_idx;
DROP INDEX IF EXISTS auth_sessions_user_expires_idx;
DROP INDEX IF EXISTS auth_email_log_recipient_sent_idx;
DROP INDEX IF EXISTS auth_email_log_user_sent_idx;
DROP INDEX IF EXISTS password_reset_tokens_user_expires_idx;
DROP INDEX IF EXISTS email_verification_tokens_user_expires_idx;

DROP INDEX IF EXISTS users_created_at_idx;
