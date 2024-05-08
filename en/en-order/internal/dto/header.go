package dto

import (
	"en-order/internal/consts"
	"en-order/internal/presentations"
	"en-order/pkg/httpclient"
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
