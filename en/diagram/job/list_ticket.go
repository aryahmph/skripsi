package job

import (
	"context"
	"en-ticket/internal/appctx"
	"en-ticket/internal/bootstrap"
	"en-ticket/internal/handler"
	"en-ticket/internal/repositories"
	"en-ticket/internal/ucase/job"
	"en-ticket/pkg/cache"
	"en-ticket/pkg/dq"
	"en-ticket/pkg/logger"
	"en-ticket/pkg/postgres"
	"fmt"
	"time"
)

func RunJobListTicket(ctx context.Context) {
	cfg, err := appctx.NewConfig()
	if err != nil {
		logger.Fatal(fmt.Sprintf("load config error %v", err))
	}

	bootstrap.RegistryLogger(cfg)

	var db postgres.Adapter
	if cfg.App.IsSingle {
		db = bootstrap.RegistryPostgresDBSingle(cfg.DBWrite, cfg.App.Timezone)
	} else {
		db = bootstrap.RegistryPostgresDB(cfg.DBWrite, cfg.DBRead, cfg.App.Timezone)
	}

	cacher := cache.NewCache(bootstrap.RegistryRedisNative(cfg))
	dqWatcher := dq.NewWatcher(cacher)
	dqp := dq.NewProducer(cacher)

	ticketRepo := repositories.NewTicketRepository(db)

	ucase := job.NewUpdateTicketCache(cfg, ticketRepo, cacher, dqp)

	dqWatcher.Watch(&dq.WatcherContext{
		Handler:  handler.DQWatcherHandler(ucase),
		Context:  ctx,
		Name:     cfg.Job.Ticket.QueueName,
		Interval: time.Duration(cfg.Job.Ticket.IntervalSecond) * time.Second,
		Limit:    cfg.Job.Ticket.DataPerInterval,
	})
}
