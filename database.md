# ProjectBook Database Design (PostgreSQL + MongoDB)

## 1. Scope

This document defines the production database model for ProjectBook based on:

- API contracts in [API-GUIDELINES.md](API-GUIDELINES.md)
- Schema contracts in [openapi.yaml](openapi.yaml)
- Shared app types in [src/app.d.ts](src/app.d.ts)
- Canonical RBAC mask specification in [rbac.md](rbac.md)
- Default role mask seed data in [src/lib/server/data/default-role-permissions.ts](src/lib/server/data/default-role-permissions.ts)
- Runtime mutation contracts in [src/lib/remote](src/lib/remote)
- Auth/session model in [src/lib/server/auth](src/lib/server/auth)

The design covers auth, workspace/project tenancy, RBAC, all artifact types, links, activity, and notifications.

## 2. Architecture Decision

Final decision: use a hybrid model (PostgreSQL + MongoDB), not SQL-only.

| Option | Pros | Cons | Decision |
| --- | --- | --- | --- |
| SQL-only | Strong joins, strict constraints, easy transactions | Artifact detail payloads are sparse and evolve frequently, leading to wide nullable tables and costly migrations | Rejected |
| SQL + NoSQL | SQL keeps relational integrity and query speed; MongoDB stores sparse, evolving artifact documents cleanly | Requires sync strategy between SQL metadata and MongoDB documents | Selected |

Reason selected: artifact details in stories/journeys/problems/ideas/tasks/feedback/pages/resources are highly optional and shape-flexible. SQL should store indexed metadata and relationships; MongoDB should store rich document content.

## 3. Data Ownership Model

- PostgreSQL is the source of truth for identity, membership, permissions, relationships, statuses, and list views.
- MongoDB is the source of truth for rich artifact content and versioned document payloads.
- SQL artifact metadata rows include a document pointer and revision fields:
  - `document_id` (nullable)
  - `document_revision` (NOT NULL)
  - `content_hash` (nullable)

## 4. PostgreSQL Schema

### 4.1 Enum Types

Create these enums first:

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

### 4.2 Identity and Auth

#### `users`

| Column | Type | Nullability | Default | Notes |
| --- | --- | --- | --- | --- |
| id | uuid | NOT NULL | gen_random_uuid() | Primary key |
| email | citext | NOT NULL | - | Unique |
| name | text | NOT NULL | - | Display name |
| password_hash | text | NOT NULL | - | Hashed password |
| is_email_verified | boolean | NOT NULL | false | Verification flag |
| last_login_at | timestamptz | NULL | - | Optional |
| created_at | timestamptz | NOT NULL | now() | Audit |
| updated_at | timestamptz | NOT NULL | now() | Audit |

Constraints:

- `PRIMARY KEY (id)`
- `UNIQUE (email)`

#### `auth_sessions`

| Column | Type | Nullability | Default | Notes |
| --- | --- | --- | --- | --- |
| id | uuid | NOT NULL | gen_random_uuid() | Primary key |
| user_id | uuid | NOT NULL | - | FK to users |
| token_hash | text | NOT NULL | - | Unique token hash |
| expires_at | timestamptz | NOT NULL | - | TTL boundary |
| revoked_at | timestamptz | NULL | - | Null when active |
| created_at | timestamptz | NOT NULL | now() | Created |
| user_agent | text | NULL | - | Optional metadata |
| ip_address | inet | NULL | - | Optional metadata |
| device_label | text | NULL | - | Optional metadata |

Constraints:

- `FOREIGN KEY (user_id) REFERENCES users(id)`
- `UNIQUE (token_hash)`

#### `email_verification_tokens`

| Column | Type | Nullability | Default | Notes |
| --- | --- | --- | --- | --- |
| id | uuid | NOT NULL | gen_random_uuid() | Primary key |
| user_id | uuid | NOT NULL | - | FK to users |
| token_hash | text | NOT NULL | - | Unique |
| expires_at | timestamptz | NOT NULL | - | 24h TTL |
| used_at | timestamptz | NULL | - | Null when unused |
| created_at | timestamptz | NOT NULL | now() | Created |

#### `password_reset_tokens`

