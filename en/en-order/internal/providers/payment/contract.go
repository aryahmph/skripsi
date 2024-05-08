package payment

import (
	"context"
	"en-order/pkg/httpclient"
)

type PaymentProvider interface {
	GetPayment(ctx context.Context, request GetPaymentRequest) (req httpclient.RequestOptions, resp GetPaymentResponse, err error)
}
