package admission

import (
	"sync"
	"time"

	"github.com/aasimkhan02/Valvo/internal"
	"github.com/aasimkhan02/Valvo/internal/fastpath"
)

type LocalLimiter struct {
	mu sync.Mutex

	buckets map[internal.RateLimitKey]*fastpath.TokenBucket

	capacity   int64
	refillRate int64
}

func NewLocalLimiter(capacity, refillRate int64) *LocalLimiter {
	return &LocalLimiter{
		buckets:    make(map[internal.RateLimitKey]*fastpath.TokenBucket),
		capacity:   capacity,
		refillRate: refillRate,
	}
}

func (l *LocalLimiter) Check(
	key internal.RateLimitKey,
	now int64,
) RateLimitResult {

	l.mu.Lock()
	defer l.mu.Unlock()

	bucket, ok := l.buckets[key]
	if !ok {
		bucket = fastpath.NewTokenBucket(
			l.capacity,
			l.refillRate,
			now,
		)
		l.buckets[key] = bucket
	}

	if !bucket.Allow(now) {
		return RateLimitResult{
			Decision:        SOFT_DENY,
			RemainingTokens: 0,
			RetryAfterMs:    estimateRetry(bucket),
		}
	}

	return RateLimitResult{
		Decision:        ALLOW,
		RemainingTokens: bucket.Tokens(),
		RetryAfterMs:    0,
	}
}

func estimateRetry(b *fastpath.TokenBucket) int64 {
	if b == nil || b.RefillRate() <= 0 {
		return 0
	}
	return int64(time.Second/time.Millisecond) / b.RefillRate()
}
