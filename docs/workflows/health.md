# Health Module Workflows

Module path: `internal/modules/health`

## Route Inventory

1. `GET /healthz`
2. `GET /readyz`

## Workflow Details

### `GET /healthz`

Policy chain:
- none (public route)

Flow:
1. Router dispatches directly to `healthz` handler.
2. Handler returns static payload `{status: "ok"}`.

Side effects:
- none
- no store access
- no cache/rate-limit/auth policy

### `GET /readyz`

Policy chain:
- none (public route)

Flow:
1. Router dispatches to `readyz` handler.
2. Handler reads readiness service report (if wired).
3. Handler observes readiness metrics (if metrics service wired).
4. Handler returns:
   - `200` when readiness status is ready
   - `503` when readiness status is not ready

Side effects:
- readiness metrics observation
- no DB mutation

## Route Workflow Matrix

| Route | Policy details | Handler flow | Output |
|---|---|---|---|
| `GET /healthz` | public, no policies | `healthz` handler returns static map | `200` with `{status:"ok"}` |
| `GET /readyz` | public, no policies | `readyz` handler builds readiness report, optionally runs readiness checks, observes readiness metrics | `200` when ready, `503` when not ready |

## Troubleshooting Scenarios

1. `/healthz` failing:
- check process is running and HTTP server bound correctly.
2. `/readyz` returns `503`:
- inspect readiness dependencies (`postgres`, `redis`, `mongo`) and their health checks.
3. readiness intermittency:
- inspect dependency timeouts and startup probes in runtime config.

## What To Check During Changes

- keep health routes public
- keep readiness status mapping (`ready` -> 200, otherwise 503)
- avoid adding business/domain logic to health module