| Column | Type | Nullability | Default | Notes |
| --- | --- | --- | --- | --- |
| id | uuid | NOT NULL | gen_random_uuid() | Primary key |
| user_id | uuid | NOT NULL | - | FK to users |
| token_hash | text | NOT NULL | - | Unique |
| expires_at | timestamptz | NOT NULL | - | 1h TTL |
| used_at | timestamptz | NULL | - | Null when unused |
| created_at | timestamptz | NOT NULL | now() | Created |

#### `auth_email_log`

| Column | Type | Nullability | Default | Notes |
| --- | --- | --- | --- | --- |
| id | uuid | NOT NULL | gen_random_uuid() | Primary key |
| user_id | uuid | NULL | - | Nullable for unknown email cases |
| kind | text | NOT NULL | - | `verify` or `reset` |
| recipient_email | citext | NOT NULL | - | Destination |
| link | text | NOT NULL | - | Sent link |
| sent_at | timestamptz | NOT NULL | now() | Sent timestamp |

#### `account_settings`

| Column | Type | Nullability | Default | Notes |
| --- | --- | --- | --- | --- |
| user_id | uuid | NOT NULL | - | PK and FK to users |
| display_name | text | NOT NULL | - | Editable |
| bio | text | NULL | - | Optional |
| theme | text | NOT NULL | `System` | `Light`/`Dark`/`System` |
| density | text | NOT NULL | `Comfortable` | `Comfortable`/`Compact` |
| landing | text | NOT NULL | `Last Project` | Landing preference |
| time_format | text | NOT NULL | `24-hour` | `12-hour`/`24-hour` |
| in_app_notifications | boolean | NOT NULL | true | Preference |
| email_notifications | boolean | NOT NULL | true | Preference |
| updated_at | timestamptz | NOT NULL | now() | Updated |

### 4.3 Workspace, Project, Team, RBAC

#### `workspaces`

| Column | Type | Nullability | Default | Notes |
| --- | --- | --- | --- | --- |
| id | uuid | NOT NULL | gen_random_uuid() | Primary key |
| name | text | NOT NULL | - | Workspace label |
| organization_name | text | NOT NULL | - | Display org |
| owner_user_id | uuid | NOT NULL | - | FK to users |
| created_at | timestamptz | NOT NULL | now() | Audit |
| updated_at | timestamptz | NOT NULL | now() | Audit |

#### `workspace_members`

| Column | Type | Nullability | Default | Notes |
| --- | --- | --- | --- | --- |
| workspace_id | uuid | NOT NULL | - | FK to workspaces |
| user_id | uuid | NOT NULL | - | FK to users |
| role | text | NOT NULL | `Member` | Workspace-level role |
| joined_at | timestamptz | NOT NULL | now() | Join timestamp |

Constraints:

- `PRIMARY KEY (workspace_id, user_id)`

#### `projects`

| Column | Type | Nullability | Default | Notes |
| --- | --- | --- | --- | --- |
| id | uuid | NOT NULL | gen_random_uuid() | Primary key |
| workspace_id | uuid | NOT NULL | - | FK to workspaces |
| slug | text | NOT NULL | - | Route-safe id per workspace |
| name | text | NOT NULL | - | Project name |
| organization_name | text | NOT NULL | - | Snapshot for UI |
| icon | text | NOT NULL | - | Matches `ProjectIconKey` enum |
| description | text | NULL | - | Optional |
| status | project_status | NOT NULL | `Active` | Lifecycle |
| created_by_user_id | uuid | NOT NULL | - | FK to users |
| archived_at | timestamptz | NULL | - | Set when archived |
| last_updated_at | timestamptz | NOT NULL | now() | Updated on mutations |
| created_at | timestamptz | NOT NULL | now() | Audit |
| updated_at | timestamptz | NOT NULL | now() | Audit |

Constraints:

- `UNIQUE (workspace_id, slug)`
- `UNIQUE (workspace_id, name)`

#### `project_members`

| Column | Type | Nullability | Default | Notes |
| --- | --- | --- | --- | --- |
| id | uuid | NOT NULL | gen_random_uuid() | Primary key |
| project_id | uuid | NOT NULL | - | FK to projects |
| user_id | uuid | NOT NULL | - | FK to users |
| role | project_role | NOT NULL | `Member` | Effective role |
| permission_mask | bigint | NOT NULL | 0 | Effective member mask (uint64-compatible bitset stored in bigint) |
| is_custom | boolean | NOT NULL | false | Whether member mask overrides role default mask |
| status | member_status | NOT NULL | `Active` | `Active` or `Invited` |
| joined_at | date | NULL | - | Optional for invited users |
| updated_at | timestamptz | NOT NULL | now() | Audit |

