// Package router
package router

import (
	"en-payment/internal/appctx"
	"en-payment/internal/ucase/contract"
	"en-payment/pkg/routerkit"
	"net/http"
)

// httpHandlerFunc is a contract http handler for router
type httpHandlerFunc func(request *http.Request, svc contract.UseCase, conf *appctx.Config) appctx.Response

// Router is a contract router and must implement this interface
type Router interface {
	Route() *routerkit.Router
}
