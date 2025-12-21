package history

import (
	"testing"
)

func TestContainerName(t *testing.T) {
	tests := []struct {
		name      string
		stackName string
		expected  string
	}{
		{
			name:      "simple stack name",
			stackName: "webapp",
			expected:  "webapp-history",
		},
		{
			name:      "stack with hyphens",
			stackName: "my-app-prod",
			expected:  "my-app-prod-history",
		},
		{
			name:      "empty stack name",
			stackName: "",
			expected:  "-history",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainerName(tt.stackName)
			if result != tt.expected {
				t.Errorf("ContainerName() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestVolumeName(t *testing.T) {
	tests := []struct {
		name      string
		stackName string
		expected  string
	}{
		{
			name:      "simple stack name",
			stackName: "webapp",
			expected:  "webapp_history_data",
		},
		{
			name:      "stack with hyphens",
			stackName: "my-app-prod",
			expected:  "my-app-prod_history_data",
		},
		{
			name:      "empty stack name",
			stackName: "",
			expected:  "_history_data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VolumeName(tt.stackName)
			if result != tt.expected {
				t.Errorf("VolumeName() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGenerateComposeService(t *testing.T) {
	stackName := "webapp"
	expected := `  webapp-history:
    image: docker.io/marcelsud/swarmctl-history:latest
    container_name: webapp-history
    volumes:
      - webapp_history_data:/data
    restart: unless-stopped
`

	result := GenerateComposeService(stackName)
	if result != expected {
		t.Errorf("GenerateComposeService() = %q, want %q", result, expected)
	}

	// Check if it contains all required elements
	requiredElements := []string{
		"webapp-history:",
		"image: docker.io/marcelsud/swarmctl-history:latest",
		"container_name: webapp-history",
		"webapp_history_data:/data",
		"restart: unless-stopped",
	}

	for _, element := range requiredElements {
		if !containsString(result, element) {
			t.Errorf("Compose service should contain: %s", element)
		}
	}
}

func TestGenerateComposeVolume(t *testing.T) {
	stackName := "webapp"
	expected := `  webapp_history_data:
    driver: local
`

	result := GenerateComposeVolume(stackName)
	if result != expected {
		t.Errorf("GenerateComposeVolume() = %q, want %q", result, expected)
	}

	// Check if it contains all required elements
	requiredElements := []string{
		"webapp_history_data:",
		"driver: local",
	}

	for _, element := range requiredElements {
		if !containsString(result, element) {
			t.Errorf("Compose volume should contain: %s", element)
		}
	}
}

func TestGenerateComposeService_DifferentStackNames(t *testing.T) {
	tests := []struct {
		name      string
		stackName string
	}{
		{"single word", "api"},
		{"with hyphens", "my-app"},
		{"with numbers", "app-v2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateComposeService(tt.stackName)

			// Should contain the stack name history container
			expectedService := tt.stackName + "-history:"
			if !containsString(result, expectedService) {
				t.Errorf("Should contain service %s", expectedService)
			}

			// Should contain the correct volume name
			expectedVolume := tt.stackName + "_history_data:/data"
			if !containsString(result, expectedVolume) {
				t.Errorf("Should contain volume %s", expectedVolume)
			}

			// Should contain the container name
			expectedContainer := "container_name: " + tt.stackName + "-history"
			if !containsString(result, expectedContainer) {
				t.Errorf("Should contain container name %s", expectedContainer)
			}
		})
	}
}

func TestGenerateComposeVolume_DifferentStackNames(t *testing.T) {
	tests := []struct {
		name      string
		stackName string
	}{
		{"single word", "api"},
		{"with hyphens", "my-app"},
		{"with numbers", "app-v2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateComposeVolume(tt.stackName)

			expectedVolume := tt.stackName + "_history_data:"
			if !containsString(result, expectedVolume) {
				t.Errorf("Should contain volume %s", expectedVolume)
			}

			// Should always use local driver
			if !containsString(result, "driver: local") {
				t.Error("Should use local driver")
			}
		})
	}
}
