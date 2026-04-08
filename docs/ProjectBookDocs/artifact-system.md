# Artifact System

## Core Principle

Each artifact type has:

- An index page (list and creation entry point).
- A detail page (editable state with explicit Save).
- A full-save editing model through remote commands.

## Route Pattern

- Stories: `/project/[projectId]/stories` and `/project/[projectId]/stories/[slug]`
- Journeys: `/project/[projectId]/journeys` and `/project/[projectId]/journeys/[slug]`
- Problems: `/project/[projectId]/problem-statement` and `/project/[projectId]/problem-statement/[slug]`
- Ideas: `/project/[projectId]/ideas` and `/project/[projectId]/ideas/[slug]`
- Tasks: `/project/[projectId]/tasks` and `/project/[projectId]/tasks/[slug]`
- Feedback: `/project/[projectId]/feedback` and `/project/[projectId]/feedback/[slug]`
- Pages: `/project/[projectId]/pages` and `/project/[projectId]/pages/[slug]`

## Linking Model

Design Thinking chain:

Story -> Problem -> Idea -> Task -> Feedback

Supporting artifacts (journeys/pages/resources/calendar) provide additional context and execution visibility.

## References

- Problems reference linked source titles from stories and journeys.
- Ideas reference locked problem statements.
- Tasks reference ideas (and derive linked problem/persona context).
- Feedback references tasks, ideas, and problem statements.

## Orphan State

Every core artifact supports `isOrphan`.

- `true`: artifact exists without required downstream or upstream linkage.
- `false`: artifact has expected links for its phase position.

Orphans are valid during draft work, but should be reviewed before progressing phase outcomes.
