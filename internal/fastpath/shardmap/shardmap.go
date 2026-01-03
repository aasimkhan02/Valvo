package shardmap

import "sync"

type shard[V any] struct {
	mu sync.Mutex
	m  map[string]V
}

// ShardedMap is a fixed-size, lock-striped map
type ShardedMap[V any] struct {
	shards []shard[V]
	mask   uint64 // used for fast modulo
}

// New creates a sharded map with a fixed number of shards.
// numShards MUST be a power of two.
func New[V any](numShards int) *ShardedMap[V] {
	if numShards <= 0 {
		panic("numShards must be > 0")
	}

	// enforce power of two
	if numShards&(numShards-1) != 0 {
		panic("numShards must be power of 2")
	}

	sm := &ShardedMap[V]{
		shards: make([]shard[V], numShards),
		mask:   uint64(numShards - 1),
	}

	for i := range sm.shards {
		sm.shards[i].m = make(map[string]V)
	}

	return sm
}

// getShard returns the shard responsible for a key
func (sm *ShardedMap[V]) getShard(key string) *shard[V] {
	h := hashKey(key)
	return &sm.shards[h&sm.mask]
}

// Get retrieves a value
func (sm *ShardedMap[V]) Get(key string) (V, bool) {
	s := sm.getShard(key)
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.m[key]
	return v, ok
}

// Set inserts or updates a value
func (sm *ShardedMap[V]) Set(key string, val V) {
	s := sm.getShard(key)
	s.mu.Lock()
	s.m[key] = val
	s.mu.Unlock()
}

// Delete removes a key
func (sm *ShardedMap[V]) Delete(key string) {
	s := sm.getShard(key)
	s.mu.Lock()
	delete(s.m, key)
	s.mu.Unlock()
}

// Len returns approximate total size (not atomic)
func (sm *ShardedMap[V]) Len() int {
	total := 0
	for i := range sm.shards {
		s := &sm.shards[i]
		s.mu.Lock()
		total += len(s.m)
		s.mu.Unlock()
	}
	return total
}