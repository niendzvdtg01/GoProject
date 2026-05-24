package config

import "os"

const defaultRedisAddr = "localhost:6379"

type CacheConfig struct {
	Addr string
}

func NewCacheConfig() *CacheConfig {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = defaultRedisAddr
	}
	return &CacheConfig{Addr: addr}
}
