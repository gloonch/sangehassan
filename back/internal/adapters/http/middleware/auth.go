package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"sangehassan/back/internal/usecase"
)

type AuthMiddleware struct {
	authService *usecase.AuthService
}

func NewAuthMiddleware(authService *usecase.AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

func (m *AuthMiddleware) RequireAdmin(c *gin.Context) {
	token, err := c.Cookie("sh_admin")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "error": "missing token"})
		return
	}

	username, err := m.authService.ParseToken(token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "error": "invalid token"})
		return
	}

	c.Set("admin_username", username)
	c.Next()
}
