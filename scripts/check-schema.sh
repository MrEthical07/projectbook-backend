#!/usr/bin/env bash
# CI guard against schema drift and dishonest migrations.
#
# 1. Parity: every *.up.sql migration must have a matching *.down.sql.
# 2. Golden diff: applying every migration to an empty database must reproduce
#    db/schema.generated.sql exactly. This catches drift such as a table being
#    defined twice with divergent shapes (as `users` was across 000003/000005),
#    or a migration whose net effect no single file explains.
#
# Self-contained: spins up its own throwaway Postgres via Docker so the pg_dump
# version always matches the server and CI needs no external database. The image
# is pinned so the golden snapshot stays reproducible.
#
# Requires: docker, go.
# Regenerate the golden file after an intentional schema change:
#   ./scripts/check-schema.sh --update
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

PG_IMAGE="${PG_IMAGE:-postgres:16-alpine}"
CONTAINER="pb-schema-check-$$"
HOST_PORT="${HOST_PORT:-55499}"
GOLDEN="db/schema.generated.sql"
IN_CONTAINER_URL="postgres://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable"
HOST_URL="postgres://postgres:postgres@127.0.0.1:${HOST_PORT}/postgres?sslmode=disable"
UPDATE=0
[[ "${1:-}" == "--update" ]] && UPDATE=1

cleanup() { docker rm -f "${CONTAINER}" >/dev/null 2>&1 || true; }
trap cleanup EXIT

echo "==> Checking up/down migration parity"
missing=0
for up in db/migrations/*.up.sql; do
  down="${up%.up.sql}.down.sql"
  if [[ ! -f "${down}" ]]; then
    echo "MISSING down migration for: ${up}"
    missing=1
  fi
done
if [[ "${missing}" -ne 0 ]]; then
  echo "ERROR: every *.up.sql must have a matching *.down.sql"
  exit 1
fi
echo "    ok"

echo "==> Starting throwaway Postgres (${PG_IMAGE})"
docker run -d --name "${CONTAINER}" \
  -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=postgres \
  -p "${HOST_PORT}:5432" "${PG_IMAGE}" >/dev/null
ready=0
for _ in $(seq 1 60); do
  if docker exec "${CONTAINER}" pg_isready -U postgres -d postgres >/dev/null 2>&1 \
     && (exec 3<>"/dev/tcp/127.0.0.1/${HOST_PORT}") 2>/dev/null; then
    ready=1
    break
  fi
  sleep 1
done
if [[ "${ready}" -ne 1 ]]; then
  echo "ERROR: Postgres did not become ready on 127.0.0.1:${HOST_PORT}"
  docker logs "${CONTAINER}" 2>&1 | tail -20
  exit 1
fi

echo "==> Applying migrations to the fresh database"
POSTGRES_ENABLED=true POSTGRES_URL="${HOST_URL}" go run ./cmd/migrate up

echo "==> Dumping normalized schema"
actual="$(mktemp)"
trap 'rm -f "${actual}"; cleanup' EXIT
docker exec "${CONTAINER}" pg_dump "${IN_CONTAINER_URL}" \
  --schema-only --no-owner --no-privileges \
  --schema=public --exclude-table=schema_migrations \
  | bash scripts/normalize-schema.sh > "${actual}"

if [[ "${UPDATE}" -eq 1 ]]; then
  cp "${actual}" "${GOLDEN}"
  echo "==> Updated ${GOLDEN}"
  exit 0
fi

echo "==> Diffing live schema against ${GOLDEN}"
if ! diff -u "${GOLDEN}" "${actual}"; then
  cat <<'EOF'

ERROR: live schema does not match db/schema.generated.sql.
If this change is intentional, regenerate the golden snapshot:
  ./scripts/check-schema.sh --update
and commit the result.
EOF
  exit 1
fi

echo "    ok — schema matches golden snapshot"
