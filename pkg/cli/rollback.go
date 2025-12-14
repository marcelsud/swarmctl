package cli

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/marcelsud/swarmctl/internal/config"
	"github.com/marcelsud/swarmctl/internal/deployment"
	"github.com/marcelsud/swarmctl/internal/executor"
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
	bold := color.New(color.Bold).SprintFunc()

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

	// Create executor
	exec, err := executor.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to connect: %v\n", red("✗"), err)
		os.Exit(1)
	}
	defer exec.Close()

	// Create deployment manager
	mgr := deployment.New(cfg, exec)

	// Show mode info
	modeStr := mgr.GetMode()
	fmt.Printf("%s Stack: %s (%s mode)\n", cyan("→"), bold(cfg.Stack), modeStr)

	// Check if stack exists
	exists, err := mgr.Exists()
	if err != nil || !exists {
		fmt.Fprintf(os.Stderr, "%s Stack %s not found\n", red("✗"), cfg.Stack)
		os.Exit(1)
	}

	// Check if rollback is supported
	if !mgr.SupportsRollback() {
		fmt.Fprintf(os.Stderr, "%s Rollback is not supported in %s mode\n", red("✗"), modeStr)
		os.Exit(1)
	}

	// In compose mode, rollback is all-or-nothing (uses history)
	if cfg.Mode == config.ModeCompose {
		if targetService != "" {
			fmt.Printf("%s In compose mode, rollback affects all services (individual service rollback not supported)\n", yellow("!"))
		}

		fmt.Printf("%s Rolling back to previous deploy...\n", cyan("→"))
		if err := mgr.RollbackAll(); err != nil {
			fmt.Fprintf(os.Stderr, "%s Failed to rollback: %v\n", red("✗"), err)
			os.Exit(1)
		}
	} else {
		// Swarm mode: can rollback individual services
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
