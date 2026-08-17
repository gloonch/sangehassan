package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
)

type requestIDContextKey struct{}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDContextKey{}, requestID)
}

func RequestIDFromContext(ctx context.Context) string {
	value, _ := ctx.Value(requestIDContextKey{}).(string)
	return value
}

type ApplicationSetting struct {
	Key         string          `json:"key"`
	Value       json.RawMessage `json:"value"`
	Description string          `json:"description"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type settingRule struct {
	Kind    string
	Min     int
	Max     int
	Allowed map[string]bool
}

var applicationSettingRules = map[string]settingRule{
	"default_currency":                {Kind: "string", Allowed: map[string]bool{"IRR": true, "USD": true, "EUR": true, "AED": true, "OMR": true}},
	"default_country_code":            {Kind: "string", Allowed: map[string]bool{"IR": true}},
	"default_phone_country":           {Kind: "string", Allowed: map[string]bool{"+98": true}},
	"default_timezone":                {Kind: "timezone"},
	"customer_portal_enabled":         {Kind: "bool"},
	"sms_enabled":                     {Kind: "bool"},
	"installation_module_enabled":     {Kind: "bool"},
	"inventory_module_enabled":        {Kind: "bool"},
	"supplier_module_enabled":         {Kind: "bool"},
	"allow_manager_force_close":       {Kind: "bool"},
	"allow_manager_workflow_override": {Kind: "bool"},
	"default_payment_due_days":        {Kind: "int", Min: 1, Max: 365},
	"default_workflow_warning_hours":  {Kind: "int", Min: 1, Max: 720},
	"max_upload_size_mb":              {Kind: "int", Min: 1, Max: 100},
}

func validateSettingValue(key string, raw json.RawMessage) error {
	rule, ok := applicationSettingRules[key]
	if !ok {
		return conflict("UNKNOWN_SETTING", "تنظیم ناشناخته است")
	}
	switch rule.Kind {
	case "bool":
		var value bool
		if err := json.Unmarshal(raw, &value); err != nil {
			return conflict("VALIDATION_FAILED", "مقدار این تنظیم باید روشن یا خاموش باشد")
		}
	case "int":
		var value int
		if err := json.Unmarshal(raw, &value); err != nil || value < rule.Min || value > rule.Max {
			return conflict("VALIDATION_FAILED", fmt.Sprintf("مقدار باید بین %d و %d باشد", rule.Min, rule.Max))
		}
	case "string":
		var value string
		if err := json.Unmarshal(raw, &value); err != nil || !rule.Allowed[value] {
			return conflict("VALIDATION_FAILED", "مقدار انتخاب‌شده معتبر نیست")
		}
	case "timezone":
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return conflict("VALIDATION_FAILED", "منطقه زمانی معتبر نیست")
		}
		if _, err := time.LoadLocation(value); err != nil {
			return conflict("VALIDATION_FAILED", "منطقه زمانی معتبر نیست")
		}
	}
	return nil
}

func (s *OperationsService) ListApplicationSettings(ctx context.Context) ([]ApplicationSetting, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT setting_key,setting_value_json,description,updated_at FROM application_settings ORDER BY setting_key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ApplicationSetting{}
	for rows.Next() {
		var x ApplicationSetting
		if err = rows.Scan(&x.Key, &x.Value, &x.Description, &x.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (s *OperationsService) UpdateApplicationSetting(ctx context.Context, actor, key string, value json.RawMessage, reason string) (ApplicationSetting, error) {
	if err := validateSettingValue(key, value); err != nil {
		return ApplicationSetting{}, err
	}
	if err := requireReason(reason); err != nil {
		return ApplicationSetting{}, conflict("REASON_REQUIRED", "ثبت دلیل تغییر تنظیم الزامی است")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ApplicationSetting{}, err
	}
	defer tx.Rollback()
	var before json.RawMessage
	if err = tx.QueryRowContext(ctx, `SELECT setting_value_json FROM application_settings WHERE setting_key=$1 FOR UPDATE`, key).Scan(&before); err != nil {
		return ApplicationSetting{}, err
	}
	var out ApplicationSetting
	err = tx.QueryRowContext(ctx, `UPDATE application_settings SET setting_value_json=$2::jsonb,updated_by_user_id=$3,updated_at=NOW() WHERE setting_key=$1 RETURNING setting_key,setting_value_json,description,updated_at`, key, string(value), actor).Scan(&out.Key, &out.Value, &out.Description, &out.UpdatedAt)
	if err != nil {
		return ApplicationSetting{}, err
	}
	s.auditTx(ctx, tx, actor, "settings.update", "application_setting", key, map[string]any{"value": json.RawMessage(before)}, map[string]any{"value": json.RawMessage(value), "reason": strings.TrimSpace(reason)})
	return out, tx.Commit()
}

func (s *OperationsService) FeatureEnabled(ctx context.Context, key string) bool {
	if _, ok := applicationSettingRules[key]; !ok {
		return false
	}
	var enabled bool
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE((setting_value_json #>> '{}')::boolean,FALSE) FROM application_settings WHERE setting_key=$1`, key).Scan(&enabled); err != nil {
		return false
	}
	return enabled
}

