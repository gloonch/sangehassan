package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type DocumentRecord struct {
	ID              string     `json:"id"`
	DocumentNumber  string     `json:"document_number"`
	DocumentType    string     `json:"document_type"`
	ScopeType       string     `json:"scope_type"`
	ScopeID         string     `json:"scope_id"`
	OrderID         *string    `json:"order_id,omitempty"`
	Version         int        `json:"version_number"`
	Status          string     `json:"status"`
	CustomerVisible bool       `json:"customer_visible"`
	CreatedAt       time.Time  `json:"created_at"`
	IssuedAt        *time.Time `json:"issued_at,omitempty"`
}
type DocumentRequirementPayload struct {
	StepID          *int64 `json:"workflow_template_step_id"`
	DocumentType    string `json:"document_type"`
	TitleFA         string `json:"title_fa"`
	IsRequired      bool   `json:"is_required"`
	IsBlocking      bool   `json:"is_blocking"`
	CustomerVisible bool   `json:"customer_visible"`
	SortOrder       int    `json:"sort_order"`
}

var allowedDocumentTypes = map[string]bool{
	"PROFORMA": true, "SALES_INVOICE": true, "PAYMENT_RECEIPT": true, "ORDER_SUMMARY": true,
	"PACKING_LIST": true, "DELIVERY_NOTE": true, "TRANSPORT_RECEIPT": true, "INSTALLATION_REPORT": true,
	"QUALITY_REPORT": true, "CUSTOMER_ACCEPTANCE": true, "COMMERCIAL_INVOICE": true, "EXPORT_PACKING_LIST": true,
	"CUSTOMS_DOCUMENT": true, "CERTIFICATE_OF_ORIGIN": true, "CUSTOMS_DECLARATION": true,
	"BILL_OF_LADING": true, "CERTIFICATE": true, "OTHER": true,
}

func (s *OperationsService) canAccessDocumentOrder(ctx context.Context, actor, orderID string) bool {
	if s.HasPermission(ctx, actor, "documents.view_all") || s.HasPermission(ctx, actor, "documents.view") {
		return true
	}
	if !s.HasPermission(ctx, actor, "documents.view_assigned") {
		return false
	}
	var allowed bool
	_ = s.db.QueryRowContext(ctx, `SELECT
		EXISTS(SELECT 1 FROM orders WHERE id=$1 AND sales_owner_user_id=$2)
		OR EXISTS(SELECT 1 FROM shipments WHERE order_id=$1 AND driver_user_id=$2)
		OR EXISTS(SELECT 1 FROM workflow_step_instances si JOIN workflow_instances wi ON wi.id=si.workflow_instance_id WHERE wi.order_id=$1 AND si.assigned_user_id=$2)
		OR EXISTS(SELECT 1 FROM fulfillment_batches b JOIN user_roles ur ON ur.user_id=$2 JOIN roles r ON r.id=ur.role_id AND r.code='SUPPLY' WHERE b.order_id=$1)`, orderID, actor).Scan(&allowed)
	return allowed
}

func (s *OperationsService) validateDocumentScope(ctx context.Context, orderID, documentType, scopeType, scopeID string) error {
	expectedScope := map[string]string{
		"PROFORMA": "ORDER", "ORDER_SUMMARY": "ORDER", "PAYMENT_RECEIPT": "PAYMENT",
		"PACKING_LIST": "SHIPMENT", "DELIVERY_NOTE": "SHIPMENT",
	}[documentType]
	if expectedScope != "" && scopeType != expectedScope {
		return ErrValidation
	}
	var valid bool
	switch scopeType {
	case "ORDER":
		valid = scopeID == orderID
	case "PAYMENT":
		if err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM customer_payments WHERE id=$1 AND order_id=$2)`, scopeID, orderID).Scan(&valid); err != nil {
			return err
		}
	case "SHIPMENT":
		if err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM shipments WHERE id=$1 AND order_id=$2)`, scopeID, orderID).Scan(&valid); err != nil {
			return err
		}
	case "BATCH":
		if err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM fulfillment_batches WHERE id=$1 AND order_id=$2)`, scopeID, orderID).Scan(&valid); err != nil {
			return err
		}
	case "WORKFLOW":
		if err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM workflow_instances WHERE id=$1 AND order_id=$2)`, scopeID, orderID).Scan(&valid); err != nil {
			return err
		}
	default:
		return ErrValidation
	}
	if !valid {
		return ErrForbidden
	}
	return nil
}

