# RBAC (Permission Mask Model)

This document defines the breaking RBAC model for ProjectBook.

RBAC is now mask-based and uses a fixed 64-bit strategy.

## 1) Source Of Truth

- Frontend mapping/constants: [src/lib/constants/permissions.ts](src/lib/constants/permissions.ts)
- Permission helper/conversion functions: [src/lib/utils/permission.ts](src/lib/utils/permission.ts)
- Authorization resolution: [src/lib/server/auth/authorization.ts](src/lib/server/auth/authorization.ts)
- Team/roles API behavior: [src/lib/remote/project.remote.ts](src/lib/remote/project.remote.ts)
- Seed/default role setup: [src/lib/server/data/project.data.ts](src/lib/server/data/project.data.ts)
- Roles page UX: [src/routes/project/[projectId]/team/roles/+page.svelte](src/routes/project/[projectId]/team/roles/+page.svelte)

## 2) Ground Rules (Non-Negotiable)

1. Max = 64 bits (`uint64` compatible space)
2. Never reuse bits
3. Never reorder bits
4. Only append new permissions
5. Frontend + backend must share exact mapping

## 3) Fixed Allocation Strategy

Bit index formula:

```text
bit = domain_index * 6 + action_index
```

### Action index mapping (global)

```text
0 = VIEW
1 = CREATE
2 = EDIT
3 = DELETE
4 = ARCHIVE
5 = STATUS_CHANGE
```

### Domain index mapping (global)

```text
0 = PROJECT
1 = STORY
2 = PROBLEM
3 = IDEA
4 = TASK
5 = FEEDBACK
6 = RESOURCE
7 = PAGE
8 = CALENDAR
9 = MEMBER
```

Current system uses bits `0..59` (60 permissions). Bits `60..63` are reserved for future append-only growth.

## 4) Full Bit Map

### PROJECT (0-5)

```go
PermProjectView         = 1 << 0
PermProjectCreate       = 1 << 1
PermProjectEdit         = 1 << 2
PermProjectDelete       = 1 << 3
PermProjectArchive      = 1 << 4
PermProjectStatusChange = 1 << 5
```

### STORY (6-11)

```go
PermStoryView         = 1 << 6
PermStoryCreate       = 1 << 7
PermStoryEdit         = 1 << 8
PermStoryDelete       = 1 << 9
PermStoryArchive      = 1 << 10
PermStoryStatusChange = 1 << 11
```

### PROBLEM (12-17)

```go
PermProblemView         = 1 << 12
PermProblemCreate       = 1 << 13
PermProblemEdit         = 1 << 14
PermProblemDelete       = 1 << 15
PermProblemArchive      = 1 << 16
PermProblemStatusChange = 1 << 17
```

### IDEA (18-23)

```go
PermIdeaView         = 1 << 18
PermIdeaCreate       = 1 << 19
PermIdeaEdit         = 1 << 20
PermIdeaDelete       = 1 << 21
PermIdeaArchive      = 1 << 22
PermIdeaStatusChange = 1 << 23
```

### TASK (24-29)

```go
PermTaskView         = 1 << 24
PermTaskCreate       = 1 << 25
PermTaskEdit         = 1 << 26
PermTaskDelete       = 1 << 27
PermTaskArchive      = 1 << 28
PermTaskStatusChange = 1 << 29
```

### FEEDBACK (30-35)

```go
PermFeedbackView         = 1 << 30
PermFeedbackCreate       = 1 << 31
PermFeedbackEdit         = 1 << 32
PermFeedbackDelete       = 1 << 33
PermFeedbackArchive      = 1 << 34
PermFeedbackStatusChange = 1 << 35
```

### RESOURCE (36-41)

```go
PermResourceView         = 1 << 36
PermResourceCreate       = 1 << 37
PermResourceEdit         = 1 << 38
PermResourceDelete       = 1 << 39
PermResourceArchive      = 1 << 40
PermResourceStatusChange = 1 << 41
```

### PAGE (42-47)

```go
PermPageView         = 1 << 42
PermPageCreate       = 1 << 43
PermPageEdit         = 1 << 44
PermPageDelete       = 1 << 45
PermPageArchive      = 1 << 46
PermPageStatusChange = 1 << 47
```

### CALENDAR (48-53)

