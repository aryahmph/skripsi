package bootstrap

import (
	"ecst-order-expiration/internal/appctx"
	"ecst-order-expiration/pkg/logger"
	"ecst-order-expiration/pkg/util"
)

func RegistryLogger(cfg *appctx.Config) {
	lc := logger.Config{
		Environment: util.EnvironmentTransform(cfg.App.Env),
		Debug:       cfg.App.Debug,
	}

	logger.Setup(lc, nil)
}
