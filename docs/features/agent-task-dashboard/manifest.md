---
feature: "agent-task-dashboard"
status: tasks
---

# Feature: agent-task-dashboard

<!-- Status flow: prd → design → tasks → in-progress → completed -->

## Documents

| Document | Path | Summary |
|----------|------|---------|
| PRD Spec | prd/prd-spec.md | Read-only web dashboard for monitoring forge-generated tasks across projects. Project cards, feature swimlanes with dependency DAG, task detail panel, activity sidebar, and agent query API. |
| User Stories | prd/prd-user-stories.md | 7 stories: project overview, blocker diagnosis, cross-feature deps, execution records, activity monitoring, agent API queries, and project configuration. |
| UI Functions | prd/prd-ui-functions.md | 4 UI functions: project card grid (landing), feature swimlane board, task detail slide-over panel, activity sidebar. 2 pages: `/` and `/projects/:id`. |
| Tech Design | design/tech-design.md | Go + Gin binary with Go templates + vanilla JS. dagre + SVG for DAG rendering. Filesystem reader with in-memory cache. 6 interfaces, 6 data models, 7 API endpoints. |
| API Handbook | design/api-handbook.md | 7 query-only REST endpoints for projects, features, tasks, and dependency graphs. JSON responses with consistent error format. |
| UI Design | ui/ui-design.md | Shadcn-style design for 4 components: project card grid, feature swimlane board, task detail slide-over panel, activity sidebar. Status color system, dependency arrow rendering, responsive layout. |

## Traceability

| PRD Section | Design Section | UI Component | Placement | Tasks |
|-------------|----------------|--------------|-----------|-------|
| Func Specs > 5.6 Project Config | Tech Design > Interface 1: Config | — | — | 1.1, 1.3 |
| Func Specs > 5.1 Landing Page (data) | Tech Design > Data Models 1-8 | — | — | 1.2, 1.4 |
| Func Specs > 5.2 Swimlane (sort) | Tech Design > Scanner.SortFeatures | — | — | 1.4 |
| Func Specs > 5.5 Agent API (errors) | Tech Design > Error Handling | — | — | 1.5 |
| UI Functions > UF-1 Project Card Grid | UI Design > Component 1 | Project Card Grid | new-page:`/` | 3.2 |
| UI Functions > UF-2 Feature Swimlane Board | UI Design > Component 2 | Feature Swimlane Board | new-page:`/projects/:id` | 3.3 |
| UI Functions > UF-3 Task Detail Slide-Over Panel | UI Design > Component 3 | Task Detail Slide-Over Panel | existing-page:`/projects/:id` | 3.4 |
| UI Functions > UF-4 Activity Sidebar | UI Design > Component 4 | Activity Sidebar | existing-page:`/projects/:id` | 3.5 |
| Func Specs > 5.3 Task Detail Panel (parsing) | Tech Design > Interface 5: Task Markdown Parser | — | — | 2.1 |
| Func Specs > 5.5 Agent API (endpoints) | Tech Design > Interface 4: API Handlers | — | — | 2.2 |
| Func Specs > 5.1 Landing Page (template) | Tech Design > Interface 3: Page Handlers | — | new-page:`/` | 2.3 |
| Func Specs > 5.1 Landing Page (serve) | Tech Design > Component Diagram | — | — | 2.4 |
| CSS + Status Colors | UI Design > Design Tokens | — | — | 3.1 |
