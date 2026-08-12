package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var ErrTemplateImmutable = errors.New("published or archived template is immutable")
var codePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{1,79}$`)

func (s *OperationsService) ensureDraft(ctx context.Context, id int64) error {
	var status string
	if err := s.db.QueryRowContext(ctx, `SELECT status FROM workflow_templates WHERE id=$1`, id).Scan(&status); err != nil {
		return err
	}
	if status != "DRAFT" {
		return ErrTemplateImmutable
	}
	return nil
}
func (s *OperationsService) templateIDForStep(ctx context.Context, stepID int64) (int64, error) {
	var id int64
	err := s.db.QueryRowContext(ctx, `SELECT workflow_template_id FROM workflow_template_steps WHERE id=$1`, stepID).Scan(&id)
	return id, err
}

func (s *OperationsService) ListWorkflowStepCatalogue(ctx context.Context) ([]WorkflowStepCatalogueItem, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,step_code,title_fa,customer_title_fa,description_fa,COALESCE(default_role_code,''),COALESCE(default_permission_code,''),default_duration_hours FROM workflow_step_catalogue ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []WorkflowStepCatalogueItem{}
	for rows.Next() {
		var item WorkflowStepCatalogueItem
		if err := rows.Scan(&item.ID, &item.StepCode, &item.TitleFA, &item.CustomerTitleFA, &item.DescriptionFA, &item.DefaultRoleCode, &item.DefaultPermissionCode, &item.DefaultDurationHours); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *OperationsService) ListWorkflowTemplateVersions(ctx context.Context) ([]WorkflowTemplateVersion, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,template_group_code,version_number,code,name_fa,description_fa,icon_key,status,start_permission_code,is_active,created_from_template_id,published_at,created_at,updated_at,scope_type,max_iterations FROM workflow_templates ORDER BY template_group_code,version_number DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []WorkflowTemplateVersion{}
	for rows.Next() {
		var t WorkflowTemplateVersion
		var from sql.NullInt64
		var published sql.NullTime
		if err := rows.Scan(&t.ID, &t.TemplateGroupCode, &t.VersionNumber, &t.Code, &t.NameFA, &t.DescriptionFA, &t.IconKey, &t.Status, &t.StartPermissionCode, &t.IsActive, &from, &published, &t.CreatedAt, &t.UpdatedAt, &t.ScopeType, &t.MaxIterations); err != nil {
			return nil, err
		}
		if from.Valid {
			v := from.Int64
			t.CreatedFromTemplateID = &v
		}
		if published.Valid {
			v := published.Time
			t.PublishedAt = &v
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *OperationsService) GetWorkflowTemplateVersion(ctx context.Context, id int64) (WorkflowTemplateVersion, error) {
	var t WorkflowTemplateVersion
	var from sql.NullInt64
	var published sql.NullTime
	err := s.db.QueryRowContext(ctx, `SELECT id,template_group_code,version_number,code,name_fa,description_fa,icon_key,status,start_permission_code,is_active,created_from_template_id,published_at,created_at,updated_at,scope_type,max_iterations FROM workflow_templates WHERE id=$1`, id).Scan(&t.ID, &t.TemplateGroupCode, &t.VersionNumber, &t.Code, &t.NameFA, &t.DescriptionFA, &t.IconKey, &t.Status, &t.StartPermissionCode, &t.IsActive, &from, &published, &t.CreatedAt, &t.UpdatedAt, &t.ScopeType, &t.MaxIterations)
	if err != nil {
		return t, err
	}
	if from.Valid {
		v := from.Int64
		t.CreatedFromTemplateID = &v
	}
	if published.Valid {
		v := published.Time
		t.PublishedAt = &v
	}
	rows, err := s.db.QueryContext(ctx, `SELECT s.id,s.workflow_template_id,s.step_code,s.internal_title_fa,s.internal_description_fa,s.customer_title_fa,s.customer_description_fa,s.sequence_number,s.responsible_role_id,COALESCE(r.code,''),COALESCE(s.required_permission_code,''),s.customer_visible,s.requires_approval,s.approval_role_id,s.is_optional,s.is_skippable,s.is_active,s.default_duration_hours,s.starts_automatically,s.is_entry,s.domain_event_code FROM workflow_template_steps s LEFT JOIN roles r ON r.id=s.responsible_role_id WHERE s.workflow_template_id=$1 ORDER BY s.sequence_number`, id)
	if err != nil {
		return t, err
	}
	defer rows.Close()
	for rows.Next() {
		var st WorkflowTemplateStepV2
		var role, approval sql.NullInt64
		var domainEvent sql.NullString
		if err := rows.Scan(&st.ID, &st.WorkflowTemplateID, &st.StepCode, &st.InternalTitleFA, &st.InternalDescriptionFA, &st.CustomerTitleFA, &st.CustomerDescriptionFA, &st.SequenceNumber, &role, &st.ResponsibleRoleCode, &st.RequiredPermissionCode, &st.CustomerVisible, &st.RequiresApproval, &approval, &st.IsOptional, &st.IsSkippable, &st.IsActive, &st.DefaultDurationHours, &st.StartsAutomatically, &st.IsEntry, &domainEvent); err != nil {
			return t, err
		}
		if role.Valid {
			v := role.Int64
			st.ResponsibleRoleID = &v
		}
		if approval.Valid {
			v := approval.Int64
			st.ApprovalRoleID = &v
		}
		st.DomainEventCode = scanNullableString(domainEvent)
		st.Fields, _ = s.listTemplateFields(ctx, st.ID)
		st.Tasks, _ = s.listTemplateTasks(ctx, st.ID)
		t.Steps = append(t.Steps, st)
	}
	t.Metrics, _ = s.ListHandoffMetrics(ctx, id)
	t.Tasks, _ = s.ListWorkflowLevelTasks(ctx, id)
	t.Transitions, _ = s.ListWorkflowTransitions(ctx, id)
	return t, rows.Err()
}

func (s *OperationsService) CreateWorkflowTemplateDraft(ctx context.Context, actor string, p WorkflowTemplatePayload) (WorkflowTemplateVersion, error) {
	p.TemplateGroupCode = strings.ToLower(strings.TrimSpace(p.TemplateGroupCode))
	if !codePattern.MatchString(p.TemplateGroupCode) || strings.TrimSpace(p.NameFA) == "" || strings.TrimSpace(p.StartPermissionCode) == "" {
		return WorkflowTemplateVersion{}, errors.New("invalid template payload")
	}
	if p.Code == "" {
		p.Code = p.TemplateGroupCode + "_v1"
	}
	if p.ScopeType == "" {
		p.ScopeType = "ORDER"
	}
	p.ScopeType = normalizeCode(p.ScopeType)
	if p.MaxIterations == 0 {
		p.MaxIterations = 20
	}
	var id int64
	err := s.db.QueryRowContext(ctx, `INSERT INTO workflow_templates(template_group_code,version_number,code,name_fa,description_fa,icon_key,status,start_permission_code,is_active,created_by_user_id,scope_type,max_iterations) SELECT $1,COALESCE(MAX(version_number),0)+1,$2,$3,$4,$5,'DRAFT',$6,$7,$8,$9,$10 FROM workflow_templates WHERE template_group_code=$1 RETURNING id`, p.TemplateGroupCode, p.Code, p.NameFA, p.DescriptionFA, p.IconKey, p.StartPermissionCode, p.IsActive, actor, p.ScopeType, p.MaxIterations).Scan(&id)
	if err != nil {
		return WorkflowTemplateVersion{}, err
	}
	s.audit(ctx, actor, "workflow_templates.create", "workflow_template", fmt.Sprint(id), p)
	return s.GetWorkflowTemplateVersion(ctx, id)
}

func (s *OperationsService) UpdateWorkflowTemplateDraft(ctx context.Context, actor string, id int64, p WorkflowTemplatePayload) error {
	if err := s.ensureDraft(ctx, id); err != nil {
		return err
	}
	if p.ScopeType == "" {
		p.ScopeType = "ORDER"
	}
	if p.MaxIterations == 0 {
		p.MaxIterations = 20
	}
	_, err := s.db.ExecContext(ctx, `UPDATE workflow_templates SET name_fa=$2,description_fa=$3,icon_key=$4,start_permission_code=$5,is_active=$6,scope_type=$7,max_iterations=$8,updated_at=NOW() WHERE id=$1`, id, p.NameFA, p.DescriptionFA, p.IconKey, p.StartPermissionCode, p.IsActive, normalizeCode(p.ScopeType), p.MaxIterations)
	if err == nil {
		s.audit(ctx, actor, "workflow_templates.update", "workflow_template", fmt.Sprint(id), p)
	}
	return err
}

func (s *OperationsService) CloneWorkflowTemplate(ctx context.Context, actor string, sourceID int64) (WorkflowTemplateVersion, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return WorkflowTemplateVersion{}, err
	}
	defer tx.Rollback()
	var group, name, desc, icon, start, scope string
	var maxIterations int
	var active bool
	if err = tx.QueryRowContext(ctx, `SELECT template_group_code,name_fa,description_fa,icon_key,start_permission_code,is_active,scope_type,max_iterations FROM workflow_templates WHERE id=$1`, sourceID).Scan(&group, &name, &desc, &icon, &start, &active, &scope, &maxIterations); err != nil {
		return WorkflowTemplateVersion{}, err
	}
	var version int
	if err = tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(version_number),0)+1 FROM workflow_templates WHERE template_group_code=$1`, group).Scan(&version); err != nil {
		return WorkflowTemplateVersion{}, err
	}
	var id int64
	code := fmt.Sprintf("%s_v%d", group, version)
	if err = tx.QueryRowContext(ctx, `INSERT INTO workflow_templates(template_group_code,version_number,code,name_fa,description_fa,icon_key,status,start_permission_code,is_active,created_from_template_id,created_by_user_id,scope_type,max_iterations) VALUES($1,$2,$3,$4,$5,$6,'DRAFT',$7,$8,$9,$10,$11,$12) RETURNING id`, group, version, code, name, desc, icon, start, active, sourceID, actor, scope, maxIterations).Scan(&id); err != nil {
		return WorkflowTemplateVersion{}, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO workflow_template_steps(workflow_template_id,step_code,internal_title_fa,internal_description_fa,customer_title_fa,customer_description_fa,sequence_number,responsible_role_code,responsible_role_id,required_permission_code,customer_visible,is_first_step,requires_approval,approval_role_id,is_optional,is_skippable,is_active,default_duration_hours,starts_automatically,is_entry,domain_event_code) SELECT $1,step_code,internal_title_fa,internal_description_fa,customer_title_fa,customer_description_fa,sequence_number,responsible_role_code,responsible_role_id,required_permission_code,customer_visible,is_first_step,requires_approval,approval_role_id,is_optional,is_skippable,is_active,default_duration_hours,starts_automatically,is_entry,domain_event_code FROM workflow_template_steps WHERE workflow_template_id=$2`, id, sourceID)
	if err != nil {
		return WorkflowTemplateVersion{}, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO workflow_step_field_definitions(workflow_template_step_id,field_key,label_fa,description_fa,field_type,is_required,is_customer_visible,is_sales_visible,is_internal_cost,unit_code,currency_code,placeholder_fa,default_value,options_json,validation_json,sort_order,handoff_metric_key,handoff_direction) SELECT ns.id,f.field_key,f.label_fa,f.description_fa,f.field_type,f.is_required,f.is_customer_visible,f.is_sales_visible,f.is_internal_cost,f.unit_code,f.currency_code,f.placeholder_fa,f.default_value,f.options_json,f.validation_json,f.sort_order,f.handoff_metric_key,f.handoff_direction FROM workflow_step_field_definitions f JOIN workflow_template_steps os ON os.id=f.workflow_template_step_id JOIN workflow_template_steps ns ON ns.workflow_template_id=$1 AND ns.step_code=os.step_code WHERE os.workflow_template_id=$2`, id, sourceID)
	if err != nil {
		return WorkflowTemplateVersion{}, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO workflow_step_task_templates(workflow_template_step_id,trigger_type,title_fa,description_fa,assigned_role_id,required_permission_code,priority,due_offset_hours,blocks_step_completion) SELECT ns.id,t.trigger_type,t.title_fa,t.description_fa,t.assigned_role_id,t.required_permission_code,t.priority,t.due_offset_hours,t.blocks_step_completion FROM workflow_step_task_templates t JOIN workflow_template_steps os ON os.id=t.workflow_template_step_id JOIN workflow_template_steps ns ON ns.workflow_template_id=$1 AND ns.step_code=os.step_code WHERE os.workflow_template_id=$2`, id, sourceID)
	if err != nil {
		return WorkflowTemplateVersion{}, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO workflow_handoff_metric_definitions(workflow_template_id,metric_key,label_fa,unit_code,absolute_tolerance,percentage_tolerance,blocking_on_mismatch) SELECT $1,metric_key,label_fa,unit_code,absolute_tolerance,percentage_tolerance,blocking_on_mismatch FROM workflow_handoff_metric_definitions WHERE workflow_template_id=$2`, id, sourceID)
	if err != nil {
		return WorkflowTemplateVersion{}, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO workflow_task_templates(workflow_template_id,trigger_type,title_fa,description_fa,assigned_role_id,required_permission_code,priority,due_offset_hours,blocks_workflow_progress) SELECT $1,trigger_type,title_fa,description_fa,assigned_role_id,required_permission_code,priority,due_offset_hours,blocks_workflow_progress FROM workflow_task_templates WHERE workflow_template_id=$2`, id, sourceID)
	if err != nil {
		return WorkflowTemplateVersion{}, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO workflow_step_transitions(workflow_template_id,source_step_id,target_step_id,transition_code,label_fa,transition_type,result_code,is_default,requires_permission_code,requires_reason,sort_order) SELECT $1,ns.id,nt.id,t.transition_code,t.label_fa,t.transition_type,t.result_code,t.is_default,t.requires_permission_code,t.requires_reason,t.sort_order FROM workflow_step_transitions t JOIN workflow_template_steps os ON os.id=t.source_step_id JOIN workflow_template_steps ot ON ot.id=t.target_step_id JOIN workflow_template_steps ns ON ns.workflow_template_id=$1 AND ns.step_code=os.step_code JOIN workflow_template_steps nt ON nt.workflow_template_id=$1 AND nt.step_code=ot.step_code WHERE t.workflow_template_id=$2`, id, sourceID)
	if err != nil {
		return WorkflowTemplateVersion{}, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO workflow_template_document_requirements(workflow_template_id,workflow_template_step_id,document_type,title_fa,is_required,is_blocking,customer_visible,sort_order) SELECT $1,ns.id,r.document_type,r.title_fa,r.is_required,r.is_blocking,r.customer_visible,r.sort_order FROM workflow_template_document_requirements r LEFT JOIN workflow_template_steps os ON os.id=r.workflow_template_step_id LEFT JOIN workflow_template_steps ns ON ns.workflow_template_id=$1 AND ns.step_code=os.step_code WHERE r.workflow_template_id=$2`, id, sourceID)
	if err != nil {
		return WorkflowTemplateVersion{}, err
	}
	s.auditTx(ctx, tx, actor, "workflow_templates.clone", "workflow_template", fmt.Sprint(id), nil, map[string]any{"source_id": sourceID})
	if err = tx.Commit(); err != nil {
		return WorkflowTemplateVersion{}, err
	}
	return s.GetWorkflowTemplateVersion(ctx, id)
}

func (s *OperationsService) PublishWorkflowTemplate(ctx context.Context, actor string, id int64) error {
	t, err := s.GetWorkflowTemplateVersion(ctx, id)
	if err != nil {
		return err
	}
	if err = validateTemplateForPublish(t); err != nil {
		return err
	}
	if err = s.validateTemplateReferences(ctx, id); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var status string
	if err = tx.QueryRowContext(ctx, `SELECT status FROM workflow_templates WHERE id=$1 FOR UPDATE`, id).Scan(&status); err != nil {
		return err
	}
	if status != "DRAFT" {
		return ErrTemplateImmutable
	}
	result, err := tx.ExecContext(ctx, `UPDATE workflow_templates SET status='PUBLISHED',published_at=NOW(),published_by_user_id=$2,updated_at=NOW() WHERE id=$1 AND status='DRAFT'`, id, actor)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return ErrTemplateImmutable
	}
	s.auditTx(ctx, tx, actor, "workflow_templates.publish", "workflow_template", fmt.Sprint(id), map[string]any{"status": "DRAFT"}, map[string]any{"status": "PUBLISHED"})
	return tx.Commit()
}

func (s *OperationsService) validateTemplateReferences(ctx context.Context, templateID int64) error {
	var invalidSteps int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM workflow_template_steps s LEFT JOIN roles r ON r.id=s.responsible_role_id AND r.is_active LEFT JOIN permissions p ON p.code=s.required_permission_code AND p.is_active LEFT JOIN roles ar ON ar.id=s.approval_role_id AND ar.is_active WHERE s.workflow_template_id=$1 AND s.is_active AND (s.sequence_number<=0 OR r.id IS NULL OR p.id IS NULL OR (s.requires_approval AND ar.id IS NULL))`, templateID).Scan(&invalidSteps)
	if err != nil {
		return err
	}
	if invalidSteps > 0 {
		return errors.New("template contains an invalid role, permission, approval role, or sequence")
	}
	var invalidTasks int
	err = s.db.QueryRowContext(ctx, `SELECT (SELECT COUNT(*) FROM workflow_step_task_templates t JOIN workflow_template_steps s ON s.id=t.workflow_template_step_id LEFT JOIN roles r ON r.id=t.assigned_role_id AND r.is_active LEFT JOIN permissions p ON p.code=t.required_permission_code AND p.is_active WHERE s.workflow_template_id=$1 AND (r.id IS NULL OR p.id IS NULL))+(SELECT COUNT(*) FROM workflow_task_templates t LEFT JOIN roles r ON r.id=t.assigned_role_id AND r.is_active LEFT JOIN permissions p ON p.code=t.required_permission_code AND p.is_active WHERE t.workflow_template_id=$1 AND (r.id IS NULL OR p.id IS NULL))`, templateID).Scan(&invalidTasks)
	if err != nil {
		return err
	}
	if invalidTasks > 0 {
		return errors.New("template contains a task with an invalid role or permission")
	}
	var invalidDocuments int
	err = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM workflow_template_document_requirements r LEFT JOIN workflow_template_steps st ON st.id=r.workflow_template_step_id AND st.workflow_template_id=r.workflow_template_id AND st.is_active WHERE r.workflow_template_id=$1 AND (TRIM(r.title_fa)='' OR r.document_type NOT IN ('PROFORMA','SALES_INVOICE','PAYMENT_RECEIPT','ORDER_SUMMARY','PACKING_LIST','DELIVERY_NOTE','TRANSPORT_RECEIPT','INSTALLATION_REPORT','QUALITY_REPORT','CUSTOMER_ACCEPTANCE','COMMERCIAL_INVOICE','EXPORT_PACKING_LIST','CUSTOMS_DOCUMENT','CERTIFICATE_OF_ORIGIN','CUSTOMS_DECLARATION','BILL_OF_LADING','CERTIFICATE','OTHER') OR (r.workflow_template_step_id IS NOT NULL AND st.id IS NULL))`, templateID).Scan(&invalidDocuments)
	if err != nil {
		return err
	}
	if invalidDocuments > 0 {
		return errors.New("template contains an invalid document requirement")
	}
	return s.validateWorkflowTransitions(ctx, templateID)
}
func (s *OperationsService) ArchiveWorkflowTemplate(ctx context.Context, actor string, id int64) error {
	res, err := s.db.ExecContext(ctx, `UPDATE workflow_templates SET status='ARCHIVED',is_active=FALSE,updated_at=NOW() WHERE id=$1 AND status='PUBLISHED'`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("only published templates can be archived")
	}
	s.audit(ctx, actor, "workflow_templates.archive", "workflow_template", fmt.Sprint(id), nil)
	return nil
}

func validateTemplateForPublish(t WorkflowTemplateVersion) error {
	active := 0
	seenSeq := map[int]bool{}
	seenCode := map[string]bool{}
	metrics := map[string]bool{}
	for _, m := range t.Metrics {
		metrics[m.MetricKey] = true
	}
	for _, st := range t.Steps {
		if !st.IsActive {
			continue
		}
		active++
		if seenSeq[st.SequenceNumber] || seenCode[st.StepCode] {
			return errors.New("duplicate step sequence or code")
		}
		seenSeq[st.SequenceNumber] = true
		seenCode[st.StepCode] = true
		if st.ResponsibleRoleID == nil || st.RequiredPermissionCode == "" {
			return fmt.Errorf("step %s requires role and permission", st.StepCode)
		}
		if st.RequiresApproval && st.ApprovalRoleID == nil {
			return fmt.Errorf("step %s requires approval role", st.StepCode)
		}
		for _, f := range st.Fields {
			if err := validateFieldDefinition(f.FieldKey, f.FieldType, f.OptionsJSON, f.ValidationJSON, f.HandoffMetricKey, f.HandoffDirection, f.UnitCode, f.CurrencyCode, f.IsInternalCost, f.IsCustomerVisible, metrics); err != nil {
				return fmt.Errorf("field %s: %w", f.FieldKey, err)
			}
		}
		for _, task := range st.Tasks {
			if !stepTriggerTypes[task.TriggerType] {
				return errors.New("invalid task trigger")
			}
			if task.BlocksStepCompletion && (task.TriggerType == "ON_STEP_APPROVE" || task.TriggerType == "ON_STEP_COMPLETE") {
				return errors.New("late trigger cannot block step completion")
			}
		}
	}
	if active == 0 {
		return errors.New("template requires at least one active step")
	}
	return nil
}
func validateFieldDefinition(key, kind string, options, validation json.RawMessage, metric, direction, unit, currency *string, internalCost, customerVisible bool, metrics map[string]bool) error {
	if !codePattern.MatchString(key) {
		return errors.New("invalid field key")
	}
	if !fieldTypes[kind] {
		return errors.New("invalid field type")
	}
	if kind == "SELECT" || kind == "MULTI_SELECT" {
		var vals []any
		if len(options) == 0 || json.Unmarshal(options, &vals) != nil || len(vals) == 0 {
			return errors.New("select options required")
		}
	}
	if (kind == "WEIGHT" || kind == "AREA" || kind == "VOLUME" || kind == "QUANTITY") && (unit == nil || strings.TrimSpace(*unit) == "") {
		return errors.New("measurement unit is required")
	}
	if kind == "MONEY" && (currency == nil || len(strings.TrimSpace(*currency)) != 3) {
		return errors.New("money currency is required")
	}
	if internalCost && customerVisible {
		return errors.New("internal cost field cannot be customer visible")
	}
	if len(validation) > 0 && !json.Valid(validation) {
		return errors.New("invalid validation json")
	}
	if metric != nil && !metrics[*metric] {
		return errors.New("unknown handoff metric")
	}
	if direction != nil && metric == nil {
		return errors.New("handoff direction requires a metric")
	}
	if direction != nil && *direction != "IN" && *direction != "OUT" {
		return errors.New("handoff direction must be IN or OUT")
	}
	return nil
}

func (s *OperationsService) AddWorkflowStep(ctx context.Context, actor string, templateID int64, p WorkflowStepPayload) (WorkflowTemplateStepV2, error) {
	if err := s.ensureDraft(ctx, templateID); err != nil {
		return WorkflowTemplateStepV2{}, err
	}
	if !codePattern.MatchString(strings.ToLower(p.StepCode)) || p.InternalTitleFA == "" || p.ResponsibleRoleID == nil || p.RequiredPermissionCode == "" {
		return WorkflowTemplateStepV2{}, errors.New("invalid step")
	}
	if p.DefaultDurationHours <= 0 {
		p.DefaultDurationHours = 24
	}
	var seq int
	_ = s.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(sequence_number),0)+1 FROM workflow_template_steps WHERE workflow_template_id=$1`, templateID).Scan(&seq)
	var id int64
	err := s.db.QueryRowContext(ctx, `INSERT INTO workflow_template_steps(workflow_template_id,step_code,internal_title_fa,internal_description_fa,customer_title_fa,customer_description_fa,sequence_number,responsible_role_id,responsible_role_code,required_permission_code,customer_visible,is_first_step,requires_approval,approval_role_id,is_optional,is_skippable,is_active,default_duration_hours,starts_automatically,is_entry,domain_event_code) SELECT $1,UPPER($2),$3,$4,$5,$6,$7,$8,r.code,$9,$10,$18,$11,$12,$13,$14,$15,$16,$17,$18,$19 FROM roles r WHERE r.id=$8 RETURNING id`, templateID, p.StepCode, p.InternalTitleFA, p.InternalDescriptionFA, p.CustomerTitleFA, p.CustomerDescriptionFA, seq, p.ResponsibleRoleID, p.RequiredPermissionCode, p.CustomerVisible, p.RequiresApproval, p.ApprovalRoleID, p.IsOptional, p.IsSkippable, p.IsActive, p.DefaultDurationHours, p.StartsAutomatically, p.IsEntry, p.DomainEventCode).Scan(&id)
	if err != nil {
		return WorkflowTemplateStepV2{}, err
	}
	s.audit(ctx, actor, "workflow_steps.create", "workflow_template_step", fmt.Sprint(id), p)
	t, _ := s.GetWorkflowTemplateVersion(ctx, templateID)
	for _, st := range t.Steps {
		if st.ID == id {
			return st, nil
		}
	}
	return WorkflowTemplateStepV2{}, sql.ErrNoRows
}
func (s *OperationsService) UpdateWorkflowStep(ctx context.Context, actor string, stepID int64, p WorkflowStepPayload) error {
	templateID, err := s.templateIDForStep(ctx, stepID)
	if err != nil {
		return err
	}
	if err = s.ensureDraft(ctx, templateID); err != nil {
		return err
	}
	if p.DefaultDurationHours <= 0 {
		p.DefaultDurationHours = 24
	}
	_, err = s.db.ExecContext(ctx, `UPDATE workflow_template_steps SET internal_title_fa=$2,internal_description_fa=$3,customer_title_fa=$4,customer_description_fa=$5,responsible_role_id=$6,responsible_role_code=(SELECT code FROM roles WHERE id=$6),required_permission_code=$7,customer_visible=$8,requires_approval=$9,approval_role_id=$10,is_optional=$11,is_skippable=$12,is_active=$13,default_duration_hours=$14,starts_automatically=$15,is_entry=$16,domain_event_code=$17,updated_at=NOW() WHERE id=$1`, stepID, p.InternalTitleFA, p.InternalDescriptionFA, p.CustomerTitleFA, p.CustomerDescriptionFA, p.ResponsibleRoleID, p.RequiredPermissionCode, p.CustomerVisible, p.RequiresApproval, p.ApprovalRoleID, p.IsOptional, p.IsSkippable, p.IsActive, p.DefaultDurationHours, p.StartsAutomatically, p.IsEntry, p.DomainEventCode)
	if err == nil {
		s.audit(ctx, actor, "workflow_steps.update", "workflow_template_step", fmt.Sprint(stepID), p)
	}
	return err
}
func (s *OperationsService) DeleteWorkflowStep(ctx context.Context, actor string, stepID int64) error {
	templateID, err := s.templateIDForStep(ctx, stepID)
	if err != nil {
		return err
	}
	if err = s.ensureDraft(ctx, templateID); err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `DELETE FROM workflow_template_steps WHERE id=$1`, stepID)
	if err == nil {
		s.audit(ctx, actor, "workflow_steps.delete", "workflow_template_step", fmt.Sprint(stepID), nil)
	}
	return err
}
func (s *OperationsService) ReorderWorkflowSteps(ctx context.Context, actor string, templateID int64, ids []int64) error {
	if err := s.ensureDraft(ctx, templateID); err != nil {
		return err
	}
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM workflow_template_steps WHERE workflow_template_id=$1`, templateID).Scan(&count); err != nil {
		return err
	}
	if count != len(ids) {
		return errors.New("reorder must include every step")
	}
	seenIDs := map[int64]bool{}
	for _, id := range ids {
		if seenIDs[id] {
			return errors.New("reorder contains duplicate step id")
		}
		seenIDs[id] = true
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for i, id := range ids {
		res, err := tx.ExecContext(ctx, `UPDATE workflow_template_steps SET sequence_number=-$3 WHERE id=$1 AND workflow_template_id=$2`, id, templateID, i+1)
		if err != nil {
			return err
		}
		n, _ := res.RowsAffected()
		if n != 1 {
			return errors.New("invalid step id")
		}
	}
	_, err = tx.ExecContext(ctx, `UPDATE workflow_template_steps SET sequence_number=-sequence_number,is_first_step=(-sequence_number=1),is_entry=(-sequence_number=1),updated_at=NOW() WHERE workflow_template_id=$1`, templateID)
	if err != nil {
		return err
	}
	s.auditTx(ctx, tx, actor, "workflow_steps.reorder", "workflow_template", fmt.Sprint(templateID), nil, map[string]any{"step_ids": ids})
	return tx.Commit()
}
func (s *OperationsService) DuplicateWorkflowStep(ctx context.Context, actor string, stepID int64) (WorkflowTemplateStepV2, error) {
	templateID, err := s.templateIDForStep(ctx, stepID)
	if err != nil {
		return WorkflowTemplateStepV2{}, err
	}
	if err = s.ensureDraft(ctx, templateID); err != nil {
		return WorkflowTemplateStepV2{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return WorkflowTemplateStepV2{}, err
	}
	defer tx.Rollback()
	var newID int64
	err = tx.QueryRowContext(ctx, `INSERT INTO workflow_template_steps(workflow_template_id,step_code,internal_title_fa,internal_description_fa,customer_title_fa,customer_description_fa,sequence_number,responsible_role_code,responsible_role_id,required_permission_code,customer_visible,is_first_step,requires_approval,approval_role_id,is_optional,is_skippable,is_active,default_duration_hours,starts_automatically,is_entry,domain_event_code) SELECT workflow_template_id,step_code||'_COPY_'||EXTRACT(EPOCH FROM NOW())::bigint,internal_title_fa||' (کپی)',internal_description_fa,customer_title_fa,customer_description_fa,(SELECT COALESCE(MAX(sequence_number),0)+1 FROM workflow_template_steps WHERE workflow_template_id=$2),responsible_role_code,responsible_role_id,required_permission_code,customer_visible,FALSE,requires_approval,approval_role_id,is_optional,is_skippable,is_active,default_duration_hours,starts_automatically,FALSE,domain_event_code FROM workflow_template_steps WHERE id=$1 RETURNING id`, stepID, templateID).Scan(&newID)
	if err != nil {
		return WorkflowTemplateStepV2{}, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO workflow_step_field_definitions(workflow_template_step_id,field_key,label_fa,description_fa,field_type,is_required,is_customer_visible,is_sales_visible,is_internal_cost,unit_code,currency_code,placeholder_fa,default_value,options_json,validation_json,sort_order,handoff_metric_key,handoff_direction) SELECT $1,field_key,label_fa,description_fa,field_type,is_required,is_customer_visible,is_sales_visible,is_internal_cost,unit_code,currency_code,placeholder_fa,default_value,options_json,validation_json,sort_order,handoff_metric_key,handoff_direction FROM workflow_step_field_definitions WHERE workflow_template_step_id=$2`, newID, stepID)
	if err != nil {
		return WorkflowTemplateStepV2{}, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO workflow_step_task_templates(workflow_template_step_id,trigger_type,title_fa,description_fa,assigned_role_id,required_permission_code,priority,due_offset_hours,blocks_step_completion) SELECT $1,trigger_type,title_fa,description_fa,assigned_role_id,required_permission_code,priority,due_offset_hours,blocks_step_completion FROM workflow_step_task_templates WHERE workflow_template_step_id=$2`, newID, stepID)
	if err != nil {
		return WorkflowTemplateStepV2{}, err
	}
	if err = tx.Commit(); err != nil {
		return WorkflowTemplateStepV2{}, err
	}
	t, _ := s.GetWorkflowTemplateVersion(ctx, templateID)
	for _, st := range t.Steps {
		if st.ID == newID {
			return st, nil
		}
	}
	return WorkflowTemplateStepV2{}, sql.ErrNoRows
}

func (s *OperationsService) listTemplateFields(ctx context.Context, stepID int64) ([]WorkflowFieldDefinition, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,workflow_template_step_id,field_key,label_fa,description_fa,field_type,is_required,is_customer_visible,is_sales_visible,is_internal_cost,unit_code,currency_code,placeholder_fa,default_value,options_json,validation_json,sort_order,handoff_metric_key,handoff_direction FROM workflow_step_field_definitions WHERE workflow_template_step_id=$1 ORDER BY sort_order,id`, stepID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []WorkflowFieldDefinition{}
	for rows.Next() {
		var f WorkflowFieldDefinition
		var unit, currency, placeholder, metric, direction sql.NullString
		var def, opts, val []byte
		if err := rows.Scan(&f.ID, &f.WorkflowTemplateStepID, &f.FieldKey, &f.LabelFA, &f.DescriptionFA, &f.FieldType, &f.IsRequired, &f.IsCustomerVisible, &f.IsSalesVisible, &f.IsInternalCost, &unit, &currency, &placeholder, &def, &opts, &val, &f.SortOrder, &metric, &direction); err != nil {
			return nil, err
		}
		if unit.Valid {
			f.UnitCode = &unit.String
		}
		if currency.Valid {
			f.CurrencyCode = &currency.String
		}
		if placeholder.Valid {
			f.PlaceholderFA = &placeholder.String
		}
		if metric.Valid {
			f.HandoffMetricKey = &metric.String
		}
		if direction.Valid {
			f.HandoffDirection = &direction.String
		}
		f.DefaultValue = def
		f.OptionsJSON = opts
		f.ValidationJSON = val
		out = append(out, f)
	}
	return out, rows.Err()
}
func (s *OperationsService) AddWorkflowField(ctx context.Context, actor string, stepID int64, p WorkflowFieldPayload) (WorkflowFieldDefinition, error) {
	templateID, err := s.templateIDForStep(ctx, stepID)
	if err != nil {
		return WorkflowFieldDefinition{}, err
	}
	if err = s.ensureDraft(ctx, templateID); err != nil {
		return WorkflowFieldDefinition{}, err
	}
	metrics, _ := s.ListHandoffMetrics(ctx, templateID)
	mm := map[string]bool{}
	for _, m := range metrics {
		mm[m.MetricKey] = true
	}
	if err = validateFieldDefinition(p.FieldKey, p.FieldType, p.OptionsJSON, p.ValidationJSON, p.HandoffMetricKey, p.HandoffDirection, p.UnitCode, p.CurrencyCode, p.IsInternalCost, p.IsCustomerVisible, mm); err != nil {
		return WorkflowFieldDefinition{}, err
	}
	if len(p.ValidationJSON) == 0 {
		p.ValidationJSON = []byte(`{}`)
	}
	var id int64
	err = s.db.QueryRowContext(ctx, `INSERT INTO workflow_step_field_definitions(workflow_template_step_id,field_key,label_fa,description_fa,field_type,is_required,is_customer_visible,is_sales_visible,is_internal_cost,unit_code,currency_code,placeholder_fa,default_value,options_json,validation_json,sort_order,handoff_metric_key,handoff_direction) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18) RETURNING id`, stepID, p.FieldKey, p.LabelFA, p.DescriptionFA, p.FieldType, p.IsRequired, p.IsCustomerVisible, p.IsSalesVisible, p.IsInternalCost, p.UnitCode, p.CurrencyCode, p.PlaceholderFA, nullableJSON(p.DefaultValue), nullableJSON(p.OptionsJSON), p.ValidationJSON, p.SortOrder, p.HandoffMetricKey, p.HandoffDirection).Scan(&id)
	if err != nil {
		return WorkflowFieldDefinition{}, err
	}
	s.audit(ctx, actor, "workflow_fields.create", "workflow_field_definition", fmt.Sprint(id), p)
	fields, _ := s.listTemplateFields(ctx, stepID)
	for _, f := range fields {
		if f.ID == id {
			return f, nil
		}
	}
	return WorkflowFieldDefinition{}, sql.ErrNoRows
}
func (s *OperationsService) UpdateWorkflowField(ctx context.Context, actor string, fieldID int64, p WorkflowFieldPayload) error {
	var stepID, templateID int64
	if err := s.db.QueryRowContext(ctx, `SELECT f.workflow_template_step_id,s.workflow_template_id FROM workflow_step_field_definitions f JOIN workflow_template_steps s ON s.id=f.workflow_template_step_id WHERE f.id=$1`, fieldID).Scan(&stepID, &templateID); err != nil {
		return err
	}
	if err := s.ensureDraft(ctx, templateID); err != nil {
		return err
	}
	metrics, _ := s.ListHandoffMetrics(ctx, templateID)
	mm := map[string]bool{}
	for _, m := range metrics {
		mm[m.MetricKey] = true
	}
	if err := validateFieldDefinition(p.FieldKey, p.FieldType, p.OptionsJSON, p.ValidationJSON, p.HandoffMetricKey, p.HandoffDirection, p.UnitCode, p.CurrencyCode, p.IsInternalCost, p.IsCustomerVisible, mm); err != nil {
		return err
	}
	if len(p.ValidationJSON) == 0 {
		p.ValidationJSON = []byte(`{}`)
	}
	_, err := s.db.ExecContext(ctx, `UPDATE workflow_step_field_definitions SET field_key=$2,label_fa=$3,description_fa=$4,field_type=$5,is_required=$6,is_customer_visible=$7,is_sales_visible=$8,is_internal_cost=$9,unit_code=$10,currency_code=$11,placeholder_fa=$12,default_value=$13,options_json=$14,validation_json=$15,sort_order=$16,handoff_metric_key=$17,handoff_direction=$18,updated_at=NOW() WHERE id=$1`, fieldID, p.FieldKey, p.LabelFA, p.DescriptionFA, p.FieldType, p.IsRequired, p.IsCustomerVisible, p.IsSalesVisible, p.IsInternalCost, p.UnitCode, p.CurrencyCode, p.PlaceholderFA, nullableJSON(p.DefaultValue), nullableJSON(p.OptionsJSON), p.ValidationJSON, p.SortOrder, p.HandoffMetricKey, p.HandoffDirection)
	if err == nil {
		s.audit(ctx, actor, "workflow_fields.update", "workflow_field_definition", fmt.Sprint(fieldID), p)
	}
	return err
}
func (s *OperationsService) DeleteWorkflowField(ctx context.Context, actor string, fieldID int64) error {
	var templateID int64
	if err := s.db.QueryRowContext(ctx, `SELECT s.workflow_template_id FROM workflow_step_field_definitions f JOIN workflow_template_steps s ON s.id=f.workflow_template_step_id WHERE f.id=$1`, fieldID).Scan(&templateID); err != nil {
		return err
	}
	if err := s.ensureDraft(ctx, templateID); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM workflow_step_field_definitions WHERE id=$1`, fieldID)
	if err == nil {
		s.audit(ctx, actor, "workflow_fields.delete", "workflow_field_definition", fmt.Sprint(fieldID), nil)
	}
	return err
}
func (s *OperationsService) ReorderWorkflowFields(ctx context.Context, actor string, stepID int64, ids []int64) error {
	templateID, err := s.templateIDForStep(ctx, stepID)
	if err != nil {
		return err
	}
	if err = s.ensureDraft(ctx, templateID); err != nil {
		return err
	}
	var count int
	if err = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM workflow_step_field_definitions WHERE workflow_template_step_id=$1`, stepID).Scan(&count); err != nil {
		return err
	}
	if count != len(ids) {
		return errors.New("reorder must include every field")
	}
	seenIDs := map[int64]bool{}
	for _, id := range ids {
		if seenIDs[id] {
			return errors.New("reorder contains duplicate field id")
		}
		seenIDs[id] = true
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for i, id := range ids {
		res, e := tx.ExecContext(ctx, `UPDATE workflow_step_field_definitions SET sort_order=$3 WHERE id=$1 AND workflow_template_step_id=$2`, id, stepID, i+1)
		if e != nil {
			return e
		}
		n, _ := res.RowsAffected()
		if n != 1 {
			return errors.New("invalid field id")
		}
	}
	s.auditTx(ctx, tx, actor, "workflow_fields.reorder", "workflow_template_step", fmt.Sprint(stepID), nil, map[string]any{"field_ids": ids})
	return tx.Commit()
}

