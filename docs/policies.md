# Policies

This document is the canonical reference for route policies in ProjectBook Backend.

Policy type:

```go
type Policy func(http.Handler) http.Handler
```

Core policy files:
- `internal/core/policy/*.go`
- `internal/core/policy/validator.go`
- `internal/core/policy/validator_rules.go`

Static verification tool:
- `cmd/superapi-verify`

## 1. Why Policies Exist

Policies provide explicit, per-route behavioral guarantees for:
- authentication
- project isolation
- permission resolution
- RBAC enforcement
- rate limiting
- cache correctness
- response cache directives

Route behavior is declared where the route is registered, not hidden in global middleware.

## 2. Enforced Ordering Model

Protected route stage order is enforced:

1. auth
2. project scope
3. permission resolver
4. RBAC
5. rate limit
6. cache read/invalidate
7. cache-control

If a route violates ordering/dependency rules, route registration fails with an invalid route config panic.

## 3. Validation Rules (Fail-Fast)

Validator checks include:
- policy stages cannot regress in order
- RBAC/project/resolver stacks require `AuthRequired`
- RBAC checks require `ResolvePermissions`
- `ProjectMatchFromPath` requires `ProjectRequired`
- routes with `{project_id}` or `{projectId}` must include project policies
- authenticated `CacheRead` must vary by `UserID` or `ProjectID`

These rules are applied:
- at runtime route registration
- by static verifier (`go run ./cmd/superapi-verify ./...`)

## 4. Policy Catalog

### 4.1 `AuthRequired(engine, mode)`

Purpose:
- enforce bearer authentication
- inject `auth.AuthContext`

Behavior:
- missing/invalid auth -> `401`
- mode-aware guard behavior (`jwt_only`, `hybrid`, `strict`)

Key implementation:
- `internal/core/policy/auth.go`

### 4.2 `ProjectRequired()`

Purpose:
- enforce project scope presence

Behavior:
- missing auth context -> `401`
- missing project scope in request path -> `403`
- scope mismatch -> `403`

### 4.3 `ProjectMatchFromPath(paramName)`

Purpose:
- bind resolved principal project to a specific route path param

Behavior:
- missing path param -> `400`
- mismatch -> `403`

Typical path param in this repo: `projectId`

### 4.4 `ResolvePermissions(resolver)`

Purpose:
- resolve effective permission mask for user + project before RBAC checks

Behavior:
- no membership -> `403`
- inconsistent mask -> `500`
- resolver dependency failure -> `503`

### 4.5 RBAC policies

- `RequirePermission(perm)`
- `RequireAnyPermission(perms...)`
- `RequireAllPermissions(perms...)`

Behavior:
- missing required permission(s) -> `403`

### 4.6 Rate limit policies

- `RateLimit(limiter, rule)`
- `RateLimitWithKeyer(limiter, name, rule, keyer)`

Behavior:
- over budget -> `429` (+ `Retry-After` when available)
- limiter dependency failure in fail-closed path -> `503`

### 4.7 Cache policies

- `CacheRead(manager, cfg)`
- `CacheInvalidate(manager, cfg)`
- `CacheControl(cfg)`

Optional wrappers used in modules that tolerate cache manager absence:
- `CacheReadOptional(...)`
- `CacheInvalidateOptional(...)`
- `CacheControlOptional(...)`

### 4.8 Utility policies

- `RequireJSON()`
- `WithHeader(key, value)`
- `Noop()`

## 5. Policy Config Parameters

### 5.1 Auth policy

Inputs:
- auth engine pointer
- auth mode (`jwt_only`, `hybrid`, `strict`)

### 5.2 Rate limit policy

`ratelimit.Rule` key fields:
- `Limit`
- `Window`
- `Scope`
- optional `Keyer`

Common scopes in this API:
- IP for public auth routes
- User for user-account scoped routes
- Project for project mutation routes
- user/project/token-hash fallback for selected endpoints

### 5.3 Cache read config

Important fields:
- `TTL`
- `TagSpecs`
- `VaryBy`
- `AllowAuthenticated`
- `Methods`
- `CacheStatuses`
- optional `FailOpen`

### 5.4 Cache invalidate config

Important fields:
- `TagSpecs`

Invalidation occurs only for successful write responses (`2xx`) after downstream handler execution.

## 6. Real Route Stack Patterns

### 6.1 Project-scoped read route (strict chain)

Typical chain:
1. `AuthRequired`
2. `ProjectRequired`
3. `ProjectMatchFromPath("projectId")`
4. `ResolvePermissions`
5. `RequirePermission(...)`
6. `CacheRead` or `CacheReadOptional`
7. `CacheControl` or `CacheControlOptional`

### 6.2 Project-scoped write route (strict chain)

Typical chain:
1. `AuthRequired`
2. `ProjectRequired`
3. `ProjectMatchFromPath("projectId")`
4. `ResolvePermissions`
5. `RequirePermission(...)` or `RequireAnyPermission(...)`
6. `RequireJSON` (if body)
7. `RateLimit` (when enabled)
8. `CacheInvalidate` or `CacheInvalidateOptional`

### 6.3 Public auth route chain

Typical chain:
1. `RequireJSON`
2. `RateLimitWithKeyer(..., KeyByIP())` when limiter exists

No project/resolver/RBAC policies are applied to public auth entry routes.

## 7. Where To Change Policy Behavior

- Metadata and policy types:
  - `internal/core/policy/metadata.go`
- Validator rules and stage model:
  - `internal/core/policy/validator_rules.go`
- Auth/project policies:
  - `internal/core/policy/auth.go`
- Permission resolver policy:
  - `internal/core/policy/permissions.go`
- RBAC policies:
  - `internal/core/policy/rbac.go`
- Rate limit policy wrappers:
  - `internal/core/policy/ratelimit.go`
- Cache policies:
  - `internal/core/policy/cache.go`
  - `internal/core/policy/cache_optional.go`
- Route stacks (module usage):
  - `internal/modules/*/routes.go`

## 8. Troubleshooting Guide

### 8.1 Panic: invalid route config during startup

Check:
1. policy order for the failing route
2. missing required dependencies (auth, resolver, project matcher)
3. authenticated cache route missing `VaryBy.UserID` or `VaryBy.ProjectID`

### 8.2 Protected project route returns 401

Check:
1. bearer token parsing and auth mode
2. auth engine wiring in app dependencies
3. `AuthRequired` included in route chain

### 8.3 Route returns 403 despite valid login

Check:
1. `ProjectRequired` / `ProjectMatchFromPath` path param alignment
2. membership exists for `user + project`
3. resolver output and effective permission mask
4. RBAC permission bit required by route

### 8.4 Route returns 503 for resolver/cache/rate-limit

Check:
1. Redis availability for cache/rate-limit/resolver hybrid path
2. fail-open/fail-closed config
3. dependency startup/health probes

### 8.5 JSON endpoint returns 415

Check:
1. `Content-Type: application/json`
2. route uses `RequireJSON`
3. client payload type and method

## 9. Operational Best Practices

- Keep policy stacks explicit per route.
- Favor canonical chain order even when optional components are disabled.
- Do not move RBAC checks into handlers/services.
- For authenticated cache routes, always vary by identity dimension.
- Keep route policy changes paired with verifier + integration test runs.

## 10. Verification Commands

```bash
go test ./...
go build ./...
go run ./cmd/superapi-verify ./...
```

## 11. Related Documents

- [architecture.md](architecture.md)
- [auth.md](auth.md)
- [cache-guide.md](cache-guide.md)
- [environment-variables.md](environment-variables.md)
- [workflows/README.md](workflows/README.md)
