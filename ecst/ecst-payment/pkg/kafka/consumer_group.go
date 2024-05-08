package kafka

import (
	"context"
	"ecst-payment/pkg/logger"
	"fmt"
	"github.com/IBM/sarama"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type consumerGroup struct {
	config     *sarama.Config
	brokers    []string
	autoCommit bool
}

// NewConsumerGroup return consumer message broker
func NewConsumerGroup(cfg *Config) Consumer {
	m := &consumerGroup{}

	/**
	 * Construct a new Sarama configuration.
	 * The Kafka cluster version has to be defined before the consumer/producer is initialized.
	 */
	config := sarama.NewConfig()

	if cfg.Version == "" {
		cfg.Version = defaultVersion
	}

	version, err := sarama.ParseKafkaVersion(cfg.Version)
	if err != nil {
		logger.Fatal(fmt.Sprintf("parse kafka version got: %v", err))
	}

	config.Version = version

	config.Consumer.Offsets.Initial = cfg.Consumer.OffsetInitial
	config.Consumer.Return.Errors = true
	config.Consumer.Group.Session.Timeout = time.Duration(cfg.Consumer.SessionTimeoutSecond) * time.Second
	config.Consumer.Group.Heartbeat.Interval = time.Duration(cfg.Consumer.HeartbeatInterval) * time.Millisecond

	if len(strings.Trim(cfg.Consumer.RebalanceStrategy, " ")) == 0 {
		cfg.Consumer.RebalanceStrategy = sarama.RangeBalanceStrategyName
	}

	st, ok := balanceStrategies[cfg.Consumer.RebalanceStrategy]

	if !ok {
		logger.Fatal(fmt.Sprintf(
			`rebalance strateggy only available : "%s", "%s", "%s",   on setting value : "%s"`,
			sarama.RoundRobinBalanceStrategyName,
			sarama.RangeBalanceStrategyName,
			sarama.StickyBalanceStrategyName,
			cfg.Consumer.RebalanceStrategy,
		))
	}

	if cfg.ChannelBufferSize > 0 {
		config.ChannelBufferSize = cfg.ChannelBufferSize
	}

	config.Consumer.IsolationLevel = sarama.IsolationLevel(cfg.Consumer.IsolationLevel)
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{st}

	m.brokers = cfg.Brokers
	m.config = config
	m.autoCommit = cfg.Consumer.AutoCommit
	return m
}

// Subscribe message
func (k *consumerGroup) Subscribe(ctx *ConsumerContext) {
	fields := []logger.Field{
		logger.Any("topics", ctx.Topics),
	}

	client, err := sarama.NewConsumerGroup(k.brokers, ctx.GroupID, k.config)

	if err != nil {
		logger.Fatal(err.Error(), fields...)
	}

	handler := NewConsumerHandler(ctx.Handler, k.autoCommit)

	// kafka consumer client
	nCtx, cancel := context.WithCancel(ctx.Context)

	defer func() {
		_ = client.Close()
	}()

	// subscriber errors
	go func() {
		for err := range client.Errors() {
			logger.Error(fmt.Sprintf("[consumer] error %s", err.Error()), fields...)
		}
	}()

	go func() {
		for {
			select {
			case <-nCtx.Done():
				logger.Warn(logger.SetMessageFormat("[consumer] stopped consume topics %v", ctx.Topics))
				return
			default:
				err := client.Consume(nCtx, ctx.Topics, handler)
				if err != nil {
					logger.Error(logger.SetMessageFormat("[consumer] topic %v consume message error %s", ctx.Topics, err.Error()))
				}
			}
		}
	}()

	logger.Info(fmt.Sprintf("[consumer] sarama consumer up and running!... group %s, queue %v", ctx.Context, ctx.Topics), fields...)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	<-sigterm // Await a sigterm signal before safely closing the consumer

	cancel()
	logger.Info("[consumer] Cancelled message without marking offsets", fields...)

}
