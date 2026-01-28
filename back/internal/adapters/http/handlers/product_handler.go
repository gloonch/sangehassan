package handlers

import (
	"net/http"
	"strconv"

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
	TitleEN     string  `json:"title_en"`
	TitleFA     string  `json:"title_fa"`
	TitleAR     string  `json:"title_ar"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	ImageURL    string  `json:"image_url"`
	CategoryID  int64   `json:"category_id"`
	IsPopular   bool    `json:"is_popular"`
}

func (h *ProductHandler) List(c *gin.Context) {
	popularOnly := c.Query("popular") == "true"
	var (
		products []domain.Product
		err      error
	)
	if popularOnly {
		products, err = h.service.ListPopular(c.Request.Context())
	} else {
		products, err = h.service.List(c.Request.Context())
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load products")
		return
	}
	respondOK(c, products)
}

func (h *ProductHandler) Create(c *gin.Context) {
	var payload productPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}

	product, err := h.service.Create(c.Request.Context(), domain.Product{
		TitleEN:     payload.TitleEN,
		TitleFA:     payload.TitleFA,
		TitleAR:     payload.TitleAR,
		Description: payload.Description,
		Price:       payload.Price,
		ImageURL:    payload.ImageURL,
		CategoryID:  payload.CategoryID,
		IsPopular:   payload.IsPopular,
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

	product, err := h.service.Update(c.Request.Context(), domain.Product{
		ID:          id,
		TitleEN:     payload.TitleEN,
		TitleFA:     payload.TitleFA,
		TitleAR:     payload.TitleAR,
		Description: payload.Description,
		Price:       payload.Price,
		ImageURL:    payload.ImageURL,
		CategoryID:  payload.CategoryID,
		IsPopular:   payload.IsPopular,
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
