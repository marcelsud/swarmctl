package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	initLLMForce bool
	initLLMAll   bool
)

type llmFile struct {
	key         string
	path        string
	description string
	content     string
}

var llmFiles = []llmFile{
	{
		key:         "CLAUDE.md",
		path:        "CLAUDE.md",
		description: "Claude Code instructions",
		content:     claudeMDTemplate,
	},
	{
		key:         ".cursorrules",
		path:        ".cursorrules",
		description: "Cursor AI rules",
		content:     cursorRulesTemplate,
	},
	{
		key:         "AGENTS.md",
		path:        "AGENTS.md",
		description: "Agent documentation",
		content:     agentsMDTemplate,
	},
	{
		key:         "copilot-instructions.md",
		path:        ".github/copilot-instructions.md",
		description: "GitHub Copilot instructions",
		content:     copilotInstructionsTemplate,
	},
	{
		key:         "swarmctl skill",
		path:        ".claude/skills/swarmctl/SKILL.md",
		description: "Claude skill for swarmctl",
		content:     skillTemplate,
	},
}

var initLLMCmd = &cobra.Command{
	Use:   "init-llm",
	Short: "Initialize LLM documentation templates",
	Long: `Initialize LLM documentation templates for AI assistants.

Available files:
  - CLAUDE.md                           Claude Code instructions
  - .cursorrules                        Cursor AI rules
  - AGENTS.md                           Agent documentation
  - .github/copilot-instructions.md     GitHub Copilot instructions
  - .claude/skills/swarmctl/SKILL.md    Claude skill for swarmctl

Use --all to create all files without prompting.
Use --force to overwrite existing files.`,
	Run: runInitLLM,
}

func init() {
	initLLMCmd.Flags().BoolVar(&initLLMForce, "force", false, "overwrite existing files")
	initLLMCmd.Flags().BoolVar(&initLLMAll, "all", false, "create all files without prompting")
}

func runInitLLM(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	var selectedFiles []llmFile

	if initLLMAll {
		selectedFiles = llmFiles
	} else {
		selected, err := promptFileSelection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s %v\n", red("✗"), err)
			os.Exit(1)
		}

		if len(selected) == 0 {
			fmt.Printf("%s No files selected\n", yellow("!"))
			return
		}

		for _, f := range llmFiles {
			for _, s := range selected {
				if f.key == s {
					selectedFiles = append(selectedFiles, f)
					break
				}
			}
		}
	}

	var created, skipped []string
	var errors []string

	fmt.Printf("\n%s Creating files...\n\n", cyan("→"))

	for _, f := range selectedFiles {
		result := createLLMFile(f, initLLMForce)
		switch result.status {
		case "created":
			created = append(created, f.path)
			fmt.Printf("  %s %s\n", green("✓"), f.path)
		case "skipped":
			skipped = append(skipped, f.path)
			fmt.Printf("  %s %s (already exists)\n", yellow("-"), f.path)
		case "error":
			errors = append(errors, fmt.Sprintf("%s: %v", f.path, result.err))
			fmt.Printf("  %s %s: %v\n", red("✗"), f.path, result.err)
		}
	}

	fmt.Println()

	if len(created) > 0 {
		fmt.Printf("%s Created %d file(s)\n", green("✓"), len(created))
	}
	if len(skipped) > 0 {
		fmt.Printf("%s Skipped %d file(s) (use --force to overwrite)\n", yellow("!"), len(skipped))
	}
	if len(errors) > 0 {
		fmt.Printf("%s Failed to create %d file(s)\n", red("✗"), len(errors))
		os.Exit(1)
	}
}

func promptFileSelection() ([]string, error) {
	options := make([]string, len(llmFiles))
	for i, f := range llmFiles {
		options[i] = f.key
	}

	var selected []string
	prompt := &survey.MultiSelect{
		Message: "Select files to create:",
		Options: options,
		Default: options,
		Description: func(value string, index int) string {
			return llmFiles[index].description
		},
	}

	err := survey.AskOne(prompt, &selected)
	return selected, err
}

