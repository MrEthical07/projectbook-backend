# Pages Module Workflows

Module path: `internal/modules/pages`

## Route Inventory

1. `GET /api/v1/projects/{projectId}/pages`
2. `POST /api/v1/projects/{projectId}/pages`
3. `GET /api/v1/projects/{projectId}/pages/{slug}`
4. `PUT /api/v1/projects/{projectId}/pages/{pageId}`
5. `PUT /api/v1/projects/{projectId}/pages/{pageId}/rename`

## Policy Chain

Read routes:
1. `AuthRequired`
2. `ProjectRequired`
3. `ProjectMatchFromPath("projectId")`
4. `ResolvePermissions`
5. `RequirePermission(PermPageView)`
6. `CacheReadOptional`
7. `CacheControlOptional`

Write routes:
1. `AuthRequired`
2. `ProjectRequired`
3. `ProjectMatchFromPath("projectId")`
4. `ResolvePermissions`
5. `RequireJSON`
6. `RequirePermission(PermPageEdit|PermPageCreate)`
7. `CacheInvalidateOptional`

## Flow

- Handler parses route params/body and calls service.
- Service orchestrates page business rules.
- Repository persists page row + document payload.
- Write paths are wrapped in `store.WithTx(...)`.

## Side Effects

- cache read tags: `pages.project`, `pages.page`
- write invalidation bumps same tags
- rename route updates identity/display fields and invalidates page caches

## Route Workflow Matrix

| Route | Required permission | Policy details | Handler-service-repo flow |
|---|---|---|---|
| `GET /api/v1/projects/{projectId}/pages` | `PermPageView` | auth -> project -> resolver -> RBAC -> optional cache read -> optional cache control | `ListPages` handler parses filters -> service read path -> repo list query |
| `POST /api/v1/projects/{projectId}/pages` | `PermPageCreate` | auth -> project -> resolver -> `RequireJSON` -> RBAC -> optional cache invalidate | `CreatePage` handler -> service validation -> `store.WithTx(...)` -> repo row create + document create |
| `GET /api/v1/projects/{projectId}/pages/{slug}` | `PermPageView` | auth -> project -> resolver -> RBAC -> optional cache read -> optional cache control | `GetPage` handler resolves slug -> service -> repo detail read |
| `PUT /api/v1/projects/{projectId}/pages/{pageId}` | `PermPageEdit` | auth -> project -> resolver -> `RequireJSON` -> RBAC -> optional cache invalidate | `UpdatePage` handler -> service state checks -> transactional repo update |
| `PUT /api/v1/projects/{projectId}/pages/{pageId}/rename` | `PermPageEdit` | auth -> project -> resolver -> `RequireJSON` -> RBAC -> optional cache invalidate | `RenamePage` handler -> service uniqueness/slug rules -> transactional repo rename |

## Troubleshooting Scenarios

1. rename conflicts or not found errors:
- check slug uniqueness and id/slug resolution in repo.
2. page update saved but list stale:
- verify page cache invalidation path params.
3. detail content missing:
- verify document store read path.

## What To Check During Changes

- preserve rename semantics and validation
- preserve transaction boundaries in service for all writes
- preserve cache tags and vary dimensions
