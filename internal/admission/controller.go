package admission

import (
	"time"

	"Valvo/internal"
	"Valvo/internal/fastpath"
)

type Policy struct {
	MaxRequests int64
	Window      time.Duration

	BucketCapacity int64
	RefillRate     int64
}

type Controller struct {
	windows *fastpath.ShardedMap[*fastpath.SlidingWindow]
	buckets *fastpath.ShardedMap[*fastpath.TokenBucket]
}

func NewController() *Controller {
	return &Controller{
		windows: fastpath.NewShardedMap,
		buckets: fastpath.NewShardedMap,
	}
}


func (c *Controller) Check(
	key internal.RateLimitKey,
	p Policy,
	now int64,
) Result {

	// Sliding window
	win, ok := c.windows.Get(key)
	if !ok {
		win = fastpath.NewSlidingWindow(p.Window, 10, now)
		c.windows.Set(key, win)
	}

	if win.Count(now) >= p.MaxRequests {
		return Result{
			Decision: HARD_DENY,
			Remaining: 0,
		}
	}

	// Token bucket
	bucket, ok := c.buckets.Get(key)
	if !ok {
		bucket = fastpath.NewTokenBucket(
			p.BucketCapacity,
			p.RefillRate,
			now,
		)
		c.buckets.Set(key, bucket)
	}

	if !bucket.Allow(now) {
		return Result{
			Decision: SOFT_DENY,
			Remaining: bucket.Tokens(),
		}
	}

	win.Add(now)

	return Result{
		Decision: ALLOW,
		Remaining: bucket.Tokens(),
	}
}