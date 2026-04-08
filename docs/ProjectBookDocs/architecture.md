# Architecture

## Runtime Path

UI (`page.svelte`)
-> `page.ts`
-> remote functions
-> data files (`datastore` + `*.data.ts`)

## Why This Architecture

- No service layer: the remote layer already owns input parsing, validation, permission checks, and state mutation.
- No command-pattern abstraction: command/query functions from `$app/server` are the execution boundary.
- Remote is the boundary: UI does not mutate datastore directly.
- Simplicity is intentional: fewer layers means lower cognitive overhead and faster debugging.

## Layer Ownership

- `src/routes/**`
  - Page UI and route-level data loading.
  - Owns rendering, local state, and Save interactions.

- `src/lib/remote/**`
  - Boundary for read/write operations.
  - Owns Zod validation, permission gating, normalization, and mutation results.

- `src/lib/server/data/**`
  - In-memory datastore and seeded domain data.
  - Owns baseline domain shape and sample state.

- `src/lib/components/**`
  - UI composition primitives and domain components.

- `src/lib/utils/**`
  - Shared helpers such as permission checks for UI decisions.

## Folder Structure Snapshot

- `src/routes/project/[projectId]/**`: project-scoped pages and artifact detail screens.
- `src/lib/remote/*.remote.ts`: domain and user-home boundary functions.
- `src/lib/server/data/*.data.ts`: domain seed data.
- `src/lib/server/data/datastore.ts`: in-memory state container.
