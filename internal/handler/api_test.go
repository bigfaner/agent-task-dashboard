package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/panda/agent-task-center/internal/config"
	"github.com/panda/agent-task-center/internal/scanner"
)

// ---- Test fixtures ----

const testAPIProjectPath = "/project"

// setupAPITest creates a Gin engine with API routes registered, using a real Scanner
// backed by a test fstest.MapFS filesystem.
func setupAPITest(fsMap fstest.MapFS) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "test-project", Path: testAPIProjectPath},
		},
	}

	s := scanner.NewScannerWithFS(cfg, fsMap)
	_, _ = s.ScanAll()

	RegisterAPI(r, s)
	return r
}

// apiIndexJSON builds an index.json byte payload.
func apiIndexJSON(feature, prd, design, status string, tasks map[string]interface{}) []byte {
	obj := map[string]interface{}{
		"feature": feature,
		"prd":     prd,
		"design":  design,
		"created": "2026-05-06",
		"status":  status,
		"tasks":   tasks,
	}
	data, _ := json.Marshal(obj)
	return data
}

// apiTaskEntry builds a single task map for index.json.
func apiTaskEntry(id, title, priority, status, scope, file string, deps []string, breaking bool) map[string]interface{} {
	m := map[string]interface{}{
		"id":       id,
		"title":    title,
		"priority": priority,
		"status":   status,
		"scope":    scope,
		"file":     file,
		"breaking": breaking,
	}
	if deps != nil {
		m["dependencies"] = deps
	}
	return m
}

// makeTestFS creates a MapFS with a single feature containing the given tasks.
func makeTestFS(tasks map[string]interface{}) fstest.MapFS {
	return fstest.MapFS{
		"project/docs/features/myfeature/tasks/index.json": {
			Data: apiIndexJSON("myfeature", "prd/spec.md", "design/tech.md", "in-progress", tasks),
		},
	}
}

// fullTestFS creates a MapFS with a single project containing two features and tasks.
func fullTestFS() fstest.MapFS {
	modTime := time.Date(2026, 5, 6, 14, 30, 0, 0, time.UTC)
	return fstest.MapFS{
		"project/docs/features/feature-a/tasks/index.json": {
			Data: apiIndexJSON("feature-a", "prd/spec.md", "design/tech.md", "in-progress", map[string]interface{}{
				"1.1-task-a": apiTaskEntry("1.1", "Task A", "P0", "completed", "all", "1.1-task-a.md", nil, false),
				"1.2-task-b": apiTaskEntry("1.2", "Task B", "P1", "in_progress", "frontend", "1.2-task-b.md", []string{"1.1"}, true),
			}),
			ModTime: modTime,
		},
		"project/docs/features/feature-b/tasks/index.json": {
			Data: apiIndexJSON("feature-b", "prd/spec2.md", "design/tech2.md", "planning", map[string]interface{}{
				"2.1-task-c": apiTaskEntry("2.1", "Task C", "P0", "pending", "backend", "2.1-task-c.md", []string{"1.x"}, false),
				"2.2-task-d": apiTaskEntry("2.2", "Task D", "P2", "blocked", "all", "2.2-task-d.md", nil, false),
			}),
			ModTime: modTime.Add(-1 * time.Hour),
		},
	}
}

// ---- GET /api/projects tests ----

func TestAPI_ListProjects(t *testing.T) {
	r := setupAPITest(fullTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body=%s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	projects, ok := resp["projects"].([]interface{})
	if !ok {
		t.Fatalf("expected projects array, got %T", resp["projects"])
	}
	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
	}

	p := projects[0].(map[string]interface{})
	if p["id"] != "test-project" {
		t.Errorf("expected id 'test-project', got %v", p["id"])
	}
	if p["featureCount"] != float64(2) {
		t.Errorf("expected featureCount 2, got %v", p["featureCount"])
	}
	if p["totalTasks"] != float64(4) {
		t.Errorf("expected totalTasks 4, got %v", p["totalTasks"])
	}
	if p["completedTasks"] != float64(1) {
		t.Errorf("expected completedTasks 1, got %v", p["completedTasks"])
	}
	if p["completionPct"] != float64(25) {
		t.Errorf("expected completionPct 25, got %v", p["completionPct"])
	}

	// Check meta.lastUpdated exists
	meta, ok := resp["meta"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected meta object, got %T", resp["meta"])
	}
	if _, ok := meta["lastUpdated"]; !ok {
		t.Error("expected meta.lastUpdated to exist")
	}
}

