package swarm

import (
	"fmt"
	"strings"
)

// TaskStatus represents the status of a task (container instance)
type TaskStatus struct {
	ID           string
	Name         string
	Image        string
	Node         string
	DesiredState string
	CurrentState string
	Error        string
}

// GetStackTasks gets all tasks for the stack
func (m *Manager) GetStackTasks() ([]TaskStatus, error) {
	cmd := fmt.Sprintf("docker stack ps %s --format '{{.ID}}|{{.Name}}|{{.Image}}|{{.Node}}|{{.DesiredState}}|{{.CurrentState}}|{{.Error}}'", m.stackName)
	result, err := m.exec.Run(cmd)
	if err != nil {
		return nil, err
	}

	if result.Stdout == "" {
		return []TaskStatus{}, nil
	}

	var tasks []TaskStatus
	lines := strings.Split(strings.TrimSpace(result.Stdout), "\n")
	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) >= 6 {
			tasks = append(tasks, TaskStatus{
				ID:           parts[0],
				Name:         parts[1],
				Image:        parts[2],
				Node:         parts[3],
				DesiredState: parts[4],
				CurrentState: parts[5],
				Error:        getOrEmpty(parts, 6),
			})
		}
	}

	return tasks, nil
}

// GetServiceTasks gets tasks for a specific service
func (m *Manager) GetServiceTasks(serviceName string) ([]TaskStatus, error) {
	fullName := fmt.Sprintf("%s_%s", m.stackName, serviceName)
	cmd := fmt.Sprintf("docker service ps %s --format '{{.ID}}|{{.Name}}|{{.Image}}|{{.Node}}|{{.DesiredState}}|{{.CurrentState}}|{{.Error}}'", fullName)

	result, err := m.exec.Run(cmd)
	if err != nil {
		return nil, err
	}

	if result.Stdout == "" {
		return []TaskStatus{}, nil
	}

	var tasks []TaskStatus
	lines := strings.Split(strings.TrimSpace(result.Stdout), "\n")
	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) >= 6 {
			tasks = append(tasks, TaskStatus{
				ID:           parts[0],
				Name:         parts[1],
				Image:        parts[2],
				Node:         parts[3],
				DesiredState: parts[4],
				CurrentState: parts[5],
				Error:        getOrEmpty(parts, 6),
			})
		}
	}

	return tasks, nil
}

// WaitForConvergence waits for all services to converge
func (m *Manager) WaitForConvergence(timeout int) error {
	// Simple implementation: check if all tasks are running
	cmd := fmt.Sprintf("docker stack ps %s --filter 'desired-state=running' --format '{{.CurrentState}}' | grep -v Running | wc -l", m.stackName)

	result, err := m.exec.Run(cmd)
	if err != nil {
		return err
	}

	notRunning := strings.TrimSpace(result.Stdout)
	if notRunning != "0" {
		return fmt.Errorf("some tasks are not running yet")
	}

	return nil
}

// GetNodeInfo gets information about Swarm nodes
func (m *Manager) GetNodeInfo() (string, error) {
	result, err := m.exec.Run("docker node ls --format 'table {{.Hostname}}\t{{.Status}}\t{{.Availability}}\t{{.ManagerStatus}}'")
	if err != nil {
		return "", err
	}
	return result.Stdout, nil
}
