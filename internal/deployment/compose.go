package deployment

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/marcelsud/swarmctl/internal/executor"
	"github.com/marcelsud/swarmctl/internal/history"
)

// ComposeManager implements Manager for docker compose deployments
type ComposeManager struct {
	exec        executor.Executor
	projectName string
	history     *history.Manager
}

// NewComposeManager creates a new ComposeManager
func NewComposeManager(exec executor.Executor, projectName string) *ComposeManager {
	return &ComposeManager{
		exec:        exec,
		projectName: projectName,
		history:     history.NewManager(exec, projectName),
	}
}

// Deploy deploys using docker compose
func (m *ComposeManager) Deploy(composeContent []byte) error {
	// Ensure history container is running for rollback support
	if err := m.history.EnsureRunning(); err != nil {
		// Log warning but don't fail deploy
		fmt.Printf("Warning: failed to start history container: %v\n", err)
	}

	// Write compose file
	composePath := fmt.Sprintf("/tmp/%s-compose.yaml", m.projectName)
	if err := m.exec.WriteFile(composePath, composeContent); err != nil {
		return fmt.Errorf("failed to write compose file: %w", err)
	}

	// Deploy with docker compose
	cmd := fmt.Sprintf("docker compose -p %s -f %s up -d --remove-orphans", m.projectName, composePath)
	result, err := m.exec.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to deploy: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("deploy failed: %s", result.Stderr)
	}

	// Extract images from compose content for history
	images := m.extractImages(composeContent)

	// Record deploy in history
	if err := m.history.Record(composeContent, images); err != nil {
		// Log warning but don't fail deploy
		fmt.Printf("Warning: failed to record deploy in history: %v\n", err)
	}

	// Clean up temp file
	m.exec.Run(fmt.Sprintf("rm -f %s", composePath))

	return nil
}

// Remove removes the project
func (m *ComposeManager) Remove() error {
	cmd := fmt.Sprintf("docker compose -p %s down", m.projectName)
	result, err := m.exec.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to remove project: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("removal failed: %s", result.Stderr)
	}

	return nil
}

// Exists checks if the project has running containers
func (m *ComposeManager) Exists() (bool, error) {
	cmd := fmt.Sprintf("docker compose -p %s ps -q", m.projectName)
	result, err := m.exec.Run(cmd)
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(result.Stdout) != "", nil
}

// ListServices lists all services in the project
func (m *ComposeManager) ListServices() ([]ServiceStatus, error) {
	cmd := fmt.Sprintf("docker compose -p %s ps --format json", m.projectName)
	result, err := m.exec.Run(cmd)
	if err != nil {
		return nil, err
	}

	if result.Stdout == "" {
		return []ServiceStatus{}, nil
	}

	// Parse JSON output
	var services []ServiceStatus
	lines := strings.Split(strings.TrimSpace(result.Stdout), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var container struct {
			Name    string `json:"Name"`
			Service string `json:"Service"`
			State   string `json:"State"`
			Status  string `json:"Status"`
			Image   string `json:"Image"`
			Ports   string `json:"Ports"`
		}

		if err := json.Unmarshal([]byte(line), &container); err != nil {
			continue
		}

		// Find if service already in list
		found := false
		for i := range services {
			if services[i].Name == fmt.Sprintf("%s_%s", m.projectName, container.Service) {
				// Increment replica count
				found = true
				break
			}
		}

		if !found {
			services = append(services, ServiceStatus{
				Name:     fmt.Sprintf("%s_%s", m.projectName, container.Service),
				Mode:     "replicated",
				Replicas: "1/1", // Compose doesn't show replicas like swarm
				Image:    container.Image,
				Ports:    container.Ports,
			})
		}
	}

	return services, nil
}

