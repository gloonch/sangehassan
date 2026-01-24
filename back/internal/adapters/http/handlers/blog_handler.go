package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/usecase"
)

type BlogHandler struct {
	service *usecase.BlogService
}

func NewBlogHandler(service *usecase.BlogService) *BlogHandler {
	return &BlogHandler{service: service}
}

type blogPayload struct {
	Title         string `json:"title"`
	Excerpt       string `json:"excerpt"`
	Content       string `json:"content"`
	CoverImageURL string `json:"cover_image_url"`
}

func (h *BlogHandler) List(c *gin.Context) {
	blogs, err := h.service.List(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load blogs")
		return
	}
	respondOK(c, blogs)
}

func (h *BlogHandler) Create(c *gin.Context) {
	var payload blogPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}

	blog, err := h.service.Create(c.Request.Context(), domain.Blog{
		Title:         payload.Title,
		Excerpt:       payload.Excerpt,
		Content:       payload.Content,
		CoverImageURL: payload.CoverImageURL,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create blog")
		return
	}

	respondCreated(c, blog)
}

func (h *BlogHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid blog id")
		return
	}

	var payload blogPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}

	blog, err := h.service.Update(c.Request.Context(), domain.Blog{
		ID:            id,
		Title:         payload.Title,
		Excerpt:       payload.Excerpt,
		Content:       payload.Content,
		CoverImageURL: payload.CoverImageURL,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to update blog")
		return
	}

	respondOK(c, blog)
}

func (h *BlogHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid blog id")
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to delete blog")
		return
	}

	c.Status(http.StatusNoContent)
}