Constraints:

- `UNIQUE (project_id, user_id)`
- `CHECK (permission_mask >= 0)`

RBAC mask rules (enforced by service layer and/or DB trigger):

- If `is_custom = false`, `project_members.permission_mask` must equal the role default mask for `(project_id, role)`.
- If `is_custom = true`, `project_members.permission_mask` may differ from the role default mask.

#### `project_invites`

| Column | Type | Nullability | Default | Notes |
| --- | --- | --- | --- | --- |
| id | uuid | NOT NULL | gen_random_uuid() | Primary key |
| project_id | uuid | NOT NULL | - | FK to projects |
| email | citext | NOT NULL | - | Invite target |
| assigned_role | project_role | NOT NULL | - | Usually `Member`/`Viewer`/`Limited Access` |
| invited_by_user_id | uuid | NOT NULL | - | FK to users |
| inviter_role | project_role | NOT NULL | - | Snapshot |
| status | invite_status | NOT NULL | `pending` | Invite lifecycle |
| sent_at | timestamptz | NOT NULL | now() | Sent timestamp |
| expires_at | timestamptz | NOT NULL | - | Expiration |
| accepted_at | timestamptz | NULL | - | Optional |

Constraints:

- partial unique index for pending invites per project/email

#### `project_settings`

| Column | Type | Nullability | Default | Notes |
| --- | --- | --- | --- | --- |
| project_id | uuid | NOT NULL | - | PK and FK to projects |
| project_name | text | NOT NULL | - | Mirrors project name |
| project_description | text | NULL | - | Optional |
| project_status | project_status | NOT NULL | `Active` | Mirrors project status |
| whiteboards_enabled | boolean | NOT NULL | true | Feature flag |
| advanced_databases_enabled | boolean | NOT NULL | true | Feature flag |
| calendar_manual_events_enabled | boolean | NOT NULL | true | Feature flag |
| resource_versioning_enabled | boolean | NOT NULL | true | Feature flag |
| feedback_aggregation_enabled | boolean | NOT NULL | true | Feature flag |
| notify_artifact_created | boolean | NOT NULL | true | Notification setting |
| notify_artifact_locked | boolean | NOT NULL | true | Notification setting |
| notify_feedback_added | boolean | NOT NULL | true | Notification setting |
| notify_resource_updated | boolean | NOT NULL | true | Notification setting |
| delivery_channel | text | NOT NULL | `In-app` | `In-app` or `Email` |
| updated_at | timestamptz | NOT NULL | now() | Updated |

#### `role_permissions`

This stores per-project role default masks (`rolePermissionMasks`).

| Column | Type | Nullability | Default | Notes |
| --- | --- | --- | --- | --- |
| project_id | uuid | NOT NULL | - | FK to projects |
| role | project_role | NOT NULL | - | Role |
| permission_mask | bigint | NOT NULL | 0 | Role default mask for this project |
| updated_by_user_id | uuid | NULL | - | Optional audit actor |
| updated_at | timestamptz | NOT NULL | now() | Updated |

Constraints:

- `PRIMARY KEY (project_id, role)`
- `CHECK (permission_mask >= 0)`

Notes:

- This replaces legacy per-domain/per-action boolean RBAC storage.
- Owner role mask remains immutable at the API/service layer.
- Role rows are seeded for each project from canonical defaults and then managed through role-mask updates.

#### RBAC Model (Mask-Based)

- Role defines default permission mask.
- Member may override with custom mask (`is_custom = true`).
- `permission_mask` is the single source of truth for authorization.

Mask allocation and helpers are defined in [rbac.md](rbac.md):

- Bit strategy: `bit = domain_index * 6 + action_index`
- Domain/action mapping is append-only and shared by frontend/backend
- Helper functions include `hasPerm(...)` and `updatePerm(...)`
- Validation rule: non-view actions require view in the same domain

### 4.4 Artifact Metadata Tables (SQL)

#### `stories`

