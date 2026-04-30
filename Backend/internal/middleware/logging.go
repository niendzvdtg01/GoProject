package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type LoggingMiddleware struct {
	logger func(format string, args ...interface{})
}

type LogEntry struct {
	Time      time.Time `json:"time"`
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	IP        string    `json:"ip"`
	Status    int       `json:"status"`
	Latency   int64     `json:"latency_ms"`
	UserAgent string    `json:"user_agent"`
	RequestID string    `json:"request_id"`
}

func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: func(format string, args ...interface{}) {
			fmt.Printf(format+"\n", args...)
		},
	}
}

func (m *LoggingMiddleware) SetLogger(logger func(format string, args ...interface{})) {
	m.logger = logger
}

func (m *LoggingMiddleware) Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Get response details
		latency := time.Since(start).Milliseconds()
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		requestID := c.GetString("request_id")

		// Log the request
		if query != "" {
			path = path + "?" + query
		}

		logEntry := LogEntry{
			Time:      time.Now(),
			Method:    c.Request.Method,
			Path:      path,
			IP:        clientIP,
			Status:    status,
			Latency:   latency,
			UserAgent: userAgent,
			RequestID: requestID,
		}

		m.logger("[%s] %s %s %d %dms %s",
			logEntry.Time.Format("2006-01-02 15:04:05"),
			logEntry.Method,
			logEntry.Path,
			logEntry.Status,
			logEntry.Latency,
			logEntry.IP,
		)
	}
}

func (m *LoggingMiddleware) StructuredLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start).Milliseconds()
		status := c.Writer.Status()
		clientIP := c.ClientIP()

		entry := LogEntry{
			Time:      time.Now(),
			Method:    c.Request.Method,
			Path:      path,
			IP:        clientIP,
			Status:    status,
			Latency:   latency,
			UserAgent: c.Request.UserAgent(),
		}

		// Print structured log
		fmt.Printf("%+v\n", entry)
	}
}
