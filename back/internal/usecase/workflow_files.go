package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

type WorkflowFile struct {
	ID                     string  `json:"id"`
	WorkflowInstanceID     string  `json:"workflow_instance_id"`
	WorkflowStepInstanceID *string `json:"workflow_step_instance_id,omitempty"`
	FieldDefinitionID      *int64  `json:"field_definition_id,omitempty"`
	EntityType             string  `json:"entity_type"`
	EntityID               string  `json:"entity_id"`
	StorageKey             string  `json:"-"`
	OriginalFileName       string  `json:"original_file_name"`
	MIMEType               string  `json:"mime_type"`
	SizeBytes              int64   `json:"size_bytes"`
	CustomerVisible        bool    `json:"customer_visible"`
}

type WorkflowUploadPolicy struct {
	AllowedMIMETypes []string
	MaxSizeBytes     int64
}

func (s *OperationsService) PrepareWorkflowFileUpload(ctx context.Context, actor, stepID string, fieldID int64) (string, bool, WorkflowUploadPolicy, error) {
	var workflowID, fieldType, permission string
	var validation []byte
	var roleID sql.NullInt64
	var assigned sql.NullString
	var customerVisible bool
	err := s.db.QueryRowContext(ctx, `SELECT si.workflow_instance_id,f.field_type,f.is_customer_visible,COALESCE(si.required_permission_code,''),si.responsible_role_id,si.assigned_user_id,f.validation_json FROM workflow_step_instances si JOIN workflow_instance_field_definitions f ON f.workflow_step_instance_id=si.id WHERE si.id=$1 AND f.id=$2`, stepID, fieldID).Scan(&workflowID, &fieldType, &customerVisible, &permission, &roleID, &assigned, &validation)
	if err != nil {
		return "", false, WorkflowUploadPolicy{}, err
	}
	if fieldType != "FILE" && fieldType != "IMAGE" && fieldType != "SIGNATURE" && fieldType != "QC_CHECK" {
		return "", false, WorkflowUploadPolicy{}, ErrValidation
	}
	if !s.canOperateStep(ctx, actor, permission, roleID, assigned) {
		return "", false, WorkflowUploadPolicy{}, ErrForbidden
	}
	policy := WorkflowUploadPolicy{MaxSizeBytes: s.MaximumUploadBytes(ctx,15 << 20)}
	if fieldType == "FILE" {
		policy.AllowedMIMETypes = []string{"image/png", "image/jpeg", "application/pdf"}
	} else if fieldType == "IMAGE" || fieldType == "QC_CHECK" {
		policy.AllowedMIMETypes = []string{"image/png", "image/jpeg"}
	} else {
		policy.AllowedMIMETypes = []string{"image/png"}
	}
	var rules struct {
		AllowedMIMETypes      []string `json:"allowedMimeTypes"`
		AllowedMIMETypesSnake []string `json:"allowed_mime_types"`
		MaxSizeBytes          int64    `json:"maxSizeBytes"`
		MaxSizeBytesSnake     int64    `json:"max_size_bytes"`
	}
	if json.Unmarshal(validation, &rules) == nil {
		if len(rules.AllowedMIMETypes) > 0 {
			policy.AllowedMIMETypes = rules.AllowedMIMETypes
		} else if len(rules.AllowedMIMETypesSnake) > 0 {
			policy.AllowedMIMETypes = rules.AllowedMIMETypesSnake
		}
		if rules.MaxSizeBytes > 0 && rules.MaxSizeBytes < policy.MaxSizeBytes {
			policy.MaxSizeBytes = rules.MaxSizeBytes
		} else if rules.MaxSizeBytesSnake > 0 && rules.MaxSizeBytesSnake < policy.MaxSizeBytes {
			policy.MaxSizeBytes = rules.MaxSizeBytesSnake
		}
	}
	return workflowID, customerVisible, policy, nil
}

