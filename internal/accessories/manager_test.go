package accessories

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/marcelsud/swarmctl/internal/config"
	"github.com/marcelsud/swarmctl/internal/executor"
)

// AccessoriesMockExecutor for testing accessories
type AccessoriesMockExecutor struct {
	runCommands []string
	runResults  map[string]*executor.CommandResult
	runErrors   map[string]error
}

func NewAccessoriesMockExecutor() *AccessoriesMockExecutor {
	return &AccessoriesMockExecutor{
		runCommands: make([]string, 0),
		runResults:  make(map[string]*executor.CommandResult),
		runErrors:   make(map[string]error),
	}
}

func (m *AccessoriesMockExecutor) Run(cmd string) (*executor.CommandResult, error) {
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

func (m *AccessoriesMockExecutor) RunInteractive(cmd string) error {
	m.runCommands = append(m.runCommands, cmd)
	return nil
}

func (m *AccessoriesMockExecutor) RunStream(cmd string, stdout, stderr io.Writer) error {
	m.runCommands = append(m.runCommands, cmd)
	return nil
}

func (m *AccessoriesMockExecutor) WriteFile(path string, content []byte) error {
	return nil
}

func (m *AccessoriesMockExecutor) Close() error {
	return nil
}

func (m *AccessoriesMockExecutor) IsLocal() bool {
	return true
}

func (m *AccessoriesMockExecutor) SetRunResult(cmd string, result *executor.CommandResult) {
	m.runResults[cmd] = result
}

func (m *AccessoriesMockExecutor) SetRunError(cmd string, err error) {
	m.runErrors[cmd] = err
}

func (m *AccessoriesMockExecutor) GetRunCommands() []string {
	return m.runCommands
}

func TestAccessoriesNewManager(t *testing.T) {
	mockExec := NewAccessoriesMockExecutor()
	manager := NewManager(mockExec, "test-stack", config.ModeSwarm)

	if manager == nil {
		t.Error("NewManager() should not return nil")
	}

	if manager.stackName != "test-stack" {
		t.Errorf("stackName = %q, want %q", manager.stackName, "test-stack")
	}

	if manager.mode != config.ModeSwarm {
		t.Errorf("mode = %q, want %q", manager.mode, config.ModeSwarm)
	}
}

func TestManager_Start_InvalidName(t *testing.T) {
	mockExec := NewAccessoriesMockExecutor()
	manager := NewManager(mockExec, "test-stack", config.ModeSwarm)

	err := manager.Start("redis; rm -rf /")
	if err == nil {
		t.Error("Start() should reject invalid name with shell metacharacters")
	}

	if !containsString(err.Error(), "must contain only alphanumeric characters") {
		t.Errorf("Expected validation error, got: %v", err)
	}
}

func TestManager_Stop_InvalidName(t *testing.T) {
	mockExec := NewAccessoriesMockExecutor()
	manager := NewManager(mockExec, "test-stack", config.ModeSwarm)

	err := manager.Stop("postgres && curl evil.com")
	if err == nil {
		t.Error("Stop() should reject invalid name with shell metacharacters")
	}

	if !containsString(err.Error(), "must contain only alphanumeric characters") {
		t.Errorf("Expected validation error, got: %v", err)
	}
}

func TestManager_Restart_InvalidName(t *testing.T) {
	mockExec := NewAccessoriesMockExecutor()
	manager := NewManager(mockExec, "test-stack", config.ModeSwarm)

	err := manager.Restart("redis`id`")
	if err == nil {
		t.Error("Restart() should reject invalid name with backticks")
	}

	if !containsString(err.Error(), "must contain only alphanumeric characters") {
		t.Errorf("Expected validation error, got: %v", err)
	}
}

func TestManager_GetStatus_InvalidName(t *testing.T) {
	mockExec := NewAccessoriesMockExecutor()
	manager := NewManager(mockExec, "test-stack", config.ModeSwarm)

	_, err := manager.GetStatus("redis; cat /etc/passwd")
	if err == nil {
		t.Error("GetStatus() should reject invalid name with shell injection")
	}

	if !containsString(err.Error(), "must contain only alphanumeric characters") {
		t.Errorf("Expected validation error, got: %v", err)
	}
}

func TestAccessoriesNewManager_ComposeMode(t *testing.T) {
	mockExec := NewAccessoriesMockExecutor()
	manager := NewManager(mockExec, "test-stack", config.ModeCompose)

	if manager.mode != config.ModeCompose {
		t.Errorf("mode = %q, want %q", manager.mode, config.ModeCompose)
	}
}

func TestManager_Start_ComposeMode(t *testing.T) {
	mockExec := NewAccessoriesMockExecutor()
	manager := NewManager(mockExec, "test-stack", config.ModeCompose)

	cmd := "docker compose -p test-stack start redis"
	mockExec.SetRunResult(cmd, &executor.CommandResult{
		ExitCode: 0,
	})

	err := manager.Start("redis")
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}

	commands := mockExec.GetRunCommands()
	if !containsCommand(commands, cmd) {
		t.Error("Compose start command should be executed")
	}
}

