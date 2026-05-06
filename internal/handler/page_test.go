package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/panda/agent-task-center/internal/config"
	"github.com/panda/agent-task-center/internal/model"
	"github.com/panda/agent-task-center/internal/scanner"
)

// ---- Page test helpers ----

// setupPageTest creates a Gin engine with page routes registered, using a real Scanner
// backed by a test fstest.MapFS filesystem.
func setupPageTest(fsMap fstest.MapFS) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "test-project", Path: testAPIProjectPath},
		},
	}

	s := scanner.NewScannerWithFS(cfg, fsMap)
	_, _ = s.ScanAll()

	RegisterPages(r, s)
	return r
}

// fullPageTestFS creates a MapFS with a project containing features for page tests.
func fullPageTestFS() fstest.MapFS {
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
			}),
			ModTime: modTime.Add(-1 * time.Hour),
		},
	}
}

// ---- GET / tests (landing page) ----

func TestPage_Landing_Returns200(t *testing.T) {
	r := setupPageTest(fullPageTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body=%s", w.Code, w.Body.String())
	}
}

func TestPage_Landing_ContentTypeHTML(t *testing.T) {
	r := setupPageTest(fullPageTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("expected Content-Type to contain 'text/html', got %q", ct)
	}
}

func TestPage_Landing_ContainsProjectNames(t *testing.T) {
	r := setupPageTest(fullPageTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "test-project") {
		t.Errorf("expected body to contain 'test-project', got: %s", body)
	}
}

func TestPage_Landing_ContainsProjectData(t *testing.T) {
	r := setupPageTest(fullPageTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()

	// Should contain the JSON data for JS to consume
	if !strings.Contains(body, "test-project") {
		t.Error("expected body to contain project name")
	}

	// Should contain data injection for JS rendering
	if !strings.Contains(body, "__PROJECTS_DATA__") {
		t.Error("expected body to contain __PROJECTS_DATA__ script variable")
	}

	// Should contain the card grid container for JS to render into
	if !strings.Contains(body, "project-cards") {
		t.Error("expected body to contain project-cards container")
	}
}

func TestPage_Landing_ContainsThemeToggle(t *testing.T) {
	r := setupPageTest(fullPageTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "theme-toggle") {
		t.Error("expected body to contain theme-toggle button")
	}
}

func TestPage_Landing_ContainsEmptyState(t *testing.T) {
	r := setupPageTest(fullPageTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "No projects configured") {
		t.Error("expected body to contain empty state message")
	}
}

func TestPage_Landing_EmptyProjects(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := &config.Config{
		Projects: []config.ProjectConfig{},
	}
	s := scanner.NewScannerWithFS(cfg, fstest.MapFS{})
	_, _ = s.ScanAll()
	RegisterPages(r, s)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	body := w.Body.String()
	// Empty config should still show empty state container
	if !strings.Contains(body, "empty-state") {
		t.Error("expected body to contain empty-state container")
	}
}

func TestPage_Landing_InvalidPathShowsWarnings(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "bad-project", Path: "/nonexistent/path"},
		},
	}
	s := scanner.NewScannerWithFS(cfg, fstest.MapFS{})
	_, _ = s.ScanAll()
	RegisterPages(r, s)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	body := w.Body.String()
	// Should contain warnings in the JSON data for JS to render
	if !strings.Contains(body, "warnings") {
		t.Error("expected body to contain warnings in project data")
	}
}

func TestPage_Landing_MultipleProjects(t *testing.T) {
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
	RegisterPages(r, s)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body=%s", w.Code, w.Body.String())
	}

	body := w.Body.String()
	if !strings.Contains(body, "Alpha") {
		t.Error("expected body to contain 'Alpha'")
	}
	if !strings.Contains(body, "Beta") {
		t.Error("expected body to contain 'Beta'")
	}
}

