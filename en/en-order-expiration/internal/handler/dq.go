package handler

import (
	"context"
	"en-order-expiration/internal/appctx"
	ucaseContract "en-order-expiration/internal/ucase/contract"
	"en-order-expiration/pkg/dq"
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
