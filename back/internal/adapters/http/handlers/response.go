package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func respondOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

func respondCreated(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": data})
}

func respondError(c *gin.Context, status int, message string) {
	code := "REQUEST_FAILED"
	if status == http.StatusBadRequest {
		code = "VALIDATION_FAILED"
	}
	if status == http.StatusUnauthorized {
		code = "AUTHENTICATION_REQUIRED"
	}
	if status == http.StatusForbidden {
		code = "PERMISSION_DENIED"
	}
	if status == http.StatusNotFound {
		code = "NOT_FOUND"
	}
	if status >= 500 {
		code, message = "INTERNAL_ERROR", "خطای غیرمنتظره‌ای رخ داد. دوباره تلاش کنید."
	}
	respondErrorCode(c, status, code, message, nil)
}

func respondErrorCode(c *gin.Context, status int, code, message string, details any) {
	requestID, _ := c.Get("request_id")
	requestIDText, _ := requestID.(string)
	c.JSON(status, gin.H{"success": false, "error": gin.H{"code": code, "message": message, "details": details, "requestId": requestIDText}})
}
