package ssh

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Client represents an SSH client connection
type Client struct {
	Host    string
	Port    int
	User    string
	KeyPath string

	conn   *ssh.Client
	config *ssh.ClientConfig
}

// NewClient creates a new SSH client
func NewClient(host string, port int, user string, keyPath string) *Client {
	return &Client{
		Host:    host,
		Port:    port,
		User:    user,
		KeyPath: keyPath,
	}
}

// Connect establishes an SSH connection
func (c *Client) Connect() error {
	authMethods, err := c.getAuthMethods()
	if err != nil {
		return fmt.Errorf("failed to get auth methods: %w", err)
	}

	c.config = &ssh.ClientConfig{
		User:            c.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: proper host key verification
		Timeout:         30 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	conn, err := ssh.Dial("tcp", addr, c.config)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	c.conn = conn
	return nil
}

// Close closes the SSH connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// getAuthMethods returns available authentication methods
func (c *Client) getAuthMethods() ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod

	// Try SSH agent first
	if agentAuth := c.getAgentAuth(); agentAuth != nil {
		methods = append(methods, agentAuth)
	}

	// Try private key file
	if c.KeyPath != "" {
		keyAuth, err := c.getKeyAuth(c.KeyPath)
		if err == nil {
			methods = append(methods, keyAuth)
		}
	} else {
		// Try default key locations
		home, _ := os.UserHomeDir()
		defaultKeys := []string{
			filepath.Join(home, ".ssh", "id_ed25519"),
			filepath.Join(home, ".ssh", "id_rsa"),
		}
		for _, keyPath := range defaultKeys {
			if _, err := os.Stat(keyPath); err == nil {
				keyAuth, err := c.getKeyAuth(keyPath)
				if err == nil {
					methods = append(methods, keyAuth)
					break
				}
			}
		}
	}

	if len(methods) == 0 {
		return nil, fmt.Errorf("no authentication methods available")
	}

	return methods, nil
}

// getAgentAuth returns SSH agent authentication
func (c *Client) getAgentAuth() ssh.AuthMethod {
	socket := os.Getenv("SSH_AUTH_SOCK")
	if socket == "" {
		return nil
	}

	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil
	}

	agentClient := agent.NewClient(conn)
	return ssh.PublicKeysCallback(agentClient.Signers)
}

// getKeyAuth returns key-based authentication
func (c *Client) getKeyAuth(keyPath string) (ssh.AuthMethod, error) {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to parse key file: %w", err)
	}

	return ssh.PublicKeys(signer), nil
}
