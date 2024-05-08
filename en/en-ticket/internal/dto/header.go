package dto

import (
	"en-ticket/internal/consts"
	"en-ticket/internal/presentations"
	"en-ticket/pkg/httpclient"
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
