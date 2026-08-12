package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"sangehassan/back/internal/usecase"
)

func int64Param(c *gin.Context, key string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(key), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

func bindOperation[T any](c *gin.Context) (T, bool) {
	var payload T
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return payload, false
	}
	return payload, true
}

func (h *OperationsHandler) WorkflowTemplateVersions(c *gin.Context) {
	items, err := h.service.ListWorkflowTemplateVersions(c.Request.Context())
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, items)
}

func (h *OperationsHandler) WorkflowStepCatalogue(c *gin.Context) {
	items, err := h.service.ListWorkflowStepCatalogue(c.Request.Context())
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, items)
}

func (h *OperationsHandler) WorkflowTemplateVersion(c *gin.Context) {
	id, ok := int64Param(c, "id")
	if !ok {
		return
	}
	item, err := h.service.GetWorkflowTemplateVersion(c.Request.Context(), id)
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, item)
}

func (h *OperationsHandler) CreateWorkflowTemplate(c *gin.Context) {
	p, ok := bindOperation[usecase.WorkflowTemplatePayload](c)
	if !ok {
		return
	}
	item, err := h.service.CreateWorkflowTemplateDraft(c.Request.Context(), actorID(c), p)
	if err != nil {
		operationError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": item})
}

func (h *OperationsHandler) UpdateWorkflowTemplate(c *gin.Context) {
	id, ok := int64Param(c, "id")
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.WorkflowTemplatePayload](c)
	if !ok {
		return
	}
	if err := h.service.UpdateWorkflowTemplateDraft(c.Request.Context(), actorID(c), id, p); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"updated": true})
}

func (h *OperationsHandler) CloneWorkflowTemplate(c *gin.Context) {
	id, ok := int64Param(c, "id")
	if !ok {
		return
	}
	item, err := h.service.CloneWorkflowTemplate(c.Request.Context(), actorID(c), id)
	if err != nil {
		operationError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": item})
}

func (h *OperationsHandler) PublishWorkflowTemplate(c *gin.Context) {
	id, ok := int64Param(c, "id")
	if !ok {
		return
	}
	if err := h.service.PublishWorkflowTemplate(c.Request.Context(), actorID(c), id); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"published": true})
}

func (h *OperationsHandler) ArchiveWorkflowTemplate(c *gin.Context) {
	id, ok := int64Param(c, "id")
	if !ok {
		return
	}
	if err := h.service.ArchiveWorkflowTemplate(c.Request.Context(), actorID(c), id); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"archived": true})
}

func (h *OperationsHandler) AddWorkflowStep(c *gin.Context) {
	templateID, ok := int64Param(c, "id")
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.WorkflowStepPayload](c)
	if !ok {
		return
	}
	item, err := h.service.AddWorkflowStep(c.Request.Context(), actorID(c), templateID, p)
	if err != nil {
		operationError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": item})
}

func (h *OperationsHandler) UpdateWorkflowStep(c *gin.Context) {
	id, ok := int64Param(c, "stepId")
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.WorkflowStepPayload](c)
	if !ok {
		return
	}
	if err := h.service.UpdateWorkflowStep(c.Request.Context(), actorID(c), id, p); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"updated": true})
}

func (h *OperationsHandler) DeleteWorkflowStep(c *gin.Context) {
	id, ok := int64Param(c, "stepId")
	if !ok {
		return
	}
	if err := h.service.DeleteWorkflowStep(c.Request.Context(), actorID(c), id); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"deleted": true})
}

type reorderPayload struct {
	IDs []int64 `json:"ids"`
}

func (h *OperationsHandler) ReorderWorkflowSteps(c *gin.Context) {
	id, ok := int64Param(c, "id")
	if !ok {
		return
	}
	p, ok := bindOperation[reorderPayload](c)
	if !ok {
		return
	}
	if err := h.service.ReorderWorkflowSteps(c.Request.Context(), actorID(c), id, p.IDs); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"reordered": true})
}

func (h *OperationsHandler) DuplicateWorkflowStep(c *gin.Context) {
	id, ok := int64Param(c, "stepId")
	if !ok {
		return
	}
	item, err := h.service.DuplicateWorkflowStep(c.Request.Context(), actorID(c), id)
	if err != nil {
		operationError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": item})
}

func (h *OperationsHandler) AddWorkflowField(c *gin.Context) {
	stepID, ok := int64Param(c, "stepId")
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.WorkflowFieldPayload](c)
	if !ok {
		return
	}
	item, err := h.service.AddWorkflowField(c.Request.Context(), actorID(c), stepID, p)
	if err != nil {
		operationError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": item})
}

func (h *OperationsHandler) UpdateWorkflowField(c *gin.Context) {
	id, ok := int64Param(c, "fieldId")
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.WorkflowFieldPayload](c)
	if !ok {
		return
	}
	if err := h.service.UpdateWorkflowField(c.Request.Context(), actorID(c), id, p); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"updated": true})
}

