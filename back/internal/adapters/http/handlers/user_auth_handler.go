package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"sangehassan/back/internal/usecase"
)

const (
	accessCookieName  = "access_token"
	refreshCookieName = "refresh_token"
)

type UserAuthHandler struct {
	service      *usecase.UserAuthService
	cookieSecure bool
}

func NewUserAuthHandler(service *usecase.UserAuthService, cookieSecure bool) *UserAuthHandler {
	return &UserAuthHandler{service: service, cookieSecure: cookieSecure}
}

type userSignupPayload struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type userLoginPayload struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type userUpdatePayload struct {
	Email    *string `json:"email"`
	FullName *string `json:"full_name"`
	Phone    *string `json:"phone"`
}

func (h *UserAuthHandler) Signup(c *gin.Context) {
	var payload userSignupPayload
	if err := c.ShouldBindJSON(&payload); err != nil || payload.Phone == "" || payload.Password == "" {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}

	user, pair, err := h.service.SignUp(c.Request.Context(), payload.Phone, payload.Password)
	if err != nil {
		switch err {
		case usecase.ErrEmailExists:
			respondError(c, http.StatusConflict, "email already exists")
		case usecase.ErrPhoneExists:
			respondError(c, http.StatusConflict, "phone already exists")
		case usecase.ErrPhoneRequired:
			respondError(c, http.StatusBadRequest, "phone is required")
		default:
			respondError(c, http.StatusBadRequest, err.Error())
		}
		return
	}

	h.setCookies(c, pair)
	respondOK(c, user)
}

func (h *UserAuthHandler) Login(c *gin.Context) {
	var payload userLoginPayload
	if err := c.ShouldBindJSON(&payload); err != nil || payload.Phone == "" || payload.Password == "" {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}

	user, pair, err := h.service.Login(c.Request.Context(), payload.Phone, payload.Password)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	h.setCookies(c, pair)
	respondOK(c, user)
}

func (h *UserAuthHandler) Logout(c *gin.Context) {
	userID, _ := c.Get("user_id")
	if idStr, ok := userID.(string); ok {
		_ = h.service.Logout(c.Request.Context(), idStr)
	}
	h.clearCookies(c)
	c.Status(http.StatusNoContent)
}

func (h *UserAuthHandler) Refresh(c *gin.Context) {
	refresh, err := c.Cookie(refreshCookieName)
	if err != nil || refresh == "" {
		respondError(c, http.StatusUnauthorized, "missing refresh token")
		return
	}

	user, pair, err := h.service.Refresh(c.Request.Context(), refresh)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	h.setCookies(c, pair)
	respondOK(c, user)
}

func (h *UserAuthHandler) Session(c *gin.Context) {
	anonymous := gin.H{"authenticated": false, "user": nil}
	token, err := c.Cookie(accessCookieName)
	if err != nil || token == "" {
		respondOK(c, anonymous)
		return
	}

	userID, err := h.service.ParseAccess(token)
	if err != nil || userID == "" {
		respondOK(c, anonymous)
		return
	}

	user, err := h.service.GetMe(c.Request.Context(), userID)
	if err != nil {
		respondOK(c, anonymous)
		return
	}

	respondOK(c, gin.H{"authenticated": true, "user": user})
}

func (h *UserAuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get("user_id")
	idStr, _ := userID.(string)
	user, err := h.service.GetMe(c.Request.Context(), idStr)
	if err != nil {
		respondError(c, http.StatusNotFound, "user not found")
		return
	}
	respondOK(c, user)
}

func (h *UserAuthHandler) UpdateMe(c *gin.Context) {
	userID, _ := c.Get("user_id")
	idStr, _ := userID.(string)
	var payload userUpdatePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}
	user, err := h.service.UpdateMe(c.Request.Context(), idStr, payload.FullName, payload.Phone, payload.Email)
	if err != nil {
		switch err {
		case usecase.ErrEmailExists:
			respondError(c, http.StatusConflict, "email already exists")
		case usecase.ErrPhoneExists:
			respondError(c, http.StatusConflict, "phone already exists")
		default:
			respondError(c, http.StatusBadRequest, err.Error())
		}
		return
	}
	respondOK(c, user)
}

func (h *UserAuthHandler) Requests(c *gin.Context) {
	userID, _ := c.Get("user_id")
	idStr, _ := userID.(string)
	items, err := h.service.ListRequests(c.Request.Context(), idStr)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	respondOK(c, items)
}

func (h *UserAuthHandler) setCookies(c *gin.Context, pair usecase.TokenPair) {
	httpOnly := true
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(accessCookieName, pair.AccessToken, int(time.Until(pair.AccessExpires).Seconds()), "/", "", h.cookieSecure, httpOnly)
	c.SetCookie(refreshCookieName, pair.RefreshToken, int(time.Until(pair.RefreshExpires).Seconds()), "/", "", h.cookieSecure, httpOnly)
}

func (h *UserAuthHandler) clearCookies(c *gin.Context) {
	httpOnly := true
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(accessCookieName, "", -1, "/", "", h.cookieSecure, httpOnly)
	c.SetCookie(refreshCookieName, "", -1, "/", "", h.cookieSecure, httpOnly)
}