func TestPage_Landing_HTMLEscaping(t *testing.T) {
	// Ensure html/template auto-escaping works - no raw script injection
	gin.SetMode(gin.TestMode)
	r := gin.New()

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "test", Path: "/test"},
		},
	}

	// Use a project name that would be dangerous if not escaped
	fsMap := fstest.MapFS{
		"test/docs/features/feat/tasks/index.json": {
			Data: apiIndexJSON("feat", "", "", "in-progress", map[string]interface{}{
				"1.1-t": apiTaskEntry("1.1", "Task <script>alert('xss')</script>", "P0", "completed", "all", "1.1-t.md", nil, false),
			}),
		},
	}

	s := scanner.NewScannerWithFS(cfg, fsMap)
	_, _ = s.ScanAll()
	RegisterPages(r, s)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	// The template should escape the script tag, not render it raw
	if strings.Contains(body, "<script>alert('xss')</script>") {
		t.Error("XSS: unescaped script tag found in landing page output")
	}
}

// ---- GET /projects/:id tests (project page) ----

func TestPage_Project_Returns200(t *testing.T) {
	r := setupPageTest(fullPageTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/projects/test-project", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body=%s", w.Code, w.Body.String())
	}
}

func TestPage_Project_ContentTypeHTML(t *testing.T) {
	r := setupPageTest(fullPageTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/projects/test-project", nil)
	r.ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("expected Content-Type to contain 'text/html', got %q", ct)
	}
}

func TestPage_Project_ContainsFeatureData(t *testing.T) {
	r := setupPageTest(fullPageTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/projects/test-project", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "feature-a") {
		t.Error("expected body to contain 'feature-a'")
	}
	if !strings.Contains(body, "feature-b") {
		t.Error("expected body to contain 'feature-b'")
	}
}

func TestPage_Project_ContainsProjectID(t *testing.T) {
	r := setupPageTest(fullPageTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/projects/test-project", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "test-project") {
		t.Error("expected body to contain 'test-project'")
	}
}

func TestPage_Project_NotFound_Returns404(t *testing.T) {
	r := setupPageTest(fullPageTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/projects/nonexistent", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d; body=%s", w.Code, w.Body.String())
	}
}

func TestPage_Project_404_ContentTypeHTML(t *testing.T) {
	r := setupPageTest(fullPageTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/projects/nonexistent", nil)
	r.ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("expected Content-Type to contain 'text/html' for 404, got %q", ct)
	}
}

func TestPage_Project_ContainsSVGContainer(t *testing.T) {
	r := setupPageTest(fullPageTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/projects/test-project", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	// Should contain SVG container for DAG rendering
	if !strings.Contains(body, "svg") && !strings.Contains(body, "dag") {
		t.Error("expected body to contain SVG container or dag reference")
	}
}

func TestPage_Project_ContainsDetailPanel(t *testing.T) {
	r := setupPageTest(fullPageTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/projects/test-project", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	// Should contain detail panel placeholder
	if !strings.Contains(body, "detail") && !strings.Contains(body, "panel") {
		t.Error("expected body to contain detail panel placeholder")
	}
}

func TestPage_Project_ContainsActivitySidebar(t *testing.T) {
	r := setupPageTest(fullPageTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/projects/test-project", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	// Should contain activity sidebar placeholder
	if !strings.Contains(body, "activity") {
		t.Error("expected body to contain activity sidebar placeholder")
	}
}

func TestPage_Project_ContainsJSONDataForJS(t *testing.T) {
	r := setupPageTest(fullPageTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/projects/test-project", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()

	// Should contain JSON data in a script tag for JS to consume
	if !strings.Contains(body, "application/json") && !strings.Contains(body, "__INITIAL_DATA__") {
		t.Error("expected body to contain JSON data for JS consumption")
	}

	// Verify JSON data is valid by extracting it
	// Look for the script tag with JSON data
	startIdx := strings.Index(body, "application/json")
	if startIdx == -1 {
		startIdx = strings.Index(body, "__INITIAL_DATA__")
	}
	if startIdx == -1 {
		t.Fatal("could not find JSON data block in response")
	}

	// Extract JSON from script tag - find data between script tags
	// Look for type="application/json" pattern
	scriptStart := strings.Index(body[startIdx:], ">")
	if scriptStart == -1 {
		t.Fatal("could not find script data start")
	}

	scriptEnd := strings.Index(body[startIdx+scriptStart:], "</script>")
	if scriptEnd == -1 {
		t.Fatal("could not find script data end")
	}

	jsonData := body[startIdx+scriptStart+1 : startIdx+scriptStart+scriptEnd]
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &parsed); err != nil {
		t.Errorf("JSON data in script tag is not valid: %v\nJSON: %s", err, jsonData)
	}
}

func TestPage_Project_HTMLEscaping(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "test", Path: "/test"},
		},
	}

	fsMap := fstest.MapFS{
		"test/docs/features/feat/tasks/index.json": {
			Data: apiIndexJSON("feat", "", "", "in-progress", map[string]interface{}{
				"1.1-t": apiTaskEntry("1.1", "Task <script>alert('xss')</script>", "P0", "completed", "all", "1.1-t.md", nil, false),
			}),
		},
	}

	s := scanner.NewScannerWithFS(cfg, fsMap)
	_, _ = s.ScanAll()
	RegisterPages(r, s)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/projects/test", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	if strings.Contains(body, "<script>alert('xss')</script>") {
		t.Error("XSS: unescaped script tag found in project page output")
	}
}

