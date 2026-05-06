---
feature: "agent-task-dashboard"
generated: "2026-05-07"
status: draft
---

# Technical Specifications: Agent Task Dashboard

## API Conventions

### TECH-001: JSON Response Format

**Requirement**: All API responses return JSON with consistent envelope. Each response includes a `meta` object with `lastUpdated` timestamp.
**Scope**: [LOCAL]
**Source**: API Handbook (Data Contracts > Meta Object)

```json
{
  "meta": {
    "lastUpdated": "2026-05-06T14:30:00Z",
    "refreshAvailable": true
  }
}
```

### TECH-002: Error Response Format

**Requirement**: All errors return consistent JSON format with `error` (machine-readable code) and `message` (human-readable description) fields.
**Scope**: [LOCAL]
**Source**: API Handbook (Error Response), Tech Design (Error Handling)

Error codes: `invalid_slug` (400), `invalid_task_id` (400), `not_found` (404), `internal_error` (500).

### TECH-003: Input Validation Patterns

**Requirement**: API input validation rules:
- `:id` must match a configured project name (exact match, no filesystem interpretation).
- `:slug` must match `^[a-zA-Z0-9-]+$` regex.
- `:taskId` must not contain path traversal sequences (e.g., `../`).
**Scope**: [LOCAL]
**Source**: Tech Design (Propagation Strategy)

Validation is performed in Gin middleware before reaching handlers.

## Filesystem Architecture

### TECH-004: Filesystem Abstraction via fs.FS

**Requirement**: The Scanner component uses `io/fs.FS` interface for filesystem access. Production code uses `os.DirFS`; tests inject `fstest.MapFS` or in-memory filesystems.
**Scope**: [LOCAL]
**Source**: Tech Design (Interface 2: Scanner)

This abstraction enables deterministic unit tests without real filesystem access.

### TECH-005: In-Memory Cache Strategy

**Requirement**: Parsed index.json data is cached in-memory per project as a Go map (`map[string]*ProjectData`). Cache is invalidated only on explicit `Invalidate()` call (triggered by page refresh).
**Scope**: [LOCAL]
**Source**: PRD (Data Requirements > caching), Tech Design (Interface 2: Scanner)

No background cache refresh; no TTL-based expiry. Manual refresh triggers full re-scan.

### TECH-006: Static Asset Embedding

**Requirement**: All frontend assets (HTML templates, JS, CSS, dagre library) are embedded into the Go binary using `//go:embed` directives. Zero external file dependencies at runtime.
**Scope**: [LOCAL]
**Source**: Tech Design (Component Diagram), Tech Design (Dependencies)

The binary is self-contained: deploy a single file, no web/ directory needed.

## Security

### TECH-007: Localhost-Only Binding

**Requirement**: The HTTP server binds to `127.0.0.1` only (not `0.0.0.0`). No external network exposure. No authentication required.
**Scope**: [LOCAL]
**Source**: PRD (Security Requirements), Tech Design (Security Considerations)

### TECH-008: XSS Prevention via Template Auto-Escaping

**Requirement**: All user-controlled content (task titles, descriptions from markdown) is rendered through Go `html/template` auto-escaping. No raw HTML injection from task content.
**Scope**: [LOCAL]
**Source**: Tech Design (Security Considerations > Mitigations)

### TECH-009: Path Traversal Prevention

**Requirement**: URL path parameters are validated against exact-match allowlists (project names from config) or strict regex patterns (slug, task ID). No filesystem path interpretation from URL segments.
**Scope**: [LOCAL]
**Source**: Tech Design (Security Considerations > Mitigations)

## Performance

### TECH-010: Performance Targets

**Requirement**: Response time targets:
- Landing page load: < 3 seconds for 5 projects with 30 features each.
- Swimlane render: < 1 second for 24 features (330 tasks).
- Task detail panel: < 200ms.
- API response: < 200ms for single project query.
**Scope**: [LOCAL]
**Source**: PRD (Performance Requirements)

Minimum supported viewport: 1920x1080.

## Data Parsing

### TECH-011: Task Markdown Parsing

**Requirement**: Task `.md` files are parsed to extract:
- Acceptance criteria from `## Acceptance Criteria` sections.
- Scope from frontmatter or body.
- Description as body text excluding acceptance criteria.

Record `.md` files are parsed to extract structured sections (`## Summary`, `## Files`, `## Decisions`, `## Test Results`) identified by `##` headings. Full markdown retained as `Raw` fallback.
**Scope**: [LOCAL]
**Source**: Tech Design (Interface 5: Task Markdown Parser)

## Dependency Graph

### TECH-012: Dependency Graph Representation

**Requirement**: Dependencies are represented as a graph with nodes (tasks) and directed edges (dependencies). Edges include a `crossFeature` boolean indicating whether the dependency spans features. Wildcard dependencies are expanded to individual edges.
**Scope**: [LOCAL]
**Source**: API Handbook (Get Dependencies), Tech Design (Model 5: DependencyGraph)
