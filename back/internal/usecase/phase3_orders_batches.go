package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/lib/pq"
)

func (s *OperationsService) ListOrderItems(ctx context.Context, orderID string) ([]OrderItem, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,order_id,product_id,COALESCE(stone_category,''),stone_name,COALESCE(stone_variant,''),COALESCE(finish_type,''),COALESCE(cut_type,''),ordered_quantity::text,quantity_unit,COALESCE(quality_grade,''),progress_weight::text,requires_production,requires_packaging,COALESCE(notes,''),created_at FROM order_items WHERE order_id=$1 ORDER BY created_at,id`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []OrderItem{}
	for rows.Next() {
		var x OrderItem
		var product sql.NullInt64
		if err = rows.Scan(&x.ID, &x.OrderID, &product, &x.StoneCategory, &x.StoneName, &x.StoneVariant, &x.FinishType, &x.CutType, &x.OrderedQuantity, &x.QuantityUnit, &x.QualityGrade, &x.ProgressWeight, &x.RequiresProduction, &x.RequiresPackaging, &x.Notes, &x.CreatedAt); err != nil {
			return nil, err
		}
		if product.Valid {
			v := product.Int64
			x.ProductID = &v
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func validateOrderItemPayload(p OrderItemPayload) error {
	p.QuantityUnit = normalizeCode(p.QuantityUnit)
	if strings.TrimSpace(p.StoneName) == "" || !validPositiveDecimal(p.OrderedQuantity) || !quantityUnits[p.QuantityUnit] {
		return ErrValidation
	}
	if p.ProgressWeight != "" && !validPositiveDecimal(p.ProgressWeight) {
		return ErrValidation
	}
	return nil
}

func (s *OperationsService) CreateOrderItem(ctx context.Context, actor, orderID string, p OrderItemPayload) (OrderItem, error) {
	var out OrderItem
	if err := validateOrderItemPayload(p); err != nil {
		return out, err
	}
	p.QuantityUnit = normalizeCode(p.QuantityUnit)
	if p.ProgressWeight == "" {
		p.ProgressWeight = "1"
	}
	requiresProduction, requiresPackaging := true, true
	if p.RequiresProduction != nil {
		requiresProduction = *p.RequiresProduction
	}
	if p.RequiresPackaging != nil {
		requiresPackaging = *p.RequiresPackaging
	}
	err := s.db.QueryRowContext(ctx, `INSERT INTO order_items(order_id,product_id,stone_category,stone_name,stone_variant,finish_type,cut_type,thickness_value,thickness_unit,width_value,length_value,dimension_unit,ordered_quantity,quantity_unit,quality_grade,color,pattern,progress_weight,requires_production,requires_packaging,notes,created_by_user_id) SELECT $1,$2,$3,$4,$5,$6,$7,NULLIF($8,'')::numeric,NULLIF($9,''),NULLIF($10,'')::numeric,NULLIF($11,'')::numeric,NULLIF($12,''),$13::numeric,$14,NULLIF($15,''),NULLIF($16,''),NULLIF($17,''),$18::numeric,$19,$20,NULLIF($21,''),$22 WHERE EXISTS(SELECT 1 FROM orders WHERE id=$1) RETURNING id`, orderID, p.ProductID, p.StoneCategory, strings.TrimSpace(p.StoneName), p.StoneVariant, p.FinishType, p.CutType, valueOrEmpty(p.ThicknessValue), p.ThicknessUnit, valueOrEmpty(p.WidthValue), valueOrEmpty(p.LengthValue), p.DimensionUnit, p.OrderedQuantity, p.QuantityUnit, p.QualityGrade, p.Color, p.Pattern, p.ProgressWeight, requiresProduction, requiresPackaging, p.Notes, actor).Scan(&out.ID)
	if err != nil {
		return out, err
	}
	s.audit(ctx, actor, "order_items.create", "order_item", out.ID, p)
	items, err := s.ListOrderItems(ctx, orderID)
	if err != nil {
		return out, err
	}
	for _, x := range items {
		if x.ID == out.ID {
			return x, nil
		}
	}
	return out, sql.ErrNoRows
}

func valueOrEmpty(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func (s *OperationsService) UpdateOrderItem(ctx context.Context, actor, id string, p OrderItemPayload) error {
	if err := validateOrderItemPayload(p); err != nil {
		return err
	}
	p.QuantityUnit = normalizeCode(p.QuantityUnit)
	if p.ProgressWeight == "" {
		p.ProgressWeight = "1"
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var batchCount int
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM fulfillment_batches WHERE order_item_id=$1`, id).Scan(&batchCount); err != nil {
		return err
	}
	if batchCount > 0 {
		return conflict("ORDER_ITEM_IN_USE", "order item with batches cannot change quantity or unit")
	}
	requiresProduction, requiresPackaging := true, true
	if p.RequiresProduction != nil {
		requiresProduction = *p.RequiresProduction
	}
	if p.RequiresPackaging != nil {
		requiresPackaging = *p.RequiresPackaging
	}
	r, err := tx.ExecContext(ctx, `UPDATE order_items SET product_id=$2,stone_category=$3,stone_name=$4,stone_variant=$5,finish_type=$6,cut_type=$7,ordered_quantity=$8::numeric,quantity_unit=$9,quality_grade=NULLIF($10,''),progress_weight=$11::numeric,requires_production=$12,requires_packaging=$13,notes=NULLIF($14,''),updated_at=NOW() WHERE id=$1`, id, p.ProductID, p.StoneCategory, p.StoneName, p.StoneVariant, p.FinishType, p.CutType, p.OrderedQuantity, p.QuantityUnit, p.QualityGrade, p.ProgressWeight, requiresProduction, requiresPackaging, p.Notes)
	if err != nil {
		return err
	}
	if n, _ := r.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	s.auditTx(ctx, tx, actor, "order_items.update", "order_item", id, nil, p)
	return tx.Commit()
}

