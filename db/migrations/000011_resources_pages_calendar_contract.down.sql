DO $$
BEGIN
	IF EXISTS (
		SELECT 1
		FROM pg_constraint
		WHERE conname = 'activity_log_action_not_blank_ck'
	) THEN
		ALTER TABLE activity_log
		DROP CONSTRAINT activity_log_action_not_blank_ck;
	END IF;
END
$$;

DROP INDEX IF EXISTS calendar_events_project_event_type_idx;
DROP INDEX IF EXISTS calendar_events_project_updated_idx;

ALTER TABLE calendar_events
DROP COLUMN IF EXISTS source_title,
DROP COLUMN IF EXISTS tags,
DROP COLUMN IF EXISTS linked_artifacts,
DROP COLUMN IF EXISTS event_kind,
DROP COLUMN IF EXISTS location,
DROP COLUMN IF EXISTS end_time,
DROP COLUMN IF EXISTS start_time,
DROP COLUMN IF EXISTS all_day;

DROP INDEX IF EXISTS pages_project_orphan_idx;

ALTER TABLE pages
DROP COLUMN IF EXISTS is_orphan;

DROP INDEX IF EXISTS resources_project_status_doc_type_idx;
DROP INDEX IF EXISTS resources_project_doc_type_idx;

ALTER TABLE resources
DROP COLUMN IF EXISTS doc_type,
DROP COLUMN IF EXISTS file_type;
