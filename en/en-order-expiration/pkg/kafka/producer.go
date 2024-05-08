package kafka

import (
	"context"
	"en-order-expiration/pkg/logger"
	"en-order-expiration/pkg/tracer"
	"fmt"
	"github.com/IBM/sarama"
	"strings"
	"time"
)

const (
	defaultTimeout = 3 // in second
)

var (
	partitions = map[string]sarama.PartitionerConstructor{
		"hash":       sarama.NewHashPartitioner,
		"roundrobin": sarama.NewRoundRobinPartitioner,
		"reference":  sarama.NewReferenceHashPartitioner,
		"random":     sarama.NewRandomPartitioner,
		"manual":     sarama.NewManualPartitioner,
	}
)

type producer struct {
	config   *sarama.Config
	brokers  []string
	producer sarama.SyncProducer
}

// Publish  message synchronously
func (k *producer) Publish(ctx context.Context, msg *MessageContext) error {

	param := &sarama.ProducerMessage{
		Topic:     msg.Topic,
		Value:     sarama.StringEncoder(msg.Value),
		Partition: msg.Partition,
		Offset:    msg.Offset,
		Timestamp: msg.TimeStamp,
	}

	ctx = tracer.KafkaSpanStartWithOption(ctx, "kafka.produce",
		tracer.WithOptions("resource.name", msg.Topic),
	)

	defer tracer.SpanFinish(ctx)

	if msg.Key != nil && len(msg.Key) > 0 {
		param.Key = sarama.ByteEncoder(msg.Key)
	}

	partition, offset, err := k.producer.SendMessage(param)

	if err != nil {
		tracer.SpanError(ctx, err)
		return fmt.Errorf("[kafka-publisher] topic: %s, partition %d, offset %d, id %v, got:%s ", msg.Topic, partition, offset, msg.LogId, err.Error())
	}

	if msg.Verbose {
		logger.Info(fmt.Sprintf("[kafka-publisher] topic: %s,  partition: %d, offset: %d", msg.Topic, partition, offset), logger.Any("msg", msg.Value))
	}

	return nil
}

// NewProducer return message producer
func NewProducer(cfg *Config) Producer {

	m := &producer{}
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

	config.Producer.Idempotent = cfg.Producer.IdemPotent
	config.Producer.RequiredAcks = sarama.RequiredAcks(cfg.Producer.RequireACK)

	if cfg.Producer.IdemPotent {
		config.Producer.RequiredAcks = sarama.WaitForAll
		config.Net.MaxOpenRequests = 1
	}

	config.Version = version

	if len(strings.Trim(cfg.Producer.PartitionStrategy, " ")) == 0 {
		cfg.Producer.PartitionStrategy = "hash"
	}

	strategy, ok := partitions[cfg.Producer.PartitionStrategy]

	if !ok {
		logger.Fatal(logger.SetMessageFormat("[kafka] invalid producer partition strategy %s", cfg.Producer.PartitionStrategy))
	}

	config.Producer.Partitioner = strategy

	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	config.Producer.Timeout = time.Duration(cfg.Producer.TimeoutSecond) * time.Second

	if cfg.Producer.TimeoutSecond < 1 {
		config.Producer.Timeout = defaultTimeout * time.Second
	}

	if cfg.ChannelBufferSize > 0 {
		config.ChannelBufferSize = cfg.ChannelBufferSize
	}

	m.brokers = cfg.Brokers
	m.config = config

	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)

	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to start Sarama producer:%s", err.Error()))
	}

	m.producer = producer

	return m
}
