package shardmap

func hashKey(key string) uint64 {
	const (
		offset64 = 14695981039346656037
		prime64  = 1099511628211
	)

	var h uint64 = offset64
	for i := 0; i < len(key); i++ {
		h ^= uint64(key[i])
		h *= prime64
	}
	return h
}