func (s *OperationsService) RegisterWorkflowFile(ctx context.Context, actor, workflowID, stepID string, fieldID int64, storageKey, originalName, mimeType string, size int64, customerVisible bool) (WorkflowFile, error) {
	var file WorkflowFile
	err := s.db.QueryRowContext(ctx, `INSERT INTO workflow_files(workflow_instance_id,workflow_step_instance_id,field_definition_id,entity_type,entity_id,storage_key,original_file_name,mime_type,size_bytes,customer_visible,uploaded_by_user_id) VALUES($1,$2,$3,'WORKFLOW',$1,$4,$5,$6,$7,$8,$9) RETURNING id,workflow_instance_id,workflow_step_instance_id,field_definition_id,entity_type,entity_id,storage_key,original_file_name,mime_type,size_bytes,customer_visible`, workflowID, stepID, fieldID, storageKey, originalName, mimeType, size, customerVisible, actor).Scan(&file.ID, &file.WorkflowInstanceID, &file.WorkflowStepInstanceID, &file.FieldDefinitionID, &file.EntityType, &file.EntityID, &file.StorageKey, &file.OriginalFileName, &file.MIMEType, &file.SizeBytes, &file.CustomerVisible)
	if err == nil {
		s.audit(ctx, actor, "workflow_files.upload", "workflow_file", file.ID, map[string]any{"workflow_instance_id": workflowID, "step_instance_id": stepID})
	}
	return file, err
}

func (s *OperationsService) PrepareOperationalFileUpload(ctx context.Context, actor, entityType, entityID string, customerVisible bool) (string, bool, WorkflowUploadPolicy, error) {
	entityType = normalizeCode(entityType)
	var workflow sql.NullString
	var workflowID, owner string
	var allowed bool
	switch entityType {
	case "SHIPMENT", "DELIVERY":
		err := s.db.QueryRowContext(ctx, `SELECT sh.workflow_instance_id,o.customer_user_id FROM shipments sh JOIN orders o ON o.id=sh.order_id WHERE sh.id=$1`, entityID).Scan(&workflow, &owner)
		if err != nil {
			return "", false, WorkflowUploadPolicy{}, err
		}
		if owner == actor {
			allowed = s.HasPermission(ctx, actor, "customer_portal.shipments.confirm_delivery")
			customerVisible = true
		} else {
			allowed = s.canViewShipment(ctx, actor, entityID, false) && (s.HasPermission(ctx, actor, "shipments.update") || s.HasPermission(ctx, actor, "shipments.confirm_delivery"))
		}
	case "PACKAGE":
		err := s.db.QueryRowContext(ctx, `SELECT b.workflow_instance_id,o.customer_user_id FROM packaging_units p JOIN fulfillment_batches b ON b.id=p.batch_id JOIN orders o ON o.id=b.order_id WHERE p.id=$1`, entityID).Scan(&workflow, &owner)
		if err != nil {
			return "", false, WorkflowUploadPolicy{}, err
		}
		allowed = s.HasPermission(ctx, actor, "packaging.update")
	case "QUALITY_INSPECTION":
		err := s.db.QueryRowContext(ctx, `SELECT wi.id,o.customer_user_id FROM quality_inspections q JOIN orders o ON o.id=q.order_id LEFT JOIN workflow_step_instances si ON si.id=q.workflow_step_instance_id LEFT JOIN workflow_instances wi ON wi.id=si.workflow_instance_id WHERE q.id=$1`, entityID).Scan(&workflow, &owner)
		if err != nil {
			return "", false, WorkflowUploadPolicy{}, err
		}
		allowed = s.HasPermission(ctx, actor, "quality.inspect") || s.HasPermission(ctx, actor, "quality.override")
		customerVisible = false
	case "INSTALLATION":
		err := s.db.QueryRowContext(ctx, `SELECT j.workflow_instance_id,j.customer_user_id FROM installation_jobs j WHERE j.id=$1`, entityID).Scan(&workflow, &owner)
		if err != nil {
			return "", false, WorkflowUploadPolicy{}, err
		}
		canAccess, accessErr := s.canAccessInstallation(ctx, actor, entityID, false)
		if accessErr != nil {
			return "", false, WorkflowUploadPolicy{}, accessErr
		}
		allowed = canAccess && (s.HasPermission(ctx, actor, "installation.progress") || s.HasPermission(ctx, actor, "installation.update"))
	case "INSTALLATION_UPDATE":
		var installationID string
		err := s.db.QueryRowContext(ctx, `SELECT j.id,j.workflow_instance_id,j.customer_user_id FROM installation_updates u JOIN installation_jobs j ON j.id=u.installation_job_id WHERE u.id=$1`, entityID).Scan(&installationID, &workflow, &owner)
		if err != nil {
			return "", false, WorkflowUploadPolicy{}, err
		}
		canAccess, accessErr := s.canAccessInstallation(ctx, actor, installationID, false)
		if accessErr != nil {
			return "", false, WorkflowUploadPolicy{}, accessErr
		}
		allowed = canAccess && s.HasPermission(ctx, actor, "installation.progress")
	case "ORDER":
		err := s.db.QueryRowContext(ctx, `SELECT (SELECT id FROM workflow_instances WHERE order_id=o.id ORDER BY started_at LIMIT 1),o.customer_user_id FROM orders o WHERE o.id=$1`, entityID).Scan(&workflow, &owner)
		if err != nil {
			return "", false, WorkflowUploadPolicy{}, err
		}
		if owner == actor {
			allowed = s.HasPermission(ctx, actor, "customer_portal.acceptance.confirm_own")
			customerVisible = true
		} else {
			allowed = s.HasPermission(ctx, actor, "customer_acceptance.record")
		}
	default:
		return "", false, WorkflowUploadPolicy{}, ErrValidation
	}
	if !allowed {
		return "", false, WorkflowUploadPolicy{}, ErrForbidden
	}
	if workflow.Valid {
		workflowID = workflow.String
	}
	return workflowID, customerVisible, WorkflowUploadPolicy{AllowedMIMETypes: []string{"image/png", "image/jpeg", "application/pdf"}, MaxSizeBytes: s.MaximumUploadBytes(ctx,15 << 20)}, nil
}

