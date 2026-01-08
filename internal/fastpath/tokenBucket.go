package fastpath

import "time"

type TokenBucket struct {
	capacity   int64
	tokens     int64
	refillRate int64 // tokens per second
	lastRefill int64 // unix nanos
}

func NewTokenBucket(capacity, refillRate int64, now int64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		lastRefill: now,
	}
}

func (b *TokenBucket) Allow(now int64) bool {
	elapsed := now - b.lastRefill
	if elapsed > 0 {
		refill := (elapsed * b.refillRate) / int64(time.Second)
		if refill > 0 {
			b.tokens += refill
			if b.tokens > b.capacity {
				b.tokens = b.capacity
			}
			b.lastRefill = now
		}
	}

	if b.tokens <= 0 {
		return false
	}

	b.tokens--
	return true
}

func (b *TokenBucket) RefillRate() int64 {
	return b.refillRate
}

func (b *TokenBucket) Tokens() int64 {
	return b.tokens
}
