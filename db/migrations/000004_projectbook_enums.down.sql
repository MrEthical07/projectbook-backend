-- Reverses 000004. Drops the enum types it created. Safe only after every table
-- and function that uses these types has been dropped by later downs (000005 and
-- 000006 downs drop those tables), which holds when rolling the whole chain back.
--
-- The pgcrypto and citext extensions created by the up are intentionally left in
-- place: they are database-wide, harmless to keep, and may be relied on elsewhere.
-- feedback_status is NOT dropped here — it is created in 000013 and dropped by
-- that migration's down.
DROP TYPE IF EXISTS notification_source_type;
DROP TYPE IF EXISTS artifact_type;
DROP TYPE IF EXISTS member_status;
DROP TYPE IF EXISTS invite_status;
DROP TYPE IF EXISTS calendar_artifact_type;
DROP TYPE IF EXISTS calendar_phase;
DROP TYPE IF EXISTS calendar_event_type;
DROP TYPE IF EXISTS page_status;
DROP TYPE IF EXISTS resource_status;
DROP TYPE IF EXISTS feedback_outcome;
DROP TYPE IF EXISTS task_status;
DROP TYPE IF EXISTS idea_status;
DROP TYPE IF EXISTS problem_status;
DROP TYPE IF EXISTS journey_status;
DROP TYPE IF EXISTS story_status;
DROP TYPE IF EXISTS project_role;
DROP TYPE IF EXISTS project_status;
