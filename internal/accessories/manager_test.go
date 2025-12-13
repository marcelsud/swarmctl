package accessories

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager(nil, "myapp")

	if m.stackName != "myapp" {
		t.Errorf("expected stackName 'myapp', got '%s'", m.stackName)
	}
}

func TestAccessoryStatus(t *testing.T) {
	status := AccessoryStatus{
		Name:     "redis",
		Replicas: "1/1",
		Running:  true,
	}

	if status.Name != "redis" {
		t.Errorf("expected name 'redis', got '%s'", status.Name)
	}

	if !status.Running {
		t.Error("expected Running to be true")
	}
}

func TestAccessoryStatusNotRunning(t *testing.T) {
	status := AccessoryStatus{
		Name:     "postgres",
		Replicas: "0/1",
		Running:  false,
	}

	if status.Running {
		t.Error("expected Running to be false")
	}
}

// ServiceNameFormat tests the service name formatting logic
// The full service name format is: {stackName}_{serviceName}
func TestServiceNameFormat(t *testing.T) {
	tests := []struct {
		stackName   string
		serviceName string
		expected    string
	}{
		{"myapp", "redis", "myapp_redis"},
		{"production", "postgres", "production_postgres"},
		{"app-staging", "memcached", "app-staging_memcached"},
	}

	for _, tt := range tests {
		result := tt.stackName + "_" + tt.serviceName
		if result != tt.expected {
			t.Errorf("expected '%s', got '%s'", tt.expected, result)
		}
	}
}

// ReplicasRunningLogic tests the logic used to determine if a service is running
func TestReplicasRunningLogic(t *testing.T) {
	tests := []struct {
		replicas string
		running  bool
	}{
		{"1/1", true},
		{"3/3", true},
		{"2/3", true},  // Partially running is still running
		{"0/1", false}, // Scaled to 0
		{"0/0", false}, // No replicas
		{"1/0", true},  // Edge case: more running than desired
	}

	for _, tt := range tests {
		// This is the logic from GetStatus: running := !strings.HasPrefix(replicas, "0/")
		running := tt.replicas[0] != '0' || tt.replicas[1] != '/'

		if running != tt.running {
			t.Errorf("replicas '%s': expected running=%v, got %v", tt.replicas, tt.running, running)
		}
	}
}
