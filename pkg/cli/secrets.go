package cli

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/marcelsud/swarmctl/internal/config"
	"github.com/marcelsud/swarmctl/internal/executor"
	"github.com/marcelsud/swarmctl/internal/secrets"
	"github.com/spf13/cobra"
)

var secretsEnvFile string

var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Manage Swarm secrets",
	Long:  `Manage Docker Swarm secrets for the stack.`,
}

var secretsPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push secrets from .env to Swarm",
	Long: `Push secrets defined in swarm.yaml from .env file or environment variables.
Secrets are created with the format: {stack}_{secret_name}`,
	Run: runSecretsPush,
}

var secretsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List existing secrets",
	Run:   runSecretsList,
}

func init() {
	secretsPushCmd.Flags().StringVarP(&secretsEnvFile, "env-file", "e", ".env", "path to .env file")

	secretsCmd.AddCommand(secretsPushCmd)
	secretsCmd.AddCommand(secretsListCmd)
}

func runSecretsPush(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	// Load config
	cfg, err := config.Load(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", red("✗"), err)
		os.Exit(1)
	}

	if len(cfg.Secrets) == 0 {
		fmt.Printf("%s No secrets defined in swarm.yaml\n", yellow("!"))
		return
	}

	fmt.Printf("%s Loading secrets for: %v\n", cyan("→"), cfg.Secrets)

	// Load secrets from .env file or environment
	var secretList []secrets.Secret
	if _, err := os.Stat(secretsEnvFile); err == nil {
		secretList, err = secrets.LoadFromEnvFile(secretsEnvFile, cfg.Secrets)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s %v\n", red("✗"), err)
			os.Exit(1)
		}
		fmt.Printf("  Loaded from %s\n", secretsEnvFile)
	} else {
		secretList = secrets.LoadFromEnv(cfg.Secrets)
		fmt.Printf("  Loaded from environment variables\n")
	}

	if len(secretList) == 0 {
		fmt.Printf("%s No secrets found. Make sure they are defined in .env or environment.\n", yellow("!"))
		return
	}

	// Create executor
	exec, err := executor.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to connect: %v\n", red("✗"), err)
		os.Exit(1)
	}
	defer exec.Close()

	mgr := secrets.NewManager(exec, cfg.Stack)

	// Push secrets
	fmt.Printf("%s Pushing %d secret(s)...\n", cyan("→"), len(secretList))

	for _, secret := range secretList {
		fmt.Printf("  %s %s...", cyan("→"), secret.Name)
		if err := mgr.Create(secret.Name, secret.Value); err != nil {
			fmt.Printf(" %s (%v)\n", red("✗"), err)
			continue
		}
		fmt.Printf(" %s\n", green("✓"))
	}

	fmt.Printf("\n%s Secrets pushed successfully\n", green("✓"))
}

func runSecretsList(cmd *cobra.Command, args []string) {
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

	mgr := secrets.NewManager(exec, cfg.Stack)

	// List secrets
	secretsList, err := mgr.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s Failed to list secrets: %v\n", red("✗"), err)
		os.Exit(1)
	}

	if len(secretsList) == 0 {
		fmt.Printf("%s No secrets found for stack %s\n", cyan("→"), cfg.Stack)
		return
	}

	fmt.Printf("%s Secrets for stack %s:\n", cyan("→"), cfg.Stack)
	for _, secret := range secretsList {
		fmt.Printf("  - %s\n", secret)
	}
}
