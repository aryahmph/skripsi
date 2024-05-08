package consumer

import (
	"context"
	"en-ticket/pkg/postgres"
	"fmt"

	"en-ticket/internal/appctx"
	"en-ticket/internal/bootstrap"
	"en-ticket/internal/handler"
	"en-ticket/internal/providers/order"
	"en-ticket/internal/repositories"
	"en-ticket/internal/ucase/consumer"

	"en-ticket/pkg/cache"
	"en-ticket/pkg/kafka"
	"en-ticket/pkg/logger"
)

func RunConsumerAssignTicketOrder(ctx context.Context) {
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

	orderProvider := order.NewOrderProvider(cfg)
	ticketRepo := repositories.NewTicketRepository(db)

	ucase := consumer.NewAssignTicketOrder(cfg, cacher, orderProvider, ticketRepo)

	kc.Subscribe(&kafka.ConsumerContext{
		Handler: handler.KafkaConsumerHandler(ucase),
		Context: ctx,
		GroupID: cfg.KafkaConsumerIds.ConsumerCreateOrder,
		Topics:  []string{cfg.KafkaTopics.TopicCreateOrder},
	})
}
