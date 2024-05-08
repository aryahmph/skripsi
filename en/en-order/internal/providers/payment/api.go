package payment

import (
	"context"
	"en-order/internal/appctx"
	"en-order/internal/consts"
	"en-order/pkg/httpclient"
	"en-order/pkg/tracer"
	"en-order/pkg/util"
	"fmt"
	"github.com/spf13/cast"
	"net/http"
	"strings"
)

type paymentProvider struct {
	config *appctx.Config
}

func (api *paymentProvider) GetPayment(ctx context.Context, request GetPaymentRequest) (req httpclient.RequestOptions, resp GetPaymentResponse, err error) {
	header := httpclient.Headers{}

	endpoint := fmt.Sprintf("%s%s",
		api.config.Dependency.InternalGateway.PaymentService.BaseUrl,
		api.config.Dependency.InternalGateway.PaymentService.GetPaymentPath,
	)
	endpoint = strings.ReplaceAll(endpoint, `{id}`, request.ID)

	if request.UserID != "" {
		endpoint = fmt.Sprintf("%s?user_id=%s", endpoint, request.UserID)
	}

	header.Add(httpclient.ContentType, httpclient.MediaTypeJSON)
	header.Add(consts.HeaderXClientName, cast.ToString(request.ClientName))

	req = httpclient.RequestOptions{
		URL:           endpoint,
		TimeoutSecond: api.config.Dependency.InternalGateway.TimeoutSecond,
		Method:        http.MethodGet,
		Context:       ctx,
		Header:        header,
	}

	rsp, err := httpclient.Request(req)

	ctx = tracer.InternalAPISpanStartWithOption(ctx, consts.PaymentDependencyName, consts.PaymentGetPaymentEventName,
		tracer.WithOptions("http.url", req.URL),
		tracer.WithOptions("http.method", req.Method),
		tracer.WithOptions("http.header", util.DumpToString(req.Header)),
		tracer.WithOptions("http.request", util.DumpToString(req.Payload)),
		tracer.WithOptions("http.response", rsp.String()),
		tracer.WithOptions("http.status_code", cast.ToString(rsp.Status())),
		tracer.WithOptions("http.duration_ms", cast.ToString(rsp.Latency())),
	)
	defer tracer.SpanFinish(ctx)

	if err != nil {
		tracer.SpanError(ctx, err)
		resp.Name = consts.ResponseInternalFailure
		return req, resp, err
	}

	if rsp.Status() == http.StatusRequestTimeout {
		resp.Name = consts.ResponseRequestTimeout
		return req, resp, err
	}

	if rsp.Status() >= http.StatusInternalServerError {
		err = fmt.Errorf("error %s", rsp.String())
		tracer.SpanError(ctx, err)
		resp.Name = consts.ResponseInternalFailure
		return req, resp, err
	}

	err = rsp.DecodeJSON(&resp)
	if err != nil {
		tracer.SpanError(ctx, err)
		resp.Name = consts.ResponseInternalFailure
		err = fmt.Errorf("err: %s, raw response: %s", err.Error(), rsp.String())
		return req, resp, err
	}

	return req, resp, err
}

func NewPaymentProvider(config *appctx.Config) PaymentProvider {
	return &paymentProvider{config: config}
}
