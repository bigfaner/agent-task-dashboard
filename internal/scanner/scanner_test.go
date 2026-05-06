package scanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"testing/fstest"
	"time"

	"github.com/panda/agent-task-center/internal/config"
	"github.com/panda/agent-task-center/internal/model"
)

// testProjectPath is the consistent project path used in MapFS unit tests.
// MapFS entries must be prefixed with "project/docs/features/..." to match.
const testProjectPath = "/project"

// indexJSON is a helper to build an index.json bytes from the given data.
func indexJSON(feature, prd, design, status string, tasks map[string]interface{}) []byte {
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

// taskEntry builds a single task map for index.json.
func taskEntry(id, title, priority, status, scope, file string, deps []string, breaking bool) map[string]interface{} {
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

// featurePath returns the MapFS path for a feature's index.json under the test project.
func featurePath(slug string) string {
	return "project/docs/features/" + slug + "/tasks/index.json"
}

// ---- Unit tests using fstest.MapFS ----

func TestNewScanner(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "Test Project", Path: "/tmp/test"},
		},
	}
	s := NewScanner(cfg)
	if s == nil {
		t.Fatal("NewScanner returned nil")
	}
	if s.cache == nil {
		t.Error("cache map not initialized")
	}
}

func TestScanAll_ValidProject(t *testing.T) {
	tasks := map[string]interface{}{
		"1.1-interfaces": taskEntry("1.1", "Define interfaces", "P0", "completed", "all", "1.1-interfaces.md", nil, false),
		"1.2-models":     taskEntry("1.2", "Define models", "P0", "in_progress", "all", "1.2-models.md", []string{"1.1"}, false),
	}

	fs := fstest.MapFS{
		featurePath("my-feature"): &fstest.MapFile{
			Data: indexJSON("my-feature", "prd.md", "design.md", "in-progress", tasks),
		},
	}

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "My Project", Path: testProjectPath},
		},
	}

	s := NewScanner(cfg)
	s.fs = fs

	result, err := s.ScanAll()
	if err != nil {
		t.Fatalf("ScanAll() returned error: %v", err)
	}

	projID := "my project"
	pd, ok := result[projID]
	if !ok {
		t.Fatalf("expected project ID %q in result, got keys: %v", projID, keys(result))
	}

	if pd.Name != "My Project" {
		t.Errorf("Name = %q, want %q", pd.Name, "My Project")
	}
	if len(pd.Features) != 1 {
		t.Fatalf("expected 1 feature, got %d", len(pd.Features))
	}
	feat := pd.Features[0]
	if feat.Slug != "my-feature" {
		t.Errorf("Slug = %q, want %q", feat.Slug, "my-feature")
	}
	if feat.TotalTasks != 2 {
		t.Errorf("TotalTasks = %d, want 2", feat.TotalTasks)
	}
	if feat.CompletedTasks != 1 {
		t.Errorf("CompletedTasks = %d, want 1", feat.CompletedTasks)
	}
	if feat.CompletionPct != 50.0 {
		t.Errorf("CompletionPct = %f, want 50.0", feat.CompletionPct)
	}
	if len(feat.Tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(feat.Tasks))
	}
	if len(feat.Phases) == 0 {
		t.Error("expected at least 1 phase")
	}
	if pd.TotalTasks != 2 {
		t.Errorf("project TotalTasks = %d, want 2", pd.TotalTasks)
	}
	if pd.CompletedTasks != 1 {
		t.Errorf("project CompletedTasks = %d, want 1", pd.CompletedTasks)
	}
}

func TestScanAll_InvalidPath(t *testing.T) {
	fs := fstest.MapFS{}

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "Missing Project", Path: "/nonexistent/path"},
		},
	}

	s := NewScanner(cfg)
	s.fs = fs

	result, err := s.ScanAll()
	if err != nil {
		t.Fatalf("ScanAll() returned error: %v", err)
	}

	pd := result["missing project"]
	if len(pd.Warnings) == 0 {
		t.Error("expected warning for invalid path")
	}
	if len(pd.Features) != 0 {
		t.Errorf("expected 0 features for invalid path, got %d", len(pd.Features))
	}
}

