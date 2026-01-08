package admission

import (
	"sync"
	"testing"
	"time"

	"github.com/aasimkhan02/Valvo/internal"
)

func TestLocalLimiter_BasicFlow(t *testing.T) {
	limiter := NewLocalLimiter(3, 1)
	key := internal.RateLimitKey{
		Tenant:   "t1",
		Region:   "us",
		Resource: "/test",
	}

	now := time.Now().UnixNano()

	// Allow burst
	for i := 0; i < 3; i++ {
		res := limiter.Check(key, now)
		if res.Decision != ALLOW {
			t.Fatalf("expected ALLOW at %d, got %v", i, res.Decision)
		}
	}

	// Next should soft deny
	res := limiter.Check(key, now)
	if res.Decision != SOFT_DENY {
		t.Fatalf("expected SOFT_DENY, got %v", res.Decision)
	}
}

func TestLocalLimiter_ConcurrentAccess(t *testing.T) {
	limiter := NewLocalLimiter(10, 5)
	key := internal.RateLimitKey{
		Tenant:   "t1",
		Region:   "us",
		Resource: "/concurrent",
	}

	now := time.Now().UnixNano()
	var wg sync.WaitGroup

	allowed := 0
	mu := sync.Mutex{}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res := limiter.Check(key, now)
			if res.Decision == ALLOW {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if allowed > 10 {
		t.Fatalf("allowed too many requests: %d", allowed)
	}
}
