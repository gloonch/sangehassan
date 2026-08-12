package handlers

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"sangehassan/back/internal/usecase"
)

func (h *OperationsHandler) CommercialTerms(c *gin.Context) {
	okOrError(c, operationResult(h.service.GetCommercialTerms(c.Request.Context(), actorID(c), c.Param("id"), false)))
}
func (h *OperationsHandler) AccountCommercialTerms(c *gin.Context) {
	okOrError(c, operationResult(h.service.GetCommercialTerms(c.Request.Context(), actorID(c), c.Param("id"), true)))
}
func (h *OperationsHandler) SaveCommercialTerms(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.CommercialTermsPayload](c)
	if !ok {
		return
	}
	okOrError(c, operationResult(h.service.SaveCommercialTerms(c.Request.Context(), actorID(c), c.Param("id"), key, p)))
}
func (h *OperationsHandler) PaymentSchedule(c *gin.Context) {
	okOrError(c, operationResult(h.service.ListPaymentSchedule(c.Request.Context(), actorID(c), c.Param("id"), false)))
}
func (h *OperationsHandler) AccountPaymentSchedule(c *gin.Context) {
	okOrError(c, operationResult(h.service.ListPaymentSchedule(c.Request.Context(), actorID(c), c.Param("id"), true)))
}
func (h *OperationsHandler) SavePaymentSchedule(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.PaymentSchedulePayload](c)
	if !ok {
		return
	}
	okOrError(c, operationResult(h.service.SavePaymentSchedule(c.Request.Context(), actorID(c), c.Param("id"), key, p)))
}
func (h *OperationsHandler) ConfirmOrder(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[struct {
		Reason string `json:"reason"`
	}](c)
	if !ok {
		return
	}
	okOrError(c, operationResult(h.service.ConfirmOrder(c.Request.Context(), actorID(c), c.Param("id"), key, p.Reason)))
}
func (h *OperationsHandler) RecordPayment(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.PaymentPayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.RecordPayment(c.Request.Context(), actorID(c), c.Param("id"), key, p)))
}
func (h *OperationsHandler) Payments(c *gin.Context) {
	okOrError(c, operationResult(h.service.ListPayments(c.Request.Context(), actorID(c), c.Param("id"), false)))
}
func (h *OperationsHandler) AccountPayments(c *gin.Context) {
	okOrError(c, operationResult(h.service.ListPayments(c.Request.Context(), actorID(c), c.Param("id"), true)))
}
func (h *OperationsHandler) PaymentAllocations(c *gin.Context) {
	okOrError(c, operationResult(h.service.PaymentAllocations(c.Request.Context(), c.Param("id"))))
}
func (h *OperationsHandler) ConfirmPayment(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.PaymentConfirmPayload](c)
	if !ok {
		return
	}
	okOrError(c, operationResult(h.service.ConfirmPayment(c.Request.Context(), actorID(c), c.Param("id"), key, p)))
}
func (h *OperationsHandler) RejectPayment(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[struct {
		Reason string `json:"reason"`
	}](c)
	if !ok {
		return
	}
	if err := h.service.RejectPayment(c.Request.Context(), actorID(c), c.Param("id"), key, p.Reason); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"rejected": true})
}
func (h *OperationsHandler) RefundPayment(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.PaymentRefundPayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.RefundPayment(c.Request.Context(), actorID(c), c.Param("id"), key, p)))
}
func (h *OperationsHandler) FinancialSummary(c *gin.Context) {
	okOrError(c, operationResult(h.service.FinancialSummary(c.Request.Context(), actorID(c), c.Param("id"), c.Query("reporting_currency"), false)))
}
func (h *OperationsHandler) AccountFinancialSummary(c *gin.Context) {
	okOrError(c, operationResult(h.service.FinancialSummary(c.Request.Context(), actorID(c), c.Param("id"), "", true)))
}
func (h *OperationsHandler) CreateCost(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.CostEntryPayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.CreateCost(c.Request.Context(), actorID(c), key, p)))
}
func (h *OperationsHandler) OrderCosts(c *gin.Context) {
	okOrError(c, operationResult(h.service.ListOrderCosts(c.Request.Context(), actorID(c), c.Param("id"))))
}
func costFlow(c *gin.Context, fn func(string, usecase.CostFlowPayload) (map[string]any, error)) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.CostFlowPayload](c)
	if !ok {
		return
	}
	okOrError(c, operationResult(fn(key, p)))
}
func (h *OperationsHandler) SubmitCost(c *gin.Context) {
	costFlow(c, func(k string, _ usecase.CostFlowPayload) (map[string]any, error) {
		return h.service.SubmitCost(c.Request.Context(), actorID(c), c.Param("id"), k)
	})
}
func (h *OperationsHandler) ApproveCost(c *gin.Context) {
	costFlow(c, func(k string, _ usecase.CostFlowPayload) (map[string]any, error) {
		return h.service.ApproveCost(c.Request.Context(), actorID(c), c.Param("id"), k)
	})
}
func (h *OperationsHandler) RejectCost(c *gin.Context) {
	costFlow(c, func(k string, p usecase.CostFlowPayload) (map[string]any, error) {
		return h.service.RejectCost(c.Request.Context(), actorID(c), c.Param("id"), k, p)
	})
}
func (h *OperationsHandler) MarkCostPaid(c *gin.Context) {
	costFlow(c, func(k string, p usecase.CostFlowPayload) (map[string]any, error) {
		return h.service.MarkCostPaid(c.Request.Context(), actorID(c), c.Param("id"), k, p)
	})
}
func (h *OperationsHandler) CancelCost(c *gin.Context) {
	costFlow(c, func(k string, p usecase.CostFlowPayload) (map[string]any, error) {
		return h.service.CancelCost(c.Request.Context(), actorID(c), c.Param("id"), k, p)
	})
}
func (h *OperationsHandler) ExchangeRates(c *gin.Context) {
	okOrError(c, operationResult(h.service.ListExchangeRates(c.Request.Context())))
}
func (h *OperationsHandler) SaveExchangeRate(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.ExchangeRatePayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.UpsertExchangeRate(c.Request.Context(), actorID(c), key, p)))
}

