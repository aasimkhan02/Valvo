package redis

// TokenBucketLua is an atomic Redis Lua script implementing
// a global token bucket rate limiter.
//
// KEYS[1] : rate limit key
// ARGV[1] : now (unix nanoseconds)
// ARGV[2] : capacity (max tokens)
// ARGV[3] : refill rate (tokens per second)
//
// Returns:
//   { allowed (0/1), remaining_tokens }
const TokenBucketLua = `
-- Atomic token bucket rate limiter

local key = KEYS[1]
local now = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local refill_rate = tonumber(ARGV[3])

-- Fetch current state
local data = redis.call("HMGET", key, "tokens", "last_refill")

local tokens = tonumber(data[1])
local last_refill = tonumber(data[2])

-- Initialize if key does not exist
if tokens == nil or last_refill == nil then
    tokens = capacity
    last_refill = now
end

-- Refill tokens based on elapsed time
local elapsed = now - last_refill
if elapsed > 0 then
    local refill = math.floor(elapsed * refill_rate / 1000000000)
    if refill > 0 then
        tokens = math.min(capacity, tokens + refill)
        last_refill = now
    end
end

-- Deny if no tokens left
if tokens <= 0 then
    redis.call("HMSET", key,
        "tokens", tokens,
        "last_refill", last_refill
    )
    return {0, tokens}
end

-- Consume one token
tokens = tokens - 1

redis.call("HMSET", key,
    "tokens", tokens,
    "last_refill", last_refill
)

return {1, tokens}
`
