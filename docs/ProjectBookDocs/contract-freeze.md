# ProjectBook Contract Freeze

Date: 2026-04-10
Status: Closed (2026-04-11)

## Purpose

This document freezes the authoritative backend contract boundaries required before feature implementation.

## Frozen Decisions

1. Authentication direction: goAuth is retained for authentication only.
2. Authorization direction: ProjectBook permission-mask RBAC remains custom and project-scoped.
3. Isolation model: project is the top-level scope; tenant/workspace semantics are removed from runtime behavior.
4. Document store direction: Mongo document store will be enabled with automatic startup bootstrap and index verification.
5. Endpoint planning baseline: 84 unique endpoints tracked in endpoint-tracker artifacts.

## Authoritative Sources

Use these as source of truth for implementation behavior:

1. docs/ProjectBookDocs/API-GUIDELINES.md
2. API_plan.md
3. database.md
4. rbac.md
5. docs/ProjectBookDocs/endpoint-tracker.md
6. docs/ProjectBookDocs/endpoint-tracker.json
7. docs/ProjectBookDocs/endpoint-tracker.csv

## Legacy Documentation Policy

1. docs/auth.md is the canonical authentication document.
2. If historical references conflict with docs/ProjectBookDocs/* for route shape, storage ownership, policy ordering, or RBAC rules, ProjectBookDocs wins.
3. New backend feature work must reference the frozen ProjectBookDocs set first.

## Endpoint Tracker

Tracker artifacts are generated from API_plan and de-duplicated by method and path.

- Parsed rows: 86
- Unique endpoint target: 84

Status values:

- not_started
- in_progress
- implemented
- tested

## Exit Criteria For Freeze

Freeze is closed after completing all criteria below:

1. [x] Tenant semantics removed from runtime and schema surfaces.
2. [x] Auth repository/sqlc schema alignment complete.
3. [x] Project-scoped access resolver and project isolation policies live.
4. [x] Role-mask seed and resync lifecycle behavior implemented.
5. [x] Mongo document-store adapter and startup index/bootstrap behavior implemented.

Implementation now proceeds module-by-module using endpoint tracker status progression.
