# Changelog

All notable changes to this template are documented in this file.

## v0.7.11 (2026-04-18)

### Breaking Changes
- Project navigation read endpoint moved from `GET /api/v1/projects/{projectId}/sidebar` to `GET /api/v1/projects/{projectId}/navigation`.
- Removed dedicated sidebar mutation endpoints:
	- `POST /api/v1/projects/{projectId}/sidebar/artifacts`
	- `PUT /api/v1/projects/{projectId}/sidebar/artifacts/{artifactId}/rename`
	- `DELETE /api/v1/projects/{projectId}/sidebar/artifacts/{artifactId}`

### Changed
- Project module handler/service/repository contracts now use `Navigation` semantics and return the `current_project` + `project_list` payload shape.
- Cache tag naming migrated from `project.sidebar` to `project.navigation` for project navigation reads and dependent invalidation paths.

### Removed
- Deprecated `internal/modules/sidebar` module implementation and tests.

### Documentation
- Updated endpoint snapshots (`endpoint-tracker.md` / `.json` / `.csv`) to reflect `EP-023` as `/navigation` and removed deprecated sidebar endpoint entries.
- Updated ProjectBookDocs API snapshot and frontend handoff/error-map artifacts to remove sidebar endpoint remnants and align sidebar interactions with artifact module endpoints.
- Updated smoke and route coverage snapshots to validate navigation endpoint behavior.

## v0.7.10 (2026-04-15)

### Added
- Added release-managed typed tuning block in runtime config (`Config.Tuning`) with:
	- centralized auth route rate-limit profile values.
	- centralized cache TTL profile defaults.
- Added startup fallback warnings for sensitive configuration defaults via `config.SensitiveFallbackWarnings(...)` and bootstrap logging in `cmd/api/main.go`.
- Added CI guardrail tests in `internal/core/config/config_test.go` covering insecure production profile defaults.

### Changed
- Auth route limiter defaults in `internal/modules/auth/routes.go` now consume the typed tuning block instead of per-route literals.
- Module runtime/dependency wiring now carries tuning snapshot values (`internal/core/app/deps.go`, `internal/core/modulekit/runtime.go`).
- System permission-context token signing now surfaces explicit errors when production secret policy is violated.

### Security
- Enforced production lint requirement for `PROJECTBOOK_PERMISSION_CONTEXT_SECRET` with non-default 32+ char value.
- Enforced production lint rejection for localhost `WEB_APP_BASE_URL` values.
- Enforced strong `METRICS_AUTH_TOKEN` policy in production/profile-prod (non-placeholder, minimum length).
- Rejected production fallback behavior for permission-context signing secret in system token generation path.

## v0.7.9 (2026-04-14)

### Added
- System session-context response now includes backend-issued signed permission-context token fields:
	- `context_token`
	- `context_token_expires_utc`
	- `context_token_expires_unix`
	- `context_token_version`
- Added session-context token construction and signing helpers in `internal/modules/system/routes.go`.
- Added system token tests in `internal/modules/system/routes_test.go` for claim payload and secret resolution behavior.

### Changed
- Auth refresh route rate-limit keying now uses refresh-token hash with IP fallback (`internal/modules/auth/routes.go`) to reduce cross-user coupling under shared IPs.
- Added refresh keyer coverage in `internal/modules/auth/routes_test.go`.

### Security
- Introduced `PROJECTBOOK_PERMISSION_CONTEXT_SECRET` for backend-issued permission-context token signing.
- Updated auth and environment documentation for shared frontend/backend verification secret requirements.

## v0.7.8 (2026-04-11)

### Added
- New Resources module at `internal/modules/resources` with strict layering:
	- `dto.go`
	- `repo.go`
	- `service.go`
	- `handler.go`
	- `routes.go`
	- `module.go`
- New Pages module at `internal/modules/pages` with strict layering:
	- `dto.go`
	- `repo.go`
	- `service.go`
	- `handler.go`
	- `routes.go`
	- `module.go`
- New Calendar module at `internal/modules/calendar` with strict layering:
	- `dto.go`
	- `repo.go`
	- `service.go`
	- `handler.go`
	- `routes.go`
	- `module.go`
