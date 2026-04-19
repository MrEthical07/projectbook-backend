ALTER TABLE users
ADD COLUMN IF NOT EXISTS account_version INTEGER NOT NULL DEFAULT 1;

UPDATE users
SET account_version = 1
WHERE account_version < 1;
