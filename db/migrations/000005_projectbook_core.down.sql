-- Reverses 000005. Drops the auth + project core tables it created.
--
-- NOTE: the up's `CREATE TABLE IF NOT EXISTS users` was a no-op on a fresh
-- database (000003 already created users), so users is intentionally NOT dropped
-- here — it is owned by 000003's down. Tables are dropped children-first so no
-- CASCADE is needed.
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS project_settings;
DROP TABLE IF EXISTS project_invites;
DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS account_settings;
DROP TABLE IF EXISTS auth_email_log;
DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS email_verification_tokens;
DROP TABLE IF EXISTS auth_sessions;
DROP TABLE IF EXISTS projects;