// ---- Template rendering tests ----

func TestPage_Landing_ContainsCSSLink(t *testing.T) {
	r := setupPageTest(fullPageTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "stylesheet") && !strings.Contains(body, ".css") {
		t.Error("expected body to contain CSS link")
	}
}

func TestPage_Landing_ContainsJS(t *testing.T) {
	r := setupPageTest(fullPageTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "<script") {
		t.Error("expected body to contain script tag")
	}
}

func TestPage_Project_ContainsJS(t *testing.T) {
	r := setupPageTest(fullPageTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/projects/test-project", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "<script") {
		t.Error("expected body to contain script tag")
	}
}

// ---- Activity sidebar data tests ----

func TestPage_Project_ContainsActivityEvents(t *testing.T) {
	r := setupPageTest(fullPageTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/projects/test-project", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()

	// The JSON data should include activityEvents
	if !strings.Contains(body, "activityEvents") {
		t.Error("expected body to contain 'activityEvents' in JSON data")
	}
}

func TestPage_Project_ContainsBlockedCount(t *testing.T) {
	r := setupPageTest(fullPageTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/projects/test-project", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()

	// The JSON data should include blockedCount
	if !strings.Contains(body, "blockedCount") {
		t.Error("expected body to contain 'blockedCount' in JSON data")
	}
}

func TestPage_Project_ActivityEventsInJSON(t *testing.T) {
	r := setupPageTest(fullPageTestFS())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/projects/test-project", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()

	// Extract JSON from script tag
	startMarker := "application/json\">"
	startIdx := strings.Index(body, startMarker)
	if startIdx == -1 {
		t.Fatal("could not find JSON data block in response")
	}
	startIdx += len(startMarker)

	endIdx := strings.Index(body[startIdx:], "</script>")
	if endIdx == -1 {
		t.Fatal("could not find end of script tag")
	}

	jsonData := body[startIdx : startIdx+endIdx]
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &parsed); err != nil {
		t.Fatalf("JSON data is not valid: %v\nJSON: %s", err, jsonData)
	}

	// Verify activityEvents is present and is an array
	events, ok := parsed["activityEvents"]
	if !ok {
		t.Fatal("expected activityEvents key in JSON data")
	}
	eventsArr, ok := events.([]interface{})
	if !ok {
		t.Fatalf("expected activityEvents to be an array, got %T", events)
	}

	// feature-a has tasks: 1.1 (completed) and 1.2 (in_progress) -> 2 events
	// feature-b has tasks: 2.1 (pending) -> no event
	if len(eventsArr) != 2 {
		t.Errorf("expected 2 activity events, got %d", len(eventsArr))
	}

	// Verify blockedCount
	bc, ok := parsed["blockedCount"]
	if !ok {
		t.Fatal("expected blockedCount key in JSON data")
	}
	// No blocked tasks in the test data
	if bc != float64(0) {
		t.Errorf("expected blockedCount to be 0, got %v", bc)
	}
}

func TestPage_Project_ActivityEventsOrdering(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "test", Path: "/test"},
		},
	}

	// feature-a has older mtime, feature-b has newer mtime
	olderTime := time.Date(2026, 5, 6, 10, 0, 0, 0, time.UTC)
	newerTime := time.Date(2026, 5, 6, 14, 0, 0, 0, time.UTC)

	fsMap := fstest.MapFS{
		"test/docs/features/alpha/tasks/index.json": {
			Data: apiIndexJSON("alpha", "", "", "in-progress", map[string]interface{}{
				"1.1-t1": apiTaskEntry("1.1", "Task One", "P0", "completed", "all", "1.1-t1.md", nil, false),
			}),
			ModTime: olderTime,
		},
		"test/docs/features/beta/tasks/index.json": {
			Data: apiIndexJSON("beta", "", "", "in-progress", map[string]interface{}{
				"2.1-t2": apiTaskEntry("2.1", "Task Two", "P1", "blocked", "backend", "2.1-t2.md", nil, false),
			}),
			ModTime: newerTime,
		},
	}

	s := scanner.NewScannerWithFS(cfg, fsMap)
	_, _ = s.ScanAll()
	RegisterPages(r, s)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/projects/test", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()

	// Extract JSON
	startMarker := "application/json\">"
	startIdx := strings.Index(body, startMarker)
	if startIdx == -1 {
		t.Fatal("could not find JSON data block")
	}
	startIdx += len(startMarker)
	endIdx := strings.Index(body[startIdx:], "</script>")
	jsonData := body[startIdx : startIdx+endIdx]

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &parsed); err != nil {
		t.Fatalf("JSON parse error: %v", err)
	}

	events := parsed["activityEvents"].([]interface{})
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	// Beta event should be first (newer timestamp)
	first := events[0].(map[string]interface{})
	if first["feature"] != "beta" {
		t.Errorf("expected first event feature to be 'beta', got %v", first["feature"])
	}

	// Blocked count should be 1
	bc := parsed["blockedCount"]
	if bc != float64(1) {
		t.Errorf("expected blockedCount to be 1, got %v", bc)
	}
}

