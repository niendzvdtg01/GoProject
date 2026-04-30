package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

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
	rl := &RateLimiter{
		requests: make(map[string]*clientLimit),
		limit:    requestsPerWindow,
		window:   window,
	}

	// Cleanup old entries every minute
	go rl.cleanup()

	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
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
		key := c.ClientIP()
		now := time.Now()

		rl.mu.Lock()

		cl, exists := rl.requests[key]
		if !exists {
			rl.requests[key] = &clientLimit{
				count:     1,
				resetTime: now.Add(rl.window),
			}
			rl.mu.Unlock()
			c.Next()
			return
		}

		// Check if window has reset
		if now.After(cl.resetTime) {
			cl.count = 1
			cl.resetTime = now.Add(rl.window)
			rl.mu.Unlock()
			c.Next()
			return
		}

		// Check if limit exceeded
		if cl.count >= rl.limit {
			rl.mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests",
				"retry_after": cl.resetTime.Sub(now).Seconds(),
			})
			c.Abort()
			return
		}

		cl.count++
		rl.mu.Unlock()
		c.Next()
	}
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
		path := c.FullPath()

		erl.mu.RLock()
		limiter, exists := erl.limiters[path]
		erl.mu.RUnlock()

		if !exists {
			c.Next()
			return
		}

		key := c.ClientIP()
		now := time.Now()

		limiter.mu.Lock()

		cl, exists := limiter.requests[key]
		if !exists {
			limiter.requests[key] = &clientLimit{
				count:     1,
				resetTime: now.Add(limiter.window),
			}
			limiter.mu.Unlock()
			c.Next()
			return
		}

		if now.After(cl.resetTime) {
			cl.count = 1
			cl.resetTime = now.Add(limiter.window)
			limiter.mu.Unlock()
			c.Next()
			return
		}

		if cl.count >= limiter.limit {
			limiter.mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests for this endpoint",
			})
			c.Abort()
			return
		}

		cl.count++
		limiter.mu.Unlock()
		c.Next()
	}
}
