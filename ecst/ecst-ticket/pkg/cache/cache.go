package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

// Closure define callback, when returning error
type Closure func(bytes []byte) error

const (
	cacheNil string = `redis: nil`
)

// Cacher contract
type Cacher interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, val interface{}, duration time.Duration) error
	BulkSet(ctx context.Context, keys []string, val []interface{}, duration time.Duration) error
	Delete(ctx context.Context, key ...string) error
	Increment(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, exp time.Duration) error
	HSet(ctx context.Context, key string, val ...interface{}) error
	HGet(ctx context.Context, key, field string) ([]byte, error)
	HLen(ctx context.Context, key string) (int64, error)
	HKeys(ctx context.Context, key string) ([]string, error)
	HDel(ctx context.Context, key, field string) error
}

type cache struct {
	rds             redis.Cmdable
	retentionSecond time.Duration
}

// NewCache creates new agent redis client
func NewCache(redis redis.Cmdable) Cacher {
	return &cache{
		rds: redis,
	}
}

func (c *cache) Set(ctx context.Context, key string, val interface{}, exp time.Duration) error {
	cmd := c.rds.Set(ctx, key, val, exp)
	return cmd.Err()
}

func (c *cache) BulkSet(ctx context.Context, keys []string, vals []interface{}, exp time.Duration) error {
	pipe := c.rds.Pipeline()
	for i, key := range keys {
		err := pipe.Set(ctx, key, vals[i], exp).Err()
		if err != nil {
			return err
		}
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (c *cache) Get(ctx context.Context, key string) ([]byte, error) {
	cmd := c.rds.Get(ctx, key)
	b, e := cmd.Bytes()

	if e != nil {
		if e.Error() == cacheNil {
			return b, nil
		}
	}

	return b, e
}

func (c *cache) Delete(ctx context.Context, key ...string) error {
	cmd := c.rds.Del(ctx, key...)
	return cmd.Err()
}

func (c *cache) Increment(ctx context.Context, key string) (int64, error) {
	cmd := c.rds.Incr(ctx, key)
	return cmd.Result()
}

func (c *cache) Expire(ctx context.Context, key string, exp time.Duration) error {
	cmd := c.rds.Expire(ctx, key, exp)
	return cmd.Err()
}

func (c *cache) HSet(ctx context.Context, key string, val ...interface{}) error {
	cmd := c.rds.HSet(ctx, key, val)
	return cmd.Err()
}

func (c *cache) HGet(ctx context.Context, key, field string) ([]byte, error) {
	cmd := c.rds.HGet(ctx, key, field)
	b, e := cmd.Bytes()

	if e != nil {
		if e.Error() == cacheNil {
			return b, nil
		}
	}

	return b, e
}

func (c *cache) HLen(ctx context.Context, key string) (int64, error) {
	cmd := c.rds.HLen(ctx, key)
	return cmd.Result()
}

func (c *cache) HKeys(ctx context.Context, key string) ([]string, error) {
	cmd := c.rds.HKeys(ctx, key)
	return cmd.Result()
}

func (c *cache) HDel(ctx context.Context, key, field string) error {
	cmd := c.rds.HDel(ctx, key, field)
	return cmd.Err()
}