func (s *OperationsService) assembleDocumentSnapshot(ctx context.Context, orderID, documentType, scopeType, scopeID, number string) (map[string]any, string, error) {
	var orderNumber, customer, status, currency, subtotal, discount, tax, charges, total, paymentTerms, deliveryTerms string
	var delivery sql.NullTime
	err := s.db.QueryRowContext(ctx, `SELECT o.order_number,COALESCE(NULLIF(TRIM(CONCAT_WS(' ',u.first_name,u.last_name)),''),u.phone_normalized),o.status,t.currency,t.subtotal::text,t.discount_amount::text,t.tax_amount::text,t.additional_charge_amount::text,t.final_customer_amount::text,COALESCE(t.payment_terms_text,''),COALESCE(t.delivery_terms_text,''),o.estimated_delivery_at FROM orders o JOIN users u ON u.id=o.customer_user_id JOIN order_commercial_terms t ON t.order_id=o.id WHERE o.id=$1`, orderID).Scan(&orderNumber, &customer, &status, &currency, &subtotal, &discount, &tax, &charges, &total, &paymentTerms, &deliveryTerms, &delivery)
	if err != nil {
		return nil, "", err
	}
	snapshot := map[string]any{"document_number": number, "order_number": orderNumber, "customer_name": customer, "status": status, "currency": currency, "subtotal": subtotal, "discount": discount, "tax": tax, "charges": charges, "total": total, "payment_terms": paymentTerms, "delivery_terms": deliveryTerms, "issued_at": time.Now().UTC().Format(time.RFC3339)}
	if delivery.Valid {
		snapshot["estimated_delivery_at"] = delivery.Time.UTC().Format(time.RFC3339)
	}
	items := []map[string]any{}
	rows, err := s.db.QueryContext(ctx, `SELECT stone_name,ordered_quantity::text,quantity_unit,unit_price::text,discount_amount::text,line_amount::text,currency FROM order_items WHERE order_id=$1 ORDER BY created_at`, orderID)
	if err != nil {
		return nil, "", err
	}
	for rows.Next() {
		var name, q, u, price, d, line, c string
		if err = rows.Scan(&name, &q, &u, &price, &d, &line, &c); err != nil {
			rows.Close()
			return nil, "", err
		}
		items = append(items, map[string]any{"description": name, "quantity": q, "unit": u, "unit_price": price, "discount": d, "line_amount": line, "currency": c})
	}
	if err = rows.Close(); err != nil {
		return nil, "", err
	}
	snapshot["items"] = items
	if documentType == "PROFORMA" {
		var proformaNumber sql.NullString
		_ = s.db.QueryRowContext(ctx, `SELECT proforma_number FROM proformas WHERE order_id=$1 AND status<>'CANCELLED' ORDER BY COALESCE(issued_at,created_at) DESC LIMIT 1`, orderID).Scan(&proformaNumber)
		snapshot["proforma_number"] = scanNullableString(proformaNumber)
	}
	if documentType == "PAYMENT_RECEIPT" {
		var paymentNumber, amount, pcurrency, reference string
		var paid time.Time
		if err = s.db.QueryRowContext(ctx, `SELECT payment_number,amount::text,currency,COALESCE(reference_number,''),paid_at FROM customer_payments WHERE id=$1 AND order_id=$2`, scopeID, orderID).Scan(&paymentNumber, &amount, &pcurrency, &reference, &paid); err != nil {
			return nil, "", err
		}
		snapshot["payment_number"] = paymentNumber
		snapshot["amount"] = amount
		snapshot["currency"] = pcurrency
		snapshot["reference"] = reference
		snapshot["paid_at"] = paid.UTC().Format(time.RFC3339)
	}
	if documentType == "PACKING_LIST" || documentType == "DELIVERY_NOTE" {
		var shipmentNumber string
		if err = s.db.QueryRowContext(ctx, `SELECT shipment_number FROM shipments WHERE id=$1 AND order_id=$2`, scopeID, orderID).Scan(&shipmentNumber); err != nil {
			return nil, "", err
		}
		snapshot["shipment_number"] = shipmentNumber
		packageRows, packageErr := s.db.QueryContext(ctx, `SELECT p.package_number,p.quantity::text,p.quantity_unit,p.gross_weight::text,p.net_weight::text,COALESCE(p.weight_unit,'') FROM packaging_units p JOIN shipment_package_assignments a ON a.packaging_unit_id=p.id AND a.released_at IS NULL JOIN shipment_items si ON si.id=a.shipment_item_id WHERE si.shipment_id=$1 ORDER BY p.package_number`, scopeID)
		if packageErr != nil {
			return nil, "", packageErr
		}
		packages := []map[string]any{}
		for packageRows.Next() {
			var packageNumber, quantity, unit, weightUnit string
			var gross, net sql.NullString
			if packageErr = packageRows.Scan(&packageNumber, &quantity, &unit, &gross, &net, &weightUnit); packageErr != nil {
				packageRows.Close()
				return nil, "", packageErr
			}
			packages = append(packages, map[string]any{"package_number": packageNumber, "quantity": quantity, "unit": unit, "gross_weight": scanNullableString(gross), "net_weight": scanNullableString(net), "weight_unit": weightUnit})
		}
		if packageErr = packageRows.Close(); packageErr != nil {
			return nil, "", packageErr
		}
		snapshot["packages"] = packages
		containerRows, containerErr := s.db.QueryContext(ctx, `SELECT container_number,container_type,COALESCE(seal_number,''),gross_weight::text,net_weight::text,COALESCE(weight_unit,'') FROM shipment_containers WHERE shipment_id=$1 ORDER BY created_at`, scopeID)
		if containerErr != nil {
			return nil, "", containerErr
		}
		containers := []map[string]any{}
		for containerRows.Next() {
			var containerNumber, containerType, seal, weightUnit string
			var gross, net sql.NullString
			if containerErr = containerRows.Scan(&containerNumber, &containerType, &seal, &gross, &net, &weightUnit); containerErr != nil {
				containerRows.Close()
				return nil, "", containerErr
			}
			containers = append(containers, map[string]any{"container_number": containerNumber, "container_type": containerType, "seal_number": seal, "gross_weight": scanNullableString(gross), "net_weight": scanNullableString(net), "weight_unit": weightUnit})
		}
		if containerErr = containerRows.Close(); containerErr != nil {
			return nil, "", containerErr
		}
		snapshot["containers"] = containers
		if documentType == "DELIVERY_NOTE" {
			var receiver sql.NullString
			var delivered sql.NullTime
			_ = s.db.QueryRowContext(ctx, `SELECT receiver_name,occurred_at FROM shipment_events WHERE shipment_id=$1 AND event_type='DELIVERY' ORDER BY occurred_at DESC LIMIT 1`, scopeID).Scan(&receiver, &delivered)
			snapshot["receiver_name"] = scanNullableString(receiver)
			snapshot["delivered_at"] = nullableTime(delivered)
		}
	}
	title := map[string]string{"PROFORMA": "پیش‌فاکتور", "PAYMENT_RECEIPT": "رسید پرداخت", "ORDER_SUMMARY": "خلاصه سفارش", "PACKING_LIST": "فهرست بسته‌بندی", "DELIVERY_NOTE": "رسید تحویل"}[documentType]
	if title == "" {
		title = "سند سفارش"
	}
	return snapshot, title, nil
}

