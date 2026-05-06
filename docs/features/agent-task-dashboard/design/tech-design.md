---
created: 2026-05-06
prd: prd/prd-spec.md
status: Draft
---

# Technical Design: Agent Task Dashboard

## Overview

A single Go binary serving a web dashboard for monitoring forge-generated tasks across projects. The binary reads `index.json` files directly from configured project directories, serves HTML pages via Go templates, provides a JSON API for agents, and renders dependency DAGs client-side using dagre + SVG.

**Key technology choices:**
- **Backend**: Go + Gin HTTP framework
- **Frontend**: Go html/template + vanilla JS, static assets embedded via `go:embed`
- **DAG rendering**: dagre (graph layout) + SVG (rendering), embedded as static JS
- **Config**: YAML config file at `~/.task-dashboard.yaml` or CLI flag
- **Deployment**: Single binary, zero external dependencies

## Architecture

### Layer Placement

```
┌─────────────────────────────────────────┐
│          Browser (Client Layer)          │
│  ┌──────────┐  ┌──────────┐  ┌───────┐  │
│  │ Go tmpl  │  │  dagre   │  │  CSS  │  │
│  │ HTML     │  │ + SVG    │  │       │  │
│  └──────────┘  └──────────┘  └───────┘  │
└────────────────┬────────────────────────┘
                 │ HTTP (JSON API + HTML)
┌────────────────▼────────────────────────┐
│          Go Server (Gin)                │
│  ┌──────────┐  ┌──────────────────┐     │
│  │ HTML     │  │  JSON API        │     │
│  │ handlers │  │  handlers        │     │
│  └──────────┘  └──────────────────┘     │
│  ┌──────────────────────────────────┐   │
│  │  Scanner (filesystem reader)     │   │
│  └──────────────────────────────────┘   │
│  ┌──────────────────────────────────┐   │
│  │  Config (YAML loader)            │   │
│  └──────────────────────────────────┘   │
└────────────────┬────────────────────────┘
                 │ fs.FS interface (os.DirFS in prod)
┌────────────────▼────────────────────────┐
│          Filesystem                      │
│  ~/.task-dashboard.yaml                  │
│  project/docs/features/*/tasks/index.json│
│  project/docs/features/*/tasks/*.md      │
│  project/docs/features/*/tasks/records/  │
└─────────────────────────────────────────┘
```

### Component Diagram

```
┌─────────────────────────────────────────────────┐
│  cmd/task-dashboard/main.go   (entry point)     │
├─────────────────────────────────────────────────┤
│  internal/                                       │
│  ├── config/       Config — YAML loader         │
│  ├── scanner/      Scanner — filesystem reader   │
│  ├── handler/      Handler — HTTP handlers       │
│   │   ├── page.go      HTML page handlers        │
│   │   └── api.go       JSON API handlers         │
│  └── model/        Model — data structures       │
├─────────────────────────────────────────────────┤
│  web/                                            │
│  ├── templates/    Go HTML templates             │
│  ├── static/       JS, CSS, images               │
│  │   ├── js/                                        │
│  │   │   ├── dagre.min.js     (embedded)           │
│  │   │   ├── swimlane.js      (swimlane renderer)  │
│  │   │   ├── detail-panel.js  (slide-over logic)    │
│  │   │   └── activity.js      (sidebar logic)       │
│  │   └── css/                                       │
│  │       └── styles.css                             │
│  └── embed.go      //go:embed directive          │
└─────────────────────────────────────────────────┘
```

### Dependencies

| Dependency | Type | Purpose | Version |
|---|---|---|---|
| gin-gonic/gin | External | HTTP framework with routing, middleware, JSON binding | v1.10+ |
| gopkg.in/yaml.v3 | External | YAML config file parsing | v3 |
| Go stdlib embed | Stdlib | Embed static assets into binary | Go 1.22+ |
| Go stdlib io/fs | Stdlib | Filesystem abstraction for testability | Go 1.22+ |
| Go stdlib html/template | Stdlib | Server-side HTML rendering | Go 1.22+ |
| dagre | JS (embedded) | Graph layout algorithm for dependency DAG | Latest |

## Interfaces

### Interface 1: Config

```go
// Load reads the YAML config file and returns parsed config.
// Returns error if file not found or invalid YAML.
func Load(path string) (*Config, error)

// Config holds project list and server settings.
type Config struct {
    Projects []ProjectConfig `yaml:"projects"`
    Server   ServerConfig    `yaml:"server,omitempty"`
}

type ProjectConfig struct {
    Name string `yaml:"name"` // Display name, also used as URL ID (lowercased)
    Path string `yaml:"path"` // Absolute path to project root
}

type ServerConfig struct {
    Port int `yaml:"port,omitempty"` // Default: 8080
}
```

