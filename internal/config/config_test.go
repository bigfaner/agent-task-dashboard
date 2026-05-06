package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/panda/agent-task-center/internal/model"
)

// writeTestConfig writes a YAML string to a temp file and returns its path.
func writeTestConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test-config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}
	return path
}

func TestLoad_ValidConfig(t *testing.T) {
	path := writeTestConfig(t, `
projects:
  - name: "My Project"
    path: "/home/user/projects/my-project"
  - name: "Another"
    path: "/home/user/projects/another"
server:
  port: 9090
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if len(cfg.Projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(cfg.Projects))
	}
	if cfg.Projects[0].Name != "My Project" {
		t.Errorf("Projects[0].Name = %q, want %q", cfg.Projects[0].Name, "My Project")
	}
	if cfg.Projects[0].Path != "/home/user/projects/my-project" {
		t.Errorf("Projects[0].Path = %q, want %q", cfg.Projects[0].Path, "/home/user/projects/my-project")
	}
	if cfg.Projects[1].Name != "Another" {
		t.Errorf("Projects[1].Name = %q, want %q", cfg.Projects[1].Name, "Another")
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("Server.Port = %d, want %d", cfg.Server.Port, 9090)
	}
}

func TestLoad_DefaultPort(t *testing.T) {
	path := writeTestConfig(t, `
projects:
  - name: "Project"
    path: "/tmp/proj"
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want default 8080", cfg.Server.Port)
	}
}

func TestLoad_DefaultPortWhenServerSectionOmitted(t *testing.T) {
	path := writeTestConfig(t, `
projects:
  - name: "P"
    path: "/x"
server: {}
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want default 8080", cfg.Server.Port)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("Load() should return error for missing file")
	}
	// Should be ErrConfigInvalid
	if _, ok := err.(model.ErrConfigInvalid); !ok {
		t.Errorf("error type = %T, want model.ErrConfigInvalid", err)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	path := writeTestConfig(t, `
projects:
  - name: "Bad
    path: [invalid yaml
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() should return error for invalid YAML")
	}
	if _, ok := err.(model.ErrConfigInvalid); !ok {
		t.Errorf("error type = %T, want model.ErrConfigInvalid", err)
	}
}

func TestLoad_NoProjects(t *testing.T) {
	path := writeTestConfig(t, `
projects: []
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() should return error when no projects defined")
	}
	if _, ok := err.(model.ErrConfigInvalid); !ok {
		t.Errorf("error type = %T, want model.ErrConfigInvalid", err)
	}
}

func TestLoad_EnvExpansion(t *testing.T) {
	if err := os.Setenv("TASKDASH_TEST_HOME", "/expanded/home"); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}
	defer func() { _ = os.Unsetenv("TASKDASH_TEST_HOME") }()

	path := writeTestConfig(t, `
projects:
  - name: "Env Project"
    path: "$TASKDASH_TEST_HOME/my-project"
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if cfg.Projects[0].Path != "/expanded/home/my-project" {
		t.Errorf("Projects[0].Path = %q, want %q", cfg.Projects[0].Path, "/expanded/home/my-project")
	}
}

func TestLoad_TildeExpansion(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory")
	}
	path := writeTestConfig(t, `
projects:
  - name: "Tilde Project"
    path: "~/projects/my-project"
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	expected := filepath.Join(home, "projects/my-project")
	if cfg.Projects[0].Path != expected {
		t.Errorf("Projects[0].Path = %q, want %q", cfg.Projects[0].Path, expected)
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	path := writeTestConfig(t, "")
	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() should return error for empty config (no projects)")
	}
	if _, ok := err.(model.ErrConfigInvalid); !ok {
		t.Errorf("error type = %T, want model.ErrConfigInvalid", err)
	}
}
