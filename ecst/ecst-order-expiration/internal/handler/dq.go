package handler

import (
	"context"
	"ecst-order-expiration/internal/appctx"
	ucaseContract "ecst-order-expiration/internal/ucase/contract"
	"ecst-order-expiration/pkg/dq"
)

func DQWatcherHandler(jobHandler ucaseContract.JobProcessor) dq.JobProcessorFunc {
	return func(decoder *dq.JobDecoder) {
		err := jobHandler.Serve(context.Background(), &appctx.WatcherData{
			Name:  decoder.Name,
			Key:   decoder.Key,
			Value: decoder.Value,
			Commit: func() {
				decoder.Commit(decoder)
			},
		})
		if err != nil {
			return
		}
	}
}
