package handlers

import (
	"database/sql"
	"errors"
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

func noStoreBlogResponse(c *gin.Context) {
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.Header("Surrogate-Control", "no-store")
}

func (h *BlogHandler) ListPublic(c *gin.Context) {
	noStoreBlogResponse(c)
	blogs, err := h.service.ListPublic(c.Request.Context(), c.DefaultQuery("locale", "en"))
	if err != nil {
		h.respondBlogError(c, err, "failed to load blogs")
		return
	}
	respondOK(c, blogs)
}

func (h *BlogHandler) GetPublicBySlug(c *gin.Context) {
	noStoreBlogResponse(c)
	blog, err := h.service.GetPublicBySlug(c.Request.Context(), c.Param("locale"), c.Param("slug"))
	if err != nil {
		h.respondBlogError(c, err, "blog not found")
		return
	}
	respondOK(c, blog)
}

func (h *BlogHandler) ListAdmin(c *gin.Context) {
	noStoreBlogResponse(c)
	blogs, err := h.service.ListAdmin(c.Request.Context())
	if err != nil {
		h.respondBlogError(c, err, "failed to load blogs")
		return
	}
	respondOK(c, blogs)
}

func (h *BlogHandler) GetByID(c *gin.Context) {
	noStoreBlogResponse(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid blog id")
		return
	}
	blog, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		h.respondBlogError(c, err, "blog not found")
		return
	}
	respondOK(c, blog)
}

func (h *BlogHandler) Create(c *gin.Context) {
	noStoreBlogResponse(c)
	var blog domain.Blog
	if err := c.ShouldBindJSON(&blog); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}
	if !validBlogImages(blog) {
		respondError(c, http.StatusBadRequest, "blog images must be uploaded")
		return
	}
	blog.CoverImageURL = normalizeImageRef(blog.CoverImageURL)
	blog.OGImageURL = normalizeImageRef(blog.OGImageURL)
	created, err := h.service.Create(c.Request.Context(), blog)
	if err != nil {
		h.respondBlogError(c, err, "failed to create blog")
		return
	}
	respondCreated(c, created)
}

func (h *BlogHandler) Update(c *gin.Context) {
	noStoreBlogResponse(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid blog id")
		return
	}
	var blog domain.Blog
	if err := c.ShouldBindJSON(&blog); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}
	if !validBlogImages(blog) {
		respondError(c, http.StatusBadRequest, "blog images must be uploaded")
		return
	}
	blog.ID = id
	blog.CoverImageURL = normalizeImageRef(blog.CoverImageURL)
	blog.OGImageURL = normalizeImageRef(blog.OGImageURL)
	updated, err := h.service.Update(c.Request.Context(), blog)
	if err != nil {
		h.respondBlogError(c, err, "failed to update blog")
		return
	}
	respondOK(c, updated)
}

func (h *BlogHandler) Delete(c *gin.Context) {
	noStoreBlogResponse(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid blog id")
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		h.respondBlogError(c, err, "failed to delete blog")
		return
	}
	c.Status(http.StatusNoContent)
}

func validBlogImages(blog domain.Blog) bool {
	return isAllowedImageRef(blog.CoverImageURL) && isAllowedImageRef(blog.OGImageURL)
}

func (h *BlogHandler) respondBlogError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		respondError(c, http.StatusNotFound, fallback)
	case errors.Is(err, usecase.ErrInvalidBlog):
		respondError(c, http.StatusBadRequest, err.Error())
	default:
		respondError(c, http.StatusInternalServerError, fallback)
	}
}
