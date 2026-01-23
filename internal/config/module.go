package config

import (
	"github.com/go-core-fx/badgerfx"
	"github.com/go-core-fx/fiberfx"
	"github.com/go-core-fx/fiberfx/openapi"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"config",
		fx.Provide(New),
		fx.Provide(func(cfg Config) fiberfx.Config {
			return fiberfx.Config{
				Address:     cfg.HTTP.Address,
				ProxyHeader: cfg.HTTP.ProxyHeader,
				Proxies:     cfg.HTTP.Proxies,
			}
		}),
		fx.Provide(func(cfg Config) badgerfx.Config {
			return badgerfx.Config{
				Dir: cfg.Storage.Dir,
			}
		}),
		fx.Provide(func(cfg Config) openapi.Config {
			return openapi.Config{
				Enabled:    cfg.HTTP.OpenAPI.Enabled,
				PublicHost: cfg.HTTP.OpenAPI.PublicHost,
				PublicPath: cfg.HTTP.OpenAPI.PublicPath,
			}
		}),
	)
}
