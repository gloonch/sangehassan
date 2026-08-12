package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"sangehassan/back/internal/usecase"
)

func idempotencyKey(c *gin.Context) (string, bool) {
	key := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	if key == "" {
		respondErrorCode(c, http.StatusBadRequest, "IDEMPOTENCY_KEY_REQUIRED", "ارسال Idempotency-Key الزامی است.", nil)
		return "", false
	}
	return key, true
}

type operationResultValue struct {
	value any
	err   error
}

func operationResult[T any](value T, err error) operationResultValue {
	return operationResultValue{value: value, err: err}
}
func createdOrError(c *gin.Context, result operationResultValue) {
	value, err := result.value, result.err
	if err != nil {
		operationError(c, err)
		return
	}
	respondCreated(c, value)
}
func okOrError(c *gin.Context, result operationResultValue) {
	value, err := result.value, result.err
	if err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, value)
}

func (h *OperationsHandler) OrderItems(c *gin.Context) {
	okOrError(c, operationResult(h.service.ListOrderItems(c.Request.Context(), actorID(c), c.Param("id"))))
}
func (h *OperationsHandler) CreateOrderItem(c *gin.Context) {
	p, ok := bindOperation[usecase.OrderItemPayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.CreateOrderItem(c.Request.Context(), actorID(c), c.Param("id"), p)))
}
func (h *OperationsHandler) UpdateOrderItem(c *gin.Context) {
	p, ok := bindOperation[usecase.OrderItemPayload](c)
	if !ok {
		return
	}
	if err := h.service.UpdateOrderItem(c.Request.Context(), actorID(c), c.Param("id"), p); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"updated": true})
}
func (h *OperationsHandler) DeleteOrderItem(c *gin.Context) {
	if err := h.service.DeleteOrderItem(c.Request.Context(), actorID(c), c.Param("id")); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"deleted": true})
}
func (h *OperationsHandler) CreateOrderItemConversion(c *gin.Context) {
	p, ok := bindOperation[usecase.QuantityConversionPayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.CreateQuantityConversion(c.Request.Context(), actorID(c), c.Param("id"), p)))
}
func (h *OperationsHandler) OrderProgress(c *gin.Context) {
	okOrError(c, operationResult(h.service.OrderProgress(c.Request.Context(), actorID(c), c.Param("id"), false)))
}
func (h *OperationsHandler) AccountOrderProgress(c *gin.Context) {
	okOrError(c, operationResult(h.service.OrderProgress(c.Request.Context(), actorID(c), c.Param("orderId"), true)))
}

func (h *OperationsHandler) Batches(c *gin.Context) {
	okOrError(c, operationResult(h.service.ListBatches(c.Request.Context(), actorID(c), c.Query("status"), c.Query("order_id"), c.Query("source"), c.Query("location_id"))))
}
func (h *OperationsHandler) Batch(c *gin.Context) {
	okOrError(c, operationResult(h.service.GetBatch(c.Request.Context(), actorID(c), c.Param("id"))))
}
func (h *OperationsHandler) CreateBatch(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.BatchPayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.CreateBatch(c.Request.Context(), actorID(c), c.Param("id"), key, p)))
}
func (h *OperationsHandler) UpdateBatch(c *gin.Context) {
	p, ok := bindOperation[usecase.BatchPayload](c)
	if !ok {
		return
	}
	if err := h.service.UpdateBatch(c.Request.Context(), actorID(c), c.Param("id"), p); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"updated": true})
}
func (h *OperationsHandler) SplitBatch(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.BatchSplitPayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.SplitBatch(c.Request.Context(), actorID(c), c.Param("id"), key, p)))
}
func (h *OperationsHandler) MergeBatches(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.BatchMergePayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.MergeBatches(c.Request.Context(), actorID(c), key, p)))
}
func (h *OperationsHandler) CancelBatch(c *gin.Context) {
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
	if err := h.service.CancelBatch(c.Request.Context(), actorID(c), c.Param("id"), key, p.Reason); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"cancelled": true})
}

