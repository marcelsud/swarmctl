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

For Swarm mode, automatically detects if the container is running on a worker
node and uses SSH hop through the manager to reach it.

 Examples:
  swarmctl exec web                    # Opens shell in web container
  swarmctl exec web -- ls -la          # Run ls -la in web container
  swarmctl exec api -- rails console   # Run rails console in api container`,
	Args: cobra.MinimumNArgs(1),
	Run:  runExec,
}

func runExec(cmd *cobra.Command, args []string) {
	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

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

	// Find a running container for this service with node info
	fmt.Printf("%s Finding container for service %s...\n", cyan("→"), serviceName)

	containerInfo, err := mgr.FindRunningContainerWithNode(serviceName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", red("✗"), err)
		os.Exit(1)
	}

	// Build docker exec command
	dockerExecCmd := fmt.Sprintf("docker exec -it %s %s", containerInfo.ContainerID, strings.Join(command, " "))

	// Check if we need SSH hop (Swarm mode with container on different node)
	needsHop := false
	if cfg.Mode == config.ModeSwarm && containerInfo.NodeName != "" && !exec.IsLocal() {
		currentNode, err := mgr.GetCurrentNodeHostname()
		if err == nil && currentNode != containerInfo.NodeName {
			needsHop = true
		}
	}

	if needsHop {
		// Container is on a worker node, need SSH hop
		fmt.Printf("%s Container is on node %s (IP: %s)\n", yellow("⚡"), containerInfo.NodeName, containerInfo.NodeIP)

		// Get SSH executor to check for agent forwarding
		sshExec, ok := exec.(*executor.SSHExecutor)
		if !ok {
			fmt.Fprintf(os.Stderr, "%s SSH hop requires SSH executor\n", red("✗"))
			os.Exit(1)
		}

		if !sshExec.HasAgentForwarding() {
			fmt.Fprintf(os.Stderr, "%s SSH agent forwarding not available. Ensure ssh-agent is running and has your key loaded (ssh-add).\n", red("✗"))
			os.Exit(1)
		}

		// Determine SSH user for the worker node
		targetUser := cfg.SSH.User // fallback to manager user
		if nodeConfig, ok := cfg.Nodes[containerInfo.NodeName]; ok && nodeConfig.User != "" {
			targetUser = nodeConfig.User
		}

		fmt.Printf("%s SSH hop to %s@%s\n", cyan("→"), targetUser, containerInfo.NodeIP)
		fmt.Printf("%s Executing: %s\n\n", cyan("→"), strings.Join(command, " "))

		// Run via SSH hop
		if err := sshExec.RunInteractiveOnHost(containerInfo.NodeIP, targetUser, dockerExecCmd); err != nil {
			fmt.Fprintf(os.Stderr, "\n%s Command failed: %v\n", red("✗"), err)
			os.Exit(1)
		}
	} else {
		// Container is on current node or compose mode, exec directly
		fmt.Printf("%s Executing: %s\n\n", cyan("→"), strings.Join(command, " "))

		if err := exec.RunInteractive(dockerExecCmd); err != nil {
			fmt.Fprintf(os.Stderr, "\n%s Command failed: %v\n", red("✗"), err)
			os.Exit(1)
		}
	}
}
