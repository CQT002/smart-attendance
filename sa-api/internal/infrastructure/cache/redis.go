package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/hdbank/smart-attendance/config"
	"github.com/redis/go-redis/v9"
)

// Cache interface định nghĩa contract cho caching
type Cache interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Get(ctx context.Context, key string, dest any) error
	Delete(ctx context.Context, key string) error
	DeletePattern(ctx context.Context, pattern string) error
	Exists(ctx context.Context, key string) (bool, error)
	Incr(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
}

// RedisCache triển khai Cache interface sử dụng Redis
type RedisCache struct {
	client *redis.Client
}

// Key prefix constants - tránh key collision giữa các module
const (
	KeyPrefixUser       = "user:"
	KeyPrefixBranch     = "branch:"
	KeyPrefixAttend     = "attend:"
	KeyPrefixSession    = "session:"
	KeyPrefixRateLimit  = "rate:"
	KeyPrefixBlacklist  = "blacklist:"
	KeyPrefixDeviceLog  = "device:"
)

// NewRedisCache khởi tạo Redis client với connection pool
func NewRedisCache(cfg *config.RedisConfig) (Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
		// Timeout settings cho high-load scenario
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	slog.Info("redis connected", "host", cfg.Host, "port", cfg.Port)
	return &RedisCache{client: client}, nil
}

func (r *RedisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal error: %w", err)
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisCache) Get(ctx context.Context, key string, dest any) error {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return fmt.Errorf("cache get error: %w", err)
	}
	return json.Unmarshal(data, dest)
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("cache keys error: %w", err)
	}
	if len(keys) == 0 {
		return nil
	}
	return r.client.Del(ctx, keys...).Err()
}

func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	return count > 0, err
}

func (r *RedisCache) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

func (r *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return r.client.Expire(ctx, key, ttl).Err()
}

// ErrCacheMiss lỗi khi key không tồn tại trong cache
var ErrCacheMiss = fmt.Errorf("cache miss")

// IsCacheMiss kiểm tra có phải lỗi cache miss không
func IsCacheMiss(err error) bool {
	return err == ErrCacheMiss
}

// BuildKey tạo cache key theo chuẩn
func BuildKey(prefix, id string) string {
	return prefix + id
}
