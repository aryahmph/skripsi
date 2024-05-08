package handler

import (
	"context"
	"ecst-ticket/internal/appctx"
	"ecst-ticket/internal/consts"
	"ecst-ticket/pkg/kafka"

	uContract "ecst-ticket/internal/ucase/contract"
)

// KafkaConsumerHandler kafka consumer message processor handler
func KafkaConsumerHandler(msgHandler uContract.MessageProcessor) kafka.MessageProcessorFunc {
	return func(decoder *kafka.MessageDecoder) {
		err := msgHandler.Serve(context.Background(), &appctx.ConsumerData{
			Body:        decoder.Body,
			Key:         decoder.Key,
			ServiceType: consts.ServiceTypeConsumer,
			Topic:       decoder.Topic,
			Commit: func() {
				decoder.Commit(decoder)
			},
		})

		if err != nil {
			return
		}

	}
}
