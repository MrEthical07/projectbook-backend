-- Reverses 000018. Drops the notify_task_assigned column; the flag values are
-- discarded with it, restoring the exact pre-000018 project_settings shape.
ALTER TABLE project_settings
    DROP COLUMN IF EXISTS notify_task_assigned;
