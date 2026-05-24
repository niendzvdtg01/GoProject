package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	redis "github.com/redis/go-redis/v9"
)

var ErrCacheMiss = errors.New("cache miss")

// scanBatch caps each SCAN page so a hot pattern can't pin the Redis main
// thread on one cursor step.
const scanBatch = 256

type redisCache struct {
	client *redis.Client
}

// NewRedisCache opens a Redis client and pings it once. The ping is the only
// place we surface a connection error; afterwards a flaky broker degrades to
// per-call errors that the caller treats as a miss.
func NewRedisCache(ctx context.Context, addr string) (Cache, error) {
	client := redis.NewClient(&redis.Options{Addr: addr})
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return &redisCache{client: client}, nil
}

func (r *redisCache) Get(ctx context.Context, key string, dest any) error {
	val, err := r.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return ErrCacheMiss
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(val, dest)
}

func (r *redisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	body, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, body, ttl).Err()
}

func (r *redisCache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return r.client.Del(ctx, keys...).Err()
}

// DeletePattern walks the keyspace with SCAN so we never block Redis with a
// KEYS sweep. Failed deletes are surfaced but the scan continues, so a
// transient error on one batch doesn't strand the rest of the matches.
func (r *redisCache) DeletePattern(ctx context.Context, pattern string) error {
	var cursor uint64
	for {
		keys, next, err := r.client.Scan(ctx, cursor, pattern, scanBatch).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		if next == 0 {
			return nil
		}
		cursor = next
	}
}
