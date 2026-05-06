package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/panda/agent-task-center/internal/model"
	"gopkg.in/yaml.v3"
)

// Config holds project list and server settings.
type Config struct {
	Projects []ProjectConfig `yaml:"projects"`
	Server   ServerConfig    `yaml:"server,omitempty"`
}

// ProjectConfig defines a single project entry in the config file.
type ProjectConfig struct {
	Name string `yaml:"name"` // Display name, also used as URL ID (lowercased)
	Path string `yaml:"path"` // Absolute path to project root
}

// ServerConfig holds server settings.
type ServerConfig struct {
	Port int `yaml:"port,omitempty"` // Default: 8080
}

// Load reads the YAML config file at the given path and returns a parsed Config.
// Returns ErrConfigInvalid if the file is missing, malformed, or has no projects.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, model.ErrConfigInvalid(fmt.Sprintf("cannot read config file: %v", err))
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, model.ErrConfigInvalid(fmt.Sprintf("cannot parse config YAML: %v", err))
	}

	if len(cfg.Projects) == 0 {
		return nil, model.ErrConfigInvalid("at least one project must be defined")
	}

	// Expand environment variables and ~ in project paths
	for i := range cfg.Projects {
		cfg.Projects[i].Path = expandPath(cfg.Projects[i].Path)
	}

	// Apply default port
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}

	return &cfg, nil
}

// expandPath performs ~ and environment variable expansion on a path.
func expandPath(path string) string {
	// Expand ~ to home directory
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			path = filepath.Join(home, path[2:])
		}
	}
	// Expand environment variables ($VAR or ${VAR})
	path = os.ExpandEnv(path)
	return path
}
