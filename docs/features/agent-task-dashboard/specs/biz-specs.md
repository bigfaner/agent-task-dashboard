---
feature: "agent-task-dashboard"
generated: "2026-05-07"
status: draft
---

# Business Rules: Agent Task Dashboard

## Task Lifecycle

### BIZ-001: Task Status Values

**Rule**: Task status must be one of: `pending`, `in_progress`, `completed`, `blocked`, `skipped`.
**Context**: These statuses drive visual representation (color coding) on the swimlane and activity sidebar, and determine feature sort order.
**Scope**: [LOCAL]
**Source**: PRD 5.2 (Task Card Fields), PRD 5.4 (Event Types)

Statuses map to colors: pending=grey, in_progress=blue, completed=green, blocked=red, skipped=yellow-strikethrough.

### BIZ-002: Activity Event Types

**Rule**: Activity events are limited to four types: `claimed` (status=in_progress), `completed` (status=completed), `blocked` (status=blocked), `skipped` (status=skipped).
**Context**: The activity sidebar derives events by observing current task statuses in index.json. There is no event log — events are inferred from state.
**Scope**: [LOCAL]
**Source**: PRD 5.4 (Event Types)

## Project Health

### BIZ-003: Project Health Status Derivation

**Rule**: Project health status is derived from task activity:
- **Active** (green): At least one task is in_progress or recently completed.
- **All Complete** (blue): All tasks across all features are completed.
- **Stale** (grey): No task updates in 7+ days.
**Context**: Health status gives operators a quick visual indicator of project state without drilling into individual features.
**Scope**: [LOCAL]
**Source**: PRD 5.1 (Status Description)

### BIZ-004: Completion Percentage Calculation

**Rule**: `completionPct = (completedTasks / totalTasks) * 100`. If `totalTasks = 0`, `completionPct = 0`.
**Context**: Used for project cards and feature rows to show progress at a glance.
**Scope**: [LOCAL]
**Source**: PRD 5.1 (List Fields)

## Feature Ordering

### BIZ-005: Feature Sort Order

**Rule**: Features are sorted by:
1. Features with any blocked tasks come first.
2. Within each group, sort by completion percentage ascending (most incomplete first).
3. Ties broken by slug alphabetically.
**Context**: Ensures blockers and incomplete features are visually prominent on the swimlane, matching the operator's diagnostic workflow.
**Scope**: [LOCAL]
**Source**: PRD 5.2 (Sort Order), Tech Design (Scanner.SortFeatures)

## Phase Assignment

### BIZ-006: Phase Derivation from Task ID

**Rule**: Task phase is derived from the leading number in the task ID:
- Tasks `1.x` -> Phase 1
- Tasks `2.x` -> Phase 2
- Tasks `3.x+` -> Phase 3+
- Tasks `T-test-*` -> Testing phase
**Context**: Phase columns on the swimlane are derived, not configured. This matches the forge task ID naming convention.
**Scope**: [LOCAL]
**Source**: PRD 5.2 (Swimlane Layout)

## Configuration Validation

### BIZ-007: Config Validation Rules

**Rule**: On project config validation:
- Path must exist: skip project, show warning on landing page.
- Path must contain `docs/features/` directory: skip project, show warning.
- Feature directories must have `tasks/index.json`: skip feature, show 0 tasks.
- `index.json` must be valid JSON matching schema: skip feature, show parse error.
**Context**: The dashboard must degrade gracefully when encountering misconfigured or partially-valid projects rather than failing entirely.
**Scope**: [LOCAL]
**Source**: PRD 5.6 (Validation Rules)

## Dependency Handling

### BIZ-008: Wildcard Dependency Expansion

**Rule**: Dependencies specified as wildcards (e.g., `"1.x"`) are expanded to individual dependency edges for all tasks matching the prefix.
**Context**: Forge tasks commonly declare group dependencies. The dashboard must expand these for accurate DAG rendering.
**Scope**: [LOCAL]
**Source**: PRD 5.2 (Dependency Arrows), Tech Design (Key Test Scenarios #5)
