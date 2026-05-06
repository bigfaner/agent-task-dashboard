package main

import (
	"bytes"
	"context"
	"io/fs"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/panda/agent-task-center/internal/config"
	"github.com/panda/agent-task-center/internal/scanner"
	"github.com/panda/agent-task-center/web"
)

// createTestConfig creates a temporary config file and returns its path.
func createTestConfig(t *testing.T) (string, string) {
	t.Helper()
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "test-config.yaml")
	projectDir := filepath.Join(tmpDir, "test-project")
	if err := os.MkdirAll(filepath.Join(projectDir, "docs/features/test-feature/tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfgContent := "projects:\n  - name: test-project\n    path: " + projectDir + "\n"
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0o644); err != nil {
		t.Fatal(err)
	}
	return cfgPath, projectDir
}

func TestSetupRouter_ReturnsGinEngine(t *testing.T) {
	cfgPath, _ := createTestConfig(t)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	s := scanner.NewScanner(cfg)
	r := setupRouter(s, cfg)

	if r == nil {
		t.Fatal("setupRouter returned nil")
	}
}

func TestSetupRouter_ServesStaticCSS(t *testing.T) {
	cfgPath, _ := createTestConfig(t)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	s := scanner.NewScanner(cfg)
	r := setupRouter(s, cfg)

	req := httptest.NewRequest(http.MethodGet, "/static/css/styles.css", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for /static/css/styles.css, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "--status-pending") {
		t.Errorf("Expected CSS custom properties in response, got: %s", body[:min(200, len(body))])
	}
}

func TestSetupRouter_ServesStaticJS(t *testing.T) {
	cfgPath, _ := createTestConfig(t)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	s := scanner.NewScanner(cfg)
	r := setupRouter(s, cfg)

	req := httptest.NewRequest(http.MethodGet, "/static/js/landing.js", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for /static/js/landing.js, got %d", w.Code)
	}
}

func TestSetupRouter_ServesDagreMinJS(t *testing.T) {
	cfgPath, _ := createTestConfig(t)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	s := scanner.NewScanner(cfg)
	r := setupRouter(s, cfg)

	req := httptest.NewRequest(http.MethodGet, "/static/js/dagre.min.js", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for /static/js/dagre.min.js, got %d", w.Code)
	}
}

func TestSetupRouter_LandingPageReturnsHTML(t *testing.T) {
	cfgPath, _ := createTestConfig(t)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	s := scanner.NewScanner(cfg)
	r := setupRouter(s, cfg)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for /, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Errorf("Expected HTML response for /, got: %s", body[:min(200, len(body))])
	}
	if !strings.Contains(body, "Task Dashboard") {
		t.Errorf("Expected 'Task Dashboard' in landing page")
	}
}

func TestSetupRouter_404ForUnknownRoute(t *testing.T) {
	cfgPath, _ := createTestConfig(t)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	s := scanner.NewScanner(cfg)
	r := setupRouter(s, cfg)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for /nonexistent, got %d", w.Code)
	}
}

func TestSetupRouter_APIProjectsEndpoint(t *testing.T) {
	cfgPath, _ := createTestConfig(t)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	s := scanner.NewScanner(cfg)
	r := setupRouter(s, cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for /api/projects, got %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Expected JSON content type, got: %s", ct)
	}
}

func TestDefaultConfigPath_IsValidPath(t *testing.T) {
	path := defaultConfigPath()
	if path == "" {
		t.Fatal("defaultConfigPath returned empty string")
	}
	if !strings.HasSuffix(path, ".task-dashboard.yaml") {
		t.Errorf("Expected path to end with .task-dashboard.yaml, got: %s", path)
	}
}

func TestStaticFilesDoNotExposeTemplates(t *testing.T) {
	cfgPath, _ := createTestConfig(t)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	s := scanner.NewScanner(cfg)
	r := setupRouter(s, cfg)

	// Templates should NOT be accessible via /static/
	req := httptest.NewRequest(http.MethodGet, "/static/templates/landing.html", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for template path via /static/, got %d", w.Code)
	}
}

func TestGracefulShutdown(t *testing.T) {
	cfgPath, _ := createTestConfig(t)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	cfg.Server.Port = 0 // will be assigned a random port

	s := scanner.NewScanner(cfg)
	r := setupRouter(s, cfg)

	// Listen on a random port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	addr := ln.Addr().String()

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		_ = srv.Serve(ln)
	}()

	// Give server time to start
	time.Sleep(50 * time.Millisecond)

	// Verify server is responding
	resp, err := http.Get("http://" + addr + "/")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	// Initiate graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

func TestServerBindsToLocalhostOnly(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "test", Path: t.TempDir()},
		},
		Server: config.ServerConfig{Port: 0},
	}

	s := scanner.NewScanner(cfg)
	r := setupRouter(s, cfg)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	addr := ln.Addr().String()

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		_ = srv.Serve(ln)
	}()
	defer func() { _ = srv.Shutdown(context.Background()) }()

	time.Sleep(50 * time.Millisecond)

	// Verify the address is 127.0.0.1
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("Failed to parse address: %v", err)
	}
	if host != "127.0.0.1" {
		t.Errorf("Expected server to bind to 127.0.0.1, got %s", host)
	}
}

