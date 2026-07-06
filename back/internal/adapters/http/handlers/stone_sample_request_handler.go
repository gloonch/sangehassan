package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/ports"
	"sangehassan/back/internal/usecase"
)

type StoneSampleRequestHandler struct {
	service *usecase.StoneSampleRequestService
}

func NewStoneSampleRequestHandler(service *usecase.StoneSampleRequestService) *StoneSampleRequestHandler {
	return &StoneSampleRequestHandler{service: service}
}

type createStoneSampleRequestPayload struct {
	Boxes          [][]int64 `json:"boxes"`
	AddressID      *int64    `json:"address_id"`
	AddressText    string    `json:"address_text"`
	PhoneID        *int64    `json:"phone_id"`
	Phone          string    `json:"phone"`
	ShippingMethod string    `json:"shipping_method"`
}

type sampleRequestStatusPayload struct {
	Status    string `json:"status"`
	AdminNote string `json:"admin_note"`
}

func (h *StoneSampleRequestHandler) SampleCategories(c *gin.Context) {
	categories, err := h.service.ListSampleCategories(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load sample categories")
		return
	}
	respondOK(c, categories)
}

func (h *StoneSampleRequestHandler) SampleProducts(c *gin.Context) {
	filter := domain.StoneSampleCatalogFilter{
		Query:  c.Query("q"),
		Limit:  parseIntDefault(c.Query("limit"), 9),
		Offset: parseIntDefault(c.Query("offset"), 0),
	}
	if rawCategoryID := strings.TrimSpace(c.Query("category_id")); rawCategoryID != "" {
		id, err := strconv.ParseInt(rawCategoryID, 10, 64)
		if err != nil || id <= 0 {
			respondError(c, http.StatusBadRequest, "invalid category_id")
			return
		}
		filter.CategoryID = &id
	}
	products, total, err := h.service.ListSampleProducts(c.Request.Context(), filter)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load sample products")
		return
	}
	respondOK(c, gin.H{
		"items":  products,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

func (h *StoneSampleRequestHandler) Options(c *gin.Context) {
	userID, _ := c.Get("user_id")
	idStr, _ := userID.(string)
	options, err := h.service.Options(c.Request.Context(), idStr)
	if err != nil {
		respondStoneSampleError(c, err)
		return
	}
	respondOK(c, options)
}

func (h *StoneSampleRequestHandler) Create(c *gin.Context) {
	userID, _ := c.Get("user_id")
	idStr, _ := userID.(string)
	var payload createStoneSampleRequestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}
	request, err := h.service.Create(c.Request.Context(), domain.CreateStoneSampleRequestInput{
		UserID:         idStr,
		Boxes:          payload.Boxes,
		AddressID:      payload.AddressID,
		AddressText:    payload.AddressText,
		PhoneID:        payload.PhoneID,
		Phone:          payload.Phone,
		ShippingMethod: payload.ShippingMethod,
	})
	if err != nil {
		respondStoneSampleError(c, err)
		return
	}
	respondCreated(c, request)
}

func (h *StoneSampleRequestHandler) ListMine(c *gin.Context) {
	userID, _ := c.Get("user_id")
	idStr, _ := userID.(string)
	items, err := h.service.ListByUser(c.Request.Context(), idStr, parseIntDefault(c.Query("limit"), 50), parseIntDefault(c.Query("offset"), 0))
	if err != nil {
		respondStoneSampleError(c, err)
		return
	}
	respondOK(c, items)
}

func (h *StoneSampleRequestHandler) GetMine(c *gin.Context) {
	userID, _ := c.Get("user_id")
	idStr, _ := userID.(string)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		respondError(c, http.StatusBadRequest, "invalid sample request id")
		return
	}
	item, err := h.service.GetByUser(c.Request.Context(), idStr, id)
	if err != nil {
		respondStoneSampleError(c, err)
		return
	}
	respondOK(c, item)
}

func (h *StoneSampleRequestHandler) AdminList(c *gin.Context) {
	filter := ports.StoneSampleRequestFilter{
		Limit:  parseIntDefault(c.Query("limit"), 50),
		Offset: parseIntDefault(c.Query("offset"), 0),
	}
	if statusParam := strings.TrimSpace(c.Query("status")); statusParam != "" {
		for _, value := range strings.Split(statusParam, ",") {
			status := domain.StoneSampleRequestStatus(strings.ToUpper(strings.TrimSpace(value)))
			if status != "" {
				filter.Status = append(filter.Status, status)
			}
		}
	}
	items, err := h.service.ListAdmin(c.Request.Context(), filter)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load sample requests")
		return
	}
	respondOK(c, h.service.EnrichUsers(c.Request.Context(), items))
}

func (h *StoneSampleRequestHandler) AdminGet(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		respondError(c, http.StatusBadRequest, "invalid sample request id")
		return
	}
	item, err := h.service.GetAdmin(c.Request.Context(), id)
	if err != nil {
		respondStoneSampleError(c, err)
		return
	}
	respondOK(c, h.service.EnrichUser(c.Request.Context(), item))
}

func (h *StoneSampleRequestHandler) AdminUpdateStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		respondError(c, http.StatusBadRequest, "invalid sample request id")
		return
	}
	var payload sampleRequestStatusPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}
	status := domain.StoneSampleRequestStatus(strings.ToUpper(strings.TrimSpace(payload.Status)))
	var note *string
	if strings.TrimSpace(payload.AdminNote) != "" {
		adminNote := strings.TrimSpace(payload.AdminNote)
		note = &adminNote
	}
	if err := h.service.UpdateStatus(c.Request.Context(), id, status, note, nil); err != nil {
		respondStoneSampleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func respondStoneSampleError(c *gin.Context, err error) {
	status := http.StatusBadRequest
	message := err.Error()
	if errors.Is(err, usecase.ErrStoneSampleUserRequired) {
		status = http.StatusUnauthorized
	}
	if errors.Is(err, usecase.ErrStoneSampleProductInvalid) ||
		errors.Is(err, usecase.ErrStoneSampleBoxesRequired) ||
		errors.Is(err, usecase.ErrStoneSampleTooManyBoxes) ||
		errors.Is(err, usecase.ErrStoneSampleIncompleteBox) ||
		errors.Is(err, usecase.ErrStoneSampleDuplicateProduct) ||
		errors.Is(err, usecase.ErrStoneSampleAddressRequired) ||
		errors.Is(err, usecase.ErrStoneSamplePhoneRequired) ||
		errors.Is(err, usecase.ErrStoneSampleShippingInvalid) ||
		errors.Is(err, usecase.ErrStoneSampleStatusInvalid) {
		respondError(c, status, message)
		return
	}
	if errors.Is(err, usecase.ErrUserNotFound) {
		respondError(c, http.StatusNotFound, message)
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		respondError(c, http.StatusNotFound, "sample request not found")
		return
	}
	respondError(c, http.StatusInternalServerError, "sample request failed")
}
