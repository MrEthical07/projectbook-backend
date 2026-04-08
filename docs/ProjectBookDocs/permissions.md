# Permissions

## Permission Model

ProjectBook uses a mask-based RBAC model.

- `permissionMask` is the authorization source of truth.
- Role defaults are stored as `rolePermissionMasks`.
- Each member has `role`, `permissionMask`, and `isCustom`.

Canonical reference: [rbac.md](rbac.md)

## RBAC Resolution

Effective access is resolved as:

```text
if isCustom == true:
	effectiveMask = member.permissionMask
else:
	effectiveMask = rolePermissionMasks[member.role]
```

## Domains and Actions

The model still uses 10 domains x 6 actions, but capability checks are done via mask bits.

Domains:

- `project`
- `story`
- `problem`
- `idea`
- `task`
- `feedback`
- `resource`
- `page`
- `calendar`
- `member`

Actions:

- `view`
- `create`
- `edit`
- `delete`
- `archive`
- `statusChange`

Bit allocation is fixed and append-only:

```text
bit = domain_index * 6 + action_index
```

## Helper Functions

Permission evaluation and mutation use mask helpers:

- `hasPerm(mask, domainIndex, actionIndex)`
- `updatePerm(mask, domainIndex, actionIndex, enabled)`
- `applyPermissionDependencyRules(mask, domain, action, enabled)`
- `validatePermissionMaskValue(mask)`

## Validation

- Non-view actions require `view` in the same domain.
- Invalid masks must be rejected by backend.
- Frontend may auto-correct toggles for UX, but backend enforcement is authoritative.

## UI and Backend Responsibilities

- UI should use resolved access state to hide/disable controls.
- Backend resolves actor membership and mask from session/context.
- Remote commands must re-check authorization before mutation.

## Read vs Write vs Status Change

- Read: requires `view` for the domain.
- Write: `create`, `edit`, `delete`, `archive` are independently gated by mask bits.
- Status changes: `statusChange` is separate from `edit`.

## UI Behavior Patterns

- Hide: actions not relevant to the user should not be shown.
- Disable: visible controls can be disabled when action is blocked.
- Block: route-level access denial when `view` is missing.
