# FAQ

## Why no backend yet?

Current implementation uses in-memory datastore modules to validate product flow and architecture quickly. Remote contracts are already structured as a future API boundary.

## Why no auto-save?

Explicit Save keeps writes intentional, avoids noisy updates during drafting, and makes permission or validation failures visible at commit time.

## Why no service layer?

Remote files already contain boundary logic (validation, permissions, mutation). Adding a service layer would duplicate responsibility and increase complexity.

## Why full object replacement instead of patching?

Snapshot-style saves keep behavior deterministic and reduce drift from partial updates. Remote commands derive normalized persisted fields from a complete editor state.

## Why enforce Design Thinking structure?

The system is built for context continuity across phases. Phase-linked artifacts make decisions traceable from user signal to tested outcome.
