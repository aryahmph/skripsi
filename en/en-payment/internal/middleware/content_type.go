package middleware

import (
	"en-payment/internal/appctx"
	"en-payment/internal/consts"
	"en-payment/pkg/logger"
	"fmt"
	"net/http"
	"strings"
)

// ValidateContentType header
func ValidateContentType(r *http.Request, conf *appctx.Config) string {
	if ct := strings.ToLower(r.Header.Get(`Content-Type`)); ct != `application/json` {
		logger.Warn(fmt.Sprintf("[middleware] invalid content-type %s", ct))

		return consts.ResponseUnprocessableEntity
	}

	return consts.MiddlewarePassed
}
