package appctx

import (
	"fmt"

	"ecst-order-expiration/internal/consts"
	"ecst-order-expiration/pkg/file"
)

type (
	Config struct {
		App             *Common         `yaml:"app" json:"app"`
		Kafka           *KafkaConfig    `yaml:"kafka" json:"kafka"`
		APM             APM             `yaml:"apm" json:"apm"`
		Redis           *RedisConf      `yaml:"redis" json:"redis"`
		Job             Job             `yaml:"job" json:"job"`
		KafkaTopics     KafkaTopics     `yaml:"kafka_topics" json:"kafka_topics"`
		KafkaEETSecond  KafkaEETSecond  `yaml:"kafka_eet_second" json:"kafka_eet_second"`
		KafkaNextSecond KafkaNextSecond `yaml:"kafka_next_second" json:"kafka_next_second"`
	}

	Common struct {
		AppName string `yaml:"name" json:"name"`
		Debug   bool   `yaml:"debug" json:"debug"`
		Env     string `yaml:"env" json:"env"`
	}

	KafkaConfig struct {
		// Brokers list of brokers connection hostname or ip address
		Brokers string `yaml:"brokers" json:"brokers"`
		// kafka broker Version
		Version  string        `yaml:"version" json:"version"`
		ClientID string        `yaml:"client_id" json:"client_id"`
		Producer KafkaProducer `yaml:"producer" json:"producer"`
		Consumer KafkaConsumer `yaml:"consumer" json:"consumer"`
		// The number of events to buffer in internal and external channels. This
		// permits the producer and consumer to continue processing some messages
		// in the background while user code is working, greatly improving throughput.
		// Defaults to 256.
		ChannelBufferSize int `json:"channel_buffer_size" yaml:"channel_buffer_size"`
	}

	APM struct {
		Address string `yaml:"address" json:"address"`
		Enable  bool   `yaml:"enable" json:"enable"`
		Name    string `yaml:"name" json:"name"`
	}

	RedisConf struct {
		Hosts              string `yaml:"host"`
		DB                 int    `yaml:"db"`
		ReadTimeoutSecond  int    `yaml:"read_timeout_second"`
		WriteTimeoutSecond int    `yaml:"write_timeout_second"`
		PoolSize           int    `yaml:"pool_size"`
		PoolTimeoutSecond  int    `yaml:"pool_timeout_second"`
		MinIdleConn        int    `yaml:"min_idle_conn"`
		IdleTimeoutSecond  int    `yaml:"idle_timeout_second"`
		RouteByLatency     bool   `yaml:"route_by_latency"`
		IdleFrequencyCheck int    `yaml:"idle_frequency_check"`
		Password           string `yaml:"password"`
		ReadOnly           bool   `yaml:"read_only"`
		RouteRandomly      bool   `yaml:"route_randomly"`
	}

	Job struct {
		OrderExpire OrderExpire `yaml:"order_expire"`
	}
)

type (
	KafkaProducer struct {
		// The maximum duration the broker will wait the receipt of the number of
		// RequiredAcks (defaults to 10 seconds). This is only relevant when
		// RequiredAcks is set to WaitForAll or a number > 1. Only supports
		// millisecond resolution, nanoseconds will be truncated. Equivalent to
		// the JVM producer's `request.timeout.ms` setting.
		TimeoutSecond int `yaml:"timeout_second" json:"timeout_second"`
		// RequireACK
		// 0 = NoResponse doesn't send any response, the TCP ACK is all you get.
		// 1 =  WaitForLocal waits for only the local commit to succeed before responding.
		// - 1 =  WaitForAll
		// WaitForAll waits for all in-sync replicas to commit before responding.
		// The minimum number of in-sync replicas is configured on the broker via
		// the `min.insync.replicas` configuration key.
		RequireACK int16 `yaml:"ack" json:"require_ack"`
		// If enabled, the producer will ensure that exactly one copy of each message is
		// written.
		IdemPotent bool `yaml:"idem_potent" json:"idem_potent"`

		// Generates partitioners for choosing the partition to send messages to
		// (defaults to hashing the message key). Similar to the `partitioner.class`
		// setting for the JVM producer.
		PartitionStrategy string `yaml:"partition_strategy" json:"partition_strategy"`
	}

	KafkaConsumer struct {
		// Minimum is 10s
		SessionTimeoutSecond int    `yaml:"session_timeout_second" json:"session_timeout_second"`
		OffsetInitial        int64  `yaml:"offset_initial" json:"offset_initial"`
		HeartbeatIntervalMS  int    `yaml:"heartbeat_interval_ms" json:"heartbeat_interval_ms"`
		RebalanceStrategy    string `yaml:"rebalance_strategy" json:"rebalance_strategy"`
		AutoCommit           bool   `yaml:"auto_commit" json:"auto_commit"`
		IsolationLevel       int8   `json:"isolation_level" yaml:"isolation_level"`
	}

	KafkaTopics struct {
		TopicOrderExpirationComplete string `yaml:"topic_order_expiration_complete" json:"topic_order_expiration_complete"`
	}

	KafkaEETSecond struct {
		EETOrderExpire int64 `yaml:"eet_order_expire" json:"eet_order_expire"`
	}

	KafkaNextSecond struct {
		NextOrderExpire int64 `yaml:"next_order_expire" json:"next_order_expire"`
	}
)

type (
	OrderExpire struct {
		QueueName       string `yaml:"queue_name"`
		IntervalSecond  int64  `yaml:"interval_second"`
		DataPerInterval int64  `yaml:"data_per_interval"`
	}
)

// NewConfig initialize config object
func NewConfig() (*Config, error) {
	fpath := []string{consts.ConfigPath}
	cfg, err := readCfg("app.yaml", fpath...)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// readCfg reads the configuration from file
// args:
//
//	fname: filename
//	ps: full path of possible configuration files
//
// returns:
//
//	*config.Configuration: configuration ptr object
//	error: error operation
func readCfg(fname string, ps ...string) (*Config, error) {
	var cfg *Config
	var errs []error

	for _, p := range ps {
		f := fmt.Sprint(p, fname)

		err := file.ReadFromYAML(f, &cfg)
		if err != nil {
			errs = append(errs, fmt.Errorf("file %s error %s", f, err.Error()))
			continue
		}
		break
	}

	if cfg == nil {
		return nil, fmt.Errorf("file config parse error %v", errs)
	}

	return cfg, nil
}
