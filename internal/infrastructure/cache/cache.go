package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func NewRedis(addr string) *Redis {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	return &Redis{
		client: rdb,
	}
}

func (r *Redis) Increment(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

func (r *Redis) Expire(
	ctx context.Context,
	key string,
	expiration time.Duration,
) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

func (r *Redis) Exists(
	ctx context.Context,
	key string,
) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *Redis) Set(
	ctx context.Context,
	key string,
	value string,
	expiration time.Duration,
) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}
