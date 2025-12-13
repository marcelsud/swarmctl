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

var statusCmd = &cobra.Command{
	Use:   "status [service]",
	Short: "Show stack status",
	Long: `Show the status of the stack and its services.
If a service name is provided, shows detailed status for that service.`,
	Run: runStatus,
}

func runStatus(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

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

	mgr := swarm.NewManager(exec, cfg.Stack)

	// Check if stack exists
	exists, err := mgr.StackExists()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to check stack: %v\n", red("✗"), err)
		os.Exit(1)
	}

	if !exists {
		fmt.Printf("%s Stack %s not found. Run 'swarmctl deploy' first.\n", yellow("!"), bold(cfg.Stack))
		os.Exit(0)
	}

	fmt.Printf("%s Stack: %s\n\n", cyan("→"), bold(cfg.Stack))

	// If service specified, show detailed status
	if len(args) > 0 {
		serviceName := args[0]
		showServiceStatus(mgr, serviceName, green, red, yellow, cyan)
		return
	}

	// Show all services
	fmt.Printf("%s Services:\n", cyan("→"))
	services, err := mgr.ListServices()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to list services: %v\n", red("✗"), err)
		os.Exit(1)
	}

	fmt.Printf("  %-30s %-12s %-15s %s\n", "NAME", "MODE", "REPLICAS", "PORTS")
	for _, svc := range services {
		replicaColor := green
		if !isHealthy(svc.Replicas) {
			replicaColor = yellow
		}
		fmt.Printf("  %-30s %-12s %s %-10s %s\n",
			svc.Name,
			svc.Mode,
			replicaColor(svc.Replicas),
			"",
			svc.Ports,
		)
	}

	// Show tasks
	fmt.Printf("\n%s Tasks:\n", cyan("→"))
	tasks, err := mgr.GetStackTasks()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Failed to list tasks: %v\n", err)
		return
	}

	fmt.Printf("  %-15s %-25s %-15s %-10s %s\n", "ID", "NAME", "NODE", "STATE", "ERROR")
	for _, task := range tasks {
		stateColor := green
		if strings.Contains(strings.ToLower(task.CurrentState), "failed") {
			stateColor = red
		} else if !strings.Contains(strings.ToLower(task.CurrentState), "running") {
			stateColor = yellow
		}

		errorMsg := ""
		if task.Error != "" {
			errorMsg = red(task.Error)
		}

		fmt.Printf("  %-15s %-25s %-15s %s %s\n",
			task.ID[:12],
			truncateName(task.Name, 25),
			task.Node,
			stateColor(task.CurrentState),
			errorMsg,
		)
	}
}

func showServiceStatus(mgr *swarm.Manager, serviceName string, green, red, yellow, cyan func(a ...interface{}) string) {
	fmt.Printf("%s Service: %s\n\n", cyan("→"), serviceName)

	tasks, err := mgr.GetServiceTasks(serviceName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to get service tasks: %v\n", red("✗"), err)
		os.Exit(1)
	}

	if len(tasks) == 0 {
		fmt.Printf("  %s No tasks found for service %s\n", yellow("!"), serviceName)
		return
	}

	fmt.Printf("  %-15s %-25s %-15s %-10s %s\n", "ID", "NAME", "NODE", "STATE", "ERROR")
	for _, task := range tasks {
		stateColor := green
		if strings.Contains(strings.ToLower(task.CurrentState), "failed") {
			stateColor = red
		} else if !strings.Contains(strings.ToLower(task.CurrentState), "running") {
			stateColor = yellow
		}

		errorMsg := ""
		if task.Error != "" {
			errorMsg = red(task.Error)
		}

		fmt.Printf("  %-15s %-25s %-15s %s %s\n",
			task.ID[:min(12, len(task.ID))],
			truncateName(task.Name, 25),
			task.Node,
			stateColor(task.CurrentState),
			errorMsg,
		)
	}
}

func isHealthy(replicas string) bool {
	// Format is like "3/3" or "0/3"
	parts := strings.Split(replicas, "/")
	if len(parts) != 2 {
		return false
	}
	return parts[0] == parts[1] && parts[0] != "0"
}

func truncateName(name string, maxLen int) string {
	if len(name) > maxLen {
		return name[:maxLen-3] + "..."
	}
	return name
}
