CREATE TABLE IF NOT EXISTS search_index (
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    artifact_type TEXT NOT NULL,
    artifact_id UUID NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT '',
    href TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    search_vector TSVECTOR NOT NULL,
    PRIMARY KEY (project_id, artifact_type, artifact_id)
);

CREATE INDEX IF NOT EXISTS search_index_project_updated_idx
    ON search_index (project_id, updated_at DESC);

CREATE INDEX IF NOT EXISTS search_index_vector_idx
    ON search_index USING GIN (search_vector);

CREATE OR REPLACE FUNCTION search_index_build_vector(search_title TEXT, search_description TEXT)
RETURNS TSVECTOR
LANGUAGE SQL
IMMUTABLE
AS $$
SELECT
    setweight(to_tsvector('english', COALESCE(search_title, '')), 'A') ||
    setweight(to_tsvector('english', COALESCE(search_description, '')), 'B')
$$;

CREATE OR REPLACE FUNCTION search_index_upsert(
    search_project_id UUID,
    search_artifact_type TEXT,
    search_artifact_id UUID,
    search_title TEXT,
    search_description TEXT,
    search_status TEXT,
    search_href TEXT,
    search_updated_at TIMESTAMPTZ
)
RETURNS VOID
LANGUAGE SQL
AS $$
INSERT INTO search_index (
    project_id,
    artifact_type,
    artifact_id,
    title,
    description,
    status,
    href,
    updated_at,
    search_vector
)
VALUES (
    search_project_id,
    search_artifact_type,
    search_artifact_id,
    COALESCE(search_title, ''),
    COALESCE(search_description, ''),
    COALESCE(search_status, ''),
    COALESCE(search_href, ''),
    COALESCE(search_updated_at, NOW()),
    search_index_build_vector(search_title, search_description)
)
ON CONFLICT (project_id, artifact_type, artifact_id) DO UPDATE
SET
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    status = EXCLUDED.status,
    href = EXCLUDED.href,
    updated_at = EXCLUDED.updated_at,
    search_vector = EXCLUDED.search_vector
$$;

CREATE OR REPLACE FUNCTION search_index_delete(
    search_project_id UUID,
    search_artifact_type TEXT,
    search_artifact_id UUID
)
RETURNS VOID
LANGUAGE SQL
AS $$
DELETE FROM search_index
WHERE project_id = search_project_id
  AND artifact_type = search_artifact_type
  AND artifact_id = search_artifact_id
$$;

