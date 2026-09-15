package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	httpInternal "github.com/Alef-Daniel/go-rate-limiter/internal/adapters/http"
	"github.com/Alef-Daniel/go-rate-limiter/internal/infrastructure/cache"
	"github.com/Alef-Daniel/go-rate-limiter/internal/ratelimiter"
)

func main() {

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	window, err := time.ParseDuration(os.Getenv("RATE_LIMIT_WINDOW"))
	if err != nil {
		log.Fatal("invalid RATE_LIMIT_WINDOW:", err)
	}

	blockTime, err := time.ParseDuration(os.Getenv("RATE_LIMIT_BLOCK_TIME"))
	if err != nil {
		log.Fatal("invalid RATE_LIMIT_BLOCK_TIME:", err)
	}

	ipLimit, err := strconv.ParseInt(os.Getenv("RATE_LIMIT_IP"), 10, 64)
	if err != nil {
		log.Fatal("invalid RATE_LIMIT_IP:", err)
	}

	tokenLimit, err := strconv.ParseInt(os.Getenv("RATE_LIMIT_TOKEN"), 10, 64)
	if err != nil {
		log.Fatal("invalid RATE_LIMIT_TOKEN:", err)
	}
	clientRedis := cache.NewRedis(redisAddr)
	if clientRedis == nil {
		log.Fatal("redis client is nil")
	}

	limiter, err := ratelimiter.NewRedisLimiter(
		clientRedis,
		window,
		blockTime,
		ipLimit,
		tokenLimit,
	)
	if err != nil {
		log.Fatal(err)
	}

	handler := httpInternal.NewHandler()

	rt := httpInternal.NewRoutes(limiter, handler)
	if rt == nil {
		log.Fatal("routes is nil")
	}

	server := &http.Server{
		Addr:    ":8080",
		Handler: rt,
	}

	log.Println("server running on :8080")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
