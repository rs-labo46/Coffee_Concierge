package db

import (
	"context"
	"fmt"
	"time"

	"coffee-spa/config"

	"github.com/redis/go-redis/v9"
)

func NewRedis(c config.Cfg) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort),
		Password:     c.RedisPassword,
		DB:           c.RedisDB,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  250 * time.Millisecond,
		WriteTimeout: 250 * time.Millisecond,
		PoolSize:     50,
		MinIdleConns: 5,
	})
}

func PingRedis(ctx context.Context, rdb *redis.Client) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return rdb.Ping(ctx).Err()
}
