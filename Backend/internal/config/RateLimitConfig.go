package config

import (
	"backend/internal/middleware"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimitConfig struct{}

func NewRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{}
}

// RateLimitEnpoint applies IP-based limits per endpoint path.
// /api/auth/login and /api/users/register are throttled to prevent brute-force / mass registration.
// /api/users/import is capped at 3 per 10 minutes; pair with ImportUserRateLimit for per-user enforcement.
func (r *RateLimitConfig) RateLimitEnpoint() gin.HandlerFunc {
	erl := middleware.NewEndpointRateLimiter()
	erl.AddEndpoint("/api/auth/login", 10000, time.Minute)
	erl.AddEndpoint("/api/users/register", 10000, time.Minute)
	erl.AddEndpoint("/api/users/import", 5, 1*time.Minute)
	return erl.Middleware()
}

// ImportUserRateLimit limits import submissions per authenticated user (not per IP).
// Prevents a single account from saturating the worker pool even when requests arrive from many IPs.
func (r *RateLimitConfig) ImportUserRateLimit() gin.HandlerFunc {
	return middleware.NewUserRateLimiter(3, 10*time.Minute).RateLimit()
}
