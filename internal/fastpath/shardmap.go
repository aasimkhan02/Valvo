package fastpath

import (
	"sync"
	"Valvo/internal"
)

type shard[V any] struct {
	mu sync.Mutex
	m  map[internal.RateLimitKey]V
}

type ShardedMap[V any] struct {
	shards []shard[V]
	mask   uint64
}

func NewShardedMap[V any](numShards int) *ShardedMap[V] {
	if numShards&(numShards-1) != 0 {
		panic("numShards must be power of 2")
	}

	sm := &ShardedMap[V]{
		shards: make([]shard[V], numShards),
		mask:   uint64(numShards - 1),
	}

	for i := range sm.shards {
		sm.shards[i].m = make(map[internal.RateLimitKey]V)
	}

	return sm
}

func (sm *ShardedMap[V]) getShard(k internal.RateLimitKey) *shard[V] {
	return &sm.shards[hashKey(k)&sm.mask]
}

func (sm *ShardedMap[V]) Get(k internal.RateLimitKey) (V, bool) {
	s := sm.getShard(k)
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.m[k]
	return v, ok
}

func (sm *ShardedMap[V]) Set(k internal.RateLimitKey, v V) {
	s := sm.getShard(k)
	s.mu.Lock()
	s.m[k] = v
	s.mu.Unlock()
}