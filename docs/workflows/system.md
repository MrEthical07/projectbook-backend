# System Module Workflows

Module path: `internal/modules/system`

## Route Inventory

1. `POST /system/parse-duration`
2. `GET /api/v1/system/whoami`

## Workflow Details

### `POST /system/parse-duration`

Policy chain:
- `RequireJSON`

Flow:
1. Router dispatches to `parseDuration` handler.
2. Request DTO validates that `duration` is non-empty.
3. Handler parses duration using Go `time.ParseDuration`.
4. Handler returns:
- canonical normalized duration string
- nanoseconds and milliseconds values

Side effects:
- none
- no DB/cache/rate-limit/auth dependencies

Failure mapping:
- invalid or empty duration -> `400`

### `GET /api/v1/system/whoami`

Policy chain:
- `AuthRequired`
- `RateLimitWithKeyer(..., "system.whoami", ScopeUser, KeyByUserOrProjectOrTokenHash(16))` when limiter enabled

Flow:
1. Router dispatches to `whoami` handler.
2. `AuthRequired` injects principal context.
3. Handler reads auth context and returns identity payload:
- `user_id`
- `project_id`
- `role`
- `permission_mask`
- `permissions`

Side effects:
- rate-limit bucket increment for authenticated caller (when limiter enabled)
- no DB writes

Failure mapping:
- missing/invalid auth -> `401`
- rate-limited -> `429`

## Route Workflow Matrix

| Route | Policy details | Handler-service flow | Side effects |
|---|---|---|---|
| `POST /system/parse-duration` | `RequireJSON` | `parseDuration` handler validates DTO -> parses duration -> returns normalized output | none |
| `GET /api/v1/system/whoami` | `AuthRequired`, optional route-specific user-scoped rate-limit | `whoami` handler reads auth principal from context and returns identity payload | optional rate-limit bucket increments |

## Troubleshooting Scenarios

1. parse-duration returns 400 unexpectedly:
- verify payload is JSON and `duration` is provided (for example `"15m"`, `"1h30m"`).
2. whoami returns 401:
- verify bearer token and auth mode wiring.
3. whoami returns 429:
- inspect `system.whoami` rate-limit counters.
4. route not reachable:
- verify module registration in `internal/modules/modules.go`
- verify route mounting for both `/system/*` and `/api/v1/system/*`
5. route returns 404:
- verify route was not disabled behind build/runtime flags

## What To Check During Changes

- keep `parse-duration` deterministic and side-effect free
- keep `whoami` auth-protected and rate-limited when limiter is enabled
- preserve response shape for tooling that introspects identity context