func (s *OperationsService) FeatureFlags(ctx context.Context) map[string]bool {
	keys := []string{"customer_portal_enabled", "sms_enabled", "installation_module_enabled", "inventory_module_enabled", "supplier_module_enabled"}
	out := map[string]bool{}
	for _, key := range keys {
		out[key] = false
	}
	rows, err := s.db.QueryContext(ctx, `SELECT setting_key,(setting_value_json #>> '{}')::boolean FROM application_settings WHERE setting_key=ANY($1)`, pq.Array(keys))
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var key string
		var enabled bool
		if rows.Scan(&key, &enabled) == nil {
			out[key] = enabled
		}
	}
	return out
}

func (s *OperationsService) MaximumUploadBytes(ctx context.Context, fallback int64) int64 {
	var value int64
	if err := s.db.QueryRowContext(ctx, `SELECT (setting_value_json #>> '{}')::bigint*1024*1024 FROM application_settings WHERE setting_key='max_upload_size_mb'`).Scan(&value); err != nil || value <= 0 {
		return fallback
	}
	return value
}

type SavedView struct {
	ID        string         `json:"id"`
	ViewKey   string         `json:"view_key"`
	Name      string         `json:"name"`
	Filters   map[string]any `json:"filters"`
	UpdatedAt time.Time      `json:"updated_at"`
}

var savedViewKeys = map[string]bool{
	"orders": true, "users": true, "audit": true, "batches": true, "inventory": true, "shipments": true,
	"suppliers": true, "purchases": true, "quality": true, "installations": true,
	"payments": true, "costs": true, "notifications": true,
}

func validateSavedView(viewKey, name string, filters map[string]any) error {
	if !savedViewKeys[viewKey] || strings.TrimSpace(name) == "" || len([]rune(name)) > 80 || filters == nil {
		return conflict("VALIDATION_FAILED", "نمای ذخیره‌شده معتبر نیست")
	}
	raw, _ := json.Marshal(filters)
	if len(raw) > 8*1024 {
		return conflict("VALIDATION_FAILED", "فیلتر ذخیره‌شده بیش از حد بزرگ است")
	}
	for key, value := range filters {
		if strings.Contains(strings.ToLower(key), "password") || strings.Contains(strings.ToLower(key), "token") {
			return conflict("VALIDATION_FAILED", "ذخیره اطلاعات حساس مجاز نیست")
		}
		switch value.(type) {
		case string, bool, float64, nil, []any:
		default:
			return conflict("VALIDATION_FAILED", "ساختار فیلتر معتبر نیست")
		}
	}
	return nil
}

func (s *OperationsService) ListSavedViews(ctx context.Context, actor, viewKey string) ([]SavedView, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,view_key,name,filters_json,updated_at FROM saved_views WHERE user_id=$1 AND ($2='' OR view_key=$2) ORDER BY updated_at DESC`, actor, viewKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []SavedView{}
	for rows.Next() {
		var x SavedView
		var raw []byte
		if err = rows.Scan(&x.ID, &x.ViewKey, &x.Name, &raw, &x.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(raw, &x.Filters)
		out = append(out, x)
	}
	return out, rows.Err()
}

func (s *OperationsService) SaveSavedView(ctx context.Context, actor, id, viewKey, name string, filters map[string]any) (SavedView, error) {
	if err := validateSavedView(viewKey, name, filters); err != nil {
		return SavedView{}, err
	}
	raw, _ := json.Marshal(filters)
	var out SavedView
	var outRaw []byte
	if id == "" {
		err := s.db.QueryRowContext(ctx, `INSERT INTO saved_views(user_id,view_key,name,filters_json) VALUES($1,$2,$3,$4::jsonb) RETURNING id,view_key,name,filters_json,updated_at`, actor, viewKey, strings.TrimSpace(name), string(raw)).Scan(&out.ID, &out.ViewKey, &out.Name, &outRaw, &out.UpdatedAt)
		if err != nil {
			return out, err
		}
	} else {
		err := s.db.QueryRowContext(ctx, `UPDATE saved_views SET view_key=$3,name=$4,filters_json=$5::jsonb,updated_at=NOW() WHERE id=$1 AND user_id=$2 RETURNING id,view_key,name,filters_json,updated_at`, id, actor, viewKey, strings.TrimSpace(name), string(raw)).Scan(&out.ID, &out.ViewKey, &out.Name, &outRaw, &out.UpdatedAt)
		if err != nil {
			return out, err
		}
	}
	_ = json.Unmarshal(outRaw, &out.Filters)
	return out, nil
}

func (s *OperationsService) DeleteSavedView(ctx context.Context, actor, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM saved_views WHERE id=$1 AND user_id=$2`, id, actor)
	if err != nil {
		return err
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return ErrForbidden
	}
	return nil
}

type SearchResult struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle,omitempty"`
	Path     string `json:"path"`
}

