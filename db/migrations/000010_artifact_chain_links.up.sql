ALTER TABLE journeys
ADD COLUMN IF NOT EXISTS is_orphan BOOLEAN NOT NULL DEFAULT TRUE;

ALTER TABLE problems
ADD COLUMN IF NOT EXISTS is_orphan BOOLEAN NOT NULL DEFAULT TRUE;

ALTER TABLE ideas
ADD COLUMN IF NOT EXISTS primary_problem_id UUID NULL REFERENCES problems (id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS is_orphan BOOLEAN NOT NULL DEFAULT TRUE;

ALTER TABLE tasks
ADD COLUMN IF NOT EXISTS primary_idea_id UUID NULL REFERENCES ideas (id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS is_orphan BOOLEAN NOT NULL DEFAULT TRUE;

ALTER TABLE feedback
ADD COLUMN IF NOT EXISTS primary_task_id UUID NULL REFERENCES tasks (id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS is_orphan BOOLEAN NOT NULL DEFAULT TRUE;

CREATE INDEX IF NOT EXISTS ideas_project_primary_problem_idx ON ideas (project_id, primary_problem_id);
CREATE INDEX IF NOT EXISTS tasks_project_primary_idea_idx ON tasks (project_id, primary_idea_id);
CREATE INDEX IF NOT EXISTS feedback_project_primary_task_idx ON feedback (project_id, primary_task_id);

CREATE INDEX IF NOT EXISTS journeys_project_orphan_idx ON journeys (project_id, is_orphan);
CREATE INDEX IF NOT EXISTS problems_project_orphan_idx ON problems (project_id, is_orphan);
CREATE INDEX IF NOT EXISTS ideas_project_orphan_idx ON ideas (project_id, is_orphan);
CREATE INDEX IF NOT EXISTS tasks_project_orphan_idx ON tasks (project_id, is_orphan);
CREATE INDEX IF NOT EXISTS feedback_project_orphan_idx ON feedback (project_id, is_orphan);

CREATE OR REPLACE FUNCTION pb_cleanup_artifact_links()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    DELETE FROM artifact_links
    WHERE (source_type = TG_ARGV[0]::artifact_type AND source_id = OLD.id)
       OR (target_type = TG_ARGV[0]::artifact_type AND target_id = OLD.id);

    RETURN OLD;
END;
$$;

CREATE OR REPLACE FUNCTION pb_sync_problem_orphan(problem_uuid UUID)
RETURNS VOID
LANGUAGE plpgsql
AS $$
BEGIN
    UPDATE problems p
    SET is_orphan = NOT EXISTS (
        SELECT 1
        FROM artifact_links l
        WHERE l.project_id = p.project_id
          AND l.target_type = 'problem'::artifact_type
          AND l.target_id = p.id
          AND l.source_type IN ('story'::artifact_type, 'journey'::artifact_type)
    )
    WHERE p.id = problem_uuid;
END;
$$;

CREATE OR REPLACE FUNCTION pb_sync_problem_orphan_on_row()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.is_orphan := NOT EXISTS (
        SELECT 1
        FROM artifact_links l
        WHERE l.project_id = NEW.project_id
          AND l.target_type = 'problem'::artifact_type
          AND l.target_id = NEW.id
          AND l.source_type IN ('story'::artifact_type, 'journey'::artifact_type)
    );

    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION pb_sync_problem_orphan_on_link_change()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        IF OLD.target_type = 'problem'::artifact_type
            AND OLD.source_type IN ('story'::artifact_type, 'journey'::artifact_type) THEN
            PERFORM pb_sync_problem_orphan(OLD.target_id);
        END IF;
        RETURN OLD;
    END IF;

    IF NEW.target_type = 'problem'::artifact_type
        AND NEW.source_type IN ('story'::artifact_type, 'journey'::artifact_type) THEN
        PERFORM pb_sync_problem_orphan(NEW.target_id);
    END IF;

    IF TG_OP = 'UPDATE' THEN
        IF OLD.target_type = 'problem'::artifact_type
            AND OLD.source_type IN ('story'::artifact_type, 'journey'::artifact_type)
            AND (OLD.target_id <> NEW.target_id
                OR OLD.source_id <> NEW.source_id
                OR OLD.source_type <> NEW.source_type) THEN
            PERFORM pb_sync_problem_orphan(OLD.target_id);
        END IF;
    END IF;

    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION pb_sync_chain_orphan_flags()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_TABLE_NAME = 'ideas' THEN
        NEW.is_orphan := NEW.primary_problem_id IS NULL;
    ELSIF TG_TABLE_NAME = 'tasks' THEN
        NEW.is_orphan := NEW.primary_idea_id IS NULL;
    ELSIF TG_TABLE_NAME = 'feedback' THEN
        NEW.is_orphan := NEW.primary_task_id IS NULL;
    END IF;

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS stories_cleanup_links_trg ON stories;
CREATE TRIGGER stories_cleanup_links_trg
BEFORE DELETE ON stories
FOR EACH ROW EXECUTE FUNCTION pb_cleanup_artifact_links('story');

DROP TRIGGER IF EXISTS journeys_cleanup_links_trg ON journeys;
CREATE TRIGGER journeys_cleanup_links_trg
BEFORE DELETE ON journeys
FOR EACH ROW EXECUTE FUNCTION pb_cleanup_artifact_links('journey');

DROP TRIGGER IF EXISTS problems_cleanup_links_trg ON problems;
CREATE TRIGGER problems_cleanup_links_trg
BEFORE DELETE ON problems
FOR EACH ROW EXECUTE FUNCTION pb_cleanup_artifact_links('problem');

DROP TRIGGER IF EXISTS ideas_cleanup_links_trg ON ideas;
CREATE TRIGGER ideas_cleanup_links_trg
BEFORE DELETE ON ideas
FOR EACH ROW EXECUTE FUNCTION pb_cleanup_artifact_links('idea');

DROP TRIGGER IF EXISTS tasks_cleanup_links_trg ON tasks;
CREATE TRIGGER tasks_cleanup_links_trg
BEFORE DELETE ON tasks
FOR EACH ROW EXECUTE FUNCTION pb_cleanup_artifact_links('task');

DROP TRIGGER IF EXISTS feedback_cleanup_links_trg ON feedback;
CREATE TRIGGER feedback_cleanup_links_trg
BEFORE DELETE ON feedback
FOR EACH ROW EXECUTE FUNCTION pb_cleanup_artifact_links('feedback');

DROP TRIGGER IF EXISTS problems_sync_orphan_row_trg ON problems;
CREATE TRIGGER problems_sync_orphan_row_trg
BEFORE INSERT OR UPDATE OF project_id, id ON problems
FOR EACH ROW EXECUTE FUNCTION pb_sync_problem_orphan_on_row();

DROP TRIGGER IF EXISTS problems_sync_orphan_link_trg ON artifact_links;
CREATE TRIGGER problems_sync_orphan_link_trg
AFTER INSERT OR UPDATE OR DELETE ON artifact_links
FOR EACH ROW EXECUTE FUNCTION pb_sync_problem_orphan_on_link_change();

DROP TRIGGER IF EXISTS ideas_sync_chain_orphan_trg ON ideas;
CREATE TRIGGER ideas_sync_chain_orphan_trg
BEFORE INSERT OR UPDATE OF primary_problem_id ON ideas
FOR EACH ROW EXECUTE FUNCTION pb_sync_chain_orphan_flags();

DROP TRIGGER IF EXISTS tasks_sync_chain_orphan_trg ON tasks;
CREATE TRIGGER tasks_sync_chain_orphan_trg
BEFORE INSERT OR UPDATE OF primary_idea_id ON tasks
FOR EACH ROW EXECUTE FUNCTION pb_sync_chain_orphan_flags();

DROP TRIGGER IF EXISTS feedback_sync_chain_orphan_trg ON feedback;
CREATE TRIGGER feedback_sync_chain_orphan_trg
BEFORE INSERT OR UPDATE OF primary_task_id ON feedback
FOR EACH ROW EXECUTE FUNCTION pb_sync_chain_orphan_flags();

WITH ranked_problem_links AS (
    SELECT
        l.target_id,
        l.source_id,
        ROW_NUMBER() OVER (PARTITION BY l.target_id ORDER BY l.created_at ASC, l.id ASC) AS rn
    FROM artifact_links l
    WHERE l.source_type = 'problem'::artifact_type
      AND l.target_type = 'idea'::artifact_type
)
UPDATE ideas i
SET primary_problem_id = r.source_id
FROM ranked_problem_links r
WHERE i.id = r.target_id
  AND r.rn = 1
  AND i.primary_problem_id IS NULL;

WITH ranked_idea_links AS (
    SELECT
        l.target_id,
        l.source_id,
        ROW_NUMBER() OVER (PARTITION BY l.target_id ORDER BY l.created_at ASC, l.id ASC) AS rn
    FROM artifact_links l
    WHERE l.source_type = 'idea'::artifact_type
      AND l.target_type = 'task'::artifact_type
)
UPDATE tasks t
SET primary_idea_id = r.source_id
FROM ranked_idea_links r
WHERE t.id = r.target_id
  AND r.rn = 1
  AND t.primary_idea_id IS NULL;

WITH ranked_task_links AS (
    SELECT
        l.target_id,
        l.source_id,
        ROW_NUMBER() OVER (PARTITION BY l.target_id ORDER BY l.created_at ASC, l.id ASC) AS rn
    FROM artifact_links l
    WHERE l.source_type = 'task'::artifact_type
      AND l.target_type = 'feedback'::artifact_type
)
UPDATE feedback f
SET primary_task_id = r.source_id
FROM ranked_task_links r
WHERE f.id = r.target_id
  AND r.rn = 1
  AND f.primary_task_id IS NULL;

UPDATE ideas
SET is_orphan = (primary_problem_id IS NULL);

UPDATE tasks
SET is_orphan = (primary_idea_id IS NULL);

UPDATE feedback
SET is_orphan = (primary_task_id IS NULL);

UPDATE problems p
SET is_orphan = NOT EXISTS (
    SELECT 1
    FROM artifact_links l
    WHERE l.project_id = p.project_id
      AND l.target_type = 'problem'::artifact_type
      AND l.target_id = p.id
      AND l.source_type IN ('story'::artifact_type, 'journey'::artifact_type)
);

UPDATE journeys
SET is_orphan = TRUE
WHERE is_orphan IS NULL;
