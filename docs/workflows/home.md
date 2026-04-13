# Home Module Workflows

Module path: `internal/modules/home`

## Route Inventory

1. `GET /api/v1/home/dashboard`
2. `GET /api/v1/home/projects`
3. `GET /api/v1/home/projects/reference`
4. `GET /api/v1/home/invites`
5. `GET /api/v1/home/notifications`
6. `GET /api/v1/home/activity`
7. `GET /api/v1/home/dashboard-activity`
8. `GET /api/v1/home/account`
9. `GET /api/v1/home/docs`
10. `POST /api/v1/home/projects`
11. `POST /api/v1/home/invites/{inviteId}/accept`
12. `POST /api/v1/home/invites/{inviteId}/decline`
13. `PUT /api/v1/home/account`

## Base Policy Stacks

Read routes:
1. `AuthRequired`
2. `CacheRead` when cache manager enabled
3. `CacheControl` when cache manager enabled

Write routes:
1. `AuthRequired`
2. `RequireJSON` where body required
3. `RateLimit` when limiter enabled
4. `CacheInvalidate` when cache manager enabled

## Route Workflow Matrix

| Route | Policy stack | Handler flow | Side effects |
|---|---|---|---|
| `GET /api/v1/home/dashboard` | `AuthRequired`, optional `CacheRead(home.dashboard)`, optional `CacheControl` | `Dashboard` handler -> service aggregate read -> repo dashboard composition | user-scoped cache (`VaryBy.UserID`) |
| `GET /api/v1/home/projects` | `AuthRequired`, optional `CacheRead(home.projects)`, optional `CacheControl` | `ListProjects` handler reads query pagination -> service -> repo list | user-scoped cache; query vary (`limit`,`offset`) |
| `GET /api/v1/home/projects/reference` | `AuthRequired`, optional `CacheRead(home.reference)`, optional `CacheControl` | `ProjectReference` handler -> service -> repo lightweight project reference set | user-scoped cache |
| `GET /api/v1/home/invites` | `AuthRequired`, optional `CacheRead(home.invites)`, optional `CacheControl` | `ListInvites` handler -> service -> repo invite reads | user-scoped cache |
| `GET /api/v1/home/notifications` | `AuthRequired`, optional `CacheRead(home.notifications)`, optional `CacheControl` | `ListNotifications` handler reads `limit` -> service -> repo | user-scoped cache; query vary (`limit`) |
| `GET /api/v1/home/activity` | `AuthRequired`, optional `CacheRead(home.activity)`, optional `CacheControl` | `ListActivity` handler reads filters -> service -> repo activity feed read | user-scoped cache; query vary (`limit`,`type`,`projectId`) |
| `GET /api/v1/home/dashboard-activity` | `AuthRequired`, optional `CacheRead(home.dashboard_activity)`, optional `CacheControl` | `DashboardActivity` handler -> service -> repo reduced feed | user-scoped cache; query vary (`limit`) |
| `GET /api/v1/home/account` | `AuthRequired`, optional `CacheRead(home.account)`, optional `CacheControl` | `GetAccountSettings` handler -> service -> repo/account read | user-scoped cache |
| `GET /api/v1/home/docs` | `AuthRequired`, optional `CacheRead(home.docs)`, optional `CacheControl` | `Docs` handler -> service static/doc payload assembly | user-scoped cache |
| `POST /api/v1/home/projects` | `AuthRequired`, `RequireJSON`, optional `RateLimit(ScopeUser 10/min)`, optional `CacheInvalidate` | `CreateProject` handler -> service write workflow -> `store.WithTx(...)` -> repo creates project + membership seed | project rows created; home dashboard/projects/reference cache invalidated |
| `POST /api/v1/home/invites/{inviteId}/accept` | `AuthRequired`, optional `RateLimit(ScopeUser 30/min)`, optional `CacheInvalidate` | `AcceptInvite` handler parses `inviteId` -> service transactional accept -> repo update | invite/member state mutation; dashboard/invites/projects cache invalidated |
| `POST /api/v1/home/invites/{inviteId}/decline` | `AuthRequired`, optional `RateLimit(ScopeUser 30/min)`, optional `CacheInvalidate` | `DeclineInvite` handler parses `inviteId` -> service transactional decline -> repo update | invite state mutation; invites/dashboard cache invalidated |
| `PUT /api/v1/home/account` | `AuthRequired`, `RequireJSON`, optional `RateLimit(ScopeUser 20/min)`, optional `CacheInvalidate` | `UpdateAccountSettings` handler validates DTO -> service transactional patch -> repo update | account settings updated; account/dashboard cache invalidated |

## Data Flow Notes

1. All handlers remain transport-only.
2. All writes are service-owned transactions (`store.WithTx(...)`).
3. Repository owns SQL mapping and persistence semantics.

## Troubleshooting Scenarios

1. Home read route stale data:
- check cache tags and user-scoped vary dimensions.
2. Invite accept/decline looks successful but dashboard not updated:
- verify cache invalidation tags applied.
3. Account update rejected:
- validate enum fields (`theme`,`density`,`landing`,`timeFormat`).

## What To Check During Changes

- Preserve user-scoped cache isolation on authenticated reads.
- Preserve transaction boundary for all write routes.
- Keep invite path param extraction and error mapping explicit.
