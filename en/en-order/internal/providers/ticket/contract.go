package ticket

import (
	"context"
	"en-order/pkg/httpclient"
)

type TicketProvider interface {
	GetTicket(ctx context.Context, request GetTicketRequest) (req httpclient.RequestOptions, resp GetTicketResponse, err error)
}
