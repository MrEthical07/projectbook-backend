# Test Documentation

This folder contains runtime verification and integration test planning artifacts.

## Files
- integration-test-plan.md: Scenario packs, execution phases, and completion gates.
- route-coverage-matrix.md: Deduplicated route inventory (85 unique routes) mapped to scenario packs.
- smoke-validation-2026-04-13.md: Migration/startup/smoke validation report.

## Intended Workflow
1. Use route-coverage-matrix.md to generate one integration test case set per route.
2. Implement scenario pack assertions from integration-test-plan.md.
3. Mark execution progress in your preferred tracker (test IDs per route).
4. Keep smoke report updated whenever environment/bootstrap behavior changes.
