# Architecture

ProjectBook Backend is a production-grade modular Go API with strict layering and policy-validated routing.

This document is the canonical architecture reference for this repository.

## 1. Core Contract

Enforced data flow:

`Handler -> Service -> Repository -> Store -> Backend`

Mandatory boundaries:
- Handlers own HTTP transport only.
- Services own business orchestration and write transaction boundaries.
- Repositories own query logic and persistence mapping.
- Stores own execution semantics only.

Do not bypass this flow.

## 2. Runtime Entry And Wiring

Entrypoint:
- `cmd/api/main.go`

Core runtime composition:
- `internal/core/app`

Module registration surface:
- `internal/modules/modules.go`

Current runtime module order:
1. `auth`
2. `home`
3. `project`
4. `artifacts`
5. `resources`
6. `pages`
7. `calendar`
8. `sidebar`
9. `activity`
10. `team`
11. `health`
12. `system`

## 3. Startup Lifecycle

The API process performs fail-fast startup in this order:

1. Load env configuration (`config.Load`).
2. Run config lint (`Config.Lint`) including feature dependency checks.
3. Initialize logging and app shell.
4. Initialize dependencies in `initDependencies`:
   - Postgres pool and relational store
   - Redis client
   - metrics service
   - auth mode and goAuth engine
   - rate limiter
   - cache manager
   - permissions resolver + permissions lifecycle startup sync
   - Mongo client + database + bootstrap + Mongo document store
   - tracing service
5. Bind dependencies into modules.
6. Register routes from all modules.
7. Start HTTP server.

Startup constraints currently enforced:
- Postgres must be enabled.
- Redis must be enabled.
- Mongo must be enabled.
- Auth requires Redis + Postgres.
- Cache/rate-limit require Redis.
- Permissions require Postgres.

## 4. Global HTTP Pipeline

Global middleware is assembled in `internal/core/httpx/globalmiddleware.go`.

Execution order (outermost to innermost):
1. request id
2. client ip
3. recoverer
4. CORS
5. security headers
6. max body bytes
7. request timeout
8. tracing
9. access log
10. router dispatch

This order is intentional for diagnostics, safety, and policy correctness.

CORS origin policy notes:
- Browser origins are controlled via `allowedOrigins` (with legacy `HTTP_MIDDLEWARE_CORS_ALLOW_ORIGINS` alias support).
- `denyOrigins` is optional and evaluated before the allow-list.
- Localhost origins are allowed by default for development unless explicitly denied.

## 5. Route Policy System

Policies are route-scoped middleware decorators under `internal/core/policy`.

Each route is validated at registration with policy metadata rules.

Required protected-route stage order:
1. `AuthRequired`
2. `ProjectRequired` / `ProjectMatchFromPath`
3. `ResolvePermissions`
4. RBAC (`RequirePermission`, `RequireAnyPermission`, `RequireAllPermissions`)
5. `RateLimit`
6. cache read/invalidate
7. cache-control

Validator-enforced safety rules include:
- auth is required for project/resolver/RBAC policies
- resolver is required before RBAC checks
- project path routes must include project policies
- authenticated cache reads must vary by user or project identity

## 6. Module Runtime Pattern

Every module follows:
- `dto.go`: transport contracts and validation
- `handler.go`: request extraction and response shaping
- `service.go`: business workflows and transaction orchestration
- `repo.go`: persistence operations and mappings
- `routes.go`: route + policy registration
- `module.go`: module constructor and dependency binding

Dependency access uses `modulekit.Runtime` surfaces:
- auth engine + mode
- relational store
- document store
- cache manager
- limiter
- permissions resolver
- shared dependencies

## 7. Data Plane

### 7.1 Relational plane

- Backend: Postgres
- Store contract: `storage.RelationalStore`
- Execution entrypoint: `store.Execute(...)`
- Transaction entrypoint: `store.WithTx(...)`

### 7.2 Document plane

- Backend: MongoDB
- Store contract: `storage.DocumentStore`
- Used by hybrid modules for rich document payloads and revisions

### 7.3 Module storage model

Current module backend families:
- Relational only: `auth`, `home`, `project`, `calendar`, `activity`, `team`, `system`, `health`
- Relational + document: `artifacts`, `resources`, `pages`, `sidebar`

## 8. Transaction Model

Write paths are service-owned transactions.

Pattern:
1. Service validates request and current state.
2. Service opens `store.WithTx(ctx, fn)`.
3. Service calls repository methods inside `fn`.
4. Repository executes relational/document operations through store APIs.
5. Store commits or rolls back.

Read paths do not require forced transaction wrapping.

## 9. Authentication And Authorization Architecture

### 9.1 Authentication

- Engine: goAuth
- Integration location: `internal/core/auth`
- Provider bridge: store-backed `StoreUserProvider`
- User persistence: `UserRepository` over relational store
- System session-context route issues a backend-signed permission-context token for frontend server-side permission hydration (`internal/modules/system/routes.go`)

Auth mode surface:
- `jwt_only`
- `hybrid` (default)
- `strict`

### 9.2 Authorization

Authorization is not delegated to goAuth.

ProjectBook authorization model:
- request-scoped project isolation (`ProjectRequired`, `ProjectMatchFromPath`)
- permission resolution (`ResolvePermissions`)
- RBAC bitmask checks (`RequirePermission` family)

## 10. Cache Architecture

Cache is Redis-backed and route-opt-in.

Core components:
- manager: `internal/core/cache/manager.go`
- policies: `CacheRead`, `CacheInvalidate`, `CacheControl`
- optional wrappers: `CacheReadOptional`, `CacheInvalidateOptional`, `CacheControlOptional`

Design model:
- key isolation via `VaryBy`
- freshness/invalidation via `TagSpecs` version bumping

## 11. Rate-Limit Architecture

Rate limiting is Redis-backed per-route policy.

Core policy:
- `RateLimit`
- `RateLimitWithKeyer`

Typical scopes used in this API:
- IP (public auth endpoints)
- User (account-scoped operations)
- Project (project mutation routes)
- user/project/token-hash fallback (selected protected routes)

## 12. Permissions Lifecycle

On startup (when enabled), permissions lifecycle performs consistency tasks:
- role permission seed/resync
- project member permission mask resync for non-custom memberships
- resolver cache invalidation for impacted users/projects

This keeps RBAC enforcement deterministic across route checks.

## 13. Observability, Health, And Readiness

Health routes:
- `GET /healthz`: liveness
- `GET /readyz`: dependency readiness report

Metrics:
- request/route instrumentation
- cache and rate-limit outcome metrics

Tracing:
- OpenTelemetry lifecycle via core tracing service

Readiness probes are registered per dependency during startup wiring.

## 14. Failure Model

- Startup is fail-fast for invalid config and dependency wiring failures.
- Policy misconfiguration fails route registration immediately.
- Runtime dependency errors are mapped to explicit API errors (for example dependency unavailable).

## 15. Change Rules For Contributors

When changing architecture-sensitive code:
- keep module flow `handler -> service -> repository -> store`
- keep write transaction boundaries in services only
- do not expose db-driver/query objects in service contracts
- do not bypass policy validation
- do not bypass auth/project/resolver/RBAC order on protected routes
- avoid global mutable state

## 16. Related Documents

- [auth.md](auth.md)
- [policies.md](policies.md)
- [cache-guide.md](cache-guide.md)
- [environment-variables.md](environment-variables.md)
- [routeDetails.md](routeDetails.md)
- [workflows/README.md](workflows/README.md)
- [test/README.md](test/README.md)
