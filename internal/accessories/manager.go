package accessories

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/marcelsud/swarmctl/internal/config"
	"github.com/marcelsud/swarmctl/internal/executor"
)

// Manager handles accessory services
type Manager struct {
	exec      executor.Executor
	stackName string
	mode      config.DeploymentMode
}

// NewManager creates a new accessories manager
func NewManager(exec executor.Executor, stackName string, mode config.DeploymentMode) *Manager {
	return &Manager{
		exec:      exec,
		stackName: stackName,
		mode:      mode,
	}
}

// AccessoryStatus represents the status of an accessory service
type AccessoryStatus struct {
	Name     string
	Replicas string
	Running  bool
}

// Start starts an accessory service
func (m *Manager) Start(name string) error {
	var cmd string

	if m.mode == config.ModeCompose {
		cmd = fmt.Sprintf("docker compose -p %s start %s", m.stackName, name)
	} else {
		fullName := fmt.Sprintf("%s_%s", m.stackName, name)
		cmd = fmt.Sprintf("docker service scale %s=1", fullName)
	}

	result, err := m.exec.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to start accessory: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("start failed: %s", result.Stderr)
	}

	return nil
}

// Stop stops an accessory service
func (m *Manager) Stop(name string) error {
	var cmd string

	if m.mode == config.ModeCompose {
		cmd = fmt.Sprintf("docker compose -p %s stop %s", m.stackName, name)
	} else {
		fullName := fmt.Sprintf("%s_%s", m.stackName, name)
		cmd = fmt.Sprintf("docker service scale %s=0", fullName)
	}

	result, err := m.exec.Run(cmd)
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
	var cmd string

	if m.mode == config.ModeCompose {
		cmd = fmt.Sprintf("docker compose -p %s restart %s", m.stackName, name)
	} else {
		fullName := fmt.Sprintf("%s_%s", m.stackName, name)
		cmd = fmt.Sprintf("docker service update --force %s", fullName)
	}

	result, err := m.exec.Run(cmd)
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
	if m.mode == config.ModeCompose {
		return m.getComposeStatus(name)
	}
	return m.getSwarmStatus(name)
}

func (m *Manager) getSwarmStatus(name string) (*AccessoryStatus, error) {
	fullName := fmt.Sprintf("%s_%s", m.stackName, name)
	cmd := fmt.Sprintf("docker service ls --filter name=%s --format '{{.Name}}|{{.Replicas}}'", fullName)

	result, err := m.exec.Run(cmd)
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

func (m *Manager) getComposeStatus(name string) (*AccessoryStatus, error) {
	cmd := fmt.Sprintf("docker compose -p %s ps %s --format json", m.stackName, name)

	result, err := m.exec.Run(cmd)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(result.Stdout) == "" {
		return nil, fmt.Errorf("accessory %s not found", name)
	}

	// Parse JSON output (docker compose ps outputs one JSON per line)
	lines := strings.Split(strings.TrimSpace(result.Stdout), "\n")
	running := false
	count := 0

	for _, line := range lines {
		if line == "" {
			continue
		}

		var container struct {
			State string `json:"State"`
		}

		if err := json.Unmarshal([]byte(line), &container); err != nil {
			continue
		}

		count++
		if strings.ToLower(container.State) == "running" {
			running = true
		}
	}

	replicas := fmt.Sprintf("%d", count)
	if count == 0 {
		replicas = "not running"
	} else if running {
		replicas = fmt.Sprintf("%d/%d", count, count)
	} else {
		replicas = fmt.Sprintf("0/%d", count)
	}

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
