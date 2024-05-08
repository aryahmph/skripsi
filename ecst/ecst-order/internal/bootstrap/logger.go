package bootstrap

import (
	"ecst-order/internal/appctx"
	"ecst-order/pkg/logger"
	"ecst-order/pkg/util"
)

func RegistryLogger(cfg *appctx.Config) {
	lc := logger.Config{
		Environment: util.EnvironmentTransform(cfg.App.Env),
		Debug:       cfg.App.Debug,
	}

	logger.Setup(lc, nil)
}
