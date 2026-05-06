---
id: "disc-1"
title: "Fix e2e: TC-013/014 filter dropdown + TC-022 overlay interception"
priority: "P0"
estimated_time: "30min"
dependencies: []
status: pending
breaking: true
---

# Fix e2e: TC-013/014 filter dropdown + TC-022 overlay interception

## Root Cause

3 e2e test timeouts: TC-013/TC-014 filter dropdown auto-closes after checkbox click (element not visible for uncheck). TC-022 detail overlay intercepts pointer events on task cards. Fix test scripts to handle dropdown reopen and close overlay before card iteration.

## Reference Files

- Source: tests/e2e/features/agent-task-dashboard/ui.spec.ts
- Test script: tests/e2e/features/agent-task-dashboard/ui.spec.ts
- Test results: tests/e2e/features/agent-task-dashboard/results/latest.md

## Verification

After fixing, verify the fix works:
1. `just test [scope]` — must pass
2. If UI/page related: `just test-e2e --feature <slug>` — must also pass

When this task is recorded as completed via `task record`, the source task T-test-3 is automatically restored to pending if all its dependencies are completed.
