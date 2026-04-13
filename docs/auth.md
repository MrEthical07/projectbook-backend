# Authentication

This is the single canonical authentication document for ProjectBook Backend.

It explains:
- auth primitives used by this API
- how goAuth is integrated
- runtime auth modes
- the exact auth configuration this API uses
- route protection and troubleshooting

## 1. Auth Architecture In This API

Auth responsibility is intentionally split:

- goAuth handles authentication mechanics:
  - account creation and login
  - JWT access token issuance and validation
  - refresh session lifecycle
  - email verification
  - password reset
- ProjectBook handles authorization separately:
  - project scope enforcement
  - permission resolution
  - RBAC bitmask checks

Core code paths:
- `internal/core/auth/goauth_provider.go`
- `internal/core/auth/provider.go`
- `internal/core/auth/provider_store.go`
- `internal/core/auth/user_repository.go`
- `internal/core/policy/auth.go`
- `internal/modules/auth/routes.go`
- `internal/modules/auth/service.go`

## 2. Auth Primitives Used By The API

### 2.1 Request auth context

When `policy.AuthRequired(...)` succeeds, it injects `auth.AuthContext` into request context with:

- `UserID`
- `ProjectID`
- `Role`
- `PermissionMask`
- `Permissions`

This context is consumed by:
- `policy.ProjectRequired()`
- `policy.ProjectMatchFromPath(...)`
- `policy.ResolvePermissions(...)`
- `policy.RequirePermission(...)`
- handlers such as `system.whoami`

### 2.2 Tokens and sessions

- Access token:
  - JWT
  - short-lived
  - used as `Authorization: Bearer <token>`
- Refresh token:
  - long-lived session token
  - used by `/api/v1/auth/refresh`
  - revocation-aware via goAuth session subsystem

### 2.3 Data primitives

The API persists auth users through repository + store contracts:

- Repository contract: `UserRepository`
- Primary record model: `StoredUser`
- Required repository methods:
  - `GetByIdentifier`
  - `GetByID`
  - `UpdatePasswordHash`
  - `Create`
  - `UpdateStatus`

No module should bypass this contract and hit DB driver APIs directly.

## 3. goAuth Integration

### 3.1 Engine construction

Engine initialization happens in `auth.NewGoAuthEngine(...)` and is wired from app dependency setup.

High-level startup flow:

1. Parse auth mode (`jwt_only`, `hybrid`, `strict`).
2. Build store-backed user provider (`NewStoreUserProvider`).
3. Build goAuth engine with ProjectBook config.
4. Expose engine in app dependencies for route policies and auth module service.

### 3.2 Store-backed user provider

`StoreUserProvider` adapts ProjectBook `UserRepository` to goAuth `UserProvider`.

This keeps boundaries clean:
- goAuth depends on an abstract provider contract
- ProjectBook owns persistence via repo and store contracts
- no goAuth code reaches database drivers directly

### 3.3 MFA and backup code status

MFA and backup-code methods are intentionally stubbed in `provider_store.go` and return unauthorized until MFA persistence is implemented.

That behavior is expected for this backend state.

## 4. Auth Modes

Configured via `AUTH_MODE` and normalized in `internal/core/auth/provider.go`.

Supported values:
- `jwt_only`
- `hybrid` (default)
- `strict`

Operational semantics:

- `jwt_only`:
  - validates JWT claims/signature only
  - lowest dependency on session backing checks for protected-route validation
- `hybrid`:
  - balanced default
  - uses goAuth hybrid validation behavior
- `strict`:
  - strongest validation mode
  - intended for strict revocation/session-aware behavior

Default mode is `hybrid` when empty.

## 5. ProjectBook goAuth Configuration

ProjectBook config is set in `projectBookGoAuthConfig(mode)`.

### 5.1 Token and identity defaults

- `JWT.AccessTTL = 5m`
- `JWT.RefreshTTL = 7d`
- `JWT.Issuer = projectbook`
- `JWT.Audience = projectbook-api`
- `JWT.KeyID = v1`

### 5.2 Result and security behavior

- `Result.IncludeRole = true`
- `Result.IncludePermissions = false`
- `Security.EnablePermissionVersionCheck = false`
- `Security.EnableRoleVersionCheck = false`
- `Security.EnforceRefreshRotation = true`
- `Security.EnforceRefreshReuseDetection = true`
- `Security.EnableLoginFailureLimiter = true`
- `Security.EnableIPBinding = false`
- `Security.EnableIPSignal = false`
- `Security.ProductionMode = false`

### 5.3 Session and feature flags

- `Session.SlidingExpiration = true`
- `Session.AbsoluteSessionLifetime = 7d`
- `DeviceBinding.Enabled = false`
- `MultiTenant.Enabled = false`
- `Permission.MaxBits = 64`
- `Permission.RootBitReserved = false`

### 5.4 Auth flows enabled

