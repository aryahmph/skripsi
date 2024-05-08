package order

import (
	"context"
	"en-payment/pkg/httpclient"
)

type OrderProvider interface {
	GetOrder(ctx context.Context, request GetOrderRequest) (req httpclient.RequestOptions, resp GetOrderResponse, err error)
}
