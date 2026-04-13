# Artifacts Module Workflows

Module path: `internal/modules/artifacts`

## Route Inventory

Stories:
1. `GET /api/v1/projects/{projectId}/stories`
2. `POST /api/v1/projects/{projectId}/stories`
3. `GET /api/v1/projects/{projectId}/stories/{slug}`
4. `PUT /api/v1/projects/{projectId}/stories/{storyId}`

Journeys:
5. `GET /api/v1/projects/{projectId}/journeys`
6. `POST /api/v1/projects/{projectId}/journeys`
7. `GET /api/v1/projects/{projectId}/journeys/{slug}`
8. `PUT /api/v1/projects/{projectId}/journeys/{journeyId}`

Problems:
9. `GET /api/v1/projects/{projectId}/problems`
10. `POST /api/v1/projects/{projectId}/problems`
11. `GET /api/v1/projects/{projectId}/problems/{slug}`
12. `PUT /api/v1/projects/{projectId}/problems/{problemId}`
13. `POST /api/v1/projects/{projectId}/problems/{problemId}/lock`
14. `PUT /api/v1/projects/{projectId}/problems/{problemId}/status`

Ideas:
15. `GET /api/v1/projects/{projectId}/ideas`
16. `POST /api/v1/projects/{projectId}/ideas`
17. `GET /api/v1/projects/{projectId}/ideas/{slug}`
18. `PUT /api/v1/projects/{projectId}/ideas/{ideaId}`
19. `POST /api/v1/projects/{projectId}/ideas/{ideaId}/select`
20. `PUT /api/v1/projects/{projectId}/ideas/{ideaId}/status`

Tasks:
21. `GET /api/v1/projects/{projectId}/tasks`
22. `POST /api/v1/projects/{projectId}/tasks`
23. `GET /api/v1/projects/{projectId}/tasks/{slug}`
24. `PUT /api/v1/projects/{projectId}/tasks/{taskId}`
25. `PUT /api/v1/projects/{projectId}/tasks/{taskId}/status`

Feedback:
26. `GET /api/v1/projects/{projectId}/feedback`
27. `POST /api/v1/projects/{projectId}/feedback`
28. `GET /api/v1/projects/{projectId}/feedback/{slug}`
29. `PUT /api/v1/projects/{projectId}/feedback/{feedbackId}`

## Common Policy Chain

Read routes:
1. `AuthRequired`
2. `ProjectRequired`
3. `ProjectMatchFromPath("projectId")`
4. `ResolvePermissions`
5. `RequirePermission(...)`
6. `CacheReadOptional(...)`
7. `CacheControlOptional(...)`

Write routes:
1. `AuthRequired`
2. `ProjectRequired`
3. `ProjectMatchFromPath("projectId")`
4. `ResolvePermissions`
5. `RequireJSON` for body routes
6. `RequirePermission(...)`
7. `CacheInvalidateOptional(...)`

## Core Flow

For all routes:
1. Handler in `handler.go` extracts path/query/body.
2. Service in `service.go` enforces business/state-transition rules.
3. Repository in `repo.go` executes relational + document operations.

For writes:
- service opens `store.WithTx(...)` for transaction scope.
- repo persists relational row updates and associated document revisions.

## Storage Behavior

Hybrid storage is used:
- relational tables hold indexable artifact state/links
- Mongo document collections hold rich detail payloads and revisions

## Cache Behavior

Read tags are artifact-family specific:
- `artifacts.story`
- `artifacts.journey`
- `artifacts.problem`
- `artifacts.idea`
- `artifacts.task`
- `artifacts.feedback`
- plus `artifacts.project`

All writes invalidate artifact-family tags for project scope.

## RBAC Summary

- Story/Journey routes use story permissions.
- Problem routes use problem permissions.
- Idea routes use idea permissions.
- Task routes use task permissions.
- Feedback routes use feedback permissions.
- status/lock/select routes use status-change permission variants.

## Route Workflow Matrix

