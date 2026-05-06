---
feature: "agent-task-dashboard"
sources:
  - docs/features/agent-task-dashboard/prd/prd-user-stories.md
  - docs/features/agent-task-dashboard/prd/prd-spec.md
  - docs/features/agent-task-dashboard/prd/prd-ui-functions.md
generated: "2026-05-07"
---

# Test Cases: agent-task-dashboard

## Summary

| Type | Count |
|------|-------|
| UI   | 24   |
| **Integration** | **2** |
| API  | 11  |
| CLI  | 0  |
| **Total** | **37** |

> **Note**: Integration test count is a subset of UI count. Integration tests verify that components are correctly wired into their parent pages, using the same Playwright framework as UI tests.

---

## UI Test Cases

### Landing Page — Project Card Grid (UF-1)

## TC-001: Project cards display with correct summary statistics
- **Source**: Story 1 / AC-1, AC-2
- **Type**: UI
- **Target**: ui/landing
- **Test ID**: ui/landing/project-cards-display-with-correct-summary-statistics
- **Pre-conditions**: Dashboard is running and 2+ projects are configured with valid paths
- **Route**: `/`
- **Steps**:
  1. Open the dashboard URL in a browser
  2. Wait for project cards to render (skeleton → populated transition)
  3. Verify one card per configured project is displayed
  4. For each card, verify: project name, feature count, task completion %, last update time are visible
- **Expected**: Each project card shows name, feature count, completed/total tasks, completion percentage, and last updated timestamp. Cards are sorted alphabetically by project name.
- **Priority**: P0

## TC-002: Project card health status is correctly derived
- **Source**: Story 1 / AC-2
- **Type**: UI
- **Target**: ui/landing
- **Test ID**: ui/landing/project-card-health-status-correctly-derived
- **Pre-conditions**: Projects exist with different health states (active with in_progress tasks, all completed, stale with no recent updates)
- **Route**: `/`
- **Steps**:
  1. Open the dashboard URL
  2. Identify a project with at least one in_progress task
  3. Verify the active project shows a green health indicator
  4. Identify a project where all tasks are completed
  5. Verify it shows a blue "All Complete" health indicator
  6. Identify a project with no task updates in 7+ days
  7. Verify it shows a grey "Stale" health indicator
- **Expected**: Health status badge on each card matches the derived state: active (green), complete (blue), or stale (grey).
- **Priority**: P0

## TC-003: Clicking a project card navigates to swimlane view
- **Source**: Story 1 / AC-3
- **Type**: UI
- **Target**: ui/landing
- **Test ID**: ui/landing/clicking-project-card-navigates-to-swimlane
- **Pre-conditions**: At least one project is configured
- **Route**: `/`
- **Steps**:
  1. Open the dashboard URL
  2. Click on a project card
  3. Verify the browser navigates to `/projects/:id` for that project
- **Expected**: Navigation to the correct project swimlane URL occurs on card click.
- **Priority**: P0

## TC-004: Landing page shows warning for invalid project paths
- **Source**: Spec 5.6 Validation Rules + UI Function 1 States
- **Type**: UI
- **Target**: ui/landing
- **Test ID**: ui/landing/warning-for-invalid-project-paths
- **Pre-conditions**: Config file includes a project entry with a non-existent filesystem path
- **Route**: `/`
- **Steps**:
  1. Open the dashboard URL
  2. Look for a warning banner for the project with invalid path
  3. Verify the warning banner displays the project name and invalid path
- **Expected**: A warning banner is displayed for the project with the invalid path, showing the project name and path. The warning is non-navigable. Valid projects still render their cards.
- **Priority**: P1

## TC-005: Landing page shows empty state when no projects configured
- **Source**: UI Function 1 States (Empty)
- **Type**: UI
- **Target**: ui/landing
- **Test ID**: ui/landing/empty-state-when-no-projects-configured
- **Pre-conditions**: Config file has zero project entries
- **Route**: `/`
- **Steps**:
  1. Open the dashboard URL
  2. Verify no project cards are rendered
  3. Verify an empty state message "No projects configured" is displayed
- **Expected**: The page displays a centered empty state with "No projects configured" message and guidance text.
- **Priority**: P1

