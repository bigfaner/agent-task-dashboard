---
created: 2026-05-06
source: prd/prd-ui-functions.md
status: Draft
---

# UI Design: Agent Task Dashboard

## Design System

**Style**: Shadcn — zinc neutral, functional minimalism. Radix + Tailwind primitives.

### Visual Theme

Clean and utilitarian. Zero decoration — every visual element serves a function. Surfaces are flat with subtle 1px borders. Light mode default with dark mode toggle.

### Color Palette

**Light Mode:**

| Role | Variable | Value |
|------|----------|-------|
| Background | --background | #ffffff |
| Foreground | --foreground | #09090b (zinc-950) |
| Card | --card | #ffffff |
| Card Foreground | --card-foreground | #09090b |
| Primary | --primary | #18181b (zinc-900) |
| Primary Foreground | --primary-foreground | #fafafa |
| Secondary | --secondary | #f4f4f5 (zinc-100) |
| Muted | --muted | #f4f4f5 |
| Muted Foreground | --muted-foreground | #71717a (zinc-500) |
| Destructive | --destructive | #ef4444 |
| Border | --border | #e4e4e7 (zinc-200) |
| Ring | --ring | #18181b |

**Dark Mode:**

| Role | Variable | Value |
|------|----------|-------|
| Background | --background | #09090b |
| Foreground | --foreground | #fafafa |
| Card | --card | #09090b |
| Primary | --primary | #fafafa |
| Secondary | --secondary | #27272a (zinc-800) |
| Muted | --muted | #27272a |
| Muted Foreground | --muted-foreground | #a1a1aa (zinc-400) |
| Border | --border | #27272a |

**Status Colors (light/dark):**

| Status | Light | Dark | Usage |
|--------|-------|------|-------|
| pending | #71717a (zinc-500) | #a1a1aa (zinc-400) | Grey dot + text |
| in_progress | #2563eb (blue-600) | #3b82f6 (blue-500) | Blue dot + text |
| completed | #16a34a (green-600) | #22c55e (green-500) | Green dot + text |
| blocked | #dc2626 (red-600) | #ef4444 (red-500) | Red dot + text + pulse |
| skipped | #ca8a04 (yellow-600) | #eab308 (yellow-500) | Yellow dot + strikethrough |

### Typography

| Role | Weight | Size | Line Height |
|------|--------|------|-------------|
| H1 | 700 | 36px | 1.2 |
| H2 | 600 | 30px | 1.25 |
| H3 | 600 | 24px | 1.3 |
| H4 | 600 | 20px | 1.35 |
| Body | 400 | 16px | 1.5 |
| Small | 500 | 14px | 1.5 |
| Mono | 400 | 14px | 1.5 |

Font stack: Inter / system stack. Mono: JetBrains Mono.

### Components

- **Cards**: bg card, rounded-lg (8px), border 1px solid var(--border), p-6, no shadow
- **Badges**: rounded-full, px-2.5 py-0.5, text-xs font-medium
- **Buttons**: rounded-md (6px), h-10, px-4, text-sm font-medium, 150ms hover transition
- **Inputs**: border 1px solid var(--input), rounded-md, h-10, px-3, focus ring 2px

---

## Component 1: Project Card Grid

### Placement

- **Mode**: new-page
- **Target**: `/` (landing page)
- **Position**: Full page content area

### Layout Structure

```
┌─────────────────────────────────────────────────────────┐
│ Header: "Task Dashboard" + dark mode toggle              │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌─────────┐ │
│  │ Card     │  │ Card     │  │ Card     │  │ Card    │ │
│  │ Name     │  │ Name     │  │ Name     │  │ Name    │ │
│  │ Stats    │  │ Stats    │  │ Stats    │  │ Stats   │ │
│  │ Progress │  │ Progress │  │ Progress │  │ Progress│ │
│  │ Updated  │  │ Updated  │  │ Updated  │  │ Updated │ │
│  └──────────┘  └──────────┘  └──────────┘  └─────────┘ │
│                                                          │
│  Grid: auto-fill, minmax(320px, 1fr), gap-6             │
└─────────────────────────────────────────────────────────┘
```

**Header**: Fixed top bar, h-14, bg-background, border-b 1px var(--border). Left: "Task Dashboard" H3. Right: sun/moon icon button for dark mode toggle.

