package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"
)

const installationSelect = `SELECT j.id,j.installation_number,j.order_id,o.order_number,j.customer_user_id,COALESCE(j.project_name,''),COALESCE(j.project_address,''),COALESCE(j.contact_name,''),COALESCE(j.contact_phone,''),j.status,j.installation_lead_user_id,j.supplier_id,s.name,j.planned_start_at,j.actual_start_at,j.estimated_end_at,j.actual_end_at,j.planned_area::text,j.installed_area::text,j.area_unit,COALESCE(j.notes,''),j.workflow_instance_id,j.created_at,j.updated_at FROM installation_jobs j JOIN orders o ON o.id=j.order_id LEFT JOIN suppliers s ON s.id=j.supplier_id`

func scanInstallation(row interface{ Scan(...any) error }) (InstallationJob, error) {
	var out InstallationJob
	var lead, supplier, supplierName, workflow sql.NullString
	var plannedStart, actualStart, estimatedEnd, actualEnd sql.NullTime
	var plannedArea sql.NullString
	err := row.Scan(&out.ID, &out.InstallationNumber, &out.OrderID, &out.OrderNumber, &out.CustomerUserID, &out.ProjectName, &out.ProjectAddress, &out.ContactName, &out.ContactPhone, &out.Status, &lead, &supplier, &supplierName, &plannedStart, &actualStart, &estimatedEnd, &actualEnd, &plannedArea, &out.InstalledArea, &out.AreaUnit, &out.Notes, &workflow, &out.CreatedAt, &out.UpdatedAt)
	out.InstallationLeadUserID = scanNullableString(lead)
	out.SupplierID, out.SupplierName = scanNullableString(supplier), scanNullableString(supplierName)
	out.PlannedStartAt, out.ActualStartAt = scanNullableTime(plannedStart), scanNullableTime(actualStart)
	out.EstimatedEndAt, out.ActualEndAt = scanNullableTime(estimatedEnd), scanNullableTime(actualEnd)
	out.PlannedArea = scanNullableString(plannedArea)
	out.WorkflowInstanceID = scanNullableString(workflow)
	return out, err
}

func installationAccessClause(viewAll bool) string {
	if viewAll {
		return "TRUE"
	}
	return `(j.installation_lead_user_id=$1 OR j.created_by_user_id=$1 OR o.sales_owner_user_id=$1 OR EXISTS(SELECT 1 FROM installation_job_members m WHERE m.installation_job_id=j.id AND m.user_id=$1))`
}

