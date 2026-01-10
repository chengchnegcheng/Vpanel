// Package middleware provides HTTP middleware for the API.
package middleware

import (
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"v/internal/logger"
)

// Recovery returns a middleware that recovers from panics.
func Recovery(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := string(debug.Stack())
				log.Error("panic recovered",
					logger.F("error", err),
					logger.F("stack", stack),
					logger.F("path", c.Request.URL.Path),
					logger.F("method", c.Request.Method),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})
			}
		}()
		c.Next()
	}
}

// Logger returns a middleware that logs requests.
func Logger(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		fields := []logger.Field{
			logger.F("status", status),
			logger.F("method", c.Request.Method),
			logger.F("path", path),
			logger.F("latency", latency.String()),
			logger.F("ip", c.ClientIP()),
			logger.F("user_agent", c.Request.UserAgent()),
		}

		if query != "" {
			fields = append(fields, logger.F("query", query))
		}

		if requestID := c.GetString("request_id"); requestID != "" {
			fields = append(fields, logger.F("request_id", requestID))
		}

		if len(c.Errors) > 0 {
			fields = append(fields, logger.F("errors", c.Errors.String()))
		}

		if status >= 500 {
			log.Error("request completed", fields...)
		} else if status >= 400 {
			log.Warn("request completed", fields...)
		} else {
			log.Info("request completed", fields...)
		}
	}
}

// CORS returns a middleware that handles CORS.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		for _, o := range allowedOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if len(allowedOrigins) > 0 {
			c.Header("Access-Control-Allow-Origin", allowedOrigins[0])
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RequestID returns a middleware that adds a request ID to the context.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// RateLimit returns a simple rate limiting middleware.
func RateLimit(requestsPerSecond int) gin.HandlerFunc {
	// Simple token bucket implementation
	type client struct {
		tokens    float64
		lastCheck time.Time
	}
	clients := make(map[string]*client)
	rate := float64(requestsPerSecond)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		cl, exists := clients[ip]
		if !exists {
			cl = &client{tokens: rate, lastCheck: now}
			clients[ip] = cl
		}

		// Refill tokens
		elapsed := now.Sub(cl.lastCheck).Seconds()
		cl.tokens += elapsed * rate
		if cl.tokens > rate {
			cl.tokens = rate
		}
		cl.lastCheck = now

		if cl.tokens < 1 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			return
		}

		cl.tokens--
		c.Next()
	}
}

// SecureHeaders returns a middleware that adds security headers.
func SecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// ContentType returns a middleware that validates content type.
func ContentType(contentTypes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "DELETE" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		ct := c.ContentType()
		for _, allowed := range contentTypes {
			if strings.HasPrefix(ct, allowed) {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{
			"error": "Unsupported content type",
		})
	}
}
