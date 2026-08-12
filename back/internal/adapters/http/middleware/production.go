package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"sangehassan/back/internal/usecase"
)

func newRequestID() string {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return time.Now().UTC().Format("20060102T150405.000000000")
	}
	return hex.EncodeToString(buffer)
}

func RequestContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader("X-Request-ID"))
		if requestID == "" || len(requestID) > 128 {
			requestID = newRequestID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Request = c.Request.WithContext(usecase.WithRequestID(c.Request.Context(), requestID))
		c.Next()
	}
}

func StructuredLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()
		requestID, _ := c.Get("request_id")
		userID, _ := c.Get("user_id")
		level := slog.LevelInfo
		if c.Writer.Status() >= 500 {
			level = slog.LevelError
		} else if c.Writer.Status() >= 400 {
			level = slog.LevelWarn
		}
		logger.Log(c.Request.Context(), level, "http_request",
			"requestId", requestID,
			"userId", userID,
			"method", c.Request.Method,
			"path", c.FullPath(),
			"status", c.Writer.Status(),
			"durationMs", time.Since(started).Milliseconds(),
			"clientIp", c.ClientIP(),
		)
	}
}

func StructuredRecovery(logger *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		requestID, _ := c.Get("request_id")
		logger.Error("http_panic", "requestId", requestID, "path", c.Request.URL.Path, "panic", recovered, "stack", string(debug.Stack()))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"code": "INTERNAL_ERROR", "message": "خطای غیرمنتظره‌ای رخ داد. دوباره تلاش کنید.", "details": nil, "requestId": requestID}})
	})
}

func SecurityHeaders(production bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(self), microphone=(), geolocation=()")
		c.Header("Content-Security-Policy", "default-src 'self'; img-src 'self' data: blob: https:; style-src 'self' 'unsafe-inline'; script-src 'self'; connect-src 'self'; font-src 'self' data:; object-src 'none'; base-uri 'self'; frame-ancestors 'none'")
		if production {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}

func abortAPIError(c *gin.Context, status int, code, message string) {
	requestID, _ := c.Get("request_id")
	c.AbortWithStatusJSON(status, gin.H{"success": false, "error": gin.H{"code": code, "message": message, "details": nil, "requestId": requestID}})
}
