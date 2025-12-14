package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Load reads and parses the swarm.yaml configuration file
func Load(path string) (*Config, error) {
	cfg := NewConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Normalize mode to lowercase
	cfg.Mode = DeploymentMode(strings.ToLower(string(cfg.Mode)))

	// Expand ~ in key path
	if cfg.SSH.Key != "" {
		cfg.SSH.Key = expandPath(cfg.SSH.Key)
	}

	// Resolve compose file path relative to config file
	if cfg.ComposeFile != "" && !filepath.IsAbs(cfg.ComposeFile) {
		configDir := filepath.Dir(path)
		cfg.ComposeFile = filepath.Join(configDir, cfg.ComposeFile)
	}

	// Load registry password from environment if not set
	if cfg.Registry.Password == "" {
		cfg.Registry.Password = os.Getenv("SWARMCTL_REGISTRY_PASSWORD")
	}

	return cfg, nil
}

// LoadComposeFile reads the docker-compose.yaml file
func LoadComposeFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read compose file: %w", err)
	}
	return data, nil
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}