func TestScanAll_MalformedJSON(t *testing.T) {
	fs := fstest.MapFS{
		featurePath("bad-feature"): &fstest.MapFile{
			Data: []byte(`{invalid json}`),
		},
	}

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "Bad Project", Path: testProjectPath},
		},
	}

	s := NewScanner(cfg)
	s.fs = fs

	result, err := s.ScanAll()
	if err != nil {
		t.Fatalf("ScanAll() returned error: %v", err)
	}

	pd := result["bad project"]
	if len(pd.Features) != 0 {
		t.Errorf("expected 0 features with malformed JSON, got %d", len(pd.Features))
	}
}

func TestScanAll_MissingFeaturesDir(t *testing.T) {
	fs := fstest.MapFS{
		"project/other-file.txt": &fstest.MapFile{Data: []byte("hello")},
	}

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "Empty Project", Path: testProjectPath},
		},
	}

	s := NewScanner(cfg)
	s.fs = fs

	result, err := s.ScanAll()
	if err != nil {
		t.Fatalf("ScanAll() returned error: %v", err)
	}

	pd := result["empty project"]
	if len(pd.Features) != 0 {
		t.Errorf("expected 0 features, got %d", len(pd.Features))
	}
}

func TestScanProject_Found(t *testing.T) {
	tasks := map[string]interface{}{
		"1.1-interfaces": taskEntry("1.1", "Define interfaces", "P0", "pending", "all", "1.1-interfaces.md", nil, false),
	}

	fs := fstest.MapFS{
		featurePath("feat-a"): &fstest.MapFile{
			Data: indexJSON("feat-a", "prd.md", "design.md", "planning", tasks),
		},
	}

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "Test Project", Path: testProjectPath},
		},
	}

	s := NewScanner(cfg)
	s.fs = fs

	pd, err := s.ScanProject("test project")
	if err != nil {
		t.Fatalf("ScanProject() returned error: %v", err)
	}
	if pd.Name != "Test Project" {
		t.Errorf("Name = %q, want %q", pd.Name, "Test Project")
	}
	if len(pd.Features) != 1 {
		t.Errorf("expected 1 feature, got %d", len(pd.Features))
	}
}

func TestScanProject_NotFound(t *testing.T) {
	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "Test Project", Path: "/tmp/test"},
		},
	}

	s := NewScanner(cfg)

	_, err := s.ScanProject("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown project ID")
	}
	if _, ok := err.(model.ErrProjectNotFound); !ok {
		t.Errorf("error type = %T, want model.ErrProjectNotFound", err)
	}
}

func TestInvalidate(t *testing.T) {
	tasks := map[string]interface{}{
		"1.1-interfaces": taskEntry("1.1", "Define interfaces", "P0", "pending", "all", "1.1-interfaces.md", nil, false),
	}

	fs := fstest.MapFS{
		featurePath("feat"): &fstest.MapFile{
			Data: indexJSON("feat", "prd.md", "design.md", "planning", tasks),
		},
	}

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "Proj", Path: testProjectPath},
		},
	}

	s := NewScanner(cfg)
	s.fs = fs

	// First scan populates cache
	_, _ = s.ScanAll()
	if len(s.cache) != 1 {
		t.Errorf("expected 1 cached entry, got %d", len(s.cache))
	}

	// Invalidate clears cache
	s.Invalidate()
	if len(s.cache) != 0 {
		t.Errorf("expected 0 cached entries after Invalidate, got %d", len(s.cache))
	}
}

func TestWildcardExpansion(t *testing.T) {
	tasks := map[string]interface{}{
		"1.1-interfaces": taskEntry("1.1", "Interfaces", "P0", "completed", "all", "1.1.md", nil, false),
		"1.2-models":     taskEntry("1.2", "Models", "P0", "completed", "all", "1.2.md", nil, false),
		"1.3-scanner":    taskEntry("1.3", "Scanner", "P0", "pending", "all", "1.3.md", []string{"1.x"}, false),
		"2.1-handler":    taskEntry("2.1", "Handler", "P1", "pending", "all", "2.1.md", nil, false),
	}

	fs := fstest.MapFS{
		featurePath("feat"): &fstest.MapFile{
			Data: indexJSON("feat", "prd.md", "design.md", "in-progress", tasks),
		},
	}

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "Proj", Path: testProjectPath},
		},
	}

	s := NewScanner(cfg)
	s.fs = fs

	result, err := s.ScanAll()
	if err != nil {
		t.Fatalf("ScanAll() returned error: %v", err)
	}

	pd := result["proj"]
	if len(pd.Features) == 0 {
		t.Fatal("expected at least 1 feature")
	}
	task13 := pd.Features[0].Tasks["1.3-scanner"]
	if task13.Dependencies == nil {
		t.Fatal("expected dependencies for task 1.3")
	}

	// "1.x" should expand to "1.1", "1.2", "1.3" (all phase 1 tasks)
	sort.Strings(task13.Dependencies)
	expected := []string{"1.1", "1.2", "1.3"}
	if len(task13.Dependencies) != len(expected) {
		t.Fatalf("expected %d dependencies, got %d: %v", len(expected), len(task13.Dependencies), task13.Dependencies)
	}
	for i, dep := range task13.Dependencies {
		if dep != expected[i] {
			t.Errorf("dependency[%d] = %q, want %q", i, dep, expected[i])
		}
	}
}