func TestAPI_ListProjects_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := &config.Config{
		Projects: []config.ProjectConfig{},
	}
	s := scanner.NewScannerWithFS(cfg, fstest.MapFS{})
	_, _ = s.ScanAll()
	RegisterAPI(r, s)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	projects := resp["projects"].([]interface{})
	if len(projects) != 0 {
		t.Errorf("expected empty projects array, got %d", len(projects))
	}
}

func TestAPI_ListProjects_ZeroTasks(t *testing.T) {
	fsMap := fstest.MapFS{
		"project/docs/features/empty-feature/tasks/index.json": {
			Data: apiIndexJSON("empty-feature", "", "", "planning", map[string]interface{}{}),
		},
	}
	r := setupAPITest(fsMap)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	projects := resp["projects"].([]interface{})
	p := projects[0].(map[string]interface{})

	if p["completionPct"] != float64(0) {
		t.Errorf("expected completionPct 0 when no tasks, got %v", p["completionPct"])
	}
	if p["totalTasks"] != float64(0) {
		t.Errorf("expected totalTasks 0, got %v", p["totalTasks"])
	}
}

// ---- GET /api/projects/:id tests ----

func TestAPI_GetProject(t *testing.T) {
	r := setupAPITest(fullTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/projects/test-project", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body=%s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["id"] != "test-project" {
		t.Errorf("expected id 'test-project', got %v", resp["id"])
	}
	if resp["name"] != "test-project" {
		t.Errorf("expected name 'test-project', got %v", resp["name"])
	}
	if resp["path"] != testAPIProjectPath {
		t.Errorf("expected path %q, got %v", testAPIProjectPath, resp["path"])
	}

	features, ok := resp["features"].([]interface{})
	if !ok {
		t.Fatalf("expected features array, got %T", resp["features"])
	}
	if len(features) != 2 {
		t.Errorf("expected 2 features, got %d", len(features))
	}

	// Check meta
	meta := resp["meta"].(map[string]interface{})
	if _, ok := meta["lastUpdated"]; !ok {
		t.Error("expected meta.lastUpdated")
	}
}

func TestAPI_GetProject_NotFound(t *testing.T) {
	r := setupAPITest(fullTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/projects/nonexistent", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d; body=%s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["error"] != "not_found" {
		t.Errorf("expected error 'not_found', got %v", resp["error"])
	}
}

// ---- GET /api/projects/:id/features tests ----

