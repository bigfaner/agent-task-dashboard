---
date: "2026-05-06"
doc_dir: "Z:/project/ai/agent-task-center/docs/features/agent-task-dashboard/design/"
iteration: "2"
target_score: "90"
evaluator: Claude (automated, adversarial)
---

# Design Eval — Iteration 2

**Score: 95/100** (target: 90)

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
| 4. Testing Strategy           |  15    |  15    | PASS         |
|    Per-layer test plan        |  5/5   |        |              |
|    Coverage target numeric    |  5/5   |        |              |
|    Test tooling named         |  5/5   |        |              |
+-------------------------------+--------+--------+--------------+
| 5. Breakdown-Readiness *      |  20    |  20    | PASS         |
|    Components enumerable      |  7/7   |        |              |
|    Tasks derivable            |  7/7   |        |              |
|    PRD AC coverage            |  6/6   |        |              |
+-------------------------------+--------+--------+--------------+
| 6. Security Considerations    |   8    |  10    | WARN         |
|    Threat model present       |  4/5   |        |              |
|    Mitigations concrete       |  4/5   |        |              |
+-------------------------------+--------+--------+--------------+
| TOTAL                         |  95    |  100   |              |
+-------------------------------+--------+--------+--------------+
```

* Breakdown-Readiness < 12/20 blocks progression to `/breakdown-tasks`

---

## Deductions

| Location | Issue | Penalty |
|----------|-------|---------|
| tech-design.md: Interface 4 (lines 170-178) | API Handlers interface is still comment-only route listing, not typed Go function signatures like Interfaces 1-3 and 5. Carried over from iteration 1. | -1 pts (Interface signatures typed) |
| tech-design.md: Interface 4 (lines 170-178) | API Handlers lack request/response struct definitions; developer must cross-reference api-handbook.md to implement. | -1 pts (Directly implementable) |
| tech-design.md + api-handbook.md | Error code names inconsistent between internal design (`ERR_INVALID_SLUG`, `ERR_PROJECT_NOT_FOUND`) and API contract (`invalid_slug`, `not_found`). Developer must reconcile two different naming schemes. | -1 pts (HTTP status codes mapped) |
| api-handbook.md: Get Project response (line 74) | `ProjectData.Path` exposes full filesystem path (`"path": "Z:/project/ai/pm-work-tracker"`) in API response. Not identified in threat model. Flagged in iteration 1, still unaddressed. | -1 pts (Threat model present) |
| tech-design.md: Security Mitigations | "sanitize markdown content before rendering" still has no named library (e.g., bluemonday, goldmark). Flagged in iteration 1, still vague. | -1 pts (Mitigations concrete) |

---

## Attack Points

### Attack 1: Interface & Model Definitions — API Handlers interface remains prose, not code

**Where**: tech-design.md Interface 4 (lines 170-178): "// GET /api/projects — list all projects with summaries" (repeated for 7 routes as comments)

**Why it's weak**: Every other interface in the design uses typed Go code — function signatures with parameter types and return types. Interface 4 is the sole exception: it is a block of comments listing route paths. This was flagged in iteration 1 and remains unchanged. A developer cannot look at Interface 4 alone and know the handler function signatures, request binding logic, or response serialization. They must cross-reference api-handbook.md and mentally reconstruct the implementation. This is the most interface-heavy component in the system (7 endpoints) yet has the least formal interface definition.

**What must improve**: Replace the comment block with typed Go function signatures (e.g., `func handleListProjects(c *gin.Context)` with inline documentation of request parameters and response shape), or define request/response structs alongside the route registrations. Match the rigor of Interfaces 1-3 and 5.

### Attack 2: Security — Filesystem path leakage in API response persists from iteration 1

**Where**: api-handbook.md Get Project response (line 74): `"path": "Z:/project/ai/pm-work-tracker"` and tech-design.md ProjectData model (line 247): `Path string // Filesystem path`

**Why it's weak**: The `ProjectData.Path` field exposes absolute filesystem paths in the public API response. This was flagged as a threat gap in iteration 1's evaluation. The iteration 2 revision added no mitigation and no threat model entry for this. While the dashboard is localhost-only, the API is explicitly designed for agent consumption (PRD: "Agent client" data flow). An agent receiving this response learns the operator's full directory structure. At minimum, this should be documented as an accepted risk with rationale, or the path field should be excluded from API responses.

**What must improve**: Either (a) add `Path` to the threat model with explicit acceptance rationale ("localhost-only, agent is trusted, paths are non-sensitive"), or (b) remove `Path` from the API response JSON (keep it internal-only, omit from json tag or use a separate API response struct).

### Attack 3: Error Handling — Error code naming inconsistency between design and API contract

**Where**: tech-design.md Error Types table uses `ERR_INVALID_SLUG`, `ERR_PROJECT_NOT_FOUND`, `ERR_FEATURE_NOT_FOUND`, `ERR_FS_READ`, etc. api-handbook.md Error Codes table uses `invalid_slug`, `not_found`, `internal_error`.

**Why it's weak**: Two documents that are supposed to describe the same system use two completely different error code naming conventions. The tech design uses SCREAMING_SNAKE_CASE with `ERR_` prefix; the API handbook uses lowercase snake_case without prefix. The API handbook collapses `ERR_PROJECT_NOT_FOUND`, `ERR_FEATURE_NOT_FOUND`, and `ERR_TASK_NOT_FOUND` into a single `not_found` code, losing granularity. A developer implementing error handling must decide: which naming scheme is canonical? Does the handler translate from internal `ERR_PROJECT_NOT_FOUND` to external `not_found`? If so, where is this mapping defined? The propagation strategy in tech-design.md references internal names only; the API contract references external names only. There is no mapping layer specified between them.

**What must improve**: Pick one canonical naming scheme and use it consistently across both documents, or add an explicit mapping table showing how internal error codes translate to API-facing error codes. The api-handbook.md should show the error code values that actually appear in HTTP responses, and the tech design should show how those are derived from internal error types.

---

## Previous Issues Check

| Previous Attack | Addressed? | Evidence |
|----------------|------------|----------|
| Attack 1: Three PRD ACs unaddressed (sort order, dependency navigation, record section parsing) | YES | (1) `SortFeatures` function added at Interface 2 with explicit 3-step algorithm. (2) Interface 6 adds `navigateToTask(taskId, featureSlug)` JS function with cross-feature navigation. (3) `RecordContent` struct added with Summary, Files, Decisions, TestResults fields. PRD Coverage Map updated with entries for all three. |
| Attack 2: Mock tooling vague ("os.ReadFile mocks", JS "Manual") | YES | Scanner unit tests now specify `testing/fstest` + `fs.FS` interface injection with `fstest.MapFS`. Integration tests use `os.CreateTemp` temp dirs. JS swimlane tests use `Jest + jsdom`. All concrete tooling named. |
| Attack 3: Missing 400-level error types and input validation | YES | Error types table now includes `ERR_INVALID_SLUG` (400) and `ERR_INVALID_TASK_ID` (400). Propagation strategy includes validation middleware layer. |

---

## Verdict

- **Score**: 95/100
- **Target**: 90/100
- **Gap**: +5 points (target exceeded)
- **Breakdown-Readiness**: 20/20 — CAN proceed to /breakdown-tasks
- **Action**: Target reached. Design is ready for breakdown. Remaining weaknesses (API handler interface formality, error code naming consistency, path leakage threat model entry) are minor and can be resolved during implementation.