- New Sidebar module at `internal/modules/sidebar` with strict layering:
	- `dto.go`
	- `repo.go`
	- `service.go`
	- `handler.go`
	- `routes.go`
	- `module.go`
- New Activity module at `internal/modules/activity` with strict layering:
	- `dto.go`
	- `repo.go`
	- `service.go`
	- `handler.go`
	- `routes.go`
	- `module.go`
- Implemented EP-064 through EP-082 under `/api/v1/projects/{projectId}/*`:
	- resources list/create/detail/update/status flows.
	- pages list/create/detail/update/rename flows.
	- calendar list/create/detail/update/delete flows with derived-event write guards.
	- sidebar artifact create/rename/delete convenience flows with prefix-based delegation.
	- project activity read endpoint.
- Added migration pair for resources/pages/calendar contract surfaces:
	- `db/migrations/000011_resources_pages_calendar_contract.up.sql`
	- `db/migrations/000011_resources_pages_calendar_contract.down.sql`

### Changed
- Module registry now includes:
	- `resources.New()`
	- `pages.New()`
	- `calendar.New()`
	- `sidebar.New()`
	- `activity.New()`
- Mongo bootstrap now provisions `resource_documents` collection/indexes.
- Document sync processor now routes `resource` artifacts to `resource_documents`.

### Documentation
- Endpoint tracker artifacts (`md`/`json`/`csv`) now mark EP-064 through EP-082 as `tested` and point implementation ownership to their module paths.

## v0.7.7 (2026-04-11)

### Added
- New Artifacts module at `internal/modules/artifacts` with strict layering:
	- `dto.go`
	- `repo.go`
	- `service.go`
	- `handler.go`
	- `routes.go`
	- `module.go`
- Implemented EP-035 through EP-063 under `/api/v1/projects/{projectId}/*`:
	- stories, journeys, problems, ideas, tasks, and feedback list/create/detail/update flows.
	- problem lock and status transition endpoint.
	- idea select endpoint and status transition endpoint.
	- task status transition endpoint with rejected-idea start guard.
- Added migration pair for artifact primary-chain/orphan semantics:
	- `db/migrations/000010_artifact_chain_links.up.sql`
	- `db/migrations/000010_artifact_chain_links.down.sql`
- Added document sync outbox processor:
	- `internal/infrastructure/docsync/processor.go`
	- app wiring in `internal/core/app/deps.go` and lifecycle run loop in `internal/core/app/app.go`.

### Changed
- Module registry now includes `artifacts.New()` in `internal/modules/modules.go`.
- Artifact chain behavior now enforces non-cascading deletes and orphan recalculation across story/journey/problem/idea/task/feedback links.

### Documentation
- Endpoint tracker artifacts (`md`/`json`/`csv`) now mark EP-035 through EP-063 as `tested` and point implementation ownership to `internal/modules/artifacts`.

## v0.7.6 (2026-04-11)

### Added
- New Team module at `internal/modules/team` with strict layering:
	- `dto.go`
	- `repo.go`
	- `service.go`
	- `handler.go`
	- `routes.go`
	- `module.go`
- Implemented EP-028 through EP-034 under `/api/v1/projects/{projectId}/team/*`:
	- members list, roles list, single invite create, batch invite create (partial success with `207`), invite cancel, member permission update, role permission update.
- Added migration pair for team member list/read hot paths:
	- `db/migrations/000009_team_member_indexes.up.sql`
	- `db/migrations/000009_team_member_indexes.down.sql`

### Changed
- Module registry now includes `team.New()` in `internal/modules/modules.go`.

### Documentation
- Endpoint tracker artifacts (`md`/`json`/`csv`) now mark EP-028 through EP-034 as `tested`.

## v0.7.5 (2026-04-11)

### Added
- New Project module at `internal/modules/project` with strict layering:
	- `dto.go`
	- `repo.go`
	- `service.go`
	- `handler.go`
	- `routes.go`
	- `module.go`
