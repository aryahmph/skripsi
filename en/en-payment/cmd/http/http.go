package http

import (
	"context"
	"en-payment/internal/consts"
	"en-payment/internal/server"
	"en-payment/pkg/logger"
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
