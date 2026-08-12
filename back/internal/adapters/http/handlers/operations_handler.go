package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"sangehassan/back/internal/usecase"
)

type OperationsHandler struct {
	service     *usecase.OperationsService
	version     string
	commit      string
	buildTime   string
	environment string
}

func NewOperationsHandler(s *usecase.OperationsService) *OperationsHandler {
	return &OperationsHandler{service: s}
}
func (h *OperationsHandler) ConfigureBuildInfo(version, commit, buildTime, environment string) {
	if len(commit) > 12 {
		commit = commit[:12]
	}
	h.version = version
	h.commit = commit
	h.buildTime = buildTime
	h.environment = environment
}
func actorID(c *gin.Context) string { v, _ := c.Get("user_id"); id, _ := v.(string); return id }
func operationError(c *gin.Context, err error) {
	status := http.StatusBadRequest
	var operationConflict *usecase.OperationConflict
	if errors.As(err, &operationConflict) {
		respondErrorCode(c, http.StatusConflict, operationConflict.Code, operationConflict.Message, nil)
		return
	}
	if errors.Is(err, usecase.ErrForbidden) {
		status = http.StatusForbidden
	}
	if errors.Is(err, usecase.ErrActivationInvalid) {
		status = http.StatusUnauthorized
	}
	code := "VALIDATION_FAILED"
	message := err.Error()
	if errors.Is(err, usecase.ErrForbidden) {
		code, message = "PERMISSION_DENIED", "شما دسترسی لازم برای انجام این عملیات را ندارید."
	}
	if errors.Is(err, usecase.ErrActivationInvalid) {
		code, message = "ACTIVATION_INVALID", "کد فعال‌سازی نامعتبر یا منقضی شده است."
	}
	respondErrorCode(c, status, code, message, nil)
}

func (h *OperationsHandler) Me(c *gin.Context) {
	u, err := h.service.GetOperationalUser(c.Request.Context(), actorID(c))
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, u)
}
func (h *OperationsHandler) Permissions(c *gin.Context) {
	v, err := h.service.ListPermissions(c.Request.Context())
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, v)
}
func (h *OperationsHandler) Roles(c *gin.Context) {
	v, err := h.service.ListRoles(c.Request.Context())
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, v)
}
func (h *OperationsHandler) Role(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, 400, "invalid id")
		return
	}
	v, err := h.service.GetRole(c.Request.Context(), id)
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, v)
}

type rolePayload struct {
	Code            string   `json:"code"`
	NameFA          string   `json:"name_fa"`
	DescriptionFA   string   `json:"description_fa"`
	IsActive        *bool    `json:"is_active"`
	CloneRoleID     *int64   `json:"clone_role_id"`
	PermissionCodes []string `json:"permission_codes"`
}

func (h *OperationsHandler) CreateRole(c *gin.Context) {
	var p rolePayload
	if c.ShouldBindJSON(&p) != nil {
		respondError(c, 400, "invalid payload")
		return
	}
	v, err := h.service.CreateRole(c.Request.Context(), actorID(c), p.Code, p.NameFA, p.DescriptionFA, p.CloneRoleID)
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, v)
}
func (h *OperationsHandler) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, 400, "invalid id")
		return
	}
	var p rolePayload
	if c.ShouldBindJSON(&p) != nil {
		respondError(c, 400, "invalid payload")
		return
	}
	active := true
	if p.IsActive != nil {
		active = *p.IsActive
	}
	if err = h.service.UpdateRole(c.Request.Context(), actorID(c), id, p.NameFA, p.DescriptionFA, active); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"updated": true})
}
func (h *OperationsHandler) RolePermissions(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, 400, "invalid id")
		return
	}
	var p rolePayload
	if c.ShouldBindJSON(&p) != nil {
		respondError(c, 400, "invalid payload")
		return
	}
	if err = h.service.UpdateRolePermissions(c.Request.Context(), actorID(c), id, p.PermissionCodes); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"updated": true})
}

func (h *OperationsHandler) Users(c *gin.Context) {
	v, err := h.service.ListUsers(c.Request.Context(), c.Query("search"), c.Query("status"), c.Query("role"))
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, v)
}

type userPayload struct {
	Phone             string  `json:"phone"`
	FirstName         string  `json:"first_name"`
	LastName          string  `json:"last_name"`
	TemporaryPassword string  `json:"temporary_password"`
	RoleIDs           []int64 `json:"role_ids"`
	Status            string  `json:"status"`
}

