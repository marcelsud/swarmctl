package cli

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/marcelsud/swarmctl/docs"
	"github.com/spf13/cobra"
)

var docsCmd = &cobra.Command{
	Use:   "docs [topic]",
	Short: "Show documentation",
	Long:  `Show swarmctl documentation. Run without arguments to list available topics.`,
	Args:  cobra.MaximumNArgs(1),
	Run:   runDocs,
}

func runDocs(cmd *cobra.Command, args []string) {
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	if len(args) == 0 {
		listTopics(cyan, yellow)
		return
	}

	topic := args[0]
	content, err := docs.Get(topic)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s Unknown topic: %s\n", red("✗"), topic)
		fmt.Fprintf(os.Stderr, "\nRun %s to see available topics.\n", cyan("swarmctl docs"))
		os.Exit(1)
	}

	fmt.Print(content)
}

func listTopics(cyan, yellow func(a ...interface{}) string) {
	fmt.Printf("%s Available documentation topics:\n\n", cyan("→"))

	for _, t := range docs.Topics {
		fmt.Printf("  %s  %s\n", yellow(fmt.Sprintf("%-18s", t.Name)), t.Description)
	}

	fmt.Printf("\nUsage: %s\n", cyan("swarmctl docs <topic>"))
}
