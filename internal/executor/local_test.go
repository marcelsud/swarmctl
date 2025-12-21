package executor

import (
	"bytes"
	"strings"
	"testing"
)

func TestLocalExecutor_Run(t *testing.T) {
	e := NewLocal()

	tests := []struct {
		name       string
		cmd        string
		wantErr    bool
		wantStdout string
	}{
		{
			name:       "echo command",
			cmd:        "echo 'hello world'",
			wantErr:    false,
			wantStdout: "hello world\n",
		},
		{
			name:    "ls command",
			cmd:     "ls /tmp",
			wantErr: false,
		},
		{
			name:    "exit with error",
			cmd:     "exit 1",
			wantErr: false, // não retorna erro, apenas exit code != 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := e.Run(tt.cmd)

			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result.ExitCode == 0 && tt.wantStdout != "" {
				if result.Stdout != tt.wantStdout {
					t.Errorf("Run() stdout = %q, want %q", result.Stdout, tt.wantStdout)
				}
			}

			// Test that result is never nil
			if result == nil {
				t.Error("Run() result should not be nil")
			}
		})
	}
}

func TestLocalExecutor_Run_ExitCode(t *testing.T) {
	e := NewLocal()

	// Test non-zero exit code
	result, err := e.Run("exit 42")
	if err != nil {
		t.Errorf("Run() unexpected error: %v", err)
	}

	if result.ExitCode != 42 {
		t.Errorf("Run() exitCode = %d, want 42", result.ExitCode)
	}
}

func TestLocalExecutor_Run_InvalidCommand(t *testing.T) {
	e := NewLocal()

	// Test invalid command that doesn't exist - sh returns exit code 127
	result, err := e.Run("nonexistentcommand12345")
	if err != nil {
		t.Errorf("Run() unexpected error: %v", err)
	}

	if result == nil {
		t.Error("Run() should not return nil result even for invalid command")
	}

	if result.ExitCode != 127 {
		t.Errorf("Run() exitCode = %d, want 127 for command not found", result.ExitCode)
	}
}

func TestLocalExecutor_RunInteractive(t *testing.T) {
	e := NewLocal()

	// This is harder to test without stdin/stdout redirection
	// Just test that it doesn't panic
	err := e.RunInteractive("echo 'test'")
	if err != nil {
		t.Errorf("RunInteractive() error = %v", err)
	}
}

func TestLocalExecutor_RunStream(t *testing.T) {
	e := NewLocal()

	var stdout, stderr bytes.Buffer

	tests := []struct {
		name    string
		cmd     string
		wantOut string
		wantErr bool
	}{
		{
			name:    "stdout only",
			cmd:     "echo 'hello'",
			wantOut: "hello\n",
			wantErr: false,
		},
		{
			name:    "stderr output",
			cmd:     "echo 'error' >&2",
			wantOut: "error\n",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout.Reset()
			stderr.Reset()

			err := e.RunStream(tt.cmd, &stdout, &stderr)

			if (err != nil) != tt.wantErr {
				t.Errorf("RunStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check that something was written to either stdout or stderr
			output := stdout.String() + stderr.String()
			if !strings.Contains(output, strings.TrimSuffix(tt.wantOut, "\n")) {
				t.Errorf("RunStream() output = %q, want to contain %q", output, tt.wantOut)
			}
		})
	}
}

func TestLocalExecutor_WriteFile(t *testing.T) {
	e := NewLocal()

	content := []byte("test content")
	path := "/tmp/testfile.txt"

	// Write file
	err := e.WriteFile(path, content)
	if err != nil {
		t.Errorf("WriteFile() error = %v", err)
	}

	// Verify file exists and has correct content (using os directly since we're testing)
	// Note: In a real scenario, you'd use the executor itself to avoid os package
	result, err := e.Run("cat " + path)
	if err != nil {
		t.Errorf("Failed to read test file: %v", err)
	}

	if result.Stdout != "test content" {
		t.Errorf("WriteFile() content = %q, want %q", result.Stdout, "test content")
	}

	// Cleanup
	e.Run("rm -f " + path)
}

func TestLocalExecutor_Close(t *testing.T) {
	e := NewLocal()

	// Close should be a no-op for LocalExecutor
	err := e.Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}

func TestLocalExecutor_IsLocal(t *testing.T) {
	e := NewLocal()

	if !e.IsLocal() {
		t.Error("IsLocal() should return true for LocalExecutor")
	}
}

func TestNewLocal(t *testing.T) {
	e := NewLocal()

	if e == nil {
		t.Error("NewLocal() should not return nil")
	}

	// Verify it implements the interface
	var _ Executor = e
}
