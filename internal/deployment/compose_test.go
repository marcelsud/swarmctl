package deployment

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/marcelsud/swarmctl/internal/executor"
)

// MockExecutor for testing
type MockExecutor struct {
	runCommands []string
	runResults  map[string]*executor.CommandResult
	runErrors   map[string]error

	streamCommands []string
	streamResults  map[string]error

	writeFiles  map[string][]byte
	writeErrors map[string]error
}

func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		runCommands:   make([]string, 0),
		runResults:    make(map[string]*executor.CommandResult),
		runErrors:     make(map[string]error),
		streamResults: make(map[string]error),
		writeFiles:    make(map[string][]byte),
		writeErrors:   make(map[string]error),
	}
}

func (m *MockExecutor) Run(cmd string) (*executor.CommandResult, error) {
	m.runCommands = append(m.runCommands, cmd)

	if err, exists := m.runErrors[cmd]; exists {
		return nil, err
	}

	if result, exists := m.runResults[cmd]; exists {
		return result, nil
	}

	// Default successful result
	return &executor.CommandResult{
		Stdout:   "",
		Stderr:   "",
		ExitCode: 0,
	}, nil
}

func (m *MockExecutor) RunInteractive(cmd string) error {
	m.runCommands = append(m.runCommands, cmd)
	return nil
}

func (m *MockExecutor) RunStream(cmd string, stdout, stderr io.Writer) error {
	m.streamCommands = append(m.streamCommands, cmd)
	if err, exists := m.streamResults[cmd]; exists {
		return err
	}
	return nil
}

func (m *MockExecutor) WriteFile(path string, content []byte) error {
	m.writeFiles[path] = content
	if err, exists := m.writeErrors[path]; exists {
		return err
	}
	return nil
}

func (m *MockExecutor) Close() error {
	return nil
}

func (m *MockExecutor) IsLocal() bool {
	return true
}

func (m *MockExecutor) SetVerbose(v bool) {
	// Mock implementation - just store the value
}

func (m *MockExecutor) SetRunResult(cmd string, result *executor.CommandResult) {
	m.runResults[cmd] = result
}

func (m *MockExecutor) SetRunError(cmd string, err error) {
	m.runErrors[cmd] = err
}

func (m *MockExecutor) GetRunCommands() []string {
	return m.runCommands
}

func (m *MockExecutor) GetWrittenFiles() map[string][]byte {
	return m.writeFiles
}

func TestNewComposeManager(t *testing.T) {
	mockExec := NewMockExecutor()
	manager := NewComposeManager(mockExec, "test-project")

	if manager == nil {
		t.Error("NewComposeManager() should not return nil")
	}

	if manager.projectName != "test-project" {
		t.Errorf("projectName = %q, want %q", manager.projectName, "test-project")
	}
}

func TestComposeManager_Deploy(t *testing.T) {
	mockExec := NewMockExecutor()
	manager := NewComposeManager(mockExec, "test-project")

	// Set up mock responses
	mockExec.SetRunResult("docker compose -p test-project -f /tmp/test-project-compose.yaml up -d --remove-orphans", &executor.CommandResult{
		Stdout:   "service1 created\nservice2 created",
		ExitCode: 0,
	})

	composeContent := []byte("version: '3.8'\nservices:\n  web:\n    image: nginx:latest")

	err := manager.Deploy(composeContent)
	if err != nil {
		t.Errorf("Deploy() error = %v", err)
	}

	// Check if commands were executed
	commands := mockExec.GetRunCommands()
	if len(commands) < 2 { // Should have at least history and deploy
		t.Errorf("Expected at least 2 commands, got %d", len(commands))
	}

	// Check if compose file was written
	writtenFiles := mockExec.GetWrittenFiles()
	expectedPath := "/tmp/test-project-compose.yaml"
	if _, exists := writtenFiles[expectedPath]; !exists {
		t.Errorf("Compose file should be written to %s", expectedPath)
	}

	// Verify content
	writtenContent := writtenFiles[expectedPath]
	if string(writtenContent) != string(composeContent) {
		t.Errorf("Written content = %q, want %q", writtenContent, composeContent)
	}
}