func (s *OperationsService) GenerateDocument(ctx context.Context, actor, orderID, key string, p DocumentGeneratePayload) (DocumentRecord, error) {
	var out DocumentRecord
	p.DocumentType = normalizeCode(p.DocumentType)
	p.ScopeType = normalizeCode(p.ScopeType)
	if !allowedDocumentTypes[p.DocumentType] || p.ScopeType == "" || p.ScopeID == "" {
		return out, ErrValidation
	}
	if !s.canAccessDocumentOrder(ctx, actor, orderID) {
		return out, ErrForbidden
	}
	if err := s.validateDocumentScope(ctx, orderID, p.DocumentType, p.ScopeType, p.ScopeID); err != nil {
		return out, err
	}
	var owner string
	if err := s.db.QueryRowContext(ctx, `SELECT customer_user_id FROM orders WHERE id=$1`, orderID).Scan(&owner); err != nil {
		return out, err
	}
	documentID := randomUUIDText()
	number := "DOC-" + time.Now().UTC().Format("20060102-150405") + "-" + randomDigits(4)
	snapshot, title, err := s.assembleDocumentSnapshot(ctx, orderID, p.DocumentType, p.ScopeType, p.ScopeID, number)
	if err != nil {
		return out, err
	}
	pdfBytes, err := generatePersianPDF(title, snapshot)
	if err != nil {
		return out, err
	}
	if err = os.MkdirAll(s.documentDir, 0700); err != nil {
		return out, err
	}
	storageKey := filepath.Join("documents", documentID+".pdf")
	finalPath := filepath.Join(s.documentDir, storageKey)
	if err = os.MkdirAll(filepath.Dir(finalPath), 0700); err != nil {
		return out, err
	}
	temp, err := os.CreateTemp(filepath.Dir(finalPath), ".document-*.tmp")
	if err != nil {
		return out, err
	}
	tempPath := temp.Name()
	cleanup := func() { _ = os.Remove(tempPath); _ = os.Remove(finalPath) }
	if err = temp.Chmod(0600); err == nil {
		_, err = temp.Write(pdfBytes)
	}
	closeErr := temp.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		cleanup()
		return out, err
	}
	if err = os.Rename(tempPath, finalPath); err != nil {
		cleanup()
		return out, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		cleanup()
		return out, err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, "DOCUMENT_GENERATE", key, map[string]any{"order_id": orderID, "payload": p})
	if err != nil {
		cleanup()
		return out, err
	}
	if claim.Existing {
		cleanup()
		if err = json.Unmarshal(claim.Response, &out); err != nil {
			return out, err
		}
		return out, tx.Commit()
	}
	var templateID sql.NullString
	templateSnapshot := []byte(`{}`)
	if err = tx.QueryRowContext(ctx, `SELECT id,template_json FROM document_templates WHERE document_type=$1 AND is_active=TRUE ORDER BY version_number DESC LIMIT 1`, p.DocumentType).Scan(&templateID, &templateSnapshot); err != nil && !errors.Is(err, sql.ErrNoRows) {
		cleanup()
		return out, err
	}
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	var version int
	if err = tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(version_number),0)+1 FROM documents WHERE document_type=$1 AND scope_type=$2 AND scope_id=$3`, p.DocumentType, p.ScopeType, p.ScopeID).Scan(&version); err != nil {
		cleanup()
		return out, err
	}
	fileID := randomUUIDText()
	raw, _ := json.Marshal(snapshot)
	_, err = tx.ExecContext(ctx, `INSERT INTO workflow_files(id,workflow_instance_id,workflow_step_instance_id,storage_key,original_file_name,mime_type,size_bytes,customer_visible,uploaded_by_user_id,entity_type,entity_id) VALUES($1,NULL,NULL,$2,$3,'application/pdf',$4,$5,$6,'DOCUMENT',$7)`, fileID, storageKey, number+".pdf", len(pdfBytes), p.CustomerVisible, actor, documentID)
	if err != nil {
		cleanup()
		return out, err
	}
	err = tx.QueryRowContext(ctx, `INSERT INTO documents(id,document_number,document_type,scope_type,scope_id,order_id,customer_user_id,document_template_id,version_number,status,template_snapshot_json,snapshot_json,workflow_file_id,customer_visible,generated_by_user_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,'DRAFT',$10::jsonb,$11::jsonb,$12,$13,$14) RETURNING id,document_number,document_type,scope_type,scope_id,order_id,version_number,status,customer_visible,created_at,issued_at`, documentID, number, p.DocumentType, p.ScopeType, p.ScopeID, orderID, owner, templateID, version, string(templateSnapshot), string(raw), fileID, p.CustomerVisible, actor).Scan(&out.ID, &out.DocumentNumber, &out.DocumentType, &out.ScopeType, &out.ScopeID, &out.OrderID, &out.Version, &out.Status, &out.CustomerVisible, &out.CreatedAt, &out.IssuedAt)
	if err != nil {
		cleanup()
		return out, err
	}
	if err = auditTx(ctx, tx, actor, "documents.generate", "document", out.ID, nil, out); err != nil {
		cleanup()
		return out, err
	}
	if err = finishOperationTx(ctx, tx, actor, "DOCUMENT_GENERATE", key, out); err != nil {
		cleanup()
		return out, err
	}
	if err = tx.Commit(); err != nil {
		cleanup()
		return out, err
	}
	return out, nil
}

func (s *OperationsService) RegisterUploadedDocument(ctx context.Context, actor, orderID, key string, p DocumentGeneratePayload, storageKey, originalName, mimeType string, size int64) (DocumentRecord, bool, error) {
	var out DocumentRecord
	p.DocumentType = normalizeCode(p.DocumentType)
	p.ScopeType = normalizeCode(p.ScopeType)
	if !allowedDocumentTypes[p.DocumentType] || p.ScopeID == "" || mimeType != "application/pdf" {
		return out, false, ErrValidation
	}
	if !s.canAccessDocumentOrder(ctx, actor, orderID) {
		return out, false, ErrForbidden
	}
	if err := s.validateDocumentScope(ctx, orderID, p.DocumentType, p.ScopeType, p.ScopeID); err != nil {
		return out, false, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return out, false, err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, "DOCUMENT_UPLOAD", key, map[string]any{"order_id": orderID, "document_type": p.DocumentType, "scope_type": p.ScopeType, "scope_id": p.ScopeID, "original_name": originalName, "size": size})
	if err != nil {
		return out, false, err
	}
	if claim.Existing {
		if err = json.Unmarshal(claim.Response, &out); err != nil {
			return out, false, err
		}
		return out, false, tx.Commit()
	}
	var owner string
	if err = tx.QueryRowContext(ctx, `SELECT customer_user_id FROM orders WHERE id=$1 FOR SHARE`, orderID).Scan(&owner); err != nil {
		return out, false, err
	}
	documentID := randomUUIDText()
	number := "DOC-" + time.Now().UTC().Format("20060102-150405") + "-" + randomDigits(4)
	var version int
	if err = tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(version_number),0)+1 FROM documents WHERE document_type=$1 AND scope_type=$2 AND scope_id=$3`, p.DocumentType, p.ScopeType, p.ScopeID).Scan(&version); err != nil {
		return out, false, err
	}
	snapshot := map[string]any{"document_number": number, "uploaded_file_name": originalName, "uploaded_at": time.Now().UTC()}
	raw, _ := json.Marshal(snapshot)
	fileID := randomUUIDText()
	if _, err = tx.ExecContext(ctx, `INSERT INTO workflow_files(id,storage_key,original_file_name,mime_type,size_bytes,customer_visible,uploaded_by_user_id,entity_type,entity_id) VALUES($1,$2,$3,$4,$5,$6,$7,'DOCUMENT',$8)`, fileID, storageKey, originalName, mimeType, size, p.CustomerVisible, actor, documentID); err != nil {
		return out, false, err
	}
	err = tx.QueryRowContext(ctx, `INSERT INTO documents(id,document_number,document_type,scope_type,scope_id,order_id,customer_user_id,version_number,status,snapshot_json,workflow_file_id,customer_visible,generated_by_user_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8,'DRAFT',$9::jsonb,$10,$11,$12) RETURNING id,document_number,document_type,scope_type,scope_id,order_id,version_number,status,customer_visible,created_at,issued_at`, documentID, number, p.DocumentType, p.ScopeType, p.ScopeID, orderID, owner, version, string(raw), fileID, p.CustomerVisible, actor).Scan(&out.ID, &out.DocumentNumber, &out.DocumentType, &out.ScopeType, &out.ScopeID, &out.OrderID, &out.Version, &out.Status, &out.CustomerVisible, &out.CreatedAt, &out.IssuedAt)
	if err != nil {
		return out, false, err
	}
	if err = finishOperationTx(ctx, tx, actor, "DOCUMENT_UPLOAD", key, out); err != nil {
		return out, false, err
	}
	return out, true, tx.Commit()
}