func (s *OperationsService) listTemplateTasks(ctx context.Context, stepID int64) ([]WorkflowTaskTemplate, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,workflow_template_step_id,trigger_type,title_fa,description_fa,assigned_role_id,COALESCE(required_permission_code,''),priority,due_offset_hours,blocks_step_completion FROM workflow_step_task_templates WHERE workflow_template_step_id=$1 ORDER BY id`, stepID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []WorkflowTaskTemplate{}
	for rows.Next() {
		var t WorkflowTaskTemplate
		var role, due sql.NullInt64
		if err := rows.Scan(&t.ID, &t.WorkflowTemplateStepID, &t.TriggerType, &t.TitleFA, &t.DescriptionFA, &role, &t.RequiredPermissionCode, &t.Priority, &due, &t.BlocksStepCompletion); err != nil {
			return nil, err
		}
		if role.Valid {
			v := role.Int64
			t.AssignedRoleID = &v
		}
		if due.Valid {
			v := int(due.Int64)
			t.DueOffsetHours = &v
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
func (s *OperationsService) AddWorkflowTask(ctx context.Context, actor string, stepID int64, p WorkflowTaskPayload) (WorkflowTaskTemplate, error) {
	templateID, err := s.templateIDForStep(ctx, stepID)
	if err != nil {
		return WorkflowTaskTemplate{}, err
	}
	if err = s.ensureDraft(ctx, templateID); err != nil {
		return WorkflowTaskTemplate{}, err
	}
	if !stepTriggerTypes[p.TriggerType] || p.TitleFA == "" || (p.BlocksStepCompletion && (p.TriggerType == "ON_STEP_APPROVE" || p.TriggerType == "ON_STEP_COMPLETE")) {
		return WorkflowTaskTemplate{}, errors.New("invalid task")
	}
	if p.Priority == "" {
		p.Priority = "NORMAL"
	}
	var id int64
	err = s.db.QueryRowContext(ctx, `INSERT INTO workflow_step_task_templates(workflow_template_step_id,trigger_type,title_fa,description_fa,assigned_role_id,required_permission_code,priority,due_offset_hours,blocks_step_completion) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`, stepID, p.TriggerType, p.TitleFA, p.DescriptionFA, p.AssignedRoleID, p.RequiredPermissionCode, p.Priority, p.DueOffsetHours, p.BlocksStepCompletion).Scan(&id)
	if err != nil {
		return WorkflowTaskTemplate{}, err
	}
	s.audit(ctx, actor, "workflow_tasks.create", "workflow_step_task_template", fmt.Sprint(id), p)
	tasks, _ := s.listTemplateTasks(ctx, stepID)
	for _, t := range tasks {
		if t.ID == id {
			return t, nil
		}
	}
	return WorkflowTaskTemplate{}, sql.ErrNoRows
}
func (s *OperationsService) UpdateWorkflowTask(ctx context.Context, actor string, taskID int64, p WorkflowTaskPayload) error {
	var templateID int64
	if err := s.db.QueryRowContext(ctx, `SELECT s.workflow_template_id FROM workflow_step_task_templates t JOIN workflow_template_steps s ON s.id=t.workflow_template_step_id WHERE t.id=$1`, taskID).Scan(&templateID); err != nil {
		return err
	}
	if err := s.ensureDraft(ctx, templateID); err != nil {
		return err
	}
	if !stepTriggerTypes[p.TriggerType] || p.TitleFA == "" || (p.BlocksStepCompletion && (p.TriggerType == "ON_STEP_APPROVE" || p.TriggerType == "ON_STEP_COMPLETE")) {
		return errors.New("invalid task")
	}
	_, err := s.db.ExecContext(ctx, `UPDATE workflow_step_task_templates SET trigger_type=$2,title_fa=$3,description_fa=$4,assigned_role_id=$5,required_permission_code=$6,priority=$7,due_offset_hours=$8,blocks_step_completion=$9,updated_at=NOW() WHERE id=$1`, taskID, p.TriggerType, p.TitleFA, p.DescriptionFA, p.AssignedRoleID, p.RequiredPermissionCode, p.Priority, p.DueOffsetHours, p.BlocksStepCompletion)
	if err == nil {
		s.audit(ctx, actor, "workflow_tasks.update", "workflow_step_task_template", fmt.Sprint(taskID), p)
	}
	return err
}
func (s *OperationsService) DeleteWorkflowTask(ctx context.Context, actor string, taskID int64) error {
	var templateID int64
	if err := s.db.QueryRowContext(ctx, `SELECT s.workflow_template_id FROM workflow_step_task_templates t JOIN workflow_template_steps s ON s.id=t.workflow_template_step_id WHERE t.id=$1`, taskID).Scan(&templateID); err != nil {
		return err
	}
	if err := s.ensureDraft(ctx, templateID); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM workflow_step_task_templates WHERE id=$1`, taskID)
	if err == nil {
		s.audit(ctx, actor, "workflow_tasks.delete", "workflow_step_task_template", fmt.Sprint(taskID), nil)
	}
	return err
}

