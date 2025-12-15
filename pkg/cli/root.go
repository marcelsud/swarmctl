package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	configFile  string
	destination string
	verbose     bool
)

var rootCmd = &cobra.Command{
	Use:   "swarmctl",
	Short: "Deploy and manage Docker Swarm stacks",
	Long: `swarmctl is a CLI tool for deploying and managing Docker Swarm stacks.
Inspired by Kamal, it provides a simple and intuitive interface for
managing your Swarm deployments via SSH.

Multiple environments are supported via the -d flag:
  swarmctl deploy                  # Uses swarm.yaml
  swarmctl deploy -d staging       # Uses swarm.staging.yaml
  swarmctl deploy -d production    # Uses swarm.production.yaml`,
	Version: "0.1.0",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// If destination is set and config is default, use swarm.{destination}.yaml
		if destination != "" && configFile == "swarm.yaml" {
			configFile = fmt.Sprintf("swarm.%s.yaml", destination)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "swarm.yaml", "config file path")
	rootCmd.PersistentFlags().StringVarP(&destination, "destination", "d", "", "deployment destination (e.g., staging, production)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Add subcommands
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(rollbackCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(secretsCmd)
	rootCmd.AddCommand(accessoryCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(docsCmd)
	rootCmd.AddCommand(initLLMCmd)
}

func Execute() error {
	return rootCmd.Execute()
}

func exitWithError(msg string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s: %v\n", msg, err)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	}
	os.Exit(1)
}