| Column | Type | Nullability | Default | Notes |
| --- | --- | --- | --- | --- |
| id | uuid | NOT NULL | gen_random_uuid() | Primary key |
| project_id | uuid | NOT NULL | - | FK |
| slug | text | NOT NULL | - | Unique per project |
| title | text | NOT NULL | - | Required |
| persona_name | text | NULL | - | Optional |
| pain_points_count | integer | NOT NULL | 0 | Derived |
| problem_hypotheses_count | integer | NOT NULL | 0 | Derived |
| owner_user_id | uuid | NOT NULL | - | FK users |
| owner_name_cache | text | NOT NULL | - | Denormalized |
| status | story_status | NOT NULL | `Draft` | Lifecycle |
| is_orphan | boolean | NOT NULL | true | Derived |
| document_id | text | NULL | - | Mongo ref |
| document_revision | integer | NOT NULL | 1 | Sync revision |
| content_hash | text | NULL | - | Optional hash |
| created_at | timestamptz | NOT NULL | now() | Audit |
| updated_at | timestamptz | NOT NULL | now() | Audit |

#### `journeys`

| Column | Type | Nullability | Default | Notes |
| --- | --- | --- | --- | --- |
| id | uuid | NOT NULL | gen_random_uuid() | Primary key |
| project_id | uuid | NOT NULL | - | FK |
| slug | text | NOT NULL | - | Unique per project |
| title | text | NOT NULL | - | Required |
| primary_persona_name | text | NULL | - | Optional |
| stages_count | integer | NOT NULL | 0 | Derived |
| pain_points_count | integer | NOT NULL | 0 | Derived |
| owner_user_id | uuid | NOT NULL | - | FK users |
| owner_name_cache | text | NOT NULL | - | Denormalized |
| status | journey_status | NOT NULL | `Draft` | Lifecycle |
| is_orphan | boolean | NOT NULL | true | Derived |
| document_id | text | NULL | - | Mongo ref |
| document_revision | integer | NOT NULL | 1 | Sync revision |
| content_hash | text | NULL | - | Optional hash |
| created_at | timestamptz | NOT NULL | now() | Audit |
| updated_at | timestamptz | NOT NULL | now() | Audit |

#### `problems`

| Column | Type | Nullability | Default | Notes |
| --- | --- | --- | --- | --- |
| id | uuid | NOT NULL | gen_random_uuid() | Primary key |
| project_id | uuid | NOT NULL | - | FK |
| slug | text | NOT NULL | - | Unique per project |
| statement | text | NOT NULL | - | Required |
| pain_points_count | integer | NOT NULL | 0 | Derived |
| ideas_count | integer | NOT NULL | 0 | Derived |
| owner_user_id | uuid | NOT NULL | - | FK users |
| owner_name_cache | text | NOT NULL | - | Denormalized |
| status | problem_status | NOT NULL | `Draft` | Lifecycle |
| is_orphan | boolean | NOT NULL | true | Derived |
| document_id | text | NULL | - | Mongo ref |
| document_revision | integer | NOT NULL | 1 | Sync revision |
| content_hash | text | NULL | - | Optional hash |
| created_at | timestamptz | NOT NULL | now() | Audit |
| updated_at | timestamptz | NOT NULL | now() | Audit |

#### `ideas`

| Column | Type | Nullability | Default | Notes |
| --- | --- | --- | --- | --- |
| id | uuid | NOT NULL | gen_random_uuid() | Primary key |
| project_id | uuid | NOT NULL | - | FK |
| slug | text | NOT NULL | - | Unique per project |
| title | text | NOT NULL | - | Required |
| linked_problem_id | uuid | NULL | - | Optional FK to problems |
| linked_problem_statement_cache | text | NULL | - | Snapshot text |
| persona_cache | text | NULL | - | Optional |
| tasks_count | integer | NOT NULL | 0 | Derived |
| owner_user_id | uuid | NOT NULL | - | FK users |
| owner_name_cache | text | NOT NULL | - | Denormalized |
| status | idea_status | NOT NULL | `Considered` | Lifecycle |
| linked_problem_locked | boolean | NOT NULL | false | Cached constraint |
| is_orphan | boolean | NOT NULL | true | Derived |
| document_id | text | NULL | - | Mongo ref |
| document_revision | integer | NOT NULL | 1 | Sync revision |
| content_hash | text | NULL | - | Optional hash |
| created_at | timestamptz | NOT NULL | now() | Audit |
| updated_at | timestamptz | NOT NULL | now() | Audit |

