package swarm

import (
	"fmt"
	"strings"
)

// ServiceStatus represents the status of a service
type ServiceStatus struct {
	Name     string
	Mode     string
	Replicas string
	Image    string
	Ports    string
}

// ListServices lists all services in the stack
func (m *Manager) ListServices() ([]ServiceStatus, error) {
	cmd := fmt.Sprintf("docker stack services %s --format '{{.Name}}|{{.Mode}}|{{.Replicas}}|{{.Image}}|{{.Ports}}'", m.stackName)
	result, err := m.client.Run(cmd)
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

// GetServiceStatus gets detailed status for a service
func (m *Manager) GetServiceStatus(serviceName string) (*ServiceStatus, error) {
	fullName := fmt.Sprintf("%s_%s", m.stackName, serviceName)
	cmd := fmt.Sprintf("docker service inspect %s --format '{{.Spec.Name}}|{{.Spec.Mode.Replicated.Replicas}}|{{.Spec.TaskTemplate.ContainerSpec.Image}}'", fullName)

	result, err := m.client.Run(cmd)
	if err != nil {
		return nil, err
	}

	if result.ExitCode != 0 {
		return nil, fmt.Errorf("service not found: %s", serviceName)
	}

	parts := strings.Split(strings.TrimSpace(result.Stdout), "|")
	if len(parts) >= 3 {
		return &ServiceStatus{
			Name:     parts[0],
			Replicas: parts[1],
			Image:    parts[2],
		}, nil
	}

	return nil, fmt.Errorf("failed to parse service status")
}

// RollbackService rolls back a service to its previous version
func (m *Manager) RollbackService(serviceName string) error {
	fullName := fmt.Sprintf("%s_%s", m.stackName, serviceName)
	cmd := fmt.Sprintf("docker service update --rollback %s", fullName)

	result, err := m.client.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to rollback service: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("rollback failed: %s", result.Stderr)
	}

	return nil
}

// ScaleService scales a service to the specified replicas
func (m *Manager) ScaleService(serviceName string, replicas int) error {
	fullName := fmt.Sprintf("%s_%s", m.stackName, serviceName)
	cmd := fmt.Sprintf("docker service scale %s=%d", fullName, replicas)

	result, err := m.client.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to scale service: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("scale failed: %s", result.Stderr)
	}

	return nil
}

// GetServiceLogs gets logs from a service
func (m *Manager) GetServiceLogs(serviceName string, follow bool, since string, tail int) (string, error) {
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

	result, err := m.client.Run(cmd)
	if err != nil {
		return "", err
	}

	return result.Stdout + result.Stderr, nil
}

// FindRunningContainer finds a running container ID for a service
func (m *Manager) FindRunningContainer(serviceName string) (string, error) {
	fullName := fmt.Sprintf("%s_%s", m.stackName, serviceName)

	// Get running tasks for the service
	cmd := fmt.Sprintf("docker service ps %s --filter 'desired-state=running' --format '{{.ID}}' | head -1", fullName)
	result, err := m.client.Run(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to find tasks: %w", err)
	}

	taskID := strings.TrimSpace(result.Stdout)
	if taskID == "" {
		return "", fmt.Errorf("no running tasks found for service %s", serviceName)
	}

	// Get container ID from task
	cmd = fmt.Sprintf("docker inspect --format '{{.Status.ContainerStatus.ContainerID}}' %s", taskID)
	result, err = m.client.Run(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to get container ID: %w", err)
	}

	containerID := strings.TrimSpace(result.Stdout)
	if containerID == "" {
		return "", fmt.Errorf("container not found for task %s", taskID)
	}

	return containerID[:12], nil
}

func getOrEmpty(slice []string, index int) string {
	if index < len(slice) {
		return slice[index]
	}
	return ""
}
