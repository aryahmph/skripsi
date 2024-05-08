package dto

import (
	"en-payment/internal/consts"
	"en-payment/internal/presentations"
	"en-payment/pkg/httpclient"
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
