# E2E Test Report: agent-task-dashboard

**Date**: 2026-05-07
**Duration**: 1.9m

## Summary

| Type  | Total | Pass | Fail | Skip |
|-------|-------|------|------|------|
| UI    | 27    | 24   | 3    | 0    |
| API   | 11    | 11   | 0    | 0    |
| CLI   | 0     | 0    | 0    | 0    |
| **All** | **38** | **35** | **3** | **0** |

**Result**: FAIL (3 tests timed out)

---

## Results by Test Case

### UI Tests

| TC ID | Title | Status | Duration |
|-------|-------|--------|----------|
| TC-001 | Project cards display with correct summary statistics | PASS | 619ms |
| TC-002 | Project card health status is correctly derived | PASS | 488ms |
| TC-003 | Clicking a project card navigates to swimlane view | PASS | 595ms |
| TC-004 | Landing page shows warning for invalid project paths | PASS | 467ms |
| TC-005 | Landing page shows empty state when no projects configured | PASS | 485ms |
| TC-006 | Landing page shows skeleton loading state | PASS | 474ms |
| TC-007 | Dark mode toggle switches theme | PASS | 572ms |
| TC-008 | Blocked task cards appear in red and sort to top | PASS | 580ms |
| TC-009 | Clicking a blocked task opens detail panel with dependencies | PASS | 590ms |
| TC-010 | Cross-feature dependency arrows render with dashed style | PASS | 606ms |
| TC-011 | Task cards are color-coded by status | PASS | 641ms |
| TC-012 | Swimlane shows empty state when no features found | PASS | 572ms |
| TC-013 | Filter by status shows only matching feature rows | FAIL | 30.1s |
| TC-014 | Filter by priority shows only matching feature rows | FAIL | 30.1s |
| TC-015 | Collapse and expand feature rows | PASS | 740ms |
| TC-016 | Dependency arrows to non-existent tasks are skipped | PASS | 595ms |
| TC-017 | Tasks with unrecognized phase grouped in Other column | PASS | 623ms |
| TC-018 | Detail panel slides in and displays task metadata | PASS | 771ms |
| TC-019 | Detail panel renders execution record as formatted markdown | PASS | 788ms |
| TC-020 | Detail panel shows no execution record message | PASS | 697ms |
| TC-021 | Detail panel closes via X button, overlay click, and Escape key | PASS | 1.9s |
| TC-022 | Dependency links in detail panel navigate to referenced tasks | FAIL | 30.1s |
| TC-023 | Activity sidebar displays recent status change events | PASS | 677ms |
| TC-024 | Clicking activity event scrolls to task and highlights | PASS | 1.6s |
| TC-025 | Activity sidebar collapse and expand with blocked count badge | PASS | 1.3s |
| TC-026 | Integration -- Task Detail Panel visible on Swimlane Page | PASS | 746ms |
| TC-027 | Integration -- Activity Sidebar visible on Swimlane Page | PASS | 590ms |

### API Tests

| TC ID | Title | Status | Duration |
|-------|-------|--------|----------|
| TC-028 | GET /api/projects returns all configured projects | PASS | 41ms |
| TC-029 | GET /api/projects/:id returns single project with features | PASS | 12ms |
| TC-030 | GET /api/projects/:id/features returns feature list | PASS | 7ms |
| TC-031 | GET /api/projects/:id/features/:slug/tasks returns task list | PASS | 66ms |
| TC-032 | API response time is under 200ms for single project | PASS | 8ms |
| TC-033 | GET /api/projects/:id/features/:slug returns feature with tasks | PASS | 5ms |
| TC-034 | GET /api/projects/:id/features/:slug/tasks/:taskId returns task details | PASS | 7ms |
| TC-035 | GET /api/projects/:id/features/:slug/dependencies returns dependency graph | PASS | 120ms |
| TC-036 | API returns 404 for non-existent project | PASS | 3ms |
| TC-037 | API returns 404 for non-existent feature | PASS | 4ms |
| TC-038 | API returns 404 for non-existent task | PASS | 4ms |

---

## Failed Tests Detail

### TC-013: Filter by status shows only matching feature rows

**Status**: TIMEOUT (30s)
**Root Cause**: After checking the "blocked" checkbox in the status filter dropdown, the dropdown auto-closes. The test then tries to uncheck the checkbox, but the element is no longer visible because the dropdown has closed.

