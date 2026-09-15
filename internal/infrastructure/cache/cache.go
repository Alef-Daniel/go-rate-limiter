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