## TC-006: Landing page shows skeleton loading state
- **Source**: UI Function 1 States (Loading) + UI Design Component 1 States
- **Type**: UI
- **Target**: ui/landing
- **Test ID**: ui/landing/skeleton-loading-state
- **Pre-conditions**: Dashboard is running with at least one project configured
- **Route**: `/`
- **Steps**:
  1. Open the dashboard URL
  2. Immediately observe the page before data loads
  3. Verify skeleton cards (pulse animation) are displayed
- **Expected**: Skeleton placeholder cards are shown during initial page load, transitioning to populated cards once data loads.
- **Priority**: P2

## TC-007: Dark mode toggle switches theme
- **Source**: UI Design Component 1 Interactions (Dark mode toggle)
- **Type**: UI
- **Target**: ui/landing
- **Test ID**: ui/landing/dark-mode-toggle-switches-theme
- **Pre-conditions**: Dashboard is running with at least one project
- **Route**: `/`
- **Steps**:
  1. Open the dashboard URL
  2. Click the dark mode toggle button in the header
  3. Verify the page theme switches to dark mode (dark backgrounds, light text)
  4. Click the toggle again
  5. Verify the page returns to light mode
- **Expected**: Theme toggles between light and dark mode, with CSS variable swap applied across all elements.
- **Priority**: P2

### Swimlane View — Feature Task Board (UF-2)

## TC-008: Blocked task cards appear in red and sort to top
- **Source**: Story 2 / AC-1, AC-2
- **Type**: UI
- **Target**: ui/swimlane
- **Test ID**: ui/swimlane/blocked-tasks-appear-red-and-sort-to-top
- **Pre-conditions**: A project has at least one blocked task
- **Route**: `/projects/:id`
- **Steps**:
  1. Navigate to a project swimlane page that has blocked tasks
  2. Verify blocked task cards display with red status color
  3. Verify features with blocked tasks appear at the top of the swimlane (sorted by blocked count descending)
- **Expected**: Blocked task cards are red. Feature rows containing blocked tasks are sorted to the top of the swimlane view.
- **Priority**: P0

## TC-009: Clicking a blocked task opens detail panel with dependencies
- **Source**: Story 2 / AC-3, AC-4
- **Type**: UI
- **Target**: ui/swimlane
- **Test ID**: ui/swimlane/clicking-blocked-task-opens-detail-panel
- **Pre-conditions**: A project has at least one blocked task with dependencies listed
- **Route**: `/projects/:id`
- **Steps**:
  1. Navigate to a project swimlane page with blocked tasks
  2. Click on a blocked task card
  3. Verify the detail panel (UF-3) slides in from the right
  4. Verify the dependency list is displayed in the panel
  5. Click on a dependency link
  6. Verify the swimlane scrolls to the upstream task and highlights it
- **Expected**: Detail panel opens showing task metadata including dependency list. Each dependency is a clickable link that navigates to the referenced task on the swimlane with a highlight ring animation.
- **Priority**: P0

## TC-010: Cross-feature dependency arrows render with dashed style
- **Source**: Story 3 / AC-1, AC-2
- **Type**: UI
- **Target**: ui/swimlane
- **Test ID**: ui/swimlane/cross-feature-dependency-arrows-render-dashed
- **Pre-conditions**: Task A in feature X depends on task B in feature Y; both features are visible
- **Route**: `/projects/:id`
- **Steps**:
  1. Navigate to a project swimlane page with cross-feature dependencies
  2. Verify a dashed dependency arrow connects task A to task B across the feature rows
  3. Verify the arrow is directional (pointing from dependent to dependency)
- **Expected**: Cross-feature dependency arrows appear as dashed directional lines between task cards in different feature rows.
- **Priority**: P0

## TC-011: Task cards are color-coded by status
- **Source**: Spec 5.2 Status Color Coding
- **Type**: UI
- **Target**: ui/swimlane
- **Test ID**: ui/swimlane/task-cards-color-coded-by-status
- **Pre-conditions**: A project has tasks in multiple statuses (pending, in_progress, completed, blocked, skipped)
- **Route**: `/projects/:id`
- **Steps**:
  1. Navigate to a project swimlane page with tasks in various statuses
  2. Verify pending tasks display with grey status color
  3. Verify in_progress tasks display with blue status color
  4. Verify completed tasks display with green status color
  5. Verify blocked tasks display with red status color (with pulse animation)
  6. Verify skipped tasks display with yellow strikethrough status