- Implemented EP-021 through EP-027 under `/api/v1/projects/{projectId}*`:
	- dashboard, access, sidebar, settings (get/update), archive, delete.
- Added migration pair for project sidebar/read hot paths:
	- `db/migrations/000008_project_sidebar_indexes.up.sql`
	- `db/migrations/000008_project_sidebar_indexes.down.sql`

### Changed
- Permission resolver relational fallback now accepts slug-or-UUID project scope inputs and resolves canonical project identity.
- Resolve-permissions policy now stores canonical resolver project ID in auth context when available.
- Module registry now includes `project.New()` in `internal/modules/modules.go`.

### Documentation
- Endpoint tracker artifacts (`md`/`json`/`csv`) now mark EP-021 through EP-027 as `tested`.

## v0.7.4 (2026-04-11)

### Added
- New Home module at `internal/modules/home` with strict layering:
	- `dto.go`
	- `repo.go`
	- `service.go`
	- `handler.go`
	- `routes.go`
	- `module.go`
- Implemented EP-008 through EP-020 under `/api/v1/home/*`:
	- dashboard/projects/create-project/reference/invites/notifications/activity/dashboard-activity/account/docs.

### Changed
- Module registry now includes `home.New()` in `internal/modules/modules.go`.
- Home write workflows use service-owned transaction boundaries with repository-only data access.
- Home routes apply authenticated cache variation by `UserID` and write-path cache invalidation for related home tags.

### Documentation
- Endpoint tracker artifacts (`md`/`json`/`csv`) now mark EP-008 through EP-020 as `tested`.
- Home implementation follows contract-first verification against API guidelines and relational schema surfaces before route wiring.

## v0.7.3 (2026-04-11)

### Breaking Changes
- Auth HTTP routes were migrated from system paths to module-owned auth paths.
	- Removed `POST /api/v1/system/auth/login`.
	- Removed `POST /api/v1/system/auth/refresh`.
	- Added module-owned routes under `/api/v1/auth/*`.

### Added
- New auth module at `internal/modules/auth` with strict layering:
	- `dto.go`
	- `handler.go`
	- `service.go`
	- `repo.go`
	- `routes.go`
- Implemented EP-001 through EP-007 in the auth module:
	- signup, login, logout, verify-email, resend-verification, forgot-password, reset-password.
- Added compatibility refresh endpoint `POST /api/v1/auth/refresh` for performance tooling continuity.

### Changed
- API envelope serialization migrated from `ok` to `success` in core response flow.
- Updated typed/timeout/cache test assertions and generator expectations to `success`.
- goAuth provider config now enables account creation, password reset, and email verification flows.
- Signup flow now preserves API name semantics via post-create repository name update.
- Module registration now includes auth module in `internal/modules/modules.go`.

### Documentation
- Updated docs and performance assets that referenced `/api/v1/system/auth/*` to `/api/v1/auth/*`.
- Endpoint tracker artifacts (`md`/`json`/`csv`) now mark EP-001 through EP-007 as `tested`.
- Added contract-freeze evidence updates for the completed auth implementation wave.

## v0.7.2 (2026-04-10)

### Breaking Changes
- Protected route policy contract now requires project-scoped permission resolution before RBAC.
	- Required order: auth -> project -> resolve permissions -> rbac -> rate limit -> cache -> cache-control
	- RBAC policy stacks without `ResolvePermissions(...)` now fail validator checks.

### Added
- New permissions resolver package and policy stage.
	- Added `internal/core/permissions` with hybrid Redis-first + relational fallback resolution.
	- Added `policy.ResolvePermissions(...)` middleware with explicit status mapping:
		- unauthenticated -> `401`
		- missing membership / missing project scope -> `403`
		- inconsistent permission state -> `500`
		- dependency failure -> `503`
- Added permission-tag invalidation helpers for user/project-scoped cache freshness.
- Added permission lifecycle startup sync for role-mask consistency.
	- Seeds canonical role masks into `role_permissions` for all projects.
	- Resyncs non-custom member masks from role masks.
	- Invalidates resolver keys and permission cache tags for affected scopes.
