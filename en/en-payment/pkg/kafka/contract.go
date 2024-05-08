package kafka

import (
	"context"
	"github.com/IBM/sarama"
	"time"
)

// Producer represents kafka publisher message topic
type Producer interface {
	Publish(ctx context.Context, msg *MessageContext) error
}

// Consumer represents a Sarama consumer consumer interface
type Consumer interface {
	Subscribe(*ConsumerContext)
}

type MessageContext struct {
	Value     string
	Key       []byte
	LogId     interface{}
	Topic     string
	Partition int32
	Offset    int64
	TimeStamp time.Time
	Verbose   bool
}

type ConsumerContext struct {
	Handler MessageProcessorFunc
	Topics  []string
	GroupID string
	Context context.Context
}

var balanceStrategies = map[string]sarama.BalanceStrategy{
	sarama.RoundRobinBalanceStrategyName: sarama.NewBalanceStrategyRoundRobin(),
	sarama.RangeBalanceStrategyName:      sarama.NewBalanceStrategyRange(),
	sarama.StickyBalanceStrategyName:     sarama.NewBalanceStrategySticky(),
}

var offsetInitials = map[string]int64{
	"oldest": -2,
	"newest": -1,
}
