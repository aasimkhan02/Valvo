-- KEYS[1] = rate limit key
-- ARGV[1] = now (unix nanos)
-- ARGV[2] = capacity
-- ARGV[3] = refill_rate (tokens per second)

local key = KEYS[1]
local now = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local refill_rate = tonumber(ARGV[3])

local data = redis.call("HMGET", key, "tokens", "last_refill")

local tokens = tonumber(data[1])
local last_refill = tonumber(data[2])

if tokens == nil then
    tokens = capacity
    last_refill = now
end

local elapsed = now - last_refill
if elapsed > 0 then
    local refill = math.floor(elapsed * refill_rate / 1000000000)
    if refill > 0 then
        tokens = math.min(capacity, tokens + refill)
        last_refill = now
    end
end

if tokens <= 0 then
    redis.call("HMSET", key,
        "tokens", tokens,
        "last_refill", last_refill
    )
    return {0, tokens}
end

tokens = tokens - 1

redis.call("HMSET", key,
    "tokens", tokens,
    "last_refill", last_refill
)

local ttl_seconds = 60

if tokens <= 0 then
    redis.call("HMSET", key,
        "tokens", tokens,
        "last_refill", last_refill
    )
    redis.call("EXPIRE", key, ttl_seconds)
    return {0, tokens}
end

tokens = tokens - 1

redis.call("HMSET", key,
    "tokens", tokens,
    "last_refill", last_refill
)
redis.call("EXPIRE", key, ttl_seconds)

return {1, tokens}