```go
PermCalendarView         = 1 << 48
PermCalendarCreate       = 1 << 49
PermCalendarEdit         = 1 << 50
PermCalendarDelete       = 1 << 51
PermCalendarArchive      = 1 << 52
PermCalendarStatusChange = 1 << 53
```

### MEMBER (54-59)

```go
PermMemberView         = 1 << 54
PermMemberCreate       = 1 << 55
PermMemberEdit         = 1 << 56
PermMemberDelete       = 1 << 57
PermMemberArchive      = 1 << 58
PermMemberStatusChange = 1 << 59
```

## 5) Frontend Mapping Contract

Do not hardcode duplicate mappings in random places.

Use domain/action arrays and compute bit from indexes.

```ts
const ACTIONS = ["view", "create", "edit", "delete", "archive", "status"];

const DOMAINS = [
  "project",
  "story",
  "problem",
  "idea",
  "task",
  "feedback",
  "resource",
  "page",
  "calendar",
  "member"
];

function getBit(domainIndex: number, actionIndex: number) {
  return 1n << BigInt(domainIndex * 6 + actionIndex);
}
```

In codebase naming, action key is `statusChange` while action index label is `status`.

## 6) Required Helper Functions

Implemented in [src/lib/utils/permission.ts](src/lib/utils/permission.ts):

- `hasPerm(mask, domainIndex, actionIndex)`
- `updatePerm(mask, domainIndex, actionIndex, enabled)`
- `applyPermissionDependencyRules(mask, domain, action, enabled)`
- `maskToPermissions(mask)`
- `permissionsToMask(permissions)`
- `normalizePermissionMask(value)`
- `validatePermissionMask(mask: bigint)`
- `validatePermissionMaskValue(mask)`
- `enforcePermissionMaskDependencies(mask)`

Backend should implement equivalent helpers using the exact same mapping.

## 7) Role Masks (Current Defaults)

These are computed from the canonical default role-mask policy used by the app.

| Role | Decimal mask | Hex mask |
| --- | ---: | ---: |
| Owner | 1152921504606846975 | 0xFFFFFFFFFFFFFFF |
| Admin | 864691128455135221 | 0xBFFFFFFFFFFFFF5 |
| Editor | 20016033248999873 | 0x471C79E79E79C1 |
| Member | 875734824153537 | 0x31C79E71C71C1 |
| Viewer | 18300341342965825 | 0x41041041041041 |
| Limited Access | 0 | 0x0 |

## 8) Member RBAC State

Each team member now carries:

- `role`
- `isCustom`
- `permissionMask`

Resolution rule:

```text
if isCustom == true:
  effectiveMask = member.permissionMask
else:
  effectiveMask = rolePermissionMask[member.role]
```

`isCustom` is true only when member mask differs from role mask.

## 9) Team Roles UX Contract

### 9.1 Member table

- No inline role selector or per-row Save button.
- Row action is `Edit permissions`.
- Role column shows role + `Custom` badge when `isCustom = true`.

### 9.2 Edit permissions dialog

- First control: role selector.
- Then: `Customize permissions?` toggle.
- When enabled: full 60-bit permission grid.
- If custom member toggles custom OFF, show confirmation alert:

```text
This will revert from custom permissions to role specific permissions. Still want to continue?
```

- If member is custom and role changes while custom stays active, show alert:

```text
Custom permissions are active. Role change won't affect permissions. To update permissions, edit the permissions explicitely.
```

### 9.3 Role Permission Mask Editor Location

- Moved below member table.
- On role permission save, show alert:

```text
For the users whos custom permissions are active, Role change won't affect permissions. To update permissions, edit the permissions explicitely.
```

## 10) API Contract (Backend Must Match)

### 10.1 Get team roles

Return for project:

- members with role/isCustom/permissionMask + existing member fields
- rolePermissionMasks map (role -> mask)

### 10.2 Update role permissions

Input:

- `projectId`
- `role`
- `permissionMask`

Rules:

- Owner role mask immutable from this API.
- Non-custom members with that role follow new role mask.
- Custom members remain unchanged.

### 10.3 Update member permissions

Input:

- `projectId`
- `memberId`
- `role`
- `isCustom`
- `permissionMask`

Rules:

- If `isCustom = false`, persisted member mask should be role mask.
- If `isCustom = true`, persisted member mask can differ from role mask.
- If provided custom mask equals role mask, member should be treated as non-custom.