CREATE OR REPLACE FUNCTION search_index_sync_stories()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        PERFORM search_index_delete(OLD.project_id, 'story', OLD.id);
        RETURN OLD;
    END IF;

    PERFORM search_index_upsert(
        NEW.project_id,
        'story',
        NEW.id,
        NEW.title,
        '',
        NEW.status::TEXT,
        '/project/' || NEW.project_id::TEXT || '/stories/' || NEW.id::TEXT,
        NEW.updated_at
    );
    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION search_index_sync_journeys()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        PERFORM search_index_delete(OLD.project_id, 'journey', OLD.id);
        RETURN OLD;
    END IF;

    PERFORM search_index_upsert(
        NEW.project_id,
        'journey',
        NEW.id,
        NEW.title,
        '',
        NEW.status::TEXT,
        '/project/' || NEW.project_id::TEXT || '/journeys/' || NEW.id::TEXT,
        NEW.updated_at
    );
    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION search_index_sync_problems()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        PERFORM search_index_delete(OLD.project_id, 'problem', OLD.id);
        RETURN OLD;
    END IF;

    PERFORM search_index_upsert(
        NEW.project_id,
        'problem',
        NEW.id,
        NEW.title,
        '',
        NEW.status::TEXT,
        '/project/' || NEW.project_id::TEXT || '/problem-statement/' || NEW.id::TEXT,
        NEW.updated_at
    );
    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION search_index_sync_ideas()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        PERFORM search_index_delete(OLD.project_id, 'idea', OLD.id);
        RETURN OLD;
    END IF;

    PERFORM search_index_upsert(
        NEW.project_id,
        'idea',
        NEW.id,
        NEW.title,
        '',
        NEW.status::TEXT,
        '/project/' || NEW.project_id::TEXT || '/ideas/' || NEW.id::TEXT,
        NEW.updated_at
    );
    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION search_index_sync_tasks()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        PERFORM search_index_delete(OLD.project_id, 'task', OLD.id);
        RETURN OLD;
    END IF;

    PERFORM search_index_upsert(
        NEW.project_id,
        'task',
        NEW.id,
        NEW.title,
        '',
        NEW.status::TEXT,
        '/project/' || NEW.project_id::TEXT || '/tasks/' || NEW.id::TEXT,
        NEW.updated_at
    );
    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION search_index_sync_feedback()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        PERFORM search_index_delete(OLD.project_id, 'feedback', OLD.id);
        RETURN OLD;
    END IF;

    PERFORM search_index_upsert(
        NEW.project_id,
        'feedback',
        NEW.id,
        NEW.title,
        '',
        NEW.status::TEXT,
        '/project/' || NEW.project_id::TEXT || '/feedback/' || NEW.id::TEXT,
        NEW.updated_at
    );
    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION search_index_sync_resources()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        PERFORM search_index_delete(OLD.project_id, 'resource', OLD.id);
        RETURN OLD;
    END IF;

    PERFORM search_index_upsert(
        NEW.project_id,
        'resource',
        NEW.id,
        NEW.title,
        '',
        NEW.status::TEXT,
        '/project/' || NEW.project_id::TEXT || '/resources/' || NEW.id::TEXT,
        NEW.updated_at
    );
    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION search_index_sync_pages()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        PERFORM search_index_delete(OLD.project_id, 'page', OLD.id);
        RETURN OLD;
    END IF;

    PERFORM search_index_upsert(
        NEW.project_id,
        'page',
        NEW.id,
        NEW.title,
        '',
        NEW.status::TEXT,
        '/project/' || NEW.project_id::TEXT || '/pages/' || NEW.id::TEXT,
        NEW.updated_at
    );
    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION search_index_sync_calendar_events()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        PERFORM search_index_delete(OLD.project_id, 'calendar', OLD.id);
        RETURN OLD;
    END IF;

    PERFORM search_index_upsert(
        NEW.project_id,
        'calendar',
        NEW.id,
        NEW.title,
        COALESCE(NEW.description, ''),
        '',
        '/project/' || NEW.project_id::TEXT || '/calendar/' || NEW.id::TEXT,
        NEW.updated_at
    );
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_search_index_stories ON stories;
CREATE TRIGGER trg_search_index_stories
AFTER INSERT OR UPDATE OR DELETE ON stories
FOR EACH ROW
EXECUTE FUNCTION search_index_sync_stories();

DROP TRIGGER IF EXISTS trg_search_index_journeys ON journeys;
CREATE TRIGGER trg_search_index_journeys
AFTER INSERT OR UPDATE OR DELETE ON journeys
FOR EACH ROW
EXECUTE FUNCTION search_index_sync_journeys();

DROP TRIGGER IF EXISTS trg_search_index_problems ON problems;
CREATE TRIGGER trg_search_index_problems
AFTER INSERT OR UPDATE OR DELETE ON problems
FOR EACH ROW
EXECUTE FUNCTION search_index_sync_problems();

DROP TRIGGER IF EXISTS trg_search_index_ideas ON ideas;
CREATE TRIGGER trg_search_index_ideas
AFTER INSERT OR UPDATE OR DELETE ON ideas
FOR EACH ROW
EXECUTE FUNCTION search_index_sync_ideas();

DROP TRIGGER IF EXISTS trg_search_index_tasks ON tasks;
CREATE TRIGGER trg_search_index_tasks
AFTER INSERT OR UPDATE OR DELETE ON tasks
FOR EACH ROW
EXECUTE FUNCTION search_index_sync_tasks();

DROP TRIGGER IF EXISTS trg_search_index_feedback ON feedback;
CREATE TRIGGER trg_search_index_feedback
AFTER INSERT OR UPDATE OR DELETE ON feedback
FOR EACH ROW
EXECUTE FUNCTION search_index_sync_feedback();

DROP TRIGGER IF EXISTS trg_search_index_resources ON resources;
CREATE TRIGGER trg_search_index_resources
AFTER INSERT OR UPDATE OR DELETE ON resources
FOR EACH ROW
EXECUTE FUNCTION search_index_sync_resources();

DROP TRIGGER IF EXISTS trg_search_index_pages ON pages;
CREATE TRIGGER trg_search_index_pages
AFTER INSERT OR UPDATE OR DELETE ON pages
FOR EACH ROW
EXECUTE FUNCTION search_index_sync_pages();

