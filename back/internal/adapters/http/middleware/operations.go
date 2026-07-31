package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sangehassan/back/internal/usecase"
)

type OperationsMiddleware struct {
	auth       *usecase.UserAuthService
	operations *usecase.OperationsService
}

func NewOperationsMiddleware(auth *usecase.UserAuthService, operations *usecase.OperationsService) *OperationsMiddleware {
	return &OperationsMiddleware{auth: auth, operations: operations}
}
func (m *OperationsMiddleware) authenticate(c *gin.Context) bool {
	token, err := c.Cookie("access_token")
	if err != nil || token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return false
	}
	id, err := m.auth.ParseAccess(token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return false
	}
	u, err := m.operations.GetOperationalUser(c.Request.Context(), id)
	if err != nil || u.Status != "ACTIVE" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return false
	}
	c.Set("user_id", id)
	c.Set("operational_user", u)
	return true
}
func (m *OperationsMiddleware) RequireUser(c *gin.Context) { m.authenticate(c) }
func (m *OperationsMiddleware) RequireInternal(c *gin.Context) {
	if !m.authenticate(c) {
		return
	}
	id, _ := c.Get("user_id")
	u, _ := c.Get("operational_user")
	if u.(usecase.OperationalUser).MustChangePassword {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "error": "password change required"})
		return
	}
	if !m.operations.IsInternal(c.Request.Context(), id.(string)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "error": "forbidden"})
		return
	}
}
func (m *OperationsMiddleware) RequirePermission(code string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.authenticate(c) {
			return
		}
		id, _ := c.Get("user_id")
		u, _ := c.Get("operational_user")
		if u.(usecase.OperationalUser).MustChangePassword {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "error": "password change required"})
			return
		}
		if !m.operations.HasPermission(c.Request.Context(), id.(string), code) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "error": "forbidden"})
			return
		}
	}
}

func (m *OperationsMiddleware) RequireAnyPermission(codes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.authenticate(c) {
			return
		}
		id, _ := c.Get("user_id")
		u, _ := c.Get("operational_user")
		if u.(usecase.OperationalUser).MustChangePassword {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "error": "password change required"})
			return
		}
		for _, code := range codes {
			if m.operations.HasPermission(c.Request.Context(), id.(string), code) {
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "error": "forbidden"})
	}
}