func TestManager_Start_SwarmMode(t *testing.T) {
	mockExec := NewAccessoriesMockExecutor()
	manager := NewManager(mockExec, "test-stack", config.ModeSwarm)

	cmd := "docker service scale test-stack_redis=1"
	mockExec.SetRunResult(cmd, &executor.CommandResult{
		ExitCode: 0,
	})

	err := manager.Start("redis")
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}

	commands := mockExec.GetRunCommands()
	if !containsCommand(commands, cmd) {
		t.Error("Swarm start command should be executed")
	}
}

func TestManager_Start_Failure(t *testing.T) {
	mockExec := NewAccessoriesMockExecutor()
	manager := NewManager(mockExec, "test-stack", config.ModeCompose)

	cmd := "docker compose -p test-stack start redis"
	mockExec.SetRunResult(cmd, &executor.CommandResult{
		Stdout:   "",
		Stderr:   "service not found",
		ExitCode: 1,
	})

	err := manager.Start("redis")
	if err == nil {
		t.Error("Start() should return error when command fails")
	}

	if !containsString(err.Error(), "start failed") {
		t.Errorf("Error should contain 'start failed': %v", err)
	}
}

func TestManager_Stop_ComposeMode(t *testing.T) {
	mockExec := NewAccessoriesMockExecutor()
	manager := NewManager(mockExec, "test-stack", config.ModeCompose)

	cmd := "docker compose -p test-stack stop redis"
	mockExec.SetRunResult(cmd, &executor.CommandResult{
		ExitCode: 0,
	})

	err := manager.Stop("redis")
	if err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	commands := mockExec.GetRunCommands()
	if !containsCommand(commands, cmd) {
		t.Error("Compose stop command should be executed")
	}
}

func TestManager_Stop_SwarmMode(t *testing.T) {
	mockExec := NewAccessoriesMockExecutor()
	manager := NewManager(mockExec, "test-stack", config.ModeSwarm)

	cmd := "docker service scale test-stack_redis=0"
	mockExec.SetRunResult(cmd, &executor.CommandResult{
		ExitCode: 0,
	})

	err := manager.Stop("redis")
	if err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	commands := mockExec.GetRunCommands()
	if !containsCommand(commands, cmd) {
		t.Error("Swarm stop command should be executed")
	}
}

func TestManager_Stop_Failure(t *testing.T) {
	mockExec := NewAccessoriesMockExecutor()
	manager := NewManager(mockExec, "test-stack", config.ModeCompose)

	cmd := "docker compose -p test-stack stop redis"
	mockExec.SetRunResult(cmd, &executor.CommandResult{
		Stdout:   "",
		Stderr:   "service not found",
		ExitCode: 1,
	})

	err := manager.Stop("redis")
	if err == nil {
		t.Error("Stop() should return error when command fails")
	}

	if !containsString(err.Error(), "stop failed") {
		t.Errorf("Error should contain 'stop failed': %v", err)
	}
}

func TestManager_Restart_ComposeMode(t *testing.T) {
	mockExec := NewAccessoriesMockExecutor()
	manager := NewManager(mockExec, "test-stack", config.ModeCompose)

	cmd := "docker compose -p test-stack restart redis"
	mockExec.SetRunResult(cmd, &executor.CommandResult{
		ExitCode: 0,
	})

	err := manager.Restart("redis")
	if err != nil {
		t.Errorf("Restart() error = %v", err)
	}

	commands := mockExec.GetRunCommands()
	if !containsCommand(commands, cmd) {
		t.Error("Compose restart command should be executed")
	}
}

