---
created: 2026-05-06
related: design/tech-design.md
---

# API Handbook: Agent Task Dashboard

## API Overview

Query-only REST API served by the dashboard binary. All endpoints return JSON. No authentication required (localhost-only).

## Endpoints

### List Projects

**Method**: `GET`
**Path**: `/api/projects`
**Auth**: none

#### Request

No parameters.

#### Response (200)

```json
{
  "projects": [
    {
      "id": "pm-work-tracker",
      "name": "pm-work-tracker",
      "featureCount": 24,
      "completedTasks": 312,
      "totalTasks": 330,
      "completionPct": 94.5,
      "lastUpdated": "2026-05-06T14:30:00Z",
      "healthStatus": "active"
    }
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| id | string | URL-safe project identifier (lowercased name) |
| name | string | Display name from config |
| featureCount | number | Number of features with valid index.json |
| completedTasks | number | Sum of completed tasks across features |
| totalTasks | number | Sum of all tasks across features |
| completionPct | number | (completedTasks / totalTasks) * 100; 0 if totalTasks = 0 |
| lastUpdated | string | ISO 8601 timestamp of most recent index.json mtime |
| healthStatus | string | "active" / "complete" / "stale" |

---

### Get Project

**Method**: `GET`
**Path**: `/api/projects/:id`
**Auth**: none

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| id | string (path) | Yes | Project identifier |

#### Response (200)

```json
{
  "id": "pm-work-tracker",
  "name": "pm-work-tracker",
  "path": "Z:/project/ai/pm-work-tracker",
  "featureCount": 24,
  "completedTasks": 312,
  "totalTasks": 330,
  "completionPct": 94.5,
  "lastUpdated": "2026-05-06T14:30:00Z",
  "healthStatus": "active",
  "features": [
    {
      "slug": "improve-ui",
      "status": "in-progress",
      "completedTasks": 13,
      "totalTasks": 18,
      "lastUpdated": "2026-05-06T12:00:00Z"
    }
  ],
  "warnings": []
}
```

---

### List Features

**Method**: `GET`
**Path**: `/api/projects/:id/features`
**Auth**: none

#### Response (200)

Array of feature summaries:

| Field | Type | Description |
|-------|------|-------------|
| slug | string | Feature identifier |
| status | string | "planning" / "in-progress" / "completed" |
| completedTasks | number | Count of completed tasks |
| totalTasks | number | Total task count |
| lastUpdated | string | ISO 8601 timestamp |

---

### Get Feature

**Method**: `GET`
**Path**: `/api/projects/:id/features/:slug`
**Auth**: none

#### Response (200)

Feature object with nested task summaries:

```json
{
  "slug": "improve-ui",
  "status": "in-progress",
  "prdPath": "prd/prd-spec.md",
  "designPath": "design/tech-design.md",
  "completedTasks": 13,
  "totalTasks": 18,
  "lastUpdated": "2026-05-06T12:00:00Z",
  "phases": [
    { "number": 1, "label": "Phase 1", "taskKeys": ["1.1-frontend", "1.2-backend"] },
    { "number": 2, "label": "Phase 2", "taskKeys": ["2.1-components"] }
  ],
  "tasks": {
    "1.1-interfaces": {
      "id": "1.1",
      "key": "1.1-interfaces",
      "title": "Define core interfaces",
      "status": "completed",
      "phase": 1
    }
  }
}
```

---

### List Tasks

**Method**: `GET`
**Path**: `/api/projects/:id/features/:slug/tasks`
**Auth**: none

#### Response (200)

Array of task objects:

| Field | Type | Description |
|-------|------|-------------|
| id | string | Task ID (e.g., "1.1") |
| key | string | Task key (e.g., "1.1-interfaces") |
| title | string | Task title |
| priority | string | "P0" / "P1" / "P2" |
| status | string | "pending" / "in_progress" / "completed" / "blocked" / "skipped" |
| scope | string | "frontend" / "backend" / "all" |
| estimatedTime | string | Optional estimate |
| dependencies | []string | Task IDs or wildcards |
| breaking | boolean | Triggers full test suite |
| phase | number | Derived from ID |
| file | string | Relative path to task .md |
| record | string | Relative path to record .md |

---

### Get Task

**Method**: `GET`
**Path**: `/api/projects/:id/features/:slug/tasks/:taskId`
**Auth**: none

#### Response (200)

Full task object with additional detail fields:

```json
{
  "id": "1.1",
  "key": "1.1-interfaces",
  "title": "Define core interfaces",
  "priority": "P0",
  "status": "completed",
  "scope": "all",
  "estimatedTime": "1-2h",
  "dependencies": [],
  "breaking": true,
  "phase": 1,
  "file": "1.1-interfaces.md",
  "record": "records/1.1-interfaces.md",
  "acceptanceCriteria": ["Interface defines all CRUD methods", "Types validated against schema"],
  "executionRecord": {
    "summary": "Implemented core interfaces",
    "files": ["internal/scanner/scanner.go", "internal/model/types.go"],
    "decisions": "Used fs.FS abstraction for testability",
    "testResults": "All 12 tests passing",
    "raw": "## Summary\nImplemented core interfaces..."
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| acceptanceCriteria | []string | Parsed from task .md file |
| executionRecord | object or null | Structured record sections; null if no record file exists |
| executionRecord.summary | string | Content under ## Summary heading |
| executionRecord.files | []string | File paths listed under ## Files heading |
| executionRecord.decisions | string | Content under ## Decisions heading |
| executionRecord.testResults | string | Content under ## Test Results heading |
| executionRecord.raw | string | Full markdown content for fallback rendering |

---

### Get Dependencies

**Method**: `GET`
**Path**: `/api/projects/:id/features/:slug/dependencies`
**Auth**: none

#### Response (200)

```json
{
  "nodes": [
    {
      "id": "1.1",
      "key": "1.1-interfaces",
      "title": "Define core interfaces",
      "status": "completed",
      "phase": 1,
      "feature": "improve-ui"
    }
  ],
  "edges": [
    {
      "source": "1.2",
      "target": "1.1",
      "crossFeature": false
    }
  ]
}
```

Wildcard dependencies ("1.x") are expanded to individual edges for each matching task.

---

## Data Contracts

### Meta Object

Every response includes a `meta` object:

```json
{
  "meta": {
    "lastUpdated": "2026-05-06T14:30:00Z",
    "refreshAvailable": true
  }
}
```

### Error Response

All errors return consistent format:

```json
{
  "error": "not_found",
  "message": "Project 'nonexistent' not found in configuration"
}
```

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| invalid_slug | 400 | Feature slug contains characters other than alphanumeric and hyphens |
| invalid_task_id | 400 | Task ID format is malformed or contains path traversal sequences |
| not_found | 404 | Project, feature, or task ID not found |
| internal_error | 500 | Filesystem read failure or JSON parse error |