- Added concrete Mongo document backend wiring.
	- Added Mongo client/database dependency initialization and readiness checks.
	- Added startup collection/index bootstrap verification.
	- Added concrete `MongoDocumentStore` for module document operations.

### Changed
- Project isolation behavior updated in auth policies.
	- `AuthRequired()` no longer seeds project scope from auth provider tenant semantics.
	- `ProjectRequired()` now requires explicit project path scope and rejects path/auth mismatches with `403`.
	- `ProjectMatchFromPath(...)` mismatch behavior changed to `403`.
- Added runtime/config wiring for permissions resolver in app dependencies and module runtime.
- Updated route validator, static analyzer, and `superapi-verify` hints for resolver-stage requirements.
- Simplified scaffold tooling by removing tenant options from:
	- `cmd/modulegen`
	- `cmd/authgen`
	- `make module` passthrough flags

### Documentation
- Rewrote `docs/policies.md` to the project-scoped resolver model.
- Updated policy/cache/architecture/module/auth guides to remove tenant-era route ordering guidance.
- Updated ProjectBook RBAC docs to the enforced resolver-aware protected route chain.

## v0.7.1 (2026-04-06)

### Fixed
- System auth demo routes were aligned with goAuth v0.3.0 error semantics.
	- Removed usage of deleted `goauth.ErrRefreshRateLimited`.
	- Added canonical auth error translation for login/refresh endpoints based on goAuth `AuthError` categories.
	- Mapped auth abuse/state/system categories to stable API responses (`429`/`403`/`503`/`500`) while preserving unauthorized defaults.

### Added
- New focused tests for system auth route error translation.
	- Added category-based mapping coverage for `AUTH_ABUSE`, `AUTH_STATE`, `SYSTEM_INTERNAL_ERROR`, and `SYSTEM_UNAVAILABLE`.
	- Added fallback coverage for legacy `goauth.ErrLoginRateLimited` sentinel matching.

### Changed
- goAuth dependency was upgraded to `v0.3.0`.
- Auth docs were migrated to the v0.3.0 model, including:
	- Canonical `AuthError` boundary and code registry documentation.
	- Updated limiter and config field naming (`EnableLoginFailureLimiter`, request/confirm limiter split fields, and creation limiter toggle).
	- Refresh-throttle removal guidance and v0.3.0 migration notes.

## v0.7.0 (2026-04-05)

### Breaking Changes
- Enforced store-first data-layer architecture across runtime wiring and module guidance.
	- Required flow: Service -> Repository -> Store -> Backend
- Removed legacy core DB helper APIs from `internal/core/db`.
	- Removed `db.NewQueries(...)`
	- Removed `db.QueriesFrom(...)`
	- Removed `db.QueriesFromTx(...)`
	- Removed `db.WithTx(...)`
	- Removed `db.WithTxResult(...)`
- `modulekit.Runtime` storage access surface changed.
	- Removed `Runtime.Postgres()` accessor
	- Added `Runtime.Store()`, `Runtime.RelationalStore()`, `Runtime.DocumentStore()`
- goAuth provider constructor and wiring path changed.
	- Replaced `auth.NewSQLCUserProvider(...)` with `auth.NewStoreUserProvider(...)`
	- Auth persistence now goes through repository + store contracts

### Added
- New storage contracts package at `internal/core/storage`.
	- Backend kind contract (`Store.Kind()`)
	- Mandatory transaction contract (`TransactionalStore.WithTx(...)`)
	- Relational/document operation execution contracts
- New relational store implementation over pgx.
	- `storage.PostgresRelationalStore`
- New document contract placeholder implementation.
	- `storage.NoopDocumentStore`
- New auth repository over relational store.
	- `internal/core/auth/user_repository.go`
- New operation helpers for repository-defined execution.
	- `storage.RelationalExec(...)`
	- `storage.RelationalQueryOne(...)`
	- `storage.RelationalQueryMany(...)`
	- `storage.DocumentRun(...)`

### Changed
- App dependency wiring now initializes store surfaces when Postgres is enabled.
	- Added `Dependencies.Store`
	- Added `Dependencies.RelationalStore`
	- Added `Dependencies.DocumentStore`