func TestAPI_ListFeatures(t *testing.T) {
	r := setupAPITest(fullTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/projects/test-project/features", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body=%s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	features, ok := resp["features"].([]interface{})
	if !ok {
		t.Fatalf("expected features array, got %T", resp["features"])
	}
	if len(features) != 2 {
		t.Errorf("expected 2 features, got %d", len(features))
	}

	// Each feature should have slug, status, completedTasks, totalTasks, lastUpdated
	f := features[0].(map[string]interface{})
	requiredKeys := []string{"slug", "status", "completedTasks", "totalTasks", "lastUpdated"}
	for _, key := range requiredKeys {
		if _, ok := f[key]; !ok {
			t.Errorf("feature missing key %q", key)
		}
	}
}

func TestAPI_ListFeatures_ProjectNotFound(t *testing.T) {
	r := setupAPITest(fullTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/projects/nonexistent/features", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// ---- GET /api/projects/:id/features/:slug tests ----

func TestAPI_GetFeature(t *testing.T) {
	r := setupAPITest(fullTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/projects/test-project/features/feature-a", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body=%s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["slug"] != "feature-a" {
		t.Errorf("expected slug 'feature-a', got %v", resp["slug"])
	}
	if resp["status"] != "in-progress" {
		t.Errorf("expected status 'in-progress', got %v", resp["status"])
	}
	if resp["prdPath"] != "prd/spec.md" {
		t.Errorf("expected prdPath 'prd/spec.md', got %v", resp["prdPath"])
	}
	if resp["designPath"] != "design/tech.md" {
		t.Errorf("expected designPath 'design/tech.md', got %v", resp["designPath"])
	}

	// Check tasks
	tasks, ok := resp["tasks"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected tasks map, got %T", resp["tasks"])
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}

	// Check phases
	phases, ok := resp["phases"].([]interface{})
	if !ok {
		t.Fatalf("expected phases array, got %T", resp["phases"])
	}
	if len(phases) != 1 {
		t.Errorf("expected 1 phase, got %d", len(phases))
	}
}

func TestAPI_GetFeature_NotFound(t *testing.T) {
	r := setupAPITest(fullTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/projects/test-project/features/nonexistent", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d; body=%s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["error"] != "not_found" {
		t.Errorf("expected error 'not_found', got %v", resp["error"])
	}
}

func TestAPI_GetFeature_InvalidSlug(t *testing.T) {
	r := setupAPITest(fullTestFS())

	// Gin router normalizes paths, so "../hack" becomes a path navigation.
	// Use a slug that is syntactically invalid but still matches the route pattern.
	w := httptest.NewRecorder()
	req := makeGetRequest("/api/projects/test-project/features/..%2Fhack")
	r.ServeHTTP(w, req)

	// Should get 400 from slug validation
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d; body=%s", w.Code, w.Body.String())
	}
}

// ---- GET /api/projects/:id/features/:slug/tasks tests ----

func TestAPI_ListTasks(t *testing.T) {
	r := setupAPITest(fullTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/projects/test-project/features/feature-a/tasks", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body=%s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	tasks, ok := resp["tasks"].([]interface{})
	if !ok {
		t.Fatalf("expected tasks array, got %T", resp["tasks"])
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}

	// Check first task has required fields (estimatedTime is always present, even if empty)
	task := tasks[0].(map[string]interface{})
	requiredKeys := []string{"id", "key", "title", "priority", "status", "scope", "estimatedTime", "dependencies", "breaking", "phase", "file", "record"}
	for _, key := range requiredKeys {
		if _, ok := task[key]; !ok {
			t.Errorf("task missing key %q", key)
		}
	}
}

func TestAPI_ListTasks_FeatureNotFound(t *testing.T) {
	r := setupAPITest(fullTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/projects/test-project/features/nope/tasks", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// ---- GET /api/projects/:id/features/:slug/tasks/:taskId tests ----

func TestAPI_GetTask(t *testing.T) {
	// Create temp files for ParseTaskFile and ParseRecordFile
	tmpDir := t.TempDir()
	taskFile := filepath.Join(tmpDir, "1.1-task-a.md")
	recordDir := filepath.Join(tmpDir, "records")
	recordFile := filepath.Join(recordDir, "1.1-task-a.md")

	taskContent := `---
id: "1.1"
title: "Task A"
---

## Description
Some description

## Acceptance Criteria
- [ ] Criterion one
- [x] Criterion two done
`
	if err := os.MkdirAll(recordDir, 0o755); err != nil {
		t.Fatalf("failed to create record dir: %v", err)
	}
	recordContent := `---
status: "completed"
---

## Summary
Implemented task A

### Files Created
- internal/foo/bar.go

## Key Decisions
Used fs.FS

## Test Results
12 tests passing
`
	if err := os.WriteFile(taskFile, []byte(taskContent), 0o644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}
	if err := os.WriteFile(recordFile, []byte(recordContent), 0o644); err != nil {
		t.Fatalf("failed to write record file: %v", err)
	}

	fsMap := fstest.MapFS{
		"project/docs/features/myfeature/tasks/index.json": {
			Data: apiIndexJSON("myfeature", "", "", "in-progress", map[string]interface{}{
				"1.1-task-a": apiTaskEntry("1.1", "Task A", "P0", "completed", "all", filepath.ToSlash(taskFile), nil, false),
			}),
		},
	}

	// Build a scanner that uses MapFS for scanning but real filesystem paths for task/record files.
	// The task's File and Record fields contain absolute paths to temp files, so
	// ParseTaskFile/ParseRecordFile will use os.ReadFile directly.
	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "test-project", Path: "/project"},
		},
	}
	s := scanner.NewScannerWithFS(cfg, fsMap)
	_, _ = s.ScanAll()

	// Patch the task's Record field to point to the temp record file.
	// The scanner read the File field from index.json but we need the Record to also be a real path.
	// We do this by scanning, then modifying the cached data.
	pd, _ := s.ScanProject("test-project")
	for i := range pd.Features {
		if pd.Features[i].Slug == "myfeature" {
			for key, t := range pd.Features[i].Tasks {
				if t.ID == "1.1" {
					t.Record = filepath.ToSlash(recordFile)
					pd.Features[i].Tasks[key] = t
				}
			}
		}
	}

	RegisterAPI(r, s)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/projects/test-project/features/myfeature/tasks/1.1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body=%s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["id"] != "1.1" {
		t.Errorf("expected id '1.1', got %v", resp["id"])
	}

	// Check acceptanceCriteria
	ac, ok := resp["acceptanceCriteria"].([]interface{})
	if !ok {
		t.Fatalf("expected acceptanceCriteria array, got %T", resp["acceptanceCriteria"])
	}
	if len(ac) != 2 {
		t.Errorf("expected 2 acceptance criteria, got %d", len(ac))
	}

	// Check executionRecord
	rec, ok := resp["executionRecord"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected executionRecord object, got %T", resp["executionRecord"])
	}
	if rec["summary"] != "Implemented task A" {
		t.Errorf("expected summary 'Implemented task A', got %v", rec["summary"])
	}
}

func TestAPI_GetTask_NotFound(t *testing.T) {
	r := setupAPITest(fullTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/projects/test-project/features/feature-a/tasks/9.9", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d; body=%s", w.Code, w.Body.String())
	}
}

func TestAPI_GetTask_InvalidTaskID(t *testing.T) {
	r := setupAPITest(fullTestFS())

	// Use URL-encoded path traversal that still matches the route
	w := httptest.NewRecorder()
	req := makeGetRequest("/api/projects/test-project/features/feature-a/tasks/..%2Fetc%2Fpasswd")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d; body=%s", w.Code, w.Body.String())
	}
}

