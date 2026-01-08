package admission

import (
	"context"
	"fmt"
	"time"

	"github.com/aasimkhan02/Valvo/internal"
	"github.com/redis/go-redis/v9"
)

type DistributedLimiter struct {
	local *LocalLimiter
	redis *redis.Client

	capacity   int64
	refillRate int64
}

func NewDistributedLimiter(
	local *LocalLimiter,
	redis *redis.Client,
	capacity, refillRate int64,
) *DistributedLimiter {
	return &DistributedLimiter{
		local:      local,
		redis:     redis,
		capacity:  capacity,
		refillRate: refillRate,
	}
}

func redisKey(k internal.RateLimitKey) string {
	return fmt.Sprintf(
		"rl:%s:%s:%s",
		k.Tenant,
		k.Region,
		k.Resource,
	)
}

func (d *DistributedLimiter) Check(
	key internal.RateLimitKey,
	now int64,
) RateLimitResult {

	// 1️⃣ Local fast path
	localResult := d.local.Check(key, now)
	if localResult.Decision == SOFT_DENY {
		return localResult
	}

	// 2️⃣ Global authoritative check
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	res, err := d.redis.Eval(
		ctx,
		tokenBucketLua,
		[]string{redisKey(key)},
		now,
		d.capacity,
		d.refillRate,
	).Result()

	if err != nil {
		// Redis failure → degrade gracefully
		return RateLimitResult{
			Decision:        SOFT_DENY,
			RemainingTokens: localResult.RemainingTokens,
			RetryAfterMs:    100,
		}
	}

	values := res.([]interface{})
	allowed := values[0].(int64)
	remaining := values[1].(int64)

	if allowed == 0 {
		return RateLimitResult{
			Decision:        HARD_DENY,
			RemainingTokens: remaining,
			RetryAfterMs:    1000,
		}
	}

	return RateLimitResult{
		Decision:        ALLOW,
		RemainingTokens: remaining,
		RetryAfterMs:    0,
	}
}