### Interface 2: Scanner

```go
// Scanner reads filesystem to build project, feature, and task data.
// Uses io/fs.FS interface for testability — production uses os.DirFS,
// unit tests inject fstest.MapFS or memfs.
type Scanner struct {
    config *Config
    fs     fs.FS                    // Abstracted filesystem
    cache  map[string]*ProjectData // keyed by project name
}

// NewScanner creates a scanner for the given config.
func NewScanner(cfg *Config) *Scanner

// ScanAll reads all configured projects. Returns map keyed by project ID.
// Skips projects with invalid paths (logs warning).
func (s *Scanner) ScanAll() (map[string]*ProjectData, error)

// ScanProject reads a single project by ID.
// Returns ErrProjectNotFound if ID not in config.
func (s *Scanner) ScanProject(id string) (*ProjectData, error)

// Invalidate clears the in-memory cache, forcing fresh filesystem reads.
func (s *Scanner) Invalidate()

// SortFeatures sorts features in-place by PRD rule:
//   1. Features with any blocked tasks come first
//   2. Within each group, sort by completion % ascending (most incomplete first)
//   3. Ties broken by slug alphabetically
func SortFeatures(features []FeatureData)
```

### Interface 3: Page Handlers

```go
// RegisterPages registers HTML page routes on the Gin engine.
func RegisterPages(r *gin.Engine, scanner *Scanner)

// GET / — landing page with project cards
func handleLanding(c *gin.Context)

// GET /projects/:id — swimlane view for a project
func handleProject(c *gin.Context)
```

### Interface 4: API Handlers

```go
// RegisterAPI registers JSON API routes under /api on the Gin engine.
func RegisterAPI(r *gin.Engine, scanner *Scanner)

// GET /api/projects — list all projects with summaries
// GET /api/projects/:id — single project with feature list
// GET /api/projects/:id/features — list features
// GET /api/projects/:id/features/:slug — single feature with tasks
// GET /api/projects/:id/features/:slug/tasks — list tasks
// GET /api/projects/:id/features/:slug/tasks/:taskId — single task
// GET /api/projects/:id/features/:slug/dependencies — dependency graph
```

### Interface 5: Task Markdown Parser

```go
// ParseTaskFile reads a task .md file and extracts acceptance criteria.
func ParseTaskFile(filePath string) (*TaskFileContent, error)

// ParseRecordFile reads a record .md file and extracts structured sections.
// Sections are identified by ## headings in the markdown.
func ParseRecordFile(filePath string) (*RecordContent, error)

// RecordContent holds the structured sections parsed from an execution record.
// PRD User Story 4 requires: summary, files created/modified, key decisions, test results.
type RecordContent struct {
    Summary     string   // Content under ## Summary heading
    Files       []string // Content under ## Files heading, split by line
    Decisions   string   // Content under ## Decisions heading
    TestResults string   // Content under ## Test Results heading
    Raw         string   // Full markdown content for fallback rendering
}

type TaskFileContent struct {
    AcceptanceCriteria []string // Extracted from ## Acceptance Criteria section
    Scope              string   // Extracted from frontmatter or body
    Description        string   // Body text excluding acceptance criteria
}
```

### Interface 6: Detail Panel Navigation (JS)

```js
// detail-panel.js — Dependency click-to-navigate

// Each dependency entry in the slide-over panel renders as a clickable link.
// On click, navigateToTask closes the current panel and scrolls the swimlane
// to the upstream task card, highlighting it briefly.

/**
 * @param {string} taskId - The upstream task ID (e.g., "1.1")
 * @param {string} featureSlug - The feature slug the task belongs to
 */
function navigateToTask(taskId, featureSlug) {
  closePanel();                                    // Close slide-over
  const card = document.getElementById(`task-${taskId}`);
  if (card) {
    card.scrollIntoView({ behavior: 'smooth', block: 'center' });
    card.classList.add('highlight-flash');          // CSS animation (1s fade)
  } else if (featureSlug !== currentFeatureSlug) {
    // Cross-feature: navigate to that feature's swimlane view
    window.location.href = `/projects/${currentProjectId}?feature=${featureSlug}&task=${taskId}`;
  }
}

// Each dependency list item in the panel template:
// <a onclick="navigateToTask('${dep.taskId}', '${dep.featureSlug}')">${dep.taskId}</a>
```

