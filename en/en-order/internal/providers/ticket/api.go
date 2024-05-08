package ticket

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

type ticketProvider struct {
	config *appctx.Config
}

func (api *ticketProvider) GetTicket(ctx context.Context, request GetTicketRequest) (req httpclient.RequestOptions, resp GetTicketResponse, err error) {
	header := httpclient.Headers{}

	endpoint := fmt.Sprintf("%s%s",
		api.config.Dependency.InternalGateway.TicketService.BaseUrl,
		api.config.Dependency.InternalGateway.TicketService.GetTicketPath,
	)
	endpoint = strings.ReplaceAll(endpoint, `{id}`, request.ID)

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

	ctx = tracer.InternalAPISpanStartWithOption(ctx, consts.TicketDependencyName, consts.TicketGetTicketEventName,
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

func NewTicketProvider(config *appctx.Config) TicketProvider {
	return &ticketProvider{config: config}
}
