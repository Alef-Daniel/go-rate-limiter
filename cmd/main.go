package main

import (
	"log"
	"net/http"
	"time"

	httpInternal "github.com/Alef-Daniel/go-rate-limiter/internal/adapters/http"
	"github.com/Alef-Daniel/go-rate-limiter/internal/infrastructure/cache"
	"github.com/Alef-Daniel/go-rate-limiter/internal/ratelimiter"
)

func main() {
	clientRedis := cache.NewRedis("redis:6379")
	if clientRedis == nil {
		log.Fatal("redis client is nil")
	}

	limiter, err := ratelimiter.NewRedisLimiter(clientRedis, time.Minute*60, 3, 3)
	if err == nil {
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
