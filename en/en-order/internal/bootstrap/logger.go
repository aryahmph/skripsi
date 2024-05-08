package bootstrap

import (
	"en-order/internal/appctx"
	"en-order/pkg/logger"
	"en-order/pkg/util"
)

func RegistryLogger(cfg *appctx.Config) {
	lc := logger.Config{
		Environment: util.EnvironmentTransform(cfg.App.Env),
		Debug:       cfg.App.Debug,
	}

	logger.Setup(lc, nil)
}