## 11) Migration Notes

- Canonical storage and authorization enforcement are mask-only at role and member level.
- Backend authorization must call `getTrustedProjectPermissionMask(input)` and enforce with `hasPerm(mask, domainIndex, actionIndex)`.
- Object-shaped permissions may exist only as isolated display adapters, never as auth source of truth.
- Any new permission must be appended only (next free bit), never inserted or reordered.

## 12) Permission Mask Integrity Validation (MANDATORY)

### 12.1 Core validation rule

For every domain:

- If any of `create`, `edit`, `delete`, `archive`, or `statusChange` is enabled, then `view` must be enabled.

### 12.2 Shared validation function

Validation is centralized in [src/lib/utils/permission.ts](src/lib/utils/permission.ts):

- `validatePermissionMask(mask: bigint): { valid: boolean; errors: string[] }`

Validation iterates all 10 domains and rejects any domain where non-view actions are enabled while view is disabled.

Error text:

```text
Invalid permission: VIEW is required when other actions are enabled (domain: <domain_name>)
```

### 12.3 Frontend enforcement behavior

Frontend uses dependency-safe toggling in [src/routes/project/[projectId]/team/roles/+page.svelte](src/routes/project/[projectId]/team/roles/+page.svelte):

- Enabling a non-view action auto-enables `view` for that domain.
- Disabling `view` auto-disables `create`, `edit`, `delete`, `archive`, and `statusChange` for that domain.
- UI state hydration normalizes masks through `enforcePermissionMaskDependencies(mask)` to avoid persisting invalid combinations in client state.

### 12.4 API and backend rejection points

Validation is enforced in backend write paths:

- `updateProjectRolePermissions` in [src/lib/remote/project.remote.ts](src/lib/remote/project.remote.ts)
- `updateProjectMemberPermissions` in [src/lib/remote/project.remote.ts](src/lib/remote/project.remote.ts)

Behavior:

- Invalid masks are rejected.
- Backend never silently auto-corrects invalid masks.
- Invalid masks are never persisted to datastore.

### 12.5 Backend-only responsibilities (explicit)

These are backend tasks and are not delegated to frontend:

- Validate masks on every RBAC write path before persistence.
- Validate and normalize stored role masks during project-role-mask resolution in [src/lib/server/auth/authorization.ts](src/lib/server/auth/authorization.ts).
- Ensure project role-mask maps exist for project access resolution, project creation, and invite acceptance using default seeded masks.
- Reject invalid stored masks explicitly with clear error messages.

### 12.6 Performance constraint

- Validation cost is fixed and bounded by 10 domains x 6 actions.
- No unbounded or data-size-dependent loops are used in the mask validator.

## 13) Backend Middleware Contract (Authorization Only)

Backend route authorization is enforced by context-driven mask middleware.

Implementation references:

- `internal/core/policy/rbac.go`
- `internal/core/rbac/permissions.go`

### 13.1 Context contract

RBAC middleware reads auth context populated upstream by auth middleware:

```go
type AuthContext struct {
  UserID         string
  PermissionMask uint64
  Role           string
}
```

If `PermissionMask` is missing, middleware treats it as `0`.

### 13.2 Mandatory helper functions

```go
func GetUserMask(ctx context.Context) uint64
func HasPerm(mask uint64, perm uint64) bool
func HasRole(ctx context.Context, role string) bool
```

Permission check is strictly bitwise:

```go
return mask&perm != 0
```

### 13.3 Policy constructors

```go
func RequirePermission(perm uint64) Policy
func RequireAnyPermission(perms ...uint64) Policy
func RequireAllPermissions(perms ...uint64) Policy
```

Policy behavior:

1. Read mask once from context.
2. Evaluate bits with `HasPerm(...)`.
3. Return `403 Forbidden` on denial.

### 13.4 Non-negotiable constraints

- Middleware is stateless.
- No database/repository calls from middleware.
- No permission resolution inside middleware.
- Role is optional helper only; mask is source of truth.
- Do not leak permission names, required bits, or raw masks in errors.

### 13.5 Route order

Protected route policy order remains:

1. auth
2. tenant
3. rbac (`RequirePermission` / `RequireAnyPermission` / `RequireAllPermissions`)
4. rate limit
5. cache
6. cache-control (optional)
