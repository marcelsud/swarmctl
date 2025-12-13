package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()

	if cfg.SSH.Port != 22 {
		t.Errorf("expected default SSH port 22, got %d", cfg.SSH.Port)
	}

	if cfg.ComposeFile != "docker-compose.yaml" {
		t.Errorf("expected default compose file 'docker-compose.yaml', got %s", cfg.ComposeFile)
	}
}

func TestLoad(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create a valid swarm.yaml
	swarmYaml := `
stack: myapp
ssh:
  host: example.com
  user: deploy
  port: 2222
secrets:
  - API_KEY
  - DB_PASSWORD
accessories:
  - redis
compose_file: docker-compose.yaml
`
	swarmPath := filepath.Join(tmpDir, "swarm.yaml")
	if err := os.WriteFile(swarmPath, []byte(swarmYaml), 0644); err != nil {
		t.Fatal(err)
	}

	// Create docker-compose.yaml (needed for validation)
	composePath := filepath.Join(tmpDir, "docker-compose.yaml")
	if err := os.WriteFile(composePath, []byte("version: '3.8'\nservices: {}"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(swarmPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Stack != "myapp" {
		t.Errorf("expected stack 'myapp', got '%s'", cfg.Stack)
	}

	if cfg.SSH.Host != "example.com" {
		t.Errorf("expected SSH host 'example.com', got '%s'", cfg.SSH.Host)
	}

	if cfg.SSH.User != "deploy" {
		t.Errorf("expected SSH user 'deploy', got '%s'", cfg.SSH.User)
	}

	if cfg.SSH.Port != 2222 {
		t.Errorf("expected SSH port 2222, got %d", cfg.SSH.Port)
	}

	if len(cfg.Secrets) != 2 {
		t.Errorf("expected 2 secrets, got %d", len(cfg.Secrets))
	}

	if len(cfg.Accessories) != 1 {
		t.Errorf("expected 1 accessory, got %d", len(cfg.Accessories))
	}

	// Compose file path should be absolute now
	if !filepath.IsAbs(cfg.ComposeFile) {
		t.Errorf("expected absolute compose file path, got '%s'", cfg.ComposeFile)
	}
}

func TestLoadWithDefaultPort(t *testing.T) {
	tmpDir := t.TempDir()

	// Create swarm.yaml without port specified
	swarmYaml := `
stack: myapp
ssh:
  host: example.com
  user: deploy
compose_file: docker-compose.yaml
`
	swarmPath := filepath.Join(tmpDir, "swarm.yaml")
	if err := os.WriteFile(swarmPath, []byte(swarmYaml), 0644); err != nil {
		t.Fatal(err)
	}

	composePath := filepath.Join(tmpDir, "docker-compose.yaml")
	if err := os.WriteFile(composePath, []byte("version: '3.8'\nservices: {}"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(swarmPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.SSH.Port != 22 {
		t.Errorf("expected default SSH port 22, got %d", cfg.SSH.Port)
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/swarm.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadInvalidYaml(t *testing.T) {
	tmpDir := t.TempDir()

	// Create invalid YAML
	invalidYaml := `stack: [invalid yaml`
	swarmPath := filepath.Join(tmpDir, "swarm.yaml")
	if err := os.WriteFile(swarmPath, []byte(invalidYaml), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(swarmPath)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoadRegistryPasswordFromEnv(t *testing.T) {
	tmpDir := t.TempDir()

	swarmYaml := `
stack: myapp
ssh:
  host: example.com
  user: deploy
registry:
  url: ghcr.io
  username: myuser
compose_file: docker-compose.yaml
`
	swarmPath := filepath.Join(tmpDir, "swarm.yaml")
	if err := os.WriteFile(swarmPath, []byte(swarmYaml), 0644); err != nil {
		t.Fatal(err)
	}

	composePath := filepath.Join(tmpDir, "docker-compose.yaml")
	if err := os.WriteFile(composePath, []byte("version: '3.8'\nservices: {}"), 0644); err != nil {
		t.Fatal(err)
	}

	// Set environment variable
	os.Setenv("SWARMCTL_REGISTRY_PASSWORD", "secret123")
	defer os.Unsetenv("SWARMCTL_REGISTRY_PASSWORD")

	cfg, err := Load(swarmPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Registry.Password != "secret123" {
		t.Errorf("expected registry password 'secret123', got '%s'", cfg.Registry.Password)
	}
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot get home directory")
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"~/.ssh/id_rsa", filepath.Join(home, ".ssh/id_rsa")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"", ""},
	}

	for _, tt := range tests {
		result := expandPath(tt.input)
		if result != tt.expected {
			t.Errorf("expandPath(%s) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}

func TestLoadComposeFile(t *testing.T) {
	tmpDir := t.TempDir()

	content := "version: '3.8'\nservices:\n  web:\n    image: nginx"
	composePath := filepath.Join(tmpDir, "docker-compose.yaml")
	if err := os.WriteFile(composePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	data, err := LoadComposeFile(composePath)
	if err != nil {
		t.Fatalf("LoadComposeFile failed: %v", err)
	}

	if string(data) != content {
		t.Errorf("expected content to match, got '%s'", string(data))
	}
}

func TestLoadComposeFileNotFound(t *testing.T) {
	_, err := LoadComposeFile("/nonexistent/docker-compose.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
