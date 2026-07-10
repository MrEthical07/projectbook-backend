#!/usr/bin/env bash
# Normalize a raw `pg_dump --schema-only` stream (read from stdin) into a
# deterministic form suitable for committing as a golden file and diffing.
#
#   <raw pg_dump output> | ./scripts/normalize-schema.sh > db/schema.generated.sql
#
# Removes:
#   - comments (-- ...), which include the non-deterministic pg_dump version banner
#   - SET / SELECT pg_catalog.set_config preamble
#   - psql meta-commands (lines starting with '\', e.g. the randomized
#     \restrict / \unrestrict tokens emitted by newer pg_dump)
# and squeezes repeated blank lines.
set -euo pipefail

grep -v -E -e '^--' -e '^SET ' -e '^SELECT pg_catalog' -e '^\\' | cat -s
