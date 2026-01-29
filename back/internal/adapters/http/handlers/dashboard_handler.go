package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"sangehassan/back/internal/usecase"
)

type DashboardHandler struct {
	service *usecase.DashboardService
}

func NewDashboardHandler(service *usecase.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (h *DashboardHandler) Stats(c *gin.Context) {
	stats, err := h.service.GetStats(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load dashboard stats")
		return
	}
	respondOK(c, stats)
}
