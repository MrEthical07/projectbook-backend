# ProjectBook Database Design (PostgreSQL Metadata + Hybrid Document Sync)

## 1. Scope

This document is the schema source of truth for ProjectBook migration generation.

Design constraints:

- Mask-based RBAC only (`permission_mask` + `is_custom`)
- No workspace layer (project is the top-level tenant boundary)
- Hybrid architecture: PostgreSQL stores relational metadata; document payload synchronization is handled through `document_sync_outbox`

References:

- `database.md` (this file)
- `rbac.md` (bit allocation and mask rules)

## 2. Migration Layout

The schema is split into ordered migration concerns:

1. `000004_projectbook_enums.up.sql` / `.down.sql`
2. `000005_projectbook_core.up.sql` / `.down.sql`
3. `000006_projectbook_artifacts.up.sql` / `.down.sql`
4. `000007_projectbook_indexes.up.sql` / `.down.sql`

Separation of concerns:

- enums first
- core/auth/project/RBAC tables second
- artifact + activity + sync tables third
- indexes last

## 3. Enums

Defined in `000004_projectbook_enums.up.sql`:

- `project_status`: `Active`, `Archived`
- `project_role`: `Owner`, `Admin`, `Editor`, `Member`, `Viewer`, `Limited Access`
- `story_status`: `Draft`, `Locked`, `Archived`
- `journey_status`: `Draft`, `Archived`
- `problem_status`: `Draft`, `Locked`, `Archived`
- `idea_status`: `Considered`, `Selected`, `Rejected`, `Archived`
- `task_status`: `Planned`, `In Progress`, `Completed`, `Abandoned`, `Blocked`
- `feedback_outcome`: `Validated`, `Invalidated`, `Needs Iteration`
- `resource_status`: `Active`, `Archived`
- `page_status`: `Draft`, `Archived`
- `calendar_event_type`: `Derived`, `Manual`
- `calendar_phase`: `Empathize`, `Define`, `Ideate`, `Prototype`, `Test`, `None`
- `calendar_artifact_type`: `Task`, `Feedback`, `Manual`
- `invite_status`: `pending`, `accepted`, `declined`, `expired`, `cancelled`
- `member_status`: `Active`, `Invited`
- `artifact_type`: `story`, `journey`, `problem`, `idea`, `task`, `feedback`, `resource`, `page`, `calendar`
- `notification_source_type`: `Project Activity`, `Project Invitation`, `System Notification`

Extensions enabled:

- `pgcrypto` (for `gen_random_uuid()`)
- `citext` (for case-insensitive email fields)

## 4. Core Domain Tables

Defined in `000005_projectbook_core.up.sql`.

### 4.1 users

Columns:

- `id UUID PK`
- `email CITEXT UNIQUE NOT NULL`
- `name TEXT NOT NULL`
- `password_hash TEXT NOT NULL`
- `is_email_verified BOOLEAN NOT NULL DEFAULT FALSE`
- `last_login_at TIMESTAMPTZ NULL`
- `created_at`, `updated_at`

### 4.2 auth_sessions

Columns:

- `id UUID PK`
- `user_id UUID FK -> users(id) ON DELETE CASCADE`
- `token_hash TEXT UNIQUE NOT NULL`
- `expires_at TIMESTAMPTZ NOT NULL`
- `revoked_at TIMESTAMPTZ NULL`
- `user_agent`, `ip_address`, `device_label`
- `created_at`

### 4.3 account_settings

Columns:

- `user_id UUID PK/FK -> users(id) ON DELETE CASCADE`
- preference fields (`theme`, `density`, `landing`, `time_format`)
- notification toggles
- `updated_at`

## 5. Project Domain + RBAC Tables

Defined in `000005_projectbook_core.up.sql`.

### 5.1 projects

Columns:

- `id UUID PK`
- `slug TEXT UNIQUE NOT NULL`
- `name`, `organization_name`, `icon`, `description`
- `status project_status`
- `owner_user_id UUID FK -> users`
- `created_by_user_id UUID FK -> users`
- `archived_at`, `last_updated_at`, `created_at`, `updated_at`

### 5.2 project_members

Columns:

- `id UUID PK`
- `project_id UUID FK -> projects(id) ON DELETE CASCADE`
- `user_id UUID FK -> users(id) ON DELETE CASCADE`
- `role project_role`
- `permission_mask BIGINT NOT NULL DEFAULT 0 CHECK (permission_mask >= 0)`
- `is_custom BOOLEAN NOT NULL DEFAULT FALSE`
- `status member_status`
- `joined_at`, `created_at`, `updated_at`

Constraints:

- `UNIQUE (project_id, user_id)`

