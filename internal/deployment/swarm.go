package deployment

import (
	"fmt"
	"io"
	"strings"

	"github.com/marcelsud/swarmctl/internal/executor"
)

// SwarmManager implements Manager for Docker Swarm deployments
type SwarmManager struct {
	exec      executor.Executor
	stackName string
}

// NewSwarmManager creates a new SwarmManager
func NewSwarmManager(exec executor.Executor, stackName string) *SwarmManager {
	return &SwarmManager{
		exec:      exec,
		stackName: stackName,
	}
}

// Deploy deploys a stack using docker stack deploy
func (m *SwarmManager) Deploy(composeContent []byte) error {
	// Write compose file
	composePath := fmt.Sprintf("/tmp/%s-compose.yaml", m.stackName)
	if err := m.exec.WriteFile(composePath, composeContent); err != nil {
		return fmt.Errorf("failed to write compose file: %w", err)
	}

	// Deploy stack
	cmd := fmt.Sprintf("docker stack deploy -c %s %s --with-registry-auth", composePath, m.stackName)
	result, err := m.exec.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to deploy stack: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("stack deploy failed: %s", result.Stderr)
	}

	// Clean up temp file
	m.exec.Run(fmt.Sprintf("rm -f %s", composePath))

	return nil
}

// Remove removes the stack
func (m *SwarmManager) Remove() error {
	cmd := fmt.Sprintf("docker stack rm %s", m.stackName)
	result, err := m.exec.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to remove stack: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("stack removal failed: %s", result.Stderr)
	}

	return nil
}

// Exists checks if the stack exists
func (m *SwarmManager) Exists() (bool, error) {
	result, err := m.exec.Run("docker stack ls --format '{{.Name}}'")
	if err != nil {
		return false, err
	}

	stacks := strings.Split(strings.TrimSpace(result.Stdout), "\n")
	for _, stack := range stacks {
		if stack == m.stackName {
			return true, nil
		}
	}

	return false, nil
}

// ListServices lists all services in the stack
func (m *SwarmManager) ListServices() ([]ServiceStatus, error) {
	cmd := fmt.Sprintf("docker stack services %s --format '{{.Name}}|{{.Mode}}|{{.Replicas}}|{{.Image}}|{{.Ports}}'", m.stackName)
	result, err := m.exec.Run(cmd)
	if err != nil {
		return nil, err
	}

	if result.Stdout == "" {
		return []ServiceStatus{}, nil
	}

	var services []ServiceStatus
	lines := strings.Split(strings.TrimSpace(result.Stdout), "\n")
	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) >= 4 {
			services = append(services, ServiceStatus{
				Name:     parts[0],
				Mode:     parts[1],
				Replicas: parts[2],
				Image:    parts[3],
				Ports:    getOrEmpty(parts, 4),
			})
		}
	}

	return services, nil
}

// GetServiceLogs gets logs from a service
func (m *SwarmManager) GetServiceLogs(serviceName string, follow bool, since string, tail int) (string, error) {
	fullName := fmt.Sprintf("%s_%s", m.stackName, serviceName)

	cmd := fmt.Sprintf("docker service logs %s", fullName)
	if tail > 0 {
		cmd += fmt.Sprintf(" --tail %d", tail)
	}
	if since != "" {
		cmd += fmt.Sprintf(" --since %s", since)
	}
	if follow {
		cmd += " --follow"
	}

	result, err := m.exec.Run(cmd)
	if err != nil {
		return "", err
	}

	return result.Stdout + result.Stderr, nil
}

// StreamServiceLogs streams logs from a service
func (m *SwarmManager) StreamServiceLogs(serviceName string, follow bool, tail int, stdout, stderr io.Writer) error {
	fullName := fmt.Sprintf("%s_%s", m.stackName, serviceName)

	cmd := fmt.Sprintf("docker service logs %s", fullName)
	if tail > 0 {
		cmd += fmt.Sprintf(" --tail %d", tail)
	}
	if follow {
		cmd += " --follow"
	}

	return m.exec.RunStream(cmd, stdout, stderr)
}