func (s *OperationsService) ListInstallations(ctx context.Context, actor, status, orderID string) ([]InstallationJob, error) {
	viewAll := s.HasPermission(ctx, actor, "installation.view_all")
	rows, err := s.db.QueryContext(ctx, installationSelect+` WHERE `+installationAccessClause(viewAll)+` AND ($2='' OR j.status=UPPER($2)) AND ($3='' OR j.order_id=$3::uuid) ORDER BY COALESCE(j.planned_start_at,j.created_at),j.created_at DESC`, actor, status, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []InstallationJob{}
	for rows.Next() {
		x, scanErr := scanInstallation(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		x.Notes = ""
		out = append(out, x)
	}
	return out, rows.Err()
}

func (s *OperationsService) canAccessInstallation(ctx context.Context, actor, id string, customer bool) (bool, error) {
	var ok bool
	if customer {
		err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM installation_jobs WHERE id=$1 AND customer_user_id=$2)`, id, actor).Scan(&ok)
		return ok, err
	}
	if s.HasPermission(ctx, actor, "installation.view_all") {
		return true, nil
	}
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM installation_jobs j JOIN orders o ON o.id=j.order_id WHERE j.id=$1 AND (j.installation_lead_user_id=$2 OR j.created_by_user_id=$2 OR o.sales_owner_user_id=$2 OR EXISTS(SELECT 1 FROM installation_job_members m WHERE m.installation_job_id=j.id AND m.user_id=$2)))`, id, actor).Scan(&ok)
	return ok, err
}

func (s *OperationsService) GetInstallation(ctx context.Context, actor, id string, customer bool) (InstallationJob, error) {
	var out InstallationJob
	ok, err := s.canAccessInstallation(ctx, actor, id, customer)
	if err != nil {
		return out, err
	}
	if !ok {
		return out, ErrForbidden
	}
	out, err = scanInstallation(s.db.QueryRowContext(ctx, installationSelect+` WHERE j.id=$1`, id))
	if err != nil {
		return out, err
	}
	if customer {
		out.Notes = ""
	}
	if !customer {
		rows, queryErr := s.db.QueryContext(ctx, `SELECT m.id,m.user_id,COALESCE(NULLIF(TRIM(CONCAT_WS(' ',u.first_name,u.last_name)),''),m.name_override,''),COALESCE(m.role_label,'') FROM installation_job_members m LEFT JOIN users u ON u.id=m.user_id WHERE m.installation_job_id=$1 ORDER BY m.created_at`, id)
		if queryErr != nil {
			return out, queryErr
		}
		out.Members = []map[string]any{}
		for rows.Next() {
			var memberID, name, role string
			var user sql.NullString
			if err = rows.Scan(&memberID, &user, &name, &role); err != nil {
				rows.Close()
				return out, err
			}
			out.Members = append(out.Members, map[string]any{"id": memberID, "user_id": scanNullableString(user), "name": name, "role_label": role})
		}
		if err = rows.Close(); err != nil {
			return out, err
		}
	}
	visibility := ""
	if customer {
		visibility = " AND customer_visible"
	}
	rows, err := s.db.QueryContext(ctx, `SELECT u.id,u.update_date,u.installed_quantity::text,u.quantity_unit,u.status,COALESCE(u.description,''),u.customer_visible,u.created_at,COALESCE((SELECT jsonb_agg(a.activity_type ORDER BY a.activity_type) FROM installation_update_activities a WHERE a.installation_update_id=u.id),'[]'::jsonb) FROM installation_updates u WHERE u.installation_job_id=$1`+visibility+` ORDER BY u.update_date DESC,u.created_at DESC`, id)
	if err != nil {
		return out, err
	}
	out.Updates = []map[string]any{}
	for rows.Next() {
		var updateID, date, quantity, unit, status, description string
		var visible bool
		var created time.Time
		var activities []byte
		if err = rows.Scan(&updateID, &date, &quantity, &unit, &status, &description, &visible, &created, &activities); err != nil {
			rows.Close()
			return out, err
		}
		out.Updates = append(out.Updates, map[string]any{"id": updateID, "update_date": date, "installed_quantity": quantity, "quantity_unit": unit, "status": status, "description": description, "customer_visible": visible, "activities": json.RawMessage(activities), "created_at": created})
	}
	if err = rows.Close(); err != nil {
		return out, err
	}
	issueVisibility := ""
	if customer {
		issueVisibility = " AND customer_visible"
	}
	rows, err = s.db.QueryContext(ctx, `SELECT id,issue_type,severity,description,status,customer_visible,COALESCE(resolution_note,''),created_at,resolved_at FROM installation_issues WHERE installation_job_id=$1`+issueVisibility+` ORDER BY status='OPEN' DESC,created_at DESC`, id)
	if err != nil {
		return out, err
	}
	out.Issues = []map[string]any{}
	for rows.Next() {
		var issueID, kind, severity, description, status, resolution string
		var visible bool
		var created time.Time
		var resolved sql.NullTime
		if err = rows.Scan(&issueID, &kind, &severity, &description, &status, &visible, &resolution, &created, &resolved); err != nil {
			rows.Close()
			return out, err
		}
		out.Issues = append(out.Issues, map[string]any{"id": issueID, "issue_type": kind, "severity": severity, "description": description, "status": status, "customer_visible": visible, "resolution_note": resolution, "created_at": created, "resolved_at": scanNullableTime(resolved)})
	}
	if err = rows.Close(); err != nil {
		return out, err
	}
	if !customer {
		rows, err = s.db.QueryContext(ctx, `SELECT id,material_name,quantity::text,unit,cost_entry_id,COALESCE(notes,''),created_at FROM installation_material_usage WHERE installation_job_id=$1 ORDER BY created_at DESC`, id)
		if err != nil {
			return out, err
		}
		out.Materials = []map[string]any{}
		for rows.Next() {
			var materialID, name, quantity, unit, notes string
			var cost sql.NullString
			var created time.Time
			if err = rows.Scan(&materialID, &name, &quantity, &unit, &cost, &notes, &created); err != nil {
				rows.Close()
				return out, err
			}
			out.Materials = append(out.Materials, map[string]any{"id": materialID, "material_name": name, "quantity": quantity, "unit": unit, "cost_entry_id": scanNullableString(cost), "notes": notes, "created_at": created})
		}
		if err = rows.Close(); err != nil {
			return out, err
		}
	}
	fileVisibility := ""
	if customer {
		fileVisibility = " AND f.customer_visible"
	}
	rows, err = s.db.QueryContext(ctx, `SELECT f.id,f.entity_type,f.entity_id,f.original_file_name,f.mime_type,f.customer_visible,f.created_at FROM workflow_files f WHERE ((f.entity_type='INSTALLATION' AND f.entity_id=$1) OR (f.entity_type='INSTALLATION_UPDATE' AND EXISTS(SELECT 1 FROM installation_updates u WHERE u.id=f.entity_id AND u.installation_job_id=$1)))`+fileVisibility+` ORDER BY f.created_at DESC`, id)
	if err != nil {
		return out, err
	}
	out.Files = []map[string]any{}
	for rows.Next() {
		var fileID, entityType, entityID, name, mime string
		var visible bool
		var created time.Time
		if err = rows.Scan(&fileID, &entityType, &entityID, &name, &mime, &visible, &created); err != nil {
			rows.Close()
			return out, err
		}
		out.Files = append(out.Files, map[string]any{"id": fileID, "entity_type": entityType, "entity_id": entityID, "original_file_name": name, "mime_type": mime, "customer_visible": visible, "created_at": created})
	}
	if err = rows.Close(); err != nil {
		return out, err
	}
	rows, err = s.db.QueryContext(ctx, `SELECT id,customer_name,accepted,COALESCE(comment,''),accepted_at FROM customer_order_acceptances WHERE installation_job_id=$1 ORDER BY accepted_at DESC`, id)
	if err != nil {
		return out, err
	}
	out.Acceptances = []map[string]any{}
	for rows.Next() {
		var acceptanceID, name, comment string
		var accepted bool
		var acceptedAt time.Time
		if err = rows.Scan(&acceptanceID, &name, &accepted, &comment, &acceptedAt); err != nil {
			rows.Close()
			return out, err
		}
		out.Acceptances = append(out.Acceptances, map[string]any{"id": acceptanceID, "customer_name": name, "accepted": accepted, "comment": comment, "accepted_at": acceptedAt})
	}
	if err = rows.Close(); err != nil {
		return out, err
	}
	return out, nil
}

func validateInstallationPayload(p InstallationPayload) (InstallationPayload, error) {
	p.AreaUnit = normalizeCode(p.AreaUnit)
	if p.AreaUnit == "" {
		p.AreaUnit = "SQUARE_METER"
	}
	if p.AreaUnit != "SQUARE_METER" && p.AreaUnit != "PIECE" && p.AreaUnit != "SLAB" && p.AreaUnit != "TILE" {
		return p, ErrValidation
	}
	if p.PlannedArea != nil && !validNonNegativeDecimal(*p.PlannedArea) {
		return p, ErrValidation
	}
	if p.ContactPhone != "" {
		p.ContactPhone = NormalizePhone(p.ContactPhone)
		if p.ContactPhone == "" {
			return p, ErrValidation
		}
	}
	return p, nil
}

func (s *OperationsService) CreateInstallation(ctx context.Context, actor, orderID, key string, p InstallationPayload) (InstallationJob, error) {
	var out InstallationJob
	p, err := validateInstallationPayload(p)
	if err != nil {
		return out, err
	}
	if p.StartWorkflow && !s.HasPermission(ctx, actor, "workflow_start.installation") {
		return out, ErrForbidden
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return out, err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, "INSTALLATION_CREATE", key, map[string]any{"order_id": orderID, "payload": p})
	if err != nil {
		return out, err
	}
	if claim.Existing {
		if err = json.Unmarshal(claim.Response, &out); err != nil {
			return out, err
		}
		return out, tx.Commit()
	}
	var customerID, orderNumber, orderStatus string
	if err = tx.QueryRowContext(ctx, `SELECT customer_user_id,order_number,status FROM orders WHERE id=$1 FOR UPDATE`, orderID).Scan(&customerID, &orderNumber, &orderStatus); err != nil {
		return out, err
	}
	if orderStatus == "CANCELLED" || orderStatus == "CLOSED" {
		return out, conflict("INVALID_INSTALLATION_TRANSITION", "closed or cancelled order cannot receive a new installation")
	}
	if p.SupplierID != nil {
		var active bool
		if err = tx.QueryRowContext(ctx, `SELECT is_active FROM suppliers WHERE id=$1 FOR SHARE`, *p.SupplierID).Scan(&active); err != nil {
			return out, err
		}
		if !active {
			return out, conflict("INACTIVE_SUPPLIER", "supplier is inactive")
		}
	}
	number, err := nextReadableNumberTx(ctx, tx, "INS")
	if err != nil {
		return out, err
	}
	status := "DRAFT"
	if p.PlannedStartAt != nil {
		status = "PLANNED"
	}
	var id string
	err = tx.QueryRowContext(ctx, `INSERT INTO installation_jobs(installation_number,order_id,customer_user_id,project_name,project_address,contact_name,contact_phone,status,installation_lead_user_id,supplier_id,planned_start_at,estimated_end_at,planned_area,area_unit,notes,created_by_user_id) VALUES($1,$2,$3,NULLIF($4,''),NULLIF($5,''),NULLIF($6,''),NULLIF($7,''),$8,$9,$10,$11,$12,NULLIF($13,'')::numeric,$14,NULLIF($15,''),$16) RETURNING id`, number, orderID, customerID, p.ProjectName, p.ProjectAddress, p.ContactName, p.ContactPhone, status, p.InstallationLeadUserID, p.SupplierID, p.PlannedStartAt, p.EstimatedEndAt, valueOrEmpty(p.PlannedArea), p.AreaUnit, p.Notes, actor).Scan(&id)
	if err != nil {
		return out, err
	}
	var workflowID *string
	if p.StartWorkflow {
		wid, workflowErr := s.startScopedWorkflowTx(ctx, tx, actor, orderID, "INSTALLATION", id, p.WorkflowTemplateID, p.ParentWorkflowID, p.ParentStepID)
		if workflowErr != nil {
			return out, workflowErr
		}
		workflowID = &wid
		if _, err = tx.ExecContext(ctx, `UPDATE installation_jobs SET workflow_instance_id=$2 WHERE id=$1`, id, wid); err != nil {
			return out, err
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE orders SET installation_required=TRUE,updated_at=NOW() WHERE id=$1`, orderID); err != nil {
		return out, err
	}
	out = InstallationJob{ID: id, InstallationNumber: number, OrderID: orderID, OrderNumber: orderNumber, CustomerUserID: customerID, ProjectName: p.ProjectName, ProjectAddress: p.ProjectAddress, ContactName: p.ContactName, ContactPhone: p.ContactPhone, Status: status, InstallationLeadUserID: p.InstallationLeadUserID, SupplierID: p.SupplierID, PlannedStartAt: p.PlannedStartAt, EstimatedEndAt: p.EstimatedEndAt, PlannedArea: p.PlannedArea, InstalledArea: "0.0000", AreaUnit: p.AreaUnit, Notes: p.Notes, WorkflowInstanceID: workflowID, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	s.auditTx(ctx, tx, actor, "INSTALLATION_CREATED", "installation_job", id, nil, p)
	if err = finishOperationTx(ctx, tx, actor, "INSTALLATION_CREATE", key, out); err != nil {
		return out, err
	}
	return out, tx.Commit()
}

func (s *OperationsService) UpdateInstallation(ctx context.Context, actor, id string, p InstallationPayload) error {
	p, err := validateInstallationPayload(p)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var status string
	if err = tx.QueryRowContext(ctx, `SELECT status FROM installation_jobs WHERE id=$1 FOR UPDATE`, id).Scan(&status); err != nil {
		return err
	}
	if status == "COMPLETED" || status == "CANCELLED" {
		return conflict("INVALID_INSTALLATION_TRANSITION", "terminal installation cannot be edited")
	}
	if p.SupplierID != nil {
		var active bool
		if err = tx.QueryRowContext(ctx, `SELECT is_active FROM suppliers WHERE id=$1 FOR SHARE`, *p.SupplierID).Scan(&active); err != nil {
			return err
		}
		if !active {
			return conflict("INACTIVE_SUPPLIER", "supplier is inactive")
		}
	}
	newStatus := status
	if status == "DRAFT" && p.PlannedStartAt != nil {
		newStatus = "PLANNED"
	}
	_, err = tx.ExecContext(ctx, `UPDATE installation_jobs SET project_name=NULLIF($2,''),project_address=NULLIF($3,''),contact_name=NULLIF($4,''),contact_phone=NULLIF($5,''),installation_lead_user_id=$6,supplier_id=$7,planned_start_at=$8,estimated_end_at=$9,planned_area=NULLIF($10,'')::numeric,area_unit=$11,notes=NULLIF($12,''),status=$13,updated_at=NOW() WHERE id=$1`, id, p.ProjectName, p.ProjectAddress, p.ContactName, p.ContactPhone, p.InstallationLeadUserID, p.SupplierID, p.PlannedStartAt, p.EstimatedEndAt, valueOrEmpty(p.PlannedArea), p.AreaUnit, p.Notes, newStatus)
	if err != nil {
		return err
	}
	s.auditTx(ctx, tx, actor, "INSTALLATION_UPDATED", "installation_job", id, map[string]any{"status": status}, p)
	return tx.Commit()
}

func (s *OperationsService) ReplaceInstallationMembers(ctx context.Context, actor, id string, members []InstallationMemberPayload) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var status string
	if err = tx.QueryRowContext(ctx, `SELECT status FROM installation_jobs WHERE id=$1 FOR UPDATE`, id).Scan(&status); err != nil {
		return err
	}
	if status == "COMPLETED" || status == "CANCELLED" {
		return conflict("INVALID_INSTALLATION_TRANSITION", "terminal installation team is immutable")
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM installation_job_members WHERE installation_job_id=$1`, id); err != nil {
		return err
	}
	for _, member := range members {
		if member.UserID == nil && strings.TrimSpace(member.NameOverride) == "" {
			return ErrValidation
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO installation_job_members(installation_job_id,user_id,name_override,role_label) VALUES($1,$2,NULLIF($3,''),NULLIF($4,''))`, id, member.UserID, member.NameOverride, member.RoleLabel); err != nil {
			return err
		}
	}
	s.auditTx(ctx, tx, actor, "INSTALLATION_UPDATED", "installation_job", id, nil, map[string]any{"members": members})
	return tx.Commit()
}

func (s *OperationsService) installationWarningsTx(ctx context.Context, tx *sql.Tx, id string) ([]string, error) {
	warnings := []string{}
	var openIssues int
	var planned sql.NullString
	var installed string
	if err := tx.QueryRowContext(ctx, `SELECT planned_area::text,installed_area::text,(SELECT COUNT(*) FROM installation_issues WHERE installation_job_id=j.id AND status='OPEN') FROM installation_jobs j WHERE id=$1`, id).Scan(&planned, &installed, &openIssues); err != nil {
		return nil, err
	}
	if openIssues > 0 {
		warnings = append(warnings, fmt.Sprintf("%d installation issues are still open", openIssues))
	}
	if planned.Valid {
		plannedRat, _ := new(big.Rat).SetString(planned.String)
		installedRat, _ := new(big.Rat).SetString(installed)
		if plannedRat != nil && installedRat != nil && installedRat.Cmp(plannedRat) < 0 {
			warnings = append(warnings, "installed quantity is below planned quantity")
		}
	}
	return warnings, nil
}

func (s *OperationsService) transitionInstallation(ctx context.Context, actor, id, key, operation, target string, p InstallationFlowPayload) (map[string]any, error) {
	if (target == "CANCELLED" || p.Force) && requireReason(p.Reason) != nil {
		return nil, ErrValidation
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, operation, key, map[string]any{"id": id, "payload": p})
	if err != nil {
		return nil, err
	}
	if claim.Existing {
		var out map[string]any
		if err = json.Unmarshal(claim.Response, &out); err != nil {
			return nil, err
		}
		return out, tx.Commit()
	}
	var current string
	if err = tx.QueryRowContext(ctx, `SELECT status FROM installation_jobs WHERE id=$1 FOR UPDATE`, id).Scan(&current); err != nil {
		return nil, err
	}
	valid := target == "IN_PROGRESS" && (current == "DRAFT" || current == "PLANNED" || current == "READY" || current == "PAUSED") || target == "PAUSED" && current == "IN_PROGRESS" || target == "COMPLETED" && (current == "IN_PROGRESS" || current == "PAUSED") || target == "CANCELLED" && current != "COMPLETED" && current != "CANCELLED"
	if !valid {
		return nil, conflict("INVALID_INSTALLATION_TRANSITION", "installation transition is not allowed")
	}
	warnings := []string{}
	if target == "COMPLETED" {
		warnings, err = s.installationWarningsTx(ctx, tx, id)
		if err != nil {
			return nil, err
		}
		if len(warnings) > 0 && !p.Force {
			return nil, conflict("INVALID_INSTALLATION_TRANSITION", strings.Join(warnings, "; "))
		}
		if p.Force && !s.HasPermission(ctx, actor, "installation.override") {
			return nil, ErrForbidden
		}
	}
	switch target {
	case "IN_PROGRESS":
		_, err = tx.ExecContext(ctx, `UPDATE installation_jobs SET status='IN_PROGRESS',actual_start_at=COALESCE(actual_start_at,NOW()),updated_at=NOW() WHERE id=$1`, id)
	case "PAUSED":
		_, err = tx.ExecContext(ctx, `UPDATE installation_jobs SET status='PAUSED',updated_at=NOW() WHERE id=$1`, id)
	case "COMPLETED":
		_, err = tx.ExecContext(ctx, `UPDATE installation_jobs SET status='COMPLETED',actual_end_at=NOW(),updated_at=NOW() WHERE id=$1`, id)
	case "CANCELLED":
		_, err = tx.ExecContext(ctx, `UPDATE installation_jobs SET status='CANCELLED',cancelled_at=NOW(),cancelled_by_user_id=$2,cancellation_reason=$3,updated_at=NOW() WHERE id=$1`, id, actor, p.Reason)
	}
	if err != nil {
		return nil, err
	}
	event := ""
	if target == "IN_PROGRESS" && current != "PAUSED" {
		event = "INSTALLATION_STARTED"
	} else if target == "COMPLETED" {
		event = "INSTALLATION_COMPLETED"
	}
	if event != "" {
		if err = s.markDomainOperationTx(ctx, tx, actor, p.WorkflowStepInstanceID, event, "INSTALLATION", id, randomUUIDText()); err != nil {
			return nil, err
		}
	}
	action := map[string]string{"IN_PROGRESS": "INSTALLATION_STARTED", "PAUSED": "INSTALLATION_UPDATED", "COMPLETED": "INSTALLATION_COMPLETED", "CANCELLED": "INSTALLATION_CANCELLED"}[target]
	s.auditTx(ctx, tx, actor, action, "installation_job", id, map[string]any{"status": current}, map[string]any{"status": target, "reason": p.Reason, "warnings": warnings})
	out := map[string]any{"id": id, "status": target, "warnings": warnings}
	if err = finishOperationTx(ctx, tx, actor, operation, key, out); err != nil {
		return nil, err
	}
	return out, tx.Commit()
}

func (s *OperationsService) StartInstallation(ctx context.Context, actor, id, key string, p InstallationFlowPayload) (map[string]any, error) {
	return s.transitionInstallation(ctx, actor, id, key, "INSTALLATION_START", "IN_PROGRESS", p)
}
func (s *OperationsService) PauseInstallation(ctx context.Context, actor, id, key string, p InstallationFlowPayload) (map[string]any, error) {
	return s.transitionInstallation(ctx, actor, id, key, "INSTALLATION_PAUSE", "PAUSED", p)
}
func (s *OperationsService) CompleteInstallation(ctx context.Context, actor, id, key string, p InstallationFlowPayload) (map[string]any, error) {
	return s.transitionInstallation(ctx, actor, id, key, "INSTALLATION_COMPLETE", "COMPLETED", p)
}
func (s *OperationsService) CancelInstallation(ctx context.Context, actor, id, key string, p InstallationFlowPayload) (map[string]any, error) {
	return s.transitionInstallation(ctx, actor, id, key, "INSTALLATION_CANCEL", "CANCELLED", p)
}

var installationActivities = map[string]bool{"SUBSTRATE_PREPARATION": true, "CUTTING": true, "INSTALLATION": true, "RESIN": true, "POLISHING": true, "GROUTING": true, "ANCHORING": true, "BASE_INSTALLATION": true, "REPAIR": true, "CLEANUP": true, "OTHER": true}

func (s *OperationsService) AddInstallationUpdate(ctx context.Context, actor, id, key string, p InstallationUpdatePayload) (map[string]any, error) {
	p.QuantityUnit = normalizeCode(p.QuantityUnit)
	p.Status = normalizeCode(p.Status)
	if p.Status == "" {
		p.Status = "PROGRESS"
	}
	if !validNonNegativeDecimal(p.InstalledQuantity) || (p.Status != "PROGRESS" && p.Status != "DELAY" && p.Status != "PAUSED" && p.Status != "RESUMED" && p.Status != "NOTE") {
		return nil, ErrValidation
	}
	for i, activity := range p.Activities {
		p.Activities[i] = normalizeCode(activity)
		if !installationActivities[p.Activities[i]] {
			return nil, ErrValidation
		}
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, "INSTALLATION_UPDATE", key, map[string]any{"id": id, "payload": p})
	if err != nil {
		return nil, err
	}
	if claim.Existing {
		var out map[string]any
		if err = json.Unmarshal(claim.Response, &out); err != nil {
			return nil, err
		}
		return out, tx.Commit()
	}
	var status, unit, installed string
	if err = tx.QueryRowContext(ctx, `SELECT status,area_unit,installed_area::text FROM installation_jobs WHERE id=$1 FOR UPDATE`, id).Scan(&status, &unit, &installed); err != nil {
		return nil, err
	}
	if status != "IN_PROGRESS" && status != "PAUSED" {
		return nil, conflict("INVALID_INSTALLATION_TRANSITION", "installation updates require an active or paused job")
	}
	if p.QuantityUnit == "" {
		p.QuantityUnit = unit
	}
	if p.QuantityUnit != unit {
		return nil, conflict("INCOMPATIBLE_UNIT", "installation update unit differs from job unit")
	}
	date := time.Now()
	if p.UpdateDate != nil {
		date = *p.UpdateDate
	}
	var updateID string
	if err = tx.QueryRowContext(ctx, `INSERT INTO installation_updates(installation_job_id,update_date,installed_quantity,quantity_unit,status,description,customer_visible,created_by_user_id) VALUES($1,$2,$3::numeric,$4,$5,NULLIF($6,''),$7,$8) RETURNING id`, id, date, p.InstalledQuantity, p.QuantityUnit, p.Status, p.Description, p.CustomerVisible, actor).Scan(&updateID); err != nil {
		return nil, err
	}
	for _, activity := range p.Activities {
		if _, err = tx.ExecContext(ctx, `INSERT INTO installation_update_activities(installation_update_id,activity_type) VALUES($1,$2) ON CONFLICT DO NOTHING`, updateID, activity); err != nil {
			return nil, err
		}
	}
	newInstalled := addDecimal(installed, p.InstalledQuantity)
	if _, err = tx.ExecContext(ctx, `UPDATE installation_jobs SET installed_area=$2::numeric,status=CASE WHEN $3='PAUSED' THEN 'PAUSED' WHEN $3='RESUMED' THEN 'IN_PROGRESS' ELSE status END,updated_at=NOW() WHERE id=$1`, id, newInstalled, p.Status); err != nil {
		return nil, err
	}
	s.auditTx(ctx, tx, actor, "INSTALLATION_UPDATED", "installation_job", id, map[string]any{"installed_area": installed}, map[string]any{"installed_area": newInstalled, "update_id": updateID})
	out := map[string]any{"id": updateID, "installation_job_id": id, "installed_quantity": p.InstalledQuantity, "installed_area": newInstalled, "quantity_unit": unit}
	if err = finishOperationTx(ctx, tx, actor, "INSTALLATION_UPDATE", key, out); err != nil {
		return nil, err
	}
	return out, tx.Commit()
}

func (s *OperationsService) AddInstallationIssue(ctx context.Context, actor, id string, p InstallationIssuePayload) (map[string]any, error) {
	p.IssueType, p.Severity = normalizeCode(p.IssueType), normalizeCode(p.Severity)
	if p.Severity == "" {
		p.Severity = "WARNING"
	}
	validTypes := map[string]bool{"STONE_DAMAGE": true, "WRONG_DIMENSION": true, "MISSING_MATERIAL": true, "SITE_NOT_READY": true, "CUSTOMER_CHANGE": true, "SUBSTRATE_PROBLEM": true, "OTHER": true}
	if !validTypes[p.IssueType] || (p.Severity != "INFO" && p.Severity != "WARNING" && p.Severity != "CRITICAL") || strings.TrimSpace(p.Description) == "" {
		return nil, ErrValidation
	}
	var issueID string
	err := s.db.QueryRowContext(ctx, `INSERT INTO installation_issues(installation_job_id,issue_type,severity,description,customer_visible,reported_by_user_id) SELECT $1,$2,$3,$4,$5,$6 WHERE EXISTS(SELECT 1 FROM installation_jobs WHERE id=$1 AND status NOT IN ('COMPLETED','CANCELLED')) RETURNING id`, id, p.IssueType, p.Severity, p.Description, p.CustomerVisible, actor).Scan(&issueID)
	if err == nil {
		s.audit(ctx, actor, "INSTALLATION_ISSUE_CREATED", "installation_issue", issueID, p)
	}
	return map[string]any{"id": issueID, "status": "OPEN"}, err
}

func (s *OperationsService) ResolveInstallationIssue(ctx context.Context, actor, issueID string, p InstallationIssuePayload) error {
	if requireReason(p.ResolutionNote) != nil {
		return ErrValidation
	}
	result, err := s.db.ExecContext(ctx, `UPDATE installation_issues SET status='RESOLVED',resolved_by_user_id=$2,resolved_at=NOW(),resolution_note=$3 WHERE id=$1 AND status='OPEN'`, issueID, actor, p.ResolutionNote)
	if err != nil {
		return err
	}
	if count, _ := result.RowsAffected(); count == 0 {
		return conflict("INVALID_INSTALLATION_TRANSITION", "installation issue is not open")
	}
	s.audit(ctx, actor, "INSTALLATION_UPDATED", "installation_issue", issueID, p)
	return nil
}

func (s *OperationsService) AddInstallationMaterial(ctx context.Context, actor, id string, p InstallationMaterialPayload) (map[string]any, error) {
	if strings.TrimSpace(p.MaterialName) == "" || strings.TrimSpace(p.Unit) == "" || !validPositiveDecimal(p.Quantity) {
		return nil, ErrValidation
	}
	var materialID string
	err := s.db.QueryRowContext(ctx, `INSERT INTO installation_material_usage(installation_job_id,material_name,quantity,unit,cost_entry_id,notes,created_by_user_id) SELECT $1,$2,$3::numeric,$4,$5,NULLIF($6,''),$7 WHERE EXISTS(SELECT 1 FROM installation_jobs WHERE id=$1 AND status NOT IN ('COMPLETED','CANCELLED')) RETURNING id`, id, p.MaterialName, p.Quantity, p.Unit, p.CostEntryID, p.Notes, actor).Scan(&materialID)
	if err == nil {
		s.audit(ctx, actor, "INSTALLATION_UPDATED", "installation_material_usage", materialID, p)
	}
	return map[string]any{"id": materialID}, err
}

func (s *OperationsService) AccountInstallation(ctx context.Context, actor, orderID string) (*InstallationJob, error) {
	var id string
	err := s.db.QueryRowContext(ctx, `SELECT id FROM installation_jobs WHERE order_id=$1 AND customer_user_id=$2 AND status<>'CANCELLED' ORDER BY created_at DESC LIMIT 1`, orderID, actor).Scan(&id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	out, err := s.GetInstallation(ctx, actor, id, true)
	return &out, err
}

func (s *OperationsService) RecordCustomerAcceptance(ctx context.Context, actor, orderID, key string, customer bool, p CustomerAcceptancePayload) (map[string]any, error) {
	if strings.TrimSpace(p.CustomerName) == "" {
		return nil, ErrValidation
	}
	if customer && !s.HasPermission(ctx, actor, "customer_portal.acceptance.confirm_own") {
		return nil, ErrForbidden
	}
	if !customer && !s.HasPermission(ctx, actor, "customer_acceptance.record") {
		return nil, ErrForbidden
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, "CUSTOMER_ACCEPTANCE", key, map[string]any{"order_id": orderID, "payload": p, "customer": customer})
	if err != nil {
		return nil, err
	}
	if claim.Existing {
		var out map[string]any
		if err = json.Unmarshal(claim.Response, &out); err != nil {
			return nil, err
		}
		return out, tx.Commit()
	}
	var owner string
	if err = tx.QueryRowContext(ctx, `SELECT customer_user_id FROM orders WHERE id=$1 FOR SHARE`, orderID).Scan(&owner); err != nil {
		return nil, err
	}
	if customer && owner != actor {
		return nil, ErrForbidden
	}
	if p.InstallationJobID != nil {
		var belongs, completed bool
		if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM installation_jobs WHERE id=$1 AND order_id=$2),EXISTS(SELECT 1 FROM installation_jobs WHERE id=$1 AND order_id=$2 AND status='COMPLETED')`, *p.InstallationJobID, orderID).Scan(&belongs, &completed); err != nil || !belongs {
			if err != nil {
				return nil, err
			}
			return nil, conflict("SCOPE_MISMATCH", "installation belongs to another order")
		}
		if !completed {
			return nil, conflict("INVALID_INSTALLATION_TRANSITION", "installation must be completed before acceptance")
		}
	}
	if p.ShipmentID != nil {
		var belongs, delivered bool
		if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM shipments WHERE id=$1 AND order_id=$2),EXISTS(SELECT 1 FROM shipments WHERE id=$1 AND order_id=$2 AND status='DELIVERED')`, *p.ShipmentID, orderID).Scan(&belongs, &delivered); err != nil || !belongs {
			if err != nil {
				return nil, err
			}
			return nil, conflict("SCOPE_MISMATCH", "shipment belongs to another order")
		}
		if !delivered {
			return nil, conflict("INVALID_TRANSITION", "shipment must be delivered before acceptance")
		}
	}
	if p.InstallationJobID == nil && p.ShipmentID == nil {
		var ready bool
		if err = tx.QueryRowContext(ctx, `SELECT status IN ('COMPLETED','CLOSED') OR EXISTS(SELECT 1 FROM shipments WHERE order_id=$1 AND status='DELIVERED') FROM orders WHERE id=$1`, orderID).Scan(&ready); err != nil {
			return nil, err
		}
		if !ready {
			return nil, conflict("INVALID_TRANSITION", "order delivery must be completed before acceptance")
		}
	}
	for _, fileID := range []*string{p.SignatureFileID, p.PhotoFileID} {
		if fileID == nil {
			continue
		}
		var fileOrder string
		var visible bool
		err = tx.QueryRowContext(ctx, `SELECT COALESCE(wi.order_id,j.order_id,sh.order_id,direct_order.id),f.customer_visible FROM workflow_files f LEFT JOIN workflow_instances wi ON wi.id=f.workflow_instance_id LEFT JOIN installation_jobs j ON f.entity_type IN ('INSTALLATION','INSTALLATION_UPDATE') AND j.id=CASE WHEN f.entity_type='INSTALLATION' THEN f.entity_id ELSE (SELECT installation_job_id FROM installation_updates WHERE id=f.entity_id) END LEFT JOIN shipments sh ON f.entity_type IN ('SHIPMENT','DELIVERY') AND sh.id=f.entity_id LEFT JOIN orders direct_order ON f.entity_type='ORDER' AND direct_order.id=f.entity_id WHERE f.id=$1`, *fileID).Scan(&fileOrder, &visible)
		if err != nil || fileOrder != orderID || customer && !visible {
			if err != nil {
				return nil, err
			}
			return nil, ErrForbidden
		}
	}
	var acceptanceID string
	err = tx.QueryRowContext(ctx, `INSERT INTO customer_order_acceptances(order_id,installation_job_id,shipment_id,workflow_step_instance_id,customer_name,accepted,comment,signature_file_id,photo_file_id,recorded_by_user_id) VALUES($1,$2,$3,$4,$5,$6,NULLIF($7,''),$8,$9,$10) RETURNING id`, orderID, p.InstallationJobID, p.ShipmentID, p.WorkflowStepInstanceID, p.CustomerName, p.Accepted, p.Comment, p.SignatureFileID, p.PhotoFileID, actor).Scan(&acceptanceID)
	if err != nil {
		return nil, err
	}
	if p.WorkflowStepInstanceID != nil && p.Accepted {
		var scopeType, scopeID string
		if err = tx.QueryRowContext(ctx, `SELECT wi.scope_type,wi.scope_id FROM workflow_step_instances si JOIN workflow_instances wi ON wi.id=si.workflow_instance_id WHERE si.id=$1`, *p.WorkflowStepInstanceID).Scan(&scopeType, &scopeID); err != nil {
			return nil, err
		}
		if scopeType == "INSTALLATION" && p.InstallationJobID != nil && scopeID == *p.InstallationJobID {
			if err = s.markDomainOperationTx(ctx, tx, actor, p.WorkflowStepInstanceID, "CUSTOMER_ACCEPTED", "INSTALLATION", scopeID, randomUUIDText()); err != nil {
				return nil, err
			}
		}
	}
	s.auditTx(ctx, tx, actor, "CUSTOMER_ACCEPTANCE_RECORDED", "customer_order_acceptance", acceptanceID, nil, map[string]any{"order_id": orderID, "accepted": p.Accepted, "operational_signature": p.SignatureFileID != nil})
	out := map[string]any{"id": acceptanceID, "order_id": orderID, "accepted": p.Accepted, "accepted_at": time.Now()}
	if err = finishOperationTx(ctx, tx, actor, "CUSTOMER_ACCEPTANCE", key, out); err != nil {
		return nil, err
	}
	return out, tx.Commit()
}
