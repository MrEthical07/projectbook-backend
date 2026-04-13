# Auth Module Workflows

Module path: `internal/modules/auth`

## Route Inventory

1. `POST /api/v1/auth/signup`
2. `POST /api/v1/auth/login`
3. `POST /api/v1/auth/verify-email`
4. `POST /api/v1/auth/resend-verification`
5. `POST /api/v1/auth/forgot-password`
6. `POST /api/v1/auth/reset-password`
7. `POST /api/v1/auth/refresh`
8. `POST /api/v1/auth/logout`

## Route Workflows

### `POST /api/v1/auth/signup`

Purpose:
- create account via goAuth and persist display-name update.

Policy chain:
1. `RequireJSON`
2. `RateLimitWithKeyer(..., "auth.signup", ScopeIP, KeyByIP())` when limiter enabled

Flow:
1. Handler `Signup` validates `signupRequest`.
2. Service `Signup` calls `engine.CreateAccount(...)`.
3. Service updates name with repo `UpdateUserName(...)`.
4. Response returns user envelope.

Transaction:
- no explicit service transaction wrapper.

Side effects:
- auth account created in auth persistence layer
- rate-limit bucket incremented by IP

Failure mapping:
- existing account -> `409`
- invalid payload/password policy -> `400`
- rate limited -> `429`
- auth dependency failure -> `503`

### `POST /api/v1/auth/login`

Policy chain:
1. `RequireJSON`
2. `RateLimitWithKeyer(..., "auth.login", ScopeIP, KeyByIP())` when enabled

Flow:
1. Handler `Login` -> Service `Login`.
2. Service calls `engine.Login(identifier,password)`.
3. Service parses JWT `exp` claim and returns token response.

Side effects:
- session/auth state updated by goAuth

Failures:
- invalid credentials -> `401`
- rate-limited -> `429`

### `POST /api/v1/auth/verify-email`

Policy chain:
1. `RequireJSON`
2. `RateLimitWithKeyer(..., "auth.verify_email", ScopeIP, KeyByIP())` when enabled

Flow:
1. Handler `VerifyEmail` -> Service `VerifyEmail`.
2. Service calls `engine.ConfirmEmailVerification(token)`.

Failures:
- invalid token -> `400`
- rate-limited -> `429`

### `POST /api/v1/auth/resend-verification`

Policy chain:
1. `RequireJSON`
2. `RateLimitWithKeyer(..., "auth.resend_verification", ScopeIP, KeyByIP())` when enabled

Flow:
1. Handler `ResendVerification` -> Service `ResendVerification`.
2. Service calls `engine.RequestEmailVerification(email)`.

### `POST /api/v1/auth/forgot-password`

Policy chain:
1. `RequireJSON`
2. `RateLimitWithKeyer(..., "auth.forgot_password", ScopeIP, KeyByIP())` when enabled

Flow:
1. Handler `ForgotPassword` -> Service `ForgotPassword`.
2. Service calls `engine.RequestPasswordReset(email)`.

### `POST /api/v1/auth/reset-password`

Policy chain:
1. `RequireJSON`
2. `RateLimitWithKeyer(..., "auth.reset_password", ScopeIP, KeyByIP())` when enabled

Flow:
1. Handler `ResetPassword` -> Service `ResetPassword`.
2. Service calls `engine.ConfirmPasswordReset(token,password)`.

### `POST /api/v1/auth/refresh`

Policy chain:
1. `RequireJSON`
2. `RateLimitWithKeyer(..., "auth.refresh", ScopeIP, KeyByIP())` when enabled

Flow:
1. Handler `Refresh` -> Service `Refresh`.
2. Service calls `engine.Refresh(refreshToken)`.
3. Service returns normalized access/refresh payload.

Note:
- compatibility endpoint retained for tooling.

### `POST /api/v1/auth/logout`

Policy chain:
1. `AuthRequired(engine, mode)`
2. `RateLimitWithKeyer(..., "auth.logout", ScopeUser, KeyByUserOrProjectOrTokenHash(16))` when enabled

Flow:
1. Handler reads bearer token from Authorization header.
2. Service `Logout` calls `engine.LogoutByAccessToken(token)`.

Side effects:
- refresh/session invalidation in goAuth backing store.

## Troubleshooting Scenarios

1. 429 on repeated signup/login:
- Check auth IP limiter keys and traffic profile.
2. logout returns 401:
- Check Authorization header format and token freshness.
3. verify/reset loops failing:
- Check token expiry and exact token string trimming.

## What To Check During Changes

- Keep public auth routes JSON + IP rate-limited.
- Keep logout authenticated.
- Do not add project/RBAC policies to public auth entry routes.
- Keep auth service error mapping consistent with API error codes.
