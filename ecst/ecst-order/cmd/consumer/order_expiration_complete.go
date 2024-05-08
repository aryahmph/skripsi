package consumer

import (
	"context"
	"fmt"

	"ecst-order/internal/appctx"
	"ecst-order/internal/bootstrap"
	"ecst-order/internal/handler"
	"ecst-order/internal/repositories"
	"ecst-order/internal/ucase/consumer"

	"ecst-order/pkg/kafka"
	"ecst-order/pkg/logger"
)

func RunConsumerOrderExpirationComplete(ctx context.Context) {
	cfg, err := appctx.NewConfig()
	if err != nil {
		logger.Fatal(fmt.Sprintf("load config error %v", err))
	}

	bootstrap.RegistryLogger(cfg)

	kc := bootstrap.RegistryKafkaConsumer(cfg)
	db := bootstrap.RegistryPostgresDBSingle(cfg.DBWrite, cfg.App.Timezone)
	kp := bootstrap.RegistryKafkaProducer(cfg)

	orderRepo := repositories.NewOrderRepository(db)
	ucase := consumer.NewOrderExpire(cfg, kp, orderRepo)

	kc.Subscribe(&kafka.ConsumerContext{
		Handler: handler.KafkaConsumerHandler(ucase),
		Context: ctx,
		GroupID: cfg.KafkaConsumerIds.ConsumerOrderExpire,
		Topics:  []string{cfg.KafkaTopics.TopicOrderExpirationComplete},
	})
}