func (s *OperationsService) ListDocuments(ctx context.Context, actor, orderID string, customer bool) ([]DocumentRecord, error) {
	where := `d.order_id=$1`
	args := []any{orderID}
	if customer {
		where += ` AND d.customer_user_id=$2 AND d.customer_visible AND d.status='ISSUED'`
		args = append(args, actor)
	} else if !s.canAccessDocumentOrder(ctx, actor, orderID) {
		return nil, ErrForbidden
	}
	rows, err := s.db.QueryContext(ctx, `SELECT d.id,d.document_number,d.document_type,d.scope_type,d.scope_id,d.order_id,d.version_number,d.status,d.customer_visible,d.created_at,d.issued_at FROM documents d WHERE `+where+` ORDER BY d.created_at DESC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []DocumentRecord{}
	for rows.Next() {
		var d DocumentRecord
		var oid sql.NullString
		var issued sql.NullTime
		if err = rows.Scan(&d.ID, &d.DocumentNumber, &d.DocumentType, &d.ScopeType, &d.ScopeID, &oid, &d.Version, &d.Status, &d.CustomerVisible, &d.CreatedAt, &issued); err != nil {
			return nil, err
		}
		d.OrderID = scanNullableString(oid)
		if issued.Valid {
			d.IssuedAt = &issued.Time
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *OperationsService) IssueDocument(ctx context.Context, actor, id, key string) (DocumentRecord, error) {
	var out DocumentRecord
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return out, err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, "DOCUMENT_ISSUE", key, map[string]string{"id": id})
	if err != nil {
		return out, err
	}
	if claim.Existing {
		if err = json.Unmarshal(claim.Response, &out); err != nil {
			return out, err
		}
		return out, tx.Commit()
	}
	var status string
	if err = tx.QueryRowContext(ctx, `SELECT status FROM documents WHERE id=$1 FOR UPDATE`, id).Scan(&status); err != nil {
		return out, err
	}
	if status != "DRAFT" {
		return out, conflict(ErrInvalidFinancialTransition, "only draft document can be issued")
	}
	var supersededID sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT old.id FROM documents current JOIN documents old ON old.document_type=current.document_type AND old.scope_type=current.scope_type AND old.scope_id=current.scope_id WHERE current.id=$1 AND old.id<>current.id AND old.status='ISSUED' ORDER BY old.version_number DESC LIMIT 1 FOR UPDATE OF old`, id).Scan(&supersededID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return out, err
	}
	if supersededID.Valid {
		if _, err = tx.ExecContext(ctx, `UPDATE documents SET status='SUPERSEDED' WHERE id=$1`, supersededID.String); err != nil {
			return out, err
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE documents SET status='ISSUED',supersedes_document_id=$3,issued_by_user_id=$2,issued_at=NOW() WHERE id=$1`, id, actor, scanNullableString(supersededID)); err != nil {
		return out, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE workflow_instance_document_requirements r SET status='SATISFIED',satisfied_document_id=d.id,satisfied_at=NOW() FROM documents d JOIN workflow_instances wi ON wi.order_id=d.order_id WHERE d.id=$1 AND r.workflow_instance_id=wi.id AND r.document_type=d.document_type AND r.status IN ('PENDING','SATISFIED')`, id); err != nil {
		return out, err
	}
	if err = scanDocumentTx(ctx, tx, id, &out); err != nil {
		return out, err
	}
	if err = auditTx(ctx, tx, actor, "documents.issue", "document", id, map[string]string{"status": status}, out); err != nil {
		return out, err
	}
	if err = finishOperationTx(ctx, tx, actor, "DOCUMENT_ISSUE", key, out); err != nil {
		return out, err
	}
	return out, tx.Commit()
}
func scanDocumentTx(ctx context.Context, tx *sql.Tx, id string, d *DocumentRecord) error {
	var oid sql.NullString
	var issued sql.NullTime
	err := tx.QueryRowContext(ctx, `SELECT id,document_number,document_type,scope_type,scope_id,order_id,version_number,status,customer_visible,created_at,issued_at FROM documents WHERE id=$1`, id).Scan(&d.ID, &d.DocumentNumber, &d.DocumentType, &d.ScopeType, &d.ScopeID, &oid, &d.Version, &d.Status, &d.CustomerVisible, &d.CreatedAt, &issued)
	d.OrderID = scanNullableString(oid)
	if issued.Valid {
		d.IssuedAt = &issued.Time
	}
	return err
}
func (s *OperationsService) CancelDocument(ctx context.Context, actor, id, key, reason string) error {
	if requireReason(reason) != nil {
		return ErrValidation
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, "DOCUMENT_CANCEL", key, map[string]string{"id": id, "reason": reason})
	if err != nil {
		return err
	}
	if claim.Existing {
		return tx.Commit()
	}
	r, err := tx.ExecContext(ctx, `UPDATE documents SET status='CANCELLED',cancelled_by_user_id=$2,cancelled_at=NOW(),cancellation_reason=$3 WHERE id=$1 AND status IN ('DRAFT','ISSUED')`, id, actor, reason)
	if err != nil {
		return err
	}
	if n, _ := r.RowsAffected(); n == 0 {
		return conflict(ErrInvalidFinancialTransition, "document cannot be cancelled")
	}
	_, err = tx.ExecContext(ctx, `UPDATE workflow_instance_document_requirements SET status='PENDING',satisfied_document_id=NULL,satisfied_at=NULL WHERE satisfied_document_id=$1`, id)
	if err != nil {
		return err
	}
	if err = finishOperationTx(ctx, tx, actor, "DOCUMENT_CANCEL", key, map[string]bool{"cancelled": true}); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *OperationsService) DocumentDownload(ctx context.Context, actor, id string, customer bool) (string, string, error) {
	var key, name, owner, orderID string
	var visible bool
	var status string
	err := s.db.QueryRowContext(ctx, `SELECT f.storage_key,f.original_file_name,d.customer_user_id,d.customer_visible,d.status,d.order_id FROM documents d JOIN workflow_files f ON f.id=d.workflow_file_id WHERE d.id=$1`, id).Scan(&key, &name, &owner, &visible, &status, &orderID)
	if err != nil {
		return "", "", err
	}
	if customer && (owner != actor || !visible || status != "ISSUED") {
		return "", "", ErrForbidden
	}
	if !customer && !s.HasPermission(ctx, actor, "documents.download") && !s.HasPermission(ctx, actor, "documents.download_internal") {
		return "", "", ErrForbidden
	}
	if !customer && !s.canAccessDocumentOrder(ctx, actor, orderID) {
		return "", "", ErrForbidden
	}
	clean := filepath.Clean(key)
	base := filepath.Clean(s.documentDir)
	path := filepath.Join(base, clean)
	if !strings.HasPrefix(path, base+string(os.PathSeparator)) {
		return "", "", ErrForbidden
	}
	return path, name, nil
}

func (s *OperationsService) ListDocumentRequirements(ctx context.Context, templateID int64) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,workflow_template_step_id,document_type,title_fa,is_required,is_blocking,customer_visible,sort_order FROM workflow_template_document_requirements WHERE workflow_template_id=$1 ORDER BY sort_order,id`, templateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, dt, title string
		var step sql.NullInt64
		var required, blocking, visible bool
		var sort int
		if err = rows.Scan(&id, &step, &dt, &title, &required, &blocking, &visible, &sort); err != nil {
			return nil, err
		}
		var stepValue any
		if step.Valid {
			stepValue = step.Int64
		}
		out = append(out, map[string]any{"id": id, "workflow_template_step_id": stepValue, "document_type": dt, "title_fa": title, "is_required": required, "is_blocking": blocking, "customer_visible": visible, "sort_order": sort})
	}
	return out, rows.Err()
}
func (s *OperationsService) SaveDocumentRequirement(ctx context.Context, actor string, templateID int64, p DocumentRequirementPayload) (map[string]any, error) {
	p.DocumentType = normalizeCode(p.DocumentType)
	if !allowedDocumentTypes[p.DocumentType] || strings.TrimSpace(p.TitleFA) == "" {
		return nil, ErrValidation
	}
	var status string
	if err := s.db.QueryRowContext(ctx, `SELECT status FROM workflow_templates WHERE id=$1`, templateID).Scan(&status); err != nil {
		return nil, err
	}
	if status != "DRAFT" {
		return nil, ErrTemplateImmutable
	}
	if p.StepID != nil {
		var valid bool
		if err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM workflow_template_steps WHERE id=$1 AND workflow_template_id=$2)`, *p.StepID, templateID).Scan(&valid); err != nil {
			return nil, err
		}
		if !valid {
			return nil, ErrValidation
		}
	}
	var id string
	err := s.db.QueryRowContext(ctx, `INSERT INTO workflow_template_document_requirements(workflow_template_id,workflow_template_step_id,document_type,title_fa,is_required,is_blocking,customer_visible,sort_order) VALUES($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT(workflow_template_id,workflow_template_step_id,document_type) DO UPDATE SET title_fa=EXCLUDED.title_fa,is_required=EXCLUDED.is_required,is_blocking=EXCLUDED.is_blocking,customer_visible=EXCLUDED.customer_visible,sort_order=EXCLUDED.sort_order RETURNING id`, templateID, p.StepID, p.DocumentType, p.TitleFA, p.IsRequired, p.IsBlocking, p.CustomerVisible, p.SortOrder).Scan(&id)
	if err != nil {
		return nil, err
	}
	s.audit(ctx, actor, "workflow_document_requirements.save", "workflow_template", fmt.Sprint(templateID), p)
	return map[string]any{"id": id}, nil
}
func (s *OperationsService) DeleteDocumentRequirement(ctx context.Context, actor string, templateID int64, id string) error {
	r, err := s.db.ExecContext(ctx, `DELETE FROM workflow_template_document_requirements WHERE id=$1 AND workflow_template_id=$2 AND EXISTS(SELECT 1 FROM workflow_templates WHERE id=$2 AND status='DRAFT')`, id, templateID)
	if err != nil {
		return err
	}
	if n, _ := r.RowsAffected(); n == 0 {
		return ErrTemplateImmutable
	}
	s.audit(ctx, actor, "workflow_document_requirements.delete", "workflow_template", fmt.Sprint(templateID), map[string]string{"id": id})
	return nil
}

