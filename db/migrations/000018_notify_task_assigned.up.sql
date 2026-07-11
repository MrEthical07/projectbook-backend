-- Add a per-project toggle for task-assignment notifications, matching the
-- existing notify_* flags on project_settings. Defaults TRUE so existing
-- projects keep assignment notifications on.
ALTER TABLE project_settings
    ADD COLUMN IF NOT EXISTS notify_task_assigned BOOLEAN NOT NULL DEFAULT TRUE;
