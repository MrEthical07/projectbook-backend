-- Reverses 000016. Reverts users.email back to TEXT and drops the name,
-- is_email_verified and last_login_at columns that this migration added. The up's
-- name backfill lived in the name column, so dropping it restores the exact
-- pre-000016 shape (users.email as TEXT, from 000003).
ALTER TABLE users
    ALTER COLUMN email TYPE TEXT USING email::TEXT;

ALTER TABLE users
    DROP COLUMN IF EXISTS last_login_at,
    DROP COLUMN IF EXISTS is_email_verified,
    DROP COLUMN IF EXISTS name;
