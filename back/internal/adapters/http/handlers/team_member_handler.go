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

type TeamMemberHandler struct {
	service *usecase.TeamMemberService
}

func NewTeamMemberHandler(service *usecase.TeamMemberService) *TeamMemberHandler {
	return &TeamMemberHandler{service: service}
}

type teamMemberPayload struct {
	NameEN      string `json:"name_en"`
	NameFA      string `json:"name_fa"`
	NameAR      string `json:"name_ar"`
	RoleEN      string `json:"role_en"`
	RoleFA      string `json:"role_fa"`
	RoleAR      string `json:"role_ar"`
	BioEN       string `json:"bio_en"`
	BioFA       string `json:"bio_fa"`
	BioAR       string `json:"bio_ar"`
	PhotoURL    string `json:"photo_url"`
	LinkedInURL string `json:"linkedin_url"`
	OrderIndex  int    `json:"order_index"`
	IsActive    bool   `json:"is_active"`
}

func (h *TeamMemberHandler) ListPublic(c *gin.Context) {
	members, err := h.service.ListPublic(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load team members")
		return
	}
	respondOK(c, members)
}

func (h *TeamMemberHandler) List(c *gin.Context) {
	members, err := h.service.List(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load team members")
		return
	}
	respondOK(c, members)
}

func (h *TeamMemberHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid team member id")
		return
	}

	member, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondError(c, http.StatusNotFound, "team member not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to load team member")
		return
	}
	respondOK(c, member)
}

func (h *TeamMemberHandler) Create(c *gin.Context) {
	var payload teamMemberPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}
	if !isAllowedImageRef(payload.PhotoURL) {
		respondError(c, http.StatusBadRequest, "photo must be uploaded")
		return
	}

	member, err := h.service.Create(c.Request.Context(), domain.TeamMember{
		NameEN:      payload.NameEN,
		NameFA:      payload.NameFA,
		NameAR:      payload.NameAR,
		RoleEN:      payload.RoleEN,
		RoleFA:      payload.RoleFA,
		RoleAR:      payload.RoleAR,
		BioEN:       payload.BioEN,
		BioFA:       payload.BioFA,
		BioAR:       payload.BioAR,
		PhotoURL:    normalizeImageRef(payload.PhotoURL),
		LinkedInURL: payload.LinkedInURL,
		OrderIndex:  payload.OrderIndex,
		IsActive:    payload.IsActive,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create team member")
		return
	}
	respondCreated(c, member)
}

func (h *TeamMemberHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid team member id")
		return
	}

	var payload teamMemberPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}
	if !isAllowedImageRef(payload.PhotoURL) {
		respondError(c, http.StatusBadRequest, "photo must be uploaded")
		return
	}

	member, err := h.service.Update(c.Request.Context(), domain.TeamMember{
		ID:          id,
		NameEN:      payload.NameEN,
		NameFA:      payload.NameFA,
		NameAR:      payload.NameAR,
		RoleEN:      payload.RoleEN,
		RoleFA:      payload.RoleFA,
		RoleAR:      payload.RoleAR,
		BioEN:       payload.BioEN,
		BioFA:       payload.BioFA,
		BioAR:       payload.BioAR,
		PhotoURL:    normalizeImageRef(payload.PhotoURL),
		LinkedInURL: payload.LinkedInURL,
		OrderIndex:  payload.OrderIndex,
		IsActive:    payload.IsActive,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to update team member")
		return
	}
	respondOK(c, member)
}

func (h *TeamMemberHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid team member id")
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to delete team member")
		return
	}
	c.Status(http.StatusNoContent)
}
