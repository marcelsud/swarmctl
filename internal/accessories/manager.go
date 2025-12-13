package accessories

import (
	"fmt"
	"strings"

	"github.com/marcelsud/swarmctl/internal/ssh"
)

// Manager handles accessory services
type Manager struct {
	client    *ssh.Client
	stackName string
}

// NewManager creates a new accessories manager
func NewManager(client *ssh.Client, stackName string) *Manager {
	return &Manager{
		client:    client,
		stackName: stackName,
	}
}

// AccessoryStatus represents the status of an accessory service
type AccessoryStatus struct {
	Name     string
	Replicas string
	Running  bool
}

// Start starts an accessory service by scaling it to 1
func (m *Manager) Start(name string) error {
	fullName := fmt.Sprintf("%s_%s", m.stackName, name)
	cmd := fmt.Sprintf("docker service scale %s=1", fullName)

	result, err := m.client.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to start accessory: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("start failed: %s", result.Stderr)
	}

	return nil
}

// Stop stops an accessory service by scaling it to 0
func (m *Manager) Stop(name string) error {
	fullName := fmt.Sprintf("%s_%s", m.stackName, name)
	cmd := fmt.Sprintf("docker service scale %s=0", fullName)

	result, err := m.client.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to stop accessory: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("stop failed: %s", result.Stderr)
	}

	return nil
}

// Restart restarts an accessory service
func (m *Manager) Restart(name string) error {
	fullName := fmt.Sprintf("%s_%s", m.stackName, name)
	cmd := fmt.Sprintf("docker service update --force %s", fullName)

	result, err := m.client.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to restart accessory: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("restart failed: %s", result.Stderr)
	}

	return nil
}

// GetStatus gets the status of an accessory service
func (m *Manager) GetStatus(name string) (*AccessoryStatus, error) {
	fullName := fmt.Sprintf("%s_%s", m.stackName, name)
	cmd := fmt.Sprintf("docker service ls --filter name=%s --format '{{.Name}}|{{.Replicas}}'", fullName)

	result, err := m.client.Run(cmd)
	if err != nil {
		return nil, err
	}

	if result.Stdout == "" {
		return nil, fmt.Errorf("accessory %s not found", name)
	}

	parts := strings.Split(strings.TrimSpace(result.Stdout), "|")
	if len(parts) < 2 {
		return nil, fmt.Errorf("failed to parse status")
	}

	replicas := parts[1]
	running := !strings.HasPrefix(replicas, "0/")

	return &AccessoryStatus{
		Name:     name,
		Replicas: replicas,
		Running:  running,
	}, nil
}

// ListAll lists all accessories and their status
func (m *Manager) ListAll(accessoryNames []string) ([]AccessoryStatus, error) {
	var statuses []AccessoryStatus

	for _, name := range accessoryNames {
		status, err := m.GetStatus(name)
		if err != nil {
			statuses = append(statuses, AccessoryStatus{
				Name:     name,
				Replicas: "not deployed",
				Running:  false,
			})
			continue
		}
		statuses = append(statuses, *status)
	}

	return statuses, nil
}