func (s *OperationsService) RegisterOperationalFile(ctx context.Context, actor, workflowID, entityType, entityID, storageKey, originalName, mimeType string, size int64, customerVisible bool) (WorkflowFile, error) {
	var file WorkflowFile
	var storedWorkflow sql.NullString
	err := s.db.QueryRowContext(ctx, `INSERT INTO workflow_files(workflow_instance_id,entity_type,entity_id,storage_key,original_file_name,mime_type,size_bytes,customer_visible,uploaded_by_user_id) VALUES(NULLIF($1,'')::uuid,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id,workflow_instance_id,workflow_step_instance_id,field_definition_id,entity_type,entity_id,storage_key,original_file_name,mime_type,size_bytes,customer_visible`, workflowID, normalizeCode(entityType), entityID, storageKey, originalName, mimeType, size, customerVisible, actor).Scan(&file.ID, &storedWorkflow, &file.WorkflowStepInstanceID, &file.FieldDefinitionID, &file.EntityType, &file.EntityID, &file.StorageKey, &file.OriginalFileName, &file.MIMEType, &file.SizeBytes, &file.CustomerVisible)
	if storedWorkflow.Valid {
		file.WorkflowInstanceID = storedWorkflow.String
	}
	if err == nil {
		s.audit(ctx, actor, "workflow_files.upload", "workflow_file", file.ID, map[string]any{"entity_type": entityType, "entity_id": entityID})
	}
	return file, err
}