func TestSortFeatures(t *testing.T) {
	features := []model.FeatureData{
		{
			Slug:            "alpha",
			CompletionPct:   80.0,
			HasBlockedTasks: false,
		},
		{
			Slug:            "beta",
			CompletionPct:   30.0,
			HasBlockedTasks: true,
		},
		{
			Slug:            "gamma",
			CompletionPct:   30.0,
			HasBlockedTasks: false,
		},
		{
			Slug:            "delta",
			CompletionPct:   50.0,
			HasBlockedTasks: true,
		},
		{
			Slug:            "epsilon",
			CompletionPct:   30.0,
			HasBlockedTasks: false,
		},
	}

	SortFeatures(features)

	expected := []struct {
		slug    string
		blocked bool
		pct     float64
	}{
		{"beta", true, 30.0},     // blocked, 30%
		{"delta", true, 50.0},    // blocked, 50%
		{"epsilon", false, 30.0}, // not blocked, 30%, alphabetical first
		{"gamma", false, 30.0},   // not blocked, 30%, alphabetical second
		{"alpha", false, 80.0},   // not blocked, 80%
	}

	for i, exp := range expected {
		if features[i].Slug != exp.slug {
			t.Errorf("position %d: slug = %q, want %q", i, features[i].Slug, exp.slug)
		}
		if features[i].HasBlockedTasks != exp.blocked {
			t.Errorf("position %d: HasBlockedTasks = %v, want %v", i, features[i].HasBlockedTasks, exp.blocked)
		}
		if features[i].CompletionPct != exp.pct {
			t.Errorf("position %d: CompletionPct = %f, want %f", i, features[i].CompletionPct, exp.pct)
		}
	}
}

func TestPhaseDerivation(t *testing.T) {
	tasks := map[string]interface{}{
		"1.1-interfaces": taskEntry("1.1", "Interfaces", "P0", "completed", "all", "1.1.md", nil, false),
		"1.2-models":     taskEntry("1.2", "Models", "P0", "completed", "all", "1.2.md", nil, false),
		"2.1-handler":    taskEntry("2.1", "Handler", "P1", "pending", "all", "2.1.md", nil, false),
		"3.1-testing":    taskEntry("3.1", "Testing", "P2", "pending", "all", "3.1.md", nil, false),
	}

	fs := fstest.MapFS{
		featurePath("feat"): &fstest.MapFile{
			Data: indexJSON("feat", "prd.md", "design.md", "in-progress", tasks),
		},
	}

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "Proj", Path: testProjectPath},
		},
	}

	s := NewScanner(cfg)
	s.fs = fs

	result, _ := s.ScanAll()
	phases := result["proj"].Features[0].Phases

	if len(phases) != 3 {
		t.Fatalf("expected 3 phases, got %d", len(phases))
	}

	if phases[0].Number != 1 {
		t.Errorf("phase[0].Number = %d, want 1", phases[0].Number)
	}
	if len(phases[0].TaskKeys) != 2 {
		t.Errorf("phase[0] TaskKeys = %d, want 2", len(phases[0].TaskKeys))
	}
	if phases[1].Number != 2 {
		t.Errorf("phase[1].Number = %d, want 2", phases[1].Number)
	}
	if phases[2].Number != 3 {
		t.Errorf("phase[2].Number = %d, want 3", phases[2].Number)
	}
}