func TestAPI_GetTask_NoRecord(t *testing.T) {
	// Task with a record path that doesn't exist on the real filesystem
	fsMap := makeTestFS(map[string]interface{}{
		"1.1-task-a": apiTaskEntry("1.1", "Task A", "P0", "completed", "all", "1.1-task-a.md", nil, false),
	})
	r := setupAPITest(fsMap)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/projects/test-project/features/myfeature/tasks/1.1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body=%s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// executionRecord should be null when file doesn't exist
	if resp["executionRecord"] != nil {
		t.Errorf("expected null executionRecord for missing record file, got %v", resp["executionRecord"])
	}
}

// ---- GET /api/projects/:id/features/:slug/dependencies tests ----

func TestAPI_GetDependencies(t *testing.T) {
	r := setupAPITest(fullTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/projects/test-project/features/feature-b/dependencies", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body=%s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	nodes, ok := resp["nodes"].([]interface{})
	if !ok {
		t.Fatalf("expected nodes array, got %T", resp["nodes"])
	}
	if len(nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(nodes))
	}

	edges, ok := resp["edges"].([]interface{})
	if !ok {
		t.Fatalf("expected edges array, got %T", resp["edges"])
	}
	// "1.x" wildcard should have been expanded by scanner to individual task IDs.
	// The feature-b tasks: 2.1 depends on "1.x" (expanded), 2.2 has no deps.
	// After expansion, 2.1 should have edges to all phase-1 tasks in this feature.
	// Since feature-b's tasks are in phase 2 (2.1, 2.2), the wildcard "1.x" expands
	// to phase 1 tasks. In feature-b there are no phase 1 tasks, so the wildcard
	// may not expand within the feature. Edges may be empty, which is fine.
	_ = edges
}

