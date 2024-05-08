// Package router
package router

import (
	"en-ticket/internal/appctx"
	"en-ticket/internal/ucase/contract"
	"en-ticket/pkg/routerkit"
	"net/http"
)

// httpHandlerFunc is a contract http handler for router
type httpHandlerFunc func(request *http.Request, svc contract.UseCase, conf *appctx.Config) appctx.Response

// Router is a contract router and must implement this interface
type Router interface {
	Route() *routerkit.Router
}