func (s *OperationsService) GlobalSearch(ctx context.Context, actor, query string) (map[string][]SearchResult, error) {
	query = strings.TrimSpace(query)
	if len([]rune(query)) < 2 || len([]rune(query)) > 80 {
		return nil, conflict("VALIDATION_FAILED", "عبارت جست‌وجو باید بین ۲ تا ۸۰ نویسه باشد")
	}
	like := "%" + query + "%"
	out := map[string][]SearchResult{"orders": {}, "customers": {}, "proformas": {}, "shipments": {}, "batches": {}, "payments": {}}
	_, permissions, err := s.authorization(ctx, actor)
	if err != nil {
		return nil, err
	}
	permissionSet := make(map[string]bool, len(permissions))
	for _, permission := range permissions {
		permissionSet[permission] = true
	}
	has := func(code string) bool { return permissionSet[code] }
	queryRows := func(group, statement string, args ...any) error {
		rows, err := s.db.QueryContext(ctx, statement, args...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var x SearchResult
			if err = rows.Scan(&x.ID, &x.Title, &x.Subtitle, &x.Path); err != nil {
				return err
			}
			out[group] = append(out[group], x)
		}
		return rows.Err()
	}
	if has("orders.view_all") {
		if err := queryRows("orders", `SELECT id::text,order_number,COALESCE(status,''),'/panel/dashboard/orders/'||id FROM orders WHERE order_number ILIKE $1 ORDER BY created_at DESC LIMIT 8`, like); err != nil {
			return nil, err
		}
	}
	if has("customers.view") {
		if err := queryRows("customers", `SELECT id::text,COALESCE(NULLIF(CONCAT_WS(' ',first_name,last_name),''),phone_normalized),COALESCE(phone_normalized,''),'/panel/dashboard/users?search='||COALESCE(phone_normalized,'') FROM users WHERE user_type='CUSTOMER' AND CONCAT_WS(' ',first_name,last_name,phone_normalized) ILIKE $1 ORDER BY created_at DESC LIMIT 8`, like); err != nil {
			return nil, err
		}
	}
	if has("proformas.view_all") {
		if err := queryRows("proformas", `SELECT p.id::text,p.proforma_number,o.order_number,'/panel/dashboard/orders/'||p.order_id FROM proformas p JOIN orders o ON o.id=p.order_id WHERE p.proforma_number ILIKE $1 ORDER BY p.created_at DESC LIMIT 8`, like); err != nil {
			return nil, err
		}
	}
	if has("shipments.view_all") {
		if err := queryRows("shipments", `SELECT s.id::text,s.shipment_number,o.order_number,'/panel/dashboard/shipments/'||s.id FROM shipments s JOIN orders o ON o.id=s.order_id WHERE s.shipment_number ILIKE $1 ORDER BY s.created_at DESC LIMIT 8`, like); err != nil {
			return nil, err
		}
	} else if has("shipments.view_assigned") {
		if err := queryRows("shipments", `SELECT s.id::text,s.shipment_number,o.order_number,'/panel/dashboard/shipments/'||s.id FROM shipments s JOIN orders o ON o.id=s.order_id WHERE s.driver_user_id=$2 AND s.shipment_number ILIKE $1 ORDER BY s.created_at DESC LIMIT 8`, like, actor); err != nil {
			return nil, err
		}
	}
	if has("batches.view_all") {
		if err := queryRows("batches", `SELECT b.id::text,b.batch_number,o.order_number,'/panel/dashboard/batches/'||b.id FROM fulfillment_batches b JOIN orders o ON o.id=b.order_id WHERE b.batch_number ILIKE $1 ORDER BY b.created_at DESC LIMIT 8`, like); err != nil {
			return nil, err
		}
	}
	if has("finance.payments.view") || has("finance.customer_payments.view") {
		if err := queryRows("payments", `SELECT p.id::text,p.payment_number,o.order_number,'/panel/dashboard/orders/'||p.order_id FROM customer_payments p JOIN orders o ON o.id=p.order_id WHERE p.payment_number ILIKE $1 ORDER BY p.created_at DESC LIMIT 8`, like); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (s *OperationsService) DashboardHome(ctx context.Context, actor string) (map[string]any, error) {
	actions, err := s.ListActionItems(ctx, actor, false)
	if err != nil {
		return nil, err
	}
	if len(actions) > 12 {
		actions = actions[:12]
	}
	metrics, err := s.OperationsDashboardSummary(ctx, actor)
	if err != nil {
		return nil, err
	}
	roles, permissions, err := s.authorization(ctx, actor)
	if err != nil {
		return nil, err
	}
	permissionSet := make(map[string]bool, len(permissions))
	for _, permission := range permissions {
		permissionSet[permission] = true
	}
	features := s.FeatureFlags(ctx)
	quick := []map[string]string{}
	add := func(permission, feature, code, label, path string) {
		if !permissionSet[permission] {
			return
		}
		if feature != "" && !features[feature] {
			return
		}
		quick = append(quick, map[string]string{"code": code, "label": label, "path": path})
	}
	add("workflow_instances.start", "", "START_ORDER", "شروع سفارش", "/panel/dashboard")
	add("finance.payments.record", "", "RECORD_PAYMENT", "ثبت پرداخت", "/panel/dashboard/finance")
	add("shipments.view_assigned", "", "SHIPMENT_TODAY", "حمل و تحویل امروز", "/panel/dashboard/shipments")
	add("inventory.receipts.create", "inventory_module_enabled", "INVENTORY_RECEIPT", "ثبت دریافت موجودی", "/panel/dashboard/inventory")
	add("purchases.receive", "supplier_module_enabled", "PURCHASE_RECEIPT", "دریافت خرید", "/panel/dashboard/purchases")
	add("installation.view_assigned", "installation_module_enabled", "INSTALLATION_TODAY", "نصب‌های امروز", "/panel/dashboard/installations")
	return map[string]any{"roles": roles, "my_actions": actions, "alerts": metrics, "quick_actions": quick}, nil
}

func (s *OperationsService) WorkflowDiagnostics(ctx context.Context, workflowID string) (map[string]any, error) {
	var status string
	var current sql.NullString
	if err := s.db.QueryRowContext(ctx, `SELECT status,current_step_instance_id FROM workflow_instances WHERE id=$1`, workflowID).Scan(&status, &current); err != nil {
		return nil, err
	}
	findings := []map[string]any{}
	var active int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM workflow_step_instances WHERE workflow_instance_id=$1 AND status IN ('WAITING_FOR_ASSIGNEE','IN_PROGRESS','WAITING_FOR_APPROVAL','HAS_MISMATCH','NEEDS_CORRECTION','BLOCKED','WAITING_FOR_TRANSITION')`, workflowID).Scan(&active); err != nil {
		return nil, err
	}
	if status == "IN_PROGRESS" && !current.Valid {
		findings = append(findings, map[string]any{"code": "CURRENT_STEP_MISSING", "severity": "CRITICAL", "safe_repair": active == 1})
	}
	if active > 1 {
		findings = append(findings, map[string]any{"code": "MULTIPLE_ACTIVE_STEPS", "severity": "CRITICAL", "safe_repair": false})
	}
	if current.Valid {
		var open int
		if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM action_items WHERE workflow_step_instance_id=$1 AND status NOT IN ('COMPLETED','CANCELLED')`, current.String).Scan(&open); err != nil {
			return nil, err
		}
		if open == 0 {
			findings = append(findings, map[string]any{"code": "OPEN_ACTION_MISSING", "severity": "WARNING", "safe_repair": true})
		}
	}
	return map[string]any{"workflow_instance_id": workflowID, "status": status, "current_step_instance_id": scanNullableString(current), "active_step_count": active, "findings": findings}, nil
}

