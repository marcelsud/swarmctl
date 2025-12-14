package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/marcelsud/swarmctl/internal/config"
	"github.com/marcelsud/swarmctl/internal/deployment"
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

	// Create deployment manager
	mgr := deployment.New(cfg, exec)

	// Check if stack/project exists
	exists, err := mgr.Exists()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to check stack: %v\n", red("✗"), err)
		os.Exit(1)
	}

	if !exists {
		fmt.Printf("%s Stack %s not found. Run 'swarmctl deploy' first.\n", yellow("!"), bold(cfg.Stack))
		os.Exit(0)
	}

	// Show stack info with mode
	modeStr := mgr.GetMode()
	fmt.Printf("%s Stack: %s (%s mode)\n\n", cyan("→"), bold(cfg.Stack), modeStr)

	// If service specified, show detailed status
	if len(args) > 0 {
		serviceName := args[0]
		showServiceStatus(cfg, exec, mgr, serviceName, green, red, yellow, cyan)
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

	// Show tasks/containers based on mode
	if cfg.Mode == config.ModeCompose {
		showComposeContainers(mgr, green, red, yellow, cyan)
	} else {
		showSwarmTasks(exec, cfg.Stack, green, red, yellow, cyan)
	}
}

func showComposeContainers(mgr deployment.Manager, green, red, yellow, cyan func(a ...interface{}) string) {
	fmt.Printf("\n%s Containers:\n", cyan("→"))
	containers, err := mgr.GetContainerStatus()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Failed to list containers: %v\n", err)
		return
	}

	if len(containers) == 0 {
		fmt.Printf("  No containers running\n")
		return
	}

	fmt.Printf("  %-15s %-25s %-20s %s\n", "ID", "NAME", "SERVICE", "STATE")
	for _, c := range containers {
		stateColor := green
		if strings.Contains(strings.ToLower(c.State), "exit") {
			stateColor = red
		} else if !strings.Contains(strings.ToLower(c.State), "running") {
			stateColor = yellow
		}

		containerID := c.ID
		if len(containerID) > 12 {
			containerID = containerID[:12]
		}

		fmt.Printf("  %-15s %-25s %-20s %s\n",
			containerID,
			truncateName(c.Name, 25),
			c.Service,
			stateColor(c.State),
		)
	}
}

func showSwarmTasks(exec executor.Executor, stackName string, green, red, yellow, cyan func(a ...interface{}) string) {
	swarmMgr := swarm.NewManager(exec, stackName)

	fmt.Printf("\n%s Tasks:\n", cyan("→"))
	tasks, err := swarmMgr.GetStackTasks()
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

		taskID := task.ID
		if len(taskID) > 12 {
			taskID = taskID[:12]
		}

		fmt.Printf("  %-15s %-25s %-15s %s %s\n",
			taskID,
			truncateName(task.Name, 25),
			task.Node,
			stateColor(task.CurrentState),
			errorMsg,
		)
	}
}

func showServiceStatus(cfg *config.Config, exec executor.Executor, mgr deployment.Manager, serviceName string, green, red, yellow, cyan func(a ...interface{}) string) {
	fmt.Printf("%s Service: %s\n\n", cyan("→"), serviceName)

	if cfg.Mode == config.ModeCompose {
		// Compose mode: show containers for this service
		containers, err := mgr.GetContainerStatus()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Failed to get containers: %v\n", red("✗"), err)
			os.Exit(1)
		}

		// Filter containers for this service
		var serviceContainers []deployment.ContainerStatus
		for _, c := range containers {
			if c.Service == serviceName {
				serviceContainers = append(serviceContainers, c)
			}
		}

		if len(serviceContainers) == 0 {
			fmt.Printf("  %s No containers found for service %s\n", yellow("!"), serviceName)
			return
		}

		fmt.Printf("  %-15s %-25s %-20s %s\n", "ID", "NAME", "SERVICE", "STATE")
		for _, c := range serviceContainers {
			stateColor := green
			if strings.Contains(strings.ToLower(c.State), "exit") {
				stateColor = red
			} else if !strings.Contains(strings.ToLower(c.State), "running") {
				stateColor = yellow
			}

			containerID := c.ID
			if len(containerID) > 12 {
				containerID = containerID[:12]
			}

			fmt.Printf("  %-15s %-25s %-20s %s\n",
				containerID,
				truncateName(c.Name, 25),
				c.Service,
				stateColor(c.State),
			)
		}
	} else {
		// Swarm mode: show tasks for this service
		swarmMgr := swarm.NewManager(exec, cfg.Stack)
		tasks, err := swarmMgr.GetServiceTasks(serviceName)
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

			taskID := task.ID
			if len(taskID) > 12 {
				taskID = taskID[:12]
			}

			fmt.Printf("  %-15s %-25s %-15s %s %s\n",
				taskID,
				truncateName(task.Name, 25),
				task.Node,
				stateColor(task.CurrentState),
				errorMsg,
			)
		}
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
