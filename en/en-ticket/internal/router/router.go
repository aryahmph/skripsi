package router

import (
	"context"
	"en-ticket/internal/appctx"
	"en-ticket/internal/bootstrap"
	"en-ticket/internal/consts"
	"en-ticket/internal/handler"
	"en-ticket/internal/middleware"
	"en-ticket/internal/repositories"
	"en-ticket/internal/ucase"
	ucaseContract "en-ticket/internal/ucase/contract"
	"en-ticket/internal/ucase/v1/ticket"
	"en-ticket/pkg/cache"
	"en-ticket/pkg/logger"
	"en-ticket/pkg/msg"
	"en-ticket/pkg/postgres"
	"en-ticket/pkg/routerkit"
	"encoding/json"
	"net/http"
	"runtime/debug"
	"time"
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

				res.SetMessage()
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

	// main route
	root := rtr.router.PathPrefix("/").Subrouter()

	in := root.PathPrefix("/in/").Subrouter()
	ex := root.PathPrefix("/ex/").Subrouter()

	exV1 := ex.PathPrefix("/v1/").Subrouter()
	inV1 := in.PathPrefix("/v1/").Subrouter()

	// repositories
	ticketRepo := repositories.NewTicketRepository(db)

	// use case
	healthy := ucase.NewHealthCheck()
	internalGetTicket := ticket.NewInternalGetTicket(ticketRepo, cacher)
	listTicketCategories := ticket.NewListCategories(cacher, ticketRepo)
	listUnreservedTickets := ticket.NewListUnreserved(cacher, ticketRepo)

	// route
	in.HandleFunc("/health", rtr.handle(
		handler.HttpRequest,
		healthy,
	)).Methods(http.MethodGet)

	inV1.HandleFunc("/tickets/{id}", rtr.handle(
		handler.HttpRequest,
		internalGetTicket,
		middleware.ValidateClientName,
	)).Methods(http.MethodGet)

	exV1.HandleFunc("/ticket-groups/{group_id}/categories", rtr.handle(
		handler.HttpRequest,
		listTicketCategories,
		middleware.ValidateUserIDAndEmail,
	)).Methods(http.MethodGet)

	exV1.HandleFunc("/ticket-groups/{group_id}", rtr.handle(
		handler.HttpRequest,
		listUnreservedTickets,
		middleware.ValidateUserIDAndEmail,
	)).Methods(http.MethodGet)

	return rtr.router

}

func (rtr *router) defaultLang(l string) string {
	if len(l) == 0 {
		return rtr.config.App.DefaultLang
	}

	return l
}
