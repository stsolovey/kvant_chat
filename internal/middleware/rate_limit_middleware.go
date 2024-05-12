package middleware

import (
	"net/http"

	"golang.org/x/time/rate"
)

func RateLimiterMiddleware(next http.Handler) http.Handler {
	limiter := rate.NewLimiter(10, 2)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)

			return
		}

		next.ServeHTTP(w, r)
	})
}
