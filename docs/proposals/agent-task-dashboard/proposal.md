---
created: 2026-05-06
author: "panda"
status: Draft
---

# Proposal: Agent Task Dashboard

## Problem

Forge-generated tasks are scattered across `docs/features/<slug>/tasks/index.json` files within each project. With 24 features and ~330 tasks in a single project (pm-work-tracker), there is no way to see the overall progress, identify blockers, or understand cross-feature dependencies without manually navigating the filesystem.

### Evidence

- **pm-work-tracker** has 24 features, ~330 tasks across 22 feature directories. The largest feature (`pm-work-tracker`) alone has 60 tasks.
- Each feature's task data lives in an isolated `index.json` — no aggregated view exists.
- Task dependencies can span features (e.g., `rbac-permissions` tasks depend on `code-quality` outputs), but these cross-feature links are invisible without reading individual files.
- `task-cli` provides per-feature operations (`task claim`, `task status`) but no cross-feature or cross-project visibility.

### Urgency

As the number of forge-managed projects grows (pm-work-tracker, forge itself, and future projects), the cognitive overhead of tracking agent work across projects becomes the primary bottleneck. Manual status aggregation across features currently takes ~10-15 minutes per check (grepping index.json files, running `task status` per feature, mentally cross-referencing dependency chains) and is performed 3-4 times daily — consuming roughly 45 minutes/day of operator time. More critically, cross-feature blockers have gone undetected: in the last sprint, a blocked task in `rbac-permissions` sat idle for 2 days because its dependency on a `code-quality` task was invisible without reading individual feature directories. Without a dashboard, every cross-project question ("what's blocked?", "which features are behind?") requires a manual scan that does not scale.

## Proposed Solution

A read-only web dashboard (separate Go binary) that:

1. **Reads task data directly from filesystem** — scans `docs/features/<slug>/tasks/index.json` across configured project directories. No database, no sync lag.
2. **Shows projects as cards** on a landing page with summary stats (feature count, task completion %, last update time).
3. **Renders a feature-per-row swimlane** within each project: rows = features, columns = phases, task cards color-coded by status, dependency arrows drawn between tasks (including cross-feature dependencies).
4. **Displays task details in a slide-over panel** — metadata, acceptance criteria, execution record, dependency chain.
5. **Provides a collapsible activity sidebar** showing recent events (claimed, completed, blocked) across all features in the current project.
6. **Exposes a query-only REST API** for agents to programmatically retrieve project, feature, task, and dependency data.

Project list is configured via a YAML/JSON config file. Dashboard runs as a local web server, data refreshes on manual reload.

## Alternatives Considered

| Approach | Pros | Cons | Verdict |
|----------|------|------|---------|
| Do nothing | Zero development cost | Problem compounds with each new project; operator overhead grows linearly | Rejected: the pain is already real at 2 projects |
| Extend task-cli with dashboard subcommand | Single binary, no extra install | task-cli is a focused CLI tool; embedding a web server bloats its scope and binary size | Rejected: separate concern deserves separate tool |
| Database-backed dashboard | Richer queries, history tracking, trend analysis | Requires sync mechanism, introduces data staleness, adds operational complexity for what is fundamentally a read-only view of file-based data | Rejected: premature; filesystem reads are sufficient and always in sync |
| Activity-centric design (Kanban by status) | Good for daily standup-style monitoring | Columns show only status — phase assignments and dependency chains are invisible. Filtering or collapsing Kanban columns does not recover this lost spatial context, because the layout itself discards it. Swimlane rows encode phase and feature grouping natively; filtering/collapsing reduces visual density while preserving the relationship structure that Kanban erases | Rejected: swimlane preserves both progress and dependency context; mitigations (filtering, collapsing) work precisely because the underlying layout retains phase/dependency relationships |

## Scope

### In Scope

1. **Project dashboard page** — landing page with project cards showing feature count, task completion %, and last update timestamp
2. **Feature swimlane view** — per-project: rows = features, columns = phases (1, 2, 3, Testing), task cards with status color-coding and dependency arrows (SVG/Canvas-based DAG)
3. **Task detail slide-over** — click a task card to open a side panel showing: ID, title, priority, status, scope, dependencies, acceptance criteria, execution record content, links to PRD/design docs
4. **Activity sidebar** — collapsible panel on the swimlane page showing recent events across all features in the current project (task claimed, completed, blocked, skipped) with timestamps
5. **Agent query API** — RESTful endpoints: `GET /api/projects`, `GET /api/projects/:id/features`, `GET /api/projects/:id/features/:slug/tasks`, `GET /api/projects/:id/features/:slug/dependencies`
6. **Project configuration** — YAML/JSON config file listing project directories (name + filesystem path)

### Out of Scope

- Write operations (claim, record, status update) — all mutations remain through task-cli
- Authentication/authorization — localhost-only, single-user context
- Real-time updates (WebSockets, SSE) — manual page refresh
- Cross-project dependency tracking
- Mobile/responsive layout
- CI/CD integration or external notifications (Slack, email)
- Data persistence or historical trend tracking

## Key Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Swimlane DAG becomes unreadable with 24+ features and 300+ tasks | High | High | Implement feature filtering (by status, by active focus) and collapsible feature rows; render dependency edges on-demand for visible tasks only |
| Filesystem reads are slow across multiple large projects on network drives | Medium | Medium | Cache parsed index.json in memory with manual invalidation on refresh; lazy-load per-project data |
| Frontend DAG rendering performance degrades with many edges | Medium | Medium | Use Canvas or WebGL for edge rendering instead of SVG; limit visible dependency edges to selected task's direct dependencies by default |
| Config file becomes stale as projects are added/removed | Low | Low | Auto-detect feature directories within configured paths; show "no features found" warning for misconfigured entries |
| Agent API design doesn't match actual agent query patterns | Medium | Low | Start with read-only endpoints matching the UI data model; iterate based on agent usage patterns |

## Success Criteria

- [ ] Dashboard displays all configured projects on landing page with accurate task completion counts within 3 seconds of page load
- [ ] Swimlane view renders 24 features (330 tasks) with dependency arrows in < 1s; all task cards render within their assigned swimlane row with no overlap at 1920x1080 viewport; no layout shift > 5px during initial render
- [ ] Task detail slide-over displays the following fields from index.json: task ID, title, priority, status, scope, dependency list, acceptance criteria text, and execution record content (if the `executionRecord` key exists in the task's index.json entry; if absent, the panel shows "No execution record" explicitly)
- [ ] Activity sidebar displays the 50 most recent task status changes with timestamps
- [ ] Agent API returns project/feature/task data in < 200ms for a single project
- [ ] Configuration supports adding a new project by specifying name + path with zero code changes

## Next Steps

- Proceed to `/write-prd` to formalize requirements