**Error**: `locator.uncheck: Test timeout of 30000ms exceeded. Element is not visible.`

**Location**: `ui.spec.ts:264`

**Diagnostic**: The filter dropdown closes after checking a checkbox value. The test cleanup step (`blockedCheckbox.uncheck()`) fails because the dropdown is no longer open. The filter functionality itself appears to work (the assertion about visible rows passed before the uncheck attempt). The fix should re-open the dropdown before unchecking, or remove the uncheck cleanup step.

**Screenshot**: `results/features-agent-task-dashbo-6465c--only-matching-feature-rows/test-failed-1.png`

---

### TC-014: Filter by priority shows only matching feature rows

**Status**: TIMEOUT (30s)
**Root Cause**: Same pattern as TC-013. After checking the "P0" checkbox in the priority filter dropdown, the dropdown auto-closes. The test tries to uncheck, but the element is no longer visible.

**Error**: `locator.uncheck: Test timeout of 30000ms exceeded. Element is not visible.`

**Location**: `ui.spec.ts:295`

**Diagnostic**: The priority filter dropdown has the same auto-close behavior as the status filter. The test script cleanup step fails. The filter assertion passed before cleanup.

**Screenshot**: `results/features-agent-task-dashbo-14474--only-matching-feature-rows/test-failed-1.png`

---

### TC-022: Dependency links in detail panel navigate to referenced tasks

**Status**: TIMEOUT (30s)
**Root Cause**: When iterating through task cards to find one with dependencies, the detail panel overlay (`#detail-overlay`) intercepts pointer events. After opening a task's detail panel on the first iteration, clicking the next task card is blocked because the overlay covers it.

**Error**: `locator.click: Test timeout of 30000ms exceeded. <div id="detail-overlay" class="detail-overlay"> intercepts pointer events.`

**Location**: `ui.spec.ts:487`

**Diagnostic**: The test loops through task cards clicking each one to check for dependency chips. After the first click opens the detail panel, the overlay element prevents clicking subsequent task cards. The test should close the detail panel between iterations (click the overlay or press Escape), or should force-click through the overlay.

**Screenshot**: `results/features-agent-task-dashbo-9f372-avigate-to-referenced-tasks/test-failed-1.png`

---

## Screenshots

### Passed Tests

| TC ID | Screenshot |
|-------|-----------|
| TC-001 | `results/screenshots/TC-001.png` |
| TC-002 | `results/screenshots/TC-002.png` |
| TC-003 | `results/screenshots/TC-003.png` |
| TC-004 | `results/screenshots/TC-004.png` |
| TC-005 | `results/screenshots/TC-005.png` |
| TC-006 | `results/screenshots/TC-006.png` |
| TC-007 | `results/screenshots/TC-007.png` |
| TC-008 | `results/screenshots/TC-008.png` |
| TC-009 | `results/screenshots/TC-009.png` |
| TC-010 | `results/screenshots/TC-010.png` |
| TC-011 | `results/screenshots/TC-011.png` |
| TC-012 | `results/screenshots/TC-012.png` |
| TC-015 | `results/screenshots/TC-015.png` |
| TC-016 | `results/screenshots/TC-016.png` |
| TC-017 | `results/screenshots/TC-017.png` |
| TC-018 | `results/screenshots/TC-018.png` |
| TC-019 | `results/screenshots/TC-019.png` |
| TC-020 | `results/screenshots/TC-020.png` |
| TC-021 | `results/screenshots/TC-021.png` |
| TC-023 | `results/screenshots/TC-023.png` |
| TC-024 | `results/screenshots/TC-024.png` |
| TC-025 | `results/screenshots/TC-025.png` |
| TC-026 | `results/screenshots/TC-026.png` |
| TC-027 | `results/screenshots/TC-027.png` |

### Failed Tests (Auto-captured by Playwright)

| TC ID | Screenshot |
|-------|-----------|
| TC-013 | `results/features-agent-task-dashbo-6465c--only-matching-feature-rows/test-failed-1.png` |
| TC-014 | `results/features-agent-task-dashbo-14474--only-matching-feature-rows/test-failed-1.png` |
| TC-022 | `results/features-agent-task-dashbo-9f372-avigate-to-referenced-tasks/test-failed-1.png` |
