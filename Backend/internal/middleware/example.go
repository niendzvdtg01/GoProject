package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// Example demonstrates how to use all middleware in MainRouting.go

func ExampleUsage(server *gin.Engine) {
	// Initialize middlewares
	authMiddleware := NewAuthMiddleware("your-secret-key")
	corsMiddleware := NewCORSMiddleware()
	loggingMiddleware := NewLoggingMiddleware()
	rateLimiter := NewRateLimiter(100, time.Minute)
	requestIDMiddleware := NewRequestIDMiddleware()

	// Global middlewares (apply to all routes)
	server.Use(requestIDMiddleware.RequestID())
	server.Use(loggingMiddleware.Logging())
	server.Use(corsMiddleware.CORS())
	server.Use(rateLimiter.RateLimit())

	// Protected routes group
	protected := server.Group("/api/protected")
	protected.Use(authMiddleware.AuthRequired())
	{
		// User routes - require authentication
		protected.GET("/users", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "List users"})
		})

		// Admin only routes
		admin := protected.Group("/admin")
		admin.Use(authMiddleware.RoleRequired("admin"))
		{
			admin.GET("/settings", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Admin settings"})
			})
		}
	}

	// Public routes (no auth required)
	public := server.Group("/api/public")
	{
		public.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
	}
}

// Example: Generate token for login
func ExampleGenerateToken() string {
	authMiddleware := NewAuthMiddleware("your-secret-key")

	token, err := authMiddleware.GenerateToken("user123", "john", "user")
	if err != nil {
		return ""
	}
	return token
}

// Example: Custom CORS configuration
func ExampleCustomCORS() gin.HandlerFunc {
	corsMiddleware := NewCORSMiddleware().
		SetAllowedOrigins([]string{"http://localhost:3000", "https://example.com"}).
		SetAllowedMethods([]string{"GET", "POST", "PUT", "DELETE"})

	return corsMiddleware.CORS()
}

// Example: Custom rate limiting per endpoint
func ExampleEndpointRateLimiter() gin.HandlerFunc {
	erl := NewEndpointRateLimiter()

	// Different limits for different endpoints
	erl.AddEndpoint("/api/login", 5, time.Minute)   // 5 requests/minute for login
	erl.AddEndpoint("/api/search", 30, time.Minute) // 30 requests/minute for search

	return erl.Middleware()
}
