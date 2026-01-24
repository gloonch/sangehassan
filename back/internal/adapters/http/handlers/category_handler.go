package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/usecase"
)

type CategoryHandler struct {
	service *usecase.CategoryService
}

func NewCategoryHandler(service *usecase.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

type categoryPayload struct {
	TitleEN string `json:"title_en"`
	TitleFA string `json:"title_fa"`
	TitleAR string `json:"title_ar"`
}

func (h *CategoryHandler) List(c *gin.Context) {
	categories, err := h.service.List(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load categories")
		return
	}
	respondOK(c, categories)
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var payload categoryPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}

	category, err := h.service.Create(c.Request.Context(), domain.Category{
		TitleEN: payload.TitleEN,
		TitleFA: payload.TitleFA,
		TitleAR: payload.TitleAR,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create category")
		return
	}

	respondCreated(c, category)
}

func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid category id")
		return
	}

	var payload categoryPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}

	category, err := h.service.Update(c.Request.Context(), domain.Category{
		ID:      id,
		TitleEN: payload.TitleEN,
		TitleFA: payload.TitleFA,
		TitleAR: payload.TitleAR,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to update category")
		return
	}

	respondOK(c, category)
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid category id")
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to delete category")
		return
	}

	c.Status(http.StatusNoContent)
}
