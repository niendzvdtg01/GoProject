package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RequestIDMiddleware struct {
	headerName string
}

func NewRequestIDMiddleware() *RequestIDMiddleware {
	return &RequestIDMiddleware{
		headerName: "X-Request-ID",
	}
}

func (m *RequestIDMiddleware) SetHeaderName(name string) {
	m.headerName = name
}

func (m *RequestIDMiddleware) RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(m.headerName)

		// Generate new request ID if not provided
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set request ID in context
		c.Set("request_id", requestID)

		// Set response header
		c.Header(m.headerName, requestID)

		c.Next()
	}
}
