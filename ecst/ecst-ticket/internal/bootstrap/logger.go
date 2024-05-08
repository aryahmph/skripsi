package bootstrap

import (
	"ecst-ticket/internal/appctx"
	"ecst-ticket/pkg/logger"
	"ecst-ticket/pkg/util"
)

func RegistryLogger(cfg *appctx.Config) {
	lc := logger.Config{
		Environment: util.EnvironmentTransform(cfg.App.Env),
		Debug:       cfg.App.Debug,
	}

	logger.Setup(lc, nil)
}
