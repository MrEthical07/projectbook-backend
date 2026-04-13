# Resources Module Workflows

Module path: `internal/modules/resources`

## Route Inventory

1. `GET /api/v1/projects/{projectId}/resources`
2. `POST /api/v1/projects/{projectId}/resources`
3. `GET /api/v1/projects/{projectId}/resources/{resourceId}`
4. `PUT /api/v1/projects/{projectId}/resources/{resourceId}`
5. `PUT /api/v1/projects/{projectId}/resources/{resourceId}/status`

## Policy Chain

Read routes:
1. `AuthRequired`
2. `ProjectRequired`
3. `ProjectMatchFromPath("projectId")`
4. `ResolvePermissions`
5. `RequirePermission(PermResourceView)`
6. `CacheReadOptional`
7. `CacheControlOptional`

Write routes:
1. `AuthRequired`
2. `ProjectRequired`
3. `ProjectMatchFromPath("projectId")`
4. `ResolvePermissions`
5. `RequireJSON`
6. `RequirePermission(...)`
7. `CacheInvalidateOptional`

## Flow

- Handler delegates to service.
- Service validates and orchestrates.
- Repository performs relational + document operations.
- Write paths use `store.WithTx(...)` in service.

## Side Effects

- read cache tags: `resources.project`, `resources.resource`
- write invalidation bumps same tags
- document revisions updated for resource content writes

## Route Workflow Matrix

| Route | Required permission | Policy details | Handler-service-repo flow |
|---|---|---|---|
| `GET /api/v1/projects/{projectId}/resources` | `PermResourceView` | auth -> project -> resolver -> RBAC -> optional cache read -> optional cache control | `ListResources` handler parses pagination/filter -> service read path -> repo list query |
| `POST /api/v1/projects/{projectId}/resources` | `PermResourceCreate` | auth -> project -> resolver -> `RequireJSON` -> RBAC -> optional cache invalidate | `CreateResource` handler -> service validation -> `store.WithTx(...)` -> repo creates row + document |
| `GET /api/v1/projects/{projectId}/resources/{resourceId}` | `PermResourceView` | auth -> project -> resolver -> RBAC -> optional cache read -> optional cache control | `GetResource` handler reads `resourceId` -> service -> repo relational/document join |
| `PUT /api/v1/projects/{projectId}/resources/{resourceId}` | `PermResourceEdit` | auth -> project -> resolver -> `RequireJSON` -> RBAC -> optional cache invalidate | `UpdateResource` handler -> service rules -> transactional repo update |
| `PUT /api/v1/projects/{projectId}/resources/{resourceId}/status` | `PermResourceStatusChange` | auth -> project -> resolver -> `RequireJSON` -> RBAC -> optional cache invalidate | `UpdateResourceStatus` handler -> service transition checks -> transactional status write |

## Troubleshooting Scenarios

1. status update rejected:
- check status enum/transition validation.
2. read response missing rich detail fields:
- check document store wiring and resource document lookup.
3. stale list after update:
- verify cache invalidation tags and project path param extraction.

## What To Check During Changes

- preserve project-scoped policy order
- preserve transaction boundary for writes
- preserve cache isolation (`VaryBy.ProjectID` and identity safety)
