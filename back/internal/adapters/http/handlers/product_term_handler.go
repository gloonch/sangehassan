package handlers

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/usecase"
)

type ProductTermHandler struct {
	service *usecase.ProductTermService
}

func NewProductTermHandler(service *usecase.ProductTermService) *ProductTermHandler {
	return &ProductTermHandler{service: service}
}

type productTermPayload struct {
	Taxonomy string `json:"taxonomy"`
	Key      string `json:"key"`
	LabelEN  string `json:"label_en"`
	LabelFA  string `json:"label_fa"`
	LabelAR  string `json:"label_ar"`
}

func (h *ProductTermHandler) List(c *gin.Context) {
	taxonomy := strings.TrimSpace(c.Query("taxonomy"))
	terms, err := h.service.List(c.Request.Context(), taxonomy)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load product terms")
		return
	}
	respondOK(c, terms)
}

func (h *ProductTermHandler) Upsert(c *gin.Context) {
	var payload productTermPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}

	payload.Taxonomy = strings.TrimSpace(payload.Taxonomy)
	payload.Key = strings.TrimSpace(payload.Key)
	payload.LabelEN = strings.TrimSpace(payload.LabelEN)
	payload.LabelFA = strings.TrimSpace(payload.LabelFA)
	payload.LabelAR = strings.TrimSpace(payload.LabelAR)

	if payload.Taxonomy == "" {
		respondError(c, http.StatusBadRequest, "taxonomy is required")
		return
	}

	// Keep it simple: require at least one label, and derive missing ones from English.
	if payload.LabelEN == "" && payload.LabelFA == "" && payload.LabelAR == "" {
		respondError(c, http.StatusBadRequest, "at least one label is required")
		return
	}
	if payload.LabelEN == "" {
		if payload.LabelFA != "" {
			payload.LabelEN = payload.LabelFA
		} else {
			payload.LabelEN = payload.LabelAR
		}
	}
	if payload.LabelFA == "" {
		payload.LabelFA = payload.LabelEN
	}
	if payload.LabelAR == "" {
		payload.LabelAR = payload.LabelEN
	}

	if payload.Key == "" {
		payload.Key = slugify(payload.LabelEN)
	}

	term, err := h.service.Upsert(c.Request.Context(), domain.ProductTerm{
		Taxonomy: payload.Taxonomy,
		Key:      payload.Key,
		LabelEN:  payload.LabelEN,
		LabelFA:  payload.LabelFA,
		LabelAR:  payload.LabelAR,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to save product term")
		return
	}

	respondOK(c, term)
}

func (h *ProductTermHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to delete product term")
		return
	}
	c.Status(http.StatusNoContent)
}

var productTermSlugRegex = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(input string) string {
	value := strings.ToLower(strings.TrimSpace(input))
	value = productTermSlugRegex.ReplaceAllString(value, "-")
	return strings.Trim(value, "-")
}

