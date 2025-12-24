package history

import (
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/marcelsud/swarmctl/internal/executor"
)

// MockExecutor for testing
type MockExecutor struct {
	runCommands []string
	runResults  map[string]*executor.CommandResult
	runErrors   map[string]error
	writeFiles  map[string][]byte
}

func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		runCommands: make([]string, 0),
		runResults:  make(map[string]*executor.CommandResult),
		runErrors:   make(map[string]error),
		writeFiles:  make(map[string][]byte),
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
	m.runCommands = append(m.runCommands, cmd)
	return nil
}

func (m *MockExecutor) WriteFile(path string, content []byte) error {
	m.writeFiles[path] = content
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

func TestNewManager(t *testing.T) {
	mockExec := NewMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	if manager == nil {
		t.Error("NewManager() should not return nil")
	}

	if manager.stackName != "test-stack" {
		t.Errorf("stackName = %q, want %q", manager.stackName, "test-stack")
	}

	if manager.containerName != "test-stack-history" {
		t.Errorf("containerName = %q, want %q", manager.containerName, "test-stack-history")
	}
}

func TestManager_IsRunning_Running(t *testing.T) {
	mockExec := NewMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	cmd := "docker ps --filter name=^test-stack-history$ --format '{{.Names}}'"
	mockExec.SetRunResult(cmd, &executor.CommandResult{
		Stdout:   "test-stack-history",
		ExitCode: 0,
	})

	running, err := manager.IsRunning()
	if err != nil {
		t.Errorf("IsRunning() error = %v", err)
	}

	if !running {
		t.Error("IsRunning() should return true when container is running")
	}
}

func TestManager_IsRunning_NotRunning(t *testing.T) {
	mockExec := NewMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	cmd := "docker ps --filter name=^test-stack-history$ --format '{{.Names}}'"
	mockExec.SetRunResult(cmd, &executor.CommandResult{
		Stdout:   "",
		ExitCode: 0,
	})

	running, err := manager.IsRunning()
	if err != nil {
		t.Errorf("IsRunning() error = %v", err)
	}

	if running {
		t.Error("IsRunning() should return false when container is not running")
	}
}

func TestManager_EnsureRunning_AlreadyRunning(t *testing.T) {
	mockExec := NewMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	// Mock IsRunning to return true
	isRunningCmd := "docker ps --filter name=^test-stack-history$ --format '{{.Names}}'"
	mockExec.SetRunResult(isRunningCmd, &executor.CommandResult{
		Stdout:   "test-stack-history",
		ExitCode: 0,
	})

	err := manager.EnsureRunning()
	if err != nil {
		t.Errorf("EnsureRunning() error = %v", err)
	}

	// Should not try to start container
	commands := mockExec.GetRunCommands()
	startCmd := "docker run -d --name test-stack-history --restart unless-stopped -v test-stack_history_data:/data docker.io/marcelsud/swarmctl-history:latest"
	for _, cmd := range commands {
		if cmd == startCmd {
			t.Error("Should not try to start container when already running")
			break
		}
	}
}

func TestManager_EnsureRunning_NewContainer(t *testing.T) {
	mockExec := NewMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	// Mock IsRunning to return false
	isRunningCmd := "docker ps --filter name=^test-stack-history$ --format '{{.Names}}'"
	mockExec.SetRunResult(isRunningCmd, &executor.CommandResult{
		Stdout:   "",
		ExitCode: 0,
	})

	// Mock successful container creation
	startCmd := "docker run -d --name test-stack-history --restart unless-stopped -v test-stack_history_data:/data docker.io/marcelsud/swarmctl-history:latest"
	mockExec.SetRunResult(startCmd, &executor.CommandResult{
		Stdout:   "container-id",
		ExitCode: 0,
	})

	err := manager.EnsureRunning()
	if err != nil {
		t.Errorf("EnsureRunning() error = %v", err)
	}

	commands := mockExec.GetRunCommands()
	found := false
	for _, cmd := range commands {
		if cmd == startCmd {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected command %q not found", startCmd)
	}
}

func TestManager_EnsureRunning_ContainerExistsStopped(t *testing.T) {
	mockExec := NewMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	// Mock IsRunning to return false
	isRunningCmd := "docker ps --filter name=^test-stack-history$ --format '{{.Names}}'"
	mockExec.SetRunResult(isRunningCmd, &executor.CommandResult{
		Stdout:   "",
		ExitCode: 0,
	})

	// Mock docker run fails (container exists but stopped)
	startCmd := "docker run -d --name test-stack-history --restart unless-stopped -v test-stack_history_data:/data docker.io/marcelsud/swarmctl-history:latest"
	mockExec.SetRunResult(startCmd, &executor.CommandResult{
		Stdout:   "",
		Stderr:   "container already exists",
		ExitCode: 125,
	})

	// Mock docker start succeeds
	startExistingCmd := "docker start test-stack-history"
	mockExec.SetRunResult(startExistingCmd, &executor.CommandResult{
		Stdout:   "test-stack-history",
		ExitCode: 0,
	})

	err := manager.EnsureRunning()
	if err != nil {
		t.Errorf("EnsureRunning() error = %v", err)
	}

	commands := mockExec.GetRunCommands()
	foundStart := false
	foundStartExisting := false
	for _, cmd := range commands {
		if cmd == startCmd {
			foundStart = true
		}
		if cmd == startExistingCmd {
			foundStartExisting = true
		}
	}
	if !foundStart {
		t.Error("Should attempt to create container")
	}
	if !foundStartExisting {
		t.Error("Should attempt to start existing container when create fails")
	}
}

func TestManager_Record(t *testing.T) {
	mockExec := NewMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	// Mock EnsureRunning success
	isRunningCmd := "docker ps --filter name=^test-stack-history$ --format '{{.Names}}'"
	mockExec.SetRunResult(isRunningCmd, &executor.CommandResult{
		Stdout:   "test-stack-history",
		ExitCode: 0,
	})

	// Mock successful docker cp and exec commands
	composeContent := []byte("version: '3.8'\nservices:\n  web:\n    image: nginx:latest")
	images := map[string]string{"web": "nginx:latest"}

	copyCmd := "docker cp /tmp/test-stack-compose-record.yaml test-stack-history:/tmp/compose-record.yaml"
	mockExec.SetRunResult(copyCmd, &executor.CommandResult{
		ExitCode: 0,
	})

	recordCmd := "docker exec test-stack-history /app/history record --stack test-stack --compose-file /tmp/compose-record.yaml --images '{\"web\":\"nginx:latest\"}'"
	mockExec.SetRunResult(recordCmd, &executor.CommandResult{
		ExitCode: 0,
	})

	err := manager.Record(composeContent, images)
	if err != nil {
		t.Errorf("Record() error = %v", err)
	}

	// Check if temp file was written
	writtenFiles := mockExec.GetWrittenFiles()
	tempPath := "/tmp/test-stack-compose-record.yaml"
	if _, exists := writtenFiles[tempPath]; !exists {
		t.Error("Temp compose file should be written")
	}

	// Check commands
	commands := mockExec.GetRunCommands()
	if !containsCommand(commands, copyCmd) {
		t.Error("Should execute docker cp command")
	}
	if !containsCommand(commands, recordCmd) {
		t.Error("Should execute history record command")
	}
}

func TestManager_List(t *testing.T) {
	mockExec := NewMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	// Mock EnsureRunning success
	isRunningCmd := "docker ps --filter name=^test-stack-history$ --format '{{.Names}}'"
	mockExec.SetRunResult(isRunningCmd, &executor.CommandResult{
		Stdout:   "test-stack-history",
		ExitCode: 0,
	})

	// Mock history list response
	expectedRecords := []DeployRecord{
		{
			ID:         1,
			StackName:  "test-stack",
			DeployedAt: time.Now(),
			Images:     map[string]string{"web": "nginx:latest"},
		},
	}
	recordsJSON, _ := json.Marshal(expectedRecords)

	listCmd := "docker exec test-stack-history /app/history list --stack test-stack --limit 5 --format json"
	mockExec.SetRunResult(listCmd, &executor.CommandResult{
		Stdout:   string(recordsJSON),
		ExitCode: 0,
	})

	records, err := manager.List(5)
	if err != nil {
		t.Errorf("List() error = %v", err)
	}

	if len(records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(records))
	}

	if records[0].ID != 1 {
		t.Errorf("Expected record ID 1, got %d", records[0].ID)
	}
}

func TestManager_GetPrevious_NoPrevious(t *testing.T) {
	mockExec := NewMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	// Mock List with only one record
	isRunningCmd := "docker ps --filter name=^test-stack-history$ --format '{{.Names}}'"
	mockExec.SetRunResult(isRunningCmd, &executor.CommandResult{
		Stdout:   "test-stack-history",
		ExitCode: 0,
	})

	singleRecord := []DeployRecord{{ID: 1}}
	recordsJSON, _ := json.Marshal(singleRecord)

	listCmd := "docker exec test-stack-history /app/history list --stack test-stack --limit 2 --format json"
	mockExec.SetRunResult(listCmd, &executor.CommandResult{
		Stdout:   string(recordsJSON),
		ExitCode: 0,
	})

	_, err := manager.GetPrevious()
	if err == nil {
		t.Error("GetPrevious() should return error when no previous deploy exists")
	}

	if !containsString(err.Error(), "no previous deploy") {
		t.Errorf("Expected 'no previous deploy' error, got: %v", err)
	}
}

func TestManager_GetComposeContent(t *testing.T) {
	mockExec := NewMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	// Mock Get response
	expectedCompose := "version: '3.8'\nservices:\n  web:\n    image: nginx:latest"
	expectedRecord := DeployRecord{
		ID:             1,
		StackName:      "test-stack",
		ComposeContent: expectedCompose,
	}
	recordJSON, _ := json.Marshal(expectedRecord)

	getCmd := "docker exec test-stack-history /app/history get --stack test-stack --offset -1 --format json"
	mockExec.SetRunResult(getCmd, &executor.CommandResult{
		Stdout:   string(recordJSON),
		ExitCode: 0,
	})

	composeContent, err := manager.GetComposeContent(-1)
	if err != nil {
		t.Errorf("GetComposeContent() error = %v", err)
	}

	if string(composeContent) != expectedCompose {
		t.Errorf("GetComposeContent() = %q, want %q", string(composeContent), expectedCompose)
	}
}

func TestManager_Stop(t *testing.T) {
	mockExec := NewMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	stopCmd := "docker stop test-stack-history"
	mockExec.SetRunResult(stopCmd, &executor.CommandResult{
		ExitCode: 0,
	})

	err := manager.Stop()
	if err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	commands := mockExec.GetRunCommands()
	if !containsCommand(commands, stopCmd) {
		t.Error("Should execute docker stop command")
	}
}

func TestManager_Remove(t *testing.T) {
	mockExec := NewMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	// Mock all removal commands
	mockExec.SetRunResult("docker stop test-stack-history", &executor.CommandResult{ExitCode: 0})
	mockExec.SetRunResult("docker rm test-stack-history", &executor.CommandResult{ExitCode: 0})
	mockExec.SetRunResult("docker volume rm test-stack_history_data", &executor.CommandResult{ExitCode: 0})

	err := manager.Remove()
	if err != nil {
		t.Errorf("Remove() error = %v", err)
	}

	commands := mockExec.GetRunCommands()
	expectedCommands := []string{
		"docker stop test-stack-history",
		"docker rm test-stack-history",
		"docker volume rm test-stack_history_data",
	}

	for _, expectedCmd := range expectedCommands {
		if !containsCommand(commands, expectedCmd) {
			t.Errorf("Should execute command: %s", expectedCmd)
		}
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

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsString(s[1:len(s)-1], substr))))
}