**Card grid**: CSS Grid, auto-fill columns, minmax(320px, 1fr), gap-6 (24px). Padding: px-6 py-8.

**Project Card internals**:
- **Top**: Project name (H4, font-semibold) + health status badge (right-aligned)
- **Middle**: Stat row — "X features" (muted) | "XX / XX tasks" (body)
- **Progress bar**: Full width, h-2, rounded-full. Track: var(--secondary). Fill: var(--primary) with width = completion %. Green fill when >= 80%.
- **Bottom**: "Updated 2h ago" (muted-foreground, text-xs). Relative time display.
- **Hover**: Cursor pointer, border color transitions to var(--ring), 150ms
- **Click**: Navigate to `/projects/:id`

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Loading | 6 skeleton cards (pulse animation, bg-muted rounded-lg, h-48) | Shown on initial page load |
| Populated | Cards with project data, progress bars, health badges | Transition from Loading: skeletons fade out (150ms opacity), cards fade in (150ms opacity). Grid renders all configured projects. |
| Error | Warning banner per errored project, rendered above the card grid. Banner: bg-secondary, border-l 4px var(--destructive), rounded-md, p-3, flex row. Alert-triangle icon (var(--destructive)) + "Path not found: {project name}" (text-sm) + file path in mono (text-xs, muted-foreground). Non-navigable. Multiple banners stack vertically for multiple errored projects. | Click on file path copies it to clipboard; banner persists until config is fixed |
| Empty | Centered layout: folder-open icon (48px, muted) + "No projects configured" H4 + "Add projects to ~/.task-dashboard.yaml" muted text | Transition from Loading: skeletons fade out (150ms opacity), empty state fades in (150ms opacity). No grid rendered. |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| Click card | Navigate to `/projects/:id` | Border highlights on hover |
| Dark mode toggle | Switch theme | CSS variable swap, 150ms transition |
| Page load | Fetch project data | Skeleton → populated transition |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| Project name | name | Config file |
| Feature count | features.length | Derived from dirs with index.json |
| Completed / Total | completedTasks / totalTasks | Derived sum |
| Progress bar width | (completedTasks / totalTasks) * 100 | Calculated |
| Last updated | maxMtime | Max mtime of all index.json files |
| Health badge | healthStatus | Derived: active (green dot) / complete (blue dot) / stale (grey dot) |

---

## Component 2: Feature Swimlane Board

### Placement

- **Mode**: new-page
- **Target**: `/projects/:id`
- **Position**: Full page content area (main body), shares page with UF-3 and UF-4

### Layout Structure

```
┌──────────────────────────────────────────────────────────────────────┐
│ Header: Breadcrumb "Dashboard > Project Name" + filter controls      │
├──────────────────────────────────────────────────────────────────┬───┤
│                                                                  │A  │
│  Phase Headers: | Phase 1 | Phase 2 | Phase 3+ | Testing | Other|  │c  │
│  ─────────────────────────────────────────────────────────────   │t  │
│  ▼ feature-slug-1  [====35%====]                                 │i  │
│    ┌─────┐  ┌─────┐  ┌─────┐  ┌─────┐                          │v  │
│    │1.1  │──│2.1  │  │3.1  │  │T-1  │                          │i  │
│    └─────┘  └─────┘  └─────┘  └─────┘                          │t  │
│  ─────────────────────────────────────────────────────────────   │y  │
│  ▼ feature-slug-2  [====80%====]                                 │   │
│    ┌─────┐  ┌─────┐                                             │S  │
│    │1.1  │  │2.1  │                                             │i  │
│    └─────┘  └─────┘                                             │d  │
│  ···                                                             │e  │
│                                                                  │b  │
│  ← Dependency arrows between task cards →                        │a  │
│                                                                  │r  │
└──────────────────────────────────────────────────────────────────┴───┘
```

**Page layout**: Flex row. Swimlane area: flex-1. Activity sidebar: w-80 (320px, collapsible to w-12).

**Responsive breakpoints**:
- **Min supported viewport**: 768px. Below 768px, display a full-width message: "This dashboard requires a wider screen (minimum 768px)."
- **>= 1280px** (default): Sidebar expanded (w-80), swimlane flex-1, phase columns evenly distributed.
- **1024px -- 1279px**: Sidebar auto-collapses to w-12 (48px) icon bar. Swimlane area gains the freed 272px. Phase columns remain evenly distributed.
- **768px -- 1023px**: Sidebar hidden entirely (display:none). Phase columns become individually scrollable -- the swimlane area uses `overflow-x: auto` and phase columns have a min-width of 160px each (5 columns = 800px scrollable width). A horizontal scrollbar appears within the swimlane area only.

