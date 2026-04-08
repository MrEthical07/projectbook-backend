CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'project_status') THEN
        CREATE TYPE project_status AS ENUM ('Active', 'Archived');
    END IF;
END
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'project_role') THEN
        CREATE TYPE project_role AS ENUM ('Owner', 'Admin', 'Editor', 'Member', 'Viewer', 'Limited Access');
    END IF;
END
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'story_status') THEN
        CREATE TYPE story_status AS ENUM ('Draft', 'Locked', 'Archived');
    END IF;
END
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'journey_status') THEN
        CREATE TYPE journey_status AS ENUM ('Draft', 'Archived');
    END IF;
END
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'problem_status') THEN
        CREATE TYPE problem_status AS ENUM ('Draft', 'Locked', 'Archived');
    END IF;
END
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'idea_status') THEN
        CREATE TYPE idea_status AS ENUM ('Considered', 'Selected', 'Rejected', 'Archived');
    END IF;
END
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'task_status') THEN
        CREATE TYPE task_status AS ENUM ('Planned', 'In Progress', 'Completed', 'Abandoned', 'Blocked');
    END IF;
END
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'feedback_outcome') THEN
        CREATE TYPE feedback_outcome AS ENUM ('Validated', 'Invalidated', 'Needs Iteration');
    END IF;
END
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'resource_status') THEN
        CREATE TYPE resource_status AS ENUM ('Active', 'Archived');
    END IF;
END
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'page_status') THEN
        CREATE TYPE page_status AS ENUM ('Draft', 'Archived');
    END IF;
END
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'calendar_event_type') THEN
        CREATE TYPE calendar_event_type AS ENUM ('Derived', 'Manual');
    END IF;
END
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'calendar_phase') THEN
        CREATE TYPE calendar_phase AS ENUM ('Empathize', 'Define', 'Ideate', 'Prototype', 'Test', 'None');
    END IF;
END
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'calendar_artifact_type') THEN
        CREATE TYPE calendar_artifact_type AS ENUM ('Task', 'Feedback', 'Manual');
    END IF;
END
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'invite_status') THEN
        CREATE TYPE invite_status AS ENUM ('pending', 'accepted', 'declined', 'expired', 'cancelled');
    END IF;
END
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'member_status') THEN
        CREATE TYPE member_status AS ENUM ('Active', 'Invited');
    END IF;
END
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'artifact_type') THEN
        CREATE TYPE artifact_type AS ENUM ('story', 'journey', 'problem', 'idea', 'task', 'feedback', 'resource', 'page', 'calendar');
    END IF;
END
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'notification_source_type') THEN
        CREATE TYPE notification_source_type AS ENUM ('Project Activity', 'Project Invitation', 'System Notification');
    END IF;
END
$$;
