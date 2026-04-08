# Remote Functions

## What A Remote File Is

A remote file (`*.remote.ts`) is the execution boundary between routes and data.

- Reads use `query(...)`.
- Writes use `command(...)`.
- Both operate over datastore-backed domain state.

## Rules

- No service layer under remote.
- No command-pattern wrapper around remote.
- No remote chaining for domain actions.
- No cross-imports between domain remotes.

Practical note:

- `sidebar.remote.ts` currently imports `access.remote.ts` to build combined sidebar/access behavior. Treat this as a boundary composition exception, not a pattern to expand.

## Read Function Structure

Standard read path:

1. Parse/normalize input scope (for example, `projectId`).
2. Guard existence and visibility constraints.
3. Read from datastore collections.
4. Return page-specific payload shape.

## Write Function Structure

Standard write path:

1. Parse input with Zod.
2. Check action permission for the domain.
3. Validate status transitions and linked references.
4. Update row state and detail cache state.
5. Return `{ success, data | error }`.

## Validation With Zod

Every mutation command parses unknown input using a schema before state mutation. This blocks malformed payloads and keeps mutation contracts explicit.

## Permission Checks Inside Remote

All writes enforce permissions in remote, not in page code. UI checks are UX controls; remote checks are authority controls.

## Remote As Future API Boundary

Current implementation is in-process and datastore-backed. The remote contract is intentionally shaped so it can map to external API handlers later without changing route mental models.