- **Expected**: Each task card's status dot and text color matches the defined color palette for its status.
- **Priority**: P0

## TC-012: Swimlane shows empty state when no features found
- **Source**: UI Function 2 States (Empty)
- **Type**: UI
- **Target**: ui/swimlane
- **Test ID**: ui/swimlane/empty-state-when-no-features
- **Pre-conditions**: A project exists but has zero features with valid index.json files
- **Route**: `/projects/:id`
- **Steps**:
  1. Navigate to the project swimlane page
  2. Verify no feature rows are rendered
  3. Verify an empty state message "No features found in this project" is displayed
- **Expected**: The page displays a centered empty state with "No features found in this project" message.
- **Priority**: P1

## TC-013: Filter by status shows only matching feature rows
- **Source**: Spec 5.2 Filter Controls + UI Function 2 States (Filtered)
- **Type**: UI
- **Target**: ui/swimlane
- **Test ID**: ui/swimlane/filter-by-status-shows-matching-rows
- **Pre-conditions**: A project has features with tasks in multiple statuses
- **Route**: `/projects/:id`
- **Steps**:
  1. Navigate to a project swimlane page
  2. Open the status filter dropdown
  3. Select "blocked" only
  4. Verify only feature rows containing blocked tasks are visible
  5. Non-matching feature rows are hidden (display:none)
  6. Clear the filter
  7. Verify all feature rows return to visible
- **Expected**: Filter toggles feature row visibility based on task-level status matching. Non-matching rows are removed from layout. Clearing filters restores all rows.
- **Priority**: P1

## TC-014: Filter by priority shows only matching feature rows
- **Source**: Spec 5.2 Filter Controls + UI Function 2 States (Filtered)
- **Type**: UI
- **Target**: ui/swimlane
- **Test ID**: ui/swimlane/filter-by-priority-shows-matching-rows
- **Pre-conditions**: A project has features with tasks of different priorities
- **Route**: `/projects/:id`
- **Steps**:
  1. Navigate to a project swimlane page
  2. Open the priority filter dropdown
  3. Select "P0" only
  4. Verify only feature rows containing P0 tasks are visible
  5. Clear the filter
  6. Verify all feature rows return to visible
- **Expected**: Priority filter toggles feature row visibility based on task-level priority matching.
- **Priority**: P1

## TC-015: Collapse and expand feature rows
- **Source**: Spec 5.2 Feature Row Controls + UI Function 2 States (Collapsed row)
- **Type**: UI
- **Target**: ui/swimlane
- **Test ID**: ui/swimlane/collapse-and-expand-feature-rows
- **Pre-conditions**: A project has at least one feature with tasks
- **Route**: `/projects/:id`
- **Steps**:
  1. Navigate to a project swimlane page
  2. Click the chevron icon on a feature row to collapse it
  3. Verify the row collapses to a summary bar showing feature slug, progress bar, and completion percentage
  4. Click the chevron to expand the row
  5. Verify the row expands to show all task cards
- **Expected**: Feature rows toggle between expanded (full task cards) and collapsed (summary bar with progress) states with animated transition.
- **Priority**: P1

## TC-016: Dependency arrows to non-existent tasks are skipped
- **Source**: UI Function 2 Validation Rules
- **Type**: UI
- **Target**: ui/swimlane
- **Test ID**: ui/swimlane/dependency-arrows-to-nonexistent-tasks-skipped
- **Pre-conditions**: A task has a dependency reference to a task ID that does not exist in the project
- **Route**: `/projects/:id`
- **Steps**:
  1. Navigate to a project swimlane page
  2. Verify no broken or dangling arrows are rendered for invalid dependency references
  3. Verify valid arrows still render correctly
- **Expected**: Dependency references to non-existent tasks are silently skipped. No broken arrows or error messages displayed.
- **Priority**: P2

## TC-017: Tasks with unrecognized phase grouped in Other column
- **Source**: UI Function 2 Validation Rules
- **Type**: UI
- **Target**: ui/swimlane
- **Test ID**: ui/swimlane/tasks-unrecognized-phase-grouped-other-column
- **Pre-conditions**: A project has tasks whose IDs do not match standard phase patterns (1.x, 2.x, 3.x+, T-*)
- **Route**: `/projects/:id`
- **Steps**:
  1. Navigate to a project swimlane page with non-standard task IDs
  2. Verify an "Other" phase column appears
  3. Verify tasks with unrecognized phase numbers are placed in the "Other" column
  4. Verify the "Other" column is hidden when no such tasks exist