| Route | Required permission | Policy details | Handler-service-repo flow |
|---|---|---|---|
| `GET /api/v1/projects/{projectId}/stories` | `PermStoryView` | auth -> project -> resolver -> RBAC -> optional cache read (`artifacts.story`) -> optional cache control | `ListStories` handler reads filters -> service read orchestration -> repo list query |
| `POST /api/v1/projects/{projectId}/stories` | `PermStoryCreate` | auth -> project -> resolver -> `RequireJSON` -> RBAC -> optional cache invalidate | `CreateStory` handler -> service write rules -> `store.WithTx(...)` -> repo relational + document writes |
| `GET /api/v1/projects/{projectId}/stories/{slug}` | `PermStoryView` | auth -> project -> resolver -> RBAC -> optional cache read (`artifacts.story`) -> optional cache control | `GetStory` handler -> service lookup -> repo slug read + detail document |
| `PUT /api/v1/projects/{projectId}/stories/{storyId}` | `PermStoryEdit` | auth -> project -> resolver -> `RequireJSON` -> RBAC -> optional cache invalidate | `UpdateStory` handler -> service transition rules -> transactional repo update |
| `GET /api/v1/projects/{projectId}/journeys` | `PermStoryView` | same as story reads with `artifacts.journey` tags | `ListJourneys` handler -> service -> repo list |
| `POST /api/v1/projects/{projectId}/journeys` | `PermStoryCreate` | same as story creates | `CreateJourney` handler -> service -> transactional repo write |
| `GET /api/v1/projects/{projectId}/journeys/{slug}` | `PermStoryView` | same as journey reads with slug vary | `GetJourney` handler -> service -> repo detail read |
| `PUT /api/v1/projects/{projectId}/journeys/{journeyId}` | `PermStoryEdit` | same as journey writes | `UpdateJourney` handler -> service -> transactional repo update |
| `GET /api/v1/projects/{projectId}/problems` | `PermProblemView` | auth -> project -> resolver -> RBAC -> optional cache read (`artifacts.problem`) -> optional cache control | `ListProblems` handler -> service -> repo list |
| `POST /api/v1/projects/{projectId}/problems` | `PermProblemCreate` | auth -> project -> resolver -> `RequireJSON` -> RBAC -> optional cache invalidate | `CreateProblem` handler -> service -> transactional relational/document create |
| `GET /api/v1/projects/{projectId}/problems/{slug}` | `PermProblemView` | problem read chain with slug vary | `GetProblem` handler -> service -> repo read |
| `PUT /api/v1/projects/{projectId}/problems/{problemId}` | `PermProblemEdit` | write chain with JSON | `UpdateProblem` handler -> service validation -> transactional update |
| `POST /api/v1/projects/{projectId}/problems/{problemId}/lock` | `PermProblemStatusChange` | write chain without JSON | `LockProblem` handler -> service lock-state workflow -> transactional update |
| `PUT /api/v1/projects/{projectId}/problems/{problemId}/status` | `PermProblemStatusChange` | write chain with JSON | `UpdateProblemStatus` handler -> service transition map check -> transactional update |
| `GET /api/v1/projects/{projectId}/ideas` | `PermIdeaView` | auth -> project -> resolver -> RBAC -> optional cache read (`artifacts.idea`) -> optional cache control | `ListIdeas` handler -> service -> repo list |
| `POST /api/v1/projects/{projectId}/ideas` | `PermIdeaCreate` | write chain with JSON | `CreateIdea` handler -> service -> transactional create |
| `GET /api/v1/projects/{projectId}/ideas/{slug}` | `PermIdeaView` | idea read chain with slug vary | `GetIdea` handler -> service -> repo detail read |
| `PUT /api/v1/projects/{projectId}/ideas/{ideaId}` | `PermIdeaEdit` | write chain with JSON | `UpdateIdea` handler -> service immutable-state check -> transactional update |
| `POST /api/v1/projects/{projectId}/ideas/{ideaId}/select` | `PermIdeaStatusChange` | write chain without JSON | `SelectIdea` handler -> service enforces linked locked problem precondition -> transactional update |
| `PUT /api/v1/projects/{projectId}/ideas/{ideaId}/status` | `PermIdeaStatusChange` | write chain with JSON | `UpdateIdeaStatus` handler -> service transition validation -> transactional update |
| `GET /api/v1/projects/{projectId}/tasks` | `PermTaskView` | auth -> project -> resolver -> RBAC -> optional cache read (`artifacts.task`) -> optional cache control | `ListTasks` handler -> service -> repo list |
| `POST /api/v1/projects/{projectId}/tasks` | `PermTaskCreate` | write chain with JSON | `CreateTask` handler -> service -> transactional create |
| `GET /api/v1/projects/{projectId}/tasks/{slug}` | `PermTaskView` | task read chain with slug vary | `GetTask` handler -> service -> repo detail read |
| `PUT /api/v1/projects/{projectId}/tasks/{taskId}` | `PermTaskEdit` | write chain with JSON | `UpdateTask` handler -> service immutable/validation checks -> transactional update |
| `PUT /api/v1/projects/{projectId}/tasks/{taskId}/status` | `PermTaskStatusChange` | write chain with JSON | `UpdateTaskStatus` handler -> service transition rules -> transactional update |
| `GET /api/v1/projects/{projectId}/feedback` | `PermFeedbackView` | auth -> project -> resolver -> RBAC -> optional cache read (`artifacts.feedback`) -> optional cache control | `ListFeedback` handler reads filters -> service -> repo list |
| `POST /api/v1/projects/{projectId}/feedback` | `PermFeedbackCreate` | write chain with JSON | `CreateFeedback` handler -> service -> transactional create |
| `GET /api/v1/projects/{projectId}/feedback/{slug}` | `PermFeedbackView` | feedback read chain with slug vary | `GetFeedback` handler -> service -> repo detail read |
| `PUT /api/v1/projects/{projectId}/feedback/{feedbackId}` | `PermFeedbackEdit` | write chain with JSON | `UpdateFeedback` handler -> service outcome/state rules -> transactional update |

## Route-Specific Notes

- Problem lock endpoint is write path without JSON body.
- Idea select endpoint has business precondition requiring linked locked problem.
- Immutable-state protections are enforced in service before repo update.
- Status transition validity is enforced by service transition maps.

## Troubleshooting Scenarios

1. Idea select returns 400 unexpectedly:
- verify idea is linked to a locked problem.
2. Update rejected for locked/archived entities:
- check immutable-state guard rules in service.
3. Data mismatch between list and detail:
- verify document revision upsert on write paths.
4. Cache appears stale after write:
- check artifact tag invalidation path params (`projectId`).

## What To Check During Changes

- Keep transition matrices in service and not in handler.
- Keep writes inside `WithTx` scopes.
- Keep cache tag specs aligned with route path params.
- Keep document write + relational state update consistency.
