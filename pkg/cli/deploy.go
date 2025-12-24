package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/marcelsud/swarmctl/internal/config"
	"github.com/marcelsud/swarmctl/internal/deployment"
	"github.com/marcelsud/swarmctl/internal/executor"
	"github.com/marcelsud/swarmctl/internal/secrets"
	"github.com/marcelsud/swarmctl/internal/swarm"
	"github.com/spf13/cobra"
)

var (
	deployService         string
	deploySkipAccessories bool
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy the stack",
	Long: `Deploy the stack using Docker Swarm or docker compose.
This command will:
- Load and validate configuration
- Connect via SSH (if configured)
- Push secrets if changed
- Deploy the stack (swarm mode or compose mode)
- Wait for services to become healthy
- Show final status`,
	Run: runDeploy,
}

func init() {
	deployCmd.Flags().StringVarP(&deployService, "service", "s", "", "deploy only this service")
	deployCmd.Flags().BoolVar(&deploySkipAccessories, "skip-accessories", false, "skip accessory services")
}

func runDeploy(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	startTime := time.Now()

	// Load config
	fmt.Printf("%s Loading configuration...\n", cyan("→"))
	cfg, err := config.Load(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", red("✗"), err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", red("✗"), err)
		os.Exit(1)
	}

	fmt.Printf("  Stack: %s\n", bold(cfg.Stack))

	// Show deployment mode
	modeStr := "swarm"
	if cfg.Mode == config.ModeCompose {
		modeStr = "compose"
	}
	fmt.Printf("  Mode:  %s\n", bold(modeStr))

	// Load compose file
	fmt.Printf("%s Loading compose file...\n", cyan("→"))
	composeContent, err := config.LoadComposeFile(cfg.ComposeFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", red("✗"), err)
		os.Exit(1)
	}
	fmt.Printf("  %s %s\n", green("✓"), cfg.ComposeFile)

	// Create executor
	exec, err := executor.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to connect: %v\n", red("✗"), err)
		os.Exit(1)
	}
	defer exec.Close()

	// Set verbose mode if requested
	exec.SetVerbose(verbose)

	if exec.IsLocal() {
		fmt.Printf("%s Running locally\n", cyan("→"))
	} else {
		fmt.Printf("%s Connected to %s\n", green("✓"), cfg.SSH.Host)
	}

	// Create deployment manager
	mgr := deployment.New(cfg, exec)

	// Push secrets if any are defined and we're in Swarm mode
	if len(cfg.Secrets) > 0 {
		fmt.Printf("%s Checking secrets...\n", cyan("→"))

		// Check current secrets on the remote
		secretsMgr := secrets.NewManager(exec, cfg.Stack)
		existingSecrets, err := secretsMgr.List()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Failed to check existing secrets: %v\n", yellow("!"), err)
			// Continue with deploy even if we can't check secrets
		} else {
			fmt.Printf("  %d secrets exist on remote\n", len(existingSecrets))
		}

		// Load secrets from .env file or environment
		var secretList []secrets.Secret
		if _, err := os.Stat(".env"); err == nil {
			secretList, err = secrets.LoadFromEnvFile(".env", cfg.Secrets)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s Failed to load secrets from .env: %v\n", yellow("!"), err)
			} else {
				fmt.Printf("  %d secrets loaded from .env\n", len(secretList))
			}
		} else {
			secretList = secrets.LoadFromEnv(cfg.Secrets)
			fmt.Printf("  %d secrets loaded from environment\n", len(secretList))
		}

		// Push secrets if we found any
		if len(secretList) > 0 {
			fmt.Printf("%s Pushing %d secret(s)...\n", cyan("→"), len(secretList))

			for _, secret := range secretList {
				fmt.Printf("  %s %s...", cyan("→"), secret.Name)
				if err := secretsMgr.Create(secret.Name, secret.Value); err != nil {
					fmt.Printf(" %s (%v)\n", red("✗"), err)
					// Continue with deploy even if some secrets fail
				} else {
					fmt.Printf(" %s\n", green("✓"))
				}
			}
		}
	}

	// Login to registry if configured (use swarm manager for registry login)
	if cfg.Registry.URL != "" && cfg.Registry.Username != "" {
		fmt.Printf("%s Logging into registry...\n", cyan("→"))
		swarmMgr := swarm.NewManager(exec, cfg.Stack)
		if err := swarmMgr.RegistryLogin(cfg.Registry.URL, cfg.Registry.Username, cfg.Registry.Password); err != nil {
			fmt.Fprintf(os.Stderr, "%s Failed to login: %v\n", red("✗"), err)
			os.Exit(1)
		}
		fmt.Printf("  %s Logged in\n", green("✓"))
	}

	// Deploy stack
	if deployService != "" {
		fmt.Printf("%s Deploying service %s in stack %s...\n", cyan("→"), deployService, bold(cfg.Stack))
	} else {
		fmt.Printf("%s Deploying stack %s...\n", cyan("→"), bold(cfg.Stack))
	}

	// Remove accessories from compose if skip flag is set
	if deploySkipAccessories {
		fmt.Printf("%s Filtering out accessories...\n", cyan("→"))
		
		// For now, just show a warning and continue
		// TODO: Implement proper YAML filtering
		fmt.Printf("  %s Warning: Skip accessories not fully implemented - deploying all services\n", yellow("!"))
	}

	// Filter compose content if specific service
	if deployService != "" {
		composeContent = filterServiceFromCompose(composeContent, deployService)
	}

	if err := mgr.Deploy(composeContent); err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to deploy: %v\n", red("✗"), err)
		os.Exit(1)
	}
	fmt.Printf("  %s Stack deployed\n", green("✓"))

	// Wait for services to become healthy
	fmt.Printf("%s Waiting for services to become healthy...\n", cyan("→"))
	timeout := 2 * time.Minute
	if err := mgr.WaitForHealthy(timeout); err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", red("✗"), err)
		fmt.Printf("%s Deploy completed with health check timeout\n", yellow("!"))
	} else {
		fmt.Printf("  %s All services are healthy\n", green("✓"))
	}

	// Show status
	fmt.Printf("\n%s Services:\n", cyan("→"))
	services, err := mgr.ListServices()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Failed to list services: %v\n", err)
	} else {
		fmt.Printf("  %-30s %-12s %-15s %s\n", "NAME", "MODE", "REPLICAS", "IMAGE")
		for _, svc := range services {
			fmt.Printf("  %-30s %-12s %-15s %s\n", svc.Name, svc.Mode, svc.Replicas, truncateImage(svc.Image))
		}
	}

	elapsed := time.Since(startTime)
	fmt.Printf("\n%s Deploy completed in %s\n", green("✓"), elapsed.Round(time.Millisecond))
}