- **Expected**: Non-standard phase tasks appear in a conditionally rendered "Other" column. The column is absent when no such tasks exist.
- **Priority**: P2

### Task Detail Slide-Over Panel (UF-3)

## TC-018: Detail panel slides in and displays task metadata
- **Source**: Story 4 / AC-1, Story 2 / AC-3 + Spec 5.3
- **Type**: UI
- **Target**: ui/swimlane
- **Test ID**: ui/swimlane/detail-panel-slides-in-displays-metadata
- **Pre-conditions**: A project has at least one task with a valid task markdown file
- **Route**: `/projects/:id`
- **Steps**:
  1. Navigate to a project swimlane page
  2. Click on a task card
  3. Verify the detail panel slides in from the right (~40% viewport width)
  4. Verify the panel displays: Task ID, Title, Status badge, Priority badge, Scope badge, Estimated time, Dependencies, Breaking indicator, File path, Record path
- **Expected**: The slide-over panel opens from the right side showing all task metadata fields from index.json with correct values and formatting.
- **Priority**: P0

## TC-019: Detail panel renders execution record as formatted markdown
- **Source**: Story 4 / AC-2, AC-3
- **Type**: UI
- **Target**: ui/swimlane
- **Test ID**: ui/swimlane/detail-panel-renders-execution-record-markdown
- **Pre-conditions**: A completed task has a record file referenced in index.json
- **Route**: `/projects/:id`
- **Steps**:
  1. Navigate to a project swimlane page
  2. Click on a completed task card that has a record file
  3. Verify the execution record section renders formatted markdown content
  4. Verify the record shows: summary, files created/modified, key decisions, test results
- **Expected**: The execution record section displays the full markdown-rendered content from the record file including summary, file changes, decisions, and test results.
- **Priority**: P0

## TC-020: Detail panel shows no execution record message
- **Source**: Story 4 / AC-4 + Spec 5.3 (No Record state)
- **Type**: UI
- **Target**: ui/swimlane
- **Test ID**: ui/swimlane/detail-panel-shows-no-execution-record
- **Pre-conditions**: A task exists with no record file path or a missing record file
- **Route**: `/projects/:id`
- **Steps**:
  1. Navigate to a project swimlane page
  2. Click on a task card that has no execution record
  3. Verify the execution record section displays "No execution record" message
- **Expected**: The panel explicitly displays "No execution record" with muted styling when no record exists.
- **Priority**: P1

## TC-021: Detail panel closes via X button, overlay click, and Escape key
- **Source**: Spec 5.3 Close Behavior
- **Type**: UI
- **Target**: ui/swimlane
- **Test ID**: ui/swimlane/detail-panel-closes-via-x-overlay-escape
- **Pre-conditions**: A project has at least one task
- **Route**: `/projects/:id`
- **Steps**:
  1. Navigate to a project swimlane page
  2. Click a task card to open the detail panel
  3. Click the X button — verify the panel closes
  4. Click another task card to reopen the panel
  5. Click outside the panel (on the overlay) — verify the panel closes
  6. Click another task card to reopen the panel
  7. Press the Escape key — verify the panel closes
- **Expected**: The detail panel closes successfully via all three methods (X button, overlay click, Escape key) with a slide-out animation.
- **Priority**: P0

## TC-022: Dependency links in detail panel navigate to referenced tasks
- **Source**: Story 2 / AC-4 + Spec 5.3 Dependencies
- **Type**: UI
- **Target**: ui/swimlane
- **Test ID**: ui/swimlane/dependency-links-navigate-to-referenced-tasks
- **Pre-conditions**: A task has at least one dependency on another task in the same or different feature
- **Route**: `/projects/:id`
- **Steps**:
  1. Open a task card that has dependencies
  2. Verify dependency IDs are displayed as clickable chips
  3. Click a dependency chip
  4. Verify the panel closes and the swimlane scrolls to the referenced task
  5. Verify the referenced task is highlighted with a ring animation
