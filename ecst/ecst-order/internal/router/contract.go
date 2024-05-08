// Package router
package router

import (
	"ecst-order/internal/appctx"
	"ecst-order/internal/ucase/contract"
	"ecst-order/pkg/routerkit"
	"net/http"
)

// httpHandlerFunc is a contract http handler for router
type httpHandlerFunc func(request *http.Request, svc contract.UseCase, conf *appctx.Config) appctx.Response

// Router is a contract router and must implement this interface
type Router interface {
	Route() *routerkit.Router
}