func (s *OperationsService) ListWorkflowLevelTasks(ctx context.Context, templateID int64) ([]WorkflowLevelTask, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,workflow_template_id,trigger_type,title_fa,description_fa,assigned_role_id,required_permission_code,priority,due_offset_hours,blocks_workflow_progress FROM workflow_task_templates WHERE workflow_template_id=$1 ORDER BY id`, templateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []WorkflowLevelTask{}
	for rows.Next() {
		var item WorkflowLevelTask
		var role, due sql.NullInt64
		if err := rows.Scan(&item.ID, &item.WorkflowTemplateID, &item.TriggerType, &item.TitleFA, &item.DescriptionFA, &role, &item.RequiredPermissionCode, &item.Priority, &due, &item.BlocksWorkflowProgress); err != nil {
			return nil, err
		}
		if role.Valid {
			v := role.Int64
			item.AssignedRoleID = &v
		}
		if due.Valid {
			v := int(due.Int64)
			item.DueOffsetHours = &v
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *OperationsService) AddWorkflowLevelTask(ctx context.Context, actor string, templateID int64, p WorkflowLevelTaskPayload) (WorkflowLevelTask, error) {
	if err := s.ensureDraft(ctx, templateID); err != nil {
		return WorkflowLevelTask{}, err
	}
	if p.TriggerType != "ON_WORKFLOW_START" || strings.TrimSpace(p.TitleFA) == "" || p.AssignedRoleID == nil || p.RequiredPermissionCode == "" {
		return WorkflowLevelTask{}, errors.New("invalid workflow task")
	}
	if p.Priority == "" {
		p.Priority = "NORMAL"
	}
	var id int64
	if err := s.db.QueryRowContext(ctx, `INSERT INTO workflow_task_templates(workflow_template_id,trigger_type,title_fa,description_fa,assigned_role_id,required_permission_code,priority,due_offset_hours,blocks_workflow_progress) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`, templateID, p.TriggerType, p.TitleFA, p.DescriptionFA, p.AssignedRoleID, p.RequiredPermissionCode, p.Priority, p.DueOffsetHours, p.BlocksWorkflowProgress).Scan(&id); err != nil {
		return WorkflowLevelTask{}, err
	}
	s.audit(ctx, actor, "workflow_tasks.create", "workflow_task_template", fmt.Sprint(id), p)
	items, _ := s.ListWorkflowLevelTasks(ctx, templateID)
	for _, item := range items {
		if item.ID == id {
			return item, nil
		}
	}
	return WorkflowLevelTask{}, sql.ErrNoRows
}

func (s *OperationsService) UpdateWorkflowLevelTask(ctx context.Context, actor string, id int64, p WorkflowLevelTaskPayload) error {
	var templateID int64
	if err := s.db.QueryRowContext(ctx, `SELECT workflow_template_id FROM workflow_task_templates WHERE id=$1`, id).Scan(&templateID); err != nil {
		return err
	}
	if err := s.ensureDraft(ctx, templateID); err != nil {
		return err
	}
	if p.TriggerType != "ON_WORKFLOW_START" || strings.TrimSpace(p.TitleFA) == "" || p.AssignedRoleID == nil || p.RequiredPermissionCode == "" {
		return errors.New("invalid workflow task")
	}
	_, err := s.db.ExecContext(ctx, `UPDATE workflow_task_templates SET trigger_type=$2,title_fa=$3,description_fa=$4,assigned_role_id=$5,required_permission_code=$6,priority=$7,due_offset_hours=$8,blocks_workflow_progress=$9,updated_at=NOW() WHERE id=$1`, id, p.TriggerType, p.TitleFA, p.DescriptionFA, p.AssignedRoleID, p.RequiredPermissionCode, p.Priority, p.DueOffsetHours, p.BlocksWorkflowProgress)
	if err == nil {
		s.audit(ctx, actor, "workflow_tasks.update", "workflow_task_template", fmt.Sprint(id), p)
	}
	return err
}

func (s *OperationsService) DeleteWorkflowLevelTask(ctx context.Context, actor string, id int64) error {
	var templateID int64
	if err := s.db.QueryRowContext(ctx, `SELECT workflow_template_id FROM workflow_task_templates WHERE id=$1`, id).Scan(&templateID); err != nil {
		return err
	}
	if err := s.ensureDraft(ctx, templateID); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM workflow_task_templates WHERE id=$1`, id)
	if err == nil {
		s.audit(ctx, actor, "workflow_tasks.delete", "workflow_task_template", fmt.Sprint(id), nil)
	}
	return err
}