- **Expected**: Clicking a dependency chip closes the panel, scrolls the swimlane to the referenced task, and highlights it with a ring animation (persists 2000ms, fades over 300ms).
- **Priority**: P0

### Activity Sidebar (UF-4)

## TC-023: Activity sidebar displays recent status change events
- **Source**: Story 5 / AC-1, AC-2
- **Type**: UI
- **Target**: ui/swimlane
- **Test ID**: ui/swimlane/activity-sidebar-displays-recent-events
- **Pre-conditions**: A project has tasks with various statuses (in_progress, completed, blocked, skipped)
- **Route**: `/projects/:id`
- **Steps**:
  1. Navigate to a project swimlane page
  2. Verify the activity sidebar is visible on the right edge
  3. Verify events display with: timestamp, task ID, task title (truncated 40 chars), feature name, event type (claimed/completed/blocked/skipped)
  4. Verify events are sorted by timestamp descending
  5. Verify at most 50 events are displayed
- **Expected**: The activity sidebar shows up to 50 recent events sorted by timestamp descending, each with task ID, title, feature, and event type colored appropriately.
- **Priority**: P0

## TC-024: Clicking activity event scrolls to task and highlights
- **Source**: Story 5 / AC-3
- **Type**: UI
- **Target**: ui/swimlane
- **Test ID**: ui/swimlane/clicking-activity-event-scrolls-to-task
- **Pre-conditions**: The activity sidebar has at least one event
- **Route**: `/projects/:id`
- **Steps**:
  1. Navigate to a project swimlane page
  2. Expand the activity sidebar
  3. Click on an event entry
  4. Verify the swimlane scrolls to the corresponding task card
  5. Verify the task card is highlighted with a ring animation
- **Expected**: Clicking an activity event scrolls the swimlane to the referenced task card and highlights it with a ring animation.
- **Priority**: P0

## TC-025: Activity sidebar collapse and expand with blocked count badge
- **Source**: Spec 5.4 Toggle + UI Function 4 States (Collapsed/Expanded)
- **Type**: UI
- **Target**: ui/swimlane
- **Test ID**: ui/swimlane/activity-sidebar-collapse-expand-with-badge
- **Pre-conditions**: A project has at least one blocked task and activity events
- **Route**: `/projects/:id`
- **Steps**:
  1. Navigate to a project swimlane page
  2. Click the collapse button on the activity sidebar
  3. Verify the sidebar collapses to a thin bar (w-12)
  4. Verify a blocked count badge is displayed when collapsed (if blocked tasks exist)
  5. Click the expand button
  6. Verify the sidebar expands to full width (w-80) with the event list
- **Expected**: Sidebar toggles between expanded (event list) and collapsed (thin bar with blocked count badge) states with 200ms width transition.
- **Priority**: P1

### Integration Test Cases

## TC-026: Integration — Task Detail Panel visible on Swimlane Page
- **Source**: PRD UI Function "Task Detail Slide-Over Panel" Placement + Integration Spec
- **Type**: UI
- **Target**: ui/swimlane
- **Test ID**: ui/swimlane/integration-task-detail-panel
- **Pre-conditions**: Component build complete, integration task complete
- **Route**: `/projects/:id`
- **Steps**:
  1. Navigate to `/projects/:id`
  2. Click on any task card in the swimlane
  3. Verify the Task Detail Panel slides in from the right at ~40% viewport width
  4. Verify the panel overlays the swimlane content
  5. Verify the panel renders with expected task data (ID, title, status, priority)
- **Expected**: The Task Detail Panel appears at the right side of the swimlane page, overlaying the swimlane, and displays correct task data.
- **Priority**: P0

## TC-027: Integration — Activity Sidebar visible on Swimlane Page
- **Source**: PRD UI Function "Activity Sidebar" Placement + Integration Spec
- **Type**: UI
- **Target**: ui/swimlane
- **Test ID**: ui/swimlane/integration-activity-sidebar
- **Pre-conditions**: Component build complete, integration task complete
- **Route**: `/projects/:id`
- **Steps**:
  1. Navigate to `/projects/:id`
  2. Verify the Activity Sidebar is visible at the right edge of the page
  3. Verify the sidebar renders with event data (timestamps, task IDs, event types)
  4. Verify the sidebar is collapsible via the collapse button
