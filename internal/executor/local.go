package executor

import (
	"bytes"
	"io"
	"os"
	"os/exec"
)

// LocalExecutor executes commands on the local machine
type LocalExecutor struct{}

// NewLocal creates a new LocalExecutor
func NewLocal() *LocalExecutor {
	return &LocalExecutor{}
}

// Run executes a command locally and returns the result
func (e *LocalExecutor) Run(cmd string) (*CommandResult, error) {
	c := exec.Command("sh", "-c", cmd)

	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()

	result := &CommandResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: 0,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			// Not a fatal error, just non-zero exit code
			err = nil
		} else {
			return nil, err
		}
	}

	return result, nil
}

// RunInteractive runs a command with stdin/stdout/stderr attached
func (e *LocalExecutor) RunInteractive(cmd string) error {
	c := exec.Command("sh", "-c", cmd)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c.Run()
}

// RunStream runs a command and streams output to the provided writers
func (e *LocalExecutor) RunStream(cmd string, stdout, stderr io.Writer) error {
	c := exec.Command("sh", "-c", cmd)
	c.Stdout = stdout
	c.Stderr = stderr

	return c.Run()
}

// WriteFile writes content to a local file
func (e *LocalExecutor) WriteFile(path string, content []byte) error {
	return os.WriteFile(path, content, 0644)
}

// Close is a no-op for local execution
func (e *LocalExecutor) Close() error {
	return nil
}

// IsLocal returns true for LocalExecutor
func (e *LocalExecutor) IsLocal() bool {
	return true
}
