CREATE INDEX IF NOT EXISTS stories_project_status_idx ON stories (project_id, status);
CREATE INDEX IF NOT EXISTS journeys_project_status_idx ON journeys (project_id, status);
CREATE INDEX IF NOT EXISTS problems_project_status_idx ON problems (project_id, status);
CREATE INDEX IF NOT EXISTS ideas_project_status_idx ON ideas (project_id, status);
CREATE INDEX IF NOT EXISTS tasks_project_status_idx ON tasks (project_id, status);
CREATE INDEX IF NOT EXISTS resources_project_status_idx ON resources (project_id, status);
CREATE INDEX IF NOT EXISTS pages_project_status_idx ON pages (project_id, status);