DROP TRIGGER IF EXISTS trg_search_index_calendar_events ON calendar_events;
CREATE TRIGGER trg_search_index_calendar_events
AFTER INSERT OR UPDATE OR DELETE ON calendar_events
FOR EACH ROW
EXECUTE FUNCTION search_index_sync_calendar_events();

DELETE FROM search_index;

INSERT INTO search_index (
    project_id,
    artifact_type,
    artifact_id,
    title,
    description,
    status,
    href,
    updated_at,
    search_vector
)
SELECT
    s.project_id,
    'story',
    s.id,
    s.title,
    '',
    s.status::TEXT,
    '/project/' || s.project_id::TEXT || '/stories/' || s.id::TEXT,
    s.updated_at,
    search_index_build_vector(s.title, '')
FROM stories s
UNION ALL
SELECT
    j.project_id,
    'journey',
    j.id,
    j.title,
    '',
    j.status::TEXT,
    '/project/' || j.project_id::TEXT || '/journeys/' || j.id::TEXT,
    j.updated_at,
    search_index_build_vector(j.title, '')
FROM journeys j
UNION ALL
SELECT
    p.project_id,
    'problem',
    p.id,
    p.title,
    '',
    p.status::TEXT,
    '/project/' || p.project_id::TEXT || '/problem-statement/' || p.id::TEXT,
    p.updated_at,
    search_index_build_vector(p.title, '')
FROM problems p
UNION ALL
SELECT
    i.project_id,
    'idea',
    i.id,
    i.title,
    '',
    i.status::TEXT,
    '/project/' || i.project_id::TEXT || '/ideas/' || i.id::TEXT,
    i.updated_at,
    search_index_build_vector(i.title, '')
FROM ideas i
UNION ALL
SELECT
    t.project_id,
    'task',
    t.id,
    t.title,
    '',
    t.status::TEXT,
    '/project/' || t.project_id::TEXT || '/tasks/' || t.id::TEXT,
    t.updated_at,
    search_index_build_vector(t.title, '')
FROM tasks t
UNION ALL
SELECT
    f.project_id,
    'feedback',
    f.id,
    f.title,
    '',
    f.status::TEXT,
    '/project/' || f.project_id::TEXT || '/feedback/' || f.id::TEXT,
    f.updated_at,
    search_index_build_vector(f.title, '')
FROM feedback f
UNION ALL
SELECT
    r.project_id,
    'resource',
    r.id,
    r.title,
    '',
    r.status::TEXT,
    '/project/' || r.project_id::TEXT || '/resources/' || r.id::TEXT,
    r.updated_at,
    search_index_build_vector(r.title, '')
FROM resources r
UNION ALL
SELECT
    pg.project_id,
    'page',
    pg.id,
    pg.title,
    '',
    pg.status::TEXT,
    '/project/' || pg.project_id::TEXT || '/pages/' || pg.id::TEXT,
    pg.updated_at,
    search_index_build_vector(pg.title, '')
FROM pages pg
UNION ALL
SELECT
    e.project_id,
    'calendar',
    e.id,
    e.title,
    COALESCE(e.description, ''),
    '',
    '/project/' || e.project_id::TEXT || '/calendar/' || e.id::TEXT,
    e.updated_at,
    search_index_build_vector(e.title, COALESCE(e.description, ''))
FROM calendar_events e;

CREATE TABLE IF NOT EXISTS global_feedback_submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    project_id UUID NULL REFERENCES projects (id) ON DELETE SET NULL,
    subject TEXT NOT NULL,
    message TEXT NOT NULL,
    page_path TEXT NOT NULL DEFAULT '',
    context JSONB NOT NULL DEFAULT '{}'::JSONB,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    email_queued_at TIMESTAMPTZ NULL,
    email_sent_at TIMESTAMPTZ NULL,
    email_error TEXT NULL,
    CHECK (length(trim(subject)) > 0),
    CHECK (length(trim(message)) > 0)
);

CREATE INDEX IF NOT EXISTS global_feedback_submissions_user_submitted_idx
    ON global_feedback_submissions (user_id, submitted_at DESC);

CREATE INDEX IF NOT EXISTS global_feedback_submissions_project_submitted_idx
    ON global_feedback_submissions (project_id, submitted_at DESC);