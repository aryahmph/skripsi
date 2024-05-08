package middleware

import (
	"ecst-order/internal/appctx"
	"ecst-order/internal/consts"
	"ecst-order/pkg/logger"
	"net/http"
)

func ValidateUserIDAndEmail(r *http.Request, _ *appctx.Config) string {
	ctx := r.Context()
	userId := r.Header.Get(consts.HeaderXUserId)
	userEmail := r.Header.Get(consts.HeaderXUserEmail)

	lf := []logger.Field{
		logger.EventName(consts.LogEventNameMiddlewareValidateUserIDAndEmail),
		logger.Any(consts.HeaderXUserId, userId),
		logger.Any(consts.HeaderXUserEmail, userEmail),
	}

	if userId == "" {
		logger.WarnWithContext(ctx, logger.SetMessageFormat("%s header not found", consts.HeaderXUserId), lf...)
		return consts.ResponseAuthenticationFailure
	}

	if userEmail == "" {
		logger.WarnWithContext(ctx, logger.SetMessageFormat("%s header not found", consts.HeaderXUserEmail), lf...)
		return consts.ResponseAuthenticationFailure
	}

	return consts.MiddlewarePassed
}
