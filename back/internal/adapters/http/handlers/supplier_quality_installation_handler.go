package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"sangehassan/back/internal/usecase"
)

func (h *OperationsHandler) Suppliers(c *gin.Context) {
	okOrError(c, operationResult(h.service.ListSuppliers(c.Request.Context(), c.Query("search"), c.Query("type"), c.Query("include_inactive") == "true")))
}
func (h *OperationsHandler) Supplier(c *gin.Context) {
	okOrError(c, operationResult(h.service.GetSupplier(c.Request.Context(), c.Param("id"))))
}
func (h *OperationsHandler) CreateSupplier(c *gin.Context) {
	p, ok := bindOperation[usecase.SupplierPayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.CreateSupplier(c.Request.Context(), actorID(c), p)))
}
func (h *OperationsHandler) UpdateSupplier(c *gin.Context) {
	p, ok := bindOperation[usecase.SupplierPayload](c)
	if !ok {
		return
	}
	okOrError(c, operationResult(h.service.UpdateSupplier(c.Request.Context(), actorID(c), c.Param("id"), p)))
}
func (h *OperationsHandler) DisableSupplier(c *gin.Context) {
	if err := h.service.DisableSupplier(c.Request.Context(), actorID(c), c.Param("id")); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"disabled": true})
}

func (h *OperationsHandler) Purchases(c *gin.Context) {
	okOrError(c, operationResult(h.service.ListPurchases(c.Request.Context(), actorID(c), c.Query("status"), c.Query("order_id"), c.Query("supplier_id"))))
}
func (h *OperationsHandler) Purchase(c *gin.Context) {
	okOrError(c, operationResult(h.service.GetPurchase(c.Request.Context(), actorID(c), c.Param("id"))))
}
func (h *OperationsHandler) CreatePurchase(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.PurchasePayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.CreatePurchase(c.Request.Context(), actorID(c), c.Param("id"), key, p)))
}
func (h *OperationsHandler) UpdatePurchase(c *gin.Context) {
	p, ok := bindOperation[usecase.PurchasePayload](c)
	if !ok {
		return
	}
	okOrError(c, operationResult(h.service.UpdatePurchase(c.Request.Context(), actorID(c), c.Param("id"), p)))
}
func (h *OperationsHandler) ConfirmPurchase(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	okOrError(c, operationResult(h.service.ConfirmPurchase(c.Request.Context(), actorID(c), c.Param("id"), key)))
}
func (h *OperationsHandler) ReceivePurchase(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.PurchaseReceiptPayload](c)
	if !ok {
		return
	}
	okOrError(c, operationResult(h.service.ReceivePurchase(c.Request.Context(), actorID(c), c.Param("id"), key, p)))
}
func (h *OperationsHandler) CancelPurchase(c *gin.Context) {
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
	okOrError(c, operationResult(h.service.CancelPurchase(c.Request.Context(), actorID(c), c.Param("id"), key, p.Reason)))
}

func (h *OperationsHandler) QualityInspections(c *gin.Context) {
	okOrError(c, operationResult(h.service.ListQualityInspections(c.Request.Context(), actorID(c), c.Query("status"), c.Query("order_id"), c.Query("batch_id"))))
}
func (h *OperationsHandler) QualityInspection(c *gin.Context) {
	okOrError(c, operationResult(h.service.GetQualityInspection(c.Request.Context(), actorID(c), c.Param("id"))))
}
func (h *OperationsHandler) CreateQualityInspection(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.QualityInspectionPayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.CreateQualityInspection(c.Request.Context(), actorID(c), key, p)))
}
func qualityDecision(c *gin.Context, fn func(string, usecase.QualityDecisionPayload) (map[string]any, error)) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.QualityDecisionPayload](c)
	if !ok {
		return
	}
	okOrError(c, operationResult(fn(key, p)))
}
func (h *OperationsHandler) PassQualityInspection(c *gin.Context) {
	qualityDecision(c, func(k string, p usecase.QualityDecisionPayload) (map[string]any, error) {
		return h.service.PassQualityInspection(c.Request.Context(), actorID(c), c.Param("id"), k, p)
	})
}
func (h *OperationsHandler) FailQualityInspection(c *gin.Context) {
	qualityDecision(c, func(k string, p usecase.QualityDecisionPayload) (map[string]any, error) {
		return h.service.FailQualityInspection(c.Request.Context(), actorID(c), c.Param("id"), k, p)
	})
}
func (h *OperationsHandler) RequestQualityRework(c *gin.Context) {
	qualityDecision(c, func(k string, p usecase.QualityDecisionPayload) (map[string]any, error) {
		return h.service.RequestQualityRework(c.Request.Context(), actorID(c), c.Param("id"), k, p)
	})
}
func (h *OperationsHandler) RejectQualityInspection(c *gin.Context) {
	qualityDecision(c, func(k string, p usecase.QualityDecisionPayload) (map[string]any, error) {
		return h.service.RejectQualityInspection(c.Request.Context(), actorID(c), c.Param("id"), k, p)
	})
}
func (h *OperationsHandler) OverrideQualityInspection(c *gin.Context) {
	qualityDecision(c, func(k string, p usecase.QualityDecisionPayload) (map[string]any, error) {
		return h.service.OverrideQualityInspection(c.Request.Context(), actorID(c), c.Param("id"), k, p)
	})
}

