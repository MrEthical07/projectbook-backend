# Calendar Module Workflows

Module path: `internal/modules/calendar`

## Route Inventory

1. `GET /api/v1/projects/{projectId}/calendar`
2. `POST /api/v1/projects/{projectId}/calendar`
3. `GET /api/v1/projects/{projectId}/calendar/{eventId}`
4. `PUT /api/v1/projects/{projectId}/calendar/{eventId}`
5. `DELETE /api/v1/projects/{projectId}/calendar/{eventId}`

## Policy Chain

All routes:
1. `AuthRequired`
2. `ProjectRequired`
3. `ProjectMatchFromPath("projectId")`
4. `ResolvePermissions`
5. RBAC permission by operation

Write routes also include:
- `RequireJSON` for create/update

## RBAC Mapping

- list/get -> `PermCalendarView`
- create -> `PermCalendarCreate`
- update -> `PermCalendarEdit`
- delete -> `PermCalendarDelete`

## Flow

- Handler delegates to service.
- Service delegates to repo.
- Service wraps writes in `store.WithTx(...)`.
- Repository persists calendar event rows in relational store.

## Derived Events

`GET /calendar` returns two kinds of event:

- **Manual** events, stored in `calendar_events` and fully editable via the
  create/update/delete routes.
- **Derived** events (`type = "Derived"`), projected at query time from other
  artifacts and **not** stored:
  - task deadlines — one per task with a non-null `due_at` (`artifactType = "Task"`).
  - captured feedback — one per `Active` feedback row (`artifactType = "Feedback"`).

  Derived events are read-only: they carry the source artifact's id, so
  get/update/delete on that id return `404` from the calendar module (edit the
  source task/feedback instead). Because they are computed from the source on
  every read, they can never go stale relative to it.

## Side Effects

- no route cache policies currently configured.
- no document store path for calendar module.
- notification fan-out (`notify_artifact_created`) on manual event creation.

## Route Workflow Matrix

| Route | Required permission | Policy details | Handler-service-repo flow |
|---|---|---|---|
| `GET /api/v1/projects/{projectId}/calendar` | `PermCalendarView` | auth -> project -> resolver -> RBAC | `ListCalendarData` handler -> service read path -> repo list query |
| `POST /api/v1/projects/{projectId}/calendar` | `PermCalendarCreate` | auth -> project -> resolver -> `RequireJSON` -> RBAC | `CreateCalendarEvent` handler -> service validation -> `store.WithTx(...)` -> repo insert |
| `GET /api/v1/projects/{projectId}/calendar/{eventId}` | `PermCalendarView` | auth -> project -> resolver -> RBAC | `GetCalendarEvent` handler parses `eventId` -> service -> repo read |
| `PUT /api/v1/projects/{projectId}/calendar/{eventId}` | `PermCalendarEdit` | auth -> project -> resolver -> `RequireJSON` -> RBAC | `UpdateCalendarEvent` handler -> service update rules -> transactional repo update |
| `DELETE /api/v1/projects/{projectId}/calendar/{eventId}` | `PermCalendarDelete` | auth -> project -> resolver -> RBAC | `DeleteCalendarEvent` handler -> service delete workflow -> transactional repo delete |

## Troubleshooting Scenarios

1. create/update returns 400:
- verify date/time fields and enum values in request DTO validation.
2. unauthorized/forbidden on calendar routes:
- verify project scope chain and calendar permission bits.
3. delete returns not found after successful create:
- verify event id path parameter mapping and repository identity resolution.

## What To Check During Changes

- keep project policy chain order unchanged
- keep write transactions in service
- keep calendar permission mapping explicit per operation
