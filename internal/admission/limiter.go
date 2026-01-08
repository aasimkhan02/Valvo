package admission

import "github.com/aasimkhan02/Valvo/internal"

type RateLimiter interface {
	Decide ( 
		key internal.RateLimitKey,
		now int64
	) RateLimitResult
}