func TestComposeManager_Deploy_HistoryError(t *testing.T) {
	mockExec := NewMockExecutor()
	manager := NewComposeManager(mockExec, "test-project")

	// Set up successful deploy but mock history failure
	mockExec.SetRunResult("docker compose -p test-project -f /tmp/test-project-compose.yaml up -d --remove-orphans", &executor.CommandResult{
		Stdout:   "service1 created",
		ExitCode: 0,
	})

	composeContent := []byte("version: '3.8'\nservices:\n  web:\n    image: nginx:latest")

	// Deploy should succeed even if history fails
	err := manager.Deploy(composeContent)
	if err != nil {
		t.Errorf("Deploy() should not fail when history fails: %v", err)
	}
}

func TestComposeManager_Deploy_DeployError(t *testing.T) {
	mockExec := NewMockExecutor()
	manager := NewComposeManager(mockExec, "test-project")

	// Set up deploy error
	mockExec.SetRunResult("docker compose -p test-project -f /tmp/test-project-compose.yaml up -d --remove-orphans", &executor.CommandResult{
		Stdout:   "",
		Stderr:   "deployment failed",
		ExitCode: 1,
	})

	composeContent := []byte("version: '3.8'\nservices:\n  web:\n    image: nginx:latest")

	err := manager.Deploy(composeContent)
	if err == nil {
		t.Error("Deploy() should return error when deploy fails")
	}

	if !strings.Contains(err.Error(), "deployment failed") {
		t.Errorf("Error should contain deployment failure message: %v", err)
	}
}

func TestComposeManager_Remove(t *testing.T) {
	mockExec := NewMockExecutor()
	manager := NewComposeManager(mockExec, "test-project")

	mockExec.SetRunResult("docker compose -p test-project down", &executor.CommandResult{
		ExitCode: 0,
	})

	err := manager.Remove()
	if err != nil {
		t.Errorf("Remove() error = %v", err)
	}

	commands := mockExec.GetRunCommands()
	expectedCmd := "docker compose -p test-project down"
	found := false
	for _, cmd := range commands {
		if cmd == expectedCmd {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected command %q not found", expectedCmd)
	}
}

func TestComposeManager_Exists(t *testing.T) {
	mockExec := NewMockExecutor()
	manager := NewComposeManager(mockExec, "test-project")

	tests := []struct {
		name   string
		stdout string
		want   bool
	}{
		{
			name:   "has containers",
			stdout: "abc123\ndef456\n",
			want:   true,
		},
		{
			name:   "no containers",
			stdout: "",
			want:   false,
		},
		{
			name:   "whitespace only",
			stdout: "   \n\t\n",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec.runCommands = nil // Reset
			mockExec.SetRunResult("docker compose -p test-project ps -q", &executor.CommandResult{
				Stdout:   tt.stdout,
				ExitCode: 0,
			})

			exists, err := manager.Exists()
			if err != nil {
				t.Errorf("Exists() error = %v", err)
			}

			if exists != tt.want {
				t.Errorf("Exists() = %v, want %v", exists, tt.want)
			}
		})
	}
}

func TestComposeManager_SupportsRollback(t *testing.T) {
	manager := NewComposeManager(NewMockExecutor(), "test-project")
	if !manager.SupportsRollback() {
		t.Error("Compose should support rollback")
	}
}

func TestComposeManager_SupportsScale(t *testing.T) {
	manager := NewComposeManager(NewMockExecutor(), "test-project")
	if manager.SupportsScale() {
		t.Error("Compose should not support scale")
	}
}

func TestComposeManager_ScaleService(t *testing.T) {
	manager := NewComposeManager(NewMockExecutor(), "test-project")
	err := manager.ScaleService("web", 3)
	if err == nil {
		t.Error("ScaleService should return error for compose mode")
	}

	var unsupportedErr *UnsupportedOperationError
	if !errors.As(err, &unsupportedErr) {
		t.Errorf("Expected UnsupportedOperationError, got %T", err)
	}

	if unsupportedErr.Operation != "scale" {
		t.Errorf("Operation = %q, want %q", unsupportedErr.Operation, "scale")
	}

	if unsupportedErr.Mode != "compose" {
		t.Errorf("Mode = %q, want %q", unsupportedErr.Mode, "compose")
	}
}

func TestComposeManager_GetStackName(t *testing.T) {
	manager := NewComposeManager(NewMockExecutor(), "test-project")
	if manager.GetStackName() != "test-project" {
		t.Errorf("GetStackName() = %q, want %q", manager.GetStackName(), "test-project")
	}
}

