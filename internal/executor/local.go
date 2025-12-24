package executor

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// LocalExecutor executes commands on the local machine
type LocalExecutor struct {
	verbose bool
}

// NewLocal creates a new LocalExecutor
func NewLocal() *LocalExecutor {
	return &LocalExecutor{verbose: false}
}

// SetVerbose sets verbose mode for command output
func (e *LocalExecutor) SetVerbose(v bool) {
	e.verbose = v
}

// Run executes a command locally and returns the result
func (e *LocalExecutor) Run(cmd string) (*CommandResult, error) {
	if e.verbose {
		fmt.Fprintf(os.Stderr, "→ Running: %s\n", cmd)
	}

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

	if e.verbose {
		if result.Stdout != "" {
			fmt.Fprintf(os.Stderr, "→ Stdout:\n%s\n", result.Stdout)
		}
		if result.Stderr != "" {
			fmt.Fprintf(os.Stderr, "→ Stderr:\n%s\n", result.Stderr)
		}
		fmt.Fprintf(os.Stderr, "→ Exit code: %d\n", result.ExitCode)
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
