# Environment Variables

This is the canonical runtime configuration reference for ProjectBook Backend.

Source of truth:
- `internal/core/config/config.go`
- `internal/core/config/profile.go`

## 1. Resolution Order

For each variable, resolution order is:
1. explicit environment variable
2. profile default (`APP_PROFILE`)
3. hard-coded fallback in config loader

Startup always runs `Config.Lint()` and fails fast on invalid values and invalid feature combinations.

## 2. Profiles (`APP_PROFILE`)

Supported profiles:
- `minimal`
- `dev`
- `prod`

Profile behavior summary:

- `minimal`:
  - disables auth/cache/rate-limit/permissions
  - keeps Postgres/Redis/Mongo enabled
- `dev`:
  - enables auth/cache/rate-limit/permissions
  - `AUTH_MODE=jwt_only`
  - fail-open defaults for cache/rate-limit
- `prod`:
  - enables auth/cache/rate-limit/permissions
  - `AUTH_MODE=strict`
  - fail-closed defaults for cache/rate-limit

Explicit env values override profile defaults.

## 3. Core Application

| Variable | Default |
|---|---|
| `APP_PROFILE` | empty |
| `APP_ENV` | `dev` |
| `APP_SERVICE_NAME` | `api-template` |

## 4. HTTP Transport

| Variable | Default |
|---|---|
| `HTTP_ADDR` | `:8080` |
| `HTTP_READ_HEADER_TIMEOUT` | `5s` |
| `HTTP_READ_TIMEOUT` | `15s` |
| `HTTP_WRITE_TIMEOUT` | `15s` |
| `HTTP_IDLE_TIMEOUT` | `60s` |
| `HTTP_SHUTDOWN_TIMEOUT` | `10s` |
| `HTTP_MAX_HEADER_BYTES` | `1048576` |

## 5. Global Middleware

### 5.1 Middleware toggles and limits

| Variable | Default |
|---|---|
| `HTTP_MIDDLEWARE_REQUEST_ID_ENABLED` | `true` |
| `HTTP_MIDDLEWARE_RECOVERER_ENABLED` | `true` |
| `HTTP_MIDDLEWARE_MAX_BODY_BYTES` | `1048576` |
| `HTTP_MIDDLEWARE_SECURITY_HEADERS_ENABLED` | `true` in prod, else `false` |
| `HTTP_MIDDLEWARE_REQUEST_TIMEOUT` | `10s` |
| `HTTP_MIDDLEWARE_TRACING_EXCLUDE_PATHS` | `/healthz,/readyz,/metrics` |

### 5.2 Access log

| Variable | Default |
|---|---|
| `HTTP_MIDDLEWARE_ACCESS_LOG_ENABLED` | `true` |
| `HTTP_MIDDLEWARE_ACCESS_LOG_SAMPLE_RATE` | `0.05` |
| `HTTP_MIDDLEWARE_ACCESS_LOG_EXCLUDE_PATHS` | `/healthz,/readyz,/metrics` |
| `HTTP_MIDDLEWARE_ACCESS_LOG_SLOW_THRESHOLD` | `2s` |
| `HTTP_MIDDLEWARE_ACCESS_LOG_INCLUDE_USER_AGENT` | `false` |
| `HTTP_MIDDLEWARE_ACCESS_LOG_INCLUDE_REMOTE_IP` | `false` |

### 5.3 Client IP

| Variable | Default |
|---|---|
| `HTTP_TRUSTED_PROXIES` | empty |

### 5.4 CORS

| Variable | Default |
|---|---|
| `HTTP_MIDDLEWARE_CORS_ENABLED` | `true` |
| `HTTP_MIDDLEWARE_CORS_ALLOW_ORIGINS` | empty |
| `HTTP_MIDDLEWARE_CORS_DENY_ORIGINS` | empty |
| `HTTP_MIDDLEWARE_CORS_ALLOW_METHODS` | empty |
| `HTTP_MIDDLEWARE_CORS_ALLOW_HEADERS` | empty |
| `HTTP_MIDDLEWARE_CORS_EXPOSE_HEADERS` | empty |
| `HTTP_MIDDLEWARE_CORS_ALLOW_CREDENTIALS` | `false` |
| `HTTP_MIDDLEWARE_CORS_MAX_AGE` | `0` |
| `HTTP_MIDDLEWARE_CORS_ALLOW_PRIVATE_NETWORK` | `false` |

## 6. Logging

| Variable | Default |
|---|---|
| `LOG_LEVEL` | `info` |
| `LOG_FORMAT` | `json` |

## 7. Authentication

| Variable | Default |
|---|---|
| `AUTH_ENABLED` | `true` |
| `AUTH_MODE` | `hybrid` |
| `AUTH_TEST_SHARED_SECRET` | empty |
| `PROJECTBOOK_PERMISSION_CONTEXT_SECRET` | `projectbook-dev-permission-context-secret` |

`AUTH_MODE` valid values:
- `jwt_only`
- `hybrid`
- `strict`

Auth lint dependencies:
- auth enabled requires redis enabled
- auth enabled requires postgres enabled

## 8. Rate Limiting

