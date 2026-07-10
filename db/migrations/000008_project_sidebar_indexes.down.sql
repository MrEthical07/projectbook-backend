-- Reverses 000008. Drops the per-status sidebar indexes it created.
DROP INDEX IF EXISTS stories_project_status_idx;
DROP INDEX IF EXISTS journeys_project_status_idx;
DROP INDEX IF EXISTS problems_project_status_idx;
DROP INDEX IF EXISTS ideas_project_status_idx;
DROP INDEX IF EXISTS tasks_project_status_idx;
DROP INDEX IF EXISTS resources_project_status_idx;
DROP INDEX IF EXISTS pages_project_status_idx;
