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

// 1. KEYS[1]の値をINCR
// 2. 初回作成時だけEXPIREを設定
// 3. 現在のTTLを取得
// 4. TTLが不正値ならwindowSec
// 5. {current, ttl}を返す
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

// RateLimitStoreを生成。
func NewRateLimitStore(rdb *redis.Client) *RateLimitStore {
	return &RateLimitStore{
		rdb: rdb,
	}
}

// keyに対して1回分のアクセスを記録。
// count: 現在の窓内hit回数
// ttl: 窓が切れるまでの残り秒数
func (r *RateLimitStore) Hit(key string, windowSec int64) (int64, int64, error) {
	if r == nil || r.rdb == nil {
		return 0, 0, errors.New("redis client is nil")
	}

	// 窓秒数は正の値。
	if windowSec <= 0 {
		return 0, 0, errors.New("invalid window sec")
	}

	// LuaスクリプトをRedis上で実行。
	// KEYS[1]にrate limit用キー、ARGV[1]にwindow 秒数を渡す。
	luaRes, err := hitScript.Run(
		context.Background(),
		r.rdb,
		[]string{key},
		windowSec,
	).Result()
	if err != nil {
		return 0, 0, err
	}

	// 戻り値は[count, ttl]の2要素配列。
	xs, ok := luaRes.([]interface{})
	if !ok || len(xs) != 2 {
		return 0, 0, errors.New("invalid lua result")
	}

	// 1要素目をcountとしてint64に変換。
	count, err := toInt64(xs[0])
	if err != nil {
		return 0, 0, err
	}

	// 2要素目をttlとしてint64に変換。
	ttl, err := toInt64(xs[1])
	if err != nil {
		return 0, 0, err
	}

	return count, ttl, nil
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
