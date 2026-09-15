package middleware

import (
	"net/http"

	"github.com/Alef-Daniel/go-rate-limiter/internal/ratelimiter"
)

func RateLimitMiddleware(limiter ratelimiter.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			req := ratelimiter.Request{
				IP:    r.RemoteAddr,
				Token: r.Header.Get("API_KEY"),
			}
			allowed, err := limiter.Allow(r.Context(), req)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			if !allowed {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte("you have reached the maximum number of requests or actions allowed within a certain time frame"))
				return
			}

			next.ServeHTTP(w, r)

		})
	}
}
