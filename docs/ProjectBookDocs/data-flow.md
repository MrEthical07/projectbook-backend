# Data Flow

## Read Flow

page -> `page.ts` -> remote -> data

1. `+page.svelte` renders props.
2. `+page.ts` calls a remote query.
3. Remote query validates scope and loads from datastore.
4. Page receives shaped data.

## Write Flow

UI edit -> Save button -> remote update command -> datastore update

1. User edits local page state.
2. User clicks Save.
3. UI sends a full editor state payload to remote command.
4. Remote validates input and permissions.
5. Remote updates normalized row fields and stored detail state.

## Full Replacement Strategy

Write commands are snapshot-based, not patch endpoint-based.

- UI sends full editable state for the screen.
- Remote derives persisted fields from that state.
- Detail state caches are replaced with the new normalized snapshot.

This keeps behavior deterministic and reduces partial-update drift.

## Why No Auto-Save

- Prevents noisy writes while users are still drafting.
- Makes permission failures explicit at commit time.
- Keeps status transitions and linked artifact rules intentional.

## Why Validation Lives In Remote

- UI state can be stale or manipulated.
- Remote is the trust boundary before mutation.
- Centralized validation keeps behavior consistent across pages.
