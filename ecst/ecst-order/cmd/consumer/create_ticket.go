package consumer

import (
	"context"
	"ecst-order/internal/appctx"
	"ecst-order/internal/bootstrap"
	"ecst-order/internal/handler"
	"ecst-order/internal/repositories"
	"ecst-order/internal/ucase/consumer"
	"ecst-order/pkg/kafka"
	"ecst-order/pkg/logger"
	"fmt"
)

func RunConsumerCreateTicket(ctx context.Context) {
	cfg, err := appctx.NewConfig()
	if err != nil {
		logger.Fatal(fmt.Sprintf("load config error %v", err))
	}

	bootstrap.RegistryLogger(cfg)

	kc := bootstrap.RegistryKafkaConsumer(cfg)
	db := bootstrap.RegistryPostgresDBSingle(cfg.DBWrite, cfg.App.Timezone)

	ticketRepo := repositories.NewTicketRepository(db)
	ucase := consumer.NewCreateTicket(cfg, ticketRepo)

	kc.Subscribe(&kafka.ConsumerContext{
		Handler: handler.KafkaConsumerHandler(ucase),
		Context: ctx,
		GroupID: cfg.KafkaConsumerIds.ConsumerCreateTicket,
		Topics:  []string{cfg.KafkaTopics.TopicCreateTicket},
	})
}
