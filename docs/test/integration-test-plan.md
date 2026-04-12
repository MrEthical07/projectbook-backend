# Integration Test Plan

## Goal
Implement exhaustive integration coverage for every registered API route (85 unique routes). Coverage is route-driven and policy-aware:

- auth and project isolation
- permissions resolver and RBAC enforcement
- request validation and error contracts
- relational metadata + Mongo document consistency
- cache read behavior and write invalidation behavior
- delete/link side effects

Use this plan with docs/test/route-coverage-matrix.md.

## Preconditions
- Postgres, Redis, and Mongo are running.
- Migrations are applied and clean (version 11, dirty=false).
- API starts with production-like policy stack.
- Test runner can obtain auth tokens for at least these personas:
  - owner/admin
  - project member with read-only permissions
  - unauthorized user (no membership)
  - anonymous user (no token)

## Test Harness Requirements
- Framework: Go integration suite (httpexpect/resty + testify) or k6/http checks for selected throughput paths.
- Isolation:
  - one project fixture per test class
  - deterministic user IDs and permission masks
  - cleanup per test class (project-level cascading cleanup)
- Assertions:
  - status code and error code
  - response schema and required fields
  - side effects in Postgres and Mongo
  - cache headers and cache tag invalidation behavior

## Scenario Packs (Applied Per Route)

### Pack SMOKE-HEALTH
For /healthz and /readyz:
- service liveness shape
- readiness dependency status map
- fail behavior when dependency is intentionally unavailable (separate env run)

### Pack AUTH-PUBLIC
For auth public endpoints:
- valid payload success path
- malformed payload 400
- replay/idempotency where applicable
- token lifecycle checks (refresh, expiry behavior)
- abuse baseline (rate-limit if configured)

### Pack SYSTEM-UTILITY
For /system/parse-duration:
- valid input parsing
- malformed units
- overflow/negative edge cases
- deterministic output schema

### Pack PROTECTED-READ
For protected GET routes:
- anonymous -> 401
- authenticated non-member -> 403 or scoped denial
- member without RBAC bit -> 403
- member with permission -> 200 schema-valid response
- project path mismatch handling
- pagination/filter edge cases where query params exist
- cache behavior:
  - first hit miss then set
  - second hit served from cache (or equivalent cache outcome metric)
  - Cache-Control and Vary headers

### Pack PROTECTED-WRITE
For protected POST/PUT/DELETE routes:
- anonymous -> 401
- authenticated without permission -> 403
- invalid body -> 400 (schema and enum validation)
- valid body -> success status + schema
- idempotency/conflict behavior (repeat requests)
- transactional integrity:
  - relational row changes are correct
  - Mongo document upsert/delete reflects new revision/content
- cache invalidation:
  - tags for impacted read routes are bumped
  - subsequent reads return fresh values (no stale cache)

## Domain-Specific Deep Checks

### Artifacts (Stories/Journeys/Problems/Ideas/Tasks/Feedback)
- status transition matrix enforcement
- immutable-state restrictions (locked/archived/completed rules)
- slug and ID lookup parity
- document content is read from Mongo collections, not outbox payloads
- list/detail response parity after writes

### Pages
- create/update/rename flows update page_documents
- linked artifacts count remains coherent with artifact_links
- archived-state mutation restrictions

### Resources
- resource status transition checks
- content updates reflected in resource_documents
- file_type/doc_type validation

### Sidebar
- create/rename/delete route behavior for mixed artifact types
- delete side effects remove Mongo document from correct collection
- cross-module cache invalidation tags (artifacts/pages/resources) on writes

### Project and Team
- project settings/archive/delete authorization and membership rules
- project delete cascade expectations validated (relational + document presence)
- invite/member/role permission mutation correctness

### Home and Activity
- dashboard/reference/feed consistency with project/team updates
- account update behavior and authorization correctness

## Coverage Gates
A route is complete only when all conditions hold:
- mapped scenario pack tests are implemented
- positive and negative auth paths pass
- schema assertions pass
- side effects asserted in backing stores as applicable
- cache semantics tested for mapped read/write routes

Global pass criteria:
- 100% route matrix coverage (85/85)
- 0 flaky tests over 3 consecutive runs
- no dirty migration state after test suite

## Execution Phases
1. Foundation
- shared fixture factory (users, projects, auth tokens)
- helper assertions for common error contract and auth failures

2. Core route families
- health/system/auth
- project/team/home/activity

3. Content modules
- artifacts
- pages
- resources
- sidebar

4. Cross-cutting validation
- cache behavior and invalidation probes
- Mongo-vs-relational consistency probes

5. CI hardening
- split into parallel-safe packages
- add retry budget only for external dependency startup, not assertions

## Deliverables to Implement Next
- integration test package scaffolding under internal/tests/integration
- fixture seeders for users/projects/permissions
- per-module *_integration_test.go files matching route matrix rows
- CI target: make test-integration
