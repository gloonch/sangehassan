package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/usecase"
)

type CatalogHandler struct {
	service *usecase.CatalogService
}

func NewCatalogHandler(service *usecase.CatalogService) *CatalogHandler {
	return &CatalogHandler{service: service}
}

func (h *CatalogHandler) Hub(c *gin.Context) {
	hub, err := h.service.Hub(c.Request.Context(), catalogLocale(c))
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load product catalog")
		return
	}
	respondOK(c, hub)
}

func (h *CatalogHandler) Page(c *gin.Context) {
	limit := parseCatalogInt(c.Query("limit"), 24, 1, 100)
	offset := parseCatalogInt(c.Query("offset"), 0, 0, 100000)
	filters := make(map[string][]string)
	for _, key := range []string{"color", "application", "finish", "form", "origin", "pattern", "availability"} {
		values := c.QueryArray(key)
		if len(values) == 1 && strings.Contains(values[0], ",") {
			values = strings.Split(values[0], ",")
		}
		if len(values) > 0 {
			filters[key] = values
		}
	}
	if query := strings.TrimSpace(c.Query("q")); query != "" {
		filters["__search"] = []string{query}
	}
	page, err := h.service.Page(
		c.Request.Context(), catalogLocale(c), c.Param("categorySlug"),
		c.Param("facet"), c.Param("value"), filters, limit, offset,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(c, http.StatusNotFound, "catalog page not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to load catalog page")
		return
	}
	respondOK(c, page)
}

func (h *CatalogHandler) Routes(c *gin.Context) {
	routes, err := h.service.Routes(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load catalog routes")
		return
	}
	respondOK(c, routes)
}

type catalogFacetPagePayload struct {
	CategoryID    int64  `json:"category_id"`
	TermID        int64  `json:"term_id"`
	TitleEN       string `json:"title_en"`
	TitleFA       string `json:"title_fa"`
	TitleAR       string `json:"title_ar"`
	DescriptionEN string `json:"description_en"`
	DescriptionFA string `json:"description_fa"`
	DescriptionAR string `json:"description_ar"`
	H1EN          string `json:"h1_en"`
	H1FA          string `json:"h1_fa"`
	H1AR          string `json:"h1_ar"`
	IntroEN       string `json:"intro_en"`
	IntroFA       string `json:"intro_fa"`
	IntroAR       string `json:"intro_ar"`
	IsActive      *bool  `json:"is_active"`
	IsIndexable   *bool  `json:"is_indexable"`
}

func (h *CatalogHandler) AdminListFacetPages(c *gin.Context) {
	pages, err := h.service.ListFacetPages(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load catalog facet pages")
		return
	}
	respondOK(c, pages)
}

func (h *CatalogHandler) AdminUpsertFacetPage(c *gin.Context) {
	var payload catalogFacetPagePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}
	isActive, isIndexable := true, true
	if payload.IsActive != nil {
		isActive = *payload.IsActive
	}
	if payload.IsIndexable != nil {
		isIndexable = *payload.IsIndexable
	}
	page, err := h.service.UpsertFacetPage(c.Request.Context(), domain.CatalogFacetPage{
		CategoryID: payload.CategoryID, TermID: payload.TermID,
		TitleEN: strings.TrimSpace(payload.TitleEN), TitleFA: strings.TrimSpace(payload.TitleFA), TitleAR: strings.TrimSpace(payload.TitleAR),
		DescriptionEN: strings.TrimSpace(payload.DescriptionEN), DescriptionFA: strings.TrimSpace(payload.DescriptionFA), DescriptionAR: strings.TrimSpace(payload.DescriptionAR),
		H1EN: strings.TrimSpace(payload.H1EN), H1FA: strings.TrimSpace(payload.H1FA), H1AR: strings.TrimSpace(payload.H1AR),
		IntroEN: strings.TrimSpace(payload.IntroEN), IntroFA: strings.TrimSpace(payload.IntroFA), IntroAR: strings.TrimSpace(payload.IntroAR),
		IsActive: isActive, IsIndexable: isIndexable,
	})
	if err != nil {
		respondError(c, http.StatusBadRequest, "failed to save catalog facet page")
		return
	}
	respondOK(c, page)
}

func (h *CatalogHandler) AdminDeleteFacetPage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.service.DeleteFacetPage(c.Request.Context(), id); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to delete catalog facet page")
		return
	}
	c.Status(http.StatusNoContent)
}

func catalogLocale(c *gin.Context) string {
	locale := strings.ToLower(strings.TrimSpace(c.DefaultQuery("locale", "fa")))
	if locale != "fa" && locale != "en" && locale != "ar" {
		return "fa"
	}
	return locale
}

func parseCatalogInt(raw string, fallback, minimum, maximum int) int {
	value, err := strconv.Atoi(raw)
	if err != nil || value < minimum {
		return fallback
	}
	if value > maximum {
		return maximum
	}
	return value
}
