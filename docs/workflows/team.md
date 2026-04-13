# Team Module Workflows

Module path: `internal/modules/team`

## Route Inventory

1. `GET /api/v1/projects/{projectId}/team/members`
2. `GET /api/v1/projects/{projectId}/team/roles`
3. `POST /api/v1/projects/{projectId}/team/invites`
4. `POST /api/v1/projects/{projectId}/team/invites/batch`
5. `DELETE /api/v1/projects/{projectId}/team/invites/{email}`
6. `PUT /api/v1/projects/{projectId}/team/members/{memberId}/permissions`
7. `PUT /api/v1/projects/{projectId}/team/roles/{role}/permissions`

## Policy Chain

Read routes:
1. `AuthRequired`
2. `ProjectRequired`
3. `ProjectMatchFromPath("projectId")`
4. `ResolvePermissions`
5. `RequirePermission(PermMemberView)`
6. `CacheRead` when cache manager enabled
7. `CacheControl` when cache manager enabled

Write routes:
1. `AuthRequired`
2. `ProjectRequired`
3. `ProjectMatchFromPath("projectId")`
4. `ResolvePermissions`
5. `RequirePermission(...)`
6. `RequireJSON` where body required
7. `RateLimit` when limiter enabled
8. `CacheInvalidate` when cache manager enabled

## Flow

- Handlers parse path/body and call service methods.
- Services enforce business rules and open `store.WithTx(...)` for writes.
- Repo handles invite/member/role persistence.

## Side Effects

- cache tags: `team.members`, `team.roles`
- write routes invalidate team tags
- permission updates trigger redis permission-scope invalidation for impacted users
- batch invites support partial success semantics (`207`)

## Route Detail Notes

| Route | Required permission | Policy details | Handler-service-repo flow |
|---|---|---|---|
| `GET /api/v1/projects/{projectId}/team/members` | `PermMemberView` | auth -> project -> resolver -> RBAC -> optional cache read -> optional cache control | `ListMembers` handler -> service read path -> repo member listing |
| `GET /api/v1/projects/{projectId}/team/roles` | `PermMemberView` | auth -> project -> resolver -> RBAC -> optional cache read -> optional cache control | `ListRoles` handler -> service read path -> repo roles listing |
| `POST /api/v1/projects/{projectId}/team/invites` | `PermMemberCreate` | auth -> project -> resolver -> RBAC -> `RequireJSON` -> optional rate-limit (`30/min` project scope) -> optional cache invalidate | `CreateInvite` handler -> service validation -> `store.WithTx(...)` -> repo invite create |
| `POST /api/v1/projects/{projectId}/team/invites/batch` | `PermMemberCreate` | auth -> project -> resolver -> RBAC -> `RequireJSON` -> optional rate-limit (`10/min` project scope) -> optional cache invalidate | `BatchInvites` handler -> service batch orchestration -> transactional repo writes with partial-success response (`207`) |
| `DELETE /api/v1/projects/{projectId}/team/invites/{email}` | `PermMemberDelete` | auth -> project -> resolver -> RBAC -> optional rate-limit (`30/min` project scope) -> optional cache invalidate | `CancelInvite` handler parses email -> service transactional cancellation -> repo update |
| `PUT /api/v1/projects/{projectId}/team/members/{memberId}/permissions` | `PermMemberEdit` | auth -> project -> resolver -> RBAC -> `RequireJSON` -> optional rate-limit (`20/min` project scope) -> optional cache invalidate | `UpdateMemberPermissions` handler -> service transactional permission mutation -> repo update + permission cache invalidation |
| `PUT /api/v1/projects/{projectId}/team/roles/{role}/permissions` | `PermMemberEdit` | auth -> project -> resolver -> RBAC -> `RequireJSON` -> optional rate-limit (`20/min` project scope) -> optional cache invalidate | `UpdateRolePermissions` handler -> service transactional role permission update -> repo update + permission cache invalidation |

## Troubleshooting Scenarios

1. invite endpoints return conflict/validation errors:
- check email normalization and existing invite/member state.
2. permission update appears saved but access unchanged:
- check permissions resolver cache invalidation for affected user.
3. team reads stale after write:
- verify team cache invalidation executed and tags are project-scoped.

## What To Check During Changes

- preserve project-scoped rate-limit rules
- preserve cache invalidation tags for team reads
- preserve permission invalidation side effects on member/role updates
