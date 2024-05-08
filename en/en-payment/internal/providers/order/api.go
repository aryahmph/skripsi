package order

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"en-payment/internal/appctx"
	"en-payment/internal/consts"
	"en-payment/pkg/httpclient"
	"en-payment/pkg/tracer"
	"en-payment/pkg/util"

	"github.com/spf13/cast"
)

type orderProvider struct {
	config *appctx.Config
}

func (api *orderProvider) GetOrder(ctx context.Context, request GetOrderRequest) (req httpclient.RequestOptions, resp GetOrderResponse, err error) {
	header := httpclient.Headers{}

	endpointUrl := fmt.Sprintf("%s%s",
		api.config.Dependency.InternalGateway.OrderService.BaseUrl,
		api.config.Dependency.InternalGateway.OrderService.GetOrderPath,
	)
	endpointUrl = strings.ReplaceAll(endpointUrl, `{id}`, request.ID)

	endpoint, _ := url.Parse(endpointUrl)
	query := endpoint.Query()

	if request.UserID != "" {
		query.Set("user_id", request.UserID)
	}

	endpoint.RawQuery = query.Encode()

	header.Add(httpclient.ContentType, httpclient.MediaTypeJSON)
	header.Add(consts.HeaderXClientName, cast.ToString(request.ClientName))

	req = httpclient.RequestOptions{
		URL:           endpoint.String(),
		TimeoutSecond: api.config.Dependency.InternalGateway.TimeoutSecond,
		Method:        http.MethodGet,
		Context:       ctx,
		Header:        header,
	}

	rsp, err := httpclient.Request(req)

	ctx = tracer.InternalAPISpanStartWithOption(ctx, consts.OrderDependencyName, consts.OrderGetOrderEventName,
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

func NewOrderProvider(config *appctx.Config) OrderProvider {
	return &orderProvider{config: config}
}
