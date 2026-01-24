package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"sangehassan/back/internal/usecase"
)

const adminCookieName = "sh_admin"

type AuthHandler struct {
	service      *usecase.AuthService
	cookieSecure bool
	cookieMaxAge int
}

func NewAuthHandler(service *usecase.AuthService, cookieSecure bool, ttlHours int) *AuthHandler {
	return &AuthHandler{
		service:      service,
		cookieSecure: cookieSecure,
		cookieMaxAge: ttlHours * 3600,
	}
}

type loginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var payload loginPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}

	token, err := h.service.Login(c.Request.Context(), payload.Username, payload.Password)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(adminCookieName, token, h.cookieMaxAge, "/", "", h.cookieSecure, true)
	respondOK(c, gin.H{"username": payload.Username})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(adminCookieName, "", -1, "/", "", h.cookieSecure, true)
	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) Session(c *gin.Context) {
	username, _ := c.Get("admin_username")
	respondOK(c, gin.H{"username": username})
}
