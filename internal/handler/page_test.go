package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/panda/agent-task-center/internal/config"
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

	// Should contain project link
	if !strings.Contains(body, "/projects/test-project") {
		t.Error("expected body to contain project link")
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
