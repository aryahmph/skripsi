package middleware

import (
	"ecst-ticket/internal/appctx"
	"ecst-ticket/internal/consts"
	"ecst-ticket/pkg/logger"
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
