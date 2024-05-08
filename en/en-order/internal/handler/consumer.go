package handler

import (
	"context"
	"en-order/internal/appctx"
	"en-order/internal/consts"
	"en-order/pkg/kafka"

	uContract "en-order/internal/ucase/contract"
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
