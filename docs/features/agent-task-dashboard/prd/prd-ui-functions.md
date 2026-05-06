---
feature: "agent-task-dashboard"
---

# Agent Task Dashboard — UI Functions

> Requirements layer: defines WHAT the UI must do. Not HOW it looks (that's ui-design.md).

## UI Scope

Three pages: landing page (project cards), project swimlane page (feature rows + dependency DAG + activity sidebar), and no additional standalone pages. The task detail panel is a slide-over overlay on the swimlane page.

## UI Function 1: Project Card Grid

### Placement

- **Mode**: new-page
- **Target Page**: `/` (landing page)
- **Position**: Full page content area

### Description

Displays all configured projects as cards in a responsive grid. Each card shows project name, feature count, task completion progress bar, and last update timestamp. Cards are clickable to navigate to the project's swimlane view.

### User Interaction Flow

1. User opens `http://localhost:PORT/` in browser
2. Dashboard reads config file and scans each project's feature directories
3. Cards render with computed statistics
4. User clicks a card → browser navigates to `/projects/:id`

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| Project name | string | Config file | Display name |
| Project path | string | Config file | Used for filesystem scanning |
| Feature count | number | Derived | Count of dirs with index.json |
| Completed tasks | number | Derived | Sum across all features |
| Total tasks | number | Derived | Sum across all features |
| Last updated | datetime | Derived | Max mtime of all index.json files |
| Health status | string | Derived | active / complete / stale |

### States

| State | Display | Trigger |
|-------|---------|---------|
| Loading | Skeleton cards | Initial page load |
| Populated | Cards with data | Data loaded successfully |
| Error | Warning banner per project | Config path invalid or no features found |
| Empty | "No projects configured" message | Config file has zero project entries |

### Validation Rules

- Project path must exist on filesystem; otherwise show warning card
- Each project must contain `docs/features/` directory; otherwise show warning card

---

## UI Function 2: Feature Swimlane Board

### Placement

- **Mode**: new-page
- **Target Page**: `/projects/:id`
- **Position**: Full page content area (main body)

### Description

Renders a feature-per-row swimlane with phase columns. Each feature is a horizontal row; task cards are placed in columns based on their phase number. Dependency arrows connect tasks with directed lines. Task cards are color-coded by status. Rows are collapsible. Filter controls allow showing only features with specific statuses or priorities.

### User Interaction Flow

1. User navigates to project page (from landing card click or direct URL)
2. Dashboard scans all `index.json` files in the project's `docs/features/` directory
3. Swimlane renders: rows sorted by blocked count (desc), then completion % (asc)
4. Task cards appear in phase columns with status colors
5. Dependency arrows render for visible tasks
6. User can: collapse/expand feature rows, filter by status/priority, click task cards
7. User clicks a task card → UF-3 (detail panel) opens

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| Feature slug | string | index.json `feature` | Row label |
| Feature status | string | index.json `status` | planning / in-progress / completed |
| Task ID | string | index.json task `id` | Card label |
| Task title | string | index.json task `title` | Card body (truncated) |
| Task status | string | index.json task `status` | Color coding |
| Task priority | string | index.json task `priority` | Card badge |
| Task phase | number | Derived from ID prefix | Column placement |
| Dependencies | array | index.json task `dependencies` | Arrow rendering |

### States

| State | Display | Trigger |
|-------|---------|---------|
| Loading | Skeleton rows | Page load |
| Populated | Full swimlane with cards and arrows | Data loaded |
| Empty | "No features found in this project" | Zero features with valid index.json |
| Filtered | Only matching feature rows visible | User applies filter |
| Collapsed row | Summary bar only (completion %) | User collapses a feature row |

### Validation Rules

- Tasks with unrecognized phase numbers grouped into "Other" column
- Dependency references to non-existent tasks are silently skipped (no broken arrows)

---

## UI Function 3: Task Detail Slide-Over Panel

### Placement

- **Mode**: existing-page
- **Target Page**: `/projects/:id`
- **Position**: Right side overlay, ~40% viewport width, over the swimlane

### Description

A slide-over panel that displays full task details when a task card is clicked. Shows task metadata from index.json, parsed acceptance criteria from the task markdown file, and rendered execution record from the record file. Dependency IDs are clickable links that scroll the swimlane to the referenced task.

### User Interaction Flow

1. User clicks a task card on the swimlane
2. Panel slides in from the right
3. Dashboard reads the task markdown file and record file (if referenced)
4. Panel displays: metadata table, acceptance criteria, execution record
5. User clicks a dependency link → swimlane scrolls to that task and highlights it
6. User closes panel via X button, clicking outside, or pressing Escape

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| Task ID | string | index.json | Header |
| Title | string | index.json | Header |
| Status | string | index.json | Color badge |
| Priority | string | index.json | P0/P1/P2 badge |
| Scope | string | index.json | frontend/backend/all |
| Estimated Time | string | index.json | Optional |
| Dependencies | array of strings | index.json | Clickable links |
| Breaking | boolean | index.json | Warning indicator if true |
| File path | string | index.json | Relative path display |
| Record path | string | index.json | Relative path display |
| Acceptance criteria | markdown | Task .md file | Parsed from body |
| Execution record | markdown | Record .md file | Rendered if exists; "No execution record" if absent |

### States

| State | Display | Trigger |
|-------|---------|---------|
| Open | Panel visible with data | Task card clicked |
| Loading | Spinner in panel | Reading task files |
| Populated | Full task details | Files loaded |
| No Record | "No execution record" message | Record path empty or file missing |
| Closed | Panel hidden | X button / outside click / Escape key |

### Validation Rules

- If task markdown file cannot be read, show "Unable to load task details" error in panel
- If record file path exists but file is missing, show "No execution record"

---

## UI Function 4: Activity Sidebar

### Placement

- **Mode**: existing-page
- **Target Page**: `/projects/:id`
- **Position**: Right edge of page, collapsible, adjacent to swimlane (outside UF-3 panel)

### Description

A collapsible sidebar showing the 50 most recent task status change events across all features in the current project. Each event shows timestamp, task ID, task title, feature name, and event type. Events are clickable to scroll the swimlane to the corresponding task. When collapsed, shows a badge with count of currently blocked tasks.

### User Interaction Flow

1. Sidebar is visible by default on the swimlane page
2. Events populate from all features' index.json data (inferred from status + mtime)
3. User clicks an event → swimlane scrolls to that task and briefly highlights it
4. User clicks collapse button → sidebar collapses to a thin bar with blocked count badge
5. User clicks expand button → sidebar reopens

### Data Requirements

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| Timestamp | datetime | index.json mtime | Event time |
| Task ID | string | index.json task `id` | Clickable |
| Task title | string | index.json task `title` | Truncated to 40 chars |
| Feature slug | string | index.json `feature` | Label |
| Event type | string | Derived from status | claimed/completed/blocked/skipped |
| Blocked count | number | Derived | Badge when collapsed |

### States

| State | Display | Trigger |
|-------|---------|---------|
| Expanded | Full event list | Default or user expands |
| Collapsed | Thin bar + blocked badge | User collapses |
| Empty | "No activity yet" message | Zero tasks with status changes |
| Loading | Skeleton events | Initial page load |

### Validation Rules

- Maximum 50 events displayed, sorted by timestamp descending
- If mtime cannot distinguish event order, group by feature alphabetically

---

## Page Composition

| Page | Type | UI Functions | Position Notes |
|------|------|-------------|----------------|
| `/` (Landing) | new | UF-1 | Full page — project card grid |
| `/projects/:id` (Swimlane) | new | UF-2, UF-3, UF-4 | UF-2: main content area; UF-3: right overlay (40% width); UF-4: right sidebar (collapsible) |