func (s *OperationsService) CleanupOrphanDocumentFiles(ctx context.Context) (int, error) {
	entries, err := os.ReadDir(filepath.Join(s.documentDir, "documents"))
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	removed := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		key := filepath.Join("documents", entry.Name())
		var exists bool
		if err = s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM workflow_files WHERE storage_key=$1)`, key).Scan(&exists); err != nil {
			return removed, err
		}
		if !exists {
			if info, e := entry.Info(); e == nil && time.Since(info.ModTime()) > time.Hour {
				if e = os.Remove(filepath.Join(s.documentDir, key)); e == nil {
					removed++
				}
			}
		}
	}
	return removed, nil
}

func (s *OperationsService) ListDocumentTemplates(ctx context.Context) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,code,name_fa,document_type,locale,template_json,allowed_variables,version_number,is_active,updated_at FROM document_templates ORDER BY document_type,version_number DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, code, name, kind, locale string
		var template, allowed []byte
		var version int
		var active bool
		var updated time.Time
		if err = rows.Scan(&id, &code, &name, &kind, &locale, &template, &allowed, &version, &active, &updated); err != nil {
			return nil, err
		}
		var t map[string]any
		var vars []string
		_ = json.Unmarshal(template, &t)
		_ = json.Unmarshal(allowed, &vars)
		out = append(out, map[string]any{"id": id, "code": code, "name_fa": name, "document_type": kind, "locale": locale, "template_json": t, "allowed_variables": vars, "version_number": version, "is_active": active, "updated_at": updated})
	}
	return out, rows.Err()
}

func validateDocumentTemplateJSON(t map[string]any) error {
	if len(t) == 0 {
		return ErrValidation
	}
	allowedSections := map[string]bool{"customer": true, "items": true, "totals": true, "terms": true, "payment": true, "delivery": true, "shipment": true, "packages": true, "receiver": true}
	for key, value := range t {
		if key != "title" && key != "sections" && key != "footer" {
			return errors.New("unsupported document template key")
		}
		if text, ok := value.(string); ok && (strings.Contains(strings.ToLower(text), "<script") || strings.Contains(text, "{{") || strings.Contains(text, "}}")) {
			return errors.New("document templates cannot contain code or expressions")
		}
	}
	if _, ok := t["title"].(string); !ok {
		return errors.New("document template title is required")
	}
	sections, ok := t["sections"].([]any)
	if !ok || len(sections) == 0 {
		return errors.New("document template sections are required")
	}
	for _, section := range sections {
		name, ok := section.(string)
		if !ok || !allowedSections[name] {
			return errors.New("unsupported document template section")
		}
	}
	return nil
}

func (s *OperationsService) UpdateDocumentTemplate(ctx context.Context, actor, id, key string, p DocumentTemplateUpdate) error {
	if strings.TrimSpace(p.NameFA) == "" {
		return ErrValidation
	}
	if err := validateDocumentTemplateJSON(p.Template); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, "DOCUMENT_TEMPLATE_UPDATE", key, map[string]any{"id": id, "payload": p})
	if err != nil {
		return err
	}
	if claim.Existing {
		return tx.Commit()
	}
	raw, _ := json.Marshal(p.Template)
	r, err := tx.ExecContext(ctx, `UPDATE document_templates SET name_fa=$2,template_json=$3::jsonb,is_active=$4,version_number=version_number+1,updated_at=NOW() WHERE id=$1`, id, p.NameFA, string(raw), p.IsActive)
	if err != nil {
		return err
	}
	if n, _ := r.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	if err = finishOperationTx(ctx, tx, actor, "DOCUMENT_TEMPLATE_UPDATE", key, map[string]bool{"updated": true}); err != nil {
		return err
	}
	return tx.Commit()
}