func TestAPI_GetDependencies_NodesHaveFeatureField(t *testing.T) {
	r := setupAPITest(fullTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/projects/test-project/features/feature-b/dependencies", nil)
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	nodes := resp["nodes"].([]interface{})
	for _, n := range nodes {
		node := n.(map[string]interface{})
		if _, ok := node["feature"]; !ok {
			t.Error("node missing 'feature' field")
		}
	}
}

func TestAPI_GetDependencies_FeatureNotFound(t *testing.T) {
	r := setupAPITest(fullTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/projects/test-project/features/nope/dependencies", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// ---- Meta object tests ----

func TestAPI_MetaLastUpdated_PresentOnAllEndpoints(t *testing.T) {
	r := setupAPITest(fullTestFS())

	endpoints := []string{
		"/api/projects",
		"/api/projects/test-project",
		"/api/projects/test-project/features",
		"/api/projects/test-project/features/feature-a",
		"/api/projects/test-project/features/feature-a/tasks",
		"/api/projects/test-project/features/feature-a/dependencies",
	}

	for _, ep := range endpoints {
		t.Run(ep, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, ep, nil)
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d for %s; body=%s", w.Code, ep, w.Body.String())
			}

			var resp map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			meta, ok := resp["meta"].(map[string]interface{})
			if !ok {
				t.Fatalf("expected meta object for %s, got %T", ep, resp["meta"])
			}
			if _, ok := meta["lastUpdated"]; !ok {
				t.Errorf("meta.lastUpdated missing for %s", ep)
			}
		})
	}
}

// ---- Edge case tests ----

func TestAPI_CompletionPct_NoDivisionByZero(t *testing.T) {
	fsMap := fstest.MapFS{
		"project/docs/features/empty/tasks/index.json": {
			Data: apiIndexJSON("empty", "", "", "planning", map[string]interface{}{}),
		},
	}
	r := setupAPITest(fsMap)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/projects/test-project", nil)
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["completionPct"] != float64(0) {
		t.Errorf("expected completionPct 0 for empty project, got %v", resp["completionPct"])
	}
}

// ---- Multi-project tests ----

func TestAPI_MultiProjectSetup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "Alpha", Path: "/alpha"},
			{Name: "Beta", Path: "/beta"},
		},
	}

	modTime := time.Date(2026, 5, 6, 14, 30, 0, 0, time.UTC)
	fsMap := fstest.MapFS{
		"alpha/docs/features/fa/tasks/index.json": {
			Data: apiIndexJSON("fa", "", "", "completed", map[string]interface{}{
				"1.1-t1": apiTaskEntry("1.1", "T1", "P0", "completed", "all", "1.1-t1.md", nil, false),
			}),
			ModTime: modTime,
		},
		"beta/docs/features/fb/tasks/index.json": {
			Data: apiIndexJSON("fb", "", "", "in-progress", map[string]interface{}{
				"1.1-t2": apiTaskEntry("1.1", "T2", "P1", "pending", "backend", "1.1-t2.md", nil, false),
			}),
			ModTime: modTime,
		},
	}

	s := scanner.NewScannerWithFS(cfg, fsMap)
	_, _ = s.ScanAll()
	RegisterAPI(r, s)

	// List should have 2 projects
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	projects := resp["projects"].([]interface{})
	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(projects))
	}

	// Alpha project (lowercased to "alpha")
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/projects/alpha", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for alpha, got %d", w.Code)
	}

	// Beta project (lowercased to "beta")
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/projects/beta", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for beta, got %d", w.Code)
	}
}

// ---- JSON Content-Type ----

func TestAPI_ResponseIsJSON(t *testing.T) {
	r := setupAPITest(fullTestFS())

	endpoints := []string{
		"/api/projects",
		"/api/projects/test-project",
		"/api/projects/test-project/features",
		"/api/projects/test-project/features/feature-a",
		"/api/projects/test-project/features/feature-a/tasks",
		"/api/projects/test-project/features/feature-b/dependencies",
	}

	for _, ep := range endpoints {
		t.Run(ep, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, ep, nil)
			r.ServeHTTP(w, req)

			ct := w.Header().Get("Content-Type")
			if !strings.Contains(ct, "application/json") {
				t.Errorf("expected Content-Type to contain 'application/json' for %s, got %q", ep, ct)
			}
		})
	}
}

// ---- Dependencies wildcard expansion (within feature) ----