func (h *OperationsHandler) Notifications(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	okOrError(c, operationResult(h.service.ListNotifications(c.Request.Context(), actorID(c), limit, offset)))
}
func (h *OperationsHandler) ReadNotification(c *gin.Context) {
	if err := h.service.ReadNotification(c.Request.Context(), actorID(c), c.Param("id")); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"read": true})
}
func (h *OperationsHandler) ReadAllNotifications(c *gin.Context) {
	if err := h.service.ReadAllNotifications(c.Request.Context(), actorID(c)); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"read": true})
}
func (h *OperationsHandler) NotificationPreferences(c *gin.Context) {
	okOrError(c, operationResult(h.service.ListNotificationPreferences(c.Request.Context(), actorID(c))))
}
func (h *OperationsHandler) SaveNotificationPreferences(c *gin.Context) {
	p, ok := bindOperation[struct {
		Items []usecase.NotificationPreferencePayload `json:"items"`
	}](c)
	if !ok {
		return
	}
	if err := h.service.SaveNotificationPreferences(c.Request.Context(), actorID(c), p.Items); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"updated": true})
}
func (h *OperationsHandler) NotificationTemplates(c *gin.Context) {
	okOrError(c, operationResult(h.service.ListNotificationTemplates(c.Request.Context())))
}
func (h *OperationsHandler) UpdateNotificationTemplate(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.NotificationTemplateUpdate](c)
	if !ok {
		return
	}
	if err := h.service.UpdateNotificationTemplate(c.Request.Context(), actorID(c), c.Param("id"), key, p); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"updated": true})
}
func (h *OperationsHandler) RetryNotification(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	if err := h.service.RetryNotificationDelivery(c.Request.Context(), actorID(c), c.Param("id"), key); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"queued": true})
}
func (h *OperationsHandler) NotificationDeliveries(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	okOrError(c, operationResult(h.service.ListNotificationDeliveries(c.Request.Context(), c.Query("status"), limit)))
}

func (h *OperationsHandler) DocumentTemplates(c *gin.Context) {
	okOrError(c, operationResult(h.service.ListDocumentTemplates(c.Request.Context())))
}
func (h *OperationsHandler) UpdateDocumentTemplate(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.DocumentTemplateUpdate](c)
	if !ok {
		return
	}
	if err := h.service.UpdateDocumentTemplate(c.Request.Context(), actorID(c), c.Param("id"), key, p); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"updated": true})
}

