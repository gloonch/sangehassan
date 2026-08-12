package handlers

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"sangehassan/back/internal/usecase"
)

func (h *OperationsHandler) DashboardHome(c *gin.Context) {
	v, err := h.service.DashboardHome(c.Request.Context(), actorID(c))
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, v)
}
func (h *OperationsHandler) FeatureFlags(c *gin.Context) {
	respondOK(c, h.service.FeatureFlags(c.Request.Context()))
}
func (h *OperationsHandler) GlobalSearch(c *gin.Context) {
	v, err := h.service.GlobalSearch(c.Request.Context(), actorID(c), c.Query("q"))
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, v)
}

type savedViewPayload struct {
	ViewKey string         `json:"view_key"`
	Name    string         `json:"name"`
	Filters map[string]any `json:"filters"`
}

func (h *OperationsHandler) SavedViews(c *gin.Context) {
	v, err := h.service.ListSavedViews(c.Request.Context(), actorID(c), c.Query("viewKey"))
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, v)
}
func (h *OperationsHandler) CreateSavedView(c *gin.Context) {
	var p savedViewPayload
	if c.ShouldBindJSON(&p) != nil {
		respondError(c, 400, "اطلاعات نما معتبر نیست.")
		return
	}
	v, err := h.service.SaveSavedView(c.Request.Context(), actorID(c), "", p.ViewKey, p.Name, p.Filters)
	if err != nil {
		operationError(c, err)
		return
	}
	respondCreated(c, v)
}
func (h *OperationsHandler) UpdateSavedView(c *gin.Context) {
	var p savedViewPayload
	if c.ShouldBindJSON(&p) != nil {
		respondError(c, 400, "اطلاعات نما معتبر نیست.")
		return
	}
	v, err := h.service.SaveSavedView(c.Request.Context(), actorID(c), c.Param("id"), p.ViewKey, p.Name, p.Filters)
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, v)
}
func (h *OperationsHandler) DeleteSavedView(c *gin.Context) {
	if err := h.service.DeleteSavedView(c.Request.Context(), actorID(c), c.Param("id")); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"deleted": true})
}

func (h *OperationsHandler) Settings(c *gin.Context) {
	v, err := h.service.ListApplicationSettings(c.Request.Context())
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, v)
}

type settingPayload struct {
	Key    string          `json:"key"`
	Value  json.RawMessage `json:"value"`
	Reason string          `json:"reason"`
}

func (h *OperationsHandler) UpdateSetting(c *gin.Context) {
	var p settingPayload
	if c.ShouldBindJSON(&p) != nil {
		respondError(c, 400, "اطلاعات تنظیم معتبر نیست.")
		return
	}
	v, err := h.service.UpdateApplicationSetting(c.Request.Context(), actorID(c), p.Key, p.Value, p.Reason)
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, v)
}
func (h *OperationsHandler) SystemInfo(c *gin.Context) {
	v, err := h.service.SystemInfo(c.Request.Context(), h.version, h.commit, h.buildTime, h.environment)
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, v)
}
func (h *OperationsHandler) Version(c *gin.Context) {
	v, err := h.service.SystemInfo(c.Request.Context(), h.version, h.commit, h.buildTime, h.environment)
	if err != nil {
		respondError(c, 503, "اطلاعات نسخه در دسترس نیست.")
		return
	}
	respondOK(c, v)
}

func (h *OperationsHandler) WorkflowDiagnostics(c *gin.Context) {
	v, err := h.service.WorkflowDiagnostics(c.Request.Context(), c.Param("id"))
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, v)
}

type repairPayload struct {
	RepairCode          string     `json:"repair_code"`
	Reason              string     `json:"reason"`
	EstimatedDeliveryAt *time.Time `json:"estimated_delivery_at"`
}

