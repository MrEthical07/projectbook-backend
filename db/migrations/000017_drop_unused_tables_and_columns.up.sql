-- Drop unused/orphaned tables and legacy columns identified in the schema audit.
-- Every object below has zero references in the application code:
--   * auth is stateless (JWT): auth_sessions, email_verification_tokens,
--     password_reset_tokens and auth_email_log were designed for a DB-backed
--     session/token model that was never implemented.
--   * document_sync_outbox has no producer and no consumer.
--   * tenants is leftover multi-tenancy scaffolding.
--   * users.role / users.permissions / users.status are superseded by the
--     per-project RBAC model (project_members + role_permissions).
--
-- All dropped tables are leaf tables (nothing references them via foreign key),
-- so the drops are safe and order-independent. DROP TABLE also removes each
-- table's indexes automatically. Dropping users.status removes users_status_idx.

DROP TABLE IF EXISTS auth_sessions;
DROP TABLE IF EXISTS email_verification_tokens;
DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS auth_email_log;
DROP TABLE IF EXISTS document_sync_outbox;
DROP TABLE IF EXISTS tenants;

ALTER TABLE users
    DROP COLUMN IF EXISTS role,
    DROP COLUMN IF EXISTS permissions,
    DROP COLUMN IF EXISTS status;
