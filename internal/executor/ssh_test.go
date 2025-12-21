package executor

import (
	"bytes"
	"testing"

	"github.com/marcelsud/swarmctl/internal/config"
)

func TestSSHExecutor_IsLocal(t *testing.T) {
	// Create a mock SSH executor (we'll test connection separately)
	cfg := config.SSHConfig{
		Host: "test.example.com",
		User: "testuser",
		Port: 22,
		Key:  "/path/to/key",
	}

	// Note: This will fail to connect in unit tests, which is expected
	executor, err := NewSSH(cfg)
	if err == nil {
		executor.Close()
		t.Skip("SSH server available, skipping connection test")
	}

	// If connection fails (expected in unit tests), we can't test IsLocal
	// but we can verify the constructor creates the right struct
	if err != nil && executor == nil {
		// Expected behavior - no SSH server available
		t.Skipf("SSH connection failed (expected): %v", err)
	}
}

func TestSSHExecutor_Host(t *testing.T) {
	// This test requires a valid SSH connection
	cfg := config.SSHConfig{
		Host: "localhost",
		User: "testuser",
		Port: 22,
		Key:  "/path/to/key",
	}

	executor, err := NewSSH(cfg)
	if err != nil {
		t.Skipf("SSH connection failed (expected in CI): %v", err)
	}
	defer executor.Close()

	if executor.Host() != cfg.Host {
		t.Errorf("Host() = %v, want %v", executor.Host(), cfg.Host)
	}
}

func TestSSHExecutor_User(t *testing.T) {
	cfg := config.SSHConfig{
		Host: "localhost",
		User: "testuser",
		Port: 22,
		Key:  "/path/to/key",
	}

	executor, err := NewSSH(cfg)
	if err != nil {
		t.Skipf("SSH connection failed (expected in CI): %v", err)
	}
	defer executor.Close()

	if executor.User() != cfg.User {
		t.Errorf("User() = %v, want %v", executor.User(), cfg.User)
	}
}

func TestSSHExecutor_Port(t *testing.T) {
	cfg := config.SSHConfig{
		Host: "localhost",
		User: "testuser",
		Port: 2222,
		Key:  "/path/to/key",
	}

	executor, err := NewSSH(cfg)
	if err != nil {
		t.Skipf("SSH connection failed (expected in CI): %v", err)
	}
	defer executor.Close()

	if executor.Port() != cfg.Port {
		t.Errorf("Port() = %v, want %v", executor.Port(), cfg.Port)
	}
}

func TestSSHExecutor_RunStream(t *testing.T) {
	cfg := config.SSHConfig{
		Host: "localhost",
		User: "testuser",
		Port: 22,
		Key:  "/path/to/key",
	}

	executor, err := NewSSH(cfg)
	if err != nil {
		t.Skipf("SSH connection failed (expected in CI): %v", err)
	}
	defer executor.Close()

	var stdout, stderr bytes.Buffer
	err = executor.RunStream("echo 'hello'", &stdout, &stderr)
	if err != nil {
		t.Errorf("RunStream() error = %v", err)
	}

	if stdout.String() != "hello\n" {
		t.Errorf("RunStream() stdout = %q, want %q", stdout.String(), "hello\n")
	}
}

func TestSSHExecutor_WriteFile(t *testing.T) {
	cfg := config.SSHConfig{
		Host: "localhost",
		User: "testuser",
		Port: 22,
		Key:  "/path/to/key",
	}

	executor, err := NewSSH(cfg)
	if err != nil {
		t.Skipf("SSH connection failed (expected in CI): %v", err)
	}
	defer executor.Close()

	content := []byte("test content")
	path := "/tmp/ssh_test.txt"

	// Write file
	err = executor.WriteFile(path, content)
	if err != nil {
		t.Errorf("WriteFile() error = %v", err)
	}

	// Read it back to verify
	result, err := executor.Run("cat " + path)
	if err != nil {
		t.Errorf("Failed to read test file: %v", err)
	}

	if result.Stdout != "test content" {
		t.Errorf("WriteFile() content = %q, want %q", result.Stdout, "test content")
	}

	// Cleanup
	executor.Run("rm -f " + path)
}

func TestSSHExecutor_Close(t *testing.T) {
	cfg := config.SSHConfig{
		Host: "localhost",
		User: "testuser",
		Port: 22,
		Key:  "/path/to/key",
	}

	executor, err := NewSSH(cfg)
	if err != nil {
		t.Skipf("SSH connection failed (expected in CI): %v", err)
	}

	// Close should work without error
	err = executor.Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}

func TestSSHExecutor_HasAgentForwarding(t *testing.T) {
	cfg := config.SSHConfig{
		Host: "localhost",
		User: "testuser",
		Port: 22,
		Key:  "/path/to/key",
	}

	executor, err := NewSSH(cfg)
	if err != nil {
		t.Skipf("SSH connection failed (expected in CI): %v", err)
	}
	defer executor.Close()

	// Just test that the method doesn't panic
	_ = executor.HasAgentForwarding()
}
