ALTER TABLE resources
ADD COLUMN IF NOT EXISTS file_type TEXT NOT NULL DEFAULT 'PDF',
ADD COLUMN IF NOT EXISTS doc_type TEXT NOT NULL DEFAULT 'Other';

CREATE INDEX IF NOT EXISTS resources_project_doc_type_idx ON resources (project_id, doc_type);
CREATE INDEX IF NOT EXISTS resources_project_status_doc_type_idx ON resources (project_id, status, doc_type);

ALTER TABLE pages
ADD COLUMN IF NOT EXISTS is_orphan BOOLEAN NOT NULL DEFAULT TRUE;

CREATE INDEX IF NOT EXISTS pages_project_orphan_idx ON pages (project_id, is_orphan);

ALTER TABLE calendar_events
ADD COLUMN IF NOT EXISTS all_day BOOLEAN NOT NULL DEFAULT TRUE,
ADD COLUMN IF NOT EXISTS start_time TEXT NULL,
ADD COLUMN IF NOT EXISTS end_time TEXT NULL,
ADD COLUMN IF NOT EXISTS location TEXT NULL,
ADD COLUMN IF NOT EXISTS event_kind TEXT NULL,
ADD COLUMN IF NOT EXISTS linked_artifacts JSONB NOT NULL DEFAULT '[]'::JSONB,
ADD COLUMN IF NOT EXISTS tags JSONB NOT NULL DEFAULT '[]'::JSONB,
ADD COLUMN IF NOT EXISTS source_title TEXT NULL;

CREATE INDEX IF NOT EXISTS calendar_events_project_updated_idx ON calendar_events (project_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS calendar_events_project_event_type_idx ON calendar_events (project_id, event_type);

DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1
		FROM pg_constraint
		WHERE conname = 'activity_log_action_not_blank_ck'
	) THEN
		ALTER TABLE activity_log
		ADD CONSTRAINT activity_log_action_not_blank_ck CHECK (length(trim(action)) > 0);
	END IF;
END
$$;
