package handler

import (
	"context"
	"en-payment/internal/appctx"
	"en-payment/internal/consts"
	"en-payment/pkg/kafka"

	uContract "en-payment/internal/ucase/contract"
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
