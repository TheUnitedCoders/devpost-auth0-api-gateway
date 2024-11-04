package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// New returns new redis.Client with ping.
func New(ctx context.Context, address, password string) (*redis.Client, error) {
	c := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
	})

	pingCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	pingRes := c.Ping(pingCtx)
	if err := pingRes.Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return c, nil
}
