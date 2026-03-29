package internal

import (
	"context"

	"github.com/backupflex/core/internal/agents"
	"github.com/backupflex/core/internal/config"
	"github.com/backupflex/core/internal/server"
	"github.com/go-core-fx/badgerfx"
	"github.com/go-core-fx/fiberfx"
	"github.com/go-core-fx/healthfx"
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Run(version healthfx.Version) {
	fx.New(
		// CORE MODULES
		logger.Module(),
		logger.WithFxDefaultLogger(),
		badgerfx.Module(),
		// goosefx.Module(),
		// bunfx.Module(),
		fiberfx.Module(),
		healthfx.Module(),
		//
		// APP MODULES
		config.Module(),
		// db.Module(),
		server.Module(),
		// bot.Module(),
		//
		// BUSINESS MODULES
		agents.Module(),
		fx.Supply(version),
		//
		fx.Invoke(func(lc fx.Lifecycle, logger *zap.Logger) {
			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					logger.Info("app started")
					return nil
				},
				OnStop: func(_ context.Context) error {
					logger.Info("app stopped")
					return nil
				},
			})
		}),
	).Run()
}
