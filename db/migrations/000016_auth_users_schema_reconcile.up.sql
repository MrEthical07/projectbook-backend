-- Reconcile legacy users schema (000003) with auth query expectations from core auth tables.
-- This keeps migration history forward-only while making fresh environments consistent.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS name TEXT;

UPDATE users
SET name = split_part(email::TEXT, '@', 1)
WHERE name IS NULL OR btrim(name) = '';

ALTER TABLE users
    ALTER COLUMN name SET NOT NULL;

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS is_email_verified BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMPTZ NULL;

ALTER TABLE users
    ALTER COLUMN email TYPE CITEXT USING email::CITEXT;
