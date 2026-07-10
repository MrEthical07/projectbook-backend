
CREATE SCHEMA public;

COMMENT ON SCHEMA public IS 'standard public schema';

CREATE TYPE public.artifact_type AS ENUM (
    'story',
    'journey',
    'problem',
    'idea',
    'task',
    'feedback',
    'resource',
    'page',
    'calendar'
);

CREATE TYPE public.calendar_artifact_type AS ENUM (
    'Task',
    'Feedback',
    'Manual'
);

CREATE TYPE public.calendar_event_type AS ENUM (
    'Derived',
    'Manual'
);

CREATE TYPE public.calendar_phase AS ENUM (
    'Empathize',
    'Define',
    'Ideate',
    'Prototype',
    'Test',
    'None'
);

CREATE TYPE public.feedback_outcome AS ENUM (
    'Validated',
    'Invalidated',
    'Needs Iteration'
);

CREATE TYPE public.feedback_status AS ENUM (
    'Active',
    'Archived'
);

CREATE TYPE public.idea_status AS ENUM (
    'Considered',
    'Selected',
    'Rejected',
    'Archived'
);

CREATE TYPE public.invite_status AS ENUM (
    'pending',
    'accepted',
    'declined',
    'expired',
    'cancelled'
);

CREATE TYPE public.journey_status AS ENUM (
    'Draft',
    'Archived',
    'Locked'
);

CREATE TYPE public.member_status AS ENUM (
    'Active',
    'Invited'
);

CREATE TYPE public.notification_source_type AS ENUM (
    'Project Activity',
    'Project Invitation',
    'System Notification'
);

CREATE TYPE public.page_status AS ENUM (
    'Draft',
    'Archived'
);

CREATE TYPE public.problem_status AS ENUM (
    'Draft',
    'Locked',
    'Archived'
);

CREATE TYPE public.project_role AS ENUM (
    'Owner',
    'Admin',
    'Editor',
    'Member',
    'Viewer',
    'Limited Access'
);

CREATE TYPE public.project_status AS ENUM (
    'Active',
    'Archived'
);

CREATE TYPE public.resource_status AS ENUM (
    'Active',
    'Archived'
);

CREATE TYPE public.story_status AS ENUM (
    'Draft',
    'Locked',
    'Archived'
);

CREATE TYPE public.task_status AS ENUM (
    'Planned',
    'In Progress',
    'Completed',
    'Abandoned',
    'Blocked'
);

CREATE FUNCTION public.pb_cleanup_artifact_links() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    DELETE FROM artifact_links
    WHERE (source_type = TG_ARGV[0]::artifact_type AND source_id = OLD.id)
       OR (target_type = TG_ARGV[0]::artifact_type AND target_id = OLD.id);

    RETURN OLD;
END;
$$;

CREATE FUNCTION public.pb_sync_chain_orphan_flags() RETURNS trigger
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

CREATE FUNCTION public.pb_sync_problem_orphan(problem_uuid uuid) RETURNS void
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

CREATE FUNCTION public.pb_sync_problem_orphan_on_link_change() RETURNS trigger
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

CREATE FUNCTION public.pb_sync_problem_orphan_on_row() RETURNS trigger
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

CREATE FUNCTION public.search_index_build_vector(search_title text, search_description text) RETURNS tsvector
    LANGUAGE sql IMMUTABLE
    AS $$
SELECT
    setweight(to_tsvector('english', COALESCE(search_title, '')), 'A') ||
    setweight(to_tsvector('english', COALESCE(search_description, '')), 'B')
$$;

CREATE FUNCTION public.search_index_delete(search_project_id uuid, search_artifact_type text, search_artifact_id uuid) RETURNS void
    LANGUAGE sql
    AS $$
DELETE FROM search_index
WHERE project_id = search_project_id
  AND artifact_type = search_artifact_type
  AND artifact_id = search_artifact_id
$$;

CREATE FUNCTION public.search_index_sync_calendar_events() RETURNS trigger
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

CREATE FUNCTION public.search_index_sync_feedback() RETURNS trigger
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

CREATE FUNCTION public.search_index_sync_ideas() RETURNS trigger
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

CREATE FUNCTION public.search_index_sync_journeys() RETURNS trigger
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

