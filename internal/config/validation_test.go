package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidationError(t *testing.T) {
	ve := &ValidationError{}

	if ve.HasErrors() {
		t.Error("expected no errors initially")
	}

	ve.Add("error 1")
	ve.Add("error 2")

	if !ve.HasErrors() {
		t.Error("expected errors after adding")
	}

	if len(ve.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(ve.Errors))
	}

	errStr := ve.Error()
	if !strings.Contains(errStr, "error 1") || !strings.Contains(errStr, "error 2") {
		t.Errorf("error string should contain both errors: %s", errStr)
	}
}

func TestValidateRequiredFields(t *testing.T) {
	cfg := &Config{}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error for empty config")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	// Only stack is required (SSH is optional for local mode)
	if len(ve.Errors) < 1 {
		t.Errorf("expected at least 1 validation error, got %d: %v", len(ve.Errors), ve.Errors)
	}

	found := false
	for _, e := range ve.Errors {
		if strings.Contains(e, "stack name is required") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'stack name is required' error, got: %v", ve.Errors)
	}
}

func TestValidateSSHUserRequiredWhenHostSet(t *testing.T) {
	tmpDir := t.TempDir()

	composePath := filepath.Join(tmpDir, "docker-compose.yaml")
	if err := os.WriteFile(composePath, []byte("version: '3.8'"), 0644); err != nil {
		t.Fatal(err)
	}

	// Host set but no user
	cfg := &Config{
		Stack: "myapp",
		SSH: SSHConfig{
			Host: "example.com",
			Port: 22,
		},
		ComposeFile: composePath,
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error when ssh.host is set without ssh.user")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	found := false
	for _, e := range ve.Errors {
		if strings.Contains(e, "ssh.user is required when ssh.host is set") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'ssh.user is required when ssh.host is set' error, got: %v", ve.Errors)
	}
}

func TestValidateLocalModeNoSSH(t *testing.T) {
	tmpDir := t.TempDir()

	composePath := filepath.Join(tmpDir, "docker-compose.yaml")
	if err := os.WriteFile(composePath, []byte("version: '3.8'"), 0644); err != nil {
		t.Fatal(err)
	}

	// No SSH config - should be valid for local mode
	cfg := &Config{
		Stack:       "myapp",
		ComposeFile: composePath,
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("expected no validation error for local mode (no SSH), got: %v", err)
	}
}

func TestValidateValidConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create compose file
	composePath := filepath.Join(tmpDir, "docker-compose.yaml")
	if err := os.WriteFile(composePath, []byte("version: '3.8'"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{
		Stack: "myapp",
		SSH: SSHConfig{
			Host: "example.com",
			User: "deploy",
			Port: 22,
		},
		ComposeFile: composePath,
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("expected no validation error, got: %v", err)
	}
}

func TestValidateInvalidPort(t *testing.T) {
	tmpDir := t.TempDir()

	composePath := filepath.Join(tmpDir, "docker-compose.yaml")
	if err := os.WriteFile(composePath, []byte("version: '3.8'"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		port     int
		hasError bool
	}{
		{0, true},
		{-1, true},
		{65536, true},
		{1, false},
		{22, false},
		{65535, false},
	}

	for _, tt := range tests {
		cfg := &Config{
			Stack: "myapp",
			SSH: SSHConfig{
				Host: "example.com",
				User: "deploy",
				Port: tt.port,
			},
			ComposeFile: composePath,
		}

		err := cfg.Validate()
		hasError := err != nil

		if hasError != tt.hasError {
			t.Errorf("port %d: expected hasError=%v, got %v (err: %v)", tt.port, tt.hasError, hasError, err)
		}
	}
}

func TestValidateComposeFileNotFound(t *testing.T) {
	cfg := &Config{
		Stack: "myapp",
		SSH: SSHConfig{
			Host: "example.com",
			User: "deploy",
			Port: 22,
		},
		ComposeFile: "/nonexistent/docker-compose.yaml",
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected validation error for nonexistent compose file")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	found := false
	for _, e := range ve.Errors {
		if strings.Contains(e, "compose file not found") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'compose file not found' error, got: %v", ve.Errors)
	}
}

func TestValidateSSHKeyNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	composePath := filepath.Join(tmpDir, "docker-compose.yaml")
	if err := os.WriteFile(composePath, []byte("version: '3.8'"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{
		Stack: "myapp",
		SSH: SSHConfig{
			Host: "example.com",
			User: "deploy",
			Port: 22,
			Key:  "/nonexistent/id_rsa",
		},
		ComposeFile: composePath,
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected validation error for nonexistent SSH key")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	found := false
	for _, e := range ve.Errors {
		if strings.Contains(e, "SSH key file not found") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'SSH key file not found' error, got: %v", ve.Errors)
	}
}

func TestValidateWithExistingSSHKey(t *testing.T) {
	tmpDir := t.TempDir()

	// Create compose file
	composePath := filepath.Join(tmpDir, "docker-compose.yaml")
	if err := os.WriteFile(composePath, []byte("version: '3.8'"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create fake SSH key
	keyPath := filepath.Join(tmpDir, "id_rsa")
	if err := os.WriteFile(keyPath, []byte("fake-key"), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{
		Stack: "myapp",
		SSH: SSHConfig{
			Host: "example.com",
			User: "deploy",
			Port: 22,
			Key:  keyPath,
		},
		ComposeFile: composePath,
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("expected no validation error, got: %v", err)
	}
}
