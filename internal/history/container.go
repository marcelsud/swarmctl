package history

import "fmt"

// ContainerName returns the history container name for a stack
func ContainerName(stackName string) string {
	return fmt.Sprintf("%s-history", stackName)
}

// VolumeName returns the history volume name for a stack
func VolumeName(stackName string) string {
	return fmt.Sprintf("%s_history_data", stackName)
}

// GenerateComposeService generates a compose service definition for the history container
// This can be used to include the history container in the compose file
func GenerateComposeService(stackName string) string {
	return fmt.Sprintf(`  %s-history:
    image: %s
    container_name: %s-history
    volumes:
      - %s_history_data:/data
    restart: unless-stopped
`, stackName, HistoryImage, stackName, stackName)
}

// GenerateComposeVolume generates a compose volume definition for the history data
func GenerateComposeVolume(stackName string) string {
	return fmt.Sprintf(`  %s_history_data:
    driver: local
`, stackName)
}
