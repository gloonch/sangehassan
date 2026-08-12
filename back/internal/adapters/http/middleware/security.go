package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type attemptWindow struct {
	Count   int
	ResetAt time.Time
}

func RateLimit(max int, window time.Duration) gin.HandlerFunc {
	var mu sync.Mutex
	attempts := map[string]attemptWindow{}
	return func(c *gin.Context) {
		key := c.ClientIP()
		now := time.Now()
		mu.Lock()
		if len(attempts) > 10000 {
			for candidate, window := range attempts {
				if window.ResetAt.Before(now) {
					delete(attempts, candidate)
				}
			}
		}
		item := attempts[key]
		if item.ResetAt.Before(now) {
			item = attemptWindow{ResetAt: now.Add(window)}
		}
		item.Count++
		attempts[key] = item
		blocked := item.Count > max
		mu.Unlock()
		if blocked {
			abortAPIError(c, http.StatusTooManyRequests, "RATE_LIMITED", "تعداد تلاش‌ها بیش از حد مجاز است. کمی بعد دوباره تلاش کنید.")
			return
		}
		c.Next()
	}
}

// OriginGuard complements SameSite cookies for state-changing browser requests.
// Non-browser clients without an Origin header remain supported.
func OriginGuard(allowed []string) gin.HandlerFunc {
	set := map[string]bool{"http://localhost:5173": true, "http://localhost:5174": true}
	for _, origin := range allowed {
		set[strings.TrimRight(origin, "/")] = true
	}
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead || c.Request.Method == http.MethodOptions {
			return
		}
		origin := strings.TrimRight(c.GetHeader("Origin"), "/")
		if origin != "" && !set[origin] {
			abortAPIError(c, http.StatusForbidden, "INVALID_ORIGIN", "منشأ درخواست مجاز نیست.")
			return
		}
		c.Next()
	}
}
