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

## Email is disabled by default

Self-host runs with `EMAIL_ENABLED=false`, so no Resend key is required and no
verification/reset emails are sent. If your build gates app access behind email
verification, either enable email (set `EMAIL_ENABLED=true` + `RESEND_API_KEY` and
the sender addresses from `.env.example`), or mark your account verified directly
in Postgres:

```bash
docker compose exec postgres psql -U projectbook -d projectbook \
  -c "UPDATE users SET status='active' WHERE email='you@example.com';"
```

> Confirm the exact column your build uses for verification state against
> `db/migrations/000003_auth_users.up.sql` before relying on this.

## Notes & limitations

- The API image is distroless (no shell), so it has no container-level health
  check; the web app simply starts after the API container starts and connects on
  first request. If the API is still booting, a refresh resolves it.
- First build compiles the Go binary and runs a full `pnpm` web build, so the
  initial `up` takes a few minutes. Subsequent runs are cached.
- This compose targets local/self-host use. For a public deployment, put it behind
  a reverse proxy with TLS, set strong secrets, and review CORS `allowedOrigins`.

<!-- TODO (follow-up): seed a demo project + pre-verified demo user on first run so
     new self-hosters land on populated data instead of an empty workspace. Needs a
     small seed step (SQL or an API-based seeder) validated against the schema. -->
