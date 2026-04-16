-- SAFE: removed legacy destructive schema reset statements; migration is additive and idempotent where possible.
CREATE TABLE IF NOT EXISTS stories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    slug TEXT NOT NULL,
    title TEXT NOT NULL,
    persona_name TEXT NULL,
    pain_points_count INTEGER NOT NULL DEFAULT 0 CHECK (pain_points_count >= 0),
    problem_hypotheses_count INTEGER NOT NULL DEFAULT 0 CHECK (problem_hypotheses_count >= 0),
    owner_user_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    status story_status NOT NULL DEFAULT 'Draft',
    is_orphan BOOLEAN NOT NULL DEFAULT TRUE,
    document_id TEXT NULL,
    document_revision INTEGER NOT NULL DEFAULT 1 CHECK (document_revision > 0),
    content_hash TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, slug)
);

CREATE TABLE IF NOT EXISTS journeys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    slug TEXT NOT NULL,
    title TEXT NOT NULL,
    owner_user_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    status journey_status NOT NULL DEFAULT 'Draft',
    document_id TEXT NULL,
    document_revision INTEGER NOT NULL DEFAULT 1 CHECK (document_revision > 0),
    content_hash TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, slug)
);

CREATE TABLE IF NOT EXISTS problems (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    slug TEXT NOT NULL,
    title TEXT NOT NULL,
    owner_user_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    status problem_status NOT NULL DEFAULT 'Draft',
    is_locked BOOLEAN NOT NULL DEFAULT FALSE,
    document_id TEXT NULL,
    document_revision INTEGER NOT NULL DEFAULT 1 CHECK (document_revision > 0),
    content_hash TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, slug)
);

CREATE TABLE IF NOT EXISTS ideas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    slug TEXT NOT NULL,
    title TEXT NOT NULL,
    owner_user_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    status idea_status NOT NULL DEFAULT 'Considered',
    selected_at TIMESTAMPTZ NULL,
    document_id TEXT NULL,
    document_revision INTEGER NOT NULL DEFAULT 1 CHECK (document_revision > 0),
    content_hash TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, slug)
);

CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    slug TEXT NOT NULL,
    title TEXT NOT NULL,
    owner_user_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    status task_status NOT NULL DEFAULT 'Planned',
    due_at TIMESTAMPTZ NULL,
    started_at TIMESTAMPTZ NULL,
    completed_at TIMESTAMPTZ NULL,
    document_id TEXT NULL,
    document_revision INTEGER NOT NULL DEFAULT 1 CHECK (document_revision > 0),
    content_hash TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, slug)
);

CREATE TABLE IF NOT EXISTS feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    slug TEXT NOT NULL,
    title TEXT NOT NULL,
    owner_user_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    outcome feedback_outcome NULL,
    document_id TEXT NULL,
    document_revision INTEGER NOT NULL DEFAULT 1 CHECK (document_revision > 0),
    content_hash TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, slug)
);

CREATE TABLE IF NOT EXISTS resources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    slug TEXT NOT NULL,
    title TEXT NOT NULL,
    owner_user_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    status resource_status NOT NULL DEFAULT 'Active',
    document_id TEXT NULL,
    document_revision INTEGER NOT NULL DEFAULT 1 CHECK (document_revision > 0),
    content_hash TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, slug)
);

CREATE TABLE IF NOT EXISTS resource_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_id UUID NOT NULL REFERENCES resources (id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    version INTEGER NOT NULL CHECK (version > 0),
    title TEXT NOT NULL,
    owner_user_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    document_id TEXT NULL,
    document_revision INTEGER NOT NULL DEFAULT 1 CHECK (document_revision > 0),
    content_hash TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (resource_id, version)
);

CREATE TABLE IF NOT EXISTS pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    slug TEXT NOT NULL,
    title TEXT NOT NULL,
    owner_user_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    status page_status NOT NULL DEFAULT 'Draft',
    document_id TEXT NULL,
    document_revision INTEGER NOT NULL DEFAULT 1 CHECK (document_revision > 0),
    content_hash TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, slug)
);

CREATE TABLE IF NOT EXISTS calendar_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT NULL,
    event_type calendar_event_type NOT NULL DEFAULT 'Manual',
    phase calendar_phase NOT NULL DEFAULT 'None',
    artifact_type calendar_artifact_type NULL,
    artifact_id UUID NULL,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    owner_user_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (ends_at >= starts_at)
);

CREATE TABLE IF NOT EXISTS artifact_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    source_type artifact_type NOT NULL,
    source_id UUID NOT NULL,
    target_type artifact_type NOT NULL,
    target_id UUID NOT NULL,
    link_kind TEXT NOT NULL DEFAULT 'related',
    created_by_user_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, source_type, source_id, target_type, target_id, link_kind),
    CHECK (NOT (source_type = target_type AND source_id = target_id))
);

CREATE TABLE IF NOT EXISTS activity_log (
    id BIGSERIAL PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    actor_user_id UUID NULL REFERENCES users (id) ON DELETE SET NULL,
    artifact_type artifact_type NULL,
    artifact_id UUID NULL,
    action TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    project_id UUID NULL REFERENCES projects (id) ON DELETE SET NULL,
    source_type notification_source_type NOT NULL,
    source_id UUID NULL,
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    read_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS document_sync_outbox (
    id BIGSERIAL PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    artifact_type artifact_type NOT NULL,
    artifact_id UUID NOT NULL,
    operation TEXT NOT NULL CHECK (operation IN ('upsert', 'delete')),
    document_id TEXT NULL,
    document_revision INTEGER NOT NULL DEFAULT 1 CHECK (document_revision > 0),
    payload JSONB NOT NULL DEFAULT '{}'::JSONB,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    attempt_count INTEGER NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_error TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, artifact_type, artifact_id, document_revision, operation)
);

