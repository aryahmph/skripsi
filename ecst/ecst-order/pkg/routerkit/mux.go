package routerkit

import (
	"ecst-order/pkg/tracer"
	"github.com/gorilla/mux"
	"net/http"
)

// Router registers routes to be matched and dispatches a handler.
type Router struct {
	*mux.Router
	config *routerConfig
}

// NewRouter returns a new router instance traced with the global tracer.
func NewRouter(opts ...RouterOption) *Router {
	cfg := new(routerConfig)
	defaults(cfg)
	for _, fn := range opts {
		fn(cfg)
	}
	return &Router{
		Router: mux.NewRouter(),
		config: cfg,
	}
}

// ServeHTTP dispatches the request to the handler
// whose pattern most closely matches the request URL.
// We only need to rewrite this function to be able to trace
// all the incoming requests to the underlying multiplexer
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var (
		match    mux.RouteMatch
		spanopts []tracer.Option
		route    = "unknown"
	)
	// get the resource associated to this request
	if r.Match(req, &match) && match.Route != nil {
		if r, err := match.Route.GetPathTemplate(); err == nil {
			route = r
		}
		if h, err := match.Route.GetHostTemplate(); err == nil {
			spanopts = append(spanopts, tracer.Option{TagKey: "mux.host", Value: h})
		}
	}
	spanopts = append(spanopts, r.config.spanOpts...)
	resource := req.Method + " " + route
	TraceAndServe(r.Router, w, req, r.config.serviceName, resource, spanopts...)
}
