package history

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/marcelsud/swarmctl/internal/executor"
)

const (
	// HistoryImage is the Docker image for the history sidecar
	HistoryImage = "docker.io/marcelsud/swarmctl-history:latest"
	// DefaultRetention is the default number of versions to keep
	DefaultRetention = 10
)

// DeployRecord represents a deploy in the history
type DeployRecord struct {
	ID             int               `json:"id"`
	StackName      string            `json:"stack_name"`
	DeployedAt     time.Time         `json:"deployed_at"`
	ComposeContent string            `json:"compose_content"`
	Images         map[string]string `json:"images"`
	CommitHash     string            `json:"commit_hash,omitempty"`
	Notes          string            `json:"notes,omitempty"`
}

// Manager manages deploy history via the history sidecar container
type Manager struct {
	exec          executor.Executor
	stackName     string
	containerName string
}

// NewManager creates a new history Manager
func NewManager(exec executor.Executor, stackName string) *Manager {
	return &Manager{
		exec:          exec,
		stackName:     stackName,
		containerName: fmt.Sprintf("%s-history", stackName),
	}
}

// IsRunning checks if the history container is running
func (m *Manager) IsRunning() (bool, error) {
	cmd := fmt.Sprintf("docker ps --filter name=^%s$ --format '{{.Names}}'", m.containerName)
	result, err := m.exec.Run(cmd)
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(result.Stdout) == m.containerName, nil
}

// EnsureRunning ensures the history container is running
func (m *Manager) EnsureRunning() error {
	running, err := m.IsRunning()
	if err != nil {
		return fmt.Errorf("failed to check history container: %w", err)
	}

	if running {
		return nil
	}

	// Start the history container
	cmd := fmt.Sprintf(
		"docker run -d --name %s --restart unless-stopped -v %s_history_data:/data %s",
		m.containerName,
		m.stackName,
		HistoryImage,
	)

	result, err := m.exec.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to start history container: %w", err)
	}

	if result.ExitCode != 0 {
		// Container might already exist but stopped, try to start it
		startCmd := fmt.Sprintf("docker start %s", m.containerName)
		result, err = m.exec.Run(startCmd)
		if err != nil {
			return fmt.Errorf("failed to start history container: %w", err)
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("failed to start history container: %s", result.Stderr)
		}
	}

	// Wait a moment for container to be ready
	time.Sleep(500 * time.Millisecond)

	return nil
}

// Record records a new deploy in the history
func (m *Manager) Record(composeContent []byte, images map[string]string) error {
	if err := m.EnsureRunning(); err != nil {
		return err
	}

	imagesJSON, err := json.Marshal(images)
	if err != nil {
		return fmt.Errorf("failed to marshal images: %w", err)
	}

	// Write compose content to temp file on host, then copy to container
	composePath := "/tmp/compose-record.yaml"
	tempPath := fmt.Sprintf("/tmp/%s-compose-record.yaml", m.stackName)
	if err := m.exec.WriteFile(tempPath, composeContent); err != nil {
		return fmt.Errorf("failed to write temp compose file: %w", err)
	}

	// Copy to container
	copyCmd := fmt.Sprintf("docker cp %s %s:%s", tempPath, m.containerName, composePath)
	result, err := m.exec.Run(copyCmd)
	if err != nil {
		return fmt.Errorf("failed to copy compose to history container: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("failed to copy compose to history container: %s", result.Stderr)
	}

	// Record the deploy
	cmd := fmt.Sprintf(
		"docker exec %s /app/history record --stack %s --compose-file %s --images '%s'",
		m.containerName,
		m.stackName,
		composePath,
		string(imagesJSON),
	)

	result, err = m.exec.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to record deploy: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("failed to record deploy: %s", result.Stderr)
	}

	// Clean up temp files
	m.exec.Run(fmt.Sprintf("rm -f %s", tempPath))
	m.exec.Run(fmt.Sprintf("docker exec %s rm -f %s", m.containerName, composePath))

	return nil
}

// List returns the deploy history
func (m *Manager) List(limit int) ([]DeployRecord, error) {
	if err := m.EnsureRunning(); err != nil {
		return nil, err
	}

	cmd := fmt.Sprintf(
		"docker exec %s /app/history list --stack %s --limit %d --format json",
		m.containerName,
		m.stackName,
		limit,
	)

	result, err := m.exec.Run(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list history: %w", err)
	}

	if result.ExitCode != 0 {
		return nil, fmt.Errorf("failed to list history: %s", result.Stderr)
	}

	var records []DeployRecord
	if err := json.Unmarshal([]byte(result.Stdout), &records); err != nil {
		return nil, fmt.Errorf("failed to parse history: %w", err)
	}

	return records, nil
}

// GetPrevious returns the previous deploy (one before the current)
func (m *Manager) GetPrevious() (*DeployRecord, error) {
	records, err := m.List(2)
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("no previous deploy found")
	}

	return &records[1], nil
}

// Get returns a specific deploy by offset (0 = current, -1 = previous, etc.)
func (m *Manager) Get(offset int) (*DeployRecord, error) {
	if err := m.EnsureRunning(); err != nil {
		return nil, err
	}

	cmd := fmt.Sprintf(
		"docker exec %s /app/history get --stack %s --offset %d --format json",
		m.containerName,
		m.stackName,
		offset,
	)

	result, err := m.exec.Run(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get deploy: %w", err)
	}

	if result.ExitCode != 0 {
		return nil, fmt.Errorf("failed to get deploy: %s", result.Stderr)
	}

	var record DeployRecord
	if err := json.Unmarshal([]byte(result.Stdout), &record); err != nil {
		return nil, fmt.Errorf("failed to parse deploy: %w", err)
	}

	return &record, nil
}

// GetComposeContent returns the compose content for a specific deploy
func (m *Manager) GetComposeContent(offset int) ([]byte, error) {
	record, err := m.Get(offset)
	if err != nil {
		return nil, err
	}

	return []byte(record.ComposeContent), nil
}

// Stop stops the history container
func (m *Manager) Stop() error {
	cmd := fmt.Sprintf("docker stop %s", m.containerName)
	result, err := m.exec.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to stop history container: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("failed to stop history container: %s", result.Stderr)
	}

	return nil
}

// Remove removes the history container and its data
func (m *Manager) Remove() error {
	// Stop and remove container
	m.exec.Run(fmt.Sprintf("docker stop %s", m.containerName))
	m.exec.Run(fmt.Sprintf("docker rm %s", m.containerName))

	// Remove volume
	m.exec.Run(fmt.Sprintf("docker volume rm %s_history_data", m.stackName))

	return nil
}