- **Expected**: The Activity Sidebar appears at the right edge of the swimlane page, displays activity events, and can be collapsed and expanded.
- **Priority**: P0

---

## API Test Cases

## TC-028: GET /api/projects returns all configured projects
- **Source**: Story 6 / AC-1, AC-2 + Spec 5.5
- **Type**: API
- **Target**: api/projects
- **Test ID**: api/projects/get-returns-all-configured-projects
- **Pre-conditions**: Dashboard is running with 2+ configured projects
- **Steps**:
  1. Send GET /api/projects
  2. Verify response status is 200
  3. Verify response body is a JSON array
  4. Verify each entry contains: project name, feature count, completed tasks, total tasks, completion percentage, last updated timestamp
- **Expected**: 200 OK with JSON array of all configured projects with summary statistics. Response includes meta.lastUpdated.
- **Priority**: P0

## TC-029: GET /api/projects/:id returns single project with features
- **Source**: Spec 5.5 Endpoints
- **Type**: API
- **Target**: api/projects-id
- **Test ID**: api/projects-id/get-returns-single-project-with-features
- **Pre-conditions**: At least one project is configured with features
- **Steps**:
  1. Send GET /api/projects/:id with a valid project ID
  2. Verify response status is 200
  3. Verify the response includes project data with nested feature list
- **Expected**: 200 OK with project object containing feature list with task summaries.
- **Priority**: P0

## TC-030: GET /api/projects/:id/features returns feature list
- **Source**: Story 6 / AC-3 + Spec 5.5 Endpoints
- **Type**: API
- **Target**: api/projects-id-features
- **Test ID**: api/projects-id-features/get-returns-feature-list
- **Pre-conditions**: A project exists with features
- **Steps**:
  1. Send GET /api/projects/:id/features with a valid project ID
  2. Verify response status is 200
  3. Verify response body is a JSON array of feature summaries
- **Expected**: 200 OK with JSON array of feature summaries including slug, status, task counts.
- **Priority**: P0

## TC-031: GET /api/projects/:id/features/:slug/tasks returns task list
- **Source**: Story 6 / AC-3, AC-4 + Spec 5.5 Endpoints
- **Type**: API
- **Target**: api/projects-id-features-slug-tasks
- **Test ID**: api/projects-id-features-slug-tasks/get-returns-task-list
- **Pre-conditions**: A feature exists with tasks
- **Steps**:
  1. Send GET /api/projects/:id/features/:slug/tasks with valid project ID and feature slug
  2. Verify response status is 200
  3. Verify response body is a JSON array of task objects with full metadata
- **Expected**: 200 OK with JSON array of task objects containing id, title, status, priority, scope, dependencies, and other metadata fields.
- **Priority**: P0

## TC-032: API response time is under 200ms for single project
- **Source**: Story 6 / AC-4 + Spec Performance Requirements
- **Type**: API
- **Target**: api/projects-id
- **Test ID**: api/projects-id/response-time-under-200ms
- **Pre-conditions**: Dashboard is running with at least one project
- **Steps**:
  1. Send GET /api/projects/:id
  2. Measure the response time
  3. Verify response time is under 200ms
- **Expected**: API response for a single project query completes in under 200ms.
- **Priority**: P1

## TC-033: GET /api/projects/:id/features/:slug returns feature with tasks
- **Source**: Spec 5.5 Endpoints
- **Type**: API
- **Target**: api/projects-id-features-slug
- **Test ID**: api/projects-id-features-slug/get-returns-feature-with-tasks
- **Pre-conditions**: A feature exists with tasks
- **Steps**:
  1. Send GET /api/projects/:id/features/:slug with valid parameters
  2. Verify response status is 200
  3. Verify the response includes the feature object with nested tasks
- **Expected**: 200 OK with feature object containing nested task array.
- **Priority**: P1

## TC-034: GET /api/projects/:id/features/:slug/tasks/:taskId returns task details
- **Source**: Spec 5.5 Endpoints
- **Type**: API
- **Target**: api/projects-id-features-slug-tasks-taskid
- **Test ID**: api/projects-id-features-slug-tasks-taskid/get-returns-task-details
- **Pre-conditions**: A task exists in a feature
- **Steps**:
  1. Send GET /api/projects/:id/features/:slug/tasks/:taskId with valid parameters
  2. Verify response status is 200
  3. Verify the response contains the full task object with all metadata fields
