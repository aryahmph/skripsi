package dq

import (
	"context"
	"en-ticket/pkg/cache"
)

type producer struct {
	client cache.Cacher
}

func NewProducer(client cache.Cacher) Producer {
	return &producer{client: client}
}

func (dq *producer) Add(ctx context.Context, job *JobContext) error {
	return dq.client.ZAdd(ctx, job.QueueName, &cache.Z{
		Score:  job.ExpiredAt,
		Member: job.Value,
	})
}
