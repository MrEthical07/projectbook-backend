# Self-hosting ProjectBook

Run the entire stack — web app, Go API, Postgres, MongoDB, Redis — with one command.

## Prerequisites

- Docker + Docker Compose v2
- Both repos cloned **side by side**:

```
your-folder/
├── projectbook/           # frontend (SvelteKit)
└── projectbook-backend/   # backend (Go) — you are here, compose file lives here
```

Clone them if you haven't:

```bash
git clone https://github.com/MrEthical07/projectbook.git
git clone https://github.com/MrEthical07/projectbook-backend.git
```

## Run

```bash
cd projectbook-backend
docker compose up --build
```

That will:

1. Start Postgres, MongoDB, and Redis (with health checks).
2. Run all database migrations (one-shot `migrate` service).
3. Start the Go API on **http://localhost:8080**.
4. Build and serve the web app on **http://localhost:3000**.

Open **http://localhost:3000** and create an account.

To run it in the background: `docker compose up --build -d`. To stop: `docker compose down`. To wipe all data: `docker compose down -v`.

## Configuration

Defaults are baked into `docker-compose.yml` for a frictionless local run. For anything beyond that:

- **Permission-context secret** — change `PROJECTBOOK_PERMISSION_CONTEXT_SECRET` (any 96 hex chars; generate with `openssl rand -hex 48`), e.g.:

  ```bash
  PROJECTBOOK_PERMISSION_CONTEXT_SECRET=$(openssl rand -hex 48) docker compose up --build
  ```

- **Database credentials** — the Postgres user/password/db default to `projectbook`; change them in the `postgres` service and the `POSTGRES_URL` in the `migrate` and `api` services together.

- **Web checkout in a different location** — override the build context:

  ```bash
  WEB_CONTEXT=/path/to/projectbook docker compose up --build
  ```

- **Serving on a different URL** — set `PUBLIC_SITE_URL` to the exact origin you
  open in the browser. It feeds SvelteKit's `ORIGIN`, which `adapter-node` uses to
  validate form POSTs; if it does not match, every login fails with
  *"Cross-site POST form submissions are forbidden"*.

  ```bash
  PUBLIC_SITE_URL=https://projectbook.example.com docker compose up --build
  ```

## Email is disabled by default — and first login needs one extra step

Self-host runs with `EMAIL_ENABLED=false`, so no Resend key is required and no
verification/reset emails are sent.

This matters, because **new accounts start unverified and the web layer redirects
unverified users to `/auth/verify`**. With email disabled, the verification OTP is
never sent, so a fresh signup cannot reach the app on its own.

Concretely:

- Signup writes `users.is_email_verified = false`.
- The API *permits* login while unverified (`EmailVerification.RequireForLogin = false`).
- The SvelteKit layer (`hooks.server.ts`) then redirects any unverified session to
  `/auth/verify`, which is a dead end without email.

### Option A — the bundled `verify` helper (easiest)

The compose file ships a one-shot `verify` service that marks every existing
account as verified. Sign up once at http://localhost:3000, then run:

```bash
docker compose run --rm verify
```

Refresh the page and you're in. Re-run it after each new signup.

It sits behind the `tools` profile, so it never runs as part of `docker compose
up` — only when you invoke it explicitly.

### Option B — verify one account by hand

```bash
docker compose exec postgres psql -U projectbook -d projectbook \
  -c "UPDATE users SET is_email_verified = true WHERE email = 'you@example.com';"
```

The column is `is_email_verified`, added by
`db/migrations/000016_auth_users_schema_reconcile.up.sql`. (The older
`000003_auth_users.up.sql` has only a `status` column, which is **not** the
verification flag — updating `status` will not let you in.)

### Option C — real email

Set `EMAIL_ENABLED=true` plus `RESEND_API_KEY` and the sender addresses from
`.env.example`, and verification works normally.

## Notes & limitations

- The API image is distroless (no shell), so it has no container-level health
  check; the web app simply starts after the API container starts and connects on
  first request. If the API is still booting, a refresh resolves it.
- First build compiles the Go binary and runs a full `pnpm` web build, so the
  initial `up` takes a few minutes. Subsequent runs are cached.
- This compose targets local/self-host use. For a public deployment, put it behind
  a reverse proxy with TLS, set strong secrets, set `PUBLIC_SITE_URL` to your public
  URL, and review CORS `allowedOrigins`.

<!-- TODO (follow-up): seed a demo project + pre-verified demo user on first run so
     new self-hosters land on populated data instead of an empty workspace. Needs a
     small seed step (SQL or an API-based seeder) validated against the schema. -->
