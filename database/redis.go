package database

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/sentrionic/OlympusGin/config"
)

type RedisConnection interface {
	Get() *redis.Client
}

type redisConnection struct {
	Redis *redis.Client
}

func NewRedisConnection(c *config.Config) RedisConnection {
	cfg := c.Get()
	host := cfg.GetString("redis.host")
	port := cfg.GetInt("redis.port")

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: "",
		DB:       0,
	})

	_, err := rdb.Ping(context.Background()).Result()

	if err != nil {
		panic("redis connection failed")
	}

	return &redisConnection{Redis: rdb}
}

func (r *redisConnection) Get() *redis.Client {
	return r.Redis
}