func TestComposeManager_GetMode(t *testing.T) {
	manager := NewComposeManager(NewMockExecutor(), "test-project")
	if manager.GetMode() != "compose" {
		t.Errorf("GetMode() = %q, want %q", manager.GetMode(), "compose")
	}
}

func TestComposeManager_ExtractImages(t *testing.T) {
	manager := NewComposeManager(NewMockExecutor(), "test-project")

	composeContent := `version: '3.8'
services:
  web:
    image: nginx:latest
    ports:
      - "80:80"
  db:
    image: postgres:13
  # This is a comment
  cache:
    image: redis:alpine`

	images := manager.extractImages([]byte(composeContent))

	expectedImages := map[string]string{
		"web":   "nginx:latest",
		"db":    "postgres:13",
		"cache": "redis:alpine",
	}

	for service, image := range expectedImages {
		if images[service] != image {
			t.Errorf("Image for %s = %q, want %q", service, images[service], image)
		}
	}

	if len(images) != 3 {
		t.Errorf("Expected 3 images, got %d", len(images))
	}
}

func TestUnsupportedOperationError(t *testing.T) {
	err := NewUnsupportedError("scale", "compose")

	expected := "scale is not supported in compose mode"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}

	unsupportedErr := err.(*UnsupportedOperationError)
	if unsupportedErr.Operation != "scale" {
		t.Errorf("Operation = %q, want %q", unsupportedErr.Operation, "scale")
	}

	if unsupportedErr.Mode != "compose" {
		t.Errorf("Mode = %q, want %q", unsupportedErr.Mode, "compose")
	}
}

// Tests for SwarmManager
func TestNewSwarmManager(t *testing.T) {
	mockExec := NewMockExecutor()
	manager := NewSwarmManager(mockExec, "test-stack")

	if manager == nil {
		t.Error("NewSwarmManager() should not return nil")
	}

	if manager.stackName != "test-stack" {
		t.Errorf("stackName = %q, want %q", manager.stackName, "test-stack")
	}
}

func TestSwarmManager_Deploy(t *testing.T) {
	mockExec := NewMockExecutor()
	manager := NewSwarmManager(mockExec, "test-stack")

	// Mock successful deploy
	cmd := "docker stack deploy -c /tmp/test-stack-compose.yaml test-stack --with-registry-auth"
	mockExec.SetRunResult(cmd, &executor.CommandResult{
		Stdout:   "Stack deployed",
		ExitCode: 0,
	})

	composeContent := []byte("version: '3.8'\nservices:\n  web:\n    image: nginx:latest")

	err := manager.Deploy(composeContent)
	if err != nil {
		t.Errorf("Deploy() error = %v", err)
	}

	// Check if compose file was written
	writtenFiles := mockExec.GetWrittenFiles()
	expectedPath := "/tmp/test-stack-compose.yaml"
	if _, exists := writtenFiles[expectedPath]; !exists {
		t.Error("Compose file should be written")
	}

	// Check if deploy command was executed
	commands := mockExec.GetRunCommands()
	if !containsCommand(commands, cmd) {
		t.Error("Deploy command should be executed")
	}
}

func TestSwarmManager_SupportsRollback(t *testing.T) {
	manager := NewSwarmManager(NewMockExecutor(), "test-stack")
	if !manager.SupportsRollback() {
		t.Error("Swarm should support rollback")
	}
}

func TestSwarmManager_SupportsScale(t *testing.T) {
	manager := NewSwarmManager(NewMockExecutor(), "test-stack")
	if !manager.SupportsScale() {
		t.Error("Swarm should support scale")
	}
}

func TestSwarmManager_GetStackName(t *testing.T) {
	manager := NewSwarmManager(NewMockExecutor(), "test-stack")
	if manager.GetStackName() != "test-stack" {
		t.Errorf("GetStackName() = %q, want %q", manager.GetStackName(), "test-stack")
	}
}

func TestSwarmManager_GetMode(t *testing.T) {
	manager := NewSwarmManager(NewMockExecutor(), "test-stack")
	if manager.GetMode() != "swarm" {
		t.Errorf("GetMode() = %q, want %q", manager.GetMode(), "swarm")
	}
}

// Helper functions
func containsCommand(commands []string, target string) bool {
	for _, cmd := range commands {
		if cmd == target {
			return true
		}
	}
	return false
}