// FindRunningContainer finds a running container ID for a service
func (m *SwarmManager) FindRunningContainer(serviceName string) (string, error) {
	fullName := fmt.Sprintf("%s_%s", m.stackName, serviceName)

	// Get running tasks for the service
	cmd := fmt.Sprintf("docker service ps %s --filter 'desired-state=running' --format '{{.ID}}' | head -1", fullName)
	result, err := m.exec.Run(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to find tasks: %w", err)
	}

	taskID := strings.TrimSpace(result.Stdout)
	if taskID == "" {
		return "", fmt.Errorf("no running tasks found for service %s", serviceName)
	}

	// Get container ID from task
	cmd = fmt.Sprintf("docker inspect --format '{{.Status.ContainerStatus.ContainerID}}' %s", taskID)
	result, err = m.exec.Run(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to get container ID: %w", err)
	}

	containerID := strings.TrimSpace(result.Stdout)
	if containerID == "" {
		return "", fmt.Errorf("container not found for task %s", taskID)
	}

	return containerID[:12], nil
}

// GetContainerStatus gets container/task status for all services
func (m *SwarmManager) GetContainerStatus() ([]ContainerStatus, error) {
	cmd := fmt.Sprintf("docker stack ps %s --format '{{.ID}}|{{.Name}}|{{.CurrentState}}|{{.Error}}'", m.stackName)
	result, err := m.exec.Run(cmd)
	if err != nil {
		return nil, err
	}

	if result.Stdout == "" {
		return []ContainerStatus{}, nil
	}

	var containers []ContainerStatus
	lines := strings.Split(strings.TrimSpace(result.Stdout), "\n")
	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) >= 3 {
			// Extract service name from task name (e.g., "myapp_web.1" -> "web")
			taskName := parts[1]
			serviceName := taskName
			if idx := strings.Index(taskName, "_"); idx != -1 {
				serviceName = taskName[idx+1:]
			}
			if idx := strings.LastIndex(serviceName, "."); idx != -1 {
				serviceName = serviceName[:idx]
			}

			containers = append(containers, ContainerStatus{
				ID:      parts[0],
				Name:    parts[1],
				Service: serviceName,
				State:   parts[2],
				Error:   getOrEmpty(parts, 3),
			})
		}
	}

	return containers, nil
}

// SupportsRollback returns true - swarm supports rollback
func (m *SwarmManager) SupportsRollback() bool {
	return true
}

// SupportsScale returns true - swarm supports scale
func (m *SwarmManager) SupportsScale() bool {
	return true
}

// RollbackService rolls back a service to its previous version
func (m *SwarmManager) RollbackService(serviceName string) error {
	fullName := fmt.Sprintf("%s_%s", m.stackName, serviceName)
	cmd := fmt.Sprintf("docker service update --rollback %s", fullName)

	result, err := m.exec.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to rollback service: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("rollback failed: %s", result.Stderr)
	}

	return nil
}

// RollbackAll rolls back all services in the stack
func (m *SwarmManager) RollbackAll() error {
	services, err := m.ListServices()
	if err != nil {
		return fmt.Errorf("failed to list services: %w", err)
	}

	for _, svc := range services {
		// Extract service name without stack prefix
		serviceName := svc.Name
		if strings.HasPrefix(serviceName, m.stackName+"_") {
			serviceName = serviceName[len(m.stackName)+1:]
		}

		if err := m.RollbackService(serviceName); err != nil {
			return fmt.Errorf("failed to rollback %s: %w", serviceName, err)
		}
	}

	return nil
}

// ScaleService scales a service to the specified replicas
func (m *SwarmManager) ScaleService(serviceName string, replicas int) error {
	fullName := fmt.Sprintf("%s_%s", m.stackName, serviceName)
	cmd := fmt.Sprintf("docker service scale %s=%d", fullName, replicas)

	result, err := m.exec.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to scale service: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("scale failed: %s", result.Stderr)
	}

	return nil
}

// GetStackName returns the stack name
func (m *SwarmManager) GetStackName() string {
	return m.stackName
}

// GetMode returns the deployment mode
func (m *SwarmManager) GetMode() string {
	return "swarm"
}

func getOrEmpty(slice []string, index int) string {
	if index < len(slice) {
		return slice[index]
	}
	return ""
}
