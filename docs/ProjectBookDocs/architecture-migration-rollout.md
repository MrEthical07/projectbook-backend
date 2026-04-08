# Architecture Migration Rollout

Date: 2026-04-08
Status: Completed
Owner: Copilot refactor execution

## Goal

Remove legacy tenant-oriented architecture from runtime, contracts, and documentation.
Final model:
- Users are global identities.
- Projects are top-level entities.
- Access is based on project membership and role permissions only.
- Product UX, APIs, and schemas use home/project terminology.

## Non-Goals

- Re-platforming to a production database in this change set.
- Changing artifact domain behavior beyond access and ownership model updates.

## Global Safety Rules

- Every phase must end with `pnpm check` and no new diagnostics.
- Keep compatibility shims only during active migration phases.
- Remove compatibility shims in the final phase.

## Phase Plan

### Phase 1: Baseline and Safety Harness
Deliverables:
- Baseline diagnostics recorded.
- Migration checklist committed.
Validation:
- `pnpm check` passes.

### Phase 2: Authorization Hardening
Deliverables:
- Access resolution no longer uses fallback role resolution from legacy collections.
- Authenticated user resolution no longer mutates legacy tenant-scoped user state.
Validation:
- Project access uses session principal + project membership only.
- `pnpm check` passes.

### Phase 3: Remove Client `actorId` Trust
Deliverables:
- Mutation handlers derive actor from authenticated request user.
- `actorId` is removed or ignored from trust decisions.
Validation:
- Permission and ownership decisions are session-based.
- `pnpm check` passes.

### Phase 4: Introduce User-Home Boundary
Deliverables:
- New user-home remote boundary replaces legacy boundary semantics.
- Home loaders read from user-home APIs.
Validation:
- Home pages work without legacy semantic coupling.
- `pnpm check` passes.

### Phase 5: Refactor Project/Artifact Remotes
Deliverables:
- Project and artifact remotes no longer reference legacy project/user collections.
- Shared helpers use project store + team membership only.
Validation:
- Artifact CRUD and project settings paths still function.
- `pnpm check` passes.

### Phase 6: Route and UI Migration
Deliverables:
- Legacy wording removed from UI pages and component labels.
- Route loaders and pages no longer import deprecated remote APIs.
Validation:
- Home/auth/invites/projects/activity/account UX uses project-centric terminology.
- `pnpm check` passes.

### Phase 7: Shared Types and Data Shape Migration
Deliverables:
- Legacy-prefixed app types replaced with user/project-centric names.
- Datastore shape no longer includes the legacy tenant branch.
Validation:
- Compile-time type usage updated across routes/remotes.
- `pnpm check` passes.

### Phase 8: API Contract Migration
Deliverables:
- OpenAPI and API guidelines moved to home/project-centric endpoints.
- Legacy schema names replaced.
Validation:
- Spec paths and schema names have no legacy tenant references.

### Phase 9: Data Model and Architecture Docs Migration
Deliverables:
- Database and architecture docs no longer describe legacy tenancy assumptions.
- Relationships and index plans updated for project-top-level model.
Validation:
- No remaining legacy-architecture claims in core design docs.

### Phase 10: Remove Compatibility Layer and Dead Code
Deliverables:
- Deprecated remote boundary and related seed modules removed.
- Transitional aliases removed.
Validation:
- No runtime imports of deprecated module names.
- `pnpm check` passes.

## Exit Criteria

- Runtime, type system, API docs, and architecture docs contain no legacy tenant dependency.
- All protected operations are session-principal based and project membership based.
- `pnpm check` passes at completion.