## Data Models

### Model 1: ProjectData

```go
type ProjectData struct {
    ID            string         // Lowercased name from config
    Name          string         // Display name from config
    Path          string         // Filesystem path
    Features      []FeatureData  // All features with valid index.json
    LastUpdated   time.Time      // Max mtime of all index.json files
    HealthStatus  string         // "active" | "complete" | "stale"
    TotalTasks    int            // Sum across all features
    CompletedTasks int           // Sum of completed tasks
    Warnings      []string       // Config validation warnings
}
```

### Model 2: FeatureData

```go
type FeatureData struct {
    Slug            string            // From index.json "feature" field
    Status          string            // "planning" | "in-progress" | "completed"
    PRDPath         string            // From index.json "prd" field
    DesignPath      string            // From index.json "design" field
    Tasks           map[string]Task   // Keyed by task key (e.g., "1.1-interfaces")
    Phases          []PhaseInfo       // Derived from task IDs
    LastUpdated     time.Time         // Mtime of index.json
    TotalTasks      int
    CompletedTasks  int
    HasBlockedTasks bool              // Derived: true if any task has status "blocked"
    CompletionPct   float64           // Derived: (CompletedTasks / TotalTasks) * 100; 0 if TotalTasks = 0
}
```

### Model 3: Task

```go
type Task struct {
    ID            string   // e.g., "1.1", "T-test-1"
    Key           string   // e.g., "1.1-interfaces"
    Title         string
    Priority      string   // "P0" | "P1" | "P2"
    Status        string   // "pending" | "in_progress" | "completed" | "blocked" | "skipped"
    Scope         string   // "frontend" | "backend" | "all"
    EstimatedTime string   // Optional
    Dependencies  []string // Task IDs, supports wildcards "1.x"
    Breaking      bool
    File          string   // Relative path to task .md
    Record        string   // Relative path to execution record .md
    Phase         int      // Derived: first number from ID (1, 2, 3...)
}
```

### Model 4: PhaseInfo

```go
type PhaseInfo struct {
    Number   int      // Phase number (1, 2, 3...)
    Label    string   // Display label ("Phase 1", "Phase 2", "Testing")
    TaskKeys []string // Task keys belonging to this phase
}
```

### Model 5: DependencyGraph

```go
type DependencyGraph struct {
    Nodes []GraphNode `json:"nodes"`
    Edges []GraphEdge `json:"edges"`
}

type GraphNode struct {
    ID       string `json:"id"`       // Task ID
    Key      string `json:"key"`      // Task key
    Title    string `json:"title"`
    Status   string `json:"status"`
    Phase    int    `json:"phase"`
    Feature  string `json:"feature"`  // Feature slug
}

type GraphEdge struct {
    Source string `json:"source"` // Dependent task ID
    Target string `json:"target"` // Dependency task ID
    CrossFeature bool `json:"crossFeature"` // True if source and target are in different features
}
```

### Model 6: ActivityEvent

```go
type ActivityEvent struct {
    Timestamp time.Time `json:"timestamp"`
    TaskID    string    `json:"taskId"`
    TaskTitle string    `json:"taskTitle"`
    Feature   string    `json:"feature"`
    EventType string    `json:"eventType"` // "claimed" | "completed" | "blocked" | "skipped"
}
```

## Error Handling

### Error Types & Codes

| Error Code | Name | Description | HTTP Status |
|---|---|---|---|
| ERR_INVALID_SLUG | InvalidSlugError | Feature slug contains characters other than alphanumeric and hyphens | 400 |
| ERR_INVALID_TASK_ID | InvalidTaskIDError | Task ID format is malformed (e.g., contains path traversal) | 400 |
| ERR_PROJECT_NOT_FOUND | ProjectNotFoundError | Project ID not in config or path invalid | 404 |
| ERR_FEATURE_NOT_FOUND | FeatureNotFoundError | Feature slug not found in project | 404 |
| ERR_TASK_NOT_FOUND | TaskNotFoundError | Task ID not found in feature | 404 |
| ERR_CONFIG_INVALID | ConfigInvalidError | YAML parse error or missing required fields | 500 |
| ERR_FS_READ | FilesystemReadError | Cannot read index.json or task files | 500 |
| ERR_PARSE_INDEX | IndexParseError | index.json is not valid JSON or missing required fields | 500 |

### Propagation Strategy

