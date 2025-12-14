package config

import (
	"fmt"
	"os"
	"strings"
)

// ValidationError holds multiple validation errors
type ValidationError struct {
	Errors []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed:\n  - %s", strings.Join(e.Errors, "\n  - "))
}

func (e *ValidationError) Add(msg string) {
	e.Errors = append(e.Errors, msg)
}

func (e *ValidationError) HasErrors() bool {
	return len(e.Errors) > 0
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	ve := &ValidationError{}

	// Required fields
	if c.Stack == "" {
		ve.Add("stack name is required")
	}

	// Validate mode
	if c.Mode != "" && c.Mode != ModeSwarm && c.Mode != ModeCompose {
		ve.Add(fmt.Sprintf("invalid mode '%s': must be 'swarm' or 'compose'", c.Mode))
	}

	// SSH is optional - if host is provided, user is required
	if c.SSH.Host != "" {
		if c.SSH.User == "" {
			ve.Add("ssh.user is required when ssh.host is set")
		}

		if c.SSH.Port <= 0 || c.SSH.Port > 65535 {
			ve.Add("ssh.port must be between 1 and 65535")
		}

		// Check if SSH key exists (if specified)
		if c.SSH.Key != "" {
			if _, err := os.Stat(c.SSH.Key); os.IsNotExist(err) {
				ve.Add(fmt.Sprintf("SSH key file not found: %s", c.SSH.Key))
			}
		}
	}

	// Check if compose file exists
	if c.ComposeFile != "" {
		if _, err := os.Stat(c.ComposeFile); os.IsNotExist(err) {
			ve.Add(fmt.Sprintf("compose file not found: %s", c.ComposeFile))
		}
	}

	if ve.HasErrors() {
		return ve
	}

	return nil
}