#### `tasks`

| Column | Type | Nullability | Default | Notes |
| --- | --- | --- | --- | --- |
| id | uuid | NOT NULL | gen_random_uuid() | Primary key |
| project_id | uuid | NOT NULL | - | FK |
| slug | text | NOT NULL | - | Unique per project |
| title | text | NOT NULL | - | Required |
| linked_idea_id | uuid | NULL | - | Optional FK to ideas |
| linked_problem_id | uuid | NULL | - | Optional FK to problems |
| linked_idea_cache | text | NULL | - | Snapshot text |
| linked_problem_statement_cache | text | NULL | - | Snapshot text |
| persona_cache | text | NULL | - | Optional |
| assignee_user_id | uuid | NULL | - | Optional FK to users |
| owner_user_id | uuid | NOT NULL | - | FK users |
| owner_name_cache | text | NOT NULL | - | Denormalized |
| deadline | date | NULL | - | Optional |
| status | task_status | NOT NULL | `Planned` | Lifecycle |
| idea_rejected | boolean | NOT NULL | false | Cached flag |
| has_feedback | boolean | NOT NULL | false | Cached flag |
| is_orphan | boolean | NOT NULL | true | Derived |
| document_id | text | NULL | - | Mongo ref |
| document_revision | integer | NOT NULL | 1 | Sync revision |
| content_hash | text | NULL | - | Optional hash |
| created_at | timestamptz | NOT NULL | now() | Audit |
| updated_at | timestamptz | NOT NULL | now() | Audit |

#### `feedback`

| Column | Type | Nullability | Default | Notes |
| --- | --- | --- | --- | --- |
| id | uuid | NOT NULL | gen_random_uuid() | Primary key |
| project_id | uuid | NOT NULL | - | FK |
| slug | text | NOT NULL | - | Unique per project |
| title | text | NOT NULL | - | Required |
| outcome | feedback_outcome | NOT NULL | `Needs Iteration` | Outcome |
| primary_linked_artifact_type | artifact_type | NULL | - | Optional primary link |
| primary_linked_artifact_id | uuid | NULL | - | Optional primary link |
| linked_task_or_idea_cache | text | NULL | - | Snapshot text |
| owner_user_id | uuid | NOT NULL | - | FK users |
| owner_name_cache | text | NOT NULL | - | Denormalized |
| created_date | date | NOT NULL | current_date | Contract field |
| has_task_link | boolean | NOT NULL | false | Cached flag |
| is_archived | boolean | NOT NULL | false | Archive support |
| is_orphan | boolean | NOT NULL | true | Derived |
| document_id | text | NULL | - | Mongo ref |
| document_revision | integer | NOT NULL | 1 | Sync revision |
| content_hash | text | NULL | - | Optional hash |
| created_at | timestamptz | NOT NULL | now() | Audit |
| updated_at | timestamptz | NOT NULL | now() | Audit |

#### `resources`

| Column | Type | Nullability | Default | Notes |
| --- | --- | --- | --- | --- |
| id | uuid | NOT NULL | gen_random_uuid() | Primary key |
| project_id | uuid | NOT NULL | - | FK |
| slug | text | NOT NULL | - | Unique per project |
| name | text | NOT NULL | - | Required |
| file_type | text | NOT NULL | - | Required |
| doc_type | text | NOT NULL | - | Required |
| owner_user_id | uuid | NOT NULL | - | FK users |
| owner_name_cache | text | NOT NULL | - | Denormalized |
| current_version | text | NOT NULL | `v1` | Current version label |
| linked_count | integer | NOT NULL | 0 | Derived |
| status | resource_status | NOT NULL | `Active` | Lifecycle |
| description | text | NULL | - | Optional |
| file_size_bytes | bigint | NULL | - | Optional |
| created_at | timestamptz | NOT NULL | now() | Audit |
| updated_at | timestamptz | NOT NULL | now() | Audit |

#### `resource_versions`

