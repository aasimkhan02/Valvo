package admission

import "github.com/aasimkhan02/Valvo/internal"

type RateLimiter interface {
	Check(
		key internal.RateLimitKey,
		now int64,
	) RateLimitResult
}