func (h *OperationsHandler) CreateUser(c *gin.Context) {
	var p userPayload
	if c.ShouldBindJSON(&p) != nil {
		respondError(c, 400, "invalid payload")
		return
	}
	u, err := h.service.CreateInternalUser(c.Request.Context(), actorID(c), p.Phone, p.FirstName, p.LastName, p.TemporaryPassword, p.RoleIDs)
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"user": u, "temporary_password": p.TemporaryPassword})
}
func (h *OperationsHandler) AssignRoles(c *gin.Context) {
	var p userPayload
	if c.ShouldBindJSON(&p) != nil {
		respondError(c, 400, "invalid payload")
		return
	}
	if err := h.service.AssignRoles(c.Request.Context(), actorID(c), c.Param("id"), p.RoleIDs); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"updated": true})
}
func (h *OperationsHandler) SetUserStatus(c *gin.Context) {
	var p userPayload
	if c.ShouldBindJSON(&p) != nil {
		respondError(c, 400, "invalid payload")
		return
	}
	if err := h.service.SetUserStatus(c.Request.Context(), actorID(c), c.Param("id"), p.Status); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"updated": true})
}
func (h *OperationsHandler) ResetPassword(c *gin.Context) {
	var p userPayload
	if c.ShouldBindJSON(&p) != nil {
		respondError(c, 400, "invalid payload")
		return
	}
	if err := h.service.ResetPassword(c.Request.Context(), actorID(c), c.Param("id"), p.TemporaryPassword); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"temporary_password": p.TemporaryPassword})
}

func (h *OperationsHandler) ActionItems(c *gin.Context) {
	v, err := h.service.ListActionItems(c.Request.Context(), actorID(c), c.Query("scope") == "all")
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, v)
}
func (h *OperationsHandler) Workflows(c *gin.Context) {
	v, err := h.service.AvailableWorkflows(c.Request.Context(), actorID(c))
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, v)
}
func (h *OperationsHandler) SearchCustomer(c *gin.Context) {
	v, err := h.service.SearchCustomer(c.Request.Context(), c.Query("phone"))
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, v)
}

type workflowPayload struct {
	WorkflowTemplateID        int64    `json:"workflow_template_id"`
	CustomerPhone             string   `json:"customer_phone"`
	CustomerName              string   `json:"customer_name"`
	ExcludedOptionalStepCodes []string `json:"excluded_optional_step_codes"`
}

func (h *OperationsHandler) StartWorkflow(c *gin.Context) {
	var p workflowPayload
	if c.ShouldBindJSON(&p) != nil {
		respondError(c, 400, "invalid payload")
		return
	}
	v, err := h.service.StartWorkflow(c.Request.Context(), actorID(c), p.WorkflowTemplateID, p.CustomerPhone, p.CustomerName, c.GetHeader("Idempotency-Key"), p.ExcludedOptionalStepCodes)
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, v)
}

type activationPayload struct {
	Phone    string `json:"phone"`
	Code     string `json:"activation_code"`
	Password string `json:"password"`
}

func (h *OperationsHandler) Activate(c *gin.Context) {
	var p activationPayload
	if c.ShouldBindJSON(&p) != nil {
		respondError(c, 400, "invalid payload")
		return
	}
	if err := h.service.ActivateCustomer(c.Request.Context(), p.Phone, p.Code, p.Password); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"activated": true})
}
func (h *OperationsHandler) RegenerateActivation(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	var payload struct {
		Reason string `json:"reason"`
	}
	if c.ShouldBindJSON(&payload) != nil {
		respondError(c, http.StatusBadRequest, "اطلاعات درخواست معتبر نیست.")
		return
	}
	result, err := h.service.RegenerateActivation(c.Request.Context(), actorID(c), c.Param("id"), key, payload.Reason)
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, result)
}

type proformaPayload struct {
	Currency       string `json:"currency"`
	Subtotal       string `json:"subtotal"`
	DiscountAmount string `json:"discount_amount"`
	Notes          string `json:"notes"`
}

func (h *OperationsHandler) CreateProforma(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	var p proformaPayload
	if c.ShouldBindJSON(&p) != nil {
		respondError(c, 400, "invalid payload")
		return
	}
	v, err := h.service.CreateProformaWithKey(c.Request.Context(), actorID(c), c.Param("id"), key, p.Currency, p.Subtotal, p.DiscountAmount, p.Notes)
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, v)
}
func (h *OperationsHandler) IssueProforma(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	if err := h.service.IssueProformaWithKey(c.Request.Context(), actorID(c), c.Param("id"), key); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"issued": true})
}
func (h *OperationsHandler) AccountOrders(c *gin.Context) {
	v, err := h.service.AccountOrders(c.Request.Context(), actorID(c))
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, v)
}
func (h *OperationsHandler) AccountProformas(c *gin.Context) {
	v, err := h.service.AccountProformas(c.Request.Context(), actorID(c))
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, v)
}
func (h *OperationsHandler) Audit(c *gin.Context) {
	v, err := h.service.AuditLogs(c.Request.Context(), c.Query("search"))
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, v)
}