| Column | Type | Nullability | Default | Notes |
| --- | --- | --- | --- | --- |
| id | uuid | NOT NULL | gen_random_uuid() | Primary key |
| resource_id | uuid | NOT NULL | - | FK to resources |
| version | text | NOT NULL | - | Version name |
| uploaded_by_user_id | uuid | NOT NULL | - | FK users |
| uploaded_by_name_cache | text | NOT NULL | - | Denormalized |
| upload_date | timestamptz | NOT NULL | now() | Upload timestamp |
| description | text | NULL | - | Optional |
| storage_document_id | text | NOT NULL | - | Mongo document pointer |
| content_hash | text | NULL | - | Optional hash |
| file_size_bytes | bigint | NULL | - | Optional |
| mime_type | text | NULL | - | Optional |

Constraints:

- `UNIQUE (resource_id, version)`

#### `pages`

| Column | Type | Nullability | Default | Notes |
| --- | --- | --- | --- | --- |
| id | uuid | NOT NULL | gen_random_uuid() | Primary key |
| project_id | uuid | NOT NULL | - | FK |
| slug | text | NOT NULL | - | Unique per project |
| title | text | NOT NULL | - | Required |
| owner_user_id | uuid | NOT NULL | - | FK users |
| owner_name_cache | text | NOT NULL | - | Denormalized |
| linked_artifacts_count | integer | NOT NULL | 0 | Derived |
| status | page_status | NOT NULL | `Draft` | Lifecycle |
| is_orphan | boolean | NOT NULL | true | Derived |
| document_id | text | NULL | - | Mongo ref |
| document_revision | integer | NOT NULL | 1 | Sync revision |
| content_hash | text | NULL | - | Optional hash |
| created_at | timestamptz | NOT NULL | now() | Audit |
| updated_at | timestamptz | NOT NULL | now() | Audit |

#### `calendar_events`

| Column | Type | Nullability | Default | Notes |
| --- | --- | --- | --- | --- |
| id | uuid | NOT NULL | gen_random_uuid() | Primary key |
| project_id | uuid | NOT NULL | - | FK |
| event_key | text | NOT NULL | - | Public event id (`evt-*`) |
| title | text | NOT NULL | - | Required |
| type | calendar_event_type | NOT NULL | `Manual` | Derived or Manual |
| start_date | date | NOT NULL | - | Required |
| end_date | date | NOT NULL | - | Required |
| start_time | time | NULL | - | Optional when not all-day |
| end_time | time | NULL | - | Optional when not all-day |
| all_day | boolean | NOT NULL | false | Required |
| owner_user_id | uuid | NULL | - | Optional user FK |
| owner_name_cache | text | NOT NULL | - | Display owner |
| phase | calendar_phase | NOT NULL | `None` | Required |
| artifact_type | calendar_artifact_type | NOT NULL | `Manual` | Required |
| source_artifact_type | artifact_type | NULL | - | For derived events |
| source_artifact_id | uuid | NULL | - | For derived events |
| source_title_cache | text | NULL | - | Optional |
| description | text | NULL | - | Optional |
| location | text | NULL | - | Optional |
| event_kind | text | NULL | - | Optional |
| created_at | timestamptz | NOT NULL | now() | Required |
| last_edited_at | timestamptz | NULL | - | Optional |

Constraints:

- `UNIQUE (project_id, event_key)`
- Derived events should be made read-only at service level.

#### `artifact_links`

Unified link graph across artifacts.

| Column | Type | Nullability | Default | Notes |
| --- | --- | --- | --- | --- |
| id | uuid | NOT NULL | gen_random_uuid() | Primary key |
| project_id | uuid | NOT NULL | - | FK to projects |
| source_type | artifact_type | NOT NULL | - | Source domain |
| source_id | uuid | NOT NULL | - | Source artifact id |
| target_type | artifact_type | NOT NULL | - | Target domain |
| target_id | uuid | NOT NULL | - | Target artifact id |
| link_kind | text | NOT NULL | `reference` | e.g. `reference`, `derived`, `primary` |
| created_by_user_id | uuid | NULL | - | Optional actor |
| created_at | timestamptz | NOT NULL | now() | Created |

Constraints:

- `UNIQUE (project_id, source_type, source_id, target_type, target_id, link_kind)`

### 4.5 Activity and Notifications

#### `activity_log`

