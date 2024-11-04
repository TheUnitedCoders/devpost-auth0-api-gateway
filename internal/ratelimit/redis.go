package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

type redisLimiter struct {
	limiter *redis_rate.Limiter
}

// NewRedis returns new Limiter.
func NewRedis(client *redis.Client) Limiter {
	return &redisLimiter{
		limiter: redis_rate.NewLimiter(client),
	}
}

func (r *redisLimiter) Allow(ctx context.Context, key Key, limit Limit) (bool, time.Duration, error) {
	res, err := r.limiter.Allow(ctx, key.string(), redis_rate.Limit{
		Rate:   int(limit.Rate),
		Burst:  int(limit.Burst),
		Period: limit.Period,
	})
	if err != nil {
		return false, time.Duration(0), fmt.Errorf("redis limiter: %w", err)
	}

	return res.RetryAfter == -1, res.RetryAfter, nil
}
