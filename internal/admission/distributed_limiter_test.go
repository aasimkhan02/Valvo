package admission

import (
	"context"
	"testing"
	"time"

	"github.com/aasimkhan02/Valvo/internal"
	"github.com/redis/go-redis/v9"
)

func TestDistributedLimiter_GlobalEnforcement(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx := context.Background()
	rdb.FlushDB(ctx)

	local := NewLocalLimiter(10, 10)

	limiter := NewDistributedLimiter(
		local,
		rdb,
		5, // global capacity
		1, // global refill/sec
	)

	key := internal.RateLimitKey{
		Tenant:   "tenant",
		Region:   "us",
		Resource: "/payments",
	}

	now := time.Now().UnixNano()

	// Allow up to global capacity
	for i := 0; i < 5; i++ {
		res := limiter.Check(key, now)
		if res.Decision != ALLOW {
			t.Fatalf("expected ALLOW at %d, got %v", i, res.Decision)
		}
	}

	// Next should HARD_DENY
	res := limiter.Check(key, now)
	if res.Decision != HARD_DENY {
		t.Fatalf("expected HARD_DENY, got %v", res.Decision)
	}
}

func TestDistributedLimiter_RedisDown(t *testing.T) {
	// invalid Redis address
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6399",
	})

	local := NewLocalLimiter(2, 1)

	limiter := NewDistributedLimiter(
		local,
		rdb,
		1,
		1,
	)

	key := internal.RateLimitKey{
		Tenant:   "t",
		Region:   "us",
		Resource: "/failover",
	}

	now := time.Now().UnixNano()

	res := limiter.Check(key, now)

	if res.Decision != SOFT_DENY && res.Decision != ALLOW {
		t.Fatalf("unexpected decision during Redis failure: %v", res.Decision)
	}
}
