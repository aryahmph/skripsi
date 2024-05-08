package consumer

import (
	"context"
	"fmt"

	"ecst-ticket/internal/appctx"
	"ecst-ticket/internal/bootstrap"
	"ecst-ticket/internal/handler"
	"ecst-ticket/internal/repositories"
	"ecst-ticket/internal/ucase/consumer"

	"ecst-ticket/pkg/kafka"
	"ecst-ticket/pkg/logger"
)

func RunConsumerRemoveTicketOrder(ctx context.Context) {
	cfg, err := appctx.NewConfig()
	if err != nil {
		logger.Fatal(fmt.Sprintf("load config error %v", err))
	}

	bootstrap.RegistryLogger(cfg)

	kc := bootstrap.RegistryKafkaConsumer(cfg)
	db := bootstrap.RegistryPostgresDBSingle(cfg.DBWrite, cfg.App.Timezone)
	kp := bootstrap.RegistryKafkaProducer(cfg)

	ticketRepo := repositories.NewTicketRepository(db)

	ucase := consumer.NewRemoveTicketOrder(cfg, kp, ticketRepo)

	kc.Subscribe(&kafka.ConsumerContext{
		Handler: handler.KafkaConsumerHandler(ucase),
		Context: ctx,
		GroupID: cfg.KafkaConsumerIds.ConsumerOrderExpire,
		Topics:  []string{cfg.KafkaTopics.TopicOrderExpire},
	})
}
