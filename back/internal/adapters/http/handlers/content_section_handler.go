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

type ContentSectionHandler struct {
	service *usecase.ContentSectionService
}

func NewContentSectionHandler(service *usecase.ContentSectionService) *ContentSectionHandler {
	return &ContentSectionHandler{service: service}
}

type contentSectionPayload struct {
	Page          string   `json:"page"`
	Key           string   `json:"key"`
	TitleEN       string   `json:"title_en"`
	TitleFA       string   `json:"title_fa"`
	TitleAR       string   `json:"title_ar"`
	SubtitleEN    string   `json:"subtitle_en"`
	SubtitleFA    string   `json:"subtitle_fa"`
	SubtitleAR    string   `json:"subtitle_ar"`
	DescriptionEN string   `json:"description_en"`
	DescriptionFA string   `json:"description_fa"`
	DescriptionAR string   `json:"description_ar"`
	CTALabelEN    string   `json:"cta_label_en"`
	CTALabelFA    string   `json:"cta_label_fa"`
	CTALabelAR    string   `json:"cta_label_ar"`
	CTAHref       string   `json:"cta_href"`
	OrderIndex    int      `json:"order_index"`
	IsActive      bool     `json:"is_active"`
	ImageURLs     []string `json:"image_urls"`
}

func (h *ContentSectionHandler) ListPublic(c *gin.Context) {
	page := c.Query("page")
	sections, err := h.service.List(c.Request.Context(), page)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load sections")
		return
	}
	filtered := make([]domain.ContentSection, 0, len(sections))
	for _, section := range sections {
		if section.IsActive {
			filtered = append(filtered, section)
		}
	}
	respondOK(c, filtered)
}

func (h *ContentSectionHandler) List(c *gin.Context) {
	page := c.Query("page")
	sections, err := h.service.List(c.Request.Context(), page)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load sections")
		return
	}
	respondOK(c, sections)
}

func (h *ContentSectionHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid section id")
		return
	}
	section, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondError(c, http.StatusNotFound, "section not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to load section")
		return
	}
	respondOK(c, section)
}

func (h *ContentSectionHandler) Create(c *gin.Context) {
	var payload contentSectionPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}

	section, err := h.service.Create(c.Request.Context(), domain.ContentSection{
		Page:          payload.Page,
		Key:           payload.Key,
		TitleEN:       payload.TitleEN,
		TitleFA:       payload.TitleFA,
		TitleAR:       payload.TitleAR,
		SubtitleEN:    payload.SubtitleEN,
		SubtitleFA:    payload.SubtitleFA,
		SubtitleAR:    payload.SubtitleAR,
		DescriptionEN: payload.DescriptionEN,
		DescriptionFA: payload.DescriptionFA,
		DescriptionAR: payload.DescriptionAR,
		CTALabelEN:    payload.CTALabelEN,
		CTALabelFA:    payload.CTALabelFA,
		CTALabelAR:    payload.CTALabelAR,
		CTAHref:       payload.CTAHref,
		OrderIndex:    payload.OrderIndex,
		IsActive:      payload.IsActive,
		Images:        payload.ImageURLs,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create section")
		return
	}
	respondCreated(c, section)
}

func (h *ContentSectionHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid section id")
		return
	}

	var payload contentSectionPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}

	section, err := h.service.Update(c.Request.Context(), domain.ContentSection{
		ID:            id,
		Page:          payload.Page,
		Key:           payload.Key,
		TitleEN:       payload.TitleEN,
		TitleFA:       payload.TitleFA,
		TitleAR:       payload.TitleAR,
		SubtitleEN:    payload.SubtitleEN,
		SubtitleFA:    payload.SubtitleFA,
		SubtitleAR:    payload.SubtitleAR,
		DescriptionEN: payload.DescriptionEN,
		DescriptionFA: payload.DescriptionFA,
		DescriptionAR: payload.DescriptionAR,
		CTALabelEN:    payload.CTALabelEN,
		CTALabelFA:    payload.CTALabelFA,
		CTALabelAR:    payload.CTALabelAR,
		CTAHref:       payload.CTAHref,
		OrderIndex:    payload.OrderIndex,
		IsActive:      payload.IsActive,
		Images:        payload.ImageURLs,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to update section")
		return
	}
	respondOK(c, section)
}

func (h *ContentSectionHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid section id")
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to delete section")
		return
	}
	c.Status(http.StatusNoContent)
}
