package bootstrap

import (
	"ecst-payment/internal/appctx"
	"ecst-payment/pkg/kafka"
	"strings"
)

func RegistryKafkaProducer(cfg *appctx.Config) kafka.Producer {
	return kafka.NewProducer(&kafka.Config{
		Producer: kafka.ProducerConfig{
			TimeoutSecond:     cfg.Kafka.Producer.TimeoutSecond,
			RequireACK:        cfg.Kafka.Producer.RequireACK,
			IdemPotent:        cfg.Kafka.Producer.IdemPotent,
			PartitionStrategy: cfg.Kafka.Producer.PartitionStrategy,
		},
		Version:           cfg.Kafka.Version,
		Brokers:           strings.Split(cfg.Kafka.Brokers, ","),
		ClientID:          cfg.Kafka.ClientID,
		ChannelBufferSize: cfg.Kafka.ChannelBufferSize,
	})
}
