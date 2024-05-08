package middleware

import (
	"en-payment/internal/appctx"
	"en-payment/internal/consts"
	"en-payment/pkg/logger"
	"net/http"
)

func ValidateClientName(r *http.Request, _ *appctx.Config) string {
	ctx := r.Context()
	clientName := r.Header.Get(consts.HeaderXClientName)
	lf := []logger.Field{
		logger.EventName(consts.LogEventNameMiddlewareValidateClientName),
		logger.Any(consts.LogFieldClientName, clientName),
	}

	if clientName == "" {
		logger.WarnWithContext(ctx, logger.SetMessageFormat("%s header not found", consts.HeaderXClientName), lf...)
		return consts.ResponseAuthenticationFailure
	}

	return consts.MiddlewarePassed
}
