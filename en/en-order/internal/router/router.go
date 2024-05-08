package router

import (
	"context"
	"en-order/pkg/cache"
	"en-order/pkg/dq"
	"en-order/pkg/postgres"
	"encoding/json"
	"net/http"
	"runtime/debug"
	"time"

	"en-order/internal/appctx"
	"en-order/internal/bootstrap"
	"en-order/internal/consts"
	"en-order/internal/handler"
	"en-order/internal/middleware"
	"en-order/internal/providers/ticket"
	repository "en-order/internal/repositories"
	"en-order/internal/ucase"
	ucaseContract "en-order/internal/ucase/contract"
	"en-order/internal/ucase/v1/order"

	"en-order/pkg/logger"
	"en-order/pkg/msg"
	"en-order/pkg/routerkit"
)

type router struct {
	config *appctx.Config
	router *routerkit.Router
}

// NewRouter initialize new router wil return Router Interface
func NewRouter(cfg *appctx.Config) Router {
	bootstrap.RegistryMessage()
	bootstrap.RegistryLogger(cfg)

	// bootstrapping OTel
	shutdown := bootstrap.RegistryOpenTelemetry(cfg)
	defer shutdown()

	return &router{
		config: cfg,
		router: routerkit.NewRouter(routerkit.WithServiceName(cfg.App.AppName)),
	}
}

func (rtr *router) handle(hfn httpHandlerFunc, svc ucaseContract.UseCase, mdws ...middleware.MiddlewareFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				w.Header().Set(consts.HeaderContentTypeKey, consts.HeaderContentTypeJSON)
				w.WriteHeader(http.StatusInternalServerError)
				res := appctx.Response{
					Name: consts.ResponseInternalFailure,
				}

				logger.Error(logger.SetMessageFormat("error %v", string(debug.Stack())))
				json.NewEncoder(w).Encode(res)
				return
			}
		}()

		var st time.Time
		var lt time.Duration

		st = time.Now()

		ctx := context.WithValue(r.Context(), "access", map[string]interface{}{
			"path":      r.URL.Path,
			"remote_ip": r.RemoteAddr,
			"method":    r.Method,
		})

		req := r.WithContext(ctx)

		if status := middleware.FilterFunc(rtr.config, req, mdws); status != consts.MiddlewarePassed {
			rtr.response(w, appctx.Response{
				Name: status,
			})

			return
		}

		resp := hfn(req, svc, rtr.config)

		resp.Lang = rtr.defaultLang(req.Header.Get(consts.HeaderLanguageKey))

		rtr.response(w, resp)

		lt = time.Since(st)
		logger.AccessLog("access log",
			logger.Any("tag", "go-access"),
			logger.Any("http.path", req.URL.Path),
			logger.Any("http.method", req.Method),
			logger.Any("http.agent", req.UserAgent()),
			logger.Any("http.referer", req.Referer()),
			logger.Any("http.status", resp.GetCode()),
			logger.Any("http.latency", lt.Seconds()),
		)
	}
}

// response prints as a json and formatted string for DGP legacy
func (rtr *router) response(w http.ResponseWriter, resp appctx.Response) {
	w.Header().Set(consts.HeaderContentTypeKey, consts.HeaderContentTypeJSON)

	if !msg.GetAvailableLang(consts.ResponseSuccess, resp.Lang) {
		resp.Lang = rtr.config.App.DefaultLang
	}

	defer func() {
		resp.BuildMessage()
		w.WriteHeader(resp.GetCode())
		json.NewEncoder(w).Encode(resp)
	}()

	return

}

// Route preparing http router and will return mux router object
func (rtr *router) Route() *routerkit.Router {
	// bootstrap
	var db postgres.Adapter
	if rtr.config.App.IsSingle {
		db = bootstrap.RegistryPostgresDBSingle(rtr.config.DBWrite, rtr.config.App.Timezone)
	} else {
		db = bootstrap.RegistryPostgresDB(rtr.config.DBWrite, rtr.config.DBRead, rtr.config.App.Timezone)
	}
	cacher := cache.NewCache(bootstrap.RegistryRedisNative(rtr.config))
	kp := bootstrap.RegistryKafkaProducer(rtr.config)
	dqp := dq.NewProducer(cacher)

	// main route
	root := rtr.router.PathPrefix("/").Subrouter()

	in := root.PathPrefix("/in/").Subrouter()
	ex := root.PathPrefix("/ex/").Subrouter()

	exV1 := ex.PathPrefix("/v1/").Subrouter()
	inV1 := in.PathPrefix("/v1/").Subrouter()

	// repositories
	orderRepo := repository.NewOrderRepository(db)

	// use case
	healthy := ucase.NewHealthCheck()
	ticketProvider := ticket.NewTicketProvider(rtr.config)
	createOrder := order.NewCreateOrder(kp, dqp, orderRepo, ticketProvider)
	internalGetOrder := order.NewInternalGetOrder(cacher, orderRepo)

	// route
	in.HandleFunc("/health", rtr.handle(
		handler.HttpRequest,
		healthy,
	)).Methods(http.MethodGet)

	exV1.HandleFunc("/orders", rtr.handle(
		handler.HttpRequest,
		createOrder,
		middleware.ValidateUserIDAndEmail,
	)).Methods(http.MethodPost)

	inV1.HandleFunc("/orders/{id}", rtr.handle(
		handler.HttpRequest,
		internalGetOrder,
		middleware.ValidateClientName,
	)).Methods(http.MethodGet)

	return rtr.router

}

func (rtr *router) defaultLang(l string) string {
	if len(l) == 0 {
		return rtr.config.App.DefaultLang
	}

	return l
}