func (h *OperationsHandler) Locations(c *gin.Context) {
	okOrError(c, operationResult(h.service.ListLocations(c.Request.Context(), c.Query("include_inactive") == "true")))
}
func (h *OperationsHandler) CreateLocation(c *gin.Context) {
	p, ok := bindOperation[usecase.LocationPayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.CreateLocation(c.Request.Context(), actorID(c), p)))
}
func (h *OperationsHandler) UpdateLocation(c *gin.Context) {
	p, ok := bindOperation[usecase.LocationPayload](c)
	if !ok {
		return
	}
	if err := h.service.UpdateLocation(c.Request.Context(), actorID(c), c.Param("id"), p); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"updated": true})
}
func (h *OperationsHandler) Lots(c *gin.Context) {
	okOrError(c, operationResult(h.service.ListLots(c.Request.Context(), c.Query("location_id"), c.Query("status"), c.Query("stone"))))
}
func (h *OperationsHandler) Lot(c *gin.Context) {
	okOrError(c, operationResult(h.service.GetLot(c.Request.Context(), c.Param("id"))))
}
func (h *OperationsHandler) UpdateLot(c *gin.Context) {
	p, ok := bindOperation[usecase.LotPayload](c)
	if !ok {
		return
	}
	if err := h.service.UpdateLotMetadata(c.Request.Context(), actorID(c), c.Param("id"), p); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"updated": true})
}
func (h *OperationsHandler) LotTraceability(c *gin.Context) {
	okOrError(c, operationResult(h.service.LotTraceability(c.Request.Context(), c.Param("id"))))
}
func (h *OperationsHandler) ReceiveInventory(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.ReceiptPayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.ReceiveInventory(c.Request.Context(), actorID(c), key, p)))
}
func (h *OperationsHandler) CreateReservation(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.ReservationPayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.CreateReservation(c.Request.Context(), actorID(c), c.Param("id"), key, p)))
}
func (h *OperationsHandler) BatchReservations(c *gin.Context) {
	okOrError(c, operationResult(h.service.ListBatchReservations(c.Request.Context(), c.Param("id"))))
}
func (h *OperationsHandler) ReleaseReservation(c *gin.Context) {
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
	if err := h.service.ReleaseReservation(c.Request.Context(), actorID(c), c.Param("id"), key, p.Reason); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"released": true})
}
func (h *OperationsHandler) TransferInventory(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.TransferPayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.TransferInventory(c.Request.Context(), actorID(c), key, p)))
}
func (h *OperationsHandler) AdjustInventory(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.AdjustmentPayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.AdjustInventory(c.Request.Context(), actorID(c), key, p)))
}
func (h *OperationsHandler) ConvertInventory(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.ConversionPayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.ConvertInventory(c.Request.Context(), actorID(c), key, p)))
}
func (h *OperationsHandler) InventoryMovements(c *gin.Context) {
	okOrError(c, operationResult(h.service.ListMovements(c.Request.Context(), c.Query("lot_id"), c.Query("batch_id"), c.Query("shipment_id"))))
}
func (h *OperationsHandler) InventorySummary(c *gin.Context) {
	okOrError(c, operationResult(h.service.InventoryDashboard(c.Request.Context(), actorID(c))))
}

