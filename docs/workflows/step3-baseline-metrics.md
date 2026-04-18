# Step 3 Baseline Metrics Snapshot

Date: 2026-04-18
Workspace: d:/Files/League/ProjectBook
Purpose: capture a single, auditable baseline artifact for call counts, latency placeholders, cache placeholders, and list payload sizing notes.

## Call-count baseline (scripted)

Source: `Web/scripts/check-route-call-budgets.mjs`

- home-navigation: 2 calls (budget 2)
- project-navigation: 3 calls (budget 3)

## Latency placeholders

No dedicated synthetic load run was available in this branch snapshot.
Latency placeholders are locked for future comparison:

- project dashboard summary load p50: TBD
- project dashboard summary load p95: TBD
- project timeline page load p50: TBD
- project timeline page load p95: TBD

## Cache placeholders

Runtime cache telemetry counters are not persisted in this branch snapshot.
Placeholders are locked for later before/after deltas:

- cache hit count: TBD
- cache miss count: TBD
- cache hit ratio: TBD

## Payload size by list endpoint (schema-level baseline)

This baseline records enforced list contract shape and pagination semantics for tracked list endpoints.
All listed endpoints now enforce `items` + optional `next_cursor` with limit default 20 and max 100.

- `/projects/{projectId}/activity`: cursor list contract active
- `/projects/{projectId}/calendar`: cursor list contract active
- `/projects/{projectId}/stories`: cursor list contract active
- `/projects/{projectId}/journeys`: cursor list contract active
- `/projects/{projectId}/problems`: cursor list contract active
- `/projects/{projectId}/ideas`: cursor list contract active
- `/projects/{projectId}/tasks`: cursor list contract active
- `/projects/{projectId}/feedback`: cursor list contract active
- `/projects/{projectId}/pages`: cursor list contract active
- `/projects/{projectId}/resources`: cursor list contract active

## Notes

- This artifact is intentionally conservative: where runtime numbers are unavailable in-repo, placeholders are explicit rather than implied.
- Gate results used for this snapshot:
  - `go test ./... -count=1` passed
  - `pnpm check` passed