- **Expected**: 200 OK with full task object including id, title, status, priority, scope, dependencies, breaking, file path, record path.
- **Priority**: P1

## TC-035: GET /api/projects/:id/features/:slug/dependencies returns dependency graph
- **Source**: Spec 5.5 Endpoints + Story 6 Agent Flow
- **Type**: API
- **Target**: api/projects-id-features-slug-dependencies
- **Test ID**: api/projects-id-features-slug-dependencies/get-returns-dependency-graph
- **Pre-conditions**: A feature has tasks with dependencies
- **Steps**:
  1. Send GET /api/projects/:id/features/:slug/dependencies with valid parameters
  2. Verify response status is 200
  3. Verify the response contains a nodes + edges representation of the dependency graph
- **Expected**: 200 OK with dependency graph data containing nodes (tasks) and edges (dependency relationships).
- **Priority**: P1

## TC-036: API returns 404 for non-existent project
- **Source**: Story 6 / AC-5 + Spec 5.5 Error Responses
- **Type**: API
- **Target**: api/projects-id
- **Test ID**: api/projects-id/returns-404-for-nonexistent-project
- **Pre-conditions**: Dashboard is running
- **Steps**:
  1. Send GET /api/projects/non-existent-project
  2. Verify response status is 404
  3. Verify response body contains {"error": "not_found", "message": "..."}
- **Expected**: 404 status with JSON error body containing "not_found" error code and descriptive message.
- **Priority**: P0

## TC-037: API returns 404 for non-existent feature
- **Source**: Story 6 / AC-5 + Spec 5.5 Error Responses
- **Type**: API
- **Target**: api/projects-id-features-slug
- **Test ID**: api/projects-id-features-slug/returns-404-for-nonexistent-feature
- **Pre-conditions**: A project exists but the feature slug does not
- **Steps**:
  1. Send GET /api/projects/:id/features/non-existent-feature
  2. Verify response status is 404
  3. Verify response body contains {"error": "not_found", "message": "..."}
- **Expected**: 404 status with JSON error body containing "not_found" error code.
- **Priority**: P0

## TC-038: API returns 404 for non-existent task
- **Source**: Story 6 / AC-5 + Spec 5.5 Error Responses
- **Type**: API
- **Target**: api/projects-id-features-slug-tasks-taskid
- **Test ID**: api/projects-id-features-slug-tasks-taskid/returns-404-for-nonexistent-task
- **Pre-conditions**: A feature exists but the task ID does not
- **Steps**:
  1. Send GET /api/projects/:id/features/:slug/tasks/non-existent-task
  2. Verify response status is 404
  3. Verify response body contains {"error": "not_found", "message": "..."}
- **Expected**: 404 status with JSON error body containing "not_found" error code.
- **Priority**: P0

---

## Traceability