### 5.3 role_permissions

Columns:

- `project_id UUID FK -> projects(id) ON DELETE CASCADE`
- `role project_role`
- `permission_mask BIGINT NOT NULL DEFAULT 0 CHECK (permission_mask >= 0)`
- `updated_by_user_id UUID NULL FK -> users(id) ON DELETE SET NULL`
- `updated_at`

Constraints:

- `PRIMARY KEY (project_id, role)`

### 5.4 project_invites

Columns:

- `id UUID PK`
- `project_id UUID FK -> projects`
- `email CITEXT NOT NULL`
- `assigned_role project_role`
- `permission_mask BIGINT NOT NULL DEFAULT 0`
- `invited_by_user_id UUID FK -> users`
- `inviter_role project_role`
- `status invite_status`
- `sent_at`, `expires_at`, `accepted_at`, `declined_at`, `cancelled_at`
- `created_at`, `updated_at`

### 5.5 project_settings

Columns:

- `project_id UUID PK/FK -> projects(id) ON DELETE CASCADE`
- mirrored project metadata (`project_name`, `project_description`, `project_status`)
- feature toggles
- notification toggles
- `delivery_channel` with constraint `In-app|Email`
- `updated_at`

## 6. Artifact Metadata Tables

Defined in `000006_projectbook_artifacts.up.sql`.

Tables:

- `stories`
- `journeys`
- `problems`
- `ideas`
- `tasks`
- `feedback`
- `resources`
- `resource_versions`
- `pages`
- `calendar_events`

Common metadata pattern (where applicable):

- `id UUID PK`
- `project_id UUID FK -> projects(id)`
- `slug TEXT` + `UNIQUE (project_id, slug)` for slug-based artifacts
- lifecycle/status enum
- owner linkage (`owner_user_id UUID FK -> users(id)`)
- document sync fields: `document_id`, `document_revision`, `content_hash`
- `created_at`, `updated_at`

Special constraints:

- `resource_versions`: `UNIQUE (resource_id, version)`
- `calendar_events`: `CHECK (ends_at >= starts_at)`

## 7. Relationship, Activity, Notification, Sync Tables

Defined in `000006_projectbook_artifacts.up.sql`.

### 7.1 artifact_links

- link rows between any artifact pairs
- uses `artifact_type` enum for source and target
- uniqueness over `(project_id, source_type, source_id, target_type, target_id, link_kind)`
- self-link guard via CHECK constraint

### 7.2 activity_log

- append-only project activity stream
- supports optional actor/artifact pointers
- JSON payload column (`payload JSONB`)

### 7.3 notifications

- user-facing notification inbox
- optional project and source linkage
- read tracking (`is_read`, `read_at`)

### 7.4 document_sync_outbox

- outbound sync queue for document payload operations
- operation constrained to `upsert|delete`
- status constrained to `pending|processing|completed|failed`
- retry metadata (`attempt_count`, `next_attempt_at`, `last_error`)
- uniqueness over `(project_id, artifact_type, artifact_id, document_revision, operation)`

## 8. Index Strategy

Defined in `000007_projectbook_indexes.up.sql`.

Coverage includes:

- auth lookup and active session filtering
- project ownership/status/update sorting
- member and invite query paths
- role permission lookups
- per-artifact project timeline/list sorting
- artifact link source/target traversal
- activity and notification feeds
- outbox polling (`status`, `next_attempt_at`)

Critical uniqueness/index requirement:

- project-scoped slug uniqueness is enforced directly on artifact tables with `UNIQUE (project_id, slug)` constraints

## 9. Explicitly Removed / Forbidden Schema

Not present in current migrations:

- `workspaces`
- `workspace_members`
- `role_permissions_default`
- `project_role_permissions`
- `permission_domain` enum
- `permission_action` enum

Also removed from relational ownership model:

- all `workspace_id` foreign keys and references

## 10. RBAC Rules (Mask-Based Only)

Persistent RBAC state is represented by:

- `role_permissions.permission_mask` (role default)
- `project_members.permission_mask` (member effective/custom)
- `project_members.is_custom` (override flag)

Rules:

- permissions are stored as BIGINT bit masks (64-bit compatible)
- no domain/action boolean row model exists in SQL
- no per-permission row expansions are used

## 11. Validation Checklist

The migration set is valid when:

- all foreign keys resolve in migration order
- no workspace-layer tables/references remain
- no legacy RBAC tables/enums exist
- all project-slug uniqueness constraints are project scoped
- all mask fields are BIGINT with non-negative checks

## 12. Execution Note

This task generated schema definition migrations only.

- Migrations were not applied.
- No repository/service code was generated.
- No seed data was added.
