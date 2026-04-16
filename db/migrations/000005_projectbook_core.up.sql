-- SAFE: removed legacy destructive schema reset statements; migration is additive and idempotent where possible.
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email CITEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    is_email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    last_login_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS auth_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_agent TEXT NULL,
    ip_address INET NULL,
    device_label TEXT NULL
);

CREATE TABLE IF NOT EXISTS email_verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS auth_email_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NULL REFERENCES users (id) ON DELETE SET NULL,
    kind TEXT NOT NULL CHECK (kind IN ('verify', 'reset')),
    recipient_email CITEXT NOT NULL,
    link TEXT NOT NULL,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS account_settings (
    user_id UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    display_name TEXT NOT NULL,
    bio TEXT NULL,
    theme TEXT NOT NULL DEFAULT 'System' CHECK (theme IN ('Light', 'Dark', 'System')),
    density TEXT NOT NULL DEFAULT 'Comfortable' CHECK (density IN ('Comfortable', 'Compact')),
    landing TEXT NOT NULL DEFAULT 'Last Project',
    time_format TEXT NOT NULL DEFAULT '24-hour' CHECK (time_format IN ('12-hour', '24-hour')),
    in_app_notifications BOOLEAN NOT NULL DEFAULT TRUE,
    email_notifications BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    organization_name TEXT NOT NULL,
    icon TEXT NOT NULL,
    description TEXT NULL,
    status project_status NOT NULL DEFAULT 'Active',
    owner_user_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    created_by_user_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    archived_at TIMESTAMPTZ NULL,
    last_updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS project_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role project_role NOT NULL DEFAULT 'Member',
    permission_mask BIGINT NOT NULL DEFAULT 0 CHECK (permission_mask >= 0),
    is_custom BOOLEAN NOT NULL DEFAULT FALSE,
    status member_status NOT NULL DEFAULT 'Active',
    joined_at DATE NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, user_id)
);

CREATE TABLE IF NOT EXISTS project_invites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    email CITEXT NOT NULL,
    assigned_role project_role NOT NULL,
    permission_mask BIGINT NOT NULL DEFAULT 0 CHECK (permission_mask >= 0),
    invited_by_user_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    inviter_role project_role NOT NULL,
    status invite_status NOT NULL DEFAULT 'pending',
    sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ NULL,
    declined_at TIMESTAMPTZ NULL,
    cancelled_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS project_settings (
    project_id UUID PRIMARY KEY REFERENCES projects (id) ON DELETE CASCADE,
    project_name TEXT NOT NULL,
    project_description TEXT NULL,
    project_status project_status NOT NULL DEFAULT 'Active',
    whiteboards_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    advanced_databases_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    calendar_manual_events_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    resource_versioning_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    feedback_aggregation_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    notify_artifact_created BOOLEAN NOT NULL DEFAULT TRUE,
    notify_artifact_locked BOOLEAN NOT NULL DEFAULT TRUE,
    notify_feedback_added BOOLEAN NOT NULL DEFAULT TRUE,
    notify_resource_updated BOOLEAN NOT NULL DEFAULT TRUE,
    delivery_channel TEXT NOT NULL DEFAULT 'In-app' CHECK (delivery_channel IN ('In-app', 'Email')),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS role_permissions (
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    role project_role NOT NULL,
    permission_mask BIGINT NOT NULL DEFAULT 0 CHECK (permission_mask >= 0),
    updated_by_user_id UUID NULL REFERENCES users (id) ON DELETE SET NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (project_id, role)
);