```
Validation middleware (Gin middleware):
                Validates :id matches a configured project name (exact match).
                Validates :slug matches ^[a-zA-Z0-9-]+$ regex.
                Validates :taskId does not contain path traversal sequences.
                On failure → 400 JSON {"error": "ERR_INVALID_SLUG", "message": "..."}
Scanner layer:  Returns typed errors (ErrProjectNotFound, ErrFSRead, ErrParseIndex)
Handler layer:  Catches scanner errors, maps to HTTP status codes
                - ErrProjectNotFound / ErrFeatureNotFound / ErrTaskNotFound → 404 JSON
                - ErrInvalidSlug / ErrInvalidTaskID → 400 JSON (caught by middleware before handler)
                - ErrFSRead / ErrParseIndex → 500 JSON
                - Unknown errors → 500 JSON with generic message
Template layer: If data fetch fails, render error page with status code and message
API layer:      Always returns JSON: {"error": "code", "message": "description"}
```

## Cross-Layer Data Map

| Field Name | Filesystem (index.json) | Backend Model (Go) | API JSON | Frontend (JS/Template) |
|---|---|---|---|---|
| task.id | `string` | `Task.ID string` | `json:"id"` | `task.id` in template, `data.id` in JS |
| task.title | `string` | `Task.Title string` | `json:"title"` | `{{.Title}}` in template |
| task.status | `string` (enum) | `Task.Status string` | `json:"status"` | Status CSS class, color mapping |
| task.priority | `string` (P0/P1/P2) | `Task.Priority string` | `json:"priority"` | Priority badge |
| task.dependencies | `[]string` | `Task.Dependencies []string` | `json:"dependencies"` | dagre edge source |
| task.phase | derived from ID | `Task.Phase int` | `json:"phase"` | dagre column placement |
| project completion % | computed | `float64` | `json:"completionPct"` | Progress bar width |
| feature last updated | file mtime | `FeatureData.LastUpdated time.Time` | `json:"lastUpdated"` (ISO 8601) | Relative time display |

## Integration Specs

No existing-page integrations — not applicable. All pages are new.

## Testing Strategy

### Per-Layer Test Plan

| Layer | Test Type | Tool | What to Test | Coverage Target |
|---|---|---|---|---|
| Scanner | Unit | testing + `testing/fstest` + `fs.FS` interface | Scanner accepts `fs.FS` abstraction; unit tests inject `fstest.MapFS` with canned index.json and task files to verify parsing, sort logic, dependency expansion | 80% |
| Scanner | Integration | testing + `os.CreateTemp` temp dirs | Real filesystem reads with test fixtures (sample index.json + task files) | Key paths |
| Handler (API) | Integration | `net/http/httptest` + Gin test mode | All 7 API endpoints: happy path, 404s, 400 validation errors, 500 error responses | 80% |
| Handler (Page) | Integration | `net/http/httptest` | HTML responses contain expected data, 404 for missing projects | Key paths |
| Model | Unit | testing | Phase derivation from task IDs, dependency wildcard expansion, health status computation, sort ordering | 90% |
| JS (swimlane) | Unit | Jest + jsdom | DAG node placement, edge rendering, sort order, dependency click-to-navigate calls | Key paths |

### Key Test Scenarios

1. **Happy path**: Load config with 2 projects → scan → API returns correct data for both
2. **Invalid project path**: Config has a path that doesn't exist → project skipped, warning returned
3. **Malformed index.json**: Feature dir has invalid JSON → feature skipped, error logged
4. **Empty project**: Project has docs/features/ but no index.json files → 0 features, no crash
5. **Wildcard dependencies**: Task depends on "1.x" → all matching tasks resolved as edges
6. **Cross-feature deps**: Task in feature A depends on task in feature B → edge has crossFeature=true
7. **Activity events**: Feature has tasks in various statuses → events derived and sorted by mtime

### Overall Coverage Target

80% for Go code. JS swimlane renderer tested with Jest + jsdom (unit-level); manual browser verification for visual layout correctness only.

## Security Considerations

### Threat Model

| Threat | Risk Level | Notes |
|---|---|---|
| Path traversal via API | Medium | Attacker crafts `:id` or `:slug` with `../` to read arbitrary files |
| XSS via task content | Medium | Task titles/descriptions from markdown rendered in HTML |
| Local network exposure | Low | Dashboard binds to localhost only |

### Mitigations

- **Path traversal**: Validate `:id` matches a configured project name (exact match, no filesystem interpretation); validate `:slug` contains only alphanumeric + hyphens
- **XSS**: Use Go template auto-escaping (default in html/template); sanitize markdown content before rendering
- **Network**: Bind to `127.0.0.1` only (not `0.0.0.0`)

