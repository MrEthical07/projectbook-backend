# Policy Reference

Policies are per-route middleware functions applied during route registration.

Type:

```go
type Policy func(http.Handler) http.Handler
```

Core files:

- `internal/core/policy/*.go`
- `internal/core/policy/validator_rules.go`
- `internal/tools/validator/analyzer.go`

## Strict guarantees

ProjectBook backend enforces fail-fast policy invariants.

- Every `r.Handle(...)` policy stack is validated at registration.
- Invalid order/dependencies panic with `invalid route config: ...`.
- Static verification (`go run ./cmd/superapi-verify ./...`) applies the same checks.

## Required protected-route order

Use this order for project-protected routes:

1. auth
2. project
3. resolve permissions
4. rbac
5. rate limit
6. cache
7. cache-control (optional)

Example:

```go
r.Handle(http.MethodGet, "/api/v1/projects/{project_id}/tasks/{id}", handler,
    policy.AuthRequired(authEngine, auth.ModeStrict),
    policy.ProjectRequired(),
    policy.ProjectMatchFromPath("project_id"),
    policy.ResolvePermissions(permissionResolver),
    policy.RequirePermission(rbac.PermProjectView),
    policy.RateLimit(limiter, ratelimit.Rule{Limit: 60, Window: time.Minute, Scope: ratelimit.ScopeProject}),
    policy.CacheRead(cacheMgr, cache.CacheReadConfig{
        TTL: 30 * time.Second,
        TagSpecs: []cache.CacheTagSpec{{Name: "task", PathParams: []string{"id"}}},
        VaryBy: cache.CacheVaryBy{ProjectID: true, PathParams: []string{"id"}},
        AllowAuthenticated: true,
    }),
)
```

## Auth and project policies

### `AuthRequired(engine, mode)`

Behavior:

- Missing/invalid bearer token -> `401 unauthorized`
- Successful validation injects `auth.AuthContext`

Current auth context fields used by policies:

- `UserID`
- `ProjectID`
- `Role`
- `PermissionMask`
- `Permissions`

Project scope authority is request-derived (`project_id`/`projectId` path params), not auth provider tenant metadata.

### `ProjectRequired()`

Behavior:

- Missing auth context -> `401 unauthorized`
- Requires project scope from request path (`project_id` / `projectId`)
- Missing scope -> `403 forbidden` (`project scope required`)
- Path/auth conflict -> `403 forbidden` (`project scope mismatch`)

### `ProjectMatchFromPath(paramName)`

Behavior:

- Missing auth context -> `401 unauthorized`
- Missing path param -> `400 bad_request`
- Missing project scope -> `403 forbidden`
- Project mismatch -> `403 forbidden` (`project scope mismatch`)

## Permission resolver policy

### `ResolvePermissions(resolver)`

Resolves effective project permission state before RBAC checks.

Behavior:

- Missing auth context -> `401 unauthorized`
- Missing project scope -> `403 forbidden`
- Membership missing -> `403 forbidden`
- Membership exists but mask inconsistent -> `500 internal_error`
- Resolver dependency failure -> `503 dependency_unavailable`
- Success injects resolved `PermissionMask` (and role when present)

Resolver contract:

```go
type Resolver interface {
    Resolve(ctx context.Context, userID, projectID string) (Resolution, error)
}
```

## RBAC policies

RBAC here is mask/bit based.

### `RequirePermission(perm uint64)`

Requires one permission bit.

### `RequireAnyPermission(perms ...uint64)`

Requires any one bit.

### `RequireAllPermissions(perms ...uint64)`

Requires every listed bit.

Shared behavior:

- Missing permission -> `403 forbidden`
- Constructors panic on invalid input (zero/empty permission sets)

## Rate limit policy

### `RateLimit(limiter, rule)`

Standard route-level limiting.

### `RateLimitWithKeyer(limiter, name, rule, keyer)`

Custom keyer path for fine-grained identity shaping.

Scopes:

- `ScopeAuto`
- `ScopeAnon`
- `ScopeIP`
- `ScopeUser`
- `ScopeProject`
- `ScopeToken`

Useful built-in keyers:

- `ratelimit.KeyByIP()`
- `ratelimit.KeyByUser()`
- `ratelimit.KeyByProject()`
- `ratelimit.KeyByTokenHash(prefixLen)`
- `ratelimit.KeyByUserOrProjectOrTokenHash(prefixLen)`

## Cache policies

### `CacheRead(manager, cfg)`

Read-through cache policy with dynamic tag-version token integration.

Authenticated safety rule (enforced by validator):

- Authenticated routes must vary by identity via `VaryBy.UserID` or `VaryBy.ProjectID`.

Important config surfaces:

- `TTL`
- `TagSpecs`
- `VaryBy`
- `AllowAuthenticated`
- `Methods`
- `CacheStatuses`

### `CacheInvalidate(manager, cfg)`

Bumps tag versions after successful write responses (`2xx`) so matching reads miss and refresh.

## Cache-Control policy

### `CacheControl(cfg)`

Adds explicit `Cache-Control` (and optional `Vary`) directives for browser/proxy behavior.

Keep this after auth/project/resolver/rbac/rate-limit/cache policies.

## Utility policies

- `RequireJSON()`
- `WithHeader(key, value)`
- `Noop()`

## Validator rules summary

Current enforced rules include:

- Policy order must be monotonic by stage.
- `AuthRequired` is required when project/resolver/RBAC policies are present.
- `ProjectMatchFromPath` requires `ProjectRequired`.
- `ResolvePermissions` requires `ProjectRequired`.
- RBAC policies require `ResolvePermissions`.
- Routes containing `{project_id}` must include:
  - `ProjectRequired`
  - `ProjectMatchFromPath("project_id")`
- Authenticated `CacheRead` must vary by user or project identity.

## Presets

Prefer validated presets for common stacks:

- `policy.ProjectRead(...)`
- `policy.ProjectWrite(...)`
- `policy.PublicRead(...)`

## Verification commands

```bash
go test ./...
go build ./...
go run ./cmd/superapi-verify ./...
```