| Column | Type | Nullability | Default | Notes |
| --- | --- | --- | --- | --- |
| id | bigserial | NOT NULL | - | Primary key |
| workspace_id | uuid | NOT NULL | - | Workspace scope |
| project_id | uuid | NULL | - | Null for workspace-only events |
| actor_user_id | uuid | NULL | - | Optional actor FK |
| actor_name_cache | text | NOT NULL | - | Display name |
| actor_initials | text | NOT NULL | - | 2-char initials |
| action | text | NOT NULL | - | Event action |
| artifact_type | artifact_type | NULL | - | Optional artifact type |
| artifact_id | uuid | NULL | - | Optional artifact id |
| artifact_name_cache | text | NULL | - | Optional artifact title |
| href | text | NULL | - | Optional deep link |
| event_type | text | NULL | - | `Artifacts`/`Tasks`/`Feedback`/`Comments` |
| occurred_at | timestamptz | NOT NULL | now() | Timeline sort key |

#### `notifications`

| Column | Type | Nullability | Default | Notes |
| --- | --- | --- | --- | --- |
| id | uuid | NOT NULL | gen_random_uuid() | Primary key |
| workspace_id | uuid | NOT NULL | - | FK to workspace |
| user_id | uuid | NOT NULL | - | Recipient |
| project_id | uuid | NULL | - | Optional project |
| source_type | notification_source_type | NULL | - | Optional source |
| title | text | NULL | - | Optional |
| text | text | NULL | - | Optional |
| description | text | NULL | - | Optional |
| url | text | NULL | - | Optional deep link |
| read | boolean | NOT NULL | false | Read state |
| dismissed | boolean | NOT NULL | false | Dismiss state |
| created_at | timestamptz | NOT NULL | now() | Created |

### 4.6 SQL and Mongo Sync Reliability

#### `document_sync_outbox`

Use an outbox table so SQL commit and Mongo update are reliably coordinated.

| Column | Type | Nullability | Default | Notes |
| --- | --- | --- | --- | --- |
| id | bigserial | NOT NULL | - | Primary key |
| project_id | uuid | NOT NULL | - | Scope |
| artifact_type | artifact_type | NOT NULL | - | Artifact type |
| artifact_id | uuid | NOT NULL | - | Artifact id |
| target_collection | text | NOT NULL | - | Mongo collection |
| intended_revision | integer | NOT NULL | - | Revision to write |
| payload | jsonb | NOT NULL | - | Document payload |
| status | text | NOT NULL | `pending` | `pending`/`processing`/`synced`/`failed` |
| attempt_count | integer | NOT NULL | 0 | Retry count |
| next_attempt_at | timestamptz | NULL | - | Backoff |
| last_error | text | NULL | - | Last failure |
| created_at | timestamptz | NOT NULL | now() | Created |
| updated_at | timestamptz | NOT NULL | now() | Updated |

## 5. MongoDB Collections

All artifact document collections should include common required metadata.

### 5.1 Common Required Fields (all collections)

| Field | Required | Type | Notes |
| --- | --- | --- | --- |
| _id | Yes | ObjectId | Mongo primary key |
| artifact_id | Yes | UUID string | SQL artifact id |
| project_id | Yes | UUID string | Tenant/project scope |
| schema_version | Yes | int | Document schema version |
| revision | Yes | int | Monotonic revision |
| updated_at | Yes | ISO date | Last write |
| updated_by_user_id | Yes | UUID string | Editor |
| content_hash | No | string | Optional hash |
| content | Yes | object | Flexible payload |

### 5.2 Collections and Content Shape

#### `story_documents`

- Optional `content` keys: `description`, `persona`, `context`, `empathyMap`, `painPoints`, `hypothesis`, `notes`, `activeModules`, `moduleContent`

#### `journey_documents`

- Optional `content` keys: `description`, `persona`, `context`, `stages`, `notes`

#### `problem_documents`

- Optional `content` keys: `title`, `finalStatement`, `selectedPainPoints`, `linkedSources`, `activeModules`, `moduleContent`, `notes`

#### `idea_documents`

- Optional `content` keys: `description`, `summary`, `notes`, `selectedProblemId`, `activeModules`, `moduleContent`

#### `task_documents`

