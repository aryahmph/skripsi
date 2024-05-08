package bootstrap

import (
	"en-ticket/internal/appctx"
	"en-ticket/pkg/logger"
	"en-ticket/pkg/util"
)

func RegistryLogger(cfg *appctx.Config) {
	lc := logger.Config{
		Environment: util.EnvironmentTransform(cfg.App.Env),
		Debug:       cfg.App.Debug,
	}

	logger.Setup(lc, nil)
}
