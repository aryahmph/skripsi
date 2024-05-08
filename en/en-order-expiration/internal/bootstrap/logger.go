package bootstrap

import (
	"en-order-expiration/internal/appctx"
	"en-order-expiration/pkg/logger"
	"en-order-expiration/pkg/util"
)

func RegistryLogger(cfg *appctx.Config) {
	lc := logger.Config{
		Environment: util.EnvironmentTransform(cfg.App.Env),
		Debug:       cfg.App.Debug,
	}

	logger.Setup(lc, nil)
}
