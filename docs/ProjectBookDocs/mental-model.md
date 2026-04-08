# Mental Model

## Design Thinking Flow

Empathize -> Define -> Ideate -> Prototype -> Test

This flow is sequential by design.

- Empathize produces user context.
- Define turns that context into explicit problem statements.
- Ideate generates options against defined problems.
- Prototype turns selected ideas into executable work.
- Test collects evidence and outcomes.

Random usage breaks traceability. If teams skip phase order, downstream artifacts lose decision context.

## Artifact Chain

Story -> Problem -> Idea -> Task -> Feedback

How the chain works:

- Story captures persona and pain points.
- Problem references source context (stories and journeys).
- Idea links to a problem statement.
- Task links to an idea.
- Feedback links to tasks, ideas, or problem statements.

## Why Context Matters

Context carries intent and constraints. Without it, teams optimize for local output instead of solving the right problem.

## Why Linking Is Enforced

Linking rules keep phase transitions valid.

- Idea selection is constrained by linked, locked problems.
- Task linking validates the idea exists and is in an allowed state.
- Feedback linking validates artifact shape and type.

## Why Orphans Are Allowed But Highlighted

Orphans are valid during drafting and exploration. New artifacts often start unlinked.

Orphan state is still surfaced (`isOrphan`) so teams can close context gaps before moving work forward.
