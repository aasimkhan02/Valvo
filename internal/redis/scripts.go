// redis/scripts.go
package redis

import "github.com/redis/go-redis/v9"

// TokenBucketScript is a cached Redis Lua script.
// go-redis will:
//   1. LOAD it once (SCRIPT LOAD)
//   2. Use EVALSHA on every call
//   3. Fallback to EVAL automatically if Redis restarts
var TokenBucketScript = redis.NewScript(TokenBucketLua)

// TokenBucketLua implements an atomic token bucket rate limiter.
//
// KEYS[1] : rate limit key
// ARGV[1] : now (unix nanoseconds)
// ARGV[2] : capacity (max tokens)
// ARGV[3] : refill rate (tokens per second)
//
// Returns:
//   { allowed (0/1), remaining_tokens }
const TokenBucketLua = `
local key = KEYS[1]
local now = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local refill_rate = tonumber(ARGV[3])

-- Fetch current state
local data = redis.call("HMGET", key, "tokens", "last_refill")
local tokens = tonumber(data[1])
local last_refill = tonumber(data[2])

-- Initialize bucket
if tokens == nil or last_refill == nil then
    tokens = capacity
    last_refill = now
end

-- Refill tokens
local elapsed = now - last_refill
if elapsed > 0 then
    local refill = math.floor(elapsed * refill_rate / 1000000000)
    if refill > 0 then
        tokens = math.min(capacity, tokens + refill)
        last_refill = now
    end
end

-- Reject if empty
if tokens <= 0 then
    redis.call("HMSET", key,
        "tokens", tokens,
        "last_refill", last_refill
    )
    return {0, tokens}
end

-- Consume token
tokens = tokens - 1

redis.call("HMSET", key,
    "tokens", tokens,
    "last_refill", last_refill
)

return {1, tokens}
`
