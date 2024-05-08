package dto

import (
	"ecst-order/internal/consts"
	"ecst-order/internal/presentations"
	"ecst-order/pkg/httpclient"
)

func TransformHeaderForLogging(req httpclient.RequestOptions) *presentations.LogHeaderOpts {
	return &presentations.LogHeaderOpts{
		Payload: req.Payload,
		URL:     req.URL,
		Header: presentations.HeaderOpts{
			ContentType: req.Header.Get(consts.HeaderContentTypeKey),
		},
		Method:        req.Method,
		TimeoutSecond: req.TimeoutSecond,
	}
}
