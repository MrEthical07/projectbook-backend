# Sidebar Module Workflows

Module path: `internal/modules/sidebar`

## Route Inventory

1. `POST /api/v1/projects/{projectId}/sidebar/artifacts`
2. `PUT /api/v1/projects/{projectId}/sidebar/artifacts/{artifactId}/rename`
3. `DELETE /api/v1/projects/{projectId}/sidebar/artifacts/{artifactId}`

## Policy Chain

All routes:
1. `AuthRequired`
2. `ProjectRequired`
3. `ProjectMatchFromPath("projectId")`
4. `ResolvePermissions`
5. `RequireJSON` (all three routes)
6. `RequireAnyPermission(...)`
7. `CacheInvalidateOptional(...)`

## Permission Mapping

Create route allows any of:
- `PermStoryCreate`
- `PermProblemCreate`
- `PermIdeaCreate`
- `PermTaskCreate`
- `PermFeedbackCreate`
- `PermPageCreate`

Rename route allows any edit equivalents.
Delete route allows any delete equivalents.

## Flow

- Handler validates sidebar artifact operation request.
- Service opens `store.WithTx(...)` for each write operation.
- Repository executes cross-artifact rename/create/delete logic.

## Side Effects

Sidebar is a cross-module invalidator.

Invalidate tags include:
- all `artifacts.*` project tags
- all `pages.*` project tags
- all `resources.*` project tags

This intentionally forces fresh reads across sidebar-affected module surfaces.

## Route Workflow Matrix

| Route | Required permissions | Policy details | Handler-service-repo flow |
|---|---|---|---|
| `POST /api/v1/projects/{projectId}/sidebar/artifacts` | any create permission (`story/problem/idea/task/feedback/page`) | auth -> project -> resolver -> `RequireJSON` -> `RequireAnyPermission(create set)` -> optional cache invalidate | `CreateSidebarArtifact` handler -> service determines target artifact family -> `store.WithTx(...)` -> repo create workflow |
| `PUT /api/v1/projects/{projectId}/sidebar/artifacts/{artifactId}/rename` | any edit permission (`story/problem/idea/task/feedback/page`) | auth -> project -> resolver -> `RequireJSON` -> `RequireAnyPermission(edit set)` -> optional cache invalidate | `RenameSidebarArtifact` handler parses `artifactId` -> service rename rules -> transactional repo update |
| `DELETE /api/v1/projects/{projectId}/sidebar/artifacts/{artifactId}` | any delete permission (`story/problem/idea/task/feedback/page`) | auth -> project -> resolver -> `RequireJSON` -> `RequireAnyPermission(delete set)` -> optional cache invalidate | `DeleteSidebarArtifact` handler parses `artifactId` -> service delete rules -> transactional repo delete + cross-module invalidation |

## Troubleshooting Scenarios

1. sidebar operation succeeds but other module lists stale:
- verify invalidation tag list remains complete.
2. permission denied despite role:
- check `RequireAnyPermission` set matches action semantics.
3. operation affects wrong artifact family:
- verify repository artifact type resolution logic.

## What To Check During Changes

- keep wide invalidation coverage for cross-module consistency
- keep permission-any semantics by action type
- keep transaction wrapping for all writes
