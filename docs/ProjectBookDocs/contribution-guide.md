# Contribution Guide

## Ground Rules

- No service layer.
- No command-pattern abstraction.
- No direct datastore imports inside route pages.
- Use remote files as the only read/write boundary for route code.

## PR Expectations

- Keep changes scoped.
- Explain domain impact in PR description.
- Include before/after behavior for permission or status changes.
- Update docs when architecture or flow behavior changes.

## Naming Conventions

- Remote files: `<domain>.remote.ts`
- Data files: `<domain>.data.ts`
- Route detail params: `[slug]` unless domain-specific id is already established.
- Remote function names:
  - Reads: `get<Domain...>`
  - Writes: `create<Domain>`, `update<Domain>`, `update<Domain>Status>`, `delete<Domain>`

## Remote Function Structure

1. Define Zod input schema.
2. Parse input.
3. Check permission.
4. Validate domain rules.
5. Mutate datastore.
6. Return typed success/error object.

## Anti-Patterns To Reject In Review

- Adding a service layer between routes and remotes.
- Writing mutation logic directly in page components.
- Introducing role-string checks in UI where action permissions already exist.
- Adding partial patch endpoints that bypass current snapshot-based save model.
