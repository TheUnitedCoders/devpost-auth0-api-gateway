package ratelimit

import (
	"context"
	"fmt"
	"time"
)

// Key ...
type Key struct {
	Service          string
	IsServiceLimiter bool
	Method           string
	Entity           string
}

func (k Key) string() string {
	if k.IsServiceLimiter {
		return fmt.Sprintf("lim_%s:%s", k.Service, k.Entity)
	}

	return fmt.Sprintf("lim_%s:%s:%s", k.Service, k.Method, k.Entity)
}

// Limit ...
type Limit struct {
	Rate   uint64
	Burst  uint64
	Period time.Duration
}

// Limiter ...
type Limiter interface {
	Allow(ctx context.Context, key Key, limit Limit) (allowed bool, retryAfter time.Duration, err error)
}
