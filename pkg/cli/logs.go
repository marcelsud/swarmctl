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

	// Connect via SSH
	client := ssh.NewClient(cfg.SSH.Host, cfg.SSH.Port, cfg.SSH.User, cfg.SSH.Key)
	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to connect: %v\n", red("✗"), err)
		os.Exit(1)
	}
	defer client.Close()

	mgr := swarm.NewManager(client, cfg.Stack)

	// If no service specified, list available services
	if len(args) == 0 {
		services, err := mgr.ListServices()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Failed to list services: %v\n", red("✗"), err)
			os.Exit(1)
		}

		fmt.Printf("%s Available services:\n", cyan("→"))
		for _, svc := range services {
			// Extract service name (remove stack prefix)
			name := svc.Name
			if len(cfg.Stack) > 0 && len(name) > len(cfg.Stack)+1 {
				name = name[len(cfg.Stack)+1:]
			}
			fmt.Printf("  - %s\n", name)
		}
		fmt.Printf("\nUsage: swarmctl logs <service>\n")
		return
	}

	serviceName := args[0]

	// For follow mode, stream logs interactively
	if logsFollow {
		fmt.Printf("%s Streaming logs for %s (Ctrl+C to stop)...\n\n", cyan("→"), serviceName)
		runStreamLogs(client, cfg.Stack, serviceName)
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

func runStreamLogs(client *ssh.Client, stackName, serviceName string) {
	fullName := fmt.Sprintf("%s_%s", stackName, serviceName)
	cmd := fmt.Sprintf("docker service logs %s --follow --tail %d", fullName, logsTail)
	if logsSince != "" {
		cmd += fmt.Sprintf(" --since %s", logsSince)
	}

	// Run interactively to stream output
	if err := client.RunStream(cmd, os.Stdout, os.Stderr); err != nil {
		// Ignore error on Ctrl+C
		return
	}
}
