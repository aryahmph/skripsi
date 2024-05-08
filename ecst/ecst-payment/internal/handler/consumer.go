package handler

import (
	"context"
	"ecst-payment/internal/appctx"
	"ecst-payment/internal/consts"
	"ecst-payment/pkg/kafka"

	uContract "ecst-payment/internal/ucase/contract"
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
