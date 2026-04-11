DROP TRIGGER IF EXISTS feedback_sync_chain_orphan_trg ON feedback;
DROP TRIGGER IF EXISTS tasks_sync_chain_orphan_trg ON tasks;
DROP TRIGGER IF EXISTS ideas_sync_chain_orphan_trg ON ideas;

DROP TRIGGER IF EXISTS problems_sync_orphan_link_trg ON artifact_links;
DROP TRIGGER IF EXISTS problems_sync_orphan_row_trg ON problems;

DROP TRIGGER IF EXISTS feedback_cleanup_links_trg ON feedback;
DROP TRIGGER IF EXISTS tasks_cleanup_links_trg ON tasks;
DROP TRIGGER IF EXISTS ideas_cleanup_links_trg ON ideas;
DROP TRIGGER IF EXISTS problems_cleanup_links_trg ON problems;
DROP TRIGGER IF EXISTS journeys_cleanup_links_trg ON journeys;
DROP TRIGGER IF EXISTS stories_cleanup_links_trg ON stories;

DROP FUNCTION IF EXISTS pb_sync_chain_orphan_flags();
DROP FUNCTION IF EXISTS pb_sync_problem_orphan_on_link_change();
DROP FUNCTION IF EXISTS pb_sync_problem_orphan_on_row();
DROP FUNCTION IF EXISTS pb_sync_problem_orphan(UUID);
DROP FUNCTION IF EXISTS pb_cleanup_artifact_links();

DROP INDEX IF EXISTS feedback_project_orphan_idx;
DROP INDEX IF EXISTS tasks_project_orphan_idx;
DROP INDEX IF EXISTS ideas_project_orphan_idx;
DROP INDEX IF EXISTS problems_project_orphan_idx;
DROP INDEX IF EXISTS journeys_project_orphan_idx;

DROP INDEX IF EXISTS feedback_project_primary_task_idx;
DROP INDEX IF EXISTS tasks_project_primary_idea_idx;
DROP INDEX IF EXISTS ideas_project_primary_problem_idx;

ALTER TABLE feedback
DROP COLUMN IF EXISTS primary_task_id,
DROP COLUMN IF EXISTS is_orphan;

ALTER TABLE tasks
DROP COLUMN IF EXISTS primary_idea_id,
DROP COLUMN IF EXISTS is_orphan;

ALTER TABLE ideas
DROP COLUMN IF EXISTS primary_problem_id,
DROP COLUMN IF EXISTS is_orphan;

ALTER TABLE problems
DROP COLUMN IF EXISTS is_orphan;

ALTER TABLE journeys
DROP COLUMN IF EXISTS is_orphan;