func TestPage_Project_ActivityEventsEmptyWhenNoNonPendingTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "test", Path: "/test"},
		},
	}

	fsMap := fstest.MapFS{
		"test/docs/features/feat/tasks/index.json": {
			Data: apiIndexJSON("feat", "", "", "planning", map[string]interface{}{
				"1.1-t1": apiTaskEntry("1.1", "Pending Task", "P0", "pending", "all", "1.1-t1.md", nil, false),
			}),
		},
	}

	s := scanner.NewScannerWithFS(cfg, fsMap)
	_, _ = s.ScanAll()
	RegisterPages(r, s)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/projects/test", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	startMarker := "application/json\">"
	startIdx := strings.Index(body, startMarker)
	if startIdx == -1 {
		t.Fatal("could not find JSON data block")
	}
	startIdx += len(startMarker)
	endIdx := strings.Index(body[startIdx:], "</script>")
	jsonData := body[startIdx : startIdx+endIdx]

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &parsed); err != nil {
		t.Fatalf("JSON parse error: %v", err)
	}

	events := parsed["activityEvents"].([]interface{})
	if len(events) != 0 {
		t.Errorf("expected 0 activity events for pending tasks, got %d", len(events))
	}
}

// ---- Activity event derivation unit tests ----

func TestDeriveActivityEvents_SortsByTimestampDesc(t *testing.T) {
	older := time.Date(2026, 5, 6, 10, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 5, 6, 14, 0, 0, 0, time.UTC)

	features := []model.FeatureData{
		{
			Slug:        "alpha",
			LastUpdated: older,
			Tasks: map[string]model.Task{
				"1.1-t": {ID: "1.1", Title: "Old Task", Status: "completed"},
			},
		},
		{
			Slug:        "beta",
			LastUpdated: newer,
			Tasks: map[string]model.Task{
				"2.1-t": {ID: "2.1", Title: "New Task", Status: "blocked"},
			},
		},
	}

	events := deriveActivityEvents(features)
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	// Newer event first
	if events[0].Feature != "beta" {
		t.Errorf("expected first event feature to be 'beta', got %s", events[0].Feature)
	}
	if events[1].Feature != "alpha" {
		t.Errorf("expected second event feature to be 'alpha', got %s", events[1].Feature)
	}
}