func (s *OperationsService) RepairWorkflow(ctx context.Context, actor, workflowID, repairCode, reason, key string) (map[string]any, error) {
	if err := requireReason(reason); err != nil {
		return nil, conflict("REASON_REQUIRED", "ثبت دلیل اصلاح الزامی است")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, "diagnostics."+repairCode, key, map[string]string{"workflow_id": workflowID, "repair_code": repairCode, "reason": reason})
	if err != nil {
		return nil, err
	}
	if claim.Existing {
		var old map[string]any
		_ = json.Unmarshal(claim.Response, &old)
		return old, nil
	}
	result := map[string]any{"workflow_instance_id": workflowID, "repair_code": repairCode, "repaired": true}
	switch repairCode {
	case "SET_SINGLE_CURRENT_STEP":
		var stepID string
		err = tx.QueryRowContext(ctx, `SELECT id FROM workflow_step_instances WHERE workflow_instance_id=$1 AND status IN ('WAITING_FOR_ASSIGNEE','IN_PROGRESS','WAITING_FOR_APPROVAL','HAS_MISMATCH','NEEDS_CORRECTION','BLOCKED','WAITING_FOR_TRANSITION') ORDER BY sequence_number,iteration_number DESC LIMIT 2`, workflowID).Scan(&stepID)
		if err != nil {
			return nil, err
		}
		var count int
		if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM workflow_step_instances WHERE workflow_instance_id=$1 AND status IN ('WAITING_FOR_ASSIGNEE','IN_PROGRESS','WAITING_FOR_APPROVAL','HAS_MISMATCH','NEEDS_CORRECTION','BLOCKED','WAITING_FOR_TRANSITION')`, workflowID).Scan(&count); err != nil || count != 1 {
			return nil, conflict("UNSAFE_REPAIR", "این مورد نیازمند بررسی دستی است")
		}
		_, err = tx.ExecContext(ctx, `UPDATE workflow_instances SET current_step_instance_id=$2,updated_at=NOW() WHERE id=$1`, workflowID, stepID)
	case "REBUILD_CURRENT_ACTION":
		_, err = tx.ExecContext(ctx, `INSERT INTO action_items(workflow_instance_id,workflow_step_instance_id,order_id,customer_user_id,title_fa,description_fa,status,priority,assigned_role_id,assigned_user_id,required_permission_code,due_at,deduplication_key,source_trigger_type) SELECT wi.id,si.id,wi.order_id,wi.customer_user_id,si.internal_title_fa,COALESCE(si.internal_description_fa,''),'OPEN','NORMAL',si.responsible_role_id,si.assigned_user_id,si.required_permission_code,si.estimated_end_at,'repair:step:'||si.id,'ADMIN_REPAIR' FROM workflow_instances wi JOIN workflow_step_instances si ON si.id=wi.current_step_instance_id WHERE wi.id=$1 ON CONFLICT(deduplication_key) WHERE deduplication_key IS NOT NULL DO NOTHING`, workflowID)
	default:
		return nil, conflict("UNSAFE_REPAIR", "Repair انتخاب‌شده مجاز نیست")
	}
	if err != nil {
		return nil, err
	}
	s.auditTx(ctx, tx, actor, "diagnostics.repair", "workflow_instance", workflowID, nil, map[string]any{"repair_code": repairCode, "reason": reason})
	if err = finishOperationTx(ctx, tx, actor, "diagnostics."+repairCode, key, result); err != nil {
		return nil, err
	}
	return result, tx.Commit()
}

func (s *OperationsService) RevokeUserSessions(ctx context.Context, actor, userID, reason, key string) (map[string]any, error) {
	if err := requireReason(reason); err != nil {
		return nil, conflict("REASON_REQUIRED", "ثبت دلیل خروج اجباری الزامی است")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, "admin.revoke_sessions", key, map[string]string{"user_id": userID, "reason": reason})
	if err != nil {
		return nil, err
	}
	if claim.Existing {
		var out map[string]any
		_ = json.Unmarshal(claim.Response, &out)
		return out, nil
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE user_id=$1`, userID); err != nil {
		return nil, err
	}
	result, err := tx.ExecContext(ctx, `UPDATE users SET auth_invalid_before=NOW(),updated_at=NOW() WHERE id=$1`, userID)
	if err != nil {
		return nil, err
	}
	if n, _ := result.RowsAffected(); n != 1 {
		return nil, ErrUserNotFound
	}
	out := map[string]any{"user_id": userID, "revoked": true}
	s.auditTx(ctx, tx, actor, "users.sessions.revoke", "user", userID, nil, map[string]any{"reason": reason})
	if err = finishOperationTx(ctx, tx, actor, "admin.revoke_sessions", key, out); err != nil {
		return nil, err
	}
	return out, tx.Commit()
}

