package contract

import (
	"context"
	"ecst-order-expiration/internal/appctx"
)

type JobProcessor interface {
	Serve(ctx context.Context, data *appctx.WatcherData) error
}