**Header bar**: h-14, border-b. Left: breadcrumb navigation (Dashboard > ProjectName, links). Right: filter controls.

**Filter controls**:
- Status filter: multi-select dropdown with checkboxes (pending / in_progress / completed / blocked / skipped). Badge shows active count.
- Priority filter: multi-select dropdown (P0 / P1 / P2).
- Both use shadcn dropdown pattern: outline button trigger, bg-popover dropdown.

**Filter matching logic**:
- A feature row is visible when it contains at least one task whose status is included in the active status filter AND whose priority is included in the active priority filter. Matching is at the task level, not the feature level -- a feature row with status "in_progress" but containing a "blocked" task will show when the "blocked" filter is active, because a child task matches.
- When no filters are active (all checkboxes unchecked, or filter dropdown not opened), all feature rows are visible.
- Non-matching feature rows are removed from layout with `display:none` (not dimmed). This avoids half-visible rows cluttering the board.
- Task cards within a visible feature row are never individually filtered out -- the filter operates on feature-row granularity based on task-level matching.

**Swimlane area**:
- **Phase header row**: sticky top, bg-muted, h-10. Columns: "Phase 1" | "Phase 2" | "Phase 3+" | "Testing" | "Other". Even width distribution. Text: small font-medium, muted-foreground, uppercase tracking-wide. "Other" column header is conditionally rendered -- only visible when at least one task in the project has a phase number that does not match Phase 1 (1.x), Phase 2 (2.x), Phase 3+ (3.x+), or Testing (T-*). When the "Other" column is absent, remaining columns redistribute evenly.
- **Feature rows**: One per feature. Each row has two states: expanded (default) and collapsed.

**Expanded row**:
- **Row header**: h-10, flex, border-b. Left: chevron-down icon + feature slug (font-medium). Right: completion badge (small, "35%").
- **Task area**: min-height 80px, padding-y 12px. CSS Grid matching phase columns. Task cards positioned in their phase column.
- **Dependency arrows**: SVG overlay positioned absolute within the swimlane area. Lines use the task card centers as endpoints. Within-feature: solid 1.5px line with arrowhead. Cross-feature: dashed 1.5px line. Arrow color: var(--muted-foreground). Blocked dependency: var(--destructive).
  - **Z-index layering**: SVG overlay at z-10, task cards at z-20 (relative). Arrows render behind task cards so cards remain clickable and arrows do not intercept pointer events (`pointer-events: none` on SVG element).
  - **Collapsed-row behavior**: When a feature row is collapsed, any dependency arrow with an endpoint in that row is hidden. An arrow does not re-render until the row is expanded again. This avoids arrows pointing to invisible targets.
  - **Dense graph simplification**: When a project has more than 50 visible dependency arrows, cross-feature arrows are collapsed into a single summary indicator per feature row -- a small badge showing the dependency count (e.g., "3 deps") at the row header level, instead of rendering individual dashed lines. Within-feature arrows always render regardless of count. The threshold prevents SVG repaint cost from degrading scroll performance.

**Collapsed row**:
- Single h-10 bar: chevron-right icon + feature slug + inline progress bar (w-32, h-1.5) + percentage text. Cursor pointer to expand.

