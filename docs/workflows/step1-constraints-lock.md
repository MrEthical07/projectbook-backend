# Step 1 Constraint Lock

Date: 2026-04-18
Scope: Step 1-28 refactor execution guardrails

## Locked constraints

1. API envelope remains canonical (`success`, `data`, `error`, `meta`) with snake_case `request_id` only.
2. Auth token contract remains canonical (`access_token`, `refresh_token`, `access_expires_unix`).
3. Project root dashboard consumption is limited to summary + my-work slices; timeline data stays on dedicated routes.
4. List endpoints in scope use cursor pagination semantics (`items` + optional `next_cursor`) with default limit 20 and max 100.
5. DELETE endpoints do not accept request bodies; selectors are path/query based.
6. Client query caching uses explicit key parts and project-scope invalidation, including project-switch cache clearing.
7. Route call budgets are enforced by script gate:
   - home navigation <= 2 remote calls
   - project navigation <= 3 remote calls
8. Step order remains strict: steps 1-26 complete before step 27 gates, then step 28 report.

## Verification commands

- Backend tests: `go test ./... -count=1`
- Web checks: `pnpm check`
- Contract checks only: `pnpm check:contracts`
- Route budget checks only: `pnpm check:call-budgets`

## Lock status

This lock is active for the current branch state and is referenced by `STEP_1_28_AUDIT_REPORT.md`.
