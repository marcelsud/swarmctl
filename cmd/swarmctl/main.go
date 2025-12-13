package main

import (
	"os"

	"github.com/marcelsud/swarmctl/pkg/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
