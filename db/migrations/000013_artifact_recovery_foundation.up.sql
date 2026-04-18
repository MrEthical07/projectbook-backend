DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'feedback_status') THEN
        CREATE TYPE feedback_status AS ENUM ('Active', 'Archived');
    END IF;
END
$$;

ALTER TABLE feedback
ADD COLUMN IF NOT EXISTS status feedback_status NOT NULL DEFAULT 'Active';

CREATE INDEX IF NOT EXISTS feedback_project_status_idx ON feedback (project_id, status);

ALTER TABLE stories
ADD COLUMN IF NOT EXISTS archived_from_status story_status NULL;

ALTER TABLE journeys
ADD COLUMN IF NOT EXISTS archived_from_status journey_status NULL;

ALTER TABLE problems
ADD COLUMN IF NOT EXISTS archived_from_status problem_status NULL;

ALTER TABLE ideas
ADD COLUMN IF NOT EXISTS archived_from_status idea_status NULL;

ALTER TABLE pages
ADD COLUMN IF NOT EXISTS archived_from_status page_status NULL;

ALTER TABLE resources
ADD COLUMN IF NOT EXISTS archived_from_status resource_status NULL;

ALTER TABLE feedback
ADD COLUMN IF NOT EXISTS archived_from_status feedback_status NULL;

CREATE TABLE IF NOT EXISTS task_assignees (
    task_id UUID NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    assigned_by_user_id UUID NULL REFERENCES users (id) ON DELETE SET NULL,
    PRIMARY KEY (task_id, user_id)
);

CREATE INDEX IF NOT EXISTS task_assignees_project_task_idx ON task_assignees (project_id, task_id);
CREATE INDEX IF NOT EXISTS task_assignees_project_user_idx ON task_assignees (project_id, user_id);

INSERT INTO task_assignees (task_id, project_id, user_id, assigned_by_user_id)
SELECT t.id, t.project_id, t.owner_user_id, t.owner_user_id
FROM tasks t
WHERE t.owner_user_id IS NOT NULL
ON CONFLICT (task_id, user_id) DO NOTHING;