- `PasswordReset.Enabled = true`
- `PasswordReset.ResetTTL = 1h`
- `EmailVerification.Enabled = true`
- `EmailVerification.RequireForLogin = true`
- `EmailVerification.VerificationTTL = 24h`
- `Account.Enabled = true`
- `Account.DefaultRole = user`

### 5.5 Test signer override

If `AUTH_TEST_SHARED_SECRET` is set:

- signing method switches to HS256
- both private/public key config are set from the shared secret bytes

This is for deterministic local/perf test behavior.

## 6. Auth Route Surface

Auth routes are registered in `internal/modules/auth/routes.go`.

| Method | Path | AuthRequired | JSON required | Rate limit |
|---|---|---|---|---|
| POST | `/api/v1/auth/signup` | no | yes | `20/min` by IP |
| POST | `/api/v1/auth/login` | no | yes | `30/min` by IP |
| POST | `/api/v1/auth/verify-email` | no | yes | `20/min` by IP |
| POST | `/api/v1/auth/resend-verification` | no | yes | `20/min` by IP |
| POST | `/api/v1/auth/forgot-password` | no | yes | `15/min` by IP |
| POST | `/api/v1/auth/reset-password` | no | yes | `15/min` by IP |
| POST | `/api/v1/auth/refresh` | no | yes | `45/min` by IP |
| POST | `/api/v1/auth/logout` | yes | no | `60/min` by user/project/token hash |

Note:
- `/api/v1/auth/refresh` exists as a compatibility endpoint and is intentionally retained.

## 7. Policy Patterns For Auth Routes

### 7.1 Public auth routes

Pattern:
1. `RequireJSON`
2. `RateLimitWithKeyer(..., KeyByIP())` when limiter exists

### 7.2 Protected auth route (`logout`)

Pattern:
1. `AuthRequired(engine, mode)`
2. `RateLimitWithKeyer(..., KeyByUserOrProjectOrTokenHash(16))` when limiter exists

Why this differs:
- Public entry routes need anti-abuse by client identity before authentication exists.
- Logout is authenticated and keyed by user-scoped identity dimensions.

## 8. Service-Level Behavior

Auth module service (`internal/modules/auth/service.go`) delegates to goAuth engine methods:

- Signup -> `CreateAccount`
- Login -> `Login`
- Refresh -> `Refresh`
- Logout -> `LogoutByAccessToken`
- VerifyEmail -> `ConfirmEmailVerification`
- ResendVerification -> `RequestEmailVerification`
- ForgotPassword -> `RequestPasswordReset`
- ResetPassword -> `ConfirmPasswordReset`

Error mapping is normalized into API error codes and HTTP statuses using `map*Error(...)` helpers in service.

## 9. Config And Lint Requirements

Environment variables:
- `AUTH_ENABLED` (default `true`)
- `AUTH_MODE` (default `hybrid`)
- `AUTH_TEST_SHARED_SECRET` (optional)

Startup lint constraints:
- if auth enabled, Redis must be enabled
- if auth enabled, Postgres must be enabled

These constraints are enforced at startup and fail fast.

## 10. Troubleshooting

### 10.1 All protected routes return 401

Check:
1. bearer token is present and formatted as `Bearer ...`
2. auth engine was initialized successfully during startup
3. selected `AUTH_MODE` matches expected runtime behavior
4. system clock skew is not invalidating JWT expiry checks

### 10.2 Login/signup flows return 429 unexpectedly

Check:
1. rate-limit is enabled and route-specific limits are low by design
2. client/IP re-use in tests is saturating IP-scoped buckets
3. Redis limiter keys for auth routes if running local stress loops

### 10.3 Logout appears successful but token still works briefly

Check:
1. understand JWT access token TTL window (`5m`)
2. strictness mode behavior (`hybrid` vs `strict`)
3. session invalidation happened in goAuth backend

### 10.4 Startup fails around auth

Check:
1. `AUTH_ENABLED`, `AUTH_MODE` values are valid
2. Redis/Postgres are reachable and enabled
3. auth config lint errors in startup logs
4. key/signing configuration for environment

## 11. Where To Change Auth Behavior

- Mode parsing/defaults:
  - `internal/core/auth/provider.go`
- Engine config values:
  - `internal/core/auth/goauth_provider.go`
- User persistence contract:
  - `internal/core/auth/user_repository.go`
- Provider bridge for goAuth:
  - `internal/core/auth/provider_store.go`
- Route-level auth policies and rate limits:
  - `internal/modules/auth/routes.go`
- API-facing auth behavior and error mapping:
  - `internal/modules/auth/service.go`

## 12. Related Docs

- [architecture.md](architecture.md)
- [policies.md](policies.md)
- [environment-variables.md](environment-variables.md)
- [cache-guide.md](cache-guide.md)
- [workflows/README.md](workflows/README.md)
