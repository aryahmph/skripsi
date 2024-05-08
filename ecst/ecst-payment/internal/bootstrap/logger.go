package bootstrap

import (
	"ecst-payment/internal/appctx"
	"ecst-payment/pkg/logger"
	"ecst-payment/pkg/util"
)

func RegistryLogger(cfg *appctx.Config) {
	lc := logger.Config{
		Environment: util.EnvironmentTransform(cfg.App.Env),
		Debug:       cfg.App.Debug,
	}

	logger.Setup(lc, nil)
}
