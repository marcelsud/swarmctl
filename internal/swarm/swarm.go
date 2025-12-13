package swarm

import (
	"fmt"
	"strings"

	"github.com/marcelsud/swarmctl/internal/executor"
)

// Manager handles Swarm operations
type Manager struct {
	exec      executor.Executor
	stackName string
}

// NewManager creates a new Swarm manager
func NewManager(exec executor.Executor, stackName string) *Manager {
	return &Manager{
		exec:      exec,
		stackName: stackName,
	}
}

// IsSwarmInitialized checks if Swarm is initialized on the manager node
func (m *Manager) IsSwarmInitialized() (bool, error) {
	result, err := m.exec.Run("docker info --format '{{.Swarm.LocalNodeState}}'")
	if err != nil {
		return false, err
	}

	state := strings.TrimSpace(result.Stdout)
	return state == "active", nil
}

// InitSwarm initializes Docker Swarm
func (m *Manager) InitSwarm() error {
	result, err := m.exec.Run("docker swarm init")
	if err != nil {
		return fmt.Errorf("failed to init swarm: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("swarm init failed: %s", result.Stderr)
	}

	return nil
}

// IsDockerInstalled checks if Docker is installed
func (m *Manager) IsDockerInstalled() (bool, error) {
	result, err := m.exec.Run("docker --version")
	if err != nil {
		return false, err
	}
	return result.ExitCode == 0, nil
}

// GetDockerVersion returns the Docker version
func (m *Manager) GetDockerVersion() (string, error) {
	result, err := m.exec.Run("docker --version")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(result.Stdout), nil
}

// CreateNetwork creates an overlay network for the stack
func (m *Manager) CreateNetwork(name string) error {
	// Check if network exists
	checkCmd := fmt.Sprintf("docker network ls --filter name=^%s$ --format '{{.Name}}'", name)
	result, err := m.exec.Run(checkCmd)
	if err != nil {
		return err
	}

	if strings.TrimSpace(result.Stdout) == name {
		return nil // Network already exists
	}

	// Create network
	createCmd := fmt.Sprintf("docker network create --driver overlay --attachable %s", name)
	result, err = m.exec.Run(createCmd)
	if err != nil {
		return fmt.Errorf("failed to create network: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("network creation failed: %s", result.Stderr)
	}

	return nil
}

// RegistryLogin logs into a container registry
func (m *Manager) RegistryLogin(url, username, password string) error {
	if username == "" || password == "" {
		return nil // Skip if no credentials
	}

	cmd := fmt.Sprintf("echo '%s' | docker login %s -u %s --password-stdin", password, url, username)
	result, err := m.exec.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to login to registry: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("registry login failed: %s", result.Stderr)
	}

	return nil
}
