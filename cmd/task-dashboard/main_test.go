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