func (s *OperationsService) DeleteOrderItem(ctx context.Context, actor, id string) error {
	r, err := s.db.ExecContext(ctx, `DELETE FROM order_items i WHERE i.id=$1 AND NOT EXISTS(SELECT 1 FROM fulfillment_batches b WHERE b.order_item_id=i.id)`, id)
	if err != nil {
		return err
	}
	if n, _ := r.RowsAffected(); n == 0 {
		return conflict("ORDER_ITEM_IN_USE", "order item is used or does not exist")
	}
	s.audit(ctx, actor, "order_items.delete", "order_item", id, nil)
	return nil
}

func (s *OperationsService) CreateQuantityConversion(ctx context.Context, actor, itemID string, p QuantityConversionPayload) (map[string]any, error) {
	p.FromUnit = normalizeCode(p.FromUnit)
	p.ToUnit = normalizeCode(p.ToUnit)
	if !validPositiveDecimal(p.FromQuantity) || !validPositiveDecimal(p.ToQuantity) || p.FromUnit == p.ToUnit || validateUnit(p.FromUnit) != nil || validateUnit(p.ToUnit) != nil || requireReason(p.Reason) != nil {
		return nil, ErrValidation
	}
	var id string
	err := s.db.QueryRowContext(ctx, `INSERT INTO order_item_quantity_conversions(order_item_id,from_quantity,from_unit,to_quantity,to_unit,reason,created_by_user_id) VALUES($1,$2::numeric,$3,$4::numeric,$5,$6,$7) ON CONFLICT(order_item_id,from_unit,to_unit) DO UPDATE SET from_quantity=EXCLUDED.from_quantity,to_quantity=EXCLUDED.to_quantity,reason=EXCLUDED.reason,created_by_user_id=EXCLUDED.created_by_user_id,created_at=NOW() RETURNING id`, itemID, p.FromQuantity, p.FromUnit, p.ToQuantity, p.ToUnit, p.Reason, actor).Scan(&id)
	if err != nil {
		return nil, err
	}
	s.audit(ctx, actor, "order_item_conversions.upsert", "order_item_quantity_conversion", id, p)
	return map[string]any{"id": id}, nil
}

func convertQuantityTx(ctx context.Context, tx *sql.Tx, itemID, quantity, fromUnit, toUnit string) (*big.Rat, error) {
	q, ok := new(big.Rat).SetString(quantity)
	if !ok {
		return nil, ErrValidation
	}
	if fromUnit == toUnit {
		return q, nil
	}
	var from, to string
	err := tx.QueryRowContext(ctx, `SELECT from_quantity::text,to_quantity::text FROM order_item_quantity_conversions WHERE order_item_id=$1 AND from_unit=$2 AND to_unit=$3`, itemID, fromUnit, toUnit).Scan(&from, &to)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, conflict("INCOMPATIBLE_UNIT", "explicit order item conversion is required")
	}
	if err != nil {
		return nil, err
	}
	fr, _ := new(big.Rat).SetString(from)
	tr, _ := new(big.Rat).SetString(to)
	return new(big.Rat).Mul(q, new(big.Rat).Quo(tr, fr)), nil
}

func (s *OperationsService) ensureBatchAllocationTx(ctx context.Context, tx *sql.Tx, actor, itemID, newQty, newUnit string, override bool, reason string) error {
	var ordered, orderUnit string
	if err := tx.QueryRowContext(ctx, `SELECT ordered_quantity::text,quantity_unit FROM order_items WHERE id=$1 FOR UPDATE`, itemID).Scan(&ordered, &orderUnit); err != nil {
		return err
	}
	total := new(big.Rat)
	rows, err := tx.QueryContext(ctx, `SELECT planned_quantity::text,quantity_unit FROM fulfillment_batches WHERE order_item_id=$1 AND status NOT IN ('SPLIT','MERGED','CANCELLED') FOR UPDATE`, itemID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var q, u string
		if err = rows.Scan(&q, &u); err != nil {
			return err
		}
		v, e := convertQuantityTx(ctx, tx, itemID, q, u, orderUnit)
		if e != nil {
			return e
		}
		total.Add(total, v)
	}
	if err = rows.Err(); err != nil {
		return err
	}
	v, err := convertQuantityTx(ctx, tx, itemID, newQty, newUnit, orderUnit)
	if err != nil {
		return err
	}
	total.Add(total, v)
	limit, _ := new(big.Rat).SetString(ordered)
	if total.Cmp(limit) > 0 {
		if !override || !s.HasPermission(ctx, actor, "batches.override") {
			return conflict("OVER_ALLOCATION", "active batch quantity exceeds order item")
		}
		if err = requireReason(reason); err != nil {
			return err
		}
	}
	return nil
}

