package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/marcelsud/swarmctl/internal/config"
	"github.com/marcelsud/swarmctl/internal/executor"
	"github.com/marcelsud/swarmctl/internal/swarm"
	"github.com/spf13/cobra"
)

var (
	deployService         string
	deploySkipAccessories bool
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy the stack to Swarm",
	Long: `Deploy the stack to the Swarm cluster.
This command will:
- Load and validate configuration
- Connect via SSH
- Push secrets if changed
- Run docker stack deploy
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

	if exec.IsLocal() {
		fmt.Printf("%s Running locally\n", cyan("→"))
	} else {
		fmt.Printf("%s Connected to %s\n", green("✓"), cfg.SSH.Host)
	}

	mgr := swarm.NewManager(exec, cfg.Stack)

	// Login to registry if configured
	if cfg.Registry.URL != "" && cfg.Registry.Username != "" {
		fmt.Printf("%s Logging into registry...\n", cyan("→"))
		if err := mgr.RegistryLogin(cfg.Registry.URL, cfg.Registry.Username, cfg.Registry.Password); err != nil {
			fmt.Fprintf(os.Stderr, "%s Failed to login: %v\n", red("✗"), err)
			os.Exit(1)
		}
		fmt.Printf("  %s Logged in\n", green("✓"))
	}

	// Deploy stack
	fmt.Printf("%s Deploying stack %s...\n", cyan("→"), bold(cfg.Stack))
	if err := mgr.DeployStack(composeContent); err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to deploy: %v\n", red("✗"), err)
		os.Exit(1)
	}
	fmt.Printf("  %s Stack deployed\n", green("✓"))

	// Wait a moment for services to start
	fmt.Printf("%s Waiting for services to start...\n", cyan("→"))
	time.Sleep(5 * time.Second)

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