**Task card**:
- Width: 100% of column cell (min 140px). bg-card, border 1px var(--border), rounded-md, p-3.
- **Top**: Task ID (mono, text-xs, muted-foreground) + priority badge (right, text-xs, variants: P0 destructive, P1 secondary, P2 outline).
- **Body**: Task title (text-sm, truncated to 2 lines, line-clamp-2).
- **Bottom**: Status dot (8px circle, color per status) + status text (text-xs). Blocked cards: pulse animation on the status dot.
- **Hover**: border → var(--ring), cursor pointer.
- **Click**: Opens UF-3 (task detail panel).
- **Active/highlighted**: ring 2px var(--ring), persists 2000ms then fades out over 300ms opacity transition — triggered when navigated via dependency link or activity event.

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Loading | Skeleton: 8 feature row placeholders (pulse animation, h-32, bg-muted, rounded-md) | Shown on initial page load |
| Populated | Full swimlane with feature rows, task cards, dependency arrows | Transition from Loading: skeletons fade out (150ms opacity), content fades in (150ms opacity). Rows sorted by blocked count desc, then completion % asc |
| Empty | Centered: "No features found in this project" H4 + "Check that docs/features/ contains valid task data" muted text | Transition from Loading: skeletons fade out (150ms opacity), empty state fades in (150ms opacity). Shown when zero features with valid index.json |
| Filtered | Only matching feature rows visible, others hidden with display:none | Transition from Populated: non-matching rows collapse with 150ms height animation then display:none. Filter badges show active filter count |
| Collapsed row | Summary bar: slug + progress bar + percentage | Click to expand, animated 200ms height transition |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| Click task card | Open UF-3 detail panel | Card gets ring highlight. Rapid clicks debounced: second click within 300ms of first is ignored. |
| Click row chevron | Toggle collapse/expand | 200ms slide animation |
| Status filter change | Show feature rows containing tasks matching selected statuses; hide non-matching rows via display:none | Non-matching rows collapse with 150ms height animation then display:none |
| Priority filter change | Show feature rows containing tasks matching selected priorities; hide non-matching rows via display:none | Non-matching rows collapse with 150ms height animation then display:none |
| Navigate to task (from UF-3/UF-4) | Scroll to task card, highlight | Ring animation: 2px solid var(--ring), persists 2000ms then fades 300ms. Auto-scroll into view. If target task is in a collapsed row, row expands first. |
| Breadcrumb "Dashboard" | Navigate to `/` | Standard navigation |
| Hover task card | Border highlight | 150ms transition |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| Feature slug | feature | index.json |
| Feature status | status | index.json |
| Completion % | (completed / total) * 100 | Derived per feature |
| Task ID | task.id | index.json task |
| Task title (truncated) | task.title | index.json task, CSS line-clamp-2 |
| Task status color | task.status | Mapped to status color palette |
| Task priority badge | task.priority | index.json task |
| Task phase column | Derived from ID prefix | 1.x → Phase 1, 2.x → Phase 2, 3.x+ → Phase 3+, T-* → Testing, any unmatched prefix → Other |
| Dependency arrows | task.dependencies | index.json task, rendered as SVG lines |

---

## Component 3: Task Detail Slide-Over Panel

### Placement

- **Mode**: existing-page
- **Target**: `/projects/:id`
- **Position**: Right side overlay, ~40% viewport width (min 480px), over the swimlane. Separate from UF-4 activity sidebar — when panel is open, activity sidebar is hidden or slides further right (off-screen).

### Layout Structure

```
┌─────────────────────────────────┐
│ ✕ Close    Task 1.1             │  ← Header: close button + task ID
│─────────────────────────────────│
│ Task Title Here                 │  ← H3, full title
│ [in_progress] [P0] [backend]    │  ← Status + Priority + Scope badges
│─────────────────────────────────│
│                                 │
│ Details                         │  ← Section H4
│ ┌─────────────────────────────┐ │
│ │ Est. Time   2h              │ │
│ │ Breaking    Yes ⚠           │ │
│ │ Dependencies 1.2, 2.1       │ │  ← Dependency IDs are clickable links
│ │ File Path   tasks/1.1.md    │ │  ← Mono font
│ │ Record      records/1.1.md  │ │  ← Mono font
│ └─────────────────────────────┘ │
│                                 │
│ Acceptance Criteria             │  ← Section H4
│ ┌─────────────────────────────┐ │
│ │ • Criterion 1               │ │  ← Rendered from task .md body
│ │ • Criterion 2               │ │
│ │ • Criterion 3               │ │
│ └─────────────────────────────┘ │
│                                 │
│ Execution Record                │  ← Section H4
│ ┌─────────────────────────────┐ │
│ │ Rendered markdown content   │ │  ← From record .md file
│ │ or "No execution record"    │ │
│ └─────────────────────────────┘ │
│                                 │
└─────────────────────────────────┘
```

**Panel container**: Fixed position, right-0, top-0, h-screen, w-[40vw] min-w-[480px]. bg-background, border-l 1px var(--border). z-40 (above swimlane, below modals).

**Overlay**: bg-black/20, backdrop-blur-sm, covers the swimlane area. Clicking overlay closes panel.

**Animation**: Slide in from right, transform translateX(100%) → translateX(0), 200ms ease-out.

