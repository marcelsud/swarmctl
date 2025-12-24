package swarm

import (
	"errors"
	"io"
	"testing"

	"github.com/marcelsud/swarmctl/internal/executor"
)

// SwarmMockExecutor for testing
type SwarmMockExecutor struct {
	runCommands []string
	runResults  map[string]*executor.CommandResult
	runErrors   map[string]error
}

func NewSwarmMockExecutor() *SwarmMockExecutor {
	return &SwarmMockExecutor{
		runCommands: make([]string, 0),
		runResults:  make(map[string]*executor.CommandResult),
		runErrors:   make(map[string]error),
	}
}

func (m *SwarmMockExecutor) Run(cmd string) (*executor.CommandResult, error) {
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

func (m *SwarmMockExecutor) RunInteractive(cmd string) error {
	m.runCommands = append(m.runCommands, cmd)
	return nil
}

func (m *SwarmMockExecutor) RunStream(cmd string, stdout, stderr io.Writer) error {
	m.runCommands = append(m.runCommands, cmd)
	return nil
}

func (m *SwarmMockExecutor) WriteFile(path string, content []byte) error {
	return nil
}

func (m *SwarmMockExecutor) Close() error {
	return nil
}

func (m *SwarmMockExecutor) IsLocal() bool {
	return true
}

func (m *SwarmMockExecutor) SetVerbose(v bool) {
	// Mock implementation - just store the value
}

func (m *SwarmMockExecutor) SetRunResult(cmd string, result *executor.CommandResult) {
	m.runResults[cmd] = result
}

func (m *SwarmMockExecutor) SetRunError(cmd string, err error) {
	m.runErrors[cmd] = err
}

func (m *SwarmMockExecutor) GetRunCommands() []string {
	return m.runCommands
}

// Tests for swarm.go (Manager)
func TestNewSwarmManager(t *testing.T) {
	mockExec := NewSwarmMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	if manager == nil {
		t.Error("NewManager() should not return nil")
	}

	if manager.stackName != "test-stack" {
		t.Errorf("stackName = %q, want %q", manager.stackName, "test-stack")
	}
}

func TestManager_IsSwarmInitialized_Active(t *testing.T) {
	mockExec := NewSwarmMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	mockExec.SetRunResult("docker info --format '{{.Swarm.LocalNodeState}}'", &executor.CommandResult{
		Stdout:   "active",
		ExitCode: 0,
	})

	initialized, err := manager.IsSwarmInitialized()
	if err != nil {
		t.Errorf("IsSwarmInitialized() error = %v", err)
	}

	if !initialized {
		t.Error("IsSwarmInitialized() should return true when state is active")
	}
}

func TestManager_IsSwarmInitialized_Inactive(t *testing.T) {
	mockExec := NewSwarmMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	mockExec.SetRunResult("docker info --format '{{.Swarm.LocalNodeState}}'", &executor.CommandResult{
		Stdout:   "inactive",
		ExitCode: 0,
	})

	initialized, err := manager.IsSwarmInitialized()
	if err != nil {
		t.Errorf("IsSwarmInitialized() error = %v", err)
	}

	if initialized {
		t.Error("IsSwarmInitialized() should return false when state is inactive")
	}
}

func TestManager_InitSwarm_Success(t *testing.T) {
	mockExec := NewSwarmMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	mockExec.SetRunResult("docker swarm init", &executor.CommandResult{
		Stdout:   "Swarm initialized",
		ExitCode: 0,
	})

	err := manager.InitSwarm()
	if err != nil {
		t.Errorf("InitSwarm() error = %v", err)
	}

	commands := mockExec.GetRunCommands()
	expectedCmd := "docker swarm init"
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

func TestManager_InitSwarm_Failure(t *testing.T) {
	mockExec := NewSwarmMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	mockExec.SetRunResult("docker swarm init", &executor.CommandResult{
		Stdout:   "",
		Stderr:   "This node is already part of a swarm",
		ExitCode: 1,
	})

	err := manager.InitSwarm()
	if err == nil {
		t.Error("InitSwarm() should return error when swarm init fails")
	}

	if !containsString(err.Error(), "swarm init failed") {
		t.Errorf("Error should contain 'swarm init failed': %v", err)
	}
}

func TestManager_CreateNetwork_Exists(t *testing.T) {
	mockExec := NewSwarmMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	checkCmd := "docker network ls --filter name=^test-net$ --format '{{.Name}}'"
	mockExec.SetRunResult(checkCmd, &executor.CommandResult{
		Stdout:   "test-net",
		ExitCode: 0,
	})

	err := manager.CreateNetwork("test-net")
	if err != nil {
		t.Errorf("CreateNetwork() error = %v", err)
	}

	// Should not try to create network
	commands := mockExec.GetRunCommands()
	createCmd := "docker network create --driver overlay --attachable test-net"
	for _, cmd := range commands {
		if cmd == createCmd {
			t.Error("Should not try to create network when it already exists")
			break
		}
	}
}

func TestManager_CreateNetwork_New(t *testing.T) {
	mockExec := NewSwarmMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	checkCmd := "docker network ls --filter name=^test-net$ --format '{{.Name}}'"
	mockExec.SetRunResult(checkCmd, &executor.CommandResult{
		Stdout:   "",
		ExitCode: 0,
	})

	createCmd := "docker network create --driver overlay --attachable test-net"
	mockExec.SetRunResult(createCmd, &executor.CommandResult{
		Stdout:   "abc123",
		ExitCode: 0,
	})

	err := manager.CreateNetwork("test-net")
	if err != nil {
		t.Errorf("CreateNetwork() error = %v", err)
	}

	commands := mockExec.GetRunCommands()
	found := false
	for _, cmd := range commands {
		if cmd == createCmd {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected command %q not found", createCmd)
	}
}

func TestManager_CreateNetwork_CreateError(t *testing.T) {
	mockExec := NewSwarmMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	checkCmd := "docker network ls --filter name=^test-net$ --format '{{.Name}}'"
	mockExec.SetRunResult(checkCmd, &executor.CommandResult{
		Stdout:   "",
		ExitCode: 0,
	})

	testErr := errors.New("network create failed")
	createCmd := "docker network create --driver overlay --attachable test-net"
	mockExec.SetRunError(createCmd, testErr)

	err := manager.CreateNetwork("test-net")
	if err == nil {
		t.Error("CreateNetwork() should return error when network creation fails")
	}

	if !containsString(err.Error(), "failed to create network") {
		t.Errorf("Error should contain 'failed to create network': %v", err)
	}
}

func TestManager_RegistryLogin_WithCredentials(t *testing.T) {
	mockExec := NewSwarmMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	cmd := "echo 'password123' | docker login registry.example.com -u user1 --password-stdin"
	mockExec.SetRunResult(cmd, &executor.CommandResult{
		Stdout:   "Login Succeeded",
		ExitCode: 0,
	})

	err := manager.RegistryLogin("registry.example.com", "user1", "password123")
	if err != nil {
		t.Errorf("RegistryLogin() error = %v", err)
	}

	commands := mockExec.GetRunCommands()
	found := false
	for _, runCmd := range commands {
		if runCmd == cmd {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected command %q not found", cmd)
	}
}

func TestManager_RegistryLogin_EmptyCredentials(t *testing.T) {
	mockExec := NewSwarmMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	err := manager.RegistryLogin("registry.example.com", "", "")
	if err != nil {
		t.Errorf("RegistryLogin() should not error with empty credentials: %v", err)
	}

	commands := mockExec.GetRunCommands()
	if len(commands) > 0 {
		t.Error("Should not execute any commands with empty credentials")
	}
}

func TestManager_GetCurrentNodeHostname(t *testing.T) {
	mockExec := NewSwarmMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	expectedHostname := "manager-node-1"
	mockExec.SetRunResult("docker node inspect self --format '{{.Description.Hostname}}'", &executor.CommandResult{
		Stdout:   expectedHostname,
		ExitCode: 0,
	})

	hostname, err := manager.GetCurrentNodeHostname()
	if err != nil {
		t.Errorf("GetCurrentNodeHostname() error = %v", err)
	}

	if hostname != expectedHostname {
		t.Errorf("GetCurrentNodeHostname() = %q, want %q", hostname, expectedHostname)
	}
}

func TestManager_GetCurrentNodeHostname_Empty(t *testing.T) {
	mockExec := NewSwarmMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	mockExec.SetRunResult("docker node inspect self --format '{{.Description.Hostname}}'", &executor.CommandResult{
		Stdout:   "",
		ExitCode: 0,
	})

	_, err := manager.GetCurrentNodeHostname()
	if err == nil {
		t.Error("GetCurrentNodeHostname() should return error when hostname is empty")
	}

	if !containsString(err.Error(), "empty hostname") {
		t.Errorf("Error should contain 'empty hostname': %v", err)
	}
}

// Tests for stack.go
func TestManager_DeployStack_Success(t *testing.T) {
	mockExec := NewSwarmMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	cmd := "docker stack deploy -c /tmp/test-stack-compose.yaml test-stack --with-registry-auth"
	mockExec.SetRunResult(cmd, &executor.CommandResult{
		Stdout:   "Stack deployed",
		ExitCode: 0,
	})

	composeContent := []byte("version: '3.8'\nservices:\n  web:\n    image: nginx:latest")

	err := manager.DeployStack(composeContent)
	if err != nil {
		t.Errorf("DeployStack() error = %v", err)
	}

	commands := mockExec.GetRunCommands()
	if !containsCommand(commands, cmd) {
		t.Error("Deploy command should be executed")
	}
}

func TestManager_DeployStack_Failure(t *testing.T) {
	mockExec := NewSwarmMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	cmd := "docker stack deploy -c /tmp/test-stack-compose.yaml test-stack --with-registry-auth"
	mockExec.SetRunResult(cmd, &executor.CommandResult{
		Stdout:   "",
		Stderr:   "Deploy failed",
		ExitCode: 1,
	})

	composeContent := []byte("version: '3.8'\nservices:\n  web:\n    image: nginx:latest")

	err := manager.DeployStack(composeContent)
	if err == nil {
		t.Error("DeployStack() should return error when deploy fails")
	}

	if !containsString(err.Error(), "stack deploy failed") {
		t.Errorf("Error should contain 'stack deploy failed': %v", err)
	}
}

func TestManager_RemoveStack_Success(t *testing.T) {
	mockExec := NewSwarmMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	cmd := "docker stack rm test-stack"
	mockExec.SetRunResult(cmd, &executor.CommandResult{
		ExitCode: 0,
	})

	err := manager.RemoveStack()
	if err != nil {
		t.Errorf("RemoveStack() error = %v", err)
	}

	commands := mockExec.GetRunCommands()
	if !containsCommand(commands, cmd) {
		t.Error("Remove command should be executed")
	}
}

func TestManager_RemoveStack_Failure(t *testing.T) {
	mockExec := NewSwarmMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	cmd := "docker stack rm test-stack"
	mockExec.SetRunResult(cmd, &executor.CommandResult{
		Stdout:   "",
		Stderr:   "Stack not found",
		ExitCode: 1,
	})

	err := manager.RemoveStack()
	if err == nil {
		t.Error("RemoveStack() should return error when remove fails")
	}

	if !containsString(err.Error(), "stack removal failed") {
		t.Errorf("Error should contain 'stack removal failed': %v", err)
	}
}

// Tests for service.go
func TestManager_ListServices_ServicesExist(t *testing.T) {
	mockExec := NewSwarmMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	cmd := "docker stack services test-stack --format '{{.Name}}|{{.Mode}}|{{.Replicas}}|{{.Image}}|{{.Ports}}'"
	mockExec.SetRunResult(cmd, &executor.CommandResult{
		Stdout:   "test-stack_web|replicated|1/1|nginx:latest|80->80/tcp\ntest-stack_db|replicated|1/1|postgres:13|5432/tcp",
		ExitCode: 0,
	})

	services, err := manager.ListServices()
	if err != nil {
		t.Errorf("ListServices() error = %v", err)
	}

	if len(services) != 2 {
		t.Errorf("Expected 2 services, got %d", len(services))
	}

	if services[0].Name != "test-stack_web" {
		t.Errorf("First service name = %q, want %q", services[0].Name, "test-stack_web")
	}
}

func TestManager_ListServices_NoServices(t *testing.T) {
	mockExec := NewSwarmMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	cmd := "docker stack services test-stack --format '{{.Name}}|{{.Mode}}|{{.Replicas}}|{{.Image}}|{{.Ports}}'"
	mockExec.SetRunResult(cmd, &executor.CommandResult{
		Stdout:   "",
		ExitCode: 0,
	})

	services, err := manager.ListServices()
	if err != nil {
		t.Errorf("ListServices() error = %v", err)
	}

	if len(services) != 0 {
		t.Errorf("Expected 0 services, got %d", len(services))
	}
}

func TestManager_ListServices_Error(t *testing.T) {
	mockExec := NewSwarmMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	cmd := "docker stack services test-stack --format '{{.Name}}|{{.Mode}}|{{.Replicas}}|{{.Image}}|{{.Ports}}'"
	testErr := errors.New("command failed")
	mockExec.SetRunError(cmd, testErr)

	_, err := manager.ListServices()
	if err == nil {
		t.Error("ListServices() should return error when command fails")
	}

	if !errors.Is(err, testErr) {
		t.Errorf("Expected error %v, got %v", testErr, err)
	}
}

// Tests for status.go
func TestManager_GetStackTasks_TasksExist(t *testing.T) {
	mockExec := NewSwarmMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	cmd := "docker stack ps test-stack --format '{{.ID}}|{{.Name}}|{{.Image}}|{{.Node}}|{{.DesiredState}}|{{.CurrentState}}|{{.Error}}'"
	mockExec.SetRunResult(cmd, &executor.CommandResult{
		Stdout:   "task123|test-stack_web.1|nginx:latest|manager1|running|running|",
		ExitCode: 0,
	})

	tasks, err := manager.GetStackTasks()
	if err != nil {
		t.Errorf("GetStackTasks() error = %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}

	if tasks[0].ID != "task123" {
		t.Errorf("Task ID = %q, want %q", tasks[0].ID, "task123")
	}
}

func TestManager_GetStackTasks_NoTasks(t *testing.T) {
	mockExec := NewSwarmMockExecutor()
	manager := NewManager(mockExec, "test-stack")

	cmd := "docker stack ps test-stack --format '{{.ID}}|{{.Name}}|{{.Image}}|{{.Node}}|{{.DesiredState}}|{{.CurrentState}}|{{.Error}}'"
	mockExec.SetRunResult(cmd, &executor.CommandResult{
		Stdout:   "",
		ExitCode: 0,
	})

	tasks, err := manager.GetStackTasks()
	if err != nil {
		t.Errorf("GetStackTasks() error = %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(tasks))
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