// GetServiceLogs gets logs from a service
func (m *ComposeManager) GetServiceLogs(serviceName string, follow bool, since string, tail int) (string, error) {
	cmd := fmt.Sprintf("docker compose -p %s logs %s", m.projectName, serviceName)
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
func (m *ComposeManager) StreamServiceLogs(serviceName string, follow bool, tail int, stdout, stderr io.Writer) error {
	cmd := fmt.Sprintf("docker compose -p %s logs %s", m.projectName, serviceName)
	if tail > 0 {
		cmd += fmt.Sprintf(" --tail %d", tail)
	}
	if follow {
		cmd += " --follow"
	}

	return m.exec.RunStream(cmd, stdout, stderr)
}

// FindRunningContainer finds a running container ID for a service
func (m *ComposeManager) FindRunningContainer(serviceName string) (string, error) {
	cmd := fmt.Sprintf("docker compose -p %s ps -q %s", m.projectName, serviceName)
	result, err := m.exec.Run(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to find container: %w", err)
	}

	containerID := strings.TrimSpace(result.Stdout)
	if containerID == "" {
		return "", fmt.Errorf("no running container found for service %s", serviceName)
	}

	// Return first 12 chars
	lines := strings.Split(containerID, "\n")
	if len(lines[0]) > 12 {
		return lines[0][:12], nil
	}
	return lines[0], nil
}

// GetContainerStatus gets container status for all services
func (m *ComposeManager) GetContainerStatus() ([]ContainerStatus, error) {
	cmd := fmt.Sprintf("docker compose -p %s ps --format json", m.projectName)
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
		if line == "" {
			continue
		}

		var container struct {
			ID      string `json:"ID"`
			Name    string `json:"Name"`
			Service string `json:"Service"`
			State   string `json:"State"`
			Status  string `json:"Status"`
		}

		if err := json.Unmarshal([]byte(line), &container); err != nil {
			continue
		}

		containers = append(containers, ContainerStatus{
			ID:      container.ID,
			Name:    container.Name,
			Service: container.Service,
			State:   container.State,
			Error:   "",
		})
	}

	return containers, nil
}

// SupportsRollback returns true - compose supports rollback via history
func (m *ComposeManager) SupportsRollback() bool {
	return true
}

// SupportsScale returns false - compose doesn't support dynamic scaling
func (m *ComposeManager) SupportsScale() bool {
	return false
}

// RollbackService rolls back a service to its previous version
func (m *ComposeManager) RollbackService(serviceName string) error {
	// Get previous compose content
	composeContent, err := m.history.GetComposeContent(-1)
	if err != nil {
		return fmt.Errorf("failed to get previous deploy: %w", err)
	}

	// Redeploy with previous content
	return m.Deploy(composeContent)
}

// RollbackAll rolls back all services to the previous version
func (m *ComposeManager) RollbackAll() error {
	// Get previous compose content
	composeContent, err := m.history.GetComposeContent(-1)
	if err != nil {
		return fmt.Errorf("failed to get previous deploy: %w", err)
	}

	// Redeploy with previous content
	return m.Deploy(composeContent)
}

// ScaleService is not supported in compose mode
func (m *ComposeManager) ScaleService(serviceName string, replicas int) error {
	return NewUnsupportedError("scale", "compose")
}

// GetStackName returns the project name
func (m *ComposeManager) GetStackName() string {
	return m.projectName
}

// GetMode returns the deployment mode
func (m *ComposeManager) GetMode() string {
	return "compose"
}

// extractImages extracts image names from compose content
func (m *ComposeManager) extractImages(composeContent []byte) map[string]string {
	images := make(map[string]string)

	// Simple extraction - look for image: lines
	lines := strings.Split(string(composeContent), "\n")
	var currentService string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if it's a service definition (ends with :)
		if strings.HasSuffix(trimmed, ":") && !strings.HasPrefix(trimmed, "#") {
			if !strings.Contains(trimmed, "image:") {
				currentService = strings.TrimSuffix(trimmed, ":")
			}
		}

		// Check for image line
		if strings.HasPrefix(trimmed, "image:") {
			image := strings.TrimSpace(strings.TrimPrefix(trimmed, "image:"))
			if currentService != "" {
				images[currentService] = image
			}
		}
	}

	return images
}

// GetHistory returns the history manager
func (m *ComposeManager) GetHistory() *history.Manager {
	return m.history
}