type fileResult struct {
	status string
	err    error
}

func createLLMFile(f llmFile, force bool) fileResult {
	if fileExists(f.path) && !force {
		return fileResult{status: "skipped"}
	}

	dir := filepath.Dir(f.path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fileResult{status: "error", err: err}
		}
	}

	if err := os.WriteFile(f.path, []byte(f.content), 0644); err != nil {
		return fileResult{status: "error", err: err}
	}

	return fileResult{status: "created"}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Templates

const claudeMDTemplate = `# [TODO: Project Name]

[TODO: Brief project description]

## Technologies

- [TODO: Primary language]
- [TODO: Frameworks]
- Docker Swarm (managed via swarmctl)

## Project Structure

` + "```" + `
[TODO: Directory structure]
` + "```" + `

## Deploy with swarmctl

This project uses [swarmctl](https://github.com/marcelsud/swarmctl) for Docker Swarm deployment.

### Configuration Files

| File | Description |
|------|-------------|
| swarm.yaml | Main config (SSH, registry, secrets) |
| docker-compose.yaml | Service definitions |
| .env | Secret values (do not commit) |

### Main Commands

` + "```bash" + `
swarmctl setup              # Initialize Swarm on server
swarmctl deploy             # Deploy the stack
swarmctl status             # Service status
swarmctl logs <service>     # View logs (-f for follow)
swarmctl rollback           # Rollback to previous version
swarmctl exec <service>     # Shell into container
swarmctl secrets push       # Push secrets from .env
` + "```" + `

### Multi-environment

` + "```bash" + `
swarmctl deploy -d staging      # Uses swarm.staging.yaml
swarmctl deploy -d production   # Uses swarm.production.yaml
` + "```" + `

## Build & Test

` + "```bash" + `
[TODO: Build commands]
[TODO: Test commands]
` + "```" + `

## Environment Variables

| Variable | Description |
|----------|-------------|
| [TODO: VAR_NAME] | [TODO: Description] |
`

const cursorRulesTemplate = `# Cursor Rules for [TODO: Project Name]

## Project Overview
[TODO: Brief project description]

## Technology Stack
- [TODO: Primary language]
- [TODO: Frameworks]
- Docker Swarm deployment via swarmctl

## Coding Conventions
[TODO: Add coding conventions]

## File Structure
[TODO: Describe important directories]

## Deployment with swarmctl

This project uses swarmctl for Docker Swarm deployments.

Key commands:
- swarmctl setup - Initialize Swarm
- swarmctl deploy - Deploy stack
- swarmctl status - Check status
- swarmctl logs <service> - View logs
- swarmctl rollback - Rollback deployment
- swarmctl exec <service> - Shell into container
- swarmctl secrets push - Push secrets

Configuration files:
- swarm.yaml - Main swarmctl config
- docker-compose.yaml - Service definitions
- .env - Secret values (do not commit)

Multi-environment:
- swarmctl deploy -d staging
- swarmctl deploy -d production

## Important Notes
[TODO: Add important notes]
`

const agentsMDTemplate = `# Agents Documentation

## Project: [TODO: Project Name]

[TODO: Brief project description]

## Available Workflows

### Deploy Workflow
Handles deployment to Docker Swarm using swarmctl.

**Commands:**
` + "```bash" + `
swarmctl setup              # Initialize Swarm
swarmctl deploy             # Deploy stack
swarmctl deploy -d staging  # Deploy to staging
swarmctl deploy -d production # Deploy to production
swarmctl status             # Check status
swarmctl rollback           # Rollback if needed
` + "```" + `

### [TODO: Additional Workflows]

[TODO: Describe other workflows]

## Configuration

### swarmctl Configuration

- **swarm.yaml**: Main configuration (SSH, registry, secrets)
- **docker-compose.yaml**: Service definitions
- **.env**: Secret values (not committed)

## Deployment Steps

1. Configure swarm.yaml and docker-compose.yaml
2. Run ` + "`swarmctl setup`" + ` to initialize
3. Run ` + "`swarmctl secrets push`" + ` to push secrets
4. Run ` + "`swarmctl deploy`" + ` to deploy
5. Monitor with ` + "`swarmctl status`" + ` and ` + "`swarmctl logs <service>`" + `
`

