---
id: "3.gate"
title: "Phase 3 Exit Gate"
priority: "P0"
estimated_time: "1h"
dependencies: ["3.summary"]
status: pending
breaking: true
---

# 3.gate: Phase 3 Exit Gate

## Description

Exit verification gate for Phase 3 (Frontend). Confirms that the swimlane DAG renders correctly, all UI interactions work, and the full dashboard is functional.

## Verification Checklist

1. [ ] Landing page renders project cards with correct stats
2. [ ] Clicking project card navigates to swimlane view
3. [ ] Swimlane renders feature rows with phase columns
4. [ ] Task cards show correct status colors (pending=grey, in_progress=blue, completed=green, blocked=red, skipped=yellow)
5. [ ] Dependency arrows render between tasks (solid within feature, dashed cross-feature)
6. [ ] Features with blocked tasks sort to top
7. [ ] Filter controls work (by status, by priority)
8. [ ] Collapsible rows work
9. [ ] Clicking task card opens detail panel with full metadata
10. [ ] Detail panel shows acceptance criteria and execution record sections
11. [ ] Shows "No execution record" when record absent
12. [ ] Dependency links in panel navigate to referenced task
13. [ ] Activity sidebar shows 50 recent events with timestamps
14. [ ] Activity sidebar collapse/expand works
15. [ ] Clicking event scrolls to task
16. [ ] Renders 24 features (330 tasks) without visual breakage at 1920x1080
17. [ ] All existing Go tests pass (`go test ./...`)
18. [ ] Binary builds and serves correctly

## Reference Files

- `prd/prd-ui-functions.md` — All 4 UI function specifications
- `ui/ui-design.md` — All 4 component designs
- Phase 3 task records: `records/3.*.md`
- Phase 3 summary: `records/3-summary.md`

## Acceptance Criteria

- [ ] All applicable verification checklist items pass
- [ ] Any deviations documented as decisions in the record
- [ ] Record created via `/record-task` with test evidence

## Implementation Notes

This is a verification-only task. No new feature code should be written.
Manual testing in browser is expected for UI verification items.
