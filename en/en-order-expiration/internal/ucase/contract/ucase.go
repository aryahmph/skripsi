package contract

import (
	"context"
	"en-order-expiration/internal/appctx"
)

type JobProcessor interface {
	Serve(ctx context.Context, data *appctx.WatcherData) error
}