| TC ID | Source | Type | Target | Priority |
|-------|--------|------|--------|----------|
| TC-001 | Story 1 / AC-1, AC-2 | UI | ui/landing | P0 |
| TC-002 | Story 1 / AC-2 | UI | ui/landing | P0 |
| TC-003 | Story 1 / AC-3 | UI | ui/landing | P0 |
| TC-004 | Spec 5.6 Validation + UF-1 States | UI | ui/landing | P1 |
| TC-005 | UF-1 States (Empty) | UI | ui/landing | P1 |
| TC-006 | UF-1 States (Loading) + Design C1 States | UI | ui/landing | P2 |
| TC-007 | Design C1 Interactions | UI | ui/landing | P2 |
| TC-008 | Story 2 / AC-1, AC-2 | UI | ui/swimlane | P0 |
| TC-009 | Story 2 / AC-3, AC-4 | UI | ui/swimlane | P0 |
| TC-010 | Story 3 / AC-1, AC-2 | UI | ui/swimlane | P0 |
| TC-011 | Spec 5.2 Status Colors | UI | ui/swimlane | P0 |
| TC-012 | UF-2 States (Empty) | UI | ui/swimlane | P1 |
| TC-013 | Spec 5.2 Filters + UF-2 States (Filtered) | UI | ui/swimlane | P1 |
| TC-014 | Spec 5.2 Filters + UF-2 States (Filtered) | UI | ui/swimlane | P1 |
| TC-015 | Spec 5.2 Row Controls + UF-2 States (Collapsed) | UI | ui/swimlane | P1 |
| TC-016 | UF-2 Validation Rules | UI | ui/swimlane | P2 |
| TC-017 | UF-2 Validation Rules | UI | ui/swimlane | P2 |
| TC-018 | Story 4 / AC-1 + Spec 5.3 | UI | ui/swimlane | P0 |
| TC-019 | Story 4 / AC-2, AC-3 | UI | ui/swimlane | P0 |
| TC-020 | Story 4 / AC-4 + Spec 5.3 (No Record) | UI | ui/swimlane | P1 |
| TC-021 | Spec 5.3 Close Behavior | UI | ui/swimlane | P0 |
| TC-022 | Story 2 / AC-4 + Spec 5.3 Dependencies | UI | ui/swimlane | P0 |
| TC-023 | Story 5 / AC-1, AC-2 | UI | ui/swimlane | P0 |
| TC-024 | Story 5 / AC-3 | UI | ui/swimlane | P0 |
| TC-025 | Spec 5.4 Toggle + UF-4 States | UI | ui/swimlane | P1 |
| TC-026 | UF-3 Placement + Integration | UI | ui/swimlane | P0 |
| TC-027 | UF-4 Placement + Integration | UI | ui/swimlane | P0 |
| TC-028 | Story 6 / AC-1, AC-2 + Spec 5.5 | API | api/projects | P0 |
| TC-029 | Spec 5.5 Endpoints | API | api/projects-id | P0 |
| TC-030 | Story 6 / AC-3 + Spec 5.5 | API | api/projects-id-features | P0 |
| TC-031 | Story 6 / AC-3, AC-4 + Spec 5.5 | API | api/projects-id-features-slug-tasks | P0 |
| TC-032 | Story 6 / AC-4 + Spec Performance | API | api/projects-id | P1 |
| TC-033 | Spec 5.5 Endpoints | API | api/projects-id-features-slug | P1 |
| TC-034 | Spec 5.5 Endpoints | API | api/projects-id-features-slug-tasks-taskid | P1 |
| TC-035 | Spec 5.5 Endpoints + Story 6 | API | api/projects-id-features-slug-dependencies | P1 |
| TC-036 | Story 6 / AC-5 + Spec 5.5 Errors | API | api/projects-id | P0 |
| TC-037 | Story 6 / AC-5 + Spec 5.5 Errors | API | api/projects-id-features-slug | P0 |
| TC-038 | Story 6 / AC-5 + Spec 5.5 Errors | API | api/projects-id-features-slug-tasks-taskid | P0 |

---

## Route Validation

| Route | Status | TC IDs | Matched Route |
|-------|--------|--------|---------------|
| `/` | Matched | TC-001 to TC-007 | `r.GET("/", handleLanding(s))` (internal/handler/page.go:36) |
| `/projects/:id` | Matched | TC-008 to TC-027 | `r.GET("/projects/:id", handleProject(s))` (internal/handler/page.go:37) |
| `/api/projects` | Matched | TC-028 | `api.GET("/projects", handleListProjects(s))` (internal/handler/api.go:18) |
| `/api/projects/:id` | Matched | TC-029, TC-032, TC-036 | `api.GET("/projects/:id", handleGetProject(s))` (internal/handler/api.go:21) |
| `/api/projects/:id/features` | Matched | TC-030 | `api.GET("/projects/:id/features", handleListFeatures(s))` (internal/handler/api.go:24) |
| `/api/projects/:id/features/:slug` | Matched | TC-033, TC-037 | `featureRoutes.GET("", handleGetFeature(s))` (internal/handler/api.go:30) |
| `/api/projects/:id/features/:slug/tasks` | Matched | TC-031 | `featureRoutes.GET("/tasks", handleListTasks(s))` (internal/handler/api.go:33) |
| `/api/projects/:id/features/:slug/tasks/:taskId` | Matched | TC-034, TC-038 | `taskRoutes.GET("", handleGetTask(s))` (internal/handler/api.go:40) |
| `/api/projects/:id/features/:slug/dependencies` | Matched | TC-035 | `featureRoutes.GET("/dependencies", handleGetDependencies(s))` (internal/handler/api.go:44) |
