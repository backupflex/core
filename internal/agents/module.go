package agents

import (
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

// Module configures dependency graph for the agent coordination module.
func Module() fx.Option {
	return fx.Module(
		"agents",
		logger.WithNamedLogger("agents"),
		fx.Provide(NewRepository, fx.Private),
		fx.Provide(NewStatusTracker, fx.Private),
		fx.Provide(New),
	)
}