func (h *OperationsHandler) Vehicles(c *gin.Context) {
	okOrError(c, operationResult(h.service.ListVehicles(c.Request.Context(), c.Query("include_inactive") == "true")))
}
func (h *OperationsHandler) CreateVehicle(c *gin.Context) {
	p, ok := bindOperation[usecase.VehiclePayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.CreateVehicle(c.Request.Context(), actorID(c), p)))
}
func (h *OperationsHandler) UpdateVehicle(c *gin.Context) {
	p, ok := bindOperation[usecase.VehiclePayload](c)
	if !ok {
		return
	}
	if err := h.service.UpdateVehicle(c.Request.Context(), actorID(c), c.Param("id"), p); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"updated": true})
}
func (h *OperationsHandler) Shipments(c *gin.Context) {
	okOrError(c, operationResult(h.service.ListShipments(c.Request.Context(), actorID(c), c.Query("status"), c.Query("order_id"), false)))
}
func (h *OperationsHandler) Shipment(c *gin.Context) {
	okOrError(c, operationResult(h.service.GetShipment(c.Request.Context(), actorID(c), c.Param("id"), false)))
}
func (h *OperationsHandler) CreateShipment(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.ShipmentPayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.CreateShipment(c.Request.Context(), actorID(c), c.Param("id"), key, p)))
}
func (h *OperationsHandler) UpdateShipment(c *gin.Context) {
	p, ok := bindOperation[usecase.ShipmentPayload](c)
	if !ok {
		return
	}
	if err := h.service.UpdateShipment(c.Request.Context(), actorID(c), c.Param("id"), p); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"updated": true})
}
func (h *OperationsHandler) AddShipmentItem(c *gin.Context) {
	p, ok := bindOperation[usecase.ShipmentItemPayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.AddShipmentItem(c.Request.Context(), actorID(c), c.Param("id"), p)))
}
func (h *OperationsHandler) shipmentOperation(c *gin.Context, operation string, customer bool) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.ShipmentOperationPayload](c)
	if !ok {
		return
	}
	var value any
	var err error
	switch operation {
	case "load":
		value, err = h.service.LoadShipment(c.Request.Context(), actorID(c), c.Param("id"), key, p)
	case "dispatch":
		value, err = h.service.DispatchShipment(c.Request.Context(), actorID(c), c.Param("id"), key, p)
	case "arrive":
		value, err = h.service.ArriveShipment(c.Request.Context(), actorID(c), c.Param("id"), key, p)
	case "deliver":
		value, err = h.service.DeliverShipment(c.Request.Context(), actorID(c), c.Param("id"), key, p, customer)
	}
	okOrError(c, operationResult(value, err))
}
func (h *OperationsHandler) LoadShipment(c *gin.Context) { h.shipmentOperation(c, "load", false) }
func (h *OperationsHandler) DispatchShipment(c *gin.Context) {
	h.shipmentOperation(c, "dispatch", false)
}
func (h *OperationsHandler) ArriveShipment(c *gin.Context)  { h.shipmentOperation(c, "arrive", false) }
func (h *OperationsHandler) DeliverShipment(c *gin.Context) { h.shipmentOperation(c, "deliver", false) }
func (h *OperationsHandler) AccountDeliverShipment(c *gin.Context) {
	h.shipmentOperation(c, "deliver", true)
}
func (h *OperationsHandler) CancelShipment(c *gin.Context) {
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
	if err := h.service.CancelShipment(c.Request.Context(), actorID(c), c.Param("id"), key, p.Reason); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"cancelled": true})
}
func (h *OperationsHandler) AccountShipments(c *gin.Context) {
	okOrError(c, operationResult(h.service.ListShipments(c.Request.Context(), actorID(c), "", c.Param("orderId"), true)))
}

func (h *OperationsHandler) Packaging(c *gin.Context) {
	okOrError(c, operationResult(h.service.ListPackaging(c.Request.Context(), c.Query("batch_id"))))
}
func (h *OperationsHandler) CreatePackaging(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.PackagingPayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.CreatePackaging(c.Request.Context(), actorID(c), c.Param("id"), key, p)))
}
func (h *OperationsHandler) UpdatePackagingStatus(c *gin.Context) {
	p, ok := bindOperation[struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}](c)
	if !ok {
		return
	}
	if err := h.service.UpdatePackagingStatus(c.Request.Context(), actorID(c), c.Param("id"), p.Status, p.Reason); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"updated": true})
}
func (h *OperationsHandler) AssignPackage(c *gin.Context) {
	p, ok := bindOperation[struct {
		ShipmentItemID string `json:"shipment_item_id"`
		Reason         string `json:"reason"`
	}](c)
	if !ok {
		return
	}
	if err := h.service.AssignPackage(c.Request.Context(), actorID(c), c.Param("id"), p.ShipmentItemID, p.Reason); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"assigned": true})
}
func (h *OperationsHandler) CreateContainer(c *gin.Context) {
	p, ok := bindOperation[usecase.ContainerPayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.CreateContainer(c.Request.Context(), actorID(c), c.Param("id"), p)))
}
func (h *OperationsHandler) UpdateContainer(c *gin.Context) {
	p, ok := bindOperation[usecase.ContainerPayload](c)
	if !ok {
		return
	}
	if err := h.service.UpdateContainer(c.Request.Context(), actorID(c), c.Param("id"), p); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"updated": true})
}
func (h *OperationsHandler) AddContainerItem(c *gin.Context) {
	p, ok := bindOperation[usecase.ContainerItemPayload](c)
	if !ok {
		return
	}
	if err := h.service.AddContainerItem(c.Request.Context(), actorID(c), c.Param("id"), p); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"created": true})
}