func (s *OperationsService) RecalculateOrderProgress(ctx context.Context, actor, orderID, reason, key string) (map[string]any, error) {
	if err := requireReason(reason); err != nil {
		return nil, conflict("REASON_REQUIRED", "ثبت دلیل اصلاح الزامی است")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, "admin.recalculate_progress", key, map[string]string{"order_id": orderID, "reason": reason})
	if err != nil {
		return nil, err
	}
	if claim.Existing {
		var out map[string]any
		_ = json.Unmarshal(claim.Response, &out)
		return out, nil
	}
	if _, err = tx.ExecContext(ctx, `SELECT id FROM orders WHERE id=$1 FOR UPDATE`, orderID); err != nil {
		return nil, err
	}
	progress, err := s.OrderProgress(ctx, actor, orderID, false)
	if err != nil {
		return nil, err
	}
	out := map[string]any{"order_id": orderID, "progress": progress, "recalculated": true}
	s.auditTx(ctx, tx, actor, "admin_tools.recalculate_progress", "order", orderID, nil, map[string]any{"reason": reason, "overall_progress": progress.OverallProgress})
	if err = finishOperationTx(ctx, tx, actor, "admin.recalculate_progress", key, out); err != nil {
		return nil, err
	}
	return out, tx.Commit()
}

func (s *OperationsService) AccessAllowed(ctx context.Context, userID string, issuedAt time.Time) bool {
	var ok bool
	err := s.db.QueryRowContext(ctx, `SELECT status='ACTIVE' AND is_active AND (auth_invalid_before IS NULL OR auth_invalid_before<=$2) FROM users WHERE id=$1`, userID, issuedAt).Scan(&ok)
	return err == nil && ok
}

func (s *OperationsService) SystemInfo(ctx context.Context, version, commit, buildTime, environment string) (map[string]any, error) {
	var migration int
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(version),0) FROM schema_migrations`).Scan(&migration); err != nil {
		return nil, err
	}
	return map[string]any{"version": version, "commit": commit, "build_time": buildTime, "environment": environment, "migration": migration}, nil
}

func (s *OperationsService) Ready(ctx context.Context) error {
	readyCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := s.db.PingContext(readyCtx); err != nil {
		return err
	}
	var exists bool
	if err := s.db.QueryRowContext(readyCtx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=20)`).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return errors.New("database migration 020 is required")
	}
	return nil
}

type PageRequest struct{ Page, PageSize int }

func ParsePage(pageRaw, sizeRaw string) PageRequest {
	page, _ := strconv.Atoi(pageRaw)
	size, _ := strconv.Atoi(sizeRaw)
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 25
	}
	if size > 100 {
		size = 100
	}
	return PageRequest{Page: page, PageSize: size}
}

func PageResult(items any, page PageRequest, total int) map[string]any {
	return map[string]any{"items": items, "page": page.Page, "pageSize": page.PageSize, "total": total}
}