func (h *OperationsHandler) DeleteWorkflowField(c *gin.Context) {
	id, ok := int64Param(c, "fieldId")
	if !ok {
		return
	}
	if err := h.service.DeleteWorkflowField(c.Request.Context(), actorID(c), id); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"deleted": true})
}

func (h *OperationsHandler) ReorderWorkflowFields(c *gin.Context) {
	stepID, ok := int64Param(c, "stepId")
	if !ok {
		return
	}
	p, ok := bindOperation[reorderPayload](c)
	if !ok {
		return
	}
	if err := h.service.ReorderWorkflowFields(c.Request.Context(), actorID(c), stepID, p.IDs); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"reordered": true})
}

func (h *OperationsHandler) AddWorkflowTask(c *gin.Context) {
	stepID, ok := int64Param(c, "stepId")
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.WorkflowTaskPayload](c)
	if !ok {
		return
	}
	item, err := h.service.AddWorkflowTask(c.Request.Context(), actorID(c), stepID, p)
	if err != nil {
		operationError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": item})
}

func (h *OperationsHandler) UpdateWorkflowTask(c *gin.Context) {
	id, ok := int64Param(c, "taskId")
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.WorkflowTaskPayload](c)
	if !ok {
		return
	}
	if err := h.service.UpdateWorkflowTask(c.Request.Context(), actorID(c), id, p); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"updated": true})
}

func (h *OperationsHandler) DeleteWorkflowTask(c *gin.Context) {
	id, ok := int64Param(c, "taskId")
	if !ok {
		return
	}
	if err := h.service.DeleteWorkflowTask(c.Request.Context(), actorID(c), id); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"deleted": true})
}

func (h *OperationsHandler) AddWorkflowLevelTask(c *gin.Context) {
	templateID, ok := int64Param(c, "id")
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.WorkflowLevelTaskPayload](c)
	if !ok {
		return
	}
	item, err := h.service.AddWorkflowLevelTask(c.Request.Context(), actorID(c), templateID, p)
	if err != nil {
		operationError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": item})
}

func (h *OperationsHandler) UpdateWorkflowLevelTask(c *gin.Context) {
	id, ok := int64Param(c, "taskId")
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.WorkflowLevelTaskPayload](c)
	if !ok {
		return
	}
	if err := h.service.UpdateWorkflowLevelTask(c.Request.Context(), actorID(c), id, p); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"updated": true})
}

func (h *OperationsHandler) DeleteWorkflowLevelTask(c *gin.Context) {
	id, ok := int64Param(c, "taskId")
	if !ok {
		return
	}
	if err := h.service.DeleteWorkflowLevelTask(c.Request.Context(), actorID(c), id); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"deleted": true})
}

func (h *OperationsHandler) AddHandoffMetric(c *gin.Context) {
	templateID, ok := int64Param(c, "id")
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.HandoffMetricPayload](c)
	if !ok {
		return
	}
	item, err := h.service.AddHandoffMetric(c.Request.Context(), actorID(c), templateID, p)
	if err != nil {
		operationError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": item})
}

func (h *OperationsHandler) UpdateHandoffMetric(c *gin.Context) {
	id, ok := int64Param(c, "metricId")
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.HandoffMetricPayload](c)
	if !ok {
		return
	}
	if err := h.service.UpdateHandoffMetric(c.Request.Context(), actorID(c), id, p); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"updated": true})
}

func (h *OperationsHandler) DeleteHandoffMetric(c *gin.Context) {
	id, ok := int64Param(c, "metricId")
	if !ok {
		return
	}
	if err := h.service.DeleteHandoffMetric(c.Request.Context(), actorID(c), id); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"deleted": true})
}

func (h *OperationsHandler) WorkflowRuntime(c *gin.Context) {
	item, err := h.service.GetWorkflowRuntime(c.Request.Context(), actorID(c), c.Param("id"))
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, item)
}

func (h *OperationsHandler) WorkflowDashboard(c *gin.Context) {
	item, err := h.service.WorkflowDashboard(c.Request.Context(), actorID(c))
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, item)
}

