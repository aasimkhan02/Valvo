package metrics

import "sync/atomic"

var (
	AllowedTotal     uint64
	SoftDeniedTotal  uint64
	HardDeniedTotal  uint64
	RedisErrorsTotal uint64
)

func IncAllowed()     { atomic.AddUint64(&AllowedTotal, 1) }
func IncSoftDenied()  { atomic.AddUint64(&SoftDeniedTotal, 1) }
func IncHardDenied()  { atomic.AddUint64(&HardDeniedTotal, 1) }
func IncRedisErrors() { atomic.AddUint64(&RedisErrorsTotal, 1) }
