package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
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
	ZRem(ctx context.Context, key string, members ...interface{}) error
	ZRangeByScoreWithScores(ctx context.Context, key string, opt ZRangeBy) (keys []int64, vals []string, err error)
	ZAdd(ctx context.Context, key string, members ...*Z) error
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

func (c *cache) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return c.rds.ZRem(ctx, key, members...).Err()
}

func (c *cache) ZRangeByScoreWithScores(ctx context.Context, key string, opt ZRangeBy) (keys []int64, vals []string, err error) {
	cmd := c.rds.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min:    opt.Min,
		Max:    opt.Max,
		Offset: opt.Offset,
		Count:  opt.Count,
	})

	members, err := cmd.Result()
	if err != nil {
		return
	}

	for _, member := range members {
		keys = append(keys, int64(member.Score))
		vals = append(vals, member.Member)
	}

	return
}

func (c *cache) ZAdd(ctx context.Context, key string, members ...*Z) error {
	zs := make([]redis.Z, len(members))
	for i, member := range members {
		zs[i] = redis.Z{
			Score:  float64(member.Score),
			Member: member.Member,
		}
	}

	return c.rds.ZAdd(ctx, key, zs...).Err()
}

type ZRangeBy struct {
	Min, Max      string
	Offset, Count int64
}

type Z struct {
	Score  int64
	Member string
}
