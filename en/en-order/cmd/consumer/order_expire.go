package consumer

import (
	"context"
	"en-order/pkg/cache"
	"en-order/pkg/postgres"
	"fmt"

	"en-order/internal/appctx"
	"en-order/internal/bootstrap"
	"en-order/internal/handler"
	"en-order/internal/repositories"
	"en-order/internal/ucase/consumer"

	"en-order/pkg/kafka"
	"en-order/pkg/logger"
)

func RunConsumerOrderExpire(ctx context.Context) {
	cfg, err := appctx.NewConfig()
	if err != nil {
		logger.Fatal(fmt.Sprintf("load config error %v", err))
	}

	bootstrap.RegistryLogger(cfg)

	kc := bootstrap.RegistryKafkaConsumer(cfg)
	kp := bootstrap.RegistryKafkaProducer(cfg)
	var db postgres.Adapter

	if cfg.App.IsSingle {
		db = bootstrap.RegistryPostgresDBSingle(cfg.DBWrite, cfg.App.Timezone)
	} else {
		db = bootstrap.RegistryPostgresDB(cfg.DBWrite, cfg.DBRead, cfg.App.Timezone)
	}
	cacher := cache.NewCache(bootstrap.RegistryRedisNative(cfg))

	orderRepo := repositories.NewOrderRepository(db)
	ucase := consumer.NewOrderExpire(cfg, cacher, orderRepo, kp)

	kc.Subscribe(&kafka.ConsumerContext{
		Handler: handler.KafkaConsumerHandler(ucase),
		Context: ctx,
		GroupID: cfg.KafkaConsumerIds.ConsumerOrderExpire,
		Topics:  []string{cfg.KafkaTopics.TopicOrderExpire},
	})
}