func (s *OperationsService) WorkflowFileForDownload(ctx context.Context, actor, fileID string) (WorkflowFile, error) {
	var file WorkflowFile
	var workflow sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT f.id,f.workflow_instance_id,f.workflow_step_instance_id,f.field_definition_id,f.entity_type,f.entity_id,f.storage_key,f.original_file_name,f.mime_type,f.size_bytes,f.customer_visible FROM workflow_files f WHERE f.id=$1`, fileID).Scan(&file.ID, &workflow, &file.WorkflowStepInstanceID, &file.FieldDefinitionID, &file.EntityType, &file.EntityID, &file.StorageKey, &file.OriginalFileName, &file.MIMEType, &file.SizeBytes, &file.CustomerVisible)
	if err != nil {
		return file, err
	}
	if workflow.Valid {
		file.WorkflowInstanceID = workflow.String
	}
	var customerID string
	if workflow.Valid {
		err = s.db.QueryRowContext(ctx, `SELECT customer_user_id FROM workflow_instances WHERE id=$1`, workflow.String).Scan(&customerID)
	} else {
		switch file.EntityType {
		case "QUALITY_INSPECTION":
			err = s.db.QueryRowContext(ctx, `SELECT o.customer_user_id FROM quality_inspections q JOIN orders o ON o.id=q.order_id WHERE q.id=$1`, file.EntityID).Scan(&customerID)
		case "INSTALLATION":
			err = s.db.QueryRowContext(ctx, `SELECT customer_user_id FROM installation_jobs WHERE id=$1`, file.EntityID).Scan(&customerID)
		case "INSTALLATION_UPDATE":
			err = s.db.QueryRowContext(ctx, `SELECT j.customer_user_id FROM installation_updates u JOIN installation_jobs j ON j.id=u.installation_job_id WHERE u.id=$1`, file.EntityID).Scan(&customerID)
		case "ORDER":
			err = s.db.QueryRowContext(ctx, `SELECT customer_user_id FROM orders WHERE id=$1`, file.EntityID).Scan(&customerID)
		default:
			err = sql.ErrNoRows
		}
	}
	if err != nil {
		return file, err
	}
	if actor == customerID {
		if file.EntityType == "QUALITY_INSPECTION" {
			return file, ErrForbidden
		}
		if !file.CustomerVisible || !s.HasPermission(ctx, actor, "workflow_files.view_customer") {
			return file, ErrForbidden
		}
		return file, nil
	}
	if !s.HasPermission(ctx, actor, "workflow_files.view_internal") {
		return file, ErrForbidden
	}
	if file.EntityType == "QUALITY_INSPECTION" && !s.HasPermission(ctx, actor, "quality.view_all") && !s.HasPermission(ctx, actor, "quality.view_assigned") {
		return file, ErrForbidden
	}
	if (file.EntityType == "INSTALLATION" || file.EntityType == "INSTALLATION_UPDATE") && !file.CustomerVisible && !s.HasPermission(ctx, actor, "installation.view_all") && !s.HasPermission(ctx, actor, "installation.progress") && !s.HasPermission(ctx, actor, "installation.update") {
		return file, ErrForbidden
	}
	if !workflow.Valid {
		switch file.EntityType {
		case "QUALITY_INSPECTION":
			if !s.HasPermission(ctx, actor, "quality.view_all") && !s.HasPermission(ctx, actor, "quality.view_assigned") {
				return file, ErrForbidden
			}
		case "INSTALLATION", "INSTALLATION_UPDATE":
			if !s.HasPermission(ctx, actor, "installation.view_all") && !s.HasPermission(ctx, actor, "installation.view_assigned") {
				return file, ErrForbidden
			}
		case "ORDER":
			if !s.HasPermission(ctx, actor, "customer_acceptance.record") {
				return file, ErrForbidden
			}
		}
		return file, nil
	}
	_, err = s.GetWorkflowRuntime(ctx, actor, file.WorkflowInstanceID)
	if errors.Is(err, sql.ErrNoRows) {
		return file, ErrForbidden
	}
	return file, err
}

func validateFileReferenceTx(ctx context.Context, tx *sql.Tx, stepID string, fieldID int64, raw map[string]any) error {
	fileID, ok := raw["fileId"].(string)
	if !ok || fileID == "" {
		return errors.New("fileId required")
	}
	var valid bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM workflow_files WHERE id=$1 AND workflow_step_instance_id=$2 AND field_definition_id=$3)`, fileID, stepID, fieldID).Scan(&valid); err != nil {
		return err
	}
	if !valid {
		return errors.New("file does not belong to this workflow field")
	}
	return nil
}