func TestHealthStatus_Active(t *testing.T) {
	tasks := map[string]interface{}{
		"1.1-interfaces": taskEntry("1.1", "Interfaces", "P0", "in_progress", "all", "1.1.md", nil, false),
	}

	fs := fstest.MapFS{
		featurePath("feat"): &fstest.MapFile{
			Data:    indexJSON("feat", "prd.md", "design.md", "in-progress", tasks),
			ModTime: time.Now(),
		},
	}

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "Proj", Path: testProjectPath},
		},
	}

	s := NewScanner(cfg)
	s.fs = fs

	result, _ := s.ScanAll()
	pd := result["proj"]
	if pd.HealthStatus != "active" {
		t.Errorf("HealthStatus = %q, want %q", pd.HealthStatus, "active")
	}
}

func TestHealthStatus_Complete(t *testing.T) {
	tasks := map[string]interface{}{
		"1.1-interfaces": taskEntry("1.1", "Interfaces", "P0", "completed", "all", "1.1.md", nil, false),
	}

	fs := fstest.MapFS{
		featurePath("feat"): &fstest.MapFile{
			Data:    indexJSON("feat", "prd.md", "design.md", "completed", tasks),
			ModTime: time.Now(),
		},
	}

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "Proj", Path: testProjectPath},
		},
	}

	s := NewScanner(cfg)
	s.fs = fs

	result, _ := s.ScanAll()
	pd := result["proj"]
	if pd.HealthStatus != "complete" {
		t.Errorf("HealthStatus = %q, want %q", pd.HealthStatus, "complete")
	}
}

func TestHealthStatus_Stale(t *testing.T) {
	tasks := map[string]interface{}{
		"1.1-interfaces": taskEntry("1.1", "Interfaces", "P0", "pending", "all", "1.1.md", nil, false),
	}

	staleTime := time.Now().Add(-8 * 24 * time.Hour)

	fs := fstest.MapFS{
		featurePath("feat"): &fstest.MapFile{
			Data:    indexJSON("feat", "prd.md", "design.md", "planning", tasks),
			ModTime: staleTime,
		},
	}

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "Proj", Path: testProjectPath},
		},
	}

	s := NewScanner(cfg)
	s.fs = fs

	result, _ := s.ScanAll()
	pd := result["proj"]
	if pd.HealthStatus != "stale" {
		t.Errorf("HealthStatus = %q, want %q", pd.HealthStatus, "stale")
	}
}

func TestScanAll_MultipleProjects(t *testing.T) {
	tasksA := map[string]interface{}{
		"1.1-task": taskEntry("1.1", "Task A", "P0", "completed", "all", "1.1.md", nil, false),
	}
	tasksB := map[string]interface{}{
		"2.1-task": taskEntry("2.1", "Task B", "P0", "pending", "all", "2.1.md", nil, false),
	}

	fs := fstest.MapFS{
		"project-a/docs/features/feat-a/tasks/index.json": &fstest.MapFile{
			Data: indexJSON("feat-a", "prd.md", "design.md", "in-progress", tasksA),
		},
		"project-b/docs/features/feat-b/tasks/index.json": &fstest.MapFile{
			Data: indexJSON("feat-b", "prd.md", "design.md", "planning", tasksB),
		},
	}

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "Project A", Path: "/project-a"},
			{Name: "Project B", Path: "/project-b"},
		},
	}

	s := NewScanner(cfg)
	s.fs = fs

	result, err := s.ScanAll()
	if err != nil {
		t.Fatalf("ScanAll() returned error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(result))
	}
	if result["project a"].Name != "Project A" {
		t.Errorf("project a Name = %q", result["project a"].Name)
	}
	if result["project b"].Name != "Project B" {
		t.Errorf("project b Name = %q", result["project b"].Name)
	}
}