**Header**: h-14, border-b, flex. Left: X button (ghost icon button). Right: Task ID (mono, text-sm, muted-foreground).

**Content**: Overflow-y-auto, p-6, space-y-6.

**Metadata table**: Two-column layout. Label (muted-foreground, text-sm, w-32) | Value (text-sm). Each row: py-2, border-b last:border-b-0.

**Dependency links**: Inline clickable chips. bg-secondary, rounded-md, px-2 py-1, text-sm, mono font. Hover: bg-accent. Click: closes panel (200ms slide-out), scrolls swimlane to referenced task at 100ms into close, highlights with ring (2px var(--ring), persists 2000ms then fades 300ms). If target task is in a collapsed feature row, the row expands first.

**Acceptance criteria**: Rendered as a list. Each item: flex row with check-circle icon (muted) + text.

**Execution record**: Rendered markdown with proper typography. Code blocks: bg-muted, rounded-md, p-3, mono font, overflow-x-auto. If no record: muted text "No execution record" with file-text icon.

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Open | Panel visible, overlay over swimlane | Slides in 200ms ease-out. Focus moves to panel container. |
| Loading | Panel frame visible, content area shows spinner (24px, border-2, animate-spin) | Shown while reading task .md and record files. Transition from Open: spinner appears immediately. |
| Populated | Full task details with all sections | Transition from Loading: spinner fades out (100ms opacity), content fades in (150ms opacity). |
| No Record | "No execution record" with muted icon | Record path empty or file missing. Rendered in place of Execution Record section. |
| Close | Panel slides out right, overlay fades | 200ms ease-in. Focus returns to originating task card. |
| Error | "Unable to load task details" alert (destructive variant) with retry button | Transition from Loading: spinner fades out (100ms opacity), error alert fades in (150ms opacity). Task .md file unreadable. |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| Click task card (UF-2) | Open panel | Slide-in animation |
| Click X button | Close panel | Slide-out animation |
| Click overlay | Close panel | Slide-out animation |
| Press Escape | Close panel | Slide-out animation |
| Click dependency link | Close panel → scroll to task → highlight | **Sequence**: (1) Panel begins 200ms slide-out. (2) At 100ms (panel 50% closed), swimlane scroll starts via `scrollIntoView({ behavior: 'smooth', block: 'center' })`. If target task is in a collapsed feature row, the row expands first (200ms height animation, blocking scroll until complete). (3) After scroll settles (~300ms from initiation), ring highlight appears on target task card: 2px solid var(--ring), persists for 2000ms, then fades out over 300ms opacity transition. Focus moves to the highlighted task card. |
| Click retry button (Error state) | Re-fetch task .md and record files | Alert fades out (100ms), spinner appears (Loading state). If retry succeeds, transitions to Populated. If retry fails, transitions back to Error with updated alert. |
| Scroll content | Vertical scroll within panel | Swimlane stays fixed underneath |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| Task ID (header) | id | index.json task |
| Title | title | index.json task |
| Status badge | status | index.json task → color mapping |
| Priority badge | priority | index.json task |
| Scope badge | scope | index.json task |
| Estimated time | estimatedTime | index.json task |
| Breaking indicator | breaking | index.json task — "Yes" shown with warning icon if true |
| Dependencies | dependencies | index.json task → clickable chips |
| File path | file | index.json task → mono display |
| Record path | record | index.json task → mono display |
| Acceptance criteria | Parsed from .md body | Task markdown file |
| Execution record | Full markdown render | Record markdown file |

---

## Component 4: Activity Sidebar

### Placement

- **Mode**: existing-page
- **Target**: `/projects/:id`
- **Position**: Right edge of page, collapsible. When expanded: w-80 (320px). When collapsed: w-12 (48px). Adjacent to swimlane, outside UF-3 panel. When UF-3 panel is open, sidebar collapses automatically.

### Layout Structure

**Expanded state:**

```
┌──────────────────┐
│ Activity    [«]  │  ← Header: "Activity" + collapse button
│──────────────────│
│ ● 14:30          │  ← Event item
│ 1.3-api-handler  │
│ completed        │  ← Green text
│ feature-auth     │  ← Muted text
│──────────────────│
│ ● 14:15          │
│ 4.1-deploy-conf  │
│ blocked          │  ← Red text
│ feature-infra    │
│──────────────────│
│ ...              │
│  (scrollable)    │
│                  │
└──────────────────┘
```