func (s *OperationsService) ExportRows(ctx context.Context, kind, search, status string, limit int) ([]string, [][]string, error) {
	if limit <= 0 || limit > 10000 {
		limit = 10000
	}
	like := "%" + strings.TrimSpace(search) + "%"
	var headers []string
	var query string
	args := []any{like, status, limit}
	switch kind {
	case "orders":
		headers = []string{"order_number", "customer_name", "customer_phone", "status", "created_at"}
		query = `SELECT o.order_number,COALESCE(NULLIF(CONCAT_WS(' ',u.first_name,u.last_name),''),''),COALESCE(u.phone_normalized,''),o.status,o.created_at::text FROM orders o JOIN users u ON u.id=o.customer_user_id WHERE ($1='%%' OR o.order_number ILIKE $1 OR CONCAT_WS(' ',u.first_name,u.last_name,u.phone_normalized) ILIKE $1) AND ($2='' OR o.status=$2) ORDER BY o.created_at DESC LIMIT $3`
	case "customers":
		headers = []string{"customer_id", "name", "phone", "status", "created_at"}
		query = `SELECT u.id::text,COALESCE(NULLIF(CONCAT_WS(' ',u.first_name,u.last_name),''),''),COALESCE(u.phone_normalized,''),u.status,u.created_at::text FROM users u WHERE u.user_type='CUSTOMER' AND ($1='%%' OR CONCAT_WS(' ',u.first_name,u.last_name,u.phone_normalized) ILIKE $1) AND ($2='' OR u.status=$2) ORDER BY u.created_at DESC LIMIT $3`
	case "payments":
		headers = []string{"payment_number", "order_number", "amount", "currency", "status", "paid_at"}
		query = `SELECT p.payment_number,o.order_number,p.amount::text,p.currency,p.status,COALESCE(p.paid_at::text,'') FROM customer_payments p JOIN orders o ON o.id=p.order_id WHERE ($1='%%' OR p.payment_number ILIKE $1 OR o.order_number ILIKE $1) AND ($2='' OR p.status=$2) ORDER BY p.created_at DESC LIMIT $3`
	case "costs":
		headers = []string{"cost_id", "order_number", "cost_type", "amount", "currency", "status", "created_at"}
		query = `SELECT c.id::text,COALESCE(o.order_number,''),c.cost_type,c.amount::text,c.currency,c.status,c.created_at::text FROM operational_cost_entries c LEFT JOIN orders o ON o.id=c.order_id WHERE ($1='%%' OR c.id::text ILIKE $1 OR o.order_number ILIKE $1) AND ($2='' OR c.status=$2) ORDER BY c.created_at DESC LIMIT $3`
	case "shipments":
		headers = []string{"shipment_number", "order_number", "status", "planned_departure_at", "estimated_arrival_at"}
		query = `SELECT s.shipment_number,o.order_number,s.status,COALESCE(s.planned_departure_at::text,''),COALESCE(s.estimated_arrival_at::text,'') FROM shipments s JOIN orders o ON o.id=s.order_id WHERE ($1='%%' OR s.shipment_number ILIKE $1 OR o.order_number ILIKE $1) AND ($2='' OR s.status=$2) ORDER BY s.created_at DESC LIMIT $3`
	default:
		return nil, nil, conflict("VALIDATION_FAILED", "نوع خروجی معتبر نیست")
	}
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	result := [][]string{}
	for rows.Next() {
		values := make([]sql.NullString, len(headers))
		targets := make([]any, len(headers))
		for i := range values {
			targets[i] = &values[i]
		}
		if err = rows.Scan(targets...); err != nil {
			return nil, nil, err
		}
		row := make([]string, len(headers))
		for i := range values {
			row[i] = values[i].String
		}
		result = append(result, row)
	}
	return headers, result, rows.Err()
}

