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

type ProjectHandler struct {
	service *usecase.ProjectService
}

func NewProjectHandler(service *usecase.ProjectService) *ProjectHandler {
	return &ProjectHandler{service: service}
}

type projectPayload struct {
	Description   string   `json:"description"`
	DescriptionEN string   `json:"description_en"`
	DescriptionFA string   `json:"description_fa"`
	DescriptionAR string   `json:"description_ar"`
	CoverImageURL string   `json:"cover_image_url"`
	GalleryImages []string `json:"gallery_images"`
	SortOrder     int      `json:"sort_order"`
}

func (h *ProjectHandler) ListPublic(c *gin.Context) {
	projects, err := h.service.ListPublic(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load projects")
		return
	}
	respondOK(c, projects)
}

func (h *ProjectHandler) List(c *gin.Context) {
	projects, err := h.service.List(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load projects")
		return
	}
	respondOK(c, projects)
}

func (h *ProjectHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid project id")
		return
	}

	project, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondError(c, http.StatusNotFound, "project not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to load project")
		return
	}
	respondOK(c, project)
}

func (h *ProjectHandler) Create(c *gin.Context) {
	var payload projectPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}
	if payload.CoverImageURL == "" || !isAllowedImageRef(payload.CoverImageURL) {
		respondError(c, http.StatusBadRequest, "cover image must be uploaded")
		return
	}
	if len(payload.GalleryImages) > 5 {
		respondError(c, http.StatusBadRequest, "gallery images cannot exceed 5")
		return
	}
	if !allAllowedImageRefs(payload.GalleryImages) {
		respondError(c, http.StatusBadRequest, "gallery images must be uploaded")
		return
	}

	descriptionEN := payload.DescriptionEN
	if descriptionEN == "" {
		descriptionEN = payload.Description
	}

	project, err := h.service.Create(c.Request.Context(), domain.Project{
		Description:   descriptionEN,
		DescriptionEN: descriptionEN,
		DescriptionFA: payload.DescriptionFA,
		DescriptionAR: payload.DescriptionAR,
		CoverImageURL: normalizeImageRef(payload.CoverImageURL),
		GalleryImages: normalizeImageRefs(payload.GalleryImages),
		SortOrder:     payload.SortOrder,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create project")
		return
	}
	respondCreated(c, project)
}

func (h *ProjectHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid project id")
		return
	}

	var payload projectPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}
	if payload.CoverImageURL == "" || !isAllowedImageRef(payload.CoverImageURL) {
		respondError(c, http.StatusBadRequest, "cover image must be uploaded")
		return
	}
	if len(payload.GalleryImages) > 5 {
		respondError(c, http.StatusBadRequest, "gallery images cannot exceed 5")
		return
	}
	if !allAllowedImageRefs(payload.GalleryImages) {
		respondError(c, http.StatusBadRequest, "gallery images must be uploaded")
		return
	}

	descriptionEN := payload.DescriptionEN
	if descriptionEN == "" {
		descriptionEN = payload.Description
	}

	project, err := h.service.Update(c.Request.Context(), domain.Project{
		ID:            id,
		Description:   descriptionEN,
		DescriptionEN: descriptionEN,
		DescriptionFA: payload.DescriptionFA,
		DescriptionAR: payload.DescriptionAR,
		CoverImageURL: normalizeImageRef(payload.CoverImageURL),
		GalleryImages: normalizeImageRefs(payload.GalleryImages),
		SortOrder:     payload.SortOrder,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to update project")
		return
	}
	respondOK(c, project)
}

func (h *ProjectHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid project id")
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to delete project")
		return
	}
	c.Status(http.StatusNoContent)
}
