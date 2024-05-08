package job

import (
	"context"
	"fmt"
	"time"

	"ecst-order-expiration/internal/appctx"
	"ecst-order-expiration/internal/bootstrap"
	"ecst-order-expiration/internal/handler"
	"ecst-order-expiration/internal/ucase/job"

	"ecst-order-expiration/pkg/cache"
	"ecst-order-expiration/pkg/dq"
	"ecst-order-expiration/pkg/logger"
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