func (s *OperationsService) DetectIntegrityFindings(ctx context.Context) (int, error) {
	checks := []struct{ code, entity, query, summary, repair string }{
		{"WORKFLOW_CURRENT_STEP_MISSING", "WORKFLOW", `SELECT id::text FROM workflow_instances WHERE status='IN_PROGRESS' AND current_step_instance_id IS NULL`, `Workflow فعال مرحله جاری ندارد`, `SET_SINGLE_CURRENT_STEP`},
		{"ACTION_ITEM_MISSING", "WORKFLOW", `SELECT wi.id::text FROM workflow_instances wi JOIN workflow_step_instances si ON si.id=wi.current_step_instance_id WHERE wi.status='IN_PROGRESS' AND NOT EXISTS(SELECT 1 FROM action_items a WHERE a.workflow_step_instance_id=si.id AND a.status NOT IN ('COMPLETED','CANCELLED'))`, `مرحله جاری Action Item باز ندارد`, `REBUILD_CURRENT_ACTION`},
		{"ORDER_PROGRESS_OVER_DELIVERY", "ORDER", `SELECT DISTINCT oi.order_id::text FROM order_items oi WHERE COALESCE((SELECT SUM(si.delivered_quantity) FROM fulfillment_batches b JOIN shipment_items si ON si.batch_id=b.id WHERE b.order_item_id=oi.id AND si.quantity_unit=oi.quantity_unit),0)>oi.ordered_quantity`, `مقدار تحویل‌شده از مقدار Line سفارش بیشتر است`, ``},
		{"PAYMENT_BALANCE_MISMATCH", "ORDER", `SELECT fs.order_id::text FROM order_financial_summaries fs JOIN order_commercial_terms t ON t.order_id=fs.order_id JOIN orders o ON o.id=fs.order_id WHERE ABS(fs.confirmed_payment_amount-COALESCE((SELECT SUM(p.amount) FROM customer_payments p WHERE p.order_id=fs.order_id AND p.currency=t.currency AND p.status IN ('CONFIRMED','PARTIALLY_REFUNDED','REFUNDED')),0))>0.0001 OR ABS(fs.refunded_amount-COALESCE((SELECT SUM(r.amount) FROM payment_refunds r JOIN customer_payments p ON p.id=r.payment_id WHERE p.order_id=fs.order_id AND r.currency=t.currency),0))>0.0001 OR ABS(fs.outstanding_amount-GREATEST(0,CASE WHEN o.status IN ('CONFIRMED','IN_PROGRESS','COMPLETED','CLOSED') THEN t.final_customer_amount ELSE 0 END-COALESCE((SELECT SUM(p.amount) FROM customer_payments p WHERE p.order_id=fs.order_id AND p.currency=t.currency AND p.status IN ('CONFIRMED','PARTIALLY_REFUNDED','REFUNDED')),0)+COALESCE((SELECT SUM(r.amount) FROM payment_refunds r JOIN customer_payments p ON p.id=r.payment_id WHERE p.order_id=fs.order_id AND r.currency=t.currency),0)))>0.0001`, `مانده مالی Order با داده‌های مرجع تطابق ندارد`, `RECONCILE_PAYMENT`},
	}
	count := 0
	for _, check := range checks {
		detected := []string{}
		rows, err := s.db.QueryContext(ctx, check.query)
		if err != nil {
			return count, err
		}
		for rows.Next() {
			var id string
			if err = rows.Scan(&id); err != nil {
				rows.Close()
				return count, err
			}
			key := check.code + ":" + id
			detected = append(detected, id)
			_, err = s.db.ExecContext(ctx, `INSERT INTO reconciliation_findings(finding_key,check_code,entity_type,entity_id,severity,summary,safe_repair_code) VALUES($1,$2,$3,$4,'WARNING',$5,$6) ON CONFLICT(finding_key) DO UPDATE SET last_detected_at=NOW(),status=CASE WHEN reconciliation_findings.status='OPEN' THEN 'OPEN' ELSE reconciliation_findings.status END`, key, check.code, check.entity, id, check.summary, check.repair)
			if err != nil {
				rows.Close()
				return count, err
			}
			count++
		}
		rows.Close()
		if _, err = s.db.ExecContext(ctx, `UPDATE reconciliation_findings SET status='RESOLVED',resolved_at=NOW(),resolution_reason='integrity check passed' WHERE check_code=$1 AND status='OPEN' AND NOT(entity_id=ANY($2))`, check.code, pq.Array(detected)); err != nil {
			return count, err
		}
	}
	return count, nil
}