**Collapsed state:**

```
┌────┐
│ [»]│  ← Expand button
│    │
│ 3  │  ← Blocked count badge (red, centered)
│    │
└────┘
```

**Expanded sidebar**: Fixed height (100vh - header height), border-l 1px var(--border), bg-background. Flex column.

**Header**: h-10, border-b, flex, items-center, px-4. "Activity" (text-sm, font-medium) + chevron-right icon button (ghost, ml-auto).

**Event list**: Overflow-y-auto, flex-1. Each event:
- **Timestamp**: text-xs, muted-foreground, mb-1
- **Status dot**: 8px circle matching event type color (same as task status colors), inline with task ID
- **Task ID + Title**: mono text-xs (ID) + body text-sm (title, truncated 40 chars), single line
- **Event type text**: text-xs, colored per event type (completed=green, blocked=red, claimed=blue, skipped=yellow)
- **Feature slug**: text-xs, muted-foreground
- Padding: px-4 py-3, border-b
- **Hover**: bg-muted/50, cursor pointer
- **Click**: Scroll swimlane to task, highlight with ring animation

**Separator**: Between events, 1px border-b var(--border).

**Collapsed sidebar**: w-12, border-l, flex column, items-center, pt-4, gap-4.
- Expand button: chevron-left icon (ghost, w-8 h-8)
- Blocked badge: red circle w-8 h-8, text-sm font-bold, centered number. Only shown if blocked > 0.

### States

| State | Visual | Behavior |
|-------|--------|----------|
| Expanded | Full event list with scrollbar | Default state on swimlane page. Transition from Collapsed: width animates from 48px to 320px over 200ms ease-out. |
| Collapsed | Thin bar + blocked count badge | User collapsed or UF-3 panel opened. Transition from Expanded: width animates from 320px to 48px over 200ms ease-in. |
| Empty | "No activity yet" centered text with activity icon (muted) | Transition from Loading: skeletons fade out (150ms opacity), empty state fades in (150ms opacity). Zero tasks with status changes. |
| Loading | 10 skeleton event rows (pulse animation) | Shown on initial page load. Transition to Empty or Populated via 150ms crossfade. |

### Interactions

| Trigger | Action | Feedback |
|---------|--------|----------|
| Click collapse button | Collapse sidebar to thin bar | 200ms width transition |
| Click expand button | Expand sidebar to full width | 200ms width transition |
| Click event | Scroll swimlane to task, highlight | Ring animation: persists 2000ms then fades 300ms. If target task is in a collapsed row, row expands first. |
| UF-3 panel opens | Auto-collapse sidebar | 200ms transition |
| UF-3 panel closes | Restore sidebar to previous state | 200ms transition |

### Data Binding

| UI Element | Data Field | Source |
|------------|-----------|--------|
| Event timestamp | mtime | index.json file mtime, formatted as relative time ("2h ago") or HH:mm |
| Task ID | task.id | index.json task |
| Task title (truncated) | task.title | index.json task, truncated 40 chars |
| Feature slug | feature | index.json |
| Event type color | Derived from status | claimed=blue, completed=green, blocked=red, skipped=yellow |
| Blocked count | Count of blocked tasks | Derived across all features |
| Max events | 50 | Sorted by timestamp descending |

---

## Edge Case Handling

### Slow Network / Prolonged Loading

If any Loading state persists beyond 10 seconds, an inline message appears below the skeleton placeholders: "Taking longer than expected" (text-sm, muted-foreground) with a manual "Retry" button (outline variant). Clicking retry re-triggers the data fetch. The skeleton animation continues during the retry. If the second attempt also exceeds 10 seconds, the message changes to "Unable to load data" (destructive variant) with the retry button.

### Stale Data

The dashboard does not poll for file changes. Data is fetched on initial page load and on explicit user action only. Each component header (UF-1, UF-2) includes a refresh icon button (outline, ghost variant, `aria-label="Refresh data"`) in the header bar. Clicking it re-fetches all data for the current page and replaces the displayed content with a Loading -> Populated/Empty/Error transition. There is no "last refreshed" timestamp -- the refresh button is sufficient for the manual-refresh-only model defined by the PRD. No optimistic updates or delta merging; each refresh is a full data replacement.

### Concurrent / Rapid Interactions

