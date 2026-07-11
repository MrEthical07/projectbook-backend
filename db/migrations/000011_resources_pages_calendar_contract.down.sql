-- Reverses 000011. Drops the constraint, indexes and columns it added to
-- resources, pages and calendar_events. Column data (file_type, doc_type,
-- calendar detail fields, ...) is discarded with the columns, restoring the exact
-- pre-000011 shape.
ALTER TABLE activity_log DROP CONSTRAINT IF EXISTS activity_log_action_not_blank_ck;

DROP INDEX IF EXISTS resources_project_doc_type_idx;
DROP INDEX IF EXISTS resources_project_status_doc_type_idx;
DROP INDEX IF EXISTS pages_project_orphan_idx;
DROP INDEX IF EXISTS calendar_events_project_updated_idx;
DROP INDEX IF EXISTS calendar_events_project_event_type_idx;

ALTER TABLE resources
    DROP COLUMN IF EXISTS file_type,
    DROP COLUMN IF EXISTS doc_type;

ALTER TABLE pages
    DROP COLUMN IF EXISTS is_orphan;

ALTER TABLE calendar_events
    DROP COLUMN IF EXISTS all_day,
    DROP COLUMN IF EXISTS start_time,
    DROP COLUMN IF EXISTS end_time,
    DROP COLUMN IF EXISTS location,
    DROP COLUMN IF EXISTS event_kind,
    DROP COLUMN IF EXISTS linked_artifacts,
    DROP COLUMN IF EXISTS tags,
    DROP COLUMN IF EXISTS source_title;
