// Package bootstrap
package bootstrap

import (
	"ecst-ticket/internal/appctx"
	"ecst-ticket/pkg/kafka"
	"strings"
)

func RegistryKafkaConsumer(cfg *appctx.Config) kafka.Consumer {
	return kafka.NewConsumerGroup(&kafka.Config{
		Consumer: kafka.ConsumerConfig{
			SessionTimeoutSecond: cfg.Kafka.Consumer.SessionTimeoutSecond,
			HeartbeatInterval:    cfg.Kafka.Consumer.HeartbeatIntervalMS,
			RebalanceStrategy:    cfg.Kafka.Consumer.RebalanceStrategy,
			OffsetInitial:        cfg.Kafka.Consumer.OffsetInitial,
		},
		Version:  cfg.Kafka.Version,
		Brokers:  strings.Split(cfg.Kafka.Brokers, ","),
		ClientID: cfg.Kafka.ClientID,

		ChannelBufferSize: cfg.Kafka.ChannelBufferSize,
	})
}