| Interaction | Guard |
|-------------|-------|
| Rapid task card clicks | Debounce: second click within 300ms of first is ignored. The detail panel opens for the first card only. |
| Rapid filter changes | Filter dropdowns use standard browser event handling -- the last selected state wins. No queuing. |
| Detail panel open during panel close animation | If the user clicks a different task card while the panel is still closing (200ms), the close animation is interrupted, and the panel reopens with the new task data. No double-panel state. |
| Multiple activity events arriving simultaneously | Event list renders in timestamp-descending order. No animation per event -- the list refreshes atomically. |

---

## Accessibility

### ARIA Labels for Icon-Only Buttons

All icon-only buttons require `aria-label` attributes:

| Button | Location | `aria-label` |
|--------|----------|--------------|
| Dark mode toggle | UF-1 header | "Toggle dark mode" |
| Close (X) button | UF-3 panel header | "Close task details" |
| Collapse sidebar | UF-4 expanded header | "Collapse activity sidebar" |
| Expand sidebar | UF-4 collapsed bar | "Expand activity sidebar" |
| Row chevron (down) | UF-2 expanded feature row | "Collapse feature {slug}" |
| Row chevron (right) | UF-2 collapsed feature row | "Expand feature {slug}" |

### Screen Reader Text for Status Indicators

Status dots convey meaning through color alone. Each dot must include an `sr-only` span:

```html
<span class="inline-block w-2 h-2 rounded-full" style="background: var(--status-color)" />
<span class="sr-only">{status label}</span>
```

Applies to: task card status dots (UF-2), event list status dots (UF-4), health badge dots (UF-1).

### Focus Management

**UF-3 slide-over panel (focus trap)**:
- On open: focus moves to the panel container element (`tabindex="-1"`, `role="dialog"`, `aria-modal="true"`, `aria-label="Task details for {id}"`). Focus trap cycles through focusable children (close button, dependency links, scrollable content) in DOM order. Tab and Shift+Tab wrap within the panel.
- On close (X button, overlay click, or Escape): focus returns to the task card that triggered the panel open. If closed via dependency link, focus moves to the referenced task card after scroll.

**UF-2 swimlane keyboard navigation**:
- Tab order: header bar controls (breadcrumb links, filter dropdowns) -> first task card in first expanded feature row -> remaining task cards in reading order (left-to-right within phase columns, top-to-bottom across features) -> activity sidebar controls.
- Arrow keys not used for card navigation; standard Tab/Shift+Tab suffices.

### Keyboard Shortcuts

| Key | Context | Action |
|-----|---------|--------|
| Escape | UF-3 panel open | Close panel, return focus to originating task card |
| Enter / Space | Task card focused | Open UF-3 panel |
| Enter / Space | Feature row chevron focused | Toggle collapse/expand |

### Reduced Motion

All animations respect `prefers-reduced-motion: reduce`:

| Animation | Default | Reduced Motion |
|-----------|---------|----------------|
| Skeleton pulse | 2s infinite pulse | Static bg-muted, no animation |
| UF-3 slide-in/out | translateX 200ms ease-out/in | Instant appear/disappear (opacity 0 to 1, 0ms) |
| Task card hover border | 150ms border-color transition | Instant color change (0ms) |
| Blocked status dot pulse | 2s infinite pulse | Static dot, no pulse |
| Filter show/hide | 150ms opacity transition | Instant display swap |
| Row collapse/expand | 200ms height animation | Instant height change |
| Activity sidebar width | 200ms width transition | Instant width change |

Implementation: wrap all transition/animation declarations in `@media (prefers-reduced-motion: no-preference) { ... }` and provide instant fallbacks outside the block.

---

## Page Composition

| Page | Route | Components | Layout Notes |
|------|-------|-----------|--------------|
| Landing | `/` | UF-1 only | Full-width card grid, centered container |
| Swimlane | `/projects/:id` | UF-2 + UF-3 + UF-4 | UF-2: flex-1 main area; UF-3: overlay panel (40vw); UF-4: right sidebar (w-80 / w-12). UF-3 takes priority over UF-4 when open. |

### Swimlane Page Layout Priority

1. **Default**: Swimlane (flex-1) + Activity sidebar expanded (w-80)
2. **Sidebar collapsed**: Swimlane (flex-1) + Activity sidebar collapsed (w-12)
3. **Detail panel open**: Swimlane (shrinks) + Activity sidebar hidden + Detail panel (40vw overlay from right)