func TestManager_Restart_SwarmMode(t *testing.T) {
	mockExec := NewAccessoriesMockExecutor()
	manager := NewManager(mockExec, "test-stack", config.ModeSwarm)

	cmd := "docker service update --force test-stack_redis"
	mockExec.SetRunResult(cmd, &executor.CommandResult{
		ExitCode: 0,
	})

	err := manager.Restart("redis")
	if err != nil {
		t.Errorf("Restart() error = %v", err)
	}

	commands := mockExec.GetRunCommands()
	if !containsCommand(commands, cmd) {
		t.Error("Swarm restart command should be executed")
	}
}

func TestManager_Restart_Failure(t *testing.T) {
	mockExec := NewAccessoriesMockExecutor()
	manager := NewManager(mockExec, "test-stack", config.ModeCompose)

	cmd := "docker compose -p test-stack restart redis"
	mockExec.SetRunResult(cmd, &executor.CommandResult{
		Stdout:   "",
		Stderr:   "service not found",
		ExitCode: 1,
	})

	err := manager.Restart("redis")
	if err == nil {
		t.Error("Restart() should return error when command fails")
	}

	if !containsString(err.Error(), "restart failed") {
		t.Errorf("Error should contain 'restart failed': %v", err)
	}
}

func TestManager_GetStatus_ComposeMode_Running(t *testing.T) {
	mockExec := NewAccessoriesMockExecutor()
	manager := NewManager(mockExec, "test-stack", config.ModeCompose)

	cmd := "docker compose -p test-stack ps redis --format json"
	mockExec.SetRunResult(cmd, &executor.CommandResult{
		Stdout:   `{"State":"running","Status":"Up 2 minutes"}`,
		ExitCode: 0,
	})

	status, err := manager.GetStatus("redis")
	if err != nil {
		t.Errorf("GetStatus() error = %v", err)
	}

	if status.Name != "redis" {
		t.Errorf("Name = %q, want %q", status.Name, "redis")
	}

	if !status.Running {
		t.Error("Running should be true when container is running")
	}
}

func TestManager_GetStatus_ComposeMode_NotRunning(t *testing.T) {
	mockExec := NewAccessoriesMockExecutor()
	manager := NewManager(mockExec, "test-stack", config.ModeCompose)

	cmd := "docker compose -p test-stack ps redis --format json"
	mockExec.SetRunResult(cmd, &executor.CommandResult{
		Stdout:   `{"State":"exited","Status":"Exited 5 minutes ago"}`,
		ExitCode: 0,
	})

	status, err := manager.GetStatus("redis")
	if err != nil {
		t.Errorf("GetStatus() error = %v", err)
	}

	if status.Running {
		t.Error("Running should be false when container is not running")
	}
}

func TestManager_GetStatus_ComposeMode_EmptyOutput(t *testing.T) {
	mockExec := NewAccessoriesMockExecutor()
	manager := NewManager(mockExec, "test-stack", config.ModeCompose)

	cmd := "docker compose -p test-stack ps redis --format json"
	mockExec.SetRunResult(cmd, &executor.CommandResult{
		Stdout:   "",
		ExitCode: 0,
	})

	_, err := manager.GetStatus("redis")
	if err == nil {
		t.Error("GetStatus() should return error when output is empty")
	}

	if !containsString(err.Error(), "not found") {
		t.Errorf("Error should contain 'not found': %v", err)
	}
}

func TestManager_GetStatus_SwarmMode_Running(t *testing.T) {
	mockExec := NewAccessoriesMockExecutor()
	manager := NewManager(mockExec, "test-stack", config.ModeSwarm)

	cmd := "docker service ls --filter name=test-stack_redis --format '{{.Name}}|{{.Replicas}}'"
	mockExec.SetRunResult(cmd, &executor.CommandResult{
		Stdout:   "test-stack_redis|2/2",
		ExitCode: 0,
	})

	status, err := manager.GetStatus("redis")
	if err != nil {
		t.Errorf("GetStatus() error = %v", err)
	}

	if status.Name != "redis" {
		t.Errorf("Name = %q, want %q", status.Name, "redis")
	}

	if !status.Running {
		t.Error("Running should be true when replicas are running")
	}

	if status.Replicas != "2/2" {
		t.Errorf("Replicas = %q, want %q", status.Replicas, "2/2")
	}
}