func (s *OperationsService) AuditLogsPage(ctx context.Context, search, actor, entity, action, orderID, from, to string, page PageRequest) (map[string]any, error) {
	where := ` WHERE ($1='' OR a.action_code ILIKE '%'||$1||'%' OR a.entity_type ILIKE '%'||$1||'%' OR COALESCE(a.entity_id,'') ILIKE '%'||$1||'%') AND ($2='' OR a.actor_user_id::text=$2) AND ($3='' OR a.entity_type=$3) AND ($4='' OR a.action_code=$4) AND ($5='' OR a.entity_id=$5 OR a.metadata->>'order_id'=$5 OR a.before_data->>'order_id'=$5 OR a.after_data->>'order_id'=$5) AND ($6='' OR a.created_at >= $6::timestamptz) AND ($7='' OR a.created_at < $7::timestamptz + INTERVAL '1 day')`
	args := []any{search, actor, entity, action, orderID, from, to}
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM audit_logs a`+where, args...).Scan(&total); err != nil {
		return nil, err
	}
	args = append(args, page.PageSize, (page.Page-1)*page.PageSize)
	rows, err := s.db.QueryContext(ctx, `SELECT a.id,COALESCE(NULLIF(CONCAT_WS(' ',u.first_name,u.last_name),''),u.phone_normalized,'سیستم'),COALESCE(a.actor_user_id::text,''),a.action_code,a.entity_type,COALESCE(a.entity_id,''),a.before_data,a.after_data,a.metadata,COALESCE(a.request_id,''),a.created_at FROM audit_logs a LEFT JOIN users u ON u.id=a.actor_user_id`+where+` ORDER BY a.created_at DESC LIMIT $8 OFFSET $9`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id int64
		var actorName, actorID, actionCode, entityType, entityID, requestID string
		var before, after, metadata []byte
		var created time.Time
		if err = rows.Scan(&id, &actorName, &actorID, &actionCode, &entityType, &entityID, &before, &after, &metadata, &requestID, &created); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{"id": id, "actor": actorName, "actor_id": actorID, "action": actionCode, "entity": entityType, "entity_id": entityID, "before": json.RawMessage(before), "after": json.RawMessage(after), "metadata": json.RawMessage(metadata), "request_id": requestID, "created_at": created})
	}
	return PageResult(items, page, total), rows.Err()
}

func (s *OperationsService) CorrectOrderEstimatedDelivery(ctx context.Context, actor, orderID string, estimated *time.Time, reason, key string) (map[string]any, error) {
	if err := requireReason(reason); err != nil {
		return nil, conflict("REASON_REQUIRED", "ثبت دلیل اصلاح الزامی است")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	payload := map[string]any{"order_id": orderID, "estimated_delivery_at": estimated, "reason": reason}
	claim, err := claimOperationTx(ctx, tx, actor, "admin.correct_estimated_delivery", key, payload)
	if err != nil {
		return nil, err
	}
	if claim.Existing {
		var out map[string]any
		_ = json.Unmarshal(claim.Response, &out)
		return out, nil
	}
	var before sql.NullTime
	if err = tx.QueryRowContext(ctx, `SELECT estimated_delivery_at FROM orders WHERE id=$1 FOR UPDATE`, orderID).Scan(&before); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE orders SET estimated_delivery_at=$2,updated_at=NOW() WHERE id=$1`, orderID, estimated); err != nil {
		return nil, err
	}
	out := map[string]any{"order_id": orderID, "estimated_delivery_at": estimated}
	s.auditTx(ctx, tx, actor, "admin_tools.correct_estimated_delivery", "order", orderID, map[string]any{"estimated_delivery_at": readinessNullableTime(before)}, map[string]any{"estimated_delivery_at": estimated, "reason": reason})
	if err = finishOperationTx(ctx, tx, actor, "admin.correct_estimated_delivery", key, out); err != nil {
		return nil, err
	}
	return out, tx.Commit()
}

func readinessNullableTime(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	v := value.Time
	return &v
}

func (s *OperationsService) ResolveStuckAction(ctx context.Context, actor, actionID, reason, key string) (map[string]any, error) {
	if err := requireReason(reason); err != nil {
		return nil, conflict("REASON_REQUIRED", "ثبت دلیل اصلاح الزامی است")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	payload := map[string]string{"action_id": actionID, "reason": reason}
	claim, err := claimOperationTx(ctx, tx, actor, "admin.resolve_action", key, payload)
	if err != nil {
		return nil, err
	}
	if claim.Existing {
		var out map[string]any
		_ = json.Unmarshal(claim.Response, &out)
		return out, nil
	}
	var status string
	if err = tx.QueryRowContext(ctx, `SELECT status FROM action_items WHERE id=$1 FOR UPDATE`, actionID).Scan(&status); err != nil {
		return nil, err
	}
	if status == "COMPLETED" || status == "CANCELLED" {
		return nil, conflict("INVALID_ACTION_STATE", "این اقدام قبلاً بسته شده است")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE action_items SET status='COMPLETED',completed_at=NOW(),completed_by_user_id=$2,updated_at=NOW() WHERE id=$1`, actionID, actor); err != nil {
		return nil, err
	}
	out := map[string]any{"action_item_id": actionID, "status": "COMPLETED"}
	s.auditTx(ctx, tx, actor, "admin_tools.resolve_action", "action_item", actionID, map[string]any{"status": status}, map[string]any{"status": "COMPLETED", "reason": reason})
	if err = finishOperationTx(ctx, tx, actor, "admin.resolve_action", key, out); err != nil {
		return nil, err
	}
	return out, tx.Commit()
}

func (s *OperationsService) ReconcileOrderPayment(ctx context.Context, actor, orderID, reason, key string) (map[string]any, error) {
	if err := requireReason(reason); err != nil {
		return nil, conflict("REASON_REQUIRED", "ثبت دلیل اصلاح الزامی است")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	payload := map[string]string{"order_id": orderID, "reason": reason}
	claim, err := claimOperationTx(ctx, tx, actor, "admin.reconcile_payment", key, payload)
	if err != nil {
		return nil, err
	}
	if claim.Existing {
		var out map[string]any
		_ = json.Unmarshal(claim.Response, &out)
		return out, nil
	}
	if _, err = tx.ExecContext(ctx, `SELECT id FROM orders WHERE id=$1 FOR UPDATE`, orderID); err != nil {
		return nil, err
	}
	if err = refreshFinancialSummaryTx(ctx, tx, orderID); err != nil {
		return nil, err
	}
	out := map[string]any{"order_id": orderID, "reconciled": true}
	s.auditTx(ctx, tx, actor, "admin_tools.reconcile_payment", "order", orderID, nil, map[string]any{"reason": reason})
	if err = finishOperationTx(ctx, tx, actor, "admin.reconcile_payment", key, out); err != nil {
		return nil, err
	}
	return out, tx.Commit()
}
