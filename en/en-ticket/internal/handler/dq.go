package handler

import (
	"context"
	"en-ticket/internal/appctx"
	ucaseContract "en-ticket/internal/ucase/contract"
	"en-ticket/pkg/dq"
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
