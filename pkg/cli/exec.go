package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/marcelsud/swarmctl/internal/config"
	"github.com/marcelsud/swarmctl/internal/deployment"
	"github.com/marcelsud/swarmctl/internal/executor"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec <service> [command]",
	Short: "Execute command in a container",
	Long: `Execute a command in a running container of the specified service.
If no command is provided, opens an interactive shell.

Examples:
  swarmctl exec web                    # Opens bash in web container
  swarmctl exec web -- ls -la          # Run ls -la in web container
  swarmctl exec api -- rails console   # Run rails console in api container`,
	Args: cobra.MinimumNArgs(1),
	Run:  runExec,
}

func runExec(cmd *cobra.Command, args []string) {
	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	serviceName := args[0]
	command := []string{"sh"}
	if len(args) > 1 {
		command = args[1:]
	}

	// Load config
	cfg, err := config.Load(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", red("✗"), err)
		os.Exit(1)
	}

	// Create executor
	exec, err := executor.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to connect: %v\n", red("✗"), err)
		os.Exit(1)
	}
	defer exec.Close()

	// Create deployment manager
	mgr := deployment.New(cfg, exec)

	// Find a running container for this service
	fmt.Printf("%s Finding container for service %s...\n", cyan("→"), serviceName)

	containerID, err := mgr.FindRunningContainer(serviceName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", red("✗"), err)
		os.Exit(1)
	}

	// Build docker exec command
	cmdStr := fmt.Sprintf("docker exec -it %s %s", containerID, strings.Join(command, " "))
	fmt.Printf("%s Executing: %s\n\n", cyan("→"), strings.Join(command, " "))

	// Run interactively
	if err := exec.RunInteractive(cmdStr); err != nil {
		fmt.Fprintf(os.Stderr, "\n%s Command failed: %v\n", red("✗"), err)
		os.Exit(1)
	}
}
