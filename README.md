[![Go Version](https://img.shields.io/badge/go-1.26+-00ADD8?logo=go)](go.mod)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Release](https://img.shields.io/badge/release-v0.7.0-brightgreen)](CHANGELOG.md)

# ProjectBook Backend

Production-grade Go backend for ProjectBook.

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

## Features

- Module system for explicit, composable API domains
- Strict startup validation for runtime and policy configuration
- goAuth integration for route-level authentication workflows
- Redis-backed response cache with dynamic TagSpecs invalidation and Redis-backed rate limiting
- Browser/proxy cache directives with policy.CacheControl(...)
- Observability stack: metrics, tracing, and structured logs
- Store-first data layer contracts in internal/core/storage
- Built-in scaffolder for generating production-oriented modules

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

## How To Build APIs

1. Create a module

```bash
make module name=projects
```

Expected output:

```text
generated module "projects" (package="projects" route=/api/v1/projects)
```

2. Confirm module wiring

internal/modules/modules.go is updated automatically with import + projects.New() entry.

3. Verify and run

```bash
go test ./internal/devx/modulegen ./internal/modules/projects
go run ./cmd/superapi-verify ./internal/modules/projects
go run ./cmd/api
```

4. Add handlers and service logic in the generated module files.

5. Add repositories that execute store operations and keep query logic inside repository.

Guides:
- Module workflows: [docs/workflows/README.md](docs/workflows/README.md)
- Route details snapshot: [docs/routeDetails.md](docs/routeDetails.md)
- Contributor playbook: [AGENTS.md](AGENTS.md)

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

## Philosophy

- Secure by default in production-sensitive paths
- Explicit policies over implicit behavior
- Fail-fast validation at startup for unsafe configurations
- One enforced data-layer architecture over compatibility layers

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