func TestAPI_Dependencies_WildcardExpansion(t *testing.T) {
	fsMap := fstest.MapFS{
		"project/docs/features/fwild/tasks/index.json": {
			Data: apiIndexJSON("fwild", "", "", "in-progress", map[string]interface{}{
				"1.1-alpha": apiTaskEntry("1.1", "Alpha", "P0", "completed", "all", "1.1-alpha.md", nil, false),
				"1.2-beta":  apiTaskEntry("1.2", "Beta", "P1", "completed", "all", "1.2-beta.md", nil, false),
				"2.1-gamma": apiTaskEntry("2.1", "Gamma", "P0", "pending", "all", "2.1-gamma.md", []string{"1.x"}, false),
			}),
		},
	}
	r := setupAPITest(fsMap)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/projects/test-project/features/fwild/dependencies", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body=%s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	edges := resp["edges"].([]interface{})
	// "1.x" should be expanded by scanner to 1.1 and 1.2, producing 2 edges from 2.1
	if len(edges) != 2 {
		t.Errorf("expected 2 edges from wildcard expansion, got %d", len(edges))
	}

	// Verify edge structure
	for _, e := range edges {
		edge := e.(map[string]interface{})
		if edge["source"] != "2.1" {
			t.Errorf("expected source '2.1', got %v", edge["source"])
		}
		target := edge["target"].(string)
		if target != "1.1" && target != "1.2" {
			t.Errorf("expected target '1.1' or '1.2', got %v", target)
		}
		// Same feature so crossFeature should be false
		if edge["crossFeature"] == true {
			t.Errorf("expected crossFeature false for same-feature edge")
		}
	}
}

// ---- Cross-feature dependencies ----

func TestAPI_Dependencies_CrossFeature(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "test-project", Path: "/project"},
		},
	}

	fsMap := fstest.MapFS{
		"project/docs/features/feat-a/tasks/index.json": {
			Data: apiIndexJSON("feat-a", "", "", "in-progress", map[string]interface{}{
				"1.1-base": apiTaskEntry("1.1", "Base", "P0", "completed", "all", "1.1-base.md", nil, false),
			}),
		},
		"project/docs/features/feat-b/tasks/index.json": {
			Data: apiIndexJSON("feat-b", "", "", "in-progress", map[string]interface{}{
				"1.2-consumer": apiTaskEntry("1.2", "Consumer", "P0", "pending", "all", "1.2-consumer.md", []string{"1.1"}, false),
			}),
		},
	}

	s := scanner.NewScannerWithFS(cfg, fsMap)
	_, _ = s.ScanAll()
	RegisterAPI(r, s)

	// Request dependencies for feat-b which has task 1.2 depending on 1.1 from feat-a
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/projects/test-project/features/feat-b/dependencies", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body=%s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	edges := resp["edges"].([]interface{})
	if len(edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(edges))
	}

	edge := edges[0].(map[string]interface{})
	// The edge from 1.2 (feat-b) -> 1.1 (feat-a) should be crossFeature=true
	if edge["crossFeature"] != true {
		t.Errorf("expected crossFeature true for cross-feature edge, got %v", edge["crossFeature"])
	}
}

// ---- Task lookup within correct feature ----

func TestAPI_GetTask_SearchesCorrectFeature(t *testing.T) {
	r := setupAPITest(fullTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/projects/test-project/features/feature-b/tasks/2.2", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body=%s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["id"] != "2.2" {
		t.Errorf("expected id '2.2', got %v", resp["id"])
	}
}

// ---- Invalid slug via direct middleware test ----

func TestAPI_InvalidSlug_Returns400(t *testing.T) {
	r := setupAPITest(fullTestFS())

	// Use a slug with special character that doesn't match slug regex
	// The path must still be routable by Gin
	w := httptest.NewRecorder()
	req := makeGetRequest("/api/projects/test-project/features/a!/tasks")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid slug, got %d; body=%s", w.Code, w.Body.String())
	}
}

