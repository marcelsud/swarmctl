package deployment

import (
	"fmt"
	"io"
)

// ServiceStatus represents the status of a service
type ServiceStatus struct {
	Name     string
	Mode     string
	Replicas string
	Image    string
	Ports    string
}

// ContainerStatus represents the status of a container/task
type ContainerStatus struct {
	ID      string
	Name    string
	Service string
	State   string
	Error   string
}

// ContainerInfo holds information about a running container including its node
type ContainerInfo struct {
	ContainerID string
	NodeName    string
	NodeIP      string
}

// Manager defines the interface for deployment operations
type Manager interface {
	// Deploy deploys the stack/project
	Deploy(composeContent []byte) error

	// Remove removes the stack/project
	Remove() error

	// Exists checks if the stack/project is deployed
	Exists() (bool, error)

	// ListServices lists all services in the stack
	ListServices() ([]ServiceStatus, error)

	// GetServiceLogs retrieves logs from a service
	GetServiceLogs(serviceName string, follow bool, since string, tail int) (string, error)

	// StreamServiceLogs streams logs from a service to the provided writers
	StreamServiceLogs(serviceName string, follow bool, tail int, stdout, stderr io.Writer) error

	// FindRunningContainer finds a container ID for exec
	FindRunningContainer(serviceName string) (string, error)

	// FindRunningContainerWithNode finds a container ID and its node information
	FindRunningContainerWithNode(serviceName string) (*ContainerInfo, error)

	// GetCurrentNodeHostname returns the hostname of the current node
	GetCurrentNodeHostname() (string, error)

	// GetContainerStatus gets container/task status for all services
	GetContainerStatus() ([]ContainerStatus, error)

	// SupportsRollback returns true if rollback is supported
	SupportsRollback() bool

	// SupportsScale returns true if scale is supported
	SupportsScale() bool

	// RollbackService rolls back a service to its previous version
	RollbackService(serviceName string) error

	// RollbackAll rolls back all services in the stack
	RollbackAll() error

	// ScaleService scales a service to the specified replicas
	ScaleService(serviceName string, replicas int) error

	// GetStackName returns the stack/project name
	GetStackName() string

	// GetMode returns the deployment mode (swarm or compose)
	GetMode() string
}

// UnsupportedOperationError is returned when an operation is not supported
type UnsupportedOperationError struct {
	Operation string
	Mode      string
}

func (e *UnsupportedOperationError) Error() string {
	return fmt.Sprintf("%s is not supported in %s mode", e.Operation, e.Mode)
}

// NewUnsupportedError creates a new UnsupportedOperationError
func NewUnsupportedError(operation, mode string) error {
	return &UnsupportedOperationError{
		Operation: operation,
		Mode:      mode,
	}
}
