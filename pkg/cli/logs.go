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

var (
	logsFollow bool
	logsSince  string
	logsTail   int
)

var logsCmd = &cobra.Command{
	Use:   "logs [service]",
	Short: "View aggregated logs",
	Long: `View aggregated logs from services.
If no service is specified, shows logs from all services.`,
	Args: cobra.MaximumNArgs(1),
	Run:  runLogs,
}

func init() {
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "follow log output")
	logsCmd.Flags().StringVar(&logsSince, "since", "", "show logs since timestamp (e.g., 1h, 30m)")
	logsCmd.Flags().IntVarP(&logsTail, "tail", "n", 100, "number of lines to show")
}

func runLogs(cmd *cobra.Command, args []string) {
	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

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

	// If no service specified, show logs from all services
	if len(args) == 0 {
		services, err := mgr.ListServices()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Failed to list services: %v\n", red("✗"), err)
			os.Exit(1)
		}

		if len(services) == 0 {
			fmt.Printf("%s No services running\n", cyan("→"))
			return
		}

		fmt.Printf("%s Showing logs from all services:\n\n", cyan("→"))
		for _, svc := range services {
			// Extract service name (remove stack/project prefix)
			name := svc.Name
			if len(cfg.Stack) > 0 && len(name) > len(cfg.Stack)+1 {
				name = name[len(cfg.Stack)+1:]
			}

			fmt.Printf("%s=== Logs for %s ===\n\n", cyan("→"), name)

			// Get logs for this service
			logs, err := mgr.GetServiceLogs(name, false, logsSince, logsTail)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Failed to get logs for %s: %v\n", name, err)
				continue
			}

			fmt.Print(logs)
			if logs != "" {
				fmt.Println()
			}
		}
		return
	}

	serviceName := args[0]

	// For follow mode, stream logs interactively
	if logsFollow {
		fmt.Printf("%s Streaming logs for %s (Ctrl+C to stop)...\n\n", cyan("→"), serviceName)
		if err := mgr.StreamServiceLogs(serviceName, true, logsTail, os.Stdout, os.Stderr); err != nil {
			// Ignore error on Ctrl+C
			return
		}
		return
	}

	// Get logs
	logs, err := mgr.GetServiceLogs(serviceName, false, logsSince, logsTail)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to get logs: %v\n", red("✗"), err)
		os.Exit(1)
	}

	fmt.Print(logs)
}
