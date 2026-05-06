---
feature: "agent-task-dashboard"
---

# User Stories: Agent Task Dashboard

## Story 1: View project overview

**As a** Operator
**I want to** see all configured projects on a single landing page with task completion statistics
**So that** I can quickly assess which projects need attention without navigating the filesystem

**Acceptance Criteria:**
- Given the dashboard is running and 2+ projects are configured
- When I open the dashboard URL in a browser
- Then I see one card per project showing: name, feature count, task completion %, last update time
- And each card health status is correctly derived (active/complete/stale)
- And clicking a project card navigates to that project's swimlane view

---

## Story 2: Diagnose a blocked task

**As a** Operator
**I want to** see blocked tasks prominently on the swimlane with their dependency chains
**So that** I can identify root causes and unblock agents within minutes instead of hours

**Acceptance Criteria:**
- Given a project has at least one blocked task
- When I view the project's swimlane
- Then blocked task cards appear in red
- And features with blocked tasks sort to the top of the swimlane
- And clicking a blocked task opens the detail panel showing its dependency list
- And each dependency is clickable to navigate to the upstream task

---

## Story 3: Trace cross-feature dependencies

**As a** Operator
**I want to** see dependency arrows between tasks across different features on the swimlane
**So that** I can understand how a change in one feature affects tasks in another

**Acceptance Criteria:**
- Given task A in feature X depends on task B in feature Y
- When I view the swimlane with both features visible
- Then a dashed dependency arrow connects task A to task B across the feature rows
- And the arrow is directional (pointing from dependent to dependency)

---

## Story 4: View task execution record

**As a** Operator
**I want to** read a task's execution record in the detail panel
**So that** I can understand what the agent did, what files were changed, and what decisions were made

**Acceptance Criteria:**
- Given a task has been completed and has a record file referenced in index.json
- When I click the task card on the swimlane
- Then the slide-over panel renders the execution record as formatted markdown
- And the panel shows: summary, files created/modified, key decisions, test results
- And if no execution record exists, the panel displays "No execution record" explicitly

---

## Story 5: Monitor recent activity

**As a** Operator
**I want to** see a chronological feed of recent task status changes across all features
**So that** I can track agent progress without checking each feature individually

**Acceptance Criteria:**
- Given I am viewing a project's swimlane
- When I expand the activity sidebar
- Then I see the 50 most recent status change events with timestamps
- And each event shows: task ID, task title, feature name, event type (claimed/completed/blocked/skipped)
- And clicking an event scrolls the swimlane to that task card

---

## Story 6: Query task state via API

**As a** Agent
**I want to** query project, feature, and task data through REST API endpoints
**So that** I can make informed decisions about which tasks to execute next without direct filesystem access

**Acceptance Criteria:**
- Given the dashboard is running
- When I send GET /api/projects
- Then I receive a JSON array of all configured projects with summary statistics
- When I send GET /api/projects/:id/features/:slug/tasks
- Then I receive a JSON array of all tasks for that feature with full metadata
- And response time is under 200ms for a single project
- And a 404 is returned for non-existent project, feature, or task IDs

---

## Story 7: Configure project list

**As a** Operator
**I want to** add or remove projects from the dashboard by editing a configuration file
**So that** I can manage which projects appear without modifying dashboard code

**Acceptance Criteria:**
- Given the config file lists 2 projects with valid paths
- When I add a third project entry with name and path, then save the file
- And I refresh the dashboard landing page
- Then the third project card appears with correct statistics
- And if a path does not exist, the project is skipped with a visible warning on the landing page

---
