package ssh

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/kballard/go-shellquote"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// CommandResult holds the result of a command execution
type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Run executes a command and returns the result
func (c *Client) Run(cmd string) (*CommandResult, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	session, err := c.conn.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(cmd)

	result := &CommandResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: 0,
	}

	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			result.ExitCode = exitErr.ExitStatus()
		} else {
			return nil, fmt.Errorf("failed to run command: %w", err)
		}
	}

	return result, nil
}

// RunInteractive runs a command with stdin/stdout/stderr attached
func (c *Client) RunInteractive(cmd string) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	session, err := c.conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	// Request pseudo-terminal
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		return fmt.Errorf("failed to request pty: %w", err)
	}

	return session.Run(cmd)
}

func validateSSHParam(param string) error {
	validParam := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_]{0,62}$`)
	if !validParam.MatchString(param) {
		return fmt.Errorf("invalid SSH parameter '%s': must contain only alphanumeric characters and underscores", param)
	}
	return nil
}

// RunInteractiveViaHost runs a command on a remote host through SSH hop with agent forwarding
func (c *Client) RunInteractiveViaHost(targetHost, targetUser, cmd string) error {
	if err := validateSSHParam(targetHost); err != nil {
		return err
	}
	if err := validateSSHParam(targetUser); err != nil {
		return err
	}
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	session, err := c.conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Enable SSH agent forwarding if available
	if c.agentConn != nil {
		if err := agent.RequestAgentForwarding(session); err != nil {
			return fmt.Errorf("failed to request agent forwarding: %w", err)
		}
		if err := agent.ForwardToAgent(c.conn, c.GetAgentClient()); err != nil {
			return fmt.Errorf("failed to forward agent: %w", err)
		}
	}

	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	// Request pseudo-terminal
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		return fmt.Errorf("failed to request pty: %w", err)
	}

	target := fmt.Sprintf("%s@%s", targetUser, targetHost)
	sshCmd := fmt.Sprintf("ssh -tt -o StrictHostKeyChecking=yes %s %s", shellquote.Join(target), shellquote.Join(cmd))

	return session.Run(sshCmd)
}

// RunStream runs a command and streams output to the provided writers
func (c *Client) RunStream(cmd string, stdout, stderr io.Writer) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	session, err := c.conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	session.Stdout = stdout
	session.Stderr = stderr

	return session.Run(cmd)
}

// CopyFile copies a local file to the remote host
func (c *Client) CopyFile(localPath, remotePath string) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	content, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to read local file: %w", err)
	}

	return c.WriteFile(remotePath, content)
}

// WriteFile writes content to a file on the remote host
func (c *Client) WriteFile(remotePath string, content []byte) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	session, err := c.conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Use cat to write file content
	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		w.Write(content)
	}()

	cmd := fmt.Sprintf("cat > %s", remotePath)
	return session.Run(cmd)
}
