package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/marcelsud/swarmctl/internal/config"
	"github.com/marcelsud/swarmctl/internal/executor"
	"github.com/marcelsud/swarmctl/internal/swarm"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup the deployment environment",
	Long: `Setup the deployment environment.

For Swarm mode:
- Verify Docker is installed
- Initialize Swarm if necessary
- Create overlay network for the stack
- Login to the registry

For Compose mode:
- Verify Docker is installed
- Verify docker compose plugin is available
- Login to the registry`,
	Run: runSetup,
}

func runSetup(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	// Load config
	fmt.Printf("%s Loading configuration...\n", cyan("→"))
	cfg, err := config.Load(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", red("✗"), err)
		os.Exit(1)
	}

	// Validate config (skip compose file check for setup)
	if cfg.Stack == "" {
		fmt.Fprintf(os.Stderr, "%s Invalid configuration: stack is required\n", red("✗"))
		os.Exit(1)
	}

	fmt.Printf("  Stack: %s\n", cfg.Stack)

	// Show deployment mode
	modeStr := "swarm"
	if cfg.Mode == config.ModeCompose {
		modeStr = "compose"
	}
	fmt.Printf("  Mode:  %s\n", bold(modeStr))

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
		fmt.Printf("  Host:  %s@%s:%d\n", cfg.SSH.User, cfg.SSH.Host, cfg.SSH.Port)
		fmt.Printf("  %s Connected\n", green("✓"))
	}

	mgr := swarm.NewManager(exec, cfg.Stack)

	// Check Docker installation
	fmt.Printf("%s Checking Docker installation...\n", cyan("→"))
	installed, err := mgr.IsDockerInstalled()
	if err != nil || !installed {
		fmt.Fprintf(os.Stderr, "%s Docker is not installed\n", red("✗"))
		os.Exit(1)
	}

	version, _ := mgr.GetDockerVersion()
	fmt.Printf("  %s %s\n", green("✓"), version)

	// Handle based on mode
	if cfg.Mode == config.ModeCompose {
		// Compose mode setup
		runSetupCompose(exec, cfg, mgr, green, red, cyan)
	} else {
		// Swarm mode setup
		runSetupSwarm(exec, cfg, mgr, green, yellow, red, cyan)
	}

	fmt.Printf("\n%s Setup complete! Run 'swarmctl deploy' to deploy your stack.\n", green("✓"))
}

func runSetupCompose(exec executor.Executor, cfg *config.Config, mgr *swarm.Manager, green, red, cyan func(a ...interface{}) string) {
	// Check docker compose plugin
	fmt.Printf("%s Checking docker compose plugin...\n", cyan("→"))
	result, err := exec.Run("docker compose version")
	if err != nil || result.ExitCode != 0 {
		fmt.Fprintf(os.Stderr, "%s docker compose plugin not found\n", red("✗"))
		fmt.Fprintf(os.Stderr, "  Install it with: apt install docker-compose-plugin\n")
		os.Exit(1)
	}
	fmt.Printf("  %s %s\n", green("✓"), strings.TrimSpace(result.Stdout))

	// Login to registry
	if cfg.Registry.URL != "" && cfg.Registry.Username != "" {
		fmt.Printf("%s Logging into registry %s...\n", cyan("→"), cfg.Registry.URL)
		if err := mgr.RegistryLogin(cfg.Registry.URL, cfg.Registry.Username, cfg.Registry.Password); err != nil {
			fmt.Fprintf(os.Stderr, "%s Failed to login to registry: %v\n", red("✗"), err)
			os.Exit(1)
		}
		fmt.Printf("  %s Logged in as %s\n", green("✓"), cfg.Registry.Username)
	}
}

func runSetupSwarm(exec executor.Executor, cfg *config.Config, mgr *swarm.Manager, green, yellow, red, cyan func(a ...interface{}) string) {
	// Check/Initialize Swarm
	fmt.Printf("%s Checking Swarm status...\n", cyan("→"))
	initialized, err := mgr.IsSwarmInitialized()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to check Swarm status: %v\n", red("✗"), err)
		os.Exit(1)
	}

	if !initialized {
		fmt.Printf("  %s Swarm not initialized, initializing...\n", yellow("!"))
		if err := mgr.InitSwarm(); err != nil {
			fmt.Fprintf(os.Stderr, "%s Failed to initialize Swarm: %v\n", red("✗"), err)
			os.Exit(1)
		}
		fmt.Printf("  %s Swarm initialized\n", green("✓"))
	} else {
		fmt.Printf("  %s Swarm already initialized\n", green("✓"))
	}

	// Create overlay network
	networkName := cfg.Stack + "-network"
	fmt.Printf("%s Creating network %s...\n", cyan("→"), networkName)
	if err := mgr.CreateNetwork(networkName); err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to create network: %v\n", red("✗"), err)
		os.Exit(1)
	}
	fmt.Printf("  %s Network ready\n", green("✓"))

	// Login to registry
	if cfg.Registry.URL != "" && cfg.Registry.Username != "" {
		fmt.Printf("%s Logging into registry %s...\n", cyan("→"), cfg.Registry.URL)
		if err := mgr.RegistryLogin(cfg.Registry.URL, cfg.Registry.Username, cfg.Registry.Password); err != nil {
			fmt.Fprintf(os.Stderr, "%s Failed to login to registry: %v\n", red("✗"), err)
			os.Exit(1)
		}
		fmt.Printf("  %s Logged in as %s\n", green("✓"), cfg.Registry.Username)
	}

	// Show node info
	fmt.Printf("\n%s Swarm nodes:\n", cyan("→"))
	nodeInfo, err := mgr.GetNodeInfo()
	if err == nil {
		fmt.Println(nodeInfo)
	}
}