## PRD Coverage Map

| PRD AC | Design Component | Interface / Model |
|---|---|---|
| Landing page shows project cards with stats | Page Handler `handleLanding` + Scanner `ScanAll` | `ProjectData` model, `landing.html` template |
| Each card: name, feature count, completion %, last update | `ProjectData` computed fields | `TotalTasks`, `CompletedTasks`, `LastUpdated`, `HealthStatus` |
| Click project card → swimlane view | Page Handler `handleProject` | Route `/projects/:id`, `swimlane.html` template |
| Feature rows with phase columns | `swimlane.js` + dagre | `FeatureData.Phases`, `Task.Phase` |
| Task cards color-coded by status | CSS status classes | `status-pending`, `status-in-progress`, `status-completed`, `status-blocked`, `status-skipped` |
| Dependency arrows between tasks | dagre + SVG in `swimlane.js` | `DependencyGraph` model, `GraphEdge.CrossFeature` |
| Cross-feature dependencies (dashed) | `GraphEdge.CrossFeature` field | CSS class `edge-cross-feature` for dashed style |
| Features sorted: blocked first, then completion % ascending | `Scanner.SortFeatures` | Sort applied after `ScanAll`/`ScanProject`, before rendering; uses `FeatureData.HasBlockedTasks` and `FeatureData.CompletionPct` |
| Click task → slide-over panel | `detail-panel.js` | API endpoint `GET /api/.../tasks/:taskId` + task file reading |
| Each dependency is clickable → navigate to upstream task | `detail-panel.js` `navigateToTask()` | Each dep renders as `<a onclick="navigateToTask(id, slug)">`; closes panel, scrolls to task card; cross-feature navigates to feature swimlane via URL param |
| Panel shows: ID, title, status, priority, scope, deps, criteria, record | API handler + `ParseTaskFile` + `ParseRecordFile` | `Task` model, `TaskFileContent`, `RecordContent` |
| Panel shows structured record sections: summary, files, decisions, test results | `ParseRecordFile` returns `RecordContent` | Sections parsed from `##` headings; each field rendered in its own panel section |
| No execution record → explicit message | `ParseRecordFile` returns nil | Template condition: if nil, show "No execution record" |
| Activity sidebar: 50 recent events | `ActivityEvent` derivation from Scanner | `activity.js`, events sorted by `Timestamp` desc |
| Events: timestamp, task ID, title, feature, type | `ActivityEvent` model | JSON from API, rendered in sidebar template |
| Click event → scroll to task | `activity.js` click handler | `document.getElementById(taskId).scrollIntoView()` |
| GET /api/projects → project list | API Handler | `RegisterAPI`, returns `[]ProjectData` |
| GET /api/.../tasks → task array | API Handler | Returns `[]Task` with full metadata |
| GET /api/.../dependencies → graph | API Handler | Returns `DependencyGraph` (nodes + edges) |
| API response < 200ms | Scanner cache | In-memory cache, invalidated on refresh |
| Config: YAML file with name + path | `Config` model + `Load` | `~/.task-dashboard.yaml`, validated on startup |
| Invalid path → warning on landing | `Scanner.ScanAll` error handling | `ProjectData.Warnings` field |

## Open Questions

- None — all technology decisions resolved.

## Appendix

### Alternatives Considered

| Approach | Pros | Cons | Why Not Chosen |
|---|---|---|---|
| Go stdlib net/http (no Gin) | Zero external dependencies | More boilerplate for routing, JSON binding, middleware | Gin provides cleaner routing with less code; single dependency is acceptable |
| React SPA | Richer interactivity, component reuse | Requires npm/pnpm build pipeline, node_modules, separate build step | PRD targets single binary; Go templates + vanilla JS keeps deployment simple |
| Cytoscape.js for DAG | Full graph visualization, pan/zoom | ~200KB, overkill for fixed swimlane layout | dagre is ~15KB, purpose-built for layered layouts that match the swimlane structure |
| SQLite cache layer | Richer queries, history tracking | Adds sync complexity, database file management | Filesystem reads are fast enough; cache in Go map with manual invalidation |

### References

- task-cli source: `Z:/project/ai/forge/task-cli/` (Go data model reference)
- task index schema: `Z:/project/ai/forge/plugins/forge/skills/breakdown-tasks/templates/index.schema.json`
- dagre documentation: https://github.com/dagrejs/dagre
- Gin documentation: https://gin-gonic.com/docs/