func TestScanAll_MultipleFeatures(t *testing.T) {
	tasksA := map[string]interface{}{
		"1.1-task": taskEntry("1.1", "Task 1", "P0", "completed", "all", "1.1.md", nil, false),
	}
	tasksB := map[string]interface{}{
		"1.2-task": taskEntry("1.2", "Task 2", "P0", "blocked", "all", "1.2.md", nil, false),
	}

	fs := fstest.MapFS{
		featurePath("feat-a"): &fstest.MapFile{
			Data: indexJSON("feat-a", "prd.md", "design.md", "completed", tasksA),
		},
		featurePath("feat-b"): &fstest.MapFile{
			Data: indexJSON("feat-b", "prd.md", "design.md", "in-progress", tasksB),
		},
	}

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "Proj", Path: testProjectPath},
		},
	}

	s := NewScanner(cfg)
	s.fs = fs

	result, _ := s.ScanAll()
	pd := result["proj"]
	if len(pd.Features) != 2 {
		t.Fatalf("expected 2 features, got %d", len(pd.Features))
	}

	// Features should be sorted: blocked first
	if !pd.Features[0].HasBlockedTasks {
		t.Errorf("first feature should have blocked tasks, got slug=%q", pd.Features[0].Slug)
	}
}

func TestFeatureWithEmptyTasks(t *testing.T) {
	fs := fstest.MapFS{
		featurePath("empty-feat"): &fstest.MapFile{
			Data: indexJSON("empty-feat", "prd.md", "design.md", "planning", map[string]interface{}{}),
		},
	}

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "Proj", Path: testProjectPath},
		},
	}

	s := NewScanner(cfg)
	s.fs = fs

	result, _ := s.ScanAll()
	pd := result["proj"]
	if len(pd.Features) != 1 {
		t.Fatalf("expected 1 feature, got %d", len(pd.Features))
	}
	feat := pd.Features[0]
	if feat.TotalTasks != 0 {
		t.Errorf("TotalTasks = %d, want 0", feat.TotalTasks)
	}
	if feat.CompletionPct != 0 {
		t.Errorf("CompletionPct = %f, want 0", feat.CompletionPct)
	}
}

// ---- Integration tests using temp directory ----

func TestIntegration_ScanRealDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	featDir := filepath.Join(tmpDir, "docs", "features", "real-feature", "tasks")
	if err := os.MkdirAll(featDir, 0755); err != nil {
		t.Fatalf("failed to create feature dir: %v", err)
	}

	idxContent := indexJSON("real-feature", "prd/spec.md", "design/tech-design.md", "in-progress", map[string]interface{}{
		"1.1-setup":    taskEntry("1.1", "Setup project", "P0", "completed", "all", "1.1-setup.md", nil, false),
		"1.2-scanner":  taskEntry("1.2", "Build scanner", "P0", "in_progress", "backend", "1.2-scanner.md", []string{"1.1"}, false),
		"2.1-handlers": taskEntry("2.1", "HTTP handlers", "P1", "pending", "backend", "2.1-handlers.md", []string{"1.2"}, true),
	})

	idxPath := filepath.Join(featDir, "index.json")
	if err := os.WriteFile(idxPath, idxContent, 0644); err != nil {
		t.Fatalf("failed to write index.json: %v", err)
	}

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "Real Project", Path: tmpDir},
		},
	}

	s := NewScanner(cfg)

	result, err := s.ScanAll()
	if err != nil {
		t.Fatalf("ScanAll() returned error: %v", err)
	}

	pd := result["real project"]
	if pd.ID != "real project" {
		t.Errorf("ID = %q, want %q", pd.ID, "real project")
	}
	if len(pd.Features) != 1 {
		t.Fatalf("expected 1 feature, got %d", len(pd.Features))
	}

	feat := pd.Features[0]
	if feat.Slug != "real-feature" {
		t.Errorf("Slug = %q, want %q", feat.Slug, "real-feature")
	}
	if feat.TotalTasks != 3 {
		t.Errorf("TotalTasks = %d, want 3", feat.TotalTasks)
	}
	if feat.CompletedTasks != 1 {
		t.Errorf("CompletedTasks = %d, want 1", feat.CompletedTasks)
	}
	expectedPct := (1.0 / 3.0) * 100
	if feat.CompletionPct < expectedPct-0.01 || feat.CompletionPct > expectedPct+0.01 {
		t.Errorf("CompletionPct = %f, want ~%f", feat.CompletionPct, expectedPct)
	}
	if feat.PRDPath != "prd/spec.md" {
		t.Errorf("PRDPath = %q, want %q", feat.PRDPath, "prd/spec.md")
	}
	if feat.DesignPath != "design/tech-design.md" {
		t.Errorf("DesignPath = %q, want %q", feat.DesignPath, "design/tech-design.md")
	}
	if feat.Status != "in-progress" {
		t.Errorf("Status = %q, want %q", feat.Status, "in-progress")
	}

	task11 := feat.Tasks["1.1-setup"]
	if task11.ID != "1.1" {
		t.Errorf("task 1.1 ID = %q", task11.ID)
	}
	if task11.Status != "completed" {
		t.Errorf("task 1.1 Status = %q", task11.Status)
	}
	if task11.Phase != 1 {
		t.Errorf("task 1.1 Phase = %d, want 1", task11.Phase)
	}

	task12 := feat.Tasks["1.2-scanner"]
	if task12.Scope != "backend" {
		t.Errorf("task 1.2 Scope = %q, want backend", task12.Scope)
	}
	if len(task12.Dependencies) != 1 || task12.Dependencies[0] != "1.1" {
		t.Errorf("task 1.2 Dependencies = %v", task12.Dependencies)
	}

	task21 := feat.Tasks["2.1-handlers"]
	if !task21.Breaking {
		t.Error("task 2.1 Breaking should be true")
	}
	if task21.Phase != 2 {
		t.Errorf("task 2.1 Phase = %d, want 2", task21.Phase)
	}

	if len(feat.Phases) != 2 {
		t.Fatalf("expected 2 phases, got %d", len(feat.Phases))
	}
	if feat.Phases[0].Number != 1 {
		t.Errorf("phase[0] Number = %d, want 1", feat.Phases[0].Number)
	}
	if len(feat.Phases[0].TaskKeys) != 2 {
		t.Errorf("phase[0] TaskKeys count = %d, want 2", len(feat.Phases[0].TaskKeys))
	}
	if feat.Phases[1].Number != 2 {
		t.Errorf("phase[1] Number = %d, want 2", feat.Phases[1].Number)
	}
}

