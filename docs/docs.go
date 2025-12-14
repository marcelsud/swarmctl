package docs

import (
	"embed"
	"fmt"
	"sort"
)

//go:embed *.md
var content embed.FS

// Topic represents a documentation topic
type Topic struct {
	Name        string
	File        string
	Description string
}

// Topics returns all available documentation topics
var Topics = []Topic{
	{Name: "getting-started", File: "getting-started.md", Description: "Installation and first deploy"},
	{Name: "commands", File: "commands.md", Description: "Command reference"},
	{Name: "configuration", File: "configuration.md", Description: "swarm.yaml configuration"},
	{Name: "compose-mode", File: "compose-mode.md", Description: "Deploy with Docker Compose"},
	{Name: "local-mode", File: "local-mode.md", Description: "Run locally without SSH"},
	{Name: "multi-environment", File: "multi-environment.md", Description: "Staging/Production setup"},
}

// Get returns the content of a topic by name
func Get(name string) (string, error) {
	for _, t := range Topics {
		if t.Name == name {
			data, err := content.ReadFile(t.File)
			if err != nil {
				return "", fmt.Errorf("failed to read %s: %w", t.File, err)
			}
			return string(data), nil
		}
	}
	return "", fmt.Errorf("unknown topic: %s", name)
}

// List returns sorted topic names
func List() []string {
	names := make([]string, len(Topics))
	for i, t := range Topics {
		names[i] = t.Name
	}
	sort.Strings(names)
	return names
}
