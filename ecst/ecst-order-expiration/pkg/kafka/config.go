package kafka

const (
	defaultVersion = "2.1.1"
)

// Config entity of kafka broker
type Config struct {
	// Brokers list of brokers connection hostname or ip address
	Brokers []string `json:"brokers" yaml:"brokers"`
	// kafka broker Version
	Version  string         `json:"version" yaml:"version"`
	ClientID string         `json:"client_id" yaml:"client_id"`
	Producer ProducerConfig `json:"producer" yaml:"producer"`
	Consumer ConsumerConfig `json:"consumer" yaml:"consumer"`
	// The number of events to buffer in internal and external channels. This
	// permits the producer and consumer to continue processing some messages
	// in the background while user code is working, greatly improving throughput.
	// Defaults to 256.
	ChannelBufferSize int `json:"channel_buffer_size" yaml:"channel_buffer_size"`
}

type ProducerConfig struct {
	// The maximum duration the broker will wait the receipt of the number of
	// RequiredAcks (defaults to 10 seconds). This is only relevant when
	// RequiredAcks is set to WaitForAll or a number > 1. Only supports
	// millisecond resolution, nanoseconds will be truncated. Equivalent to
	// the JVM producer's `request.timeout.ms` setting.
	TimeoutSecond int `json:"timeout_second" yaml:"timeout_second"`
	// RequireACK
	// 0 = NoResponse doesn't send any response, the TCP ACK is all you get.
	// 1 =  WaitForLocal waits for only the local commit to succeed before responding.
	// - 1 =  WaitForAll
	// WaitForAll waits for all in-sync replicas to commit before responding.
	// The minimum number of in-sync replicas is configured on the broker via
	// the `min.insync.replicas` configuration key.
	RequireACK int16 `json:"require_ack" yaml:"require_ack"`
	// If enabled, the producer will ensure that exactly one copy of each message is
	// written.
	IdemPotent bool `json:"idem_potent" yaml:"idem_potent"`

	// Generates partitioners for choosing the partition to send messages to
	// (defaults to hashing the message key). Similar to the `partitioner.class`
	// setting for the JVM producer.
	PartitionStrategy string `json:"partition_strategy" yaml:"partition_strategy"`
}

type ConsumerConfig struct {
	// Minimum is 10s
	SessionTimeoutSecond int    `json:"session_timeout_second" yaml:"session_timeout_second"`
	OffsetInitial        int64  `json:"offset_initial" yaml:"offset_initial"`
	HeartbeatInterval    int    `json:"heartbeat_interval" yaml:"heartbeat_interval"`
	RebalanceStrategy    string `json:"rebalance_strategy" yaml:"rebalance_strategy"`
	AutoCommit           bool   `json:"auto_commit" yaml:"auto_commit"`
	IsolationLevel       int8   `json:"isolation_level" yaml:"isolation_level"`
}
