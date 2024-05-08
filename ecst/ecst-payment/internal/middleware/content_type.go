package middleware

import (
	"ecst-payment/internal/appctx"
	"ecst-payment/internal/consts"
	"ecst-payment/pkg/logger"
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