- Optional `content` keys: `assignedToId`, `selectedIdeaId`, `deadline`, `hypothesis`, `planItems`, `executionLinks`, `notes`, `activeModules`, `abandonReason`

#### `feedback_documents`

- Optional `content` keys: `observation`, `interpretation`, `linkedArtifacts`, `activeModules`, `moduleContent`, `evidenceText`, `nextStepsText`, `notes`

#### `page_documents`

- Optional `content` keys: `description`, `tags`, `linkedArtifacts`, `docHeading`, `docBody`, `views`, `activeViewId`, `tableColumns`, `tableRows`, `databaseItems`

#### `resource_version_documents`

- Required keys in `content`: version metadata and object-storage pointer
- Suggested structure: `storageProvider`, `storageKey`, `mimeType`, `sizeBytes`, `description`, `previewText`, `extractedMetadata`

## 6. Relationships and Cardinality

- one `workspace` to many `projects`
- one `project` to many `project_members`
- one `user` to many `project_members` (many-to-many users/projects)
- one `project` to many artifacts (`stories`, `journeys`, `problems`, `ideas`, `tasks`, `feedback`, `resources`, `pages`, `calendar_events`)
- `problems` one-to-many `ideas` (optional from idea side)
- `ideas` one-to-many `tasks` (optional from task side)
- `resources` one-to-many `resource_versions`
- artifacts many-to-many through `artifact_links`
- one `workspace` to many `activity_log` entries and `notifications`

## 7. Index Plan

### 7.1 PostgreSQL

Create these high-value indexes:

- `users(email)` unique
- `auth_sessions(token_hash)` unique
- `auth_sessions(user_id, revoked_at, expires_at desc)`
- `projects(workspace_id, slug)` unique
- `projects(workspace_id, status, last_updated_at desc)`
- `project_members(project_id, user_id)` unique
- `project_members(user_id, project_id)`
- `project_invites(project_id, email)` partial unique where `status = 'pending'`
- `role_permissions(project_id, role)` primary key
- `project_members(project_id, role, is_custom)`
- for each artifact metadata table:
  - unique `(project_id, slug)`
  - index `(project_id, status, updated_at desc)`
- `tasks(project_id, status, deadline)`
- `artifact_links(project_id, source_type, source_id)`
- `artifact_links(project_id, target_type, target_id)`
- `activity_log(project_id, occurred_at desc)`
- `activity_log(workspace_id, occurred_at desc)`
- `notifications(user_id, read, created_at desc)`
- `document_sync_outbox(status, next_attempt_at)`

### 7.2 MongoDB

For every artifact collection:

- unique index on `{ artifact_id: 1 }`
- index on `{ project_id: 1, updated_at: -1 }`
- index on `{ project_id: 1, revision: -1 }`

For `resource_version_documents`:

- unique index on `{ resource_version_id: 1 }`
- index on `{ resource_id: 1, revision: -1 }`

## 8. Lifecycle and Integrity Rules

- Enforce status transition rules at service layer according to [API-GUIDELINES.md](API-GUIDELINES.md) section 7.
- Enforce RBAC checks before any mutation.
- Derived calendar events (`type = Derived`) are read-only at API/service layer.
- Owner role cannot be reassigned through role update APIs.
- Owner role mask should remain immutable.
- Reject invalid permission masks on write. Non-view actions require view in the same domain.
- Frontend may auto-correct mask toggles for UX, but backend must validate and reject invalid masks.
- Archive should be logical (status/is_archived), not hard delete for core artifacts.

## 9. Migration Strategy from In-Memory Store

1. Create all SQL enums and tables.
2. Migrate identity/auth datasets first.
3. Migrate workspace/project/member/RBAC datasets.
4. Migrate artifact row metadata into SQL tables.
5. Migrate rich artifact details into MongoDB collections.
6. Populate `document_id` and `document_revision` back into SQL rows.
7. Build `artifact_links` from existing string-based linked arrays.
8. Backfill `is_orphan` and aggregate counts.
9. Turn on outbox-driven SQL to Mongo sync for updates.
10. Run consistency checks (counts, orphans, revision parity).

## 10. Final Recommendation

Use PostgreSQL for relational correctness and query-heavy workloads, and MongoDB for optional/sparse artifact content. This directly matches ProjectBook's domain behavior where list metadata is structured but editor payloads are highly flexible.
