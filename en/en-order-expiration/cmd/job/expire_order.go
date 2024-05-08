package job

import (
	"context"
	"fmt"
	"time"

	"en-order-expiration/internal/appctx"
	"en-order-expiration/internal/bootstrap"
	"en-order-expiration/internal/handler"
	"en-order-expiration/internal/ucase/job"

	"en-order-expiration/pkg/cache"
	"en-order-expiration/pkg/dq"
	"en-order-expiration/pkg/logger"
)

func RunJobExpireOrder(ctx context.Context) {
	cfg, err := appctx.NewConfig()
	if err != nil {
		logger.Fatal(fmt.Sprintf("load config error %v", err))
	}

	bootstrap.RegistryLogger(cfg)

	kp := bootstrap.RegistryKafkaProducer(cfg)

	cacher := cache.NewCache(bootstrap.RegistryRedisNative(cfg))
	dqWatcher := dq.NewWatcher(cacher)

	ucase := job.NewExpireOrder(cfg, kp)

	dqWatcher.Watch(&dq.WatcherContext{
		Handler:  handler.DQWatcherHandler(ucase),
		Context:  ctx,
		Name:     cfg.Job.OrderExpire.QueueName,
		Interval: time.Duration(cfg.Job.OrderExpire.IntervalSecond) * time.Second,
		Limit:    cfg.Job.OrderExpire.DataPerInterval,
	})
}