func TestManager_GetStatus_SwarmMode_NotRunning(t *testing.T) {
	mockExec := NewAccessoriesMockExecutor()
	manager := NewManager(mockExec, "test-stack", config.ModeSwarm)

	cmd := "docker service ls --filter name=test-stack_redis --format '{{.Name}}|{{.Replicas}}'"
	mockExec.SetRunResult(cmd, &executor.CommandResult{
		Stdout:   "test-stack_redis|0/2",
		ExitCode: 0,
	})

	status, err := manager.GetStatus("redis")
	if err != nil {
		t.Errorf("GetStatus() error = %v", err)
	}

	if status.Running {
		t.Error("Running should be false when replicas are 0/2")
	}

	if status.Replicas != "0/2" {
		t.Errorf("Replicas = %q, want %q", status.Replicas, "0/2")
	}
}

func TestManager_GetStatus_SwarmMode_NotFound(t *testing.T) {
	mockExec := NewAccessoriesMockExecutor()
	manager := NewManager(mockExec, "test-stack", config.ModeSwarm)

	cmd := "docker service ls --filter name=test-stack_redis --format '{{.Name}}|{{.Replicas}}'"
	mockExec.SetRunResult(cmd, &executor.CommandResult{
		Stdout:   "",
		ExitCode: 0,
	})

	_, err := manager.GetStatus("redis")
	if err == nil {
		t.Error("GetStatus() should return error when service not found")
	}

	if !containsString(err.Error(), "not found") {
		t.Errorf("Error should contain 'not found': %v", err)
	}
}

func TestManager_GetStatus_SwarmMode_ParseError(t *testing.T) {
	mockExec := NewAccessoriesMockExecutor()
	manager := NewManager(mockExec, "test-stack", config.ModeSwarm)

	cmd := "docker service ls --filter name=test-stack_redis --format '{{.Name}}|{{.Replicas}}'"
	mockExec.SetRunResult(cmd, &executor.CommandResult{
		Stdout:   "test-stack_redis", // Missing replicas part
		ExitCode: 0,
	})

	_, err := manager.GetStatus("redis")
	if err == nil {
		t.Error("GetStatus() should return error when parsing fails")
	}

	if !containsString(err.Error(), "failed to parse status") {
		t.Errorf("Error should contain 'failed to parse status': %v", err)
	}
}

func TestManager_ListAll_EmptyList(t *testing.T) {
	mockExec := NewAccessoriesMockExecutor()
	manager := NewManager(mockExec, "test-stack", config.ModeSwarm)

	accessoryNames := []string{"redis", "postgres"}

	// Mock GetStatus calls to return not found for all services
	redisCmd := "docker service ls --filter name=test-stack_redis --format '{{.Name}}|{{.Replicas}}'"
	postgresCmd := "docker service ls --filter name=test-stack_postgres --format '{{.Name}}|{{.Replicas}}'"
	mockExec.SetRunResult(redisCmd, &executor.CommandResult{
		Stdout:   "",
		ExitCode: 0,
	})
	mockExec.SetRunResult(postgresCmd, &executor.CommandResult{
		Stdout:   "",
		ExitCode: 0,
	})

	statuses, err := manager.ListAll(accessoryNames)
	if err != nil {
		t.Errorf("ListAll() error = %v", err)
	}

	if len(statuses) != 2 {
		t.Errorf("Expected 2 statuses, got %d", len(statuses))
	}

	// All should be not deployed since mock returns empty output
	for _, status := range statuses {
		if status.Replicas != "not deployed" {
			t.Errorf("Expected 'not deployed', got %q", status.Replicas)
		}
		if status.Running {
			t.Error("Running should be false for not deployed services")
		}
	}
}

func TestManager_ListAll_WithServices(t *testing.T) {
	mockExec := NewAccessoriesMockExecutor()
	manager := NewManager(mockExec, "test-stack", config.ModeSwarm)

	accessoryNames := []string{"redis", "postgres"}

	// Mock GetStatus calls
	redisCmd := "docker service ls --filter name=test-stack_redis --format '{{.Name}}|{{.Replicas}}'"
	postgresCmd := "docker service ls --filter name=test-stack_postgres --format '{{.Name}}|{{.Replicas}}'"
	mockExec.SetRunResult(redisCmd, &executor.CommandResult{
		Stdout:   "test-stack_redis|1/1",
		ExitCode: 0,
	})
	mockExec.SetRunResult(postgresCmd, &executor.CommandResult{
		Stdout:   "test-stack_postgres|1/1",
		ExitCode: 0,
	})

	statuses, err := manager.ListAll(accessoryNames)
	if err != nil {
		t.Errorf("ListAll() error = %v", err)
	}

	if len(statuses) != 2 {
		t.Errorf("Expected 2 statuses, got %d", len(statuses))
	}

	// Check redis
	redisStatus := findStatus(statuses, "redis")
	if redisStatus == nil {
		t.Error("Redis status should be present")
	} else {
		if !redisStatus.Running {
			t.Error("Redis should be running")
		}
		if redisStatus.Replicas != "1/1" {
			t.Errorf("Redis replicas = %q, want '1/1'", redisStatus.Replicas)
		}
	}

	// Check postgres
	postgresStatus := findStatus(statuses, "postgres")
	if postgresStatus == nil {
		t.Error("Postgres status should be present")
	} else {
		if !postgresStatus.Running {
			t.Error("Postgres should be running")
		}
		if postgresStatus.Replicas != "1/1" {
			t.Errorf("Postgres replicas = %q, want '1/1'", postgresStatus.Replicas)
		}
	}
}