func (s *OperationsService) startScopedWorkflowTx(ctx context.Context, tx *sql.Tx, actor, orderID, scopeType, scopeID string, templateID *int64, parentID, parentStepID *string) (string, error) {
	var tid int64
	var customer string
	if templateID == nil {
		err := tx.QueryRowContext(ctx, `SELECT id FROM workflow_templates WHERE scope_type=$1 AND status='PUBLISHED' AND is_active ORDER BY version_number DESC,id LIMIT 1`, scopeType).Scan(&tid)
		if err != nil {
			return "", err
		}
	} else {
		tid = *templateID
	}
	var group, startPermission string
	var version int
	if err := tx.QueryRowContext(ctx, `SELECT template_group_code,version_number,start_permission_code FROM workflow_templates WHERE id=$1 AND scope_type=$2 AND status='PUBLISHED' AND is_active FOR SHARE`, tid, scopeType).Scan(&group, &version, &startPermission); err != nil {
		return "", err
	}
	if !s.HasPermission(ctx, actor, startPermission) {
		return "", ErrForbidden
	}
	if err := tx.QueryRowContext(ctx, `SELECT customer_user_id FROM orders WHERE id=$1`, orderID).Scan(&customer); err != nil {
		return "", err
	}
	var workflowID string
	err := tx.QueryRowContext(ctx, `INSERT INTO workflow_instances(workflow_template_id,order_id,customer_user_id,started_by_user_id,template_group_code,template_version_number,scope_type,scope_id,parent_workflow_instance_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`, tid, orderID, customer, actor, group, version, scopeType, scopeID, parentID).Scan(&workflowID)
	if err != nil {
		return "", err
	}
	if err = s.snapshotWorkflowTx(ctx, tx, workflowID, tid, actor, nil); err != nil {
		return "", err
	}
	if parentID != nil {
		var actionID sql.NullString
		if parentStepID != nil {
			_ = tx.QueryRowContext(ctx, `INSERT INTO action_items(workflow_instance_id,workflow_step_instance_id,order_id,customer_user_id,title_fa,status,priority,required_permission_code,deduplication_key,source_trigger_type,is_blocking) SELECT $1,$2,$3,$4,'تکمیل فرایند فرزند','OPEN','HIGH','workflow_instances.update','child-workflow:'||$5,'CHILD_WORKFLOW',TRUE RETURNING id`, *parentID, *parentStepID, orderID, customer, workflowID).Scan(&actionID)
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO workflow_child_dependencies(parent_workflow_instance_id,parent_step_instance_id,child_workflow_instance_id,action_item_id) VALUES($1,$2,$3,$4)`, *parentID, parentStepID, workflowID, actionID)
		if err != nil {
			return "", err
		}
	}
	return workflowID, nil
}

func (s *OperationsService) CreateBatch(ctx context.Context, actor, orderID, key string, p BatchPayload) (Batch, error) {
	var out Batch
	p.SourceType = normalizeCode(p.SourceType)
	p.QuantityUnit = normalizeCode(p.QuantityUnit)
	if p.Priority == "" {
		p.Priority = "NORMAL"
	}
	if !validPositiveDecimal(p.PlannedQuantity) || validateUnit(p.QuantityUnit) != nil || strings.TrimSpace(p.OrderItemID) == "" || strings.TrimSpace(p.StoneName) == "" {
		return out, ErrValidation
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return out, err
	}
	defer tx.Rollback()
	operation := "BATCH_CREATE:" + orderID
	idem, err := claimOperationTx(ctx, tx, actor, operation, key, p)
	if err != nil {
		return out, err
	}
	if idem.Existing {
		var stored struct {
			ID string `json:"id"`
		}
		if err = json.Unmarshal(idem.Response, &stored); err != nil {
			return out, err
		}
		return s.GetBatch(ctx, actor, stored.ID)
	}
	var orderNumber, itemOrder string
	if err = tx.QueryRowContext(ctx, `SELECT i.order_id,o.order_number FROM order_items i JOIN orders o ON o.id=i.order_id WHERE i.id=$1 AND o.id=$2 FOR UPDATE OF i,o`, p.OrderItemID, orderID).Scan(&itemOrder, &orderNumber); err != nil {
		return out, err
	}
	if err = s.ensureBatchAllocationTx(ctx, tx, actor, p.OrderItemID, p.PlannedQuantity, p.QuantityUnit, p.Override, p.Reason); err != nil {
		return out, err
	}
	var seq int
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*)+1 FROM fulfillment_batches WHERE order_id=$1`, orderID).Scan(&seq); err != nil {
		return out, err
	}
	number := fmt.Sprintf("%s-B%02d", orderNumber, seq)
	required := true
	if p.IsRequired != nil {
		required = *p.IsRequired
	}
	err = tx.QueryRowContext(ctx, `INSERT INTO fulfillment_batches(batch_number,order_id,order_item_id,parent_batch_id,source_type,source_location_id,target_location_id,stone_category,stone_name,stone_variant,finish_type,cut_type,thickness_value,thickness_unit,planned_quantity,quantity_unit,status,priority,is_required,created_by_user_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,NULLIF($13,'')::numeric,NULLIF($14,''),$15::numeric,$16,'PLANNED',$17,$18,$19) RETURNING id`, number, orderID, p.OrderItemID, p.ParentBatchID, p.SourceType, p.SourceLocationID, p.TargetLocationID, p.StoneCategory, p.StoneName, p.StoneVariant, p.FinishType, p.CutType, valueOrEmpty(p.ThicknessValue), p.ThicknessUnit, p.PlannedQuantity, p.QuantityUnit, p.Priority, required, actor).Scan(&out.ID)
	if err != nil {
		return out, err
	}
	wid, err := s.startScopedWorkflowTx(ctx, tx, actor, orderID, "BATCH", out.ID, p.WorkflowTemplateID, p.ParentWorkflowID, p.ParentStepID)
	if err != nil {
		return out, err
	}
	_, err = tx.ExecContext(ctx, `UPDATE fulfillment_batches SET workflow_instance_id=$2 WHERE id=$1`, out.ID, wid)
	if err != nil {
		return out, err
	}
	s.auditTx(ctx, tx, actor, "batches.create", "fulfillment_batch", out.ID, nil, p)
	if err = finishOperationTx(ctx, tx, actor, operation, key, map[string]string{"id": out.ID}); err != nil {
		return out, err
	}
	if err = tx.Commit(); err != nil {
		return out, err
	}
	return s.GetBatch(ctx, actor, out.ID)
}

func (s *OperationsService) ListBatches(ctx context.Context, actor, status, orderID, source, location string) ([]Batch, error) {
	viewAll := s.HasPermission(ctx, actor, "batches.view_all")
	rows, err := s.db.QueryContext(ctx, `SELECT b.id,b.batch_number,b.order_id,o.order_number,b.order_item_id,b.parent_batch_id,b.source_type,b.source_location_id,b.target_location_id,b.stone_name,COALESCE(b.stone_variant,''),b.planned_quantity::text,b.actual_quantity::text,b.quantity_unit,b.status,b.priority,b.is_required,b.workflow_instance_id,b.estimated_ready_at,b.created_at FROM fulfillment_batches b JOIN orders o ON o.id=b.order_id WHERE ($2 OR b.created_by_user_id=$1 OR EXISTS(SELECT 1 FROM workflow_step_instances si JOIN user_roles ur ON ur.role_id=si.responsible_role_id WHERE si.workflow_instance_id=b.workflow_instance_id AND ur.user_id=$1)) AND ($3='' OR b.status=$3) AND ($4='' OR b.order_id=$4::uuid) AND ($5='' OR b.source_type=$5) AND ($6='' OR b.source_location_id=$6::uuid OR b.target_location_id=$6::uuid) ORDER BY b.created_at DESC`, actor, viewAll, status, orderID, source, location)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Batch{}
	for rows.Next() {
		var b Batch
		var parent, sourceID, target, wid sql.NullString
		var est sql.NullTime
		if err = rows.Scan(&b.ID, &b.BatchNumber, &b.OrderID, &b.OrderNumber, &b.OrderItemID, &parent, &b.SourceType, &sourceID, &target, &b.StoneName, &b.StoneVariant, &b.PlannedQuantity, &b.ActualQuantity, &b.QuantityUnit, &b.Status, &b.Priority, &b.IsRequired, &wid, &est, &b.CreatedAt); err != nil {
			return nil, err
		}
		b.ParentBatchID = scanNullableString(parent)
		b.SourceLocationID = scanNullableString(sourceID)
		b.TargetLocationID = scanNullableString(target)
		b.WorkflowInstanceID = scanNullableString(wid)
		if est.Valid {
			t := est.Time
			b.EstimatedReadyAt = &t
		}
		if !s.HasPermission(ctx, actor, "inventory.lots.view") {
			b.SourceLocationID = nil
			b.TargetLocationID = nil
			b.SourceType = ""
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *OperationsService) GetBatch(ctx context.Context, actor, id string) (Batch, error) {
	var b Batch
	var parent, sourceID, target, wid sql.NullString
	var est sql.NullTime
	viewAll := s.HasPermission(ctx, actor, "batches.view_all")
	err := s.db.QueryRowContext(ctx, `SELECT b.id,b.batch_number,b.order_id,o.order_number,b.order_item_id,b.parent_batch_id,b.source_type,b.source_location_id,b.target_location_id,b.stone_name,COALESCE(b.stone_variant,''),b.planned_quantity::text,b.actual_quantity::text,b.quantity_unit,b.status,b.priority,b.is_required,b.workflow_instance_id,b.estimated_ready_at,b.created_at FROM fulfillment_batches b JOIN orders o ON o.id=b.order_id WHERE b.id=$1 AND ($3 OR b.created_by_user_id=$2 OR EXISTS(SELECT 1 FROM workflow_step_instances si JOIN user_roles ur ON ur.role_id=si.responsible_role_id WHERE si.workflow_instance_id=b.workflow_instance_id AND ur.user_id=$2))`, id, actor, viewAll).Scan(&b.ID, &b.BatchNumber, &b.OrderID, &b.OrderNumber, &b.OrderItemID, &parent, &b.SourceType, &sourceID, &target, &b.StoneName, &b.StoneVariant, &b.PlannedQuantity, &b.ActualQuantity, &b.QuantityUnit, &b.Status, &b.Priority, &b.IsRequired, &wid, &est, &b.CreatedAt)
	if err != nil {
		return b, err
	}
	b.ParentBatchID = scanNullableString(parent)
	b.SourceLocationID = scanNullableString(sourceID)
	b.TargetLocationID = scanNullableString(target)
	b.WorkflowInstanceID = scanNullableString(wid)
	if est.Valid {
		t := est.Time
		b.EstimatedReadyAt = &t
	}
	if !s.HasPermission(ctx, actor, "inventory.lots.view") {
		b.SourceLocationID = nil
		b.TargetLocationID = nil
		b.SourceType = ""
	}
	return b, nil
}

func (s *OperationsService) UpdateBatch(ctx context.Context, actor, id string, p BatchPayload) error {
	if p.Override && (!s.HasPermission(ctx, actor, "batches.override") || requireReason(p.Reason) != nil) {
		return ErrForbidden
	}
	r, err := s.db.ExecContext(ctx, `UPDATE fulfillment_batches SET source_type=COALESCE(NULLIF($2,''),source_type),source_location_id=$3,target_location_id=$4,priority=COALESCE(NULLIF($5,''),priority),stone_variant=COALESCE(NULLIF($6,''),stone_variant),updated_at=NOW() WHERE id=$1 AND status NOT IN ('SPLIT','MERGED','CANCELLED','DELIVERED')`, id, normalizeCode(p.SourceType), p.SourceLocationID, p.TargetLocationID, p.Priority, p.StoneVariant)
	if err != nil {
		return err
	}
	if n, _ := r.RowsAffected(); n == 0 {
		return conflict("INVALID_BATCH_STATE", "batch is terminal or missing")
	}
	s.audit(ctx, actor, "batches.update", "fulfillment_batch", id, p)
	return nil
}

func ratString(r *big.Rat) string { return r.FloatString(quantityScale) }

func (s *OperationsService) SplitBatch(ctx context.Context, actor, id, key string, p BatchSplitPayload) ([]Batch, error) {
	if len(p.Children) < 2 || requireReason(p.Reason) != nil {
		return nil, ErrValidation
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	idem, err := claimOperationTx(ctx, tx, actor, "BATCH_SPLIT", key, p)
	if err != nil {
		return nil, err
	}
	if idem.Existing {
		var out []Batch
		if err = json.Unmarshal(idem.Response, &out); err != nil {
			return nil, err
		}
		return out, nil
	}
	var orderID, itemID, unit, parentQty, status, stone, variant, orderNumber string
	if err = tx.QueryRowContext(ctx, `SELECT b.order_id,b.order_item_id,b.quantity_unit,b.planned_quantity::text,b.status,b.stone_name,COALESCE(b.stone_variant,''),o.order_number FROM fulfillment_batches b JOIN orders o ON o.id=b.order_id WHERE b.id=$1 FOR UPDATE OF b,o`, id).Scan(&orderID, &itemID, &unit, &parentQty, &status, &stone, &variant, &orderNumber); err != nil {
		return nil, err
	}
	if status == "SPLIT" || status == "MERGED" || status == "CANCELLED" || status == "DELIVERED" {
		return nil, conflict("INVALID_BATCH_STATE", "batch cannot be split")
	}
	total := new(big.Rat)
	for _, child := range p.Children {
		if !validPositiveDecimal(child.PlannedQuantity) {
			return nil, ErrValidation
		}
		q, _ := new(big.Rat).SetString(child.PlannedQuantity)
		total.Add(total, q)
	}
	parent, _ := new(big.Rat).SetString(parentQty)
	if total.Cmp(parent) > 0 {
		return nil, conflict("OVER_ALLOCATION", "split quantity exceeds parent")
	}
	var baseSeq int
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM fulfillment_batches WHERE order_id=$1`, orderID).Scan(&baseSeq); err != nil {
		return nil, err
	}
	out := make([]Batch, 0, len(p.Children))
	for i, child := range p.Children {
		number := fmt.Sprintf("%s-B%02d", orderNumber, baseSeq+i+1)
		source := normalizeCode(child.SourceType)
		if source == "" {
			source = "FACTORY"
		}
		var childID string
		err = tx.QueryRowContext(ctx, `INSERT INTO fulfillment_batches(batch_number,order_id,order_item_id,parent_batch_id,source_type,source_location_id,target_location_id,stone_name,stone_variant,planned_quantity,quantity_unit,status,priority,is_required,created_by_user_id) SELECT $1,order_id,order_item_id,id,$2,$3,$4,stone_name,stone_variant,$5::numeric,quantity_unit,'PLANNED',priority,is_required,$6 FROM fulfillment_batches WHERE id=$7 RETURNING id`, number, source, child.SourceLocationID, child.TargetLocationID, child.PlannedQuantity, actor, id).Scan(&childID)
		if err != nil {
			return nil, err
		}
		wid, e := s.startScopedWorkflowTx(ctx, tx, actor, orderID, "BATCH", childID, child.WorkflowTemplateID, nil, nil)
		if e != nil {
			return nil, e
		}
		_, err = tx.ExecContext(ctx, `UPDATE fulfillment_batches SET workflow_instance_id=$2 WHERE id=$1`, childID, wid)
		if err != nil {
			return nil, err
		}
		for _, allocation := range child.ReservationAllocations {
			if !validPositiveDecimal(allocation.Quantity) {
				return nil, ErrValidation
			}
			var lotID, resQty, consumed, resUnit, resStatus string
			if err = tx.QueryRowContext(ctx, `SELECT inventory_lot_id,reserved_quantity::text,consumed_quantity::text,quantity_unit,status FROM inventory_reservations WHERE id=$1 AND batch_id=$2 FOR UPDATE`, allocation.ReservationID, id).Scan(&lotID, &resQty, &consumed, &resUnit, &resStatus); err != nil {
				return nil, err
			}
			if consumed != "0.0000" && consumed != "0" {
				return nil, conflict("RESERVATION_IN_USE", "consumed reservation cannot be split")
			}
			a, _ := new(big.Rat).SetString(allocation.Quantity)
			rq, _ := new(big.Rat).SetString(resQty)
			if a.Cmp(rq) > 0 {
				return nil, conflict("OVER_ALLOCATION", "reservation allocation exceeds parent reservation")
			}
			_, err = tx.ExecContext(ctx, `UPDATE inventory_reservations SET reserved_quantity=reserved_quantity-$2::numeric,status=CASE WHEN reserved_quantity-$2::numeric=0 THEN 'CANCELLED' ELSE status END,updated_at=NOW() WHERE id=$1`, allocation.ReservationID, allocation.Quantity)
			if err != nil {
				return nil, err
			}
			_, err = tx.ExecContext(ctx, `INSERT INTO inventory_reservations(inventory_lot_id,order_id,order_item_id,batch_id,reserved_quantity,quantity_unit,status,reserved_by_user_id) VALUES($1,$2,$3,$4,$5::numeric,$6,'ACTIVE',$7)`, lotID, orderID, itemID, childID, allocation.Quantity, resUnit, actor)
			if err != nil {
				return nil, err
			}
		}
		out = append(out, Batch{ID: childID, BatchNumber: number, OrderID: orderID, OrderNumber: orderNumber, OrderItemID: itemID, ParentBatchID: &id, StoneName: stone, StoneVariant: variant, PlannedQuantity: child.PlannedQuantity, ActualQuantity: "0.0000", QuantityUnit: unit, Status: "PLANNED", WorkflowInstanceID: &wid})
	}
	remaining := new(big.Rat).Sub(parent, total)
	newStatus := status
	if remaining.Sign() == 0 {
		newStatus = "SPLIT"
	}
	_, err = tx.ExecContext(ctx, `UPDATE fulfillment_batches SET planned_quantity=$2::numeric,status=$3,updated_at=NOW() WHERE id=$1`, id, ratString(remaining), newStatus)
	if err != nil {
		return nil, err
	}
	if newStatus == "SPLIT" {
		_, _ = tx.ExecContext(ctx, `UPDATE workflow_instances SET status='CANCELLED',cancelled_at=NOW(),updated_at=NOW() WHERE scope_type='BATCH' AND scope_id=$1 AND status<>'CANCELLED'`, id)
		_, _ = tx.ExecContext(ctx, `UPDATE action_items SET status='CANCELLED',updated_at=NOW() WHERE workflow_instance_id IN (SELECT id FROM workflow_instances WHERE scope_type='BATCH' AND scope_id=$1) AND status NOT IN ('COMPLETED','CANCELLED')`, id)
	}
	s.auditTx(ctx, tx, actor, "batches.split", "fulfillment_batch", id, map[string]any{"planned_quantity": parentQty}, map[string]any{"remaining": ratString(remaining), "children": out, "reason": p.Reason})
	if err = finishOperationTx(ctx, tx, actor, "BATCH_SPLIT", key, out); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *OperationsService) MergeBatches(ctx context.Context, actor, key string, p BatchMergePayload) (Batch, error) {
	var out Batch
	if len(p.BatchIDs) < 2 || requireReason(p.Reason) != nil {
		return out, ErrValidation
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return out, err
	}
	defer tx.Rollback()
	idem, err := claimOperationTx(ctx, tx, actor, "BATCH_MERGE", key, p)
	if err != nil {
		return out, err
	}
	if idem.Existing {
		if err = json.Unmarshal(idem.Response, &out); err != nil {
			return out, err
		}
		return out, nil
	}
	rows, err := tx.QueryContext(ctx, `SELECT b.id,b.order_id,b.order_item_id,b.quantity_unit,b.stone_name,COALESCE(b.stone_variant,''),COALESCE(b.finish_type,''),COALESCE(b.cut_type,''),b.planned_quantity::text,b.status,o.order_number,EXISTS(SELECT 1 FROM shipment_items WHERE batch_id=b.id AND loaded_quantity>0) FROM fulfillment_batches b JOIN orders o ON o.id=b.order_id WHERE b.id=ANY($1::uuid[]) ORDER BY b.id FOR UPDATE OF b,o`, pq.Array(p.BatchIDs))
	if err != nil {
		return out, err
	}
	defer rows.Close()
	var orderID, itemID, unit, stone, variant, finish, cut, orderNumber string
	total := new(big.Rat)
	count := 0
	for rows.Next() {
		var id, q, status, oid, iid, u, n, v, f, c, on string
		var shipped bool
		if err = rows.Scan(&id, &oid, &iid, &u, &n, &v, &f, &c, &q, &status, &on, &shipped); err != nil {
			return out, err
		}
		if count == 0 {
			orderID, itemID, unit, stone, variant, finish, cut, orderNumber = oid, iid, u, n, v, f, c, on
		} else if orderID != oid || itemID != iid || unit != u || stone != n || variant != v || finish != f || cut != c {
			return out, conflict("INCOMPATIBLE_BATCHES", "batches are not compatible")
		}
		if status == "SHIPPED" || status == "PARTIALLY_SHIPPED" || status == "DELIVERED" || status == "PARTIALLY_DELIVERED" || status == "SPLIT" || status == "MERGED" || status == "CANCELLED" {
			return out, conflict("INVALID_BATCH_STATE", "shipped or terminal batch cannot be merged")
		}
		if shipped {
			return out, conflict("INVALID_BATCH_STATE", "loaded batch cannot be merged")
		}
		qr, _ := new(big.Rat).SetString(q)
		total.Add(total, qr)
		count++
	}
	if err = rows.Err(); err != nil {
		return out, err
	}
	if err = rows.Close(); err != nil {
		return out, err
	}
	if count != len(p.BatchIDs) {
		return out, sql.ErrNoRows
	}
	var seq int
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*)+1 FROM fulfillment_batches WHERE order_id=$1`, orderID).Scan(&seq); err != nil {
		return out, err
	}
	number := fmt.Sprintf("%s-B%02d", orderNumber, seq)
	err = tx.QueryRowContext(ctx, `INSERT INTO fulfillment_batches(batch_number,order_id,order_item_id,source_type,stone_name,stone_variant,finish_type,cut_type,planned_quantity,quantity_unit,status,priority,is_required,created_by_user_id) VALUES($1,$2,$3,'FACTORY',$4,$5,$6,$7,$8::numeric,$9,'PLANNED','NORMAL',TRUE,$10) RETURNING id`, number, orderID, itemID, stone, variant, finish, cut, ratString(total), unit, actor).Scan(&out.ID)
	if err != nil {
		return out, err
	}
	for _, sourceID := range p.BatchIDs {
		_, err = tx.ExecContext(ctx, `INSERT INTO batch_merge_members(merged_batch_id,source_batch_id,quantity) SELECT $1,id,planned_quantity FROM fulfillment_batches WHERE id=$2`, out.ID, sourceID)
		if err != nil {
			return out, err
		}
		_, err = tx.ExecContext(ctx, `UPDATE fulfillment_batches SET status='MERGED',updated_at=NOW() WHERE id=$1`, sourceID)
		if err != nil {
			return out, err
		}
		_, _ = tx.ExecContext(ctx, `UPDATE workflow_instances SET status='CANCELLED',cancelled_at=NOW(),updated_at=NOW() WHERE scope_type='BATCH' AND scope_id=$1 AND status<>'CANCELLED'`, sourceID)
		_, _ = tx.ExecContext(ctx, `UPDATE action_items SET status='CANCELLED',updated_at=NOW() WHERE workflow_instance_id IN (SELECT id FROM workflow_instances WHERE scope_type='BATCH' AND scope_id=$1) AND status NOT IN ('COMPLETED','CANCELLED')`, sourceID)
		_, err = tx.ExecContext(ctx, `UPDATE inventory_reservations SET batch_id=$2,updated_at=NOW() WHERE batch_id=$1 AND status IN ('ACTIVE','PARTIALLY_CONSUMED')`, sourceID, out.ID)
		if err != nil {
			return out, err
		}
	}
	wid, err := s.startScopedWorkflowTx(ctx, tx, actor, orderID, "BATCH", out.ID, p.WorkflowTemplateID, nil, nil)
	if err != nil {
		return out, err
	}
	_, err = tx.ExecContext(ctx, `UPDATE fulfillment_batches SET workflow_instance_id=$2 WHERE id=$1`, out.ID, wid)
	if err != nil {
		return out, err
	}
	out = Batch{ID: out.ID, BatchNumber: number, OrderID: orderID, OrderNumber: orderNumber, OrderItemID: itemID, StoneName: stone, StoneVariant: variant, PlannedQuantity: ratString(total), ActualQuantity: "0.0000", QuantityUnit: unit, Status: "PLANNED", Priority: "NORMAL", IsRequired: true, WorkflowInstanceID: &wid}
	s.auditTx(ctx, tx, actor, "batches.merge", "fulfillment_batch", out.ID, map[string]any{"source_ids": p.BatchIDs}, map[string]any{"batch": out, "reason": p.Reason})
	if err = finishOperationTx(ctx, tx, actor, "BATCH_MERGE", key, out); err != nil {
		return out, err
	}
	if err = tx.Commit(); err != nil {
		return out, err
	}
	return out, nil
}

func (s *OperationsService) CancelBatch(ctx context.Context, actor, id, key, reason string) error {
	if err := requireReason(reason); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	idem, err := claimOperationTx(ctx, tx, actor, "BATCH_CANCEL", key, map[string]string{"batch_id": id, "reason": reason})
	if err != nil {
		return err
	}
	if idem.Existing {
		return tx.Commit()
	}
	var status string
	if err = tx.QueryRowContext(ctx, `SELECT status FROM fulfillment_batches WHERE id=$1 FOR UPDATE`, id).Scan(&status); err != nil {
		return err
	}
	if status == "DELIVERED" || status == "SPLIT" || status == "MERGED" {
		return conflict("INVALID_BATCH_STATE", "batch cannot be cancelled")
	}
	rows, err := tx.QueryContext(ctx, `SELECT r.id,r.inventory_lot_id,(r.reserved_quantity-r.consumed_quantity)::text,r.quantity_unit FROM inventory_reservations r WHERE r.batch_id=$1 AND r.status IN ('ACTIVE','PARTIALLY_CONSUMED') ORDER BY r.inventory_lot_id FOR UPDATE OF r`, id)
	if err != nil {
		return err
	}
	defer rows.Close()
	type release struct{ id, lot, q, unit string }
	items := []release{}
	for rows.Next() {
		var x release
		if err = rows.Scan(&x.id, &x.lot, &x.q, &x.unit); err != nil {
			return err
		}
		items = append(items, x)
	}
	for _, x := range items {
		var beforeA, beforeR string
		if err = tx.QueryRowContext(ctx, `SELECT available_quantity::text,reserved_quantity::text FROM inventory_lots WHERE id=$1 FOR UPDATE`, x.lot).Scan(&beforeA, &beforeR); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `UPDATE inventory_lots SET available_quantity=available_quantity+$2::numeric,reserved_quantity=reserved_quantity-$2::numeric,status=CASE WHEN reserved_quantity-$2::numeric=0 THEN 'AVAILABLE' ELSE 'PARTIALLY_RESERVED' END,updated_at=NOW() WHERE id=$1`, x.lot, x.q)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `UPDATE inventory_reservations SET status='CANCELLED',released_at=NOW(),release_reason=$2,updated_at=NOW() WHERE id=$1`, x.id, reason)
		if err != nil {
			return err
		}
		group := randomUUIDText()
		if err = s.insertMovementTx(ctx, tx, actor, group, "RESERVATION_RELEASE", x.lot, nil, nil, nil, &id, nil, &x.id, x.q, x.unit, beforeA, addDecimal(beforeA, x.q), beforeR, subDecimal(beforeR, x.q), "BATCH", id, reason, nil); err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, `UPDATE fulfillment_batches SET status='CANCELLED',cancelled_at=NOW(),updated_at=NOW() WHERE id=$1`, id)
	if err != nil {
		return err
	}
	_, _ = tx.ExecContext(ctx, `UPDATE workflow_instances SET status='CANCELLED',cancelled_at=NOW(),updated_at=NOW() WHERE scope_type='BATCH' AND scope_id=$1`, id)
	s.auditTx(ctx, tx, actor, "batches.cancel", "fulfillment_batch", id, map[string]any{"status": status}, map[string]any{"status": "CANCELLED", "reason": reason})
	if err = finishOperationTx(ctx, tx, actor, "BATCH_CANCEL", key, map[string]any{"cancelled": true}); err != nil {
		return err
	}
	return tx.Commit()
}
