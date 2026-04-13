# Test Documentation

This folder contains runtime verification and integration test planning artifacts.

## Files
- integration-test-plan.md: Scenario packs, execution phases, and completion gates.
- route-coverage-matrix.md: Deduplicated route inventory (87 unique routes) mapped to scenario packs.
- smoke-validation-2026-04-13.md: Migration/startup/smoke validation report.

## Intended Workflow
1. Use route-coverage-matrix.md to generate one integration test case set per route.
2. Implement scenario pack assertions from integration-test-plan.md.
3. Mark execution progress in your preferred tracker (test IDs per route).
4. Keep smoke report updated whenever environment/bootstrap behavior changes.

## Current Execution Status (2026-04-13)

- Integration package: pass (`go test ./internal/tests/integration -count=1 -v`)
- Full repository tests: pass (`go test ./...`)
- Static route/policy verifier: pass (`go run ./cmd/superapi-verify ./...`)

Recently added deep sweep coverage includes:
- artifacts lifecycle and transition rules (including lock/select/status edge cases)
- calendar event lifecycle (create/read/update/delete)
- home/project/team lifecycle with project deletion cascade assertions
- auth signup limiter stabilization in integration harness cleanup path

## Integration Test Harness (Real Databases)

Initial executable suites are now implemented under internal/tests/integration:
- smoke + readiness validation
- system parse-duration contract checks
- auth lifecycle (signup, verify-email, login, refresh, malformed inputs)
- project settings RBAC + cache hit/miss + invalidation checks
- resources relational + Mongo consistency + cache invalidation checks
- artifacts/calendar/team/home/project deep lifecycle sweeps

The harness provisions isolated runtime resources per test run:
- a dedicated temporary Postgres database (created/dropped automatically)
- a dedicated Redis logical DB (flushed before/after run)
- a dedicated Mongo database (dropped after run)

### Run

```bash
make test-integration
```

Equivalent direct command:

```bash
INTEGRATION_TESTS=1 go test ./internal/tests/integration -count=1 -v
```

### Optional Environment Overrides

- IT_PG_ADMIN_URL (default resolves from POSTGRES_URL or local postgres)
- IT_REDIS_ADDR
- IT_REDIS_PASSWORD
- IT_REDIS_DB
- IT_MONGO_URL

If INTEGRATION_TESTS is not set to a truthy value, integration tests are skipped.