| Variable | Default |
|---|---|
| `RATELIMIT_ENABLED` | `true` |
| `RATELIMIT_FAIL_OPEN` | `true` non-prod, `false` prod |
| `RATELIMIT_DEFAULT_LIMIT` | `10` |
| `RATELIMIT_DEFAULT_WINDOW` | `1m` |

Lint dependency:
- rate-limit enabled requires redis enabled

## 9. Caching

| Variable | Default |
|---|---|
| `CACHE_ENABLED` | `true` |
| `CACHE_FAIL_OPEN` | `true` non-prod, `false` prod |
| `CACHE_DEFAULT_MAX_BYTES` | `262144` |
| `CACHE_TAG_VERSION_CACHE_TTL` | `250ms` |

Lint dependency:
- cache enabled requires redis enabled

## 10. Permissions

| Variable | Default |
|---|---|
| `PERMISSIONS_ENABLED` | `true` |
| `PERMISSIONS_DB_QUERY_TIMEOUT` | `750ms` |
| `PERMISSIONS_REDIS_TTL` | `6h` |
| `PERMISSIONS_BACKFILL_TIMEOUT` | `500ms` |

Lint dependency:
- permissions enabled requires postgres enabled

## 11. Postgres

| Variable | Default |
|---|---|
| `POSTGRES_ENABLED` | `true` |
| `POSTGRES_URL` | `postgres://superapi:superapi@127.0.0.1:5432/superapi?sslmode=disable` |
| `DATABASE_URL` | alias fallback |
| `POSTGRES_MAX_CONNS` | `10` |
| `POSTGRES_MIN_CONNS` | `0` |
| `POSTGRES_CONN_MAX_LIFETIME` | `30m` |
| `POSTGRES_CONN_MAX_IDLE_TIME` | `5m` |
| `POSTGRES_STARTUP_PING_TIMEOUT` | `3s` |
| `POSTGRES_HEALTH_CHECK_TIMEOUT` | `1s` |

## 12. Redis

| Variable | Default |
|---|---|
| `REDIS_ENABLED` | `true` |
| `REDIS_ADDR` | `127.0.0.1:6379` |
| `REDIS_URL` | alias fallback |
| `REDIS_PASSWORD` | empty |
| `REDIS_DB` | `0` |
| `REDIS_DIAL_TIMEOUT` | `2s` |
| `REDIS_READ_TIMEOUT` | `2s` |
| `REDIS_WRITE_TIMEOUT` | `2s` |
| `REDIS_POOL_SIZE` | `10` |
| `REDIS_MIN_IDLE_CONNS` | `0` |
| `REDIS_STARTUP_PING_TIMEOUT` | `3s` |
| `REDIS_HEALTH_CHECK_TIMEOUT` | `1s` |

## 13. Mongo

| Variable | Default |
|---|---|
| `MONGO_ENABLED` | `true` |
| `MONGO_URL` | `mongodb://127.0.0.1:27017` |
| `MONGO_DB` | `projectbook` |
| `MONGO_MAX_POOL_SIZE` | `50` |
| `MONGO_MIN_POOL_SIZE` | `0` |
| `MONGO_CONNECT_TIMEOUT` | `5s` |
| `MONGO_STARTUP_PING_TIMEOUT` | `3s` |
| `MONGO_HEALTH_CHECK_TIMEOUT` | `1s` |
| `MONGO_BOOTSTRAP_ENABLED` | `true` |
| `MONGO_BOOTSTRAP_TIMEOUT` | `10s` |

## 14. Metrics

| Variable | Default |
|---|---|
| `METRICS_ENABLED` | `true` |
| `METRICS_PATH` | `/metrics` |
| `METRICS_AUTH_TOKEN` | empty |
| `METRICS_EXCLUDE_PATHS` | `/healthz,/readyz` |

Prod lint rule:
- if metrics enabled in prod, auth token must be non-empty

## 15. Tracing

| Variable | Default |
|---|---|
| `TRACING_ENABLED` | `false` |
| `TRACING_SERVICE_NAME` | empty (falls back to app service name) |
| `TRACING_EXPORTER` | `otlpgrpc` |
| `TRACING_OTLP_ENDPOINT` | `localhost:4317` |
| `TRACING_SAMPLER` | `traceidratio` |
| `TRACING_SAMPLE_RATIO` | `0.05` |
| `TRACING_INSECURE` | `true` non-prod, `false` prod |

## 16. Production Constraints

When `APP_ENV=prod` or `production`:
- cache fail-open cannot be enabled
- rate-limit fail-open cannot be enabled
- metrics auth token required when metrics enabled

## 17. Troubleshooting

### 17.1 Startup config lint failure

Check:
1. invalid duration/int/bool parsing
2. missing dependency toggles for enabled features
3. prod-only fail-open restrictions

### 17.2 Auth routes failing at startup

Check:
1. `AUTH_ENABLED=true`
2. Redis/Postgres enabled and reachable
3. valid `AUTH_MODE`

### 17.3 Cache or rate-limit unexpectedly inactive

Check:
1. `CACHE_ENABLED` / `RATELIMIT_ENABLED`
2. Redis enabled + reachable
3. profile overrides versus explicit env values

## 18. Related Documents

- [architecture.md](architecture.md)
- [auth.md](auth.md)
- [policies.md](policies.md)
- [cache-guide.md](cache-guide.md)
- [workflows/README.md](workflows/README.md)
