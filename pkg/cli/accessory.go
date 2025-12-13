package cli

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/marcelsud/swarmctl/internal/accessories"
	"github.com/marcelsud/swarmctl/internal/config"
	"github.com/marcelsud/swarmctl/internal/ssh"
	"github.com/spf13/cobra"
)

var accessoryCmd = &cobra.Command{
	Use:   "accessory",
	Short: "Manage accessory services",
	Long: `Manage accessory services (Redis, PostgreSQL, etc.) for the stack.
Accessories are defined in swarm.yaml under the 'accessories' key.`,
	Run: runAccessoryList,
}

var accessoryStartCmd = &cobra.Command{
	Use:   "start <name|all>",
	Short: "Start an accessory service",
	Args:  cobra.ExactArgs(1),
	Run:   runAccessoryStart,
}

var accessoryStopCmd = &cobra.Command{
	Use:   "stop <name|all>",
	Short: "Stop an accessory service",
	Args:  cobra.ExactArgs(1),
	Run:   runAccessoryStop,
}

var accessoryRestartCmd = &cobra.Command{
	Use:   "restart <name|all>",
	Short: "Restart an accessory service",
	Args:  cobra.ExactArgs(1),
	Run:   runAccessoryRestart,
}

func init() {
	accessoryCmd.AddCommand(accessoryStartCmd)
	accessoryCmd.AddCommand(accessoryStopCmd)
	accessoryCmd.AddCommand(accessoryRestartCmd)
}

func runAccessoryList(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	cfg, err := config.Load(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", red("✗"), err)
		os.Exit(1)
	}

	if len(cfg.Accessories) == 0 {
		fmt.Printf("%s No accessories defined in swarm.yaml\n", yellow("!"))
		return
	}

	client := ssh.NewClient(cfg.SSH.Host, cfg.SSH.Port, cfg.SSH.User, cfg.SSH.Key)
	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to connect: %v\n", red("✗"), err)
		os.Exit(1)
	}
	defer client.Close()

	mgr := accessories.NewManager(client, cfg.Stack)

	statuses, _ := mgr.ListAll(cfg.Accessories)

	fmt.Printf("%s Accessories for stack %s:\n\n", cyan("→"), cfg.Stack)
	fmt.Printf("  %-20s %-15s %s\n", "NAME", "REPLICAS", "STATUS")
	for _, status := range statuses {
		statusStr := red("stopped")
		if status.Running {
			statusStr = green("running")
		}
		fmt.Printf("  %-20s %-15s %s\n", status.Name, status.Replicas, statusStr)
	}
}

func runAccessoryStart(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	target := args[0]

	cfg, err := config.Load(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", red("✗"), err)
		os.Exit(1)
	}

	client := ssh.NewClient(cfg.SSH.Host, cfg.SSH.Port, cfg.SSH.User, cfg.SSH.Key)
	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to connect: %v\n", red("✗"), err)
		os.Exit(1)
	}
	defer client.Close()

	mgr := accessories.NewManager(client, cfg.Stack)

	targets := getTargets(target, cfg.Accessories)

	for _, name := range targets {
		fmt.Printf("%s Starting %s...", cyan("→"), name)
		if err := mgr.Start(name); err != nil {
			fmt.Printf(" %s (%v)\n", red("✗"), err)
			continue
		}
		fmt.Printf(" %s\n", green("✓"))
	}
}

func runAccessoryStop(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	target := args[0]

	cfg, err := config.Load(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", red("✗"), err)
		os.Exit(1)
	}

	client := ssh.NewClient(cfg.SSH.Host, cfg.SSH.Port, cfg.SSH.User, cfg.SSH.Key)
	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to connect: %v\n", red("✗"), err)
		os.Exit(1)
	}
	defer client.Close()

	mgr := accessories.NewManager(client, cfg.Stack)

	targets := getTargets(target, cfg.Accessories)

	for _, name := range targets {
		fmt.Printf("%s Stopping %s...", cyan("→"), name)
		if err := mgr.Stop(name); err != nil {
			fmt.Printf(" %s (%v)\n", red("✗"), err)
			continue
		}
		fmt.Printf(" %s\n", green("✓"))
	}
}

func runAccessoryRestart(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	target := args[0]

	cfg, err := config.Load(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", red("✗"), err)
		os.Exit(1)
	}

	client := ssh.NewClient(cfg.SSH.Host, cfg.SSH.Port, cfg.SSH.User, cfg.SSH.Key)
	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to connect: %v\n", red("✗"), err)
		os.Exit(1)
	}
	defer client.Close()

	mgr := accessories.NewManager(client, cfg.Stack)

	targets := getTargets(target, cfg.Accessories)

	for _, name := range targets {
		fmt.Printf("%s Restarting %s...", cyan("→"), name)
		if err := mgr.Restart(name); err != nil {
			fmt.Printf(" %s (%v)\n", red("✗"), err)
			continue
		}
		fmt.Printf(" %s\n", green("✓"))
	}
}

func getTargets(target string, accessories []string) []string {
	if target == "all" {
		return accessories
	}
	return []string{target}
}
