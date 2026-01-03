package fastpath

import "hash/fnv"
import "Valvo/internal"

func hashKey(k internal.RateLimitKey) uint64 {
	h := fnv.New64a()
	h.Write([]byte(k.Tenant))
	h.Write([]byte{0})
	h.Write([]byte(k.Region))
	h.Write([]byte{0})
	h.Write([]byte(k.Resource))
	return h.Sum64()
}