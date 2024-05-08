package dto

import (
	"ecst-payment/internal/consts"
	"ecst-payment/internal/presentations"
	"ecst-payment/pkg/httpclient"
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