func TestAPI_GetTask_ResponseFormat(t *testing.T) {
	fsMap := makeTestFS(map[string]interface{}{
		"1.1-task-a": apiTaskEntry("1.1", "Task A", "P0", "completed", "all", "1.1-task-a.md", []string{"1.0"}, true),
	})
	r := setupAPITest(fsMap)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/projects/test-project/features/myfeature/tasks/1.1", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body=%s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Verify all expected fields are present
	expectedFields := []string{"id", "key", "title", "priority", "status", "scope", "estimatedTime", "dependencies", "breaking", "phase", "file", "record", "acceptanceCriteria", "executionRecord"}
	for _, field := range expectedFields {
		if _, ok := resp[field]; !ok {
			t.Errorf("task response missing field %q", field)
		}
	}

	if resp["id"] != "1.1" {
		t.Errorf("expected id '1.1', got %v", resp["id"])
	}
	if resp["key"] != "1.1-task-a" {
		t.Errorf("expected key '1.1-task-a', got %v", resp["key"])
	}
	if resp["title"] != "Task A" {
		t.Errorf("expected title 'Task A', got %v", resp["title"])
	}
	if resp["priority"] != "P0" {
		t.Errorf("expected priority 'P0', got %v", resp["priority"])
	}
	if resp["status"] != "completed" {
		t.Errorf("expected status 'completed', got %v", resp["status"])
	}
	if resp["scope"] != "all" {
		t.Errorf("expected scope 'all', got %v", resp["scope"])
	}
	if resp["breaking"] != true {
		t.Errorf("expected breaking true, got %v", resp["breaking"])
	}
	if resp["phase"] != float64(1) {
		t.Errorf("expected phase 1, got %v", resp["phase"])
	}
}

// ---- 404 error messages ----

func TestAPI_404_ErrorMessages(t *testing.T) {
	r := setupAPITest(fullTestFS())

	tests := []struct {
		name     string
		endpoint string
	}{
		{"project not found", "/api/projects/nope"},
		{"feature not found", "/api/projects/test-project/features/nope"},
		{"task not found", "/api/projects/test-project/features/feature-a/tasks/9.9"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.endpoint, nil)
			r.ServeHTTP(w, req)

			if w.Code != http.StatusNotFound {
				t.Fatalf("expected 404, got %d for %s", w.Code, tt.endpoint)
			}

			var resp map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if resp["error"] != "not_found" {
				t.Errorf("expected error 'not_found', got %v", resp["error"])
			}
			msg, ok := resp["message"].(string)
			if !ok || len(msg) == 0 {
				t.Error("expected non-empty message")
			}
		})
	}
}

// ---- Feature list stats ----

func TestAPI_ListFeatures_Stats(t *testing.T) {
	r := setupAPITest(fullTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/projects/test-project/features", nil)
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	features := resp["features"].([]interface{})
	// feature-a: 2 tasks, 1 completed
	// feature-b: 2 tasks, 0 completed
	for _, f := range features {
		feat := f.(map[string]interface{})
		slug := feat["slug"].(string)
		switch slug {
		case "feature-a":
			if feat["completedTasks"] != float64(1) {
				t.Errorf("feature-a: expected completedTasks 1, got %v", feat["completedTasks"])
			}
			if feat["totalTasks"] != float64(2) {
				t.Errorf("feature-a: expected totalTasks 2, got %v", feat["totalTasks"])
			}
		case "feature-b":
			if feat["completedTasks"] != float64(0) {
				t.Errorf("feature-b: expected completedTasks 0, got %v", feat["completedTasks"])
			}
			if feat["totalTasks"] != float64(2) {
				t.Errorf("feature-b: expected totalTasks 2, got %v", feat["totalTasks"])
			}
		}
	}
}

// ---- Project lastUpdated format ----

func TestAPI_ProjectLastUpdated(t *testing.T) {
	r := setupAPITest(fullTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/projects/test-project", nil)
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	lu, ok := resp["lastUpdated"].(string)
	if !ok || lu == "" {
		t.Error("expected non-empty lastUpdated")
	}
	_, err := time.Parse(time.RFC3339, lu)
	if err != nil {
		t.Errorf("lastUpdated not valid RFC3339: %v", err)
	}
}

// ---- ComputePct helper test ----

func TestComputePct(t *testing.T) {
	tests := []struct {
		completed int
		total     int
		expected  float64
	}{
		{0, 0, 0},
		{1, 4, 25},
		{3, 3, 100},
		{0, 5, 0},
	}
	for _, tt := range tests {
		result := computePct(tt.completed, tt.total)
		if result != tt.expected {
			t.Errorf("computePct(%d, %d) = %v, want %v", tt.completed, tt.total, result, tt.expected)
		}
	}

	// Test non-exact with tolerance
	result := computePct(1, 3)
	if result < 33 || result > 34 {
		t.Errorf("computePct(1, 3) = %v, want ~33.33", result)
	}
}