CREATE FUNCTION public.search_index_sync_pages() RETURNS trigger
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

CREATE FUNCTION public.search_index_sync_problems() RETURNS trigger
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

CREATE FUNCTION public.search_index_sync_resources() RETURNS trigger
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

CREATE FUNCTION public.search_index_sync_stories() RETURNS trigger
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

CREATE FUNCTION public.search_index_sync_tasks() RETURNS trigger
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

CREATE FUNCTION public.search_index_upsert(search_project_id uuid, search_artifact_type text, search_artifact_id uuid, search_title text, search_description text, search_status text, search_href text, search_updated_at timestamp with time zone) RETURNS void
    LANGUAGE sql
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

CREATE TABLE public.account_settings (
    user_id uuid NOT NULL,
    display_name text NOT NULL,
    bio text,
    theme text DEFAULT 'System'::text NOT NULL,
    density text DEFAULT 'Comfortable'::text NOT NULL,
    landing text DEFAULT 'Last Project'::text NOT NULL,
    time_format text DEFAULT '24-hour'::text NOT NULL,
    in_app_notifications boolean DEFAULT true NOT NULL,
    email_notifications boolean DEFAULT true NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT account_settings_density_check CHECK ((density = ANY (ARRAY['Comfortable'::text, 'Compact'::text]))),
    CONSTRAINT account_settings_theme_check CHECK ((theme = ANY (ARRAY['Light'::text, 'Dark'::text, 'System'::text]))),
    CONSTRAINT account_settings_time_format_check CHECK ((time_format = ANY (ARRAY['12-hour'::text, '24-hour'::text])))
);

CREATE TABLE public.activity_log (
    id bigint NOT NULL,
    project_id uuid NOT NULL,
    actor_user_id uuid,
    artifact_type public.artifact_type,
    artifact_id uuid,
    action text NOT NULL,
    payload jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT activity_log_action_not_blank_ck CHECK ((length(TRIM(BOTH FROM action)) > 0))
);

CREATE SEQUENCE public.activity_log_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.activity_log_id_seq OWNED BY public.activity_log.id;

CREATE TABLE public.artifact_links (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    source_type public.artifact_type NOT NULL,
    source_id uuid NOT NULL,
    target_type public.artifact_type NOT NULL,
    target_id uuid NOT NULL,
    link_kind text DEFAULT 'related'::text NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT artifact_links_check CHECK ((NOT ((source_type = target_type) AND (source_id = target_id))))
);

CREATE TABLE public.calendar_events (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    title text NOT NULL,
    description text,
    event_type public.calendar_event_type DEFAULT 'Manual'::public.calendar_event_type NOT NULL,
    phase public.calendar_phase DEFAULT 'None'::public.calendar_phase NOT NULL,
    artifact_type public.calendar_artifact_type,
    artifact_id uuid,
    starts_at timestamp with time zone NOT NULL,
    ends_at timestamp with time zone NOT NULL,
    owner_user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    all_day boolean DEFAULT true NOT NULL,
    start_time text,
    end_time text,
    location text,
    event_kind text,
    linked_artifacts jsonb DEFAULT '[]'::jsonb NOT NULL,
    tags jsonb DEFAULT '[]'::jsonb NOT NULL,
    source_title text,
    CONSTRAINT calendar_events_check CHECK ((ends_at >= starts_at))
);

CREATE TABLE public.feedback (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    slug text NOT NULL,
    title text NOT NULL,
    owner_user_id uuid NOT NULL,
    outcome public.feedback_outcome,
    document_id text,
    document_revision integer DEFAULT 1 NOT NULL,
    content_hash text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    primary_task_id uuid,
    is_orphan boolean DEFAULT true NOT NULL,
    status public.feedback_status DEFAULT 'Active'::public.feedback_status NOT NULL,
    archived_from_status public.feedback_status,
    CONSTRAINT feedback_document_revision_check CHECK ((document_revision > 0))
);

CREATE TABLE public.global_feedback_submissions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    project_id uuid,
    subject text NOT NULL,
    message text NOT NULL,
    page_path text DEFAULT ''::text NOT NULL,
    context jsonb DEFAULT '{}'::jsonb NOT NULL,
    submitted_at timestamp with time zone DEFAULT now() NOT NULL,
    email_queued_at timestamp with time zone,
    email_sent_at timestamp with time zone,
    email_error text,
    CONSTRAINT global_feedback_submissions_message_check CHECK ((length(TRIM(BOTH FROM message)) > 0)),
    CONSTRAINT global_feedback_submissions_subject_check CHECK ((length(TRIM(BOTH FROM subject)) > 0))
);

