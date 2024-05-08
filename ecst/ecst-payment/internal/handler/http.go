package handler

import (
	"context"
	"ecst-payment/internal/appctx"
	"ecst-payment/internal/consts"
	"ecst-payment/internal/ucase/contract"
	"ecst-payment/pkg/msg"
	"net/http"
)

// HttpRequest handler func wrapper
func HttpRequest(request *http.Request, svc contract.UseCase, conf *appctx.Config) appctx.Response {
	if !msg.GetAvailableLang(consts.ResponseSuccess, request.Header.Get(consts.HeaderLanguageKey)) {
		request.Header.Set(consts.HeaderLanguageKey, conf.App.DefaultLang)
	}

	ctx := context.WithValue(request.Context(), consts.CtxLang, request.Header.Get(consts.HeaderLanguageKey))

	req := request.WithContext(ctx)

	data := &appctx.Data{
		Request:     req,
		Config:      conf,
		ServiceType: consts.ServiceTypeHTTP,
	}

	return svc.Serve(data)
}
