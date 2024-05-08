package http

import (
	"context"
	"ecst-ticket/internal/consts"
	"ecst-ticket/internal/server"
	"ecst-ticket/pkg/logger"
)

func Start(ctx context.Context) {
	serve := server.NewHTTPServer()
	defer serve.Done()
	logger.Info(logger.SetMessageFormat("starting rll-account-service services... %d", serve.Config().App.Port), logger.EventName(consts.LogEventNameServiceStarting))

	if err := serve.Run(ctx); err != nil {
		logger.Warn(logger.SetMessageFormat("service stopped, err:%s", err.Error()), logger.EventName(consts.LogEventNameServiceStarting))
	}

	return
}
