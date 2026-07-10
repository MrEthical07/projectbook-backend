-- Reverses 000014. Drops the users.account_version column. The up's UPDATE only
-- normalised values within this column, so nothing outside it needs restoring.
ALTER TABLE users DROP COLUMN IF EXISTS account_version;
