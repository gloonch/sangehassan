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
	TitleEN          string `json:"title_en"`
	TitleFA          string `json:"title_fa"`
	TitleAR          string `json:"title_ar"`
	Slug             string `json:"slug"`
	Parent           *int64 `json:"parent_id"`
	ImageURL         string `json:"image_url"`
	IntroEN          string `json:"intro_en"`
	IntroFA          string `json:"intro_fa"`
	IntroAR          string `json:"intro_ar"`
	SEOTitleEN       string `json:"seo_title_en"`
	SEOTitleFA       string `json:"seo_title_fa"`
	SEOTitleAR       string `json:"seo_title_ar"`
	SEODescriptionEN string `json:"seo_description_en"`
	SEODescriptionFA string `json:"seo_description_fa"`
	SEODescriptionAR string `json:"seo_description_ar"`
	IsActive         *bool  `json:"is_active"`
	IsIndexable      *bool  `json:"is_indexable"`
}

func categoryFromPayload(payload categoryPayload) domain.Category {
	isActive := true
	isIndexable := true
	if payload.IsActive != nil {
		isActive = *payload.IsActive
	}
	if payload.IsIndexable != nil {
		isIndexable = *payload.IsIndexable
	}
	return domain.Category{
		TitleEN: payload.TitleEN, TitleFA: payload.TitleFA, TitleAR: payload.TitleAR,
		Slug: payload.Slug, ParentID: payload.Parent, ImageURL: payload.ImageURL,
		IntroEN: payload.IntroEN, IntroFA: payload.IntroFA, IntroAR: payload.IntroAR,
		SEOTitleEN: payload.SEOTitleEN, SEOTitleFA: payload.SEOTitleFA, SEOTitleAR: payload.SEOTitleAR,
		SEODescriptionEN: payload.SEODescriptionEN, SEODescriptionFA: payload.SEODescriptionFA, SEODescriptionAR: payload.SEODescriptionAR,
		IsActive: isActive, IsIndexable: isIndexable,
	}
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

	category, err := h.service.Create(c.Request.Context(), categoryFromPayload(payload))
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

	category := categoryFromPayload(payload)
	category.ID = id
	category, err = h.service.Update(c.Request.Context(), category)
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