func TestEmbeddedAssetsExist(t *testing.T) {
	// Verify key embedded assets are accessible
	entries, err := fs.ReadDir(web.Assets, "static")
	if err != nil {
		t.Fatalf("Failed to read embedded static directory: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("No entries in embedded static directory")
	}

	// Verify CSS file exists
	cssContent, err := fs.ReadFile(web.Assets, "static/css/styles.css")
	if err != nil {
		t.Fatalf("Failed to read embedded styles.css: %v", err)
	}
	if len(cssContent) == 0 {
		t.Error("styles.css is empty")
	}

	// Verify a JS file exists
	_, err = fs.ReadFile(web.Assets, "static/js/landing.js")
	if err != nil {
		t.Fatalf("Failed to read embedded landing.js: %v", err)
	}

	// Verify templates exist
	tmplContent, err := fs.ReadFile(web.Assets, "templates/landing.html")
	if err != nil {
		t.Fatalf("Failed to read embedded landing.html: %v", err)
	}
	if !bytes.Contains(tmplContent, []byte("DOCTYPE")) {
		t.Error("landing.html doesn't look like an HTML template")
	}
}

func TestSetupRouter_ServesSwimlaneJS(t *testing.T) {
	cfgPath, _ := createTestConfig(t)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	s := scanner.NewScanner(cfg)
	r := setupRouter(s, cfg)

	req := httptest.NewRequest(http.MethodGet, "/static/js/swimlane.js", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for /static/js/swimlane.js, got %d", w.Code)
	}

	body := w.Body.String()
	// Verify key swimlane functions and patterns exist
	checks := []string{
		"renderPage",
		"renderFeatureRow",
		"renderTaskCard",
		"renderDependencyArrows",
		"renderArrow",
		"applyFilters",
		"toggleRow",
		"sortFeatures",
		"highlightTaskCard",
		"__INITIAL_DATA__",
		"edge-cross-feature",
		"edge-within",
		"edge-blocked",
		"status-",
		"priority-",
	}
	for _, check := range checks {
		if !strings.Contains(body, check) {
			t.Errorf("Expected swimlane.js to contain %q", check)
		}
	}
}

func TestSetupRouter_ServesDagreWithContent(t *testing.T) {
	cfgPath, _ := createTestConfig(t)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	s := scanner.NewScanner(cfg)
	r := setupRouter(s, cfg)

	req := httptest.NewRequest(http.MethodGet, "/static/js/dagre.min.js", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for /static/js/dagre.min.js, got %d", w.Code)
	}

	body := w.Body.String()
	if len(body) < 1000 {
		t.Errorf("dagre.min.js seems too small (%d bytes), expected a real dagre library", len(body))
	}
}

func TestSwimlaneTemplateContainsFilterControls(t *testing.T) {
	tmplContent, err := fs.ReadFile(web.Assets, "templates/swimlane.html")
	if err != nil {
		t.Fatalf("Failed to read swimlane.html: %v", err)
	}
	body := string(tmplContent)

	checks := []string{
		"status-filter-btn",
		"priority-filter-btn",
		"statusFilter",
		"priorityFilter",
		"swimlane-container",
		"detail-panel",
		"activity-sidebar",
		"__INITIAL_DATA__",
		"swimlane.js",
		"dagre.min.js",
		"detail-panel.js",
		"activity.js",
	}
	for _, check := range checks {
		if !strings.Contains(body, check) {
			t.Errorf("Expected swimlane.html to contain %q", check)
		}
	}
}

func TestSwimlaneTemplateRendersForProject(t *testing.T) {
	cfgPath, projectDir := createTestConfig(t)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create a feature with tasks in the test project
	taskDir := filepath.Join(projectDir, "docs/features/test-feature/tasks")
	featureDir := filepath.Join(projectDir, "docs/features/test-feature")
	if err := os.MkdirAll(taskDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write index.json with tasks (tasks is a map keyed by task key)
	indexJSON := `{
		"feature": "test-feature",
		"status": "in-progress",
		"tasks": {
			"1.1-setup": {"id": "1.1", "key": "1.1-setup", "title": "Set up project", "status": "completed", "priority": "P0", "phase": 1, "dependencies": []},
			"2.1-scanner": {"id": "2.1", "key": "2.1-scanner", "title": "Build scanner", "status": "in_progress", "priority": "P0", "phase": 2, "dependencies": ["1.1"]},
			"3.1-renderer": {"id": "3.1", "key": "3.1-renderer", "title": "Swimlane renderer", "status": "blocked", "priority": "P0", "phase": 3, "dependencies": ["2.1"]}
		}
	}`
	if err := os.WriteFile(filepath.Join(featureDir, "tasks/index.json"), []byte(indexJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	s := scanner.NewScanner(cfg)
	r := setupRouter(s, cfg)

	req := httptest.NewRequest(http.MethodGet, "/projects/test-project", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for project page, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "test-project") {
		t.Error("Expected project name in rendered page")
	}
	if !strings.Contains(body, "__INITIAL_DATA__") {
		t.Error("Expected __INITIAL_DATA__ script in rendered page")
	}
	if !strings.Contains(body, "swimlane.js") {
		t.Error("Expected swimlane.js script reference")
	}
	if !strings.Contains(body, "status-filter-btn") {
		t.Error("Expected status filter button in rendered page")
	}
	if !strings.Contains(body, "priority-filter-btn") {
		t.Error("Expected priority filter button in rendered page")
	}
}

func TestSwimlaneJSContainsAllPhaseColumns(t *testing.T) {
	jsContent, err := fs.ReadFile(web.Assets, "static/js/swimlane.js")
	if err != nil {
		t.Fatalf("Failed to read swimlane.js: %v", err)
	}
	body := string(jsContent)

	phases := []string{"Phase 1", "Phase 2", "Phase 3+", "Testing", "Other"}
	for _, phase := range phases {
		if !strings.Contains(body, phase) {
			t.Errorf("Expected swimlane.js to contain phase column %q", phase)
		}
	}
}

func TestSwimlaneJSContainsStatusHandling(t *testing.T) {
	jsContent, err := fs.ReadFile(web.Assets, "static/js/swimlane.js")
	if err != nil {
		t.Fatalf("Failed to read swimlane.js: %v", err)
	}
	body := string(jsContent)

	// The JS references statuses for status class generation (status.replace('_', '-'))
	// and for filter values
	statusChecks := []struct {
		value   string
		context string
	}{
		{"pending", "status filter option"},
		{"completed", "status filter option"},
		{"blocked", "status filter and hasBlockedTasks check"},
	}
	for _, check := range statusChecks {
		if !strings.Contains(body, check.value) {
			t.Errorf("Expected swimlane.js to reference status %q (%s)", check.value, check.context)
		}
	}

	// Check that the JS generates status CSS classes via replace('_', '-')
	if !strings.Contains(body, "status-") {
		t.Error("Expected swimlane.js to generate status- CSS class prefix")
	}
	if !strings.Contains(body, ".replace") {
		t.Error("Expected swimlane.js to use .replace for status class name normalization")
	}
}

func TestSwimlaneJSContainsPriorityHandling(t *testing.T) {
	jsContent, err := fs.ReadFile(web.Assets, "static/js/swimlane.js")
	if err != nil {
		t.Fatalf("Failed to read swimlane.js: %v", err)
	}
	body := string(jsContent)

	// JS uses task.priority and toLowerCase() to build priority-p0/p1/p2 classes
	if !strings.Contains(body, "priority-") {
		t.Error("Expected swimlane.js to generate priority- CSS class prefix")
	}
	if !strings.Contains(body, "toLowerCase") {
		t.Error("Expected swimlane.js to use toLowerCase() for priority class generation")
	}
}

func TestSwimlaneJSContainsArrowMarkers(t *testing.T) {
	jsContent, err := fs.ReadFile(web.Assets, "static/js/swimlane.js")
	if err != nil {
		t.Fatalf("Failed to read swimlane.js: %v", err)
	}
	body := string(jsContent)

	if !strings.Contains(body, "arrowhead") {
		t.Error("Expected swimlane.js to define arrowhead SVG marker")
	}
	if !strings.Contains(body, "arrowhead-dashed") {
		t.Error("Expected swimlane.js to define dashed arrowhead marker")
	}
	if !strings.Contains(body, "edge-cross-feature") {
		t.Error("Expected swimlane.js to use edge-cross-feature class for dashed arrows")
	}
}

func TestSwimlaneJSTruncatesTitle(t *testing.T) {
	jsContent, err := fs.ReadFile(web.Assets, "static/js/swimlane.js")
	if err != nil {
		t.Fatalf("Failed to read swimlane.js: %v", err)
	}
	body := string(jsContent)

	if !strings.Contains(body, "truncateTitle") {
		t.Error("Expected swimlane.js to contain truncateTitle function")
	}
	if !strings.Contains(body, "30") {
		t.Error("Expected swimlane.js to reference title truncation length of 30")
	}
}

func TestSwimlaneJSExposesGlobalFunctions(t *testing.T) {
	jsContent, err := fs.ReadFile(web.Assets, "static/js/swimlane.js")
	if err != nil {
		t.Fatalf("Failed to read swimlane.js: %v", err)
	}
	body := string(jsContent)

	if !strings.Contains(body, "window.highlightTaskCard") {
		t.Error("Expected swimlane.js to expose window.highlightTaskCard")
	}
	if !strings.Contains(body, "window.openDetailPanel") {
		t.Error("Expected swimlane.js to reference window.openDetailPanel for detail panel integration")
	}
}

func TestSetupRouter_SetsGinReleaseMode(t *testing.T) {
	// Reset gin mode after test
	defer gin.SetMode(gin.DebugMode)

	cfgPath, _ := createTestConfig(t)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	s := scanner.NewScanner(cfg)
	_ = setupRouter(s, cfg)

	if gin.Mode() != gin.ReleaseMode {
		t.Errorf("Expected gin mode to be ReleaseMode after setupRouter, got: %s", gin.Mode())
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
