package fastpath

import (
	"testing"
	"time"
)

func TestTokenBucket_BurstAndRefill(t *testing.T) {
	now := time.Now().UnixNano()

	b := NewTokenBucket(
		5,  // capacity
		1,  // refill: 1 token/sec
		now,
	)

	// Burst: allow 5
	for i := 0; i < 5; i++ {
		if !b.Allow(now) {
			t.Fatalf("expected allow at burst %d", i)
		}
	}

	// 6th should deny
	if b.Allow(now) {
		t.Fatal("expected deny when bucket empty")
	}

	// Advance time by 2 seconds
	later := now + int64(2*time.Second)

	if !b.Allow(later) {
		t.Fatal("expected allow after refill")
	}

	if b.Tokens() != 1 {
	t.Fatalf("expected 1 token left, got %d", b.Tokens())
	}

}
