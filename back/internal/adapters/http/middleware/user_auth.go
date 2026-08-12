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
		abortAPIError(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "برای ادامه وارد حساب شوید.")
		return
	}

	userID, err := m.auth.ValidateAccess(c.Request.Context(), token)
	if err != nil {
		abortAPIError(c, http.StatusUnauthorized, "SESSION_INVALID", "نشست شما معتبر نیست؛ دوباره وارد شوید.")
		return
	}

	c.Set("user_id", userID)
	c.Next()
}
