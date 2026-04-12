# Smoke Validation Report (2026-04-13)

## Scope
Validated baseline runtime behavior after Mongo-first document cutover, route cache tag wiring, and legacy outbox cleanup.

## Environment
- Host: Windows
- Workspace: Backend
- API startup command: go run ./cmd/api
- API bind address observed: :8080

## Migration Execution
Commands executed (with .env values loaded into process env):

```powershell
go run ./cmd/migrate version
go run ./cmd/migrate up
go run ./cmd/migrate version
```

Results:
- Initial version: no_version
- Up migration: success
- Final version: 11
- Dirty flag: false

## Smoke Checks
Executed against running API at http://127.0.0.1:8080.

### Health and Readiness
- GET /healthz -> 200
- GET /readyz -> 200
- Readiness dependencies reported healthy: postgres, redis, mongo

### Auth Protection Baseline
- GET /api/v1/projects/atlas-2026/stories -> 401
- GET /api/v1/projects/atlas-2026/pages -> 401
- GET /api/v1/projects/atlas-2026/resources -> 401
- POST /api/v1/projects/atlas-2026/sidebar/artifacts -> 401

## Automated Validation
- go test ./... -> pass (after clearing process env pollution from .env keys)
- go run ./cmd/superapi-verify ./... -> verify: ok

## Legacy Cleanup Included
Removed legacy, unused outbox sync package files:
- internal/infrastructure/docsync/doc.go
- internal/infrastructure/docsync/processor.go

Also fixed Windows migration source URL handling so migrate CLI works with file sources:
- internal/core/db/migrate.go
