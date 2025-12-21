package ssh

import (
	"os"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test.example.com", 2222, "testuser", "/path/to/key")

	if client.Host != "test.example.com" {
		t.Errorf("Host = %q, want %q", client.Host, "test.example.com")
	}

	if client.Port != 2222 {
		t.Errorf("Port = %d, want %d", client.Port, 2222)
	}

	if client.User != "testuser" {
		t.Errorf("User = %q, want %q", client.User, "testuser")
	}

	if client.KeyPath != "/path/to/key" {
		t.Errorf("KeyPath = %q, want %q", client.KeyPath, "/path/to/key")
	}

	if client.conn != nil {
		t.Error("conn should be nil initially")
	}

	if client.config != nil {
		t.Error("config should be nil initially")
	}

	if client.agentConn != nil {
		t.Error("agentConn should be nil initially")
	}
}

func TestClient_HasAgentForwarding(t *testing.T) {
	client := NewClient("test.example.com", 22, "testuser", "")

	// Initially should be false (no connection)
	if client.HasAgentForwarding() {
		t.Error("HasAgentForwarding should be false without connection")
	}
}

func TestClient_GetAgentClient(t *testing.T) {
	client := NewClient("test.example.com", 22, "testuser", "")

	// Should return nil when no agent connection
	agentClient := client.GetAgentClient()
	if agentClient != nil {
		t.Error("GetAgentClient should return nil without agent connection")
	}
}

func TestClient_Close(t *testing.T) {
	client := NewClient("test.example.com", 22, "testuser", "")

	// Close should not panic even without connection
	err := client.Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}

func TestClient_getAuthMethods_NoKeyOrAgent(t *testing.T) {
	client := NewClient("test.example.com", 22, "testuser", "/nonexistent/key")

	// Temporarily unset SSH_AUTH_SOCK to ensure no agent
	oldSock := os.Getenv("SSH_AUTH_SOCK")
	os.Unsetenv("SSH_AUTH_SOCK")
	defer os.Setenv("SSH_AUTH_SOCK", oldSock)

	_, err := client.getAuthMethods()
	if err == nil {
		t.Error("getAuthMethods should return error when no key or agent available")
	}

	expectedError := "no authentication methods available"
	if err.Error() != expectedError {
		t.Errorf("Error = %q, want %q", err.Error(), expectedError)
	}
}

func TestClient_getKeyAuth_NonExistentKey(t *testing.T) {
	client := NewClient("test.example.com", 22, "testuser", "")

	_, err := client.getKeyAuth("/nonexistent/key")
	if err == nil {
		t.Error("getKeyAuth should return error for non-existent key")
	}

	expectedError := "failed to read key file"
	if !containsString(err.Error(), expectedError) {
		t.Errorf("Error should contain %q, got %q", expectedError, err.Error())
	}
}

func TestClient_ConnectionFailure(t *testing.T) {
	client := NewClient("nonexistent.invalid", 12345, "testuser", "/nonexistent/key")

	// Connect should fail gracefully
	err := client.Connect()
	if err == nil {
		t.Error("Connect should fail with invalid host/port/key")
	}

	// Should not panic on subsequent Close
	err = client.Close()
	if err != nil {
		t.Errorf("Close after failed connect should not error: %v", err)
	}
}

func TestClient_RunWithoutConnection(t *testing.T) {
	client := NewClient("test.example.com", 22, "testuser", "/path/to/key")

	// Run should fail without connection
	_, err := client.Run("ls")
	if err == nil {
		t.Error("Run should fail without connection")
	}

	// Check error message
	expectedError := "not connected"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestClient_RunInteractiveWithoutConnection(t *testing.T) {
	client := NewClient("test.example.com", 22, "testuser", "/path/to/key")

	// RunInteractive should fail without connection
	err := client.RunInteractive("ls")
	if err == nil {
		t.Error("RunInteractive should fail without connection")
	}

	expectedError := "not connected"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestClient_RunStreamWithoutConnection(t *testing.T) {
	client := NewClient("test.example.com", 22, "testuser", "/path/to/key")

	// RunStream should fail without connection
	err := client.RunStream("ls", nil, nil)
	if err == nil {
		t.Error("RunStream should fail without connection")
	}

	expectedError := "not connected"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestClient_RunInteractiveViaHostWithoutConnection(t *testing.T) {
	client := NewClient("test.example.com", 22, "testuser", "/path/to/key")

	// RunInteractiveViaHost should fail without connection
	err := client.RunInteractiveViaHost("target_example_com", "targetuser", "ls")
	if err == nil {
		t.Error("RunInteractiveViaHost should fail without connection")
	}

	expectedError := "not connected"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestClient_CopyFileWithoutConnection(t *testing.T) {
	client := NewClient("test_example_com", 22, "testuser", "/path/to/key")

	// CopyFile should fail without connection
	err := client.CopyFile("/tmp/test.txt", "/tmp/remote.txt")
	if err == nil {
		t.Error("CopyFile should fail without connection")
	}

	expectedError := "not connected"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestClient_WriteFileWithoutConnection(t *testing.T) {
	client := NewClient("test.example.com", 22, "testuser", "/path/to/key")

	// WriteFile should fail without connection
	err := client.WriteFile("/tmp/remote.txt", []byte("test"))
	if err == nil {
		t.Error("WriteFile should fail without connection")
	}

	expectedError := "not connected"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestClient_getAgentAuth_NoSocket(t *testing.T) {
	client := NewClient("test.example.com", 22, "testuser", "")

	// Temporarily unset SSH_AUTH_SOCK
	oldSock := os.Getenv("SSH_AUTH_SOCK")
	os.Unsetenv("SSH_AUTH_SOCK")
	defer os.Setenv("SSH_AUTH_SOCK", oldSock)

	// Should return nil when no socket
	auth := client.getAgentAuth()
	if auth != nil {
		t.Error("getAgentAuth should return nil when no SSH_AUTH_SOCK")
	}
}

func TestClient_GetAgentClient_WithAgentConnection(t *testing.T) {
	client := NewClient("test.example.com", 22, "testuser", "")

	// Simulate having an agent connection (in a real test, this would be set during getAgentAuth)
	// For unit tests, we just test the method exists and handles nil case
	agentClient := client.GetAgentClient()
	if agentClient != nil {
		t.Error("GetAgentClient should return nil when agentConn is nil")
	}
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsString(s[1:len(s)-1], substr))))
}
