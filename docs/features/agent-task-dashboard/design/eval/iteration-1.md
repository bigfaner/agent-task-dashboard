---
date: "2026-05-06"
doc_dir: "Z:/project/ai/agent-task-center/docs/features/agent-task-dashboard/design/"
iteration: "1"
target_score: "90"
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 1

**Score: 83/100** (target: 90)

```
+---------------------------------------------------------------+
|                     DESIGN QUALITY SCORECARD                   |
+-------------------------------+--------+--------+--------------+
| Dimension                     | Score  | Max    | Status       |
+-------------------------------+--------+--------+--------------+
| 1. Architecture Clarity       |  20    |  20    | PASS         |
|    Layer placement explicit   |  7/7   |        |              |
|    Component diagram present  |  7/7   |        |              |
|    Dependencies listed        |  6/6   |        |              |
+-------------------------------+--------+--------+--------------+
| 2. Interface & Model Defs     |  18    |  20    | WARN         |
|    Interface signatures typed |  6/7   |        |              |
|    Models concrete            |  7/7   |        |              |
|    Directly implementable     |  5/6   |        |              |
+-------------------------------+--------+--------+--------------+
| 3. Error Handling             |  14    |  15    | WARN         |
|    Error types defined        |  5/5   |        |              |
|    Propagation strategy clear |  5/5   |        |              |
|    HTTP status codes mapped   |  4/5   |        |              |
+-------------------------------+--------+--------+--------------+
| 4. Testing Strategy           |  13    |  15    | WARN         |
|    Per-layer test plan        |  5/5   |        |              |
|    Coverage target numeric    |  5/5   |        |              |
|    Test tooling named         |  3/5   |        |              |
+-------------------------------+--------+--------+--------------+
| 5. Breakdown-Readiness *      |  10    |  20    | FAIL         |
|    Components enumerable      |  7/7   |        |              |
|    Tasks derivable            |  6/7   |        |              |
|    PRD AC coverage            |  6/6   |        |              |
+-------------------------------+--------+--------+--------------+
| 6. Security Considerations    |   8    |  10    | PASS         |
|    Threat model present       |  4/5   |        |              |
|    Mitigations concrete       |  4/5   |        |              |
+-------------------------------+--------+--------+--------------+
| TOTAL                         |  83    |  100   |              |
+-------------------------------+--------+--------+--------------+
```

* Breakdown-Readiness < 12/20 blocks progression to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| tech-design.md: Interface 4 (lines 157-169) | API Handlers interface is prose route listing in comments, not typed Go function signatures like other interfaces | -1 pts (Interface signatures typed) |
| tech-design.md: Interface 3 & 4 (lines 144-169) | Page and API handler interfaces lack request/response struct definitions; developer must guess request parsing and response shape | -1 pts (Directly implementable) |
| api-handbook.md: Get Feature (lines 139) | `"tasks": { ... }` uses ellipsis instead of full type definition, forcing developer to guess task object shape in context | -1 pts (Directly implementable) |
| tech-design.md: Error Handling (lines 289-298) | No 400 Bad Request status code mapped for input validation failures (malformed slug, invalid task ID format) despite security section mentioning alphanumeric+hyphens constraint | -1 pts (HTTP status codes mapped) |
| tech-design.md: Testing Strategy (lines 332-333) | "os.ReadFile mocks" is not a real tool or library; no concrete mock framework named (testify, gomock, etc.) | -2 pts (Test tooling named) |
| tech-design.md: Testing Strategy (lines 339) | JS testing listed as "Manual" with no tooling at all — no Playwright, no Jest, no Cypress | -2 pts (Test tooling named) |
| tech-design.md + PRD | PRD User Story 2 AC: "features with blocked tasks sort to the top of the swimlane" — completely absent from design. No sort logic defined anywhere. | -3 pts (PRD AC gap) |
| tech-design.md + PRD | PRD User Story 2 AC: "each dependency is clickable to navigate to the upstream task" — no design for click-to-navigate behavior in the detail panel | -3 pts (PRD AC gap) |
| tech-design.md + PRD | PRD User Story 4 AC: "panel shows: summary, files created/modified, key decisions, test results" — ParseRecordFile returns raw string with no structured parsing into sections | -3 pts (PRD AC gap) |
| tech-design.md: Security (lines 367) | "sanitize markdown content before rendering" — no specific sanitizer library or approach named | -1 pts (Mitigations concrete) |
| tech-design.md: Security | ProjectData.Path exposed in API response (api-handbook line 74) leaks filesystem paths; not identified as threat | -1 pts (Threat model present) |

