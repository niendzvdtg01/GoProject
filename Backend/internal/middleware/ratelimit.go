package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const defaultCleanupInterval = time.Minute

type RateLimiter struct {
	requests map[string]*clientLimit
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

type clientLimit struct {
	count     int
	resetTime time.Time
}

func NewRateLimiter(requestsPerWindow int, window time.Duration) *RateLimiter {
	limiter := &RateLimiter{
		requests: make(map[string]*clientLimit),
		limit:    requestsPerWindow,
		window:   window,
	}

	go limiter.cleanup(defaultCleanupInterval)

	return limiter
}

func (rl *RateLimiter) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, cl := range rl.requests {
			if now.After(cl.resetTime) {
				delete(rl.requests, key)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		retryAfter, allowed := rl.allow(c.ClientIP())
		if allowed {
			c.Next()
			return
		}

		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":       "too many requests",
			"retry_after": retryAfter.Seconds(),
		})
		c.Abort()
	}
}

func (rl *RateLimiter) allow(key string) (time.Duration, bool) {
	now := time.Now()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	limit, exists := rl.requests[key]
	if !exists || now.After(limit.resetTime) {
		rl.requests[key] = &clientLimit{
			count:     1,
			resetTime: now.Add(rl.window),
		}
		return 0, true
	}

	if limit.count >= rl.limit {
		return time.Until(limit.resetTime), false
	}

	limit.count++
	return 0, true
}

// IP-based rate limiter with different limits per endpoint
type EndpointRateLimiter struct {
	limiters map[string]*RateLimiter
	mu       sync.RWMutex
}

func NewEndpointRateLimiter() *EndpointRateLimiter {
	return &EndpointRateLimiter{
		limiters: make(map[string]*RateLimiter),
	}
}

func (erl *EndpointRateLimiter) AddEndpoint(path string, requests int, window time.Duration) {
	erl.mu.Lock()
	defer erl.mu.Unlock()
	erl.limiters[path] = NewRateLimiter(requests, window)
}

func (erl *EndpointRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		erl.mu.RLock()
		limiter, exists := erl.limiters[c.FullPath()]
		erl.mu.RUnlock()

		if !exists {
			c.Next()
			return
		}

		retryAfter, allowed := limiter.allow(c.ClientIP())
		if allowed {
			c.Next()
			return
		}

		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":       "too many requests for this endpoint",
			"retry_after": retryAfter.Seconds(),
		})
		c.Abort()
	}
}
