package http

import (
	"net/http"

	"github.com/Alef-Daniel/go-rate-limiter/internal/adapters/http/middleware"
	"github.com/Alef-Daniel/go-rate-limiter/internal/ratelimiter"
	"github.com/go-chi/chi/v5"
)

func NewRoutes(
	limiter ratelimiter.Limiter,
	handler *Handler,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RateLimitMiddleware(limiter))

	r.Get("/", handler.Hello)

	return r
}