func (h *OperationsHandler) Documents(c *gin.Context) {
	okOrError(c, operationResult(h.service.ListDocuments(c.Request.Context(), actorID(c), c.Param("id"), false)))
}
func (h *OperationsHandler) AccountDocuments(c *gin.Context) {
	okOrError(c, operationResult(h.service.ListDocuments(c.Request.Context(), actorID(c), c.Param("id"), true)))
}
func (h *OperationsHandler) GenerateDocument(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.DocumentGeneratePayload](c)
	if !ok {
		return
	}
	if p.ScopeType == "" {
		p.ScopeType = "ORDER"
	}
	if p.ScopeID == "" {
		p.ScopeID = c.Param("id")
	}
	createdOrError(c, operationResult(h.service.GenerateDocument(c.Request.Context(), actorID(c), c.Param("id"), key, p)))
}
func (h *OperationsHandler) IssueDocument(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	okOrError(c, operationResult(h.service.IssueDocument(c.Request.Context(), actorID(c), c.Param("id"), key)))
}
func (h *OperationsHandler) CancelDocument(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[struct {
		Reason string `json:"reason"`
	}](c)
	if !ok {
		return
	}
	if err := h.service.CancelDocument(c.Request.Context(), actorID(c), c.Param("id"), key, p.Reason); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"cancelled": true})
}
func (h *OperationsHandler) DownloadDocument(c *gin.Context)        { serveDocument(c, h, false) }
func (h *OperationsHandler) DownloadAccountDocument(c *gin.Context) { serveDocument(c, h, true) }
func serveDocument(c *gin.Context, h *OperationsHandler, customer bool) {
	path, name, err := h.service.DocumentDownload(c.Request.Context(), actorID(c), c.Param("id"), customer)
	if err != nil {
		operationError(c, err)
		return
	}
	f, err := os.Open(path)
	if err != nil {
		respondError(c, http.StatusNotFound, "file not found")
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		respondError(c, http.StatusNotFound, "file not found")
		return
	}
	headers := map[string]string{"Content-Disposition": mime.FormatMediaType("inline", map[string]string{"filename": filepath.Base(name)}), "X-Content-Type-Options": "nosniff", "Cache-Control": "private, no-store"}
	c.DataFromReader(http.StatusOK, info.Size(), "application/pdf", f, headers)
}

func (h *OperationsHandler) DocumentRequirements(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, 400, "invalid id")
		return
	}
	okOrError(c, operationResult(h.service.ListDocumentRequirements(c.Request.Context(), id)))
}
func (h *OperationsHandler) SaveDocumentRequirement(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, 400, "invalid id")
		return
	}
	p, ok := bindOperation[usecase.DocumentRequirementPayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.SaveDocumentRequirement(c.Request.Context(), actorID(c), id, p)))
}
func (h *OperationsHandler) DeleteDocumentRequirement(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, 400, "invalid id")
		return
	}
	if err = h.service.DeleteDocumentRequirement(c.Request.Context(), actorID(c), id, c.Param("requirementId")); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"deleted": true})
}

func (h *OperationsHandler) ReportOverview(c *gin.Context) {
	okOrError(c, operationResult(h.service.ReportOverview(c.Request.Context())))
}
func (h *OperationsHandler) ReportReceivables(c *gin.Context) {
	okOrError(c, operationResult(h.service.ReportReceivables(c.Request.Context())))
}
func (h *OperationsHandler) ReportCosts(c *gin.Context) {
	okOrError(c, operationResult(h.service.ReportCosts(c.Request.Context())))
}
func (h *OperationsHandler) ReportProfitability(c *gin.Context) {
	okOrError(c, operationResult(h.service.ReportProfitability(c.Request.Context(), c.Query("reporting_currency"))))
}
func (h *OperationsHandler) ReportOperations(c *gin.Context) {
	okOrError(c, operationResult(h.service.ReportOperations(c.Request.Context())))
}
func (h *OperationsHandler) ReportSales(c *gin.Context) {
	okOrError(c, operationResult(h.service.ReportSales(c.Request.Context())))
}
func (h *OperationsHandler) CustomerContactLogs(c *gin.Context) {
	okOrError(c, operationResult(h.service.ListContactLogs(c.Request.Context(), c.Param("id"), "")))
}
func (h *OperationsHandler) OrderContactLogs(c *gin.Context) {
	okOrError(c, operationResult(h.service.ListContactLogs(c.Request.Context(), "", c.Param("id"))))
}
func (h *OperationsHandler) CreateCustomerContactLog(c *gin.Context) {
	p, ok := bindOperation[usecase.ContactLogPayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.CreateContactLog(c.Request.Context(), actorID(c), c.Param("id"), "", p)))
}
func (h *OperationsHandler) CreateOrderContactLog(c *gin.Context) {
	p, ok := bindOperation[usecase.ContactLogPayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.CreateContactLog(c.Request.Context(), actorID(c), "", c.Param("id"), p)))
}
