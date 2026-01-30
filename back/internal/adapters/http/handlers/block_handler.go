package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/usecase"
)

type BlockHandler struct {
	service *usecase.BlockService
}

func NewBlockHandler(service *usecase.BlockService) *BlockHandler {
	return &BlockHandler{service: service}
}

type blockPayload struct {
	TitleEN     string   `json:"title_en"`
	TitleFA     string   `json:"title_fa"`
	TitleAR     string   `json:"title_ar"`
	Slug        string   `json:"slug"`
	StoneType   string   `json:"stone_type"`
	Quarry      string   `json:"quarry"`
	Dimensions  string   `json:"dimensions"`
	WeightTon   float64  `json:"weight_ton"`
	Status      string   `json:"status"`
	Description string   `json:"description"`
	ImageURL    string   `json:"image_url"`
	ImageURLs   []string `json:"image_urls"`
	IsFeatured  bool     `json:"is_featured"`
}

func (h *BlockHandler) List(c *gin.Context) {
	featuredOnly := c.Query("featured") == "true"
	statusFilter := strings.TrimSpace(c.Query("status"))

	var (
		blocks []domain.Block
		err    error
	)
	if featuredOnly {
		blocks, err = h.service.ListFeatured(c.Request.Context())
	} else {
		blocks, err = h.service.List(c.Request.Context())
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load blocks")
		return
	}
	if statusFilter != "" {
		filtered := make([]domain.Block, 0, len(blocks))
		for _, block := range blocks {
			if strings.EqualFold(block.Status, statusFilter) {
				filtered = append(filtered, block)
			}
		}
		blocks = filtered
	}
	respondOK(c, blocks)
}

func (h *BlockHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	block, err := h.service.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondError(c, http.StatusNotFound, "block not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to load block")
		return
	}
	respondOK(c, block)
}

func (h *BlockHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid block id")
		return
	}
	block, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondError(c, http.StatusNotFound, "block not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to load block")
		return
	}
	respondOK(c, block)
}

func (h *BlockHandler) Create(c *gin.Context) {
	var payload blockPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}
	status := strings.TrimSpace(payload.Status)
	if status == "" {
		status = "available"
	}

	block, err := h.service.Create(c.Request.Context(), domain.Block{
		TitleEN:     payload.TitleEN,
		TitleFA:     payload.TitleFA,
		TitleAR:     payload.TitleAR,
		Slug:        payload.Slug,
		StoneType:   payload.StoneType,
		Quarry:      payload.Quarry,
		Dimensions:  payload.Dimensions,
		WeightTon:   payload.WeightTon,
		Status:      status,
		Description: payload.Description,
		ImageURL:    payload.ImageURL,
		Images:      payload.ImageURLs,
		IsFeatured:  payload.IsFeatured,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create block")
		return
	}
	respondCreated(c, block)
}

func (h *BlockHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid block id")
		return
	}

	var payload blockPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}
	status := strings.TrimSpace(payload.Status)
	if status == "" {
		status = "available"
	}

	block, err := h.service.Update(c.Request.Context(), domain.Block{
		ID:          id,
		TitleEN:     payload.TitleEN,
		TitleFA:     payload.TitleFA,
		TitleAR:     payload.TitleAR,
		Slug:        payload.Slug,
		StoneType:   payload.StoneType,
		Quarry:      payload.Quarry,
		Dimensions:  payload.Dimensions,
		WeightTon:   payload.WeightTon,
		Status:      status,
		Description: payload.Description,
		ImageURL:    payload.ImageURL,
		Images:      payload.ImageURLs,
		IsFeatured:  payload.IsFeatured,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to update block")
		return
	}
	respondOK(c, block)
}

func (h *BlockHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid block id")
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to delete block")
		return
	}
	c.Status(http.StatusNoContent)
}