func (s *OperationsService) ListHandoffMetrics(ctx context.Context, templateID int64) ([]HandoffMetricDefinition, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,workflow_template_id,metric_key,label_fa,unit_code,absolute_tolerance,percentage_tolerance,blocking_on_mismatch FROM workflow_handoff_metric_definitions WHERE workflow_template_id=$1 ORDER BY metric_key`, templateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []HandoffMetricDefinition{}
	for rows.Next() {
		var m HandoffMetricDefinition
		var a, p sql.NullFloat64
		if err := rows.Scan(&m.ID, &m.WorkflowTemplateID, &m.MetricKey, &m.LabelFA, &m.UnitCode, &a, &p, &m.BlockingOnMismatch); err != nil {
			return nil, err
		}
		if a.Valid {
			v := a.Float64
			m.AbsoluteTolerance = &v
		}
		if p.Valid {
			v := p.Float64
			m.PercentageTolerance = &v
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
func (s *OperationsService) AddHandoffMetric(ctx context.Context, actor string, templateID int64, p HandoffMetricPayload) (HandoffMetricDefinition, error) {
	if err := s.ensureDraft(ctx, templateID); err != nil {
		return HandoffMetricDefinition{}, err
	}
	if !codePattern.MatchString(p.MetricKey) || p.LabelFA == "" || p.UnitCode == "" {
		return HandoffMetricDefinition{}, errors.New("invalid metric")
	}
	var id int64
	err := s.db.QueryRowContext(ctx, `INSERT INTO workflow_handoff_metric_definitions(workflow_template_id,metric_key,label_fa,unit_code,absolute_tolerance,percentage_tolerance,blocking_on_mismatch) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id`, templateID, p.MetricKey, p.LabelFA, p.UnitCode, p.AbsoluteTolerance, p.PercentageTolerance, p.BlockingOnMismatch).Scan(&id)
	if err != nil {
		return HandoffMetricDefinition{}, err
	}
	s.audit(ctx, actor, "workflow_handoff_metrics.create", "workflow_handoff_metric", fmt.Sprint(id), p)
	items, _ := s.ListHandoffMetrics(ctx, templateID)
	for _, m := range items {
		if m.ID == id {
			return m, nil
		}
	}
	return HandoffMetricDefinition{}, sql.ErrNoRows
}
func (s *OperationsService) UpdateHandoffMetric(ctx context.Context, actor string, id int64, p HandoffMetricPayload) error {
	var templateID int64
	if err := s.db.QueryRowContext(ctx, `SELECT workflow_template_id FROM workflow_handoff_metric_definitions WHERE id=$1`, id).Scan(&templateID); err != nil {
		return err
	}
	if err := s.ensureDraft(ctx, templateID); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `UPDATE workflow_handoff_metric_definitions SET metric_key=$2,label_fa=$3,unit_code=$4,absolute_tolerance=$5,percentage_tolerance=$6,blocking_on_mismatch=$7,updated_at=NOW() WHERE id=$1`, id, p.MetricKey, p.LabelFA, p.UnitCode, p.AbsoluteTolerance, p.PercentageTolerance, p.BlockingOnMismatch)
	if err == nil {
		s.audit(ctx, actor, "workflow_handoff_metrics.update", "workflow_handoff_metric", fmt.Sprint(id), p)
	}
	return err
}
func (s *OperationsService) DeleteHandoffMetric(ctx context.Context, actor string, id int64) error {
	var templateID int64
	if err := s.db.QueryRowContext(ctx, `SELECT workflow_template_id FROM workflow_handoff_metric_definitions WHERE id=$1`, id).Scan(&templateID); err != nil {
		return err
	}
	if err := s.ensureDraft(ctx, templateID); err != nil {
		return err
	}
	var used bool
	_ = s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM workflow_step_field_definitions f JOIN workflow_template_steps st ON st.id=f.workflow_template_step_id WHERE st.workflow_template_id=$1 AND f.handoff_metric_key=(SELECT metric_key FROM workflow_handoff_metric_definitions WHERE id=$2))`, templateID, id).Scan(&used)
	if used {
		return errors.New("metric is used by a field")
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM workflow_handoff_metric_definitions WHERE id=$1`, id)
	if err == nil {
		s.audit(ctx, actor, "workflow_handoff_metrics.delete", "workflow_handoff_metric", fmt.Sprint(id), nil)
	}
	return err
}
func nullableJSON(raw json.RawMessage) any {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	return []byte(raw)
}
