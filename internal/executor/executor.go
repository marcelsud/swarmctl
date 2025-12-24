package executor

import (
	"io"

	"github.com/marcelsud/swarmctl/internal/config"
)

// CommandResult holds the result of a command execution
type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Executor defines the interface for executing commands
type Executor interface {
	// Run executes a command and returns the result
	Run(cmd string) (*CommandResult, error)

	// RunInteractive runs a command with stdin/stdout/stderr attached
	RunInteractive(cmd string) error

	// RunStream runs a command and streams output to the provided writers
	RunStream(cmd string, stdout, stderr io.Writer) error

	// WriteFile writes content to a file
	WriteFile(path string, content []byte) error

	// Close cleans up any resources
	Close() error

	// IsLocal returns true if executing locally
	IsLocal() bool

	// SetVerbose sets verbose mode for command output
	SetVerbose(verbose bool)
}

// New creates an Executor based on the configuration.
// If SSH host is not configured, returns a LocalExecutor.
// Otherwise, returns an SSHExecutor.
func New(cfg *config.Config) (Executor, error) {
	if cfg.SSH.Host == "" {
		return NewLocal(), nil
	}
	return NewSSH(cfg.SSH)
}
