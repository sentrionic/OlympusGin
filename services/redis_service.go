package services

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/sentrionic/OlympusGin/database"
	"strconv"
	"time"
)

type RedisService interface {
	SetResetToken(ctx context.Context, id uint) (string, error)
	GetIdFromToken(ctx context.Context, token string) (uint, error)
}

type redisService struct {
	redis *redis.Client
}

func NewRedisService(conn database.RedisConnection) RedisService {
	return &redisService{redis: conn.Get()}
}

func (r *redisService) SetResetToken(ctx context.Context, id uint) (string, error) {
	uid, err := uuid.NewRandom()

	if err != nil {
		return "", err
	}

	if err := r.redis.Set(ctx, fmt.Sprintf("forgot-password:%s", uid.String()), id, 24*time.Hour).Err(); err != nil {
		fmt.Println(err)
		return "", err
	}

	return uid.String(), nil
}

func (r *redisService) GetIdFromToken(ctx context.Context, token string) (uint, error) {

	fmt.Println(token)
	key := fmt.Sprintf("forgot-password:%s", token)
	val, err := r.redis.Get(ctx, key).Result()

	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	id, _ := strconv.Atoi(val)

	r.redis.Del(ctx, key)

	return uint(id), nil
}
