package bootstrap

import (
	"en-payment/internal/appctx"
	"en-payment/pkg/logger"
	"en-payment/pkg/util"
)

func RegistryLogger(cfg *appctx.Config) {
	lc := logger.Config{
		Environment: util.EnvironmentTransform(cfg.App.Env),
		Debug:       cfg.App.Debug,
	}

	logger.Setup(lc, nil)
}