func (h *OperationsHandler) OperationalCosts(c *gin.Context) {
	okOrError(c, operationResult(h.service.ListOperationalCosts(c.Request.Context(), actorID(c), c.Query("entity_type"), c.Query("entity_id"))))
}
func (h *OperationsHandler) CreateOperationalCost(c *gin.Context) {
	p, ok := bindOperation[usecase.CostPayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.CreateOperationalCost(c.Request.Context(), actorID(c), p)))
}
func (h *OperationsHandler) ApproveOperationalCost(c *gin.Context) {
	if err := h.service.ApproveOperationalCost(c.Request.Context(), actorID(c), c.Param("id")); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"approved": true})
}
func (h *OperationsHandler) VoidOperationalCost(c *gin.Context) {
	p, ok := bindOperation[struct {
		Reason string `json:"reason"`
	}](c)
	if !ok {
		return
	}
	if err := h.service.VoidOperationalCost(c.Request.Context(), actorID(c), c.Param("id"), p.Reason); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"voided": true})
}
func (h *OperationsHandler) OperationsDashboardSummary(c *gin.Context) {
	okOrError(c, operationResult(h.service.OperationsDashboardSummary(c.Request.Context(), actorID(c))))
}

func (h *OperationsHandler) WorkflowTransitions(c *gin.Context) {
	id, ok := int64Param(c, "id")
	if !ok {
		return
	}
	okOrError(c, operationResult(h.service.ListWorkflowTransitions(c.Request.Context(), id)))
}
func (h *OperationsHandler) CreateWorkflowTransition(c *gin.Context) {
	id, ok := int64Param(c, "id")
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.WorkflowTransitionPayload](c)
	if !ok {
		return
	}
	createdOrError(c, operationResult(h.service.CreateWorkflowTransition(c.Request.Context(), actorID(c), id, p)))
}
func (h *OperationsHandler) UpdateWorkflowTransition(c *gin.Context) {
	id, ok := int64Param(c, "transitionId")
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.WorkflowTransitionPayload](c)
	if !ok {
		return
	}
	if err := h.service.UpdateWorkflowTransition(c.Request.Context(), actorID(c), id, p); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"updated": true})
}
func (h *OperationsHandler) DeleteWorkflowTransition(c *gin.Context) {
	id, ok := int64Param(c, "transitionId")
	if !ok {
		return
	}
	if err := h.service.DeleteWorkflowTransition(c.Request.Context(), actorID(c), id); err != nil {
		operationError(c, err)
		return
	}
	respondOK(c, gin.H{"deleted": true})
}
func (h *OperationsHandler) RuntimeTransitions(c *gin.Context) {
	okOrError(c, operationResult(h.service.GetRuntimeTransitions(c.Request.Context(), actorID(c), c.Param("id"))))
}
func (h *OperationsHandler) SelectRuntimeTransition(c *gin.Context) {
	key, ok := idempotencyKey(c)
	if !ok {
		return
	}
	p, ok := bindOperation[usecase.SelectTransitionPayload](c)
	if !ok {
		return
	}
	okOrError(c, operationResult(h.service.SelectWorkflowTransition(c.Request.Context(), actorID(c), c.Param("id"), key, p)))
}
