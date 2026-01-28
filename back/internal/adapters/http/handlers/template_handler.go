package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/usecase"
)

type TemplateHandler struct {
	service *usecase.TemplateService
}

func NewTemplateHandler(service *usecase.TemplateService) *TemplateHandler {
	return &TemplateHandler{service: service}
}

type templatePayload struct {
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
	IsActive bool   `json:"is_active"`
}

func (h *TemplateHandler) List(c *gin.Context) {
	activeOnly := c.Query("active") == "true"
	var (
		templates []domain.Template
		err       error
	)
	if activeOnly {
		templates, err = h.service.ListActive(c.Request.Context())
	} else {
		templates, err = h.service.List(c.Request.Context())
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load templates")
		return
	}
	respondOK(c, templates)
}

func (h *TemplateHandler) Create(c *gin.Context) {
	var payload templatePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}

	template, err := h.service.Create(c.Request.Context(), domain.Template{
		Name:     payload.Name,
		ImageURL: payload.ImageURL,
		IsActive: payload.IsActive,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create template")
		return
	}

	respondCreated(c, template)
}

func (h *TemplateHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid template id")
		return
	}

	var payload templatePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}

	template, err := h.service.Update(c.Request.Context(), domain.Template{
		ID:       id,
		Name:     payload.Name,
		ImageURL: payload.ImageURL,
		IsActive: payload.IsActive,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to update template")
		return
	}

	respondOK(c, template)
}

func (h *TemplateHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid template id")
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to delete template")
		return
	}

	c.Status(http.StatusNoContent)
}
