// Package bootstrap
package bootstrap

import (
	"context"
	"en-order-expiration/internal/appctx"
	"en-order-expiration/pkg/logger"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

const (
	redisInitializeNil     = `redis cannot connect, please check your config or network`
	redisPingError         = `redis cannot connect, error: %v`
	logFieldHost           = "host"
	logFieldDB             = "db"
	logFieldReadOnly       = "read_only"
	logFieldRouteRandomly  = "route_randomly"
	logFieldRouteByLatency = "route_by_latency"
)

// RegistryRedisNative initiate redis session
func RegistryRedisNative(conf *appctx.Config) redis.Cmdable {
	lf := []logger.Field{
		logger.Any(logFieldHost, conf.Redis.Hosts),
		logger.Any(logFieldDB, conf.Redis.DB),
		logger.Any(logFieldReadOnly, conf.Redis.ReadOnly),
		logger.Any(logFieldRouteByLatency, conf.Redis.RouteByLatency),
		logger.Any(logFieldRouteRandomly, conf.Redis.RouteRandomly),
	}

	return registryRedisUniversal(conf, lf)
}

// registryRedisUniversal initiate redis session
func registryRedisUniversal(conf *appctx.Config, lf []logger.Field) redis.Cmdable {
	cfg := redis.UniversalOptions{
		Addrs:           strings.Split(conf.Redis.Hosts, ","),
		ReadTimeout:     time.Duration(conf.Redis.ReadTimeoutSecond) * time.Second,
		WriteTimeout:    time.Duration(conf.Redis.WriteTimeoutSecond) * time.Second,
		DB:              conf.Redis.DB,
		PoolSize:        conf.Redis.PoolSize,
		PoolTimeout:     time.Duration(conf.Redis.PoolTimeoutSecond) * time.Second,
		MinIdleConns:    conf.Redis.MinIdleConn,
		ConnMaxIdleTime: time.Duration(conf.Redis.IdleTimeoutSecond) * time.Second,
		RouteByLatency:  conf.Redis.RouteByLatency,
		Password:        conf.Redis.Password,
		ReadOnly:        conf.Redis.ReadOnly,
		RouteRandomly:   conf.Redis.RouteRandomly,
	}

	r := redis.NewUniversalClient(&cfg)

	if r == nil {
		logger.Fatal(redisInitializeNil, lf...)
	}

	c := r.Ping(context.Background())

	if err := c.Err(); err != nil {
		logger.Fatal(fmt.Sprintf(redisPingError, err), lf...)
	}

	return r
}
