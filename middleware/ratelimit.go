package middleware

import (
	"golang.org/x/time/rate"
	"net/http"
)

var Limiter = rate.NewLimiter(1, 3)

type MiddlewareFunc func(http.Handler) http.Handler

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !Limiter.Allow() {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