func (h *OperationsHandler) Installations(c *gin.Context) {
	okOrError(c, operationResult(h.service.ListInstallations(c.Request.Context(), actorID(c), c.Query("status"), c.Query("order_id"))))
}
func (h *OperationsHandler) Installation(c *gin.Context) {
	okOrError(c, operationResult(h.service.GetInstallation(c.Request.Context(), actorID(c), c.Param("id"), false)))
}
func (h *OperationsHandler) CreateInstallation(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.InstallationPayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.CreateInstallation(c.Request.Context(), actorID(c), c.Param("id"), key, p)))
}
func (h *OperationsHandler) UpdateInstallation(c *gin.Context) {
	p, ok := bindOperation[usecase.InstallationPayload](c)
	if !ok {
		return
	}
	if err := h.service.UpdateInstallation(c.Request.Context(), actorID(c), c.Param("id"), p); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"updated": true})
}
func (h *OperationsHandler) ReplaceInstallationMembers(c *gin.Context) {
	p, ok := bindOperation[struct {
		Members []usecase.InstallationMemberPayload `json:"members"`
	}](c)
	if !ok {
		return
	}
	if err := h.service.ReplaceInstallationMembers(c.Request.Context(), actorID(c), c.Param("id"), p.Members); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"updated": true})
}
func installationFlow(c *gin.Context, fn func(string, usecase.InstallationFlowPayload) (map[string]any, error)) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.InstallationFlowPayload](c)
	if !ok {
		return
	}
	okOrError(c, operationResult(fn(key, p)))
}
func (h *OperationsHandler) StartInstallation(c *gin.Context) {
	installationFlow(c, func(k string, p usecase.InstallationFlowPayload) (map[string]any, error) {
		return h.service.StartInstallation(c.Request.Context(), actorID(c), c.Param("id"), k, p)
	})
}
func (h *OperationsHandler) PauseInstallation(c *gin.Context) {
	installationFlow(c, func(k string, p usecase.InstallationFlowPayload) (map[string]any, error) {
		return h.service.PauseInstallation(c.Request.Context(), actorID(c), c.Param("id"), k, p)
	})
}
func (h *OperationsHandler) CompleteInstallation(c *gin.Context) {
	installationFlow(c, func(k string, p usecase.InstallationFlowPayload) (map[string]any, error) {
		return h.service.CompleteInstallation(c.Request.Context(), actorID(c), c.Param("id"), k, p)
	})
}
func (h *OperationsHandler) CancelInstallation(c *gin.Context) {
	installationFlow(c, func(k string, p usecase.InstallationFlowPayload) (map[string]any, error) {
		return h.service.CancelInstallation(c.Request.Context(), actorID(c), c.Param("id"), k, p)
	})
}
func (h *OperationsHandler) AddInstallationUpdate(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.InstallationUpdatePayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.AddInstallationUpdate(c.Request.Context(), actorID(c), c.Param("id"), key, p)))
}
func (h *OperationsHandler) AddInstallationIssue(c *gin.Context) {
	p, ok := bindOperation[usecase.InstallationIssuePayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.AddInstallationIssue(c.Request.Context(), actorID(c), c.Param("id"), p)))
}
func (h *OperationsHandler) ResolveInstallationIssue(c *gin.Context) {
	p, ok := bindOperation[usecase.InstallationIssuePayload](c)
	if !ok {
		return
	}
	if err := h.service.ResolveInstallationIssue(c.Request.Context(), actorID(c), c.Param("issueId"), p); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"resolved": true})
}
func (h *OperationsHandler) AddInstallationMaterial(c *gin.Context) {
	p, ok := bindOperation[usecase.InstallationMaterialPayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.AddInstallationMaterial(c.Request.Context(), actorID(c), c.Param("id"), p)))
}
func (h *OperationsHandler) AccountInstallation(c *gin.Context) {
	okOrError(c, operationResult(h.service.AccountInstallation(c.Request.Context(), actorID(c), c.Param("id"))))
}

func (h *OperationsHandler) RecordCustomerAcceptance(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.CustomerAcceptancePayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.RecordCustomerAcceptance(c.Request.Context(), actorID(c), c.Param("id"), key, false, p)))
}
func (h *OperationsHandler) AccountCustomerAcceptance(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.CustomerAcceptancePayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.RecordCustomerAcceptance(c.Request.Context(), actorID(c), c.Param("id"), key, true, p)))
}
func (h *OperationsHandler) OrderClosureReadiness(c *gin.Context) {
	okOrError(c, operationResult(h.service.OrderClosureReadiness(c.Request.Context(), c.Param("id"))))
}
func (h *OperationsHandler) CloseOrder(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.OrderClosurePayload](c)
	if !ok {
		return
	}
	okOrError(c, operationResult(h.service.CloseOrder(c.Request.Context(), actorID(c), c.Param("id"), key, p)))
}

var _ = http.StatusOK
