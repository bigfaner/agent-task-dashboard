---
id: "2.gate"
title: "Phase 2 Exit Gate"
priority: "P0"
estimated_time: "1h"
dependencies: ["2.summary"]
status: pending
breaking: true
---

# 2.gate: Phase 2 Exit Gate

## Description

Exit verification gate for Phase 2 (Backend). Confirms that all API endpoints respond correctly, page handlers render HTML, and the binary builds and serves.

## Verification Checklist

1. [ ] All 7 API endpoints return correct JSON structure matching api-handbook.md
2. [ ] API returns 404 for non-existent project/feature/task IDs
3. [ ] API returns 400 for invalid slug/taskId formats
4. [ ] Landing page handler returns HTML with project data
5. [ ] Project page handler returns HTML with feature data
6. [ ] Task markdown parser extracts acceptance criteria and record sections
7. [ ] Binary builds as single executable: `go build -o task-dashboard ./cmd/task-dashboard/`
8. [ ] Binary starts and responds to `GET /` within 3 seconds
9. [ ] `go test ./...` passes
10. [ ] No deviations from api-handbook.md response formats

## Reference Files

- `design/tech-design.md` — Interfaces 3-5
- `design/api-handbook.md` — All endpoint specifications
- Phase 2 task records: `records/2.*.md`
- Phase 2 summary: `records/2-summary.md`

## Acceptance Criteria

- [ ] All applicable verification checklist items pass
- [ ] Any deviations documented as decisions in the record
- [ ] Record created via `/record-task` with test evidence

## Implementation Notes

This is a verification-only task. No new feature code should be written.