func requireIdempotency(c *gin.Context) string {
	key := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	if key == "" {
		respondErrorCode(c, 400, "IDEMPOTENCY_KEY_REQUIRED", "ارسال Idempotency-Key الزامی است.", nil)
	}
	return key
}
func (h *OperationsHandler) RepairWorkflow(c *gin.Context) {
	key := requireIdempotency(c)
	if key == "" {
		return
	}
	var p repairPayload
	if c.ShouldBindJSON(&p) != nil {
		respondError(c, 400, "اطلاعات Repair معتبر نیست.")
		return
	}
	v, err := h.service.RepairWorkflow(c.Request.Context(), actorID(c), c.Param("id"), p.RepairCode, p.Reason, key)
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, v)
}
func (h *OperationsHandler) CorrectEstimatedDelivery(c *gin.Context) {
	key := requireIdempotency(c)
	if key == "" {
		return
	}
	var p repairPayload
	if c.ShouldBindJSON(&p) != nil {
		respondError(c, 400, "اطلاعات اصلاح معتبر نیست.")
		return
	}
	v, err := h.service.CorrectOrderEstimatedDelivery(c.Request.Context(), actorID(c), c.Param("id"), p.EstimatedDeliveryAt, p.Reason, key)
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, v)
}
func (h *OperationsHandler) ResolveStuckAction(c *gin.Context) {
	key := requireIdempotency(c)
	if key == "" {
		return
	}
	var p repairPayload
	if c.ShouldBindJSON(&p) != nil {
		respondError(c, 400, "اطلاعات اصلاح معتبر نیست.")
		return
	}
	v, err := h.service.ResolveStuckAction(c.Request.Context(), actorID(c), c.Param("id"), p.Reason, key)
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, v)
}
func (h *OperationsHandler) ReconcilePayment(c *gin.Context) {
	key := requireIdempotency(c)
	if key == "" {
		return
	}
	var p repairPayload
	if c.ShouldBindJSON(&p) != nil {
		respondError(c, 400, "اطلاعات اصلاح معتبر نیست.")
		return
	}
	v, err := h.service.ReconcileOrderPayment(c.Request.Context(), actorID(c), c.Param("id"), p.Reason, key)
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, v)
}
func (h *OperationsHandler) RecalculateProgress(c *gin.Context) {
	key := requireIdempotency(c)
	if key == "" {
		return
	}
	var p repairPayload
	if c.ShouldBindJSON(&p) != nil || strings.TrimSpace(p.Reason) == "" {
		respondErrorCode(c, 400, "REASON_REQUIRED", "ثبت دلیل اصلاح الزامی است.", nil)
		return
	}
	v, err := h.service.RecalculateOrderProgress(c.Request.Context(), actorID(c), c.Param("id"), p.Reason, key)
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, v)
}
func (h *OperationsHandler) RevokeSessions(c *gin.Context) {
	key := requireIdempotency(c)
	if key == "" {
		return
	}
	var p repairPayload
	if c.ShouldBindJSON(&p) != nil {
		respondError(c, 400, "اطلاعات درخواست معتبر نیست.")
		return
	}
	v, err := h.service.RevokeUserSessions(c.Request.Context(), actorID(c), c.Param("id"), p.Reason, key)
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, v)
}

func (h *OperationsHandler) AuditPage(c *gin.Context) {
	page := usecase.ParsePage(c.Query("page"), c.Query("pageSize"))
	v, err := h.service.AuditLogsPage(c.Request.Context(), c.Query("search"), c.Query("actor"), c.Query("entity"), c.Query("action"), c.Query("order"), c.Query("from"), c.Query("to"), page)
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, v)
}

func (h *OperationsHandler) ExportCSV(c *gin.Context) {
	kind := c.Param("kind")
	if !h.service.HasPermission(c.Request.Context(), actorID(c), "exports."+kind) {
		respondErrorCode(c, http.StatusForbidden, "PERMISSION_DENIED", "شما دسترسی لازم برای این خروجی را ندارید.", nil)
		return
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	headers, rows, err := h.service.ExportRows(c.Request.Context(), kind, c.Query("search"), c.Query("status"), limit)
	if err != nil {
		operationError(c, err)
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="`+kind+`-export.csv"`)
	c.Status(http.StatusOK)
	_, _ = c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})
	writer := csv.NewWriter(c.Writer)
	_ = writer.Write(headers)
	for _, row := range rows {
		for index, value := range row {
			trimmed := strings.TrimLeft(value, " \t\r\n")
			if trimmed != "" && strings.ContainsRune("=+-@", rune(trimmed[0])) {
				row[index] = "'" + value
			}
		}
		_ = writer.Write(row)
	}
	writer.Flush()
}
