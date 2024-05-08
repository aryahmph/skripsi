package consumer

import (
	"context"
	"ecst-payment/internal/appctx"
	"ecst-payment/internal/bootstrap"
	"ecst-payment/internal/handler"
	"ecst-payment/internal/repositories"
	"ecst-payment/internal/ucase/consumer"
	"ecst-payment/pkg/kafka"
	"ecst-payment/pkg/logger"
	"fmt"
)

func RunConsumerOrderExpire(ctx context.Context) {
	cfg, err := appctx.NewConfig()
	if err != nil {
		logger.Fatal(fmt.Sprintf("load config error %v", err))
	}

	bootstrap.RegistryLogger(cfg)

	kc := bootstrap.RegistryKafkaConsumer(cfg)
	db := bootstrap.RegistryPostgresDBSingle(cfg.DBWrite, cfg.App.Timezone)

	orderRepo := repositories.NewOrderRepository(db)
	ucase := consumer.NewOrderExpire(cfg, orderRepo)

	kc.Subscribe(&kafka.ConsumerContext{
		Handler: handler.KafkaConsumerHandler(ucase),
		Context: ctx,
		GroupID: cfg.KafkaConsumerIds.ConsumerOrderExpire,
		Topics:  []string{cfg.KafkaTopics.TopicOrderExpire},
	})
}
