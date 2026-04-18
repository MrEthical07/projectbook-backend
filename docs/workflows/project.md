# Project Module Workflows

Module path: `internal/modules/project`

## Route Inventory

1. `GET /api/v1/projects/{projectId}/dashboard`
2. `GET /api/v1/projects/{projectId}/access`
3. `GET /api/v1/projects/{projectId}/navigation`
4. `GET /api/v1/projects/{projectId}/settings`
5. `PUT /api/v1/projects/{projectId}/settings`
6. `POST /api/v1/projects/{projectId}/archive`
7. `DELETE /api/v1/projects/{projectId}`

## Protected Route Chain

All routes are project-scoped and follow:
1. `AuthRequired`
2. `ProjectRequired`
3. `ProjectMatchFromPath("projectId")`
4. `ResolvePermissions`
5. RBAC check on routes that require explicit permission
6. cache/rate-limit policies depending on route type and runtime feature toggles

## Route Workflows

### `GET /api/v1/projects/{projectId}/dashboard`

- RBAC: `PermProjectView`
- Cache read tag: `project.dashboard`

Flow:
1. Handler `Dashboard` -> Service `Dashboard`.
2. Service reads repo dashboard aggregate.

### `GET /api/v1/projects/{projectId}/access`

- No explicit `RequirePermission` in route chain.
- Cache read tag: `project.access`

Flow:
1. Handler `Access` -> Service `Access`.
2. Service reads principal access details.

### `GET /api/v1/projects/{projectId}/navigation`

- RBAC: `PermProjectView`
- Cache read tag: `project.navigation`

Flow:
1. Handler `Navigation` -> Service `Navigation`.
2. Service reads repo navigation aggregate.

### `GET /api/v1/projects/{projectId}/settings`

- RBAC: `PermProjectView`
- Cache read tag: `project.settings`

Flow:
1. Handler `GetSettings` -> Service `GetSettings`.
2. Service reads repo settings.

### `PUT /api/v1/projects/{projectId}/settings`

- RBAC: `PermProjectEdit`
- JSON required
- Rate-limited by project scope (if limiter enabled)
- Cache invalidates all project read tags

Flow:
1. Handler validates update payload.
2. Service opens `store.WithTx(...)`.
3. Repo updates settings.

### `POST /api/v1/projects/{projectId}/archive`

- RBAC: `PermProjectArchive`
- Rate-limited by project scope (if enabled)
- Cache invalidates project tags

Flow:
1. Handler `Archive` -> Service `Archive`.
2. Service opens transaction and archives project state.

### `DELETE /api/v1/projects/{projectId}`

- RBAC: `PermProjectDelete`
- Rate-limited by project scope (if enabled)
- Cache invalidates project tags

Flow:
1. Handler `Delete` -> Service `Delete`.
2. Service opens transaction and deletes project aggregate.

## Troubleshooting Scenarios

1. `403 project scope mismatch`:
- check principal project context vs path param.
2. settings update blocked unexpectedly:
- check resolver + required permission bit.
3. delete/archive appears successful but stale reads remain:
- verify project cache invalidation tags.

## What To Check During Changes

- Keep policy order strict for project routes.
- Keep writes transaction-boundary in service.
- Keep cache vary dimensions user+project for authenticated reads.
