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
	c.JSON(status, gin.H{"success": false, "error": message})
}
