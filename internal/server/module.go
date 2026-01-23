package server

import (
	"github.com/go-core-fx/fiberfx"
	"github.com/go-core-fx/fiberfx/health"
	"github.com/go-core-fx/fiberfx/statuscode"
	"github.com/go-core-fx/logger"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Module() fx.Option {
	return fx.Module(
		"server",
		logger.WithNamedLogger("server"),

		fx.Provide(func(log *zap.Logger) fiberfx.Options {
			opts := fiberfx.Options{}
			opts.WithErrorHandler(fiberfx.NewJSONErrorHandler(log))
			opts.WithMetrics()
			return opts
		}),

		fx.Provide(health.NewHandler, fx.Private),

		fx.Invoke(func(app *fiber.App, health *health.Handler) {
			health.Register(app)

			api := app.Group("/api/v1")

			// messages.Register(api.Group("/messages"))

			api.Use(statuscode.New())
		}),
	)
}
