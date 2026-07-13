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

type ContactSubmissionHandler struct {
	submissions *usecase.ContactSubmissionService
}

func NewContactSubmissionHandler(service *usecase.ContactSubmissionService) *ContactSubmissionHandler {
	return &ContactSubmissionHandler{submissions: service}
}

type contactSubmissionPayload struct {
	FullName    string `json:"full_name"`
	Email       string `json:"email"`
	CountryISO  string `json:"country_iso"`
	CountryCode string `json:"country_code"`
	PhoneNumber string `json:"phone_number"`
	Message     string `json:"message"`
	Source      string `json:"source"`
}

func (h *ContactSubmissionHandler) Create(c *gin.Context) {
	var payload contactSubmissionPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid contact submission")
		return
	}

	submission, err := h.submissions.Create(c.Request.Context(), domain.ContactSubmission{
		FullName:    payload.FullName,
		Email:       payload.Email,
		CountryISO:  payload.CountryISO,
		CountryCode: payload.CountryCode,
		PhoneNumber: payload.PhoneNumber,
		Message:     payload.Message,
		Source:      payload.Source,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrContactSubmissionInvalid) {
			respondError(c, http.StatusBadRequest, "invalid contact submission")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to create contact submission")
		return
	}

	respondCreated(c, submission)
}

func (h *ContactSubmissionHandler) AdminList(c *gin.Context) {
	limit := parseIntDefault(c.Query("limit"), 50)
	offset := parseIntDefault(c.Query("offset"), 0)
	submissions, err := h.submissions.List(c.Request.Context(), limit, offset)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load contact submissions")
		return
	}
	respondOK(c, submissions)
}

func (h *ContactSubmissionHandler) AdminGet(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		respondError(c, http.StatusBadRequest, "invalid contact submission id")
		return
	}

	submission, err := h.submissions.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondError(c, http.StatusNotFound, "contact submission not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to load contact submission")
		return
	}
	respondOK(c, submission)
}