CREATE TABLE public.ideas (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    slug text NOT NULL,
    title text NOT NULL,
    owner_user_id uuid NOT NULL,
    status public.idea_status DEFAULT 'Considered'::public.idea_status NOT NULL,
    selected_at timestamp with time zone,
    document_id text,
    document_revision integer DEFAULT 1 NOT NULL,
    content_hash text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    primary_problem_id uuid,
    is_orphan boolean DEFAULT true NOT NULL,
    archived_from_status public.idea_status,
    CONSTRAINT ideas_document_revision_check CHECK ((document_revision > 0))
);

CREATE TABLE public.journeys (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    slug text NOT NULL,
    title text NOT NULL,
    owner_user_id uuid NOT NULL,
    status public.journey_status DEFAULT 'Draft'::public.journey_status NOT NULL,
    document_id text,
    document_revision integer DEFAULT 1 NOT NULL,
    content_hash text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    is_orphan boolean DEFAULT true NOT NULL,
    archived_from_status public.journey_status,
    CONSTRAINT journeys_document_revision_check CHECK ((document_revision > 0))
);

CREATE TABLE public.notifications (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    project_id uuid,
    source_type public.notification_source_type NOT NULL,
    source_id uuid,
    title text NOT NULL,
    message text NOT NULL,
    is_read boolean DEFAULT false NOT NULL,
    read_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE public.pages (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    slug text NOT NULL,
    title text NOT NULL,
    owner_user_id uuid NOT NULL,
    status public.page_status DEFAULT 'Draft'::public.page_status NOT NULL,
    document_id text,
    document_revision integer DEFAULT 1 NOT NULL,
    content_hash text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    is_orphan boolean DEFAULT true NOT NULL,
    archived_from_status public.page_status,
    CONSTRAINT pages_document_revision_check CHECK ((document_revision > 0))
);

CREATE TABLE public.problems (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    slug text NOT NULL,
    title text NOT NULL,
    owner_user_id uuid NOT NULL,
    status public.problem_status DEFAULT 'Draft'::public.problem_status NOT NULL,
    is_locked boolean DEFAULT false NOT NULL,
    document_id text,
    document_revision integer DEFAULT 1 NOT NULL,
    content_hash text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    is_orphan boolean DEFAULT true NOT NULL,
    archived_from_status public.problem_status,
    CONSTRAINT problems_document_revision_check CHECK ((document_revision > 0))
);

CREATE TABLE public.project_invites (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    email public.citext NOT NULL,
    assigned_role public.project_role NOT NULL,
    permission_mask bigint DEFAULT 0 NOT NULL,
    invited_by_user_id uuid NOT NULL,
    inviter_role public.project_role NOT NULL,
    status public.invite_status DEFAULT 'pending'::public.invite_status NOT NULL,
    sent_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    accepted_at timestamp with time zone,
    declined_at timestamp with time zone,
    cancelled_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT project_invites_permission_mask_check CHECK ((permission_mask >= 0))
);

CREATE TABLE public.project_members (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    user_id uuid NOT NULL,
    role public.project_role DEFAULT 'Member'::public.project_role NOT NULL,
    permission_mask bigint DEFAULT 0 NOT NULL,
    is_custom boolean DEFAULT false NOT NULL,
    status public.member_status DEFAULT 'Active'::public.member_status NOT NULL,
    joined_at date,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT project_members_permission_mask_check CHECK ((permission_mask >= 0))
);

CREATE TABLE public.project_settings (
    project_id uuid NOT NULL,
    project_name text NOT NULL,
    project_description text,
    project_status public.project_status DEFAULT 'Active'::public.project_status NOT NULL,
    whiteboards_enabled boolean DEFAULT true NOT NULL,
    advanced_databases_enabled boolean DEFAULT true NOT NULL,
    calendar_manual_events_enabled boolean DEFAULT true NOT NULL,
    resource_versioning_enabled boolean DEFAULT true NOT NULL,
    feedback_aggregation_enabled boolean DEFAULT true NOT NULL,
    notify_artifact_created boolean DEFAULT true NOT NULL,
    notify_artifact_locked boolean DEFAULT true NOT NULL,
    notify_feedback_added boolean DEFAULT true NOT NULL,
    notify_resource_updated boolean DEFAULT true NOT NULL,
    delivery_channel text DEFAULT 'In-app'::text NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT project_settings_delivery_channel_check CHECK ((delivery_channel = ANY (ARRAY['In-app'::text, 'Email'::text])))
);

CREATE TABLE public.projects (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    slug text NOT NULL,
    name text NOT NULL,
    organization_name text NOT NULL,
    icon text NOT NULL,
    description text,
    status public.project_status DEFAULT 'Active'::public.project_status NOT NULL,
    owner_user_id uuid NOT NULL,
    created_by_user_id uuid NOT NULL,
    archived_at timestamp with time zone,
    last_updated_at timestamp with time zone DEFAULT now() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE public.resource_versions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    resource_id uuid NOT NULL,
    project_id uuid NOT NULL,
    version integer NOT NULL,
    title text NOT NULL,
    owner_user_id uuid NOT NULL,
    document_id text,
    document_revision integer DEFAULT 1 NOT NULL,
    content_hash text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT resource_versions_document_revision_check CHECK ((document_revision > 0)),
    CONSTRAINT resource_versions_version_check CHECK ((version > 0))
);

CREATE TABLE public.resources (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    slug text NOT NULL,
    title text NOT NULL,
    owner_user_id uuid NOT NULL,
    status public.resource_status DEFAULT 'Active'::public.resource_status NOT NULL,
    document_id text,
    document_revision integer DEFAULT 1 NOT NULL,
    content_hash text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    file_type text DEFAULT 'PDF'::text NOT NULL,
    doc_type text DEFAULT 'Other'::text NOT NULL,
    archived_from_status public.resource_status,
    CONSTRAINT resources_document_revision_check CHECK ((document_revision > 0))
);

CREATE TABLE public.role_permissions (
    project_id uuid NOT NULL,
    role public.project_role NOT NULL,
    permission_mask bigint DEFAULT 0 NOT NULL,
    updated_by_user_id uuid,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT role_permissions_permission_mask_check CHECK ((permission_mask >= 0))
);

CREATE TABLE public.search_index (
    project_id uuid NOT NULL,
    artifact_type text NOT NULL,
    artifact_id uuid NOT NULL,
    title text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    status text DEFAULT ''::text NOT NULL,
    href text NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    search_vector tsvector NOT NULL
);

CREATE TABLE public.stories (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    slug text NOT NULL,
    title text NOT NULL,
    persona_name text,
    pain_points_count integer DEFAULT 0 NOT NULL,
    problem_hypotheses_count integer DEFAULT 0 NOT NULL,
    owner_user_id uuid NOT NULL,
    status public.story_status DEFAULT 'Draft'::public.story_status NOT NULL,
    is_orphan boolean DEFAULT true NOT NULL,
    document_id text,
    document_revision integer DEFAULT 1 NOT NULL,
    content_hash text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    archived_from_status public.story_status,
    CONSTRAINT stories_document_revision_check CHECK ((document_revision > 0)),
    CONSTRAINT stories_pain_points_count_check CHECK ((pain_points_count >= 0)),
    CONSTRAINT stories_problem_hypotheses_count_check CHECK ((problem_hypotheses_count >= 0))
);

CREATE TABLE public.system_settings (
    key text NOT NULL,
    value jsonb NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE public.task_assignees (
    task_id uuid NOT NULL,
    project_id uuid NOT NULL,
    user_id uuid NOT NULL,
    assigned_at timestamp with time zone DEFAULT now() NOT NULL,
    assigned_by_user_id uuid
);

CREATE TABLE public.tasks (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    slug text NOT NULL,
    title text NOT NULL,
    owner_user_id uuid NOT NULL,
    status public.task_status DEFAULT 'Planned'::public.task_status NOT NULL,
    due_at timestamp with time zone,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    document_id text,
    document_revision integer DEFAULT 1 NOT NULL,
    content_hash text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    primary_idea_id uuid,
    is_orphan boolean DEFAULT true NOT NULL,
    CONSTRAINT tasks_document_revision_check CHECK ((document_revision > 0))
);

CREATE TABLE public.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email public.citext NOT NULL,
    password_hash text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    account_version integer DEFAULT 1 NOT NULL,
    name text NOT NULL,
    is_email_verified boolean DEFAULT false NOT NULL,
    last_login_at timestamp with time zone
);

ALTER TABLE ONLY public.activity_log ALTER COLUMN id SET DEFAULT nextval('public.activity_log_id_seq'::regclass);

ALTER TABLE ONLY public.account_settings
    ADD CONSTRAINT account_settings_pkey PRIMARY KEY (user_id);

ALTER TABLE ONLY public.activity_log
    ADD CONSTRAINT activity_log_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.artifact_links
    ADD CONSTRAINT artifact_links_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.artifact_links
    ADD CONSTRAINT artifact_links_project_id_source_type_source_id_target_type_key UNIQUE (project_id, source_type, source_id, target_type, target_id, link_kind);

ALTER TABLE ONLY public.calendar_events
    ADD CONSTRAINT calendar_events_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.feedback
    ADD CONSTRAINT feedback_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.feedback
    ADD CONSTRAINT feedback_project_id_slug_key UNIQUE (project_id, slug);

ALTER TABLE ONLY public.global_feedback_submissions
    ADD CONSTRAINT global_feedback_submissions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.ideas
    ADD CONSTRAINT ideas_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.ideas
    ADD CONSTRAINT ideas_project_id_slug_key UNIQUE (project_id, slug);

ALTER TABLE ONLY public.journeys
    ADD CONSTRAINT journeys_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.journeys
    ADD CONSTRAINT journeys_project_id_slug_key UNIQUE (project_id, slug);

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.pages
    ADD CONSTRAINT pages_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.pages
    ADD CONSTRAINT pages_project_id_slug_key UNIQUE (project_id, slug);

ALTER TABLE ONLY public.problems
    ADD CONSTRAINT problems_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.problems
    ADD CONSTRAINT problems_project_id_slug_key UNIQUE (project_id, slug);

ALTER TABLE ONLY public.project_invites
    ADD CONSTRAINT project_invites_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.project_members
    ADD CONSTRAINT project_members_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.project_members
    ADD CONSTRAINT project_members_project_id_user_id_key UNIQUE (project_id, user_id);

ALTER TABLE ONLY public.project_settings
    ADD CONSTRAINT project_settings_pkey PRIMARY KEY (project_id);

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_slug_key UNIQUE (slug);

ALTER TABLE ONLY public.resource_versions
    ADD CONSTRAINT resource_versions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.resource_versions
    ADD CONSTRAINT resource_versions_resource_id_version_key UNIQUE (resource_id, version);

ALTER TABLE ONLY public.resources
    ADD CONSTRAINT resources_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.resources
    ADD CONSTRAINT resources_project_id_slug_key UNIQUE (project_id, slug);

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_pkey PRIMARY KEY (project_id, role);

ALTER TABLE ONLY public.search_index
    ADD CONSTRAINT search_index_pkey PRIMARY KEY (project_id, artifact_type, artifact_id);

ALTER TABLE ONLY public.stories
    ADD CONSTRAINT stories_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.stories
    ADD CONSTRAINT stories_project_id_slug_key UNIQUE (project_id, slug);

ALTER TABLE ONLY public.system_settings
    ADD CONSTRAINT system_settings_pkey PRIMARY KEY (key);

ALTER TABLE ONLY public.task_assignees
    ADD CONSTRAINT task_assignees_pkey PRIMARY KEY (task_id, user_id);

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_project_id_slug_key UNIQUE (project_id, slug);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

CREATE INDEX activity_log_actor_created_idx ON public.activity_log USING btree (actor_user_id, created_at DESC);

CREATE INDEX activity_log_project_created_idx ON public.activity_log USING btree (project_id, created_at DESC);

CREATE INDEX artifact_links_source_idx ON public.artifact_links USING btree (project_id, source_type, source_id);

CREATE INDEX artifact_links_target_idx ON public.artifact_links USING btree (project_id, target_type, target_id);

CREATE INDEX calendar_events_project_event_type_idx ON public.calendar_events USING btree (project_id, event_type);

CREATE INDEX calendar_events_project_starts_idx ON public.calendar_events USING btree (project_id, starts_at);

CREATE INDEX calendar_events_project_updated_idx ON public.calendar_events USING btree (project_id, updated_at DESC);

CREATE INDEX feedback_project_orphan_idx ON public.feedback USING btree (project_id, is_orphan);

CREATE INDEX feedback_project_primary_task_idx ON public.feedback USING btree (project_id, primary_task_id);

CREATE INDEX feedback_project_status_idx ON public.feedback USING btree (project_id, status);

CREATE INDEX feedback_project_updated_idx ON public.feedback USING btree (project_id, updated_at DESC);

CREATE INDEX global_feedback_submissions_project_submitted_idx ON public.global_feedback_submissions USING btree (project_id, submitted_at DESC);

CREATE INDEX global_feedback_submissions_user_submitted_idx ON public.global_feedback_submissions USING btree (user_id, submitted_at DESC);

CREATE INDEX ideas_project_orphan_idx ON public.ideas USING btree (project_id, is_orphan);

CREATE INDEX ideas_project_primary_problem_idx ON public.ideas USING btree (project_id, primary_problem_id);

CREATE INDEX ideas_project_status_idx ON public.ideas USING btree (project_id, status);

CREATE INDEX ideas_project_updated_idx ON public.ideas USING btree (project_id, updated_at DESC);

CREATE INDEX journeys_project_orphan_idx ON public.journeys USING btree (project_id, is_orphan);

CREATE INDEX journeys_project_status_idx ON public.journeys USING btree (project_id, status);

CREATE INDEX journeys_project_updated_idx ON public.journeys USING btree (project_id, updated_at DESC);

CREATE INDEX notifications_project_created_idx ON public.notifications USING btree (project_id, created_at DESC);

CREATE INDEX notifications_user_read_created_idx ON public.notifications USING btree (user_id, is_read, created_at DESC);

CREATE INDEX pages_project_orphan_idx ON public.pages USING btree (project_id, is_orphan);

CREATE INDEX pages_project_status_idx ON public.pages USING btree (project_id, status);

CREATE INDEX pages_project_updated_idx ON public.pages USING btree (project_id, updated_at DESC);

CREATE INDEX problems_project_orphan_idx ON public.problems USING btree (project_id, is_orphan);

CREATE INDEX problems_project_status_idx ON public.problems USING btree (project_id, status);

CREATE INDEX problems_project_updated_idx ON public.problems USING btree (project_id, updated_at DESC);

CREATE INDEX project_invites_expires_at_idx ON public.project_invites USING btree (expires_at);

CREATE UNIQUE INDEX project_invites_pending_email_unique_idx ON public.project_invites USING btree (project_id, email) WHERE (status = 'pending'::public.invite_status);

CREATE INDEX project_invites_project_status_idx ON public.project_invites USING btree (project_id, status);

CREATE INDEX project_members_project_custom_idx ON public.project_members USING btree (project_id, is_custom);

CREATE INDEX project_members_project_role_idx ON public.project_members USING btree (project_id, role);

CREATE INDEX project_members_project_status_idx ON public.project_members USING btree (project_id, status);

CREATE INDEX project_members_user_id_idx ON public.project_members USING btree (user_id);

CREATE INDEX projects_last_updated_at_idx ON public.projects USING btree (last_updated_at DESC);

CREATE INDEX projects_owner_user_id_idx ON public.projects USING btree (owner_user_id);

CREATE INDEX projects_status_idx ON public.projects USING btree (status);

CREATE INDEX resource_versions_resource_created_idx ON public.resource_versions USING btree (resource_id, created_at DESC);

CREATE INDEX resources_project_doc_type_idx ON public.resources USING btree (project_id, doc_type);

CREATE INDEX resources_project_status_doc_type_idx ON public.resources USING btree (project_id, status, doc_type);

CREATE INDEX resources_project_status_idx ON public.resources USING btree (project_id, status);

CREATE INDEX resources_project_updated_idx ON public.resources USING btree (project_id, updated_at DESC);

CREATE INDEX role_permissions_project_mask_idx ON public.role_permissions USING btree (project_id, permission_mask);

CREATE INDEX search_index_project_updated_idx ON public.search_index USING btree (project_id, updated_at DESC);

CREATE INDEX search_index_vector_idx ON public.search_index USING gin (search_vector);

CREATE INDEX stories_project_status_idx ON public.stories USING btree (project_id, status);

CREATE INDEX stories_project_updated_idx ON public.stories USING btree (project_id, updated_at DESC);

CREATE INDEX task_assignees_project_task_idx ON public.task_assignees USING btree (project_id, task_id);

CREATE INDEX task_assignees_project_user_idx ON public.task_assignees USING btree (project_id, user_id);

CREATE INDEX tasks_project_orphan_idx ON public.tasks USING btree (project_id, is_orphan);

CREATE INDEX tasks_project_primary_idea_idx ON public.tasks USING btree (project_id, primary_idea_id);

CREATE INDEX tasks_project_status_idx ON public.tasks USING btree (project_id, status);

CREATE INDEX tasks_project_updated_idx ON public.tasks USING btree (project_id, updated_at DESC);

CREATE INDEX users_created_at_idx ON public.users USING btree (created_at);

CREATE UNIQUE INDEX users_email_unique_idx ON public.users USING btree (email);

CREATE TRIGGER feedback_cleanup_links_trg BEFORE DELETE ON public.feedback FOR EACH ROW EXECUTE FUNCTION public.pb_cleanup_artifact_links('feedback');

CREATE TRIGGER feedback_sync_chain_orphan_trg BEFORE INSERT OR UPDATE OF primary_task_id ON public.feedback FOR EACH ROW EXECUTE FUNCTION public.pb_sync_chain_orphan_flags();

CREATE TRIGGER ideas_cleanup_links_trg BEFORE DELETE ON public.ideas FOR EACH ROW EXECUTE FUNCTION public.pb_cleanup_artifact_links('idea');

CREATE TRIGGER ideas_sync_chain_orphan_trg BEFORE INSERT OR UPDATE OF primary_problem_id ON public.ideas FOR EACH ROW EXECUTE FUNCTION public.pb_sync_chain_orphan_flags();

CREATE TRIGGER journeys_cleanup_links_trg BEFORE DELETE ON public.journeys FOR EACH ROW EXECUTE FUNCTION public.pb_cleanup_artifact_links('journey');

CREATE TRIGGER problems_cleanup_links_trg BEFORE DELETE ON public.problems FOR EACH ROW EXECUTE FUNCTION public.pb_cleanup_artifact_links('problem');

CREATE TRIGGER problems_sync_orphan_link_trg AFTER INSERT OR DELETE OR UPDATE ON public.artifact_links FOR EACH ROW EXECUTE FUNCTION public.pb_sync_problem_orphan_on_link_change();

CREATE TRIGGER problems_sync_orphan_row_trg BEFORE INSERT OR UPDATE OF project_id, id ON public.problems FOR EACH ROW EXECUTE FUNCTION public.pb_sync_problem_orphan_on_row();

CREATE TRIGGER stories_cleanup_links_trg BEFORE DELETE ON public.stories FOR EACH ROW EXECUTE FUNCTION public.pb_cleanup_artifact_links('story');

CREATE TRIGGER tasks_cleanup_links_trg BEFORE DELETE ON public.tasks FOR EACH ROW EXECUTE FUNCTION public.pb_cleanup_artifact_links('task');

CREATE TRIGGER tasks_sync_chain_orphan_trg BEFORE INSERT OR UPDATE OF primary_idea_id ON public.tasks FOR EACH ROW EXECUTE FUNCTION public.pb_sync_chain_orphan_flags();

CREATE TRIGGER trg_search_index_calendar_events AFTER INSERT OR DELETE OR UPDATE ON public.calendar_events FOR EACH ROW EXECUTE FUNCTION public.search_index_sync_calendar_events();

CREATE TRIGGER trg_search_index_feedback AFTER INSERT OR DELETE OR UPDATE ON public.feedback FOR EACH ROW EXECUTE FUNCTION public.search_index_sync_feedback();

CREATE TRIGGER trg_search_index_ideas AFTER INSERT OR DELETE OR UPDATE ON public.ideas FOR EACH ROW EXECUTE FUNCTION public.search_index_sync_ideas();

CREATE TRIGGER trg_search_index_journeys AFTER INSERT OR DELETE OR UPDATE ON public.journeys FOR EACH ROW EXECUTE FUNCTION public.search_index_sync_journeys();

CREATE TRIGGER trg_search_index_pages AFTER INSERT OR DELETE OR UPDATE ON public.pages FOR EACH ROW EXECUTE FUNCTION public.search_index_sync_pages();

CREATE TRIGGER trg_search_index_problems AFTER INSERT OR DELETE OR UPDATE ON public.problems FOR EACH ROW EXECUTE FUNCTION public.search_index_sync_problems();

CREATE TRIGGER trg_search_index_resources AFTER INSERT OR DELETE OR UPDATE ON public.resources FOR EACH ROW EXECUTE FUNCTION public.search_index_sync_resources();

CREATE TRIGGER trg_search_index_stories AFTER INSERT OR DELETE OR UPDATE ON public.stories FOR EACH ROW EXECUTE FUNCTION public.search_index_sync_stories();

CREATE TRIGGER trg_search_index_tasks AFTER INSERT OR DELETE OR UPDATE ON public.tasks FOR EACH ROW EXECUTE FUNCTION public.search_index_sync_tasks();

ALTER TABLE ONLY public.account_settings
    ADD CONSTRAINT account_settings_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.activity_log
    ADD CONSTRAINT activity_log_actor_user_id_fkey FOREIGN KEY (actor_user_id) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.activity_log
    ADD CONSTRAINT activity_log_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.artifact_links
    ADD CONSTRAINT artifact_links_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.artifact_links
    ADD CONSTRAINT artifact_links_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.calendar_events
    ADD CONSTRAINT calendar_events_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.calendar_events
    ADD CONSTRAINT calendar_events_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.feedback
    ADD CONSTRAINT feedback_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.feedback
    ADD CONSTRAINT feedback_primary_task_id_fkey FOREIGN KEY (primary_task_id) REFERENCES public.tasks(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.feedback
    ADD CONSTRAINT feedback_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.global_feedback_submissions
    ADD CONSTRAINT global_feedback_submissions_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.global_feedback_submissions
    ADD CONSTRAINT global_feedback_submissions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.ideas
    ADD CONSTRAINT ideas_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.ideas
    ADD CONSTRAINT ideas_primary_problem_id_fkey FOREIGN KEY (primary_problem_id) REFERENCES public.problems(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.ideas
    ADD CONSTRAINT ideas_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.journeys
    ADD CONSTRAINT journeys_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.journeys
    ADD CONSTRAINT journeys_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.pages
    ADD CONSTRAINT pages_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.pages
    ADD CONSTRAINT pages_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.problems
    ADD CONSTRAINT problems_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.problems
    ADD CONSTRAINT problems_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.project_invites
    ADD CONSTRAINT project_invites_invited_by_user_id_fkey FOREIGN KEY (invited_by_user_id) REFERENCES public.users(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.project_invites
    ADD CONSTRAINT project_invites_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.project_members
    ADD CONSTRAINT project_members_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.project_members
    ADD CONSTRAINT project_members_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.project_settings
    ADD CONSTRAINT project_settings_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.resource_versions
    ADD CONSTRAINT resource_versions_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.resource_versions
    ADD CONSTRAINT resource_versions_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.resource_versions
    ADD CONSTRAINT resource_versions_resource_id_fkey FOREIGN KEY (resource_id) REFERENCES public.resources(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.resources
    ADD CONSTRAINT resources_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.resources
    ADD CONSTRAINT resources_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_updated_by_user_id_fkey FOREIGN KEY (updated_by_user_id) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.search_index
    ADD CONSTRAINT search_index_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.stories
    ADD CONSTRAINT stories_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.stories
    ADD CONSTRAINT stories_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.task_assignees
    ADD CONSTRAINT task_assignees_assigned_by_user_id_fkey FOREIGN KEY (assigned_by_user_id) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.task_assignees
    ADD CONSTRAINT task_assignees_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.task_assignees
    ADD CONSTRAINT task_assignees_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.task_assignees
    ADD CONSTRAINT task_assignees_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_primary_idea_id_fkey FOREIGN KEY (primary_idea_id) REFERENCES public.ideas(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;

