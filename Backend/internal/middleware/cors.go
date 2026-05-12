package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type CORSMiddleware struct {
	allowedOrigins []string
	allowedMethods []string
	allowedHeaders []string
	exposeHeaders  []string
	maxAge         int
}

func NewCORSMiddleware() *CORSMiddleware {
	return &CORSMiddleware{
		allowedOrigins: []string{"*"},
		allowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		allowedHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		exposeHeaders:  []string{"Content-Length", "Content-Type"},
		maxAge:         86400,
	}
}

func (m *CORSMiddleware) SetAllowedOrigins(origins []string) *CORSMiddleware {
	m.allowedOrigins = origins
	return m
}

func (m *CORSMiddleware) SetAllowedMethods(methods []string) *CORSMiddleware {
	m.allowedMethods = methods
	return m
}

func (m *CORSMiddleware) SetAllowedHeaders(headers []string) *CORSMiddleware {
	m.allowedHeaders = headers
	return m
}

func (m *CORSMiddleware) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Check if origin is allowed
		allowed := false
		for _, o := range m.allowedOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}

		if allowed {
			if origin != "" && origin != "*" {
				c.Header("Access-Control-Allow-Origin", origin)
			} else {
				c.Header("Access-Control-Allow-Origin", "*")
			}
		}

		c.Header("Access-Control-Allow-Methods", joinStrings(m.allowedMethods))
		c.Header("Access-Control-Allow-Headers", joinStrings(m.allowedHeaders))
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", strconv.Itoa(m.maxAge))

		if len(m.exposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", joinStrings(m.exposeHeaders))
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func joinStrings(slice []string) string {
	result := ""
	for i, s := range slice {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}
