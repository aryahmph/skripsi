package consumer

import (
	"context"
	"en-order/internal/appctx"
	"en-order/internal/bootstrap"
	"en-order/internal/handler"
	"en-order/internal/providers/payment"
	"en-order/internal/repositories"
	"en-order/internal/ucase/consumer"
	"en-order/pkg/cache"
	"en-order/pkg/kafka"
	"en-order/pkg/logger"
	"en-order/pkg/postgres"
	"fmt"
)

func RunConsumerOrderComplete(ctx context.Context) {
	cfg, err := appctx.NewConfig()
	if err != nil {
		logger.Fatal(fmt.Sprintf("load config error %v", err))
	}

	bootstrap.RegistryLogger(cfg)

	kc := bootstrap.RegistryKafkaConsumer(cfg)
	var db postgres.Adapter

	if cfg.App.IsSingle {
		db = bootstrap.RegistryPostgresDBSingle(cfg.DBWrite, cfg.App.Timezone)
	} else {
		db = bootstrap.RegistryPostgresDB(cfg.DBWrite, cfg.DBRead, cfg.App.Timezone)
	}
	cacher := cache.NewCache(bootstrap.RegistryRedisNative(cfg))

	paymentProvider := payment.NewPaymentProvider(cfg)

	orderRepo := repositories.NewOrderRepository(db)
	ucase := consumer.NewOrderComplete(cfg, cacher, orderRepo, paymentProvider)

	kc.Subscribe(&kafka.ConsumerContext{
		Handler: handler.KafkaConsumerHandler(ucase),
		Context: ctx,
		GroupID: cfg.KafkaConsumerIds.ConsumerOrderComplete,
		Topics:  []string{cfg.KafkaTopics.TopicCreatePayment},
	})
}
