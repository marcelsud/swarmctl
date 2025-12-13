package cli

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/marcelsud/swarmctl/internal/config"
	"github.com/marcelsud/swarmctl/internal/ssh"
	"github.com/marcelsud/swarmctl/internal/swarm"
	"github.com/spf13/cobra"
)

var rollbackService string

var rollbackCmd = &cobra.Command{
	Use:   "rollback [service]",
	Short: "Rollback to the previous version",
	Long: `Rollback services to their previous version.
If a service name is provided, only that service will be rolled back.
Otherwise, all services in the stack will be rolled back.`,
	Run: runRollback,
}

func init() {
	rollbackCmd.Flags().StringVarP(&rollbackService, "service", "s", "", "rollback only this service (deprecated, use argument)")
}

func runRollback(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	// Get service from args or flag
	targetService := rollbackService
	if len(args) > 0 {
		targetService = args[0]
	}

	// Load config
	cfg, err := config.Load(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", red("✗"), err)
		os.Exit(1)
	}

	// Connect via SSH
	fmt.Printf("%s Connecting to %s...\n", cyan("→"), cfg.SSH.Host)
	client := ssh.NewClient(cfg.SSH.Host, cfg.SSH.Port, cfg.SSH.User, cfg.SSH.Key)
	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to connect: %v\n", red("✗"), err)
		os.Exit(1)
	}
	defer client.Close()

	mgr := swarm.NewManager(client, cfg.Stack)

	// Check if stack exists
	exists, err := mgr.StackExists()
	if err != nil || !exists {
		fmt.Fprintf(os.Stderr, "%s Stack %s not found\n", red("✗"), cfg.Stack)
		os.Exit(1)
	}

	// Get services to rollback
	var servicesToRollback []string

	if targetService != "" {
		servicesToRollback = []string{targetService}
	} else {
		// Get all services
		services, err := mgr.ListServices()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Failed to list services: %v\n", red("✗"), err)
			os.Exit(1)
		}

		for _, svc := range services {
			// Extract service name without stack prefix
			name := svc.Name
			if len(cfg.Stack) > 0 && len(name) > len(cfg.Stack)+1 {
				name = name[len(cfg.Stack)+1:]
			}
			servicesToRollback = append(servicesToRollback, name)
		}
	}

	if len(servicesToRollback) == 0 {
		fmt.Printf("%s No services to rollback\n", yellow("!"))
		return
	}

	// Rollback each service
	fmt.Printf("%s Rolling back %d service(s)...\n", cyan("→"), len(servicesToRollback))

	for _, svc := range servicesToRollback {
		fmt.Printf("  %s %s...", cyan("→"), svc)
		if err := mgr.RollbackService(svc); err != nil {
			fmt.Printf(" %s (%v)\n", red("✗"), err)
			continue
		}
		fmt.Printf(" %s\n", green("✓"))
	}

	// Show status
	fmt.Printf("\n%s Services after rollback:\n", cyan("→"))
	services, err := mgr.ListServices()
	if err == nil {
		for _, svc := range services {
			fmt.Printf("  %-30s %s\n", svc.Name, svc.Replicas)
		}
	}

	fmt.Printf("\n%s Rollback completed\n", green("✓"))
}