func TestIntegration_MalformedIndexJSON(t *testing.T) {
	tmpDir := t.TempDir()

	featDir := filepath.Join(tmpDir, "docs", "features", "bad-feature", "tasks")
	if err := os.MkdirAll(featDir, 0755); err != nil {
		t.Fatalf("failed to create feature dir: %v", err)
	}

	idxPath := filepath.Join(featDir, "index.json")
	if err := os.WriteFile(idxPath, []byte(`{"broken": json}`), 0644); err != nil {
		t.Fatalf("failed to write index.json: %v", err)
	}

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "Proj", Path: tmpDir},
		},
	}

	s := NewScanner(cfg)
	result, err := s.ScanAll()
	if err != nil {
		t.Fatalf("ScanAll() returned error: %v", err)
	}

	pd := result["proj"]
	if len(pd.Features) != 0 {
		t.Errorf("expected 0 features with malformed JSON, got %d", len(pd.Features))
	}
}

func TestIntegration_MissingFeaturesDir(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "Proj", Path: tmpDir},
		},
	}

	s := NewScanner(cfg)
	result, err := s.ScanAll()
	if err != nil {
		t.Fatalf("ScanAll() returned error: %v", err)
	}

	pd := result["proj"]
	if len(pd.Features) != 0 {
		t.Errorf("expected 0 features, got %d", len(pd.Features))
	}
}

// ---- Cache behavior tests ----

func TestScanProject_UsesCache(t *testing.T) {
	tasks := map[string]interface{}{
		"1.1-task": taskEntry("1.1", "Task", "P0", "pending", "all", "1.1.md", nil, false),
	}

	fs := fstest.MapFS{
		featurePath("feat"): &fstest.MapFile{
			Data: indexJSON("feat", "prd.md", "design.md", "planning", tasks),
		},
	}

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "Proj", Path: testProjectPath},
		},
	}

	s := NewScanner(cfg)
	s.fs = fs

	// First call reads from FS
	pd1, _ := s.ScanProject("proj")
	if pd1 == nil {
		t.Fatal("first ScanProject returned nil")
	}

	// Modify the FS - second call should still return cached data
	delete(fs, featurePath("feat"))

	pd2, _ := s.ScanProject("proj")
	if len(pd2.Features) != 1 {
		t.Errorf("cached result should still have 1 feature, got %d", len(pd2.Features))
	}
}

