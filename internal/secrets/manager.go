package secrets

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/marcelsud/swarmctl/internal/ssh"
)

// Manager handles Docker Swarm secrets
type Manager struct {
	client    *ssh.Client
	stackName string
}

// NewManager creates a new secrets manager
func NewManager(client *ssh.Client, stackName string) *Manager {
	return &Manager{
		client:    client,
		stackName: stackName,
	}
}

// Secret represents a secret with its name and value
type Secret struct {
	Name  string
	Value string
}

// LoadFromEnvFile loads secrets from a .env file
func LoadFromEnvFile(path string, secretNames []string) ([]Secret, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open .env file: %w", err)
	}
	defer file.Close()

	envMap := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove quotes if present
			value = strings.Trim(value, "\"'")
			envMap[key] = value
		}
	}

	var secrets []Secret
	for _, name := range secretNames {
		value, exists := envMap[name]
		if !exists {
			// Try environment variable
			value = os.Getenv(name)
		}
		if value != "" {
			secrets = append(secrets, Secret{Name: name, Value: value})
		}
	}

	return secrets, nil
}

// LoadFromEnv loads secrets from environment variables
func LoadFromEnv(secretNames []string) []Secret {
	var secrets []Secret
	for _, name := range secretNames {
		value := os.Getenv(name)
		if value != "" {
			secrets = append(secrets, Secret{Name: name, Value: value})
		}
	}
	return secrets
}

// Create creates a new secret in Docker Swarm
func (m *Manager) Create(name, value string) error {
	secretName := fmt.Sprintf("%s_%s", m.stackName, strings.ToLower(name))

	// Check if secret exists
	checkCmd := fmt.Sprintf("docker secret ls --filter name=%s --format '{{.Name}}'", secretName)
	result, err := m.client.Run(checkCmd)
	if err != nil {
		return err
	}

	// Remove existing secret if it exists
	if strings.TrimSpace(result.Stdout) != "" {
		rmCmd := fmt.Sprintf("docker secret rm %s", secretName)
		m.client.Run(rmCmd)
	}

	// Create secret using echo and pipe
	createCmd := fmt.Sprintf("echo -n '%s' | docker secret create %s -", value, secretName)
	result, err = m.client.Run(createCmd)
	if err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("secret creation failed: %s", result.Stderr)
	}

	return nil
}

// List lists all secrets for the stack
func (m *Manager) List() ([]string, error) {
	cmd := fmt.Sprintf("docker secret ls --filter name=%s_ --format '{{.Name}}'", m.stackName)
	result, err := m.client.Run(cmd)
	if err != nil {
		return nil, err
	}

	if result.Stdout == "" {
		return []string{}, nil
	}

	secrets := strings.Split(strings.TrimSpace(result.Stdout), "\n")
	return secrets, nil
}

// Delete deletes a secret
func (m *Manager) Delete(name string) error {
	secretName := fmt.Sprintf("%s_%s", m.stackName, strings.ToLower(name))
	cmd := fmt.Sprintf("docker secret rm %s", secretName)

	result, err := m.client.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("secret deletion failed: %s", result.Stderr)
	}

	return nil
}
