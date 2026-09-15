package ratelimiter

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Limiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}

type Cache interface {
	Increment(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
}

type RedisLimiter struct {
	client Cache
	window time.Duration
	limit  int64
}

func NewRedisLimiter(client Cache, window time.Duration, limit int64) (*RedisLimiter, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}

	if window <= 0 {
		return nil, errors.New("window must be greater than zero")
	}

	if limit <= 0 {
		return nil, errors.New("limit must be greater than zero")
	}

	return &RedisLimiter{
		client: client,
		window: window,
		limit:  limit,
	}, nil

}

func (r *RedisLimiter) Allow(ctx context.Context, key string) (bool, error) {

	keyWithPrefix, err := addPrefixKey(key)
	if err != nil {
		return false, err
	}
	count, err := r.client.Increment(ctx, keyWithPrefix)
	if err != nil {
		return false, err
	}

	if count == 1 {
		err := r.client.Expire(ctx, keyWithPrefix, r.window)
		if err != nil {
			return false, err
		}
	}

	return count <= r.limit, nil
}

func addPrefixKey(key string) (string, error) {
	if key == "" {
		return "", errors.New("key is empty")
	}

	return fmt.Sprintf("%s_%s", "rate_limit", key), nil
}
