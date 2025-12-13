package swarm

import (
	"fmt"
	"strings"
)

// DeployStack deploys a stack using docker stack deploy
func (m *Manager) DeployStack(composeContent []byte) error {
	// Write compose file to remote
	remotePath := fmt.Sprintf("/tmp/%s-compose.yaml", m.stackName)
	if err := m.client.WriteFile(remotePath, composeContent); err != nil {
		return fmt.Errorf("failed to upload compose file: %w", err)
	}

	// Deploy stack
	cmd := fmt.Sprintf("docker stack deploy -c %s %s --with-registry-auth", remotePath, m.stackName)
	result, err := m.client.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to deploy stack: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("stack deploy failed: %s", result.Stderr)
	}

	// Clean up temp file
	m.client.Run(fmt.Sprintf("rm -f %s", remotePath))

	return nil
}

// RemoveStack removes a stack
func (m *Manager) RemoveStack() error {
	cmd := fmt.Sprintf("docker stack rm %s", m.stackName)
	result, err := m.client.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to remove stack: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("stack removal failed: %s", result.Stderr)
	}

	return nil
}

// ListStacks lists all stacks
func (m *Manager) ListStacks() ([]string, error) {
	result, err := m.client.Run("docker stack ls --format '{{.Name}}'")
	if err != nil {
		return nil, err
	}

	if result.Stdout == "" {
		return []string{}, nil
	}

	stacks := strings.Split(strings.TrimSpace(result.Stdout), "\n")
	return stacks, nil
}

// StackExists checks if the stack exists
func (m *Manager) StackExists() (bool, error) {
	stacks, err := m.ListStacks()
	if err != nil {
		return false, err
	}

	for _, stack := range stacks {
		if stack == m.stackName {
			return true, nil
		}
	}

	return false, nil
}
