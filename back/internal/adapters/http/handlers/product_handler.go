package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/usecase"
)

type ProductHandler struct {
	service *usecase.ProductService
}

func NewProductHandler(service *usecase.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

type productPayload struct {
	TitleEN                string   `json:"title_en"`
	TitleFA                string   `json:"title_fa"`
	TitleAR                string   `json:"title_ar"`
	Slug                   string   `json:"slug"`
	Description            string   `json:"description"`
	ShortDescription       string   `json:"short_description_html"`
	DescriptionHTMLEn      string   `json:"description_html_en"`
	DescriptionHTMLFa      string   `json:"description_html_fa"`
	DescriptionHTMLAr      string   `json:"description_html_ar"`
	ShortDescriptionHTMLEn string   `json:"short_description_html_en"`
	ShortDescriptionHTMLFa string   `json:"short_description_html_fa"`
	ShortDescriptionHTMLAr string   `json:"short_description_html_ar"`
	Price                  float64  `json:"price"`
	PriceHTML              string   `json:"price_html"`
	ImageURL               string   `json:"image_url"`
	ImageURLs              []string `json:"image_urls"`
	CategoryID             int64    `json:"category_id"`
	IsPopular              bool     `json:"is_popular"`
	TermIDs                []int64  `json:"term_ids"`
}

func (h *ProductHandler) List(c *gin.Context) {
	popularOnly := c.Query("popular") == "true"
	limit := 0
	offset := 0
	if rawLimit := c.Query("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil || parsedLimit <= 0 {
			respondError(c, http.StatusBadRequest, "invalid limit")
			return
		}
		limit = parsedLimit
	}
	if rawOffset := c.Query("offset"); rawOffset != "" {
		parsedOffset, err := strconv.Atoi(rawOffset)
		if err != nil || parsedOffset < 0 {
			respondError(c, http.StatusBadRequest, "invalid offset")
			return
		}
		offset = parsedOffset
	}
	var (
		products []domain.Product
		err      error
	)
	if popularOnly {
		products, err = h.service.ListPopular(c.Request.Context())
	} else {
		products, err = h.service.List(c.Request.Context(), limit, offset)
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load products")
		return
	}
	respondOK(c, products)
}

func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid product id")
		return
	}

	product, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, http.StatusNotFound, "product not found")
		return
	}
	respondOK(c, product)
}

func (h *ProductHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		respondError(c, http.StatusBadRequest, "invalid slug")
		return
	}

	product, err := h.service.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		respondError(c, http.StatusNotFound, "product not found")
		return
	}
	respondOK(c, product)
}

func (h *ProductHandler) Create(c *gin.Context) {
	var payload productPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}
	if !isAllowedImageRef(payload.ImageURL) || !allAllowedImageRefs(payload.ImageURLs) {
		respondError(c, http.StatusBadRequest, "images must be uploaded")
		return
	}

	descriptionEN := strings.TrimSpace(payload.DescriptionHTMLEn)
	if descriptionEN == "" {
		descriptionEN = payload.Description
	}
	shortEN := strings.TrimSpace(payload.ShortDescriptionHTMLEn)
	if shortEN == "" {
		shortEN = payload.ShortDescription
	}

	mainCategoryID := buildCategoryID(payload.CategoryID)
	product, err := h.service.Create(c.Request.Context(), domain.Product{
		TitleEN:                payload.TitleEN,
		TitleFA:                payload.TitleFA,
		TitleAR:                payload.TitleAR,
		Slug:                   payload.Slug,
		Description:            descriptionEN,
		DescriptionHTML:        descriptionEN,
		DescriptionHTMLEn:      descriptionEN,
		DescriptionHTMLFa:      payload.DescriptionHTMLFa,
		DescriptionHTMLAr:      payload.DescriptionHTMLAr,
		ShortDescriptionHTML:   shortEN,
		ShortDescriptionHTMLEn: shortEN,
		ShortDescriptionHTMLFa: payload.ShortDescriptionHTMLFa,
		ShortDescriptionHTMLAr: payload.ShortDescriptionHTMLAr,
		Price:                  payload.Price,
		PriceHTML:              payload.PriceHTML,
		ImageURL:               normalizeImageRef(payload.ImageURL),
		Images:                 normalizeImageRefs(payload.ImageURLs),
		MainCategoryID:         mainCategoryID,
		IsPopular:              payload.IsPopular,
		TermIDs:                payload.TermIDs,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create product")
		return
	}

	respondCreated(c, product)
}

func (h *ProductHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid product id")
		return
	}

	var payload productPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}
	if !isAllowedImageRef(payload.ImageURL) || !allAllowedImageRefs(payload.ImageURLs) {
		respondError(c, http.StatusBadRequest, "images must be uploaded")
		return
	}

	descriptionEN := strings.TrimSpace(payload.DescriptionHTMLEn)
	if descriptionEN == "" {
		descriptionEN = payload.Description
	}
	shortEN := strings.TrimSpace(payload.ShortDescriptionHTMLEn)
	if shortEN == "" {
		shortEN = payload.ShortDescription
	}

	mainCategoryID := buildCategoryID(payload.CategoryID)
	product, err := h.service.Update(c.Request.Context(), domain.Product{
		ID:                     id,
		TitleEN:                payload.TitleEN,
		TitleFA:                payload.TitleFA,
		TitleAR:                payload.TitleAR,
		Slug:                   payload.Slug,
		Description:            descriptionEN,
		DescriptionHTML:        descriptionEN,
		DescriptionHTMLEn:      descriptionEN,
		DescriptionHTMLFa:      payload.DescriptionHTMLFa,
		DescriptionHTMLAr:      payload.DescriptionHTMLAr,
		ShortDescriptionHTML:   shortEN,
		ShortDescriptionHTMLEn: shortEN,
		ShortDescriptionHTMLFa: payload.ShortDescriptionHTMLFa,
		ShortDescriptionHTMLAr: payload.ShortDescriptionHTMLAr,
		Price:                  payload.Price,
		PriceHTML:              payload.PriceHTML,
		ImageURL:               normalizeImageRef(payload.ImageURL),
		Images:                 normalizeImageRefs(payload.ImageURLs),
		MainCategoryID:         mainCategoryID,
		IsPopular:              payload.IsPopular,
		TermIDs:                payload.TermIDs,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to update product")
		return
	}

	respondOK(c, product)
}

func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid product id")
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to delete product")
		return
	}

	c.Status(http.StatusNoContent)
}

func buildCategoryID(value int64) *int64 {
	if value == 0 {
		return nil
	}
	return &value
}
