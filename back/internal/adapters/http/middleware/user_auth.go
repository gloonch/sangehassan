package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"sangehassan/back/internal/usecase"
)

type UserAuthMiddleware struct {
	auth *usecase.UserAuthService
}

func NewUserAuthMiddleware(auth *usecase.UserAuthService) *UserAuthMiddleware {
	return &UserAuthMiddleware{auth: auth}
}

func (m *UserAuthMiddleware) RequireUser(c *gin.Context) {
	token, err := c.Cookie("access_token")
	if err != nil || token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "error": "missing token"})
		return
	}

	userID, err := m.auth.ParseAccess(token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "error": "invalid token"})
		return
	}

	c.Set("user_id", userID)
	c.Next()
}