---

## Attack Points

### Attack 1: Breakdown-Readiness — Three PRD Acceptance Criteria completely unaddressed

**Where**: PRD User Stories 2 and 4 vs. tech-design.md PRD Coverage Map (lines 373-394)

**Why it's weak**: Three acceptance criteria from the PRD are not addressed in the design:
1. PRD 5.2 states "Features sorted by: blocked tasks first, then by completion % ascending (most incomplete first)" and User Story 2 requires "features with blocked tasks sort to the top of the swimlane." The PRD Coverage Map and tech design contain zero mention of feature sort order. A developer cannot implement correct swimlane ordering without this.
2. User Story 2 requires "each dependency is clickable to navigate to the upstream task." The detail panel shows a dependency list but there is no interface or JS function for navigating from a dependency entry to the upstream task card on the swimlane.
3. User Story 4 requires the panel to show structured sections: "summary, files created/modified, key decisions, test results." The design's ParseRecordFile returns a raw `string` — no parsing logic, no section extraction, no structured model for the record content.

These are not edge cases; they are core user-facing behaviors defined in the PRD.

**What must improve**: Add explicit design for: (a) feature sort algorithm in the Scanner or handler, (b) dependency click-to-navigate interaction in the detail panel JS, (c) a structured model for execution record sections with a parsing strategy.

### Attack 2: Testing Strategy — Mocking approach is hand-waving, not a real plan

**Where**: tech-design.md Testing Strategy (lines 332-333): "testing + os.ReadFile mocks"

**Why it's weak**: "os.ReadFile mocks" is not a library, framework, or concrete approach. In Go, mocking filesystem reads requires either: (a) an abstraction layer (fs.FS interface) that can be swapped with a memfs, (b) generating mocks via gomock/mockgen against a defined interface, (c) using testify/mock, or (d) creating temp directories with test fixtures. The design names none of these. The Scanner component is the most critical unit to test (parsing index.json, deriving health status, expanding wildcard dependencies) yet the mock strategy is a placeholder phrase. JS testing is "Manual" — meaning the swimlane DAG renderer, the most complex UI component, has zero automated test coverage planned.

**What must improve**: Name a concrete mocking approach (e.g., "abstract filesystem behind io/fs.FS interface; use testing/fstest in unit tests, real temp dirs in integration tests"). Name a JS test framework for the swimlane renderer or justify why manual-only is acceptable for the highest-risk frontend component.

### Attack 3: Error Handling — Missing 400-level status codes and input validation error types

**Where**: tech-design.md Error Types table (lines 290-296) and api-handbook.md Error Codes table (lines 269-273)

**Why it's weak**: The error design covers only two HTTP status codes: 404 and 500. The security section explicitly states that `:slug` must contain "only alphanumeric + hyphens" and `:id` must match a configured project name. But when these validations fail, there is no defined error type or status code. Should a malformed slug like `/api/projects/../../etc/passwd/features` return 404? 400? 500? The error types table has no `ERR_INVALID_INPUT`, `ERR_VALIDATION`, or equivalent. The propagation strategy maps only ErrNotFound and ErrFSRead/ErrParseIndex. Input validation errors are unclassified.

**What must improve**: Add input validation error types (e.g., `ERR_INVALID_SLUG`, `ERR_INVALID_TASK_ID`) mapped to HTTP 400 Bad Request. Update the propagation strategy to show where validation occurs (middleware vs. handler) and how validation errors flow to the response.

---

## Previous Issues Check

N/A — Iteration 1.

---

## Verdict

- **Score**: 83/100
- **Target**: 90/100
- **Gap**: 7 points
- **Breakdown-Readiness**: 10/20 — CANNOT proceed to /breakdown-tasks (below 12/20 gate)
- **Action**: Continue to iteration 2. Must address the 3 unaddressed PRD ACs (sort order, dependency navigation, record section parsing), add concrete mock tooling, and add input validation error types to reach target and unblock breakdown.
