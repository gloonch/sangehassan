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
		abortAPIError(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "برای ادامه وارد حساب شوید.")
		return false
	}
	id, issuedAt, err := m.auth.ParseAccessDetails(token)
	if err != nil {
		abortAPIError(c, http.StatusUnauthorized, "SESSION_INVALID", "نشست شما معتبر نیست؛ دوباره وارد شوید.")
		return false
	}
	if !m.operations.AccessAllowed(c.Request.Context(), id, issuedAt) {
		abortAPIError(c, http.StatusUnauthorized, "SESSION_REVOKED", "نشست شما منقضی یا باطل شده است.")
		return false
	}
	u, err := m.operations.GetOperationalUser(c.Request.Context(), id)
	if err != nil || u.Status != "ACTIVE" {
		abortAPIError(c, http.StatusUnauthorized, "ACCOUNT_INACTIVE", "این حساب فعال نیست.")
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
		abortAPIError(c, http.StatusForbidden, "PASSWORD_CHANGE_REQUIRED", "قبل از ادامه رمز عبور موقت را تغییر دهید.")
		return
	}
	if !m.operations.IsInternal(c.Request.Context(), id.(string)) {
		abortAPIError(c, http.StatusForbidden, "PERMISSION_DENIED", "شما دسترسی لازم برای این بخش را ندارید.")
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
			abortAPIError(c, http.StatusForbidden, "PASSWORD_CHANGE_REQUIRED", "قبل از ادامه رمز عبور موقت را تغییر دهید.")
			return
		}
		if !m.operations.HasPermission(c.Request.Context(), id.(string), code) {
			abortAPIError(c, http.StatusForbidden, "PERMISSION_DENIED", "شما دسترسی لازم برای این عملیات را ندارید.")
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
			abortAPIError(c, http.StatusForbidden, "PASSWORD_CHANGE_REQUIRED", "قبل از ادامه رمز عبور موقت را تغییر دهید.")
			return
		}
		for _, code := range codes {
			if m.operations.HasPermission(c.Request.Context(), id.(string), code) {
				return
			}
		}
		abortAPIError(c, http.StatusForbidden, "PERMISSION_DENIED", "شما دسترسی لازم برای این عملیات را ندارید.")
	}
}

// RequireFeature is used only on creation/mutation routes. Existing records
// remain readable while an optional module is disabled.
func (m *OperationsMiddleware) RequireFeature(settingKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.operations.FeatureEnabled(c.Request.Context(), settingKey) {
			abortAPIError(c, http.StatusConflict, "MODULE_DISABLED", "این ماژول در تنظیمات سیستم غیرفعال است.")
			return
		}
	}
}
