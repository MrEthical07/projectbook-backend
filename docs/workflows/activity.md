# Activity Module Workflows

Module path: `internal/modules/activity`

## Route Inventory

1. `GET /api/v1/projects/{projectId}/activity`

## Policy Chain

1. `AuthRequired`
2. `ProjectRequired`
3. `ProjectMatchFromPath("projectId")`
4. `ResolvePermissions`
5. `RequirePermission(PermProjectView)`

## Flow

1. Handler `ListProjectActivity` reads optional `limit` query.
2. Service `ListProjectActivity` orchestrates list retrieval.
3. Repository `ListProjectActivity` reads relational activity records.
4. Response returns list envelope.

Transaction:
- read-only, no transaction wrapper.

Side effects:
- no cache policy
- no write path

## Route Workflow Matrix

| Route | Required permission | Policy details | Handler-service-repo flow |
|---|---|---|---|
| `GET /api/v1/projects/{projectId}/activity` | `PermProjectView` | auth -> project -> project path match -> resolver -> RBAC | `ListProjectActivity` handler parses `limit` -> service read path -> repo activity query |

## Troubleshooting Scenarios

1. empty activity when data expected:
- check project scope path and membership permissions.
2. forbidden response:
- check resolved permission mask includes project view.

## What To Check During Changes

- keep route read-only
- keep strict project policy stack
- avoid adding write behavior into activity module
