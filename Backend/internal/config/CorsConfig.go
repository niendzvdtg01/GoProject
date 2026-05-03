package config

import (
	"backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

type CorsConfig struct {
	corsMiddleware *middleware.CORSMiddleware
}

func NewCorsConfig() *CorsConfig {
	return &CorsConfig{
		corsMiddleware: middleware.NewCORSMiddleware(),
	}
}

func (c *CorsConfig) CustomCORS() gin.HandlerFunc {
	corsMiddleware := c.corsMiddleware.
		SetAllowedOrigins([]string{
			"http://localhost:5173",
			"https://example.com",
		}).
		SetAllowedMethods([]string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
		}).
		SetAllowedHeaders([]string{
			"Origin",
			"Content-Type",
			"Authorization",
		})

	return corsMiddleware.CORS()
}
