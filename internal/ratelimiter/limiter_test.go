package ratelimiter_test

import (
	"context"
	"testing"
	"time"

	"github.com/Alef-Daniel/go-rate-limiter/internal/ratelimiter"
	"github.com/Alef-Daniel/go-rate-limiter/mocks"
	"github.com/stretchr/testify/mock"
)

func TestRedisLimiter_AllowsUntilIPLimit(t *testing.T) {
	cacheMock := mocks.NewCache(t)

	cacheMock.
		On(
			"Exists",
			mock.Anything,
			"rate_limit_192.168.0.1_blocked",
		).
		Return(false, nil).
		Times(4)

	cacheMock.
		On(
			"Increment",
			mock.Anything,
			"rate_limit_192.168.0.1",
		).
		Return(int64(1), nil).
		Once()

	cacheMock.
		On(
			"Expire",
			mock.Anything,
			"rate_limit_192.168.0.1",
			time.Second,
		).
		Return(nil).
		Once()

	cacheMock.
		On(
			"Increment",
			mock.Anything,
			"rate_limit_192.168.0.1",
		).
		Return(int64(2), nil).
		Once()

	cacheMock.
		On(
			"Increment",
			mock.Anything,
			"rate_limit_192.168.0.1",
		).
		Return(int64(3), nil).
		Once()

	cacheMock.
		On(
			"Increment",
			mock.Anything,
			"rate_limit_192.168.0.1",
		).
		Return(int64(4), nil).
		Once()

	cacheMock.
		On(
			"Set",
			mock.Anything,
			"rate_limit_192.168.0.1_blocked",
			"1",
			time.Minute,
		).
		Return(nil).
		Once()

	limiter, err := ratelimiter.NewRedisLimiter(
		cacheMock,
		time.Second,
		time.Minute,
		3,
		10,
	)
	if err != nil {
		t.Fatal(err)
	}

	req := ratelimiter.Request{
		IP: "192.168.0.1",
	}

	for i := 0; i < 3; i++ {
		allowed, err := limiter.Allow(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}

		if !allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}

	allowed, err := limiter.Allow(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	if allowed {
		t.Fatal("request exceeding the limit should be blocked")
	}
}
