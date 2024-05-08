package routerkit

import "ecst-payment/pkg/tracer"

type routerConfig struct {
	serviceName string
	spanOpts    []tracer.Option // additional span options to be applied
}

// RouterOption represents an option that can be passed to NewRouter.
type RouterOption func(*routerConfig)

func defaults(cfg *routerConfig) {
	cfg.serviceName = "mux.router"
}

// WithServiceName sets the given service name for the router.
func WithServiceName(name string) RouterOption {
	return func(cfg *routerConfig) {
		cfg.serviceName = name
	}
}
