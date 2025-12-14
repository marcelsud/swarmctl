package deployment

import (
	"github.com/marcelsud/swarmctl/internal/config"
	"github.com/marcelsud/swarmctl/internal/executor"
)

// New creates the appropriate Manager based on configuration
func New(cfg *config.Config, exec executor.Executor) Manager {
	switch cfg.Mode {
	case config.ModeCompose:
		return NewComposeManager(exec, cfg.Stack)
	default:
		return NewSwarmManager(exec, cfg.Stack)
	}
}