func truncateImage(image string) string {
	if len(image) > 50 {
		return image[:47] + "..."
	}
	return image
}

// filterServiceFromCompose removes all services except the specified one from compose content
func filterServiceFromCompose(composeContent []byte, serviceName string) []byte {
	// Parse the compose content to extract only the specified service
	// This is a simple implementation - for production, use a proper YAML parser

	content := string(composeContent)
	lines := strings.Split(content, "\n")

	var result []string
	inServices := false
	inTargetService := false
	serviceIndent := ""

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if we're entering services section
		if trimmed == "services:" {
			inServices = true
			result = append(result, line)
			continue
		}

		// If we're in services section
		if inServices {
			// Check if this is a service definition (ends with : and has proper indent)
			if strings.HasSuffix(trimmed, ":") && !strings.HasPrefix(trimmed, "#") {
				// Extract indent level
				indent := ""
				for _, ch := range line {
					if ch == ' ' || ch == '\t' {
						indent += string(ch)
					} else {
						break
					}
				}

				// Check if this is our target service
				serviceNameFromLine := strings.TrimSuffix(trimmed, ":")
				if serviceNameFromLine == serviceName {
					inTargetService = true
					serviceIndent = indent
					result = append(result, line)
				} else {
					inTargetService = false
				}
				continue
			}

			// If we're in the target service, include all lines with same or greater indent
			if inTargetService {
				currentIndent := ""
				for _, ch := range line {
					if ch == ' ' || ch == '\t' {
						currentIndent += string(ch)
					} else {
						break
					}
				}

				// Include line if it has same indent as service or greater (sub-items)
				if len(currentIndent) >= len(serviceIndent) {
					result = append(result, line)
				} else if trimmed == "" {
					// Empty lines are okay
					result = append(result, line)
				} else {
					// We've left the service definition
					inTargetService = false
				}
			}
		} else {
			// Not in services section, include everything
			result = append(result, line)
		}
	}

	// Reconstruct the compose content
	filteredContent := strings.Join(result, "\n")

	// Add basic services structure if we only have the target service
	if !strings.Contains(filteredContent, "services:") {
		filteredContent = "services:\n" + filteredContent
	}

	fmt.Printf("  Filtered compose to deploy only service: %s\n", serviceName)
	return []byte(filteredContent)
}