func TestExpandDependencies_NoWildcards(t *testing.T) {
	tasks := map[string]interface{}{
		"1.1-task": taskEntry("1.1", "Task", "P0", "pending", "all", "1.1.md", []string{"1.0"}, false),
	}

	fs := fstest.MapFS{
		featurePath("feat"): &fstest.MapFile{
			Data: indexJSON("feat", "prd.md", "design.md", "planning", tasks),
		},
	}

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "Proj", Path: testProjectPath},
		},
	}

	s := NewScanner(cfg)
	s.fs = fs

	result, _ := s.ScanAll()
	task := result["proj"].Features[0].Tasks["1.1-task"]
	if len(task.Dependencies) != 1 || task.Dependencies[0] != "1.0" {
		t.Errorf("Dependencies = %v, want [1.0]", task.Dependencies)
	}
}

func TestRecordPath(t *testing.T) {
	tasks := map[string]interface{}{
		"1.1-task": map[string]interface{}{
			"id":       "1.1",
			"title":    "Task",
			"priority": "P0",
			"status":   "completed",
			"scope":    "all",
			"file":     "1.1-task.md",
			"record":   "records/1.1-task.md",
			"breaking": false,
		},
	}

	fs := fstest.MapFS{
		featurePath("feat"): &fstest.MapFile{
			Data: indexJSON("feat", "prd.md", "design.md", "completed", tasks),
		},
	}

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "Proj", Path: testProjectPath},
		},
	}

	s := NewScanner(cfg)
	s.fs = fs

	result, _ := s.ScanAll()
	task := result["proj"].Features[0].Tasks["1.1-task"]
	if task.Record != "records/1.1-task.md" {
		t.Errorf("Record = %q, want %q", task.Record, "records/1.1-task.md")
	}
}

func TestEstimatedTime(t *testing.T) {
	tasks := map[string]interface{}{
		"1.1-task": map[string]interface{}{
			"id":            "1.1",
			"title":         "Task",
			"priority":      "P0",
			"status":        "pending",
			"scope":         "all",
			"file":          "1.1-task.md",
			"estimatedTime": "3h",
			"breaking":      false,
		},
	}

	fs := fstest.MapFS{
		featurePath("feat"): &fstest.MapFile{
			Data: indexJSON("feat", "prd.md", "design.md", "planning", tasks),
		},
	}

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "Proj", Path: testProjectPath},
		},
	}

	s := NewScanner(cfg)
	s.fs = fs

	result, _ := s.ScanAll()
	task := result["proj"].Features[0].Tasks["1.1-task"]
	if task.EstimatedTime != "3h" {
		t.Errorf("EstimatedTime = %q, want %q", task.EstimatedTime, "3h")
	}
}

func TestScanAll_SortFeaturesApplied(t *testing.T) {
	now := time.Now()
	fs := fstest.MapFS{}

	for i, slug := range []string{"c-feature", "a-feature", "b-feature"} {
		taskStatus := "pending"
		blocked := false
		if i == 1 {
			taskStatus = "in_progress"
		}
		if i == 2 {
			taskStatus = "blocked"
			blocked = true
		}

		tasks := map[string]interface{}{
			fmt.Sprintf("1.%d-task", i+1): map[string]interface{}{
				"id":       fmt.Sprintf("1.%d", i+1),
				"title":    fmt.Sprintf("Task %d", i+1),
				"priority": "P0",
				"status":   taskStatus,
				"scope":    "all",
				"file":     fmt.Sprintf("1.%d-task.md", i+1),
				"breaking": blocked,
			},
		}

		fs[featurePath(slug)] = &fstest.MapFile{
			Data:    indexJSON(slug, "prd.md", "design.md", "planning", tasks),
			ModTime: now,
		}
	}

	cfg := &config.Config{
		Projects: []config.ProjectConfig{
			{Name: "Proj", Path: testProjectPath},
		},
	}

	s := NewScanner(cfg)
	s.fs = fs

	result, _ := s.ScanAll()
	pd := result["proj"]

	if !pd.Features[0].HasBlockedTasks {
		t.Errorf("first feature should have blocked tasks, got slug=%q", pd.Features[0].Slug)
	}
}

// keys helper returns sorted keys of a map.
func keys(m map[string]*model.ProjectData) []string {
	k := make([]string, 0, len(m))
	for key := range m {
		k = append(k, key)
	}
	sort.Strings(k)
	return k
}