- Auth engine wiring now uses store-backed provider path.
	- `StoreUserProvider -> UserRepository -> RelationalStore -> Postgres`
- `cmd/perftoken` updated to match new auth/store wiring.
- Core DB package scope narrowed to Postgres connectivity and migrations for storage backends.

### Removed
- Legacy core DB helper files:
	- `internal/core/db/queries.go`
	- `internal/core/db/queries_test.go`
	- `internal/core/db/tx.go`
	- `internal/core/db/tx_test.go`

### Documentation
- Rewrote architecture/docs set for store-first model with beginner-focused detail:
	- `docs/overview.md`
	- `docs/architecture.md`
	- `docs/modules.md`
	- `docs/module_guide.md`
	- `docs/crud-examples.md`
	- `docs/workflows.md`
	- `docs/environment-variables.md`
	- `docs/auth-goauth.md`
- Updated auth bootstrap docs for store-backed provider and repository wiring.
- Updated governance instructions in `AGENTS.md` for enforced data-layer constraints.

## v0.6.0 (2026-04-05)

### Breaking Changes
- Cache config model changed from static tags to structured dynamic tag specs.
	- `cache.CacheReadConfig.Tags` -> `cache.CacheReadConfig.TagSpecs`
	- `cache.CacheInvalidateConfig.Tags` -> `cache.CacheInvalidateConfig.TagSpecs`
- Preset option APIs now accept `cache.CacheTagSpec` values.
	- `policy.WithCache(ttl, ...)`
	- `policy.WithInvalidateTags(...)`
- Cache invalidation metadata was renamed for analyzer/validator consistency.
	- `CacheInvalidateMetadata.TagCount` -> `CacheInvalidateMetadata.TagSpecCount`

### Added
- Dynamic scoped cache invalidation tags (`cache.CacheTagSpec`) with runtime resolution from:
	- path params
	- authenticated tenant/user context
	- literal key/value scope dimensions
- New cache key template preparation and reuse path for lower overhead on hot routes.
- In-process cache tag-version token memoization with configurable TTL.
	- New env var: `CACHE_TAG_VERSION_CACHE_TTL` (default `250ms`)
- New browser/proxy cache directive policy:
	- `policy.CacheControl(policy.CacheControlConfig{...})`
- New middleware instrumentation controls:
	- `HTTP_MIDDLEWARE_TRACING_EXCLUDE_PATHS`
	- `METRICS_EXCLUDE_PATHS`
- New tests and benchmarks across app/httpx/cache/metrics/policy modules.

### Changed
- Cache read path now resolves and validates scoped tag names before key generation.
- Cache invalidation now resolves scoped tags from request/auth context and bumps only resolved scopes after successful `2xx` writes.
- Cache and rate-limit prod defaults are now fail-closed by default, with startup lint rejecting fail-open in prod when enabled.
- Tracing middleware supports exact-path exclusion and improved response-writer capability forwarding.
- Request-timeout middleware now bypasses SSE and websocket upgrade flows.
- Metrics middleware supports excluded paths and improved route-pattern propagation through wrapped writers.
- CORS denied preflight and metrics auth failures now return standardized error envelopes.
- Adapter decode path reduced repeated generic checks for request types without bodies.

### Documentation
- Rewrote cache documentation for dynamic tag specs, bump-miss invalidation, and scoped invalidation strategies.
- Updated policy reference with:
	- `TagSpecs` model
	- `CacheControl` policy usage and validation
	- revised preset and route stack examples
- Updated module and CRUD guides to replace legacy static tag examples with scoped `TagSpecs` examples.
- Expanded environment variable docs for new cache, tracing, metrics, and prod hardening settings.

## v0.5.0
- Public template release baseline
- Module system
- Strict policy engine
- goAuth integration
- Redis-backed cache and rate limiting
- Observability foundations (metrics, tracing, structured logs)
- Scaffolder workflow for module generation

## Changelog Rules
- Every release must update this file.
- Entries must be specific and verifiable.
- Avoid vague notes like "misc fixes".
