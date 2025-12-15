package executor

import (
	"io"

	"github.com/marcelsud/swarmctl/internal/config"
	"github.com/marcelsud/swarmctl/internal/ssh"
)

// SSHExecutor executes commands on a remote machine via SSH
type SSHExecutor struct {
	client *ssh.Client
}

// NewSSH creates a new SSHExecutor and connects to the remote host
func NewSSH(cfg config.SSHConfig) (*SSHExecutor, error) {
	client := ssh.NewClient(cfg.Host, cfg.Port, cfg.User, cfg.Key)

	if err := client.Connect(); err != nil {
		return nil, err
	}

	return &SSHExecutor{client: client}, nil
}

// Run executes a command on the remote host and returns the result
func (e *SSHExecutor) Run(cmd string) (*CommandResult, error) {
	result, err := e.client.Run(cmd)
	if err != nil {
		return nil, err
	}

	return &CommandResult{
		Stdout:   result.Stdout,
		Stderr:   result.Stderr,
		ExitCode: result.ExitCode,
	}, nil
}

// RunInteractive runs a command with stdin/stdout/stderr attached
func (e *SSHExecutor) RunInteractive(cmd string) error {
	return e.client.RunInteractive(cmd)
}

// RunInteractiveOnHost runs a command on a remote host through SSH hop
func (e *SSHExecutor) RunInteractiveOnHost(host, user, cmd string) error {
	return e.client.RunInteractiveViaHost(host, user, cmd)
}

// HasAgentForwarding returns true if SSH agent forwarding is available
func (e *SSHExecutor) HasAgentForwarding() bool {
	return e.client.HasAgentForwarding()
}

// RunStream runs a command and streams output to the provided writers
func (e *SSHExecutor) RunStream(cmd string, stdout, stderr io.Writer) error {
	return e.client.RunStream(cmd, stdout, stderr)
}

// WriteFile writes content to a file on the remote host
func (e *SSHExecutor) WriteFile(path string, content []byte) error {
	return e.client.WriteFile(path, content)
}

// Close closes the SSH connection
func (e *SSHExecutor) Close() error {
	return e.client.Close()
}

// IsLocal returns false for SSHExecutor
func (e *SSHExecutor) IsLocal() bool {
	return false
}

// Host returns the remote host address
func (e *SSHExecutor) Host() string {
	return e.client.Host
}

// User returns the SSH user
func (e *SSHExecutor) User() string {
	return e.client.User
}

// Port returns the SSH port
func (e *SSHExecutor) Port() int {
	return e.client.Port
}