func (h *OperationsHandler) StartWorkflowStep(c *gin.Context) {
	p, ok := bindOperation[usecase.StepOperationPayload](c)
	if !ok {
		return
	}
	if err := h.service.StartWorkflowStep(c.Request.Context(), actorID(c), c.Param("id"), p.Reason); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"started": true})
}

func (h *OperationsHandler) DraftWorkflowStep(c *gin.Context) {
	p, ok := bindOperation[usecase.StepValuesPayload](c)
	if !ok {
		return
	}
	if err := h.service.SaveWorkflowStepDraft(c.Request.Context(), actorID(c), c.Param("id"), p); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"saved": true})
}

func (h *OperationsHandler) SubmitWorkflowStep(c *gin.Context) {
	p, ok := bindOperation[usecase.StepValuesPayload](c)
	if !ok {
		return
	}
	if err := h.service.SubmitWorkflowStep(c.Request.Context(), actorID(c), c.Param("id"), p); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"submitted": true})
}

func (h *OperationsHandler) ApproveWorkflowStep(c *gin.Context) {
	h.stepReason(c, h.service.ApproveWorkflowStep, "approved")
}
func (h *OperationsHandler) RejectWorkflowStep(c *gin.Context) {
	h.stepReason(c, h.service.RejectWorkflowStep, "rejected")
}
func (h *OperationsHandler) SkipWorkflowStep(c *gin.Context) {
	h.stepReason(c, h.service.SkipWorkflowStep, "skipped")
}
func (h *OperationsHandler) ReopenWorkflowStep(c *gin.Context) {
	h.stepReason(c, h.service.ReopenWorkflowStep, "reopened")
}

func (h *OperationsHandler) stepReason(c *gin.Context, operation func(context.Context, string, string, string) error, key string) {
	p, ok := bindOperation[usecase.StepOperationPayload](c)
	if !ok {
		return
	}
	if err := operation(c.Request.Context(), actorID(c), c.Param("id"), p.Reason); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{key: true})
}

func (h *OperationsHandler) ReassignWorkflowStep(c *gin.Context) {
	p, ok := bindOperation[usecase.StepOperationPayload](c)
	if !ok {
		return
	}
	if err := h.service.ReassignWorkflowStep(c.Request.Context(), actorID(c), c.Param("id"), p.AssignedUserID, p.Reason); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"reassigned": true})
}

type discrepancyPayload struct {
	Status string `json:"status"`
	Note   string `json:"note"`
}

func (h *OperationsHandler) WorkflowDiscrepancies(c *gin.Context) {
	items, err := h.service.ListWorkflowDiscrepancies(c.Request.Context(), actorID(c), c.Param("id"))
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, items)
}

func (h *OperationsHandler) WorkflowDiscrepancy(c *gin.Context) {
	item, err := h.service.GetWorkflowDiscrepancy(c.Request.Context(), actorID(c), c.Param("id"))
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, item)
}

func (h *OperationsHandler) CreateWorkflowDiscrepancy(c *gin.Context) {
	p, ok := bindOperation[usecase.CreateDiscrepancyPayload](c)
	if !ok {
		return
	}
	item, err := h.service.CreateWorkflowDiscrepancy(c.Request.Context(), actorID(c), c.Param("id"), p)
	if err != nil {
		operationError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": item})
}

func (h *OperationsHandler) ResolveWorkflowDiscrepancy(c *gin.Context) {
	p, ok := bindOperation[discrepancyPayload](c)
	if !ok {
		return
	}
	if err := h.service.ResolveWorkflowDiscrepancy(c.Request.Context(), actorID(c), c.Param("id"), p.Status, p.Note); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"updated": true})
}

func (h *OperationsHandler) ReviewWorkflowDiscrepancy(c *gin.Context) {
	h.discrepancyStatus(c, "UNDER_REVIEW")
}
func (h *OperationsHandler) AcceptWorkflowDiscrepancy(c *gin.Context) {
	h.discrepancyStatus(c, "ACCEPTED")
}
func (h *OperationsHandler) RequireCorrectionForDiscrepancy(c *gin.Context) {
	h.discrepancyStatus(c, "CORRECTION_REQUIRED")
}
func (h *OperationsHandler) discrepancyStatus(c *gin.Context, status string) {
	p, ok := bindOperation[discrepancyPayload](c)
	if !ok {
		return
	}
	if err := h.service.ResolveWorkflowDiscrepancy(c.Request.Context(), actorID(c), c.Param("id"), status, p.Note); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"updated": true})
}

func (h *OperationsHandler) CompleteActionItem(c *gin.Context) {
	if err := h.service.CompleteActionItem(c.Request.Context(), actorID(c), c.Param("id")); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"completed": true})
}
