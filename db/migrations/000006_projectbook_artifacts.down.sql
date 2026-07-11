-- Reverses 000006. Drops the artifact tables it created. By the time this runs
-- in a full rollback, the chain-link FK columns (000010), search triggers
-- (000015), task_assignees (000013) etc. have already been removed by their own
-- downs, so these tables have no remaining external dependents. Dropped
-- children-first (resource_versions before resources) so no CASCADE is needed.
DROP TABLE IF EXISTS document_sync_outbox;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS activity_log;
DROP TABLE IF EXISTS artifact_links;
DROP TABLE IF EXISTS calendar_events;
DROP TABLE IF EXISTS pages;
DROP TABLE IF EXISTS resource_versions;
DROP TABLE IF EXISTS resources;
DROP TABLE IF EXISTS feedback;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS ideas;
DROP TABLE IF EXISTS problems;
DROP TABLE IF EXISTS journeys;
DROP TABLE IF EXISTS stories;