func TestManager_ListAll_GetStatusError(t *testing.T) {
	mockExec := NewAccessoriesMockExecutor()
	manager := NewManager(mockExec, "test-stack", config.ModeSwarm)

	accessoryNames := []string{"redis", "postgres"}

	// Mock GetStatus to return error for redis
	redisCmd := "docker service ls --filter name=test-stack_redis --format '{{.Name}}|{{.Replicas}}'"
	mockExec.SetRunError(redisCmd, errors.New("service failed"))

	statuses, err := manager.ListAll(accessoryNames)

	// ListAll never returns errors, it swallows them and returns "not deployed" status
	if err != nil {
		t.Errorf("ListAll() should not return error even when GetStatus fails, got: %v", err)
	}

	// Should still return statuses for all accessories, with failed ones marked as "not deployed"
	if len(statuses) != 2 {
		t.Errorf("Expected 2 statuses, got %d", len(statuses))
	}

	// Check that redis is marked as "not deployed" due to the error
	var redisStatus *AccessoryStatus
	for _, status := range statuses {
		if status.Name == "redis" {
			redisStatus = &status
			break
		}
	}

	if redisStatus == nil {
		t.Error("Expected to find redis status")
	} else if redisStatus.Replicas != "not deployed" || redisStatus.Running {
		t.Errorf("Expected redis to be marked as not deployed, got replicas=%q, running=%v",
			redisStatus.Replicas, redisStatus.Running)
	}
}

func TestManager_CommandEscaping(t *testing.T) {
	mockExec := NewAccessoriesMockExecutor()
	manager := NewManager(mockExec, "test-stack", config.ModeSwarm)

	evilName := "redis; rm -rf /tmp"
	cmd := "docker service scale test-stack_redis; rm -rf /tmp=1"
	mockExec.SetRunResult(cmd, &executor.CommandResult{ExitCode: 0})

	err := manager.Start(evilName)
	if err == nil {
		t.Error("Start() should reject invalid name")
	}

	validName := "cache_redis"
	mockExec.SetRunResult("docker service scale test-stack_cache_redis=1", &executor.CommandResult{ExitCode: 0})

	err = manager.Start(validName)
	if err != nil {
		t.Errorf("Start() should accept valid name with underscores, got error: %v", err)
	}
}

func TestManager_ValidationEdgeCases(t *testing.T) {
	mockExec := NewAccessoriesMockExecutor()
	manager := NewManager(mockExec, "test-stack", config.ModeSwarm)

	testCases := []struct {
		name          string
		shouldBeValid bool
	}{
		{"redis", true},
		{"redis_01", true},
		{"cache_redis", true},
		{"redis123", true},
		{"r", true},
		{"", false},
		{"-invalid", false},
		{"cache.redis", false},
		{"cache-redis", false},
		{"a" + strings.Repeat("b", 100), false},
		{"redis;", false},
		{"redis &&", false},
		{"redis|", false},
		{"redis`", false},
		{"redis$(id)", false},
	}

	for _, tc := range testCases {
		err := manager.Start(tc.name)
		if tc.shouldBeValid && err != nil {
			t.Errorf("Expected name '%s' to be valid, got error: %v", tc.name, err)
		}
		if !tc.shouldBeValid && err == nil {
			t.Errorf("Expected name '%s' to be invalid", tc.name)
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
	return strings.Contains(s, substr)
}

func findStatus(statuses []AccessoryStatus, name string) *AccessoryStatus {
	for _, status := range statuses {
		if status.Name == name {
			return &status
		}
	}
	return nil
}
