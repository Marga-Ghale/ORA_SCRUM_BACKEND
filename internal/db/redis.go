// internal/db/redis.go
package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisDB struct {
	Client *redis.Client
}

func NewRedisDB(redisURL string) (*RedisDB, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("[Redis] ✅ Connected to Redis")
	return &RedisDB{Client: client}, nil
}

func (r *RedisDB) Close() {
	if r.Client != nil {
		r.Client.Close()
		log.Println("[Redis] Connection closed")
	}
}

// Session management
func (r *RedisDB) SetSession(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.Client.Set(ctx, "session:"+key, data, expiration).Err()
}

func (r *RedisDB) GetSession(ctx context.Context, key string, dest interface{}) error {
	data, err := r.Client.Get(ctx, "session:"+key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (r *RedisDB) DeleteSession(ctx context.Context, key string) error {
	return r.Client.Del(ctx, "session:"+key).Err()
}

// Cache methods
func (r *RedisDB) SetCache(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.Client.Set(ctx, "cache:"+key, data, expiration).Err()
}

func (r *RedisDB) GetCache(ctx context.Context, key string, dest interface{}) error {
	data, err := r.Client.Get(ctx, "cache:"+key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (r *RedisDB) InvalidateCache(ctx context.Context, pattern string) error {
	keys, err := r.Client.Keys(ctx, "cache:"+pattern).Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return r.Client.Del(ctx, keys...).Err()
	}
	return nil
}