func TestDeriveActivityEvents_SortsByFeatureSlugForEqualTimestamps(t *testing.T) {
	ts := time.Date(2026, 5, 6, 14, 0, 0, 0, time.UTC)

	features := []model.FeatureData{
		{
			Slug:        "zeta",
			LastUpdated: ts,
			Tasks: map[string]model.Task{
				"1.1-t": {ID: "1.1", Title: "Zeta Task", Status: "completed"},
			},
		},
		{
			Slug:        "alpha",
			LastUpdated: ts,
			Tasks: map[string]model.Task{
				"2.1-t": {ID: "2.1", Title: "Alpha Task", Status: "completed"},
			},
		},
	}

	events := deriveActivityEvents(features)
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	// Alpha first (alphabetically) when timestamps are equal
	if events[0].Feature != "alpha" {
		t.Errorf("expected first event feature to be 'alpha', got %s", events[0].Feature)
	}
	if events[1].Feature != "zeta" {
		t.Errorf("expected second event feature to be 'zeta', got %s", events[1].Feature)
	}
}

func TestDeriveActivityEvents_LimitsTo50(t *testing.T) {
	ts := time.Date(2026, 5, 6, 14, 0, 0, 0, time.UTC)
	tasks := make(map[string]model.Task)
	for i := 0; i < 60; i++ {
		key := fmt.Sprintf("%d-t", i)
		tasks[key] = model.Task{ID: fmt.Sprintf("%d", i), Title: fmt.Sprintf("Task %d", i), Status: "completed"}
	}

	features := []model.FeatureData{
		{
			Slug:        "feat",
			LastUpdated: ts,
			Tasks:       tasks,
		},
	}

	events := deriveActivityEvents(features)
	if len(events) != 50 {
		t.Errorf("expected 50 events (limited), got %d", len(events))
	}
}

func TestDeriveActivityEvents_SkipsPendingTasks(t *testing.T) {
	ts := time.Date(2026, 5, 6, 14, 0, 0, 0, time.UTC)

	features := []model.FeatureData{
		{
			Slug:        "feat",
			LastUpdated: ts,
			Tasks: map[string]model.Task{
				"1.1-t": {ID: "1.1", Title: "Pending", Status: "pending"},
				"1.2-t": {ID: "1.2", Title: "Completed", Status: "completed"},
			},
		},
	}

	events := deriveActivityEvents(features)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].TaskID != "1.2" {
		t.Errorf("expected event for task 1.2, got %s", events[0].TaskID)
	}
}

func TestDeriveActivityEvents_MapsStatusToEventType(t *testing.T) {
	ts := time.Date(2026, 5, 6, 14, 0, 0, 0, time.UTC)

	features := []model.FeatureData{
		{
			Slug:        "feat",
			LastUpdated: ts,
			Tasks: map[string]model.Task{
				"1-t": {ID: "1", Title: "Claimed", Status: "in_progress"},
				"2-t": {ID: "2", Title: "Done", Status: "completed"},
				"3-t": {ID: "3", Title: "Stuck", Status: "blocked"},
				"4-t": {ID: "4", Title: "Skipped", Status: "skipped"},
			},
		},
	}

	events := deriveActivityEvents(features)
	if len(events) != 4 {
		t.Fatalf("expected 4 events, got %d", len(events))
	}

	expected := map[string]string{
		"1": "claimed",
		"2": "completed",
		"3": "blocked",
		"4": "skipped",
	}

	for _, ev := range events {
		if ev.EventType != expected[ev.TaskID] {
			t.Errorf("task %s: expected eventType %s, got %s", ev.TaskID, expected[ev.TaskID], ev.EventType)
		}
	}
}

func TestCountBlockedTasks(t *testing.T) {
	features := []model.FeatureData{
		{
			Slug: "a",
			Tasks: map[string]model.Task{
				"1-t": {ID: "1", Status: "blocked"},
				"2-t": {ID: "2", Status: "completed"},
			},
		},
		{
			Slug: "b",
			Tasks: map[string]model.Task{
				"3-t": {ID: "3", Status: "blocked"},
				"4-t": {ID: "4", Status: "in_progress"},
			},
		},
	}

	count := countBlockedTasks(features)
	if count != 2 {
		t.Errorf("expected 2 blocked tasks, got %d", count)
	}
}

func TestCountBlockedTasks_None(t *testing.T) {
	features := []model.FeatureData{
		{
			Slug: "a",
			Tasks: map[string]model.Task{
				"1-t": {ID: "1", Status: "completed"},
			},
		},
	}

	count := countBlockedTasks(features)
	if count != 0 {
		t.Errorf("expected 0 blocked tasks, got %d", count)
	}
}
