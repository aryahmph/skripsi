package dq

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"en-ticket/pkg/cache"
	"en-ticket/pkg/logger"
)

type watcher struct {
	client cache.Cacher
}

func NewWatcher(client cache.Cacher) Watcher {
	return &watcher{
		client: client,
	}
}

func (dq *watcher) Watch(ctx *WatcherContext) {
	ticker := time.NewTicker(ctx.Interval)
	defer func() {
		ticker.Stop()
	}()

	newCtx, cancel := context.WithCancel(ctx.Context)

	go func() {
		for {
			select {
			case <-newCtx.Done():
				logger.Warn(logger.SetMessageFormat("[delay queue] stopped consume queue %v", ctx.Name))
				return
			case <-ticker.C:
				keys, values, err := dq.pulls(newCtx, ctx.Name, ctx.Limit)
				if err != nil {
					logger.Error(logger.SetMessageFormat("[delay queue] queue %s error pull jobs, err: %v", ctx.Name, err))
					continue
				}

				for i, key := range keys {
					go func(newCtx context.Context, key int64, value string) {
						ctx.Handler(&JobDecoder{
							Name:  ctx.Name,
							Key:   key,
							Value: value,
							Commit: func(decoder *JobDecoder) {
								err := dq.remove(newCtx, ctx.Name, decoder.Value)
								if err != nil {
									logger.Error(logger.SetMessageFormat("[delay queue] queue %s error remove jobs, err: %v", ctx.Name, err))
								}
							},
						})
					}(newCtx, key, values[i])
				}

			}
		}
	}()

	logger.Info(fmt.Sprintf("[delay queue] watch %s queue is running", ctx.Name))

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	<-sigterm

	cancel()
	logger.Info(fmt.Sprintf("[delay queue] cancelled watch %s queue", ctx.Name))
}

func (dq *watcher) pulls(ctx context.Context, queueName string, limit int64) (keys []int64, values []string, err error) {
	now := time.Now().UnixMilli()
	keys, values, err = dq.client.ZRangeByScoreWithScores(ctx, queueName, cache.ZRangeBy{
		Min:   "-inf",
		Max:   fmt.Sprintf("%d", now),
		Count: limit,
	})
	if err != nil {
		return keys, values, err
	}

	return
}

func (dq *watcher) remove(ctx context.Context, queueName string, value string) error {
	return dq.client.ZRem(ctx, queueName, value)
}
