---
id: "T-test-4"
title: "Graduate Test Scripts"
priority: "P1"
estimated_time: "30min"
dependencies: ["T-test-3"]
status: pending
---

# Graduate Test Scripts

## Description

Call `/graduate-tests` skill to migrate feature test scripts from `tests/e2e/features/agent-task-dashboard/` to the project-wide regression suite.

## Reference Files

- `tests/e2e/features/agent-task-dashboard/results/latest.md` — Must show status = PASS
- `tests/e2e/features/agent-task-dashboard/` — Source scripts

## Acceptance Criteria

- [ ] `tests/e2e/features/agent-task-dashboard/results/latest.md` shows status = PASS
- [ ] `tests/e2e/.graduated/agent-task-dashboard` marker exists
- [ ] Spec files present in `tests/e2e/<module>/`

## User Stories

No direct user story mapping. This is a standard test graduation task.

## Implementation Notes

Verify e2e passed before graduating. If still failing, mark blocked.
