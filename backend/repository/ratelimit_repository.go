package repository

import (
	"context"
	"errors"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type RateLimitStore struct {
	rdb *redis.Client
}

var hitScript = redis.NewScript(`
local current = redis.call("INCR", KEYS[1])
if current == 1 then
  redis.call("EXPIRE", KEYS[1], ARGV[1])
end

local ttl = redis.call("TTL", KEYS[1])
if ttl < 0 then
  ttl = tonumber(ARGV[1])
end

return {current, ttl}
`)

func NewRateLimitStore(rdb *redis.Client) *RateLimitStore {
	return &RateLimitStore{
		rdb: rdb,
	}
}

func (r *RateLimitStore) Hit(key string, windowSec int64) (int64, int64, error) {
	if r == nil || r.rdb == nil {
		return 0, 0, errors.New("redis client is nil")
	}
	if windowSec <= 0 {
		return 0, 0, errors.New("invalid window sec")
	}

	luaRes, err := hitScript.Run(
		context.Background(),
		r.rdb,
		[]string{key},
		windowSec,
	).Result()
	if err != nil {
		return 0, 0, err
	}

	xs, ok := luaRes.([]interface{})
	if !ok || len(xs) != 2 {
		return 0, 0, errors.New("invalid lua result")
	}

	count, err := toInt64(xs[0])
	if err != nil {
		return 0, 0, err
	}

	ttl, err := toInt64(xs[1])
	if err != nil {
		return 0, 0, err
	}

	return count, ttl, nil
}

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
