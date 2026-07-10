-- Reversal of 000017. This recreates the dropped structures exactly as they
-- were defined in migrations 000002, 000003, 000005, 000006 and 000007.
--
-- This is a schema-only rollback: the dropped tables and columns held no
-- application data (they were never written to), so nothing is lost by
-- dropping them and nothing needs restoring here beyond the structure.
-- Recreating them lets `migrate down` return the database to the exact
-- pre-000017 shape, which keeps the migration honestly reversible.

-- Legacy users columns (originally from 000003) + their index.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS role TEXT,
    ADD COLUMN IF NOT EXISTS permissions BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'active';

CREATE INDEX IF NOT EXISTS users_status_idx ON users (status);

-- tenants (originally from 000002).
CREATE TABLE IF NOT EXISTS tenants (
    id TEXT PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS tenants_created_at_idx ON tenants (created_at);

-- auth_sessions (originally from 000005) + indexes from 000007.
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

CREATE INDEX IF NOT EXISTS auth_sessions_user_expires_idx ON auth_sessions (user_id, expires_at);
CREATE INDEX IF NOT EXISTS auth_sessions_active_idx ON auth_sessions (user_id, expires_at) WHERE revoked_at IS NULL;

-- email_verification_tokens (originally from 000005) + index from 000007.
CREATE TABLE IF NOT EXISTS email_verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS email_verification_tokens_user_expires_idx ON email_verification_tokens (user_id, expires_at DESC);

-- password_reset_tokens (originally from 000005) + index from 000007.
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS password_reset_tokens_user_expires_idx ON password_reset_tokens (user_id, expires_at DESC);

-- auth_email_log (originally from 000005) + indexes from 000007.
CREATE TABLE IF NOT EXISTS auth_email_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NULL REFERENCES users (id) ON DELETE SET NULL,
    kind TEXT NOT NULL CHECK (kind IN ('verify', 'reset')),
    recipient_email CITEXT NOT NULL,
    link TEXT NOT NULL,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS auth_email_log_user_sent_idx ON auth_email_log (user_id, sent_at DESC);
CREATE INDEX IF NOT EXISTS auth_email_log_recipient_sent_idx ON auth_email_log (recipient_email, sent_at DESC);

-- document_sync_outbox (originally from 000006) + indexes from 000007.
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

CREATE INDEX IF NOT EXISTS document_sync_outbox_status_next_attempt_idx ON document_sync_outbox (status, next_attempt_at);
CREATE INDEX IF NOT EXISTS document_sync_outbox_project_artifact_idx ON document_sync_outbox (project_id, artifact_type, artifact_id);