const copilotInstructionsTemplate = `# GitHub Copilot Instructions

## Project: [TODO: Project Name]

[TODO: Brief project description]

## Technology Stack
- [TODO: Primary language]
- [TODO: Frameworks]
- Docker Swarm (deployment via swarmctl)

## Code Style
[TODO: Describe coding style]

## Project Structure
[TODO: Key directories and purposes]

## Deployment with swarmctl

This project uses swarmctl for Docker Swarm deployment.

### Key Files
- swarm.yaml: swarmctl configuration
- docker-compose.yaml: Docker service definitions
- .env: Secret values (never commit)

### Common Commands
- swarmctl setup - Initialize Swarm
- swarmctl deploy - Deploy stack
- swarmctl status - Service status
- swarmctl logs <service> - View logs
- swarmctl rollback - Rollback deployment
- swarmctl exec <service> - Shell into container
- swarmctl secrets push - Push secrets

### Multi-environment
- swarmctl deploy -d staging
- swarmctl deploy -d production

## Testing
[TODO: Describe testing approach]

## Conventions
[TODO: List conventions]
`

const skillTemplate = `---
name: swarmctl
description: CLI for Docker Swarm deployment and management. Use when deploying, checking status, viewing logs, exec into containers, managing secrets/accessories, or rollback.
allowed-tools: Read, Grep, Glob, Bash
---

# swarmctl CLI Helper

CLI for deploying and managing Docker Swarm stacks, inspired by Kamal.

## Main Commands

| Command | Description |
|---------|-------------|
| ` + "`swarmctl setup`" + ` | Initialize Swarm on server |
| ` + "`swarmctl deploy`" + ` | Deploy the stack |
| ` + "`swarmctl status`" + ` | Service status |
| ` + "`swarmctl logs <service>`" + ` | View logs (-f for follow) |
| ` + "`swarmctl rollback`" + ` | Rollback to previous version |
| ` + "`swarmctl exec <service>`" + ` | Shell into container |
| ` + "`swarmctl secrets push`" + ` | Push secrets from .env |
| ` + "`swarmctl secrets list`" + ` | List secrets |
| ` + "`swarmctl accessory start <name>`" + ` | Start accessory |
| ` + "`swarmctl accessory stop <name>`" + ` | Stop accessory |

## Configuration Files

| File | Description |
|------|-------------|
| swarm.yaml | Main config (SSH, registry, secrets, nodes) |
| docker-compose.yaml | Service definitions (standard Docker format) |
| .env | Secret values (do not commit) |

## Multi-environment

` + "```bash" + `
swarmctl deploy -d staging      # Uses swarm.staging.yaml
swarmctl deploy -d production   # Uses swarm.production.yaml
` + "```" + `

## Exec on Worker Nodes

swarmctl exec auto-detects containers on worker nodes and performs SSH hop through manager:

` + "```" + `
swarmctl -> SSH manager -> SSH worker (internal IP) -> docker exec
` + "```" + `

Requirements:
- ssh-agent running with key loaded (` + "`ssh-add`" + `)
- Workers accessible via internal IP from manager
- SSH enabled on workers

## Instructions

1. When user mentions swarmctl, identify the task
2. Use Read/Grep to inspect swarm.yaml and docker-compose.yaml
3. Provide correct swarmctl commands with appropriate flags
4. For debugging, suggest ` + "`swarmctl status`" + ` and ` + "`swarmctl logs <service>`" + `
`
