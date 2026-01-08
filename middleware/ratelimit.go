package middleware

import (
	"string"
	"net/http"
	"time"

	"github.com/aasimkhan02/Valvo/internal"
)

type Limiter interface {
	Check (key.internal.RateLimitKey, now int64) (allowed bool)
}

func RateLimit(limiter Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			now := time.Now().UnixNano()

			key := internal.RateLimitKey{
				Tenant:   r.Header.Get("X-Tenant"),
				Region:   r.Header.Get("X-Region"),
				Resource: strings.Split(r.URL.Path, "/")[1],
			}

			if !limiter.Check(key, now) {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte("rate limit exceeded"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}