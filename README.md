[![Go Version](https://img.shields.io/badge/go-1.26+-00ADD8?logo=go)](go.mod)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Release](https://img.shields.io/badge/release-v0.7.0-brightgreen)](CHANGELOG.md)

# ProjectBook Backend

Policy-driven, production-grade Go backend for ProjectBook

## Repository Context

This repository is the backend API for **ProjectBook**.

- Frontend repository: [https://github.com/MrEthical07/projectbook](https://github.com/MrEthical07/projectbook)
- Backend repository: [https://github.com/MrEthical07/projectbook-backend](https://github.com/MrEthical07/projectbook-backend)
- Local development note: if available in your workspace, the sibling `../web` folder contains the frontend code.

## What This Is

ProjectBook Backend uses a modular Go API foundation focused on production use from day one.

It provides:
- a module-oriented API architecture
- policy-based middleware wiring
- built-in auth, caching, rate limiting, and observability primitives
- a store-first data layer with strict boundaries

Start here:
- Architecture: [docs/architecture.md](docs/architecture.md)
- Authentication: [docs/auth.md](docs/auth.md)

## Role in System

This backend is the enforcement layer of ProjectBook.

It is responsible for:
- enforcing permissions and policies
- maintaining data integrity
- controlling all data access and mutations

All operations must pass through this backend.
There are no alternative execution paths or bypass mechanisms.


## Data Layer Architecture

Enforced flow:

Service -> Repository -> Store -> Backend

Hard rules:
- services call repositories for all data operations
- services may call store.WithTx(...) only to define transaction boundaries for write operations; they must not call store execution methods (Execute, Query, etc.)
- repositories own all data access logic and call store execution methods (Execute, Query, etc.)
- repositories must not control transaction boundaries
- handlers never call DB/store directly
- one storage type per module (relational or document)
- transaction API exists at store layer and is used only for write paths; services define the boundary via store.WithTx and repositories perform all store execution calls inside that scope

## System Constraints

The backend enforces strict architectural boundaries:

- Handlers never access database or store directly
- Services define transaction boundaries but do not execute queries
- Repositories exclusively own all data access logic
- No cross-module access outside defined runtime wiring
- All requests pass through a strictly ordered policy pipeline before execution
- Policy stages execute in a strictly defined and validated order

Forbidden patterns:
- Direct database access outside repositories
- Custom authentication or permission checks outside policy stack
- Bypassing cache identity dimensions for authenticated data
- Defining transactions outside service layer
- Accessing dependencies outside module runtime container

The policy pipeline is enforced at runtime and validated at startup.
These constraints are enforced through code structure, policy layers, and startup validation—not developer convention.


## System Guarantees

ProjectBook enforces the following guarantees:

- All data access passes through controlled repository/store boundaries
- No direct database access from handlers or UI layers
- Permissions are enforced consistently across all operations
- Artifact relationships remain explicit and traceable
- System behavior is deterministic under defined inputs
- Cache correctness enforced via bump-miss atomic invalidation
- Authentication and authorization always validated before execution

These guarantees are enforced by architecture, not convention.

## Tradeoffs

This architecture introduces deliberate tradeoffs:

- Increased rigidity due to strict module and data boundaries
- Higher complexity in policy wiring and middleware composition
- Reduced flexibility for custom execution paths due to enforced policy pipeline
- Additional operational overhead from hybrid persistence (PostgreSQL + MongoDB)

These tradeoffs are intentional to ensure consistency, security, and enforceable system behavior.

## Request Lifecycle

1. API request reaches handler
2. Policy pipeline executes in strict order:
   auth → project scope → permission resolution → RBAC → rate limiting → caching
3. Handler validates request input
4. Service defines transaction boundaries
5. Repository executes data operations via store
6. Store interacts with database
7. Response is returned through the same controlled path

No request can bypass policy enforcement or data boundaries.
All execution follows a single controlled path enforced by the policy pipeline.


## Core Capabilities

- Policy-driven request pipeline enforcing authentication, authorization, and rate limiting
- Store-first data architecture with strict repository boundaries
- Module-based API structure with explicit domain separation
- Bitmask-based RBAC with constant-time permission evaluation
- Redis-backed caching with bump-miss invalidation model
- Hybrid persistence (PostgreSQL + MongoDB) aligned to data shape requirements
- Observability stack providing metrics, tracing, and structured logging across all modules
- Strict startup validation ensuring safe runtime configuration

## Permission Model

- Bitmask-based RBAC system using uint64 permission masks
- Role defaults combined with per-member overrides
- Permissions resolved per request for (user, project) context
- Enforced in policy pipeline before handler execution

Permission evaluation is constant-time and consistent across all operations.

## Failure Handling

- Fail-fast startup: application does not boot on invalid configuration
- Transaction failures trigger automatic rollback
- Cache failures fall back to controlled execution modes (fail-open or fail-closed)
- Standardized error structure (`AppError`) ensures consistent error handling

System behavior under failure is explicit and controlled.


## Cache System

- Redis-backed cache with explicit route opt-in
- Bump-Miss invalidation model using atomic tag versioning
- Writes trigger version increments, forcing immediate cache misses
- Cache keys derived from canonical request context (project, user, params)

This ensures strong cache correctness without relying on TTL-based expiration.


## Philosophy

- Secure by default in production-sensitive paths
- Explicit policies over implicit behavior
- Fail-fast validation at startup for unsafe configurations
- One enforced data-layer architecture over compatibility layers

This backend prioritizes enforceability over flexibility.


## Acknowledgments

ProjectBook Backend uses **goAuth** as its authentication engine.

`goAuth` is an open-source authentication framework that powers route-level auth workflows and identity lifecycle integration in this backend.

- goAuth repository: [https://github.com/MrEthical07/goAuth](https://github.com/MrEthical07/goAuth)

## Quick Start

```bash
go run ./cmd/api
```

Default configuration enables Postgres, Redis, Mongo, auth, cache, rate-limit, and permissions.
Ensure Postgres, Redis, and Mongo are running locally before using default startup.

After startup:
- Liveness: GET /healthz
- Readiness: GET /readyz

## Docker (Production)

Build the production image from the backend repository root:

```bash
docker build -t projectbook-backend:prod .
```

Run the image with runtime-injected configuration:

```bash
docker run --rm -p 8080:8080 \
	-e APP_ENV=prod \
	-e HTTP_ADDR=:8080 \
	-e POSTGRES_URL=postgres://user:pass@postgres:5432/projectbook?sslmode=disable \
	-e REDIS_ADDR=redis:6379 \
	-e MONGO_URL=mongodb://mongo:27017 \
	-e MONGO_DB=projectbook \
	-e PROJECTBOOK_PERMISSION_CONTEXT_SECRET=replace-with-strong-secret \
	-e WEB_APP_BASE_URL=https://app.example.com \
	-e METRICS_ENABLED=false \
	projectbook-backend:prod
```

Container notes:
- No secrets are baked into the image.
- `.env` files are not required for image build.
- Postgres, Redis, and MongoDB are external runtime dependencies and are not bundled in the image.

### Minimal profile (lean features, core stores still required)

Use the profile that keeps Postgres/Redis/Mongo active but disables auth/cache/rate-limit/permissions:

```bash
cp .env.example .env
# edit .env and enable:
# APP_PROFILE=minimal

go run ./cmd/api
```

### Full mode (Postgres + Redis + Mongo + auth)

Use default settings (or .env.example full-mode values), then run:

```bash
go run ./cmd/api
```

Equivalent explicit full-mode toggles are shown in .env.example:
- POSTGRES_ENABLED=true with valid POSTGRES_URL
- REDIS_ENABLED=true with valid REDIS_ADDR
- MONGO_ENABLED=true with valid MONGO_URL
- AUTH_ENABLED=true
- RATELIMIT_ENABLED=true
- CACHE_ENABLED=true
- PERMISSIONS_ENABLED=true

CORS configuration notes:
- `allowedOrigins` (comma-separated) controls which browser origins are accepted.
- `denyOrigins` (optional, comma-separated) explicitly blocks origins before allow-list checks.
- localhost origins are accepted by default for development unless explicitly denied.

### Note
For a production deployment, ensure all environment variables are properly set and that Postgres, Redis, and MongoDB are running and accessible with the provided connection details before starting the backend application. It is recommended to set all env variables properly before testing the frontend application, as the frontend relies on the backend API for authentication and data operations.

## Docs Navigation

- Architecture: [docs/architecture.md](docs/architecture.md)
- Auth: [docs/auth.md](docs/auth.md)
- Policies: [docs/policies.md](docs/policies.md)
- Cache guide: [docs/cache-guide.md](docs/cache-guide.md)
- Performance runbook: [docs/performance-testing.md](docs/performance-testing.md)
- Environment variables: [docs/environment-variables.md](docs/environment-variables.md)
- Workflows: [docs/workflows/README.md](docs/workflows/README.md)
- Test docs: [docs/test/README.md](docs/test/README.md)
- Contributor playbook: [AGENTS.md](AGENTS.md)

## Versioning And Updates

- This repository was initially bootstrapped from the SuperAPI template baseline.
- It is now maintained as the ProjectBook backend and does not auto-sync with upstream template changes.
- Upgrades from upstream are manual: compare changes, port intentionally, and validate with tests/build.
- Initial bootstrap baseline: SuperAPI v0.7.0.

## Release Hygiene

Before publishing a downstream release:

```bash
go test ./...
go build ./...
```

## Contributing

Contribution process and governance rules are documented in [CONTRIBUTING.md](CONTRIBUTING.md).
