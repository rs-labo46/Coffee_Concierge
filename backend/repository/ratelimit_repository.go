package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimitStore struct {
	rdb *redis.Client
}

var allowScript = redis.NewScript(`
local key = KEYS[1]

local rate = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local cost = tonumber(ARGV[3])
local now = tonumber(ARGV[4])
local ttl = tonumber(ARGV[5])

if not rate or rate <= 0 then
	return redis.error_reply("invalid rate")
end
if not capacity or capacity <= 0 then
	return redis.error_reply("invalid capacity")
end
if not cost or cost <= 0 then
	return redis.error_reply("invalid cost")
end
if not now or now < 0 then
	return redis.error_reply("invalid now")
end
if not ttl or ttl <= 0 then
	return redis.error_reply("invalid ttl")
end

local vals = redis.call("HMGET", key, "tokens", "ts")
local tokens = tonumber(vals[1])
local ts = tonumber(vals[2])

if not tokens then
	tokens = capacity
end
if not ts then
	ts = now
end

local elapsed = now - ts
if elapsed < 0 then
	elapsed = 0
end

tokens = math.min(capacity, tokens + (elapsed * rate))

local allowed = 0
local retry_after = 0

if tokens >= cost then
	tokens = tokens - cost
	allowed = 1
else
	local deficit = cost - tokens
	retry_after = math.ceil(deficit / rate)
end

redis.call("HSET", key, "tokens", tostring(tokens), "ts", tostring(now))
redis.call("EXPIRE", key, ttl)

local remaining = math.floor(tokens)
if remaining < 0 then
	remaining = 0
end

return {allowed, remaining, retry_after}
`)

// RateLimitStoreを生成。
func NewRateLimitStore(rdb *redis.Client) *RateLimitStore {
	return &RateLimitStore{
		rdb: rdb,
	}
}

// AllowのtokenBucket方式
func (r *RateLimitStore) Allow(
	key string,
	rate float64,
	capacity float64,
	cost float64,
	now time.Time,
) (bool, int, error) {
	if r == nil || r.rdb == nil {
		return false, 0, errors.New("redis client is nil")
	}
	if key == "" {
		return false, 0, errors.New("key is empty")
	}
	if rate <= 0 {
		return false, 0, errors.New("invalid rate")
	}
	if capacity <= 0 {
		return false, 0, errors.New("invalid capacity")
	}
	if cost <= 0 {
		return false, 0, errors.New("invalid cost")
	}

	drainSec := int64(math.Ceil(capacity / rate))
	if drainSec <= 0 {
		drainSec = 1
	}
	ttlSec := drainSec * 2
	if ttlSec <= 0 {
		ttlSec = 1
	}
	luaRes, err := allowScript.Run(
		context.Background(),
		r.rdb,
		[]string{key},
		fmt.Sprintf("%.6f", rate),
		fmt.Sprintf("%.6f", capacity),
		fmt.Sprintf("%.6f", cost),
		now.Unix(),
		ttlSec,
	).Result()
	if err != nil {
		return false, 0, err
	}

	xs, ok := luaRes.([]interface{})
	if !ok || len(xs) != 3 {
		return false, 0, errors.New("invalid lua result")
	}

	allowedInt, err := toInt64(xs[0])
	if err != nil {
		return false, 0, err
	}
	retryAfterSec, err := toInt(xs[2])
	if err != nil {
		return false, 0, err
	}

	return allowedInt == 1, retryAfterSec, nil
}

// toInt64はLuaから返る値をint64に変換。
func toInt64(v interface{}) (int64, error) {
	switch x := v.(type) {
	case int64:
		return x, nil
	case string:
		return strconv.ParseInt(x, 10, 64)
	default:
		return 0, errors.New("invalid lua value")
	}
}

// Luaの戻り値をintに変換
func toInt(v interface{}) (int, error) {
	switch x := v.(type) {
	case int:
		return x, nil
	case int64:
		return int(x), nil
	case float64:
		return int(x), nil
	case string:
		n, err := strconv.Atoi(x)
		if err != nil {
			return 0, err
		}
		return n, nil
	default:
		return 0, errors.New("invalid lua value")
	}
}
