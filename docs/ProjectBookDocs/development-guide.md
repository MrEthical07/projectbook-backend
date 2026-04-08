# Development Guide

## Run The Project

```bash
npm install
npm run dev
```

## Folder Structure (Working View)

- `src/routes/**`: route UI and loaders.
- `src/lib/remote/**`: read/write boundary functions.
- `src/lib/server/data/**`: datastore and seeded data.
- `src/lib/components/**`: shared UI.
- `src/lib/utils/**`: small shared helpers.

## Add A New Artifact

1. Create a data file in `src/lib/server/data` and register rows in `datastore`.
2. Create a remote file in `src/lib/remote/<artifact>.remote.ts`.
3. Add read and write functions (`query` for reads, `command` for writes) with Zod validation and permission checks.
4. Add route pages:
   - Index: `src/routes/project/[projectId]/<artifact>/+page.ts` and `+page.svelte`
   - Detail: `src/routes/project/[projectId]/<artifact>/[slug]/+page.ts` and `+page.svelte`
5. Hook artifact into sidebar/navigation data so users can reach index and detail routes.

## Add A New Field To An Existing Artifact

1. Add the field to shared type definitions in `src/app.d.ts`.
2. Add default/sample values in the matching `*.data.ts` file.
3. Update remote read payload shape.
4. Update remote write validation and normalization logic.
5. Bind field in detail page UI and include it in Save payload/signature tracking.

## Modify Remote Functions Safely

1. Keep input as `unknown` and parse with Zod.
2. Enforce permission checks before mutation.
3. Validate linked references and status transitions.
4. Return consistent mutation result shape.
5. Avoid introducing service-layer indirection.
