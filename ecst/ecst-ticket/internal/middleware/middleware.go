// Package middleware
package middleware

import (
	"ecst-ticket/internal/appctx"
	"ecst-ticket/internal/consts"
	"net/http"
)

// MiddlewareFunc is contract for middleware and must implement this type for http if need middleware http request
type MiddlewareFunc func(r *http.Request, conf *appctx.Config) string

// FilterFunc is a iterator resolver in each middleware registered
func FilterFunc(conf *appctx.Config, r *http.Request, mfs []MiddlewareFunc) string {
	for _, mf := range mfs {
		if status := mf(r, conf); status != consts.MiddlewarePassed {
			return status
		}
	}

	return consts.MiddlewarePassed
}
