# Sidebar Module Workflows (Deprecated)

This workflow document is retained for historical reference.

Current active behavior:

1. Project sidebar read data is served by `GET /api/v1/projects/{projectId}/navigation` in the project module.
2. Sidebar create actions in the web app dispatch to artifact module create endpoints (`/stories`, `/journeys`, `/problems`, `/ideas`, `/tasks`, `/feedback`, `/pages`).
3. Sidebar rename and delete actions are not part of the current sidebar UI flow.

For active route behavior, refer to:

- `docs/workflows/project.md`
- `docs/test/route-coverage-matrix.md`
