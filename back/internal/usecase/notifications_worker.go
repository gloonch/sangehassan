package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

type SMSProvider interface {
	SendMessage(ctx context.Context, recipient, message string) (string, error)
	GetDeliveryStatus(ctx context.Context, providerMessageID string) (string, error)
	Name() string
}

var ErrSMSProviderDisabled = errors.New("sms provider disabled")

type DisabledSMSProvider struct{}

func (DisabledSMSProvider) SendMessage(context.Context, string, string) (string, error) {
	return "", ErrSMSProviderDisabled
}
func (DisabledSMSProvider) GetDeliveryStatus(context.Context, string) (string, error) {
	return "CANCELLED", ErrSMSProviderDisabled
}
func (DisabledSMSProvider) Name() string { return "disabled" }

type FakeSMSProvider struct {
	mu       sync.Mutex
	Messages []FakeSMSMessage
	Failures int
}
type FakeSMSMessage struct {
	Recipient  string
	Body       string
	ProviderID string
}

func (p *FakeSMSProvider) SendMessage(_ context.Context, recipient, message string) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Failures > 0 {
		p.Failures--
		return "", errors.New("fake provider failure")
	}
	id := fmt.Sprintf("fake-%d", len(p.Messages)+1)
	p.Messages = append(p.Messages, FakeSMSMessage{recipient, message, id})
	return id, nil
}
func (p *FakeSMSProvider) GetDeliveryStatus(context.Context, string) (string, error) {
	return "DELIVERED", nil
}
func (*FakeSMSProvider) Name() string { return "fake" }

var templateVariablePattern = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_]+)\s*\}\}`)

func renderNotificationTemplate(template string, allowed []string, values map[string]string) (string, error) {
	allowedSet := map[string]bool{}
	for _, x := range allowed {
		allowedSet[x] = true
	}
	var renderErr error
	out := templateVariablePattern.ReplaceAllStringFunc(template, func(token string) string {
		m := templateVariablePattern.FindStringSubmatch(token)
		key := m[1]
		if !allowedSet[key] {
			renderErr = fmt.Errorf("template variable %s is not allowed", key)
			return token
		}
		value, ok := values[key]
		if !ok || strings.TrimSpace(value) == "" {
			renderErr = fmt.Errorf("template variable %s is unresolved", key)
			return token
		}
		return value
	})
	if renderErr != nil {
		return "", renderErr
	}
	if strings.Contains(out, "{{") || strings.Contains(out, "}}") {
		return "", errors.New("unresolved template expression")
	}
	return out, nil
}

func emitNotificationTx(ctx context.Context, tx *sql.Tx, userID, eventType, eventKey, entityType, entityID, deepLink string, values map[string]string) error {
	if _, err := tx.ExecContext(ctx, `SAVEPOINT notification_side_effect`); err != nil {
		return err
	}
	if err := emitNotificationUnsafeTx(ctx, tx, userID, eventType, eventKey, entityType, entityID, deepLink, values); err != nil {
		_, rollbackErr := tx.ExecContext(ctx, `ROLLBACK TO SAVEPOINT notification_side_effect`)
		_, releaseErr := tx.ExecContext(ctx, `RELEASE SAVEPOINT notification_side_effect`)
		slog.WarnContext(ctx, "notification_side_effect_skipped", "eventType", eventType, "entityType", entityType, "entityId", entityID, "error", err)
		if rollbackErr != nil {
			return rollbackErr
		}
		return releaseErr
	}
	_, err := tx.ExecContext(ctx, `RELEASE SAVEPOINT notification_side_effect`)
	return err
}

func emitNotificationUnsafeTx(ctx context.Context, tx *sql.Tx, userID, eventType, eventKey, entityType, entityID, deepLink string, values map[string]string) error {
	if strings.Contains(deepLink, "://") || (!strings.HasPrefix(deepLink, "/panel/dashboard") && !strings.HasPrefix(deepLink, "/account") && deepLink != "") {
		return errors.New("unsafe notification deep link")
	}
	rows, err := tx.QueryContext(ctx, `SELECT channel,title_template,body_template,allowed_variables FROM notification_templates WHERE event_type=$1 AND locale='fa' AND is_active=TRUE ORDER BY channel`, eventType)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var channel, titleTemplate, bodyTemplate string
		var raw []byte
		if err = rows.Scan(&channel, &titleTemplate, &bodyTemplate, &raw); err != nil {
			return err
		}
		var allowed []string
		if err = json.Unmarshal(raw, &allowed); err != nil {
			return err
		}
		title, err := renderNotificationTemplate(titleTemplate, allowed, values)
		if err != nil {
			return err
		}
		body, err := renderNotificationTemplate(bodyTemplate, allowed, values)
		if err != nil {
			return err
		}
		var inApp, sms bool
		err = tx.QueryRowContext(ctx, `SELECT COALESCE((SELECT in_app_enabled FROM notification_preferences WHERE user_id=$1 AND event_type=$2),TRUE),COALESCE((SELECT sms_enabled FROM notification_preferences WHERE user_id=$1 AND event_type=$2),FALSE)`, userID, eventType).Scan(&inApp, &sms)
		if err != nil {
			return err
		}
		if eventType == "CUSTOMER_ACCOUNT_CREATED" || eventType == "CUSTOMER_ACCOUNT_ACTIVATED" {
			inApp, sms = true, true
		}
		if channel == "IN_APP" && inApp {
			data, _ := json.Marshal(values)
			_, err = tx.ExecContext(ctx, `INSERT INTO notifications(user_id,type,payload,event_type,event_key,title,body,entity_type,entity_id,deep_link,data_json) VALUES($1,$2,$9::jsonb,$2,$3,$4,$5,NULLIF($6,''),NULLIF($7,'')::uuid,NULLIF($8,''),$9::jsonb) ON CONFLICT(event_key) DO NOTHING`, userID, eventType, userID+":"+eventKey, title, body, entityType, entityID, deepLink, string(data))
			if err != nil {
				return err
			}
		}
		if channel == "SMS" && sms {
			var phone string
			if err = tx.QueryRowContext(ctx, `SELECT phone_normalized FROM users WHERE id=$1`, userID).Scan(&phone); err != nil {
				return err
			}
			_, err = tx.ExecContext(ctx, `INSERT INTO notification_outbox(notification_id,user_id,channel,event_key,recipient,message_body) VALUES((SELECT id FROM notifications WHERE event_key=$2),$1,'SMS',$3,$4,$5) ON CONFLICT(event_key) DO NOTHING`, userID, userID+":"+eventKey, userID+":"+eventKey+":sms", phone, body)
			if err != nil {
				return err
			}
		}
	}
	return rows.Err()
}

func emitNotificationToRoleTx(ctx context.Context, tx *sql.Tx, roleCode, eventType, eventKey, entityType, entityID, deepLink string, values map[string]string) error {
	if _, err := tx.ExecContext(ctx, `SAVEPOINT notification_role_side_effect`); err != nil {
		return err
	}
	if err := emitNotificationToRoleUnsafeTx(ctx, tx, roleCode, eventType, eventKey, entityType, entityID, deepLink, values); err != nil {
		_, rollbackErr := tx.ExecContext(ctx, `ROLLBACK TO SAVEPOINT notification_role_side_effect`)
		_, releaseErr := tx.ExecContext(ctx, `RELEASE SAVEPOINT notification_role_side_effect`)
		slog.WarnContext(ctx, "role_notification_side_effect_skipped", "role", roleCode, "eventType", eventType, "entityType", entityType, "entityId", entityID, "error", err)
		if rollbackErr != nil {
			return rollbackErr
		}
		return releaseErr
	}
	_, err := tx.ExecContext(ctx, `RELEASE SAVEPOINT notification_role_side_effect`)
	return err
}

func emitNotificationToRoleUnsafeTx(ctx context.Context, tx *sql.Tx, roleCode, eventType, eventKey, entityType, entityID, deepLink string, values map[string]string) error {
	rows, err := tx.QueryContext(ctx, `SELECT DISTINCT u.id FROM users u JOIN user_roles ur ON ur.user_id=u.id JOIN roles r ON r.id=ur.role_id WHERE r.code=$1 AND r.is_active AND u.status='ACTIVE'`, roleCode)
	if err != nil {
		return err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	if err = rows.Close(); err != nil {
		return err
	}
	for _, id := range ids {
		if err = emitNotificationTx(ctx, tx, id, eventType, eventKey, entityType, entityID, deepLink, values); err != nil {
			return err
		}
	}
	return nil
}

func emitShipmentCustomerNotificationTx(ctx context.Context, tx *sql.Tx, shipmentID, eventType string) error {
	if _, err := tx.ExecContext(ctx, `SAVEPOINT notification_shipment_side_effect`); err != nil {
		return err
	}
	var customer string
	if err := tx.QueryRowContext(ctx, `SELECT o.customer_user_id FROM shipments s JOIN orders o ON o.id=s.order_id WHERE s.id=$1`, shipmentID).Scan(&customer); err != nil {
		_, rollbackErr := tx.ExecContext(ctx, `ROLLBACK TO SAVEPOINT notification_shipment_side_effect`)
		_, releaseErr := tx.ExecContext(ctx, `RELEASE SAVEPOINT notification_shipment_side_effect`)
		slog.WarnContext(ctx, "shipment_notification_side_effect_skipped", "eventType", eventType, "shipmentId", shipmentID, "error", err)
		if rollbackErr != nil {
			return rollbackErr
		}
		return releaseErr
	}
	if err := emitNotificationTx(ctx, tx, customer, eventType, strings.ToLower(eventType)+":"+shipmentID, "SHIPMENT", shipmentID, "/account", map[string]string{}); err != nil {
		_, _ = tx.ExecContext(ctx, `ROLLBACK TO SAVEPOINT notification_shipment_side_effect`)
		_, _ = tx.ExecContext(ctx, `RELEASE SAVEPOINT notification_shipment_side_effect`)
		return err
	}
	_, err := tx.ExecContext(ctx, `RELEASE SAVEPOINT notification_shipment_side_effect`)
	return err
}

func (s *OperationsService) ListNotifications(ctx context.Context, userID string, limit, offset int) (map[string]any, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,event_type,title,body,entity_type,entity_id::text,deep_link,data_json,status,created_at FROM notifications WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Notification{}
	for rows.Next() {
		var n Notification
		var et, eid, link sql.NullString
		var raw []byte
		if err = rows.Scan(&n.ID, &n.EventType, &n.Title, &n.Body, &et, &eid, &link, &raw, &n.Status, &n.CreatedAt); err != nil {
			return nil, err
		}
		n.EntityType = scanNullableString(et)
		n.EntityID = scanNullableString(eid)
		n.DeepLink = scanNullableString(link)
		_ = json.Unmarshal(raw, &n.Data)
		items = append(items, n)
	}
	var unread int
	if err = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id=$1 AND status='UNREAD'`, userID).Scan(&unread); err != nil {
		return nil, err
	}
	return map[string]any{"items": items, "unread_count": unread}, rows.Err()
}
func (s *OperationsService) ReadNotification(ctx context.Context, userID, id string) error {
	r, err := s.db.ExecContext(ctx, `UPDATE notifications SET status='READ',read_at=COALESCE(read_at,NOW()) WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return err
	}
	if n, _ := r.RowsAffected(); n == 0 {
		return ErrForbidden
	}
	return nil
}
func (s *OperationsService) ReadAllNotifications(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE notifications SET status='READ',read_at=COALESCE(read_at,NOW()) WHERE user_id=$1 AND status='UNREAD'`, userID)
	return err
}
func (s *OperationsService) ListNotificationPreferences(ctx context.Context, userID string) ([]NotificationPreferencePayload, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT t.event_type,COALESCE(p.in_app_enabled,TRUE),COALESCE(p.sms_enabled,FALSE) FROM (SELECT DISTINCT event_type FROM notification_templates WHERE is_active=TRUE)t LEFT JOIN notification_preferences p ON p.event_type=t.event_type AND p.user_id=$1 ORDER BY t.event_type`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []NotificationPreferencePayload{}
	for rows.Next() {
		var p NotificationPreferencePayload
		if err = rows.Scan(&p.EventType, &p.InAppEnabled, &p.SMSEnabled); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
func (s *OperationsService) SaveNotificationPreferences(ctx context.Context, userID string, items []NotificationPreferencePayload) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, p := range items {
		if p.EventType == "CUSTOMER_ACCOUNT_CREATED" || p.EventType == "CUSTOMER_ACCOUNT_ACTIVATED" {
			p.InAppEnabled = true
			p.SMSEnabled = true
		}
		var exists bool
		if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM notification_templates WHERE event_type=$1)`, p.EventType).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return ErrValidation
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO notification_preferences(user_id,event_type,in_app_enabled,sms_enabled,updated_at) VALUES($1,$2,$3,$4,NOW()) ON CONFLICT(user_id,event_type) DO UPDATE SET in_app_enabled=EXCLUDED.in_app_enabled,sms_enabled=EXCLUDED.sms_enabled,updated_at=NOW()`, userID, p.EventType, p.InAppEnabled, p.SMSEnabled)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *OperationsService) ListNotificationTemplates(ctx context.Context) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,event_type,channel,locale,title_template,body_template,allowed_variables,is_active,updated_at FROM notification_templates ORDER BY event_type,channel`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, event, channel, locale, title, body string
		var allowed []byte
		var active bool
		var updated time.Time
		if err = rows.Scan(&id, &event, &channel, &locale, &title, &body, &allowed, &active, &updated); err != nil {
			return nil, err
		}
		var vars []string
		_ = json.Unmarshal(allowed, &vars)
		out = append(out, map[string]any{"id": id, "event_type": event, "channel": channel, "locale": locale, "title_template": title, "body_template": body, "allowed_variables": vars, "is_active": active, "updated_at": updated})
	}
	return out, rows.Err()
}

type NotificationTemplateUpdate struct {
	TitleTemplate string `json:"title_template"`
	BodyTemplate  string `json:"body_template"`
	IsActive      bool   `json:"is_active"`
}

func (s *OperationsService) UpdateNotificationTemplate(ctx context.Context, actor, id, key string, p NotificationTemplateUpdate) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, "NOTIFICATION_TEMPLATE_UPDATE", key, map[string]any{"id": id, "payload": p})
	if err != nil {
		return err
	}
	if claim.Existing {
		return tx.Commit()
	}
	var raw []byte
	if err = tx.QueryRowContext(ctx, `SELECT allowed_variables FROM notification_templates WHERE id=$1 FOR UPDATE`, id).Scan(&raw); err != nil {
		return err
	}
	var allowed []string
	if err = json.Unmarshal(raw, &allowed); err != nil {
		return err
	}
	dummy := map[string]string{}
	for _, x := range allowed {
		dummy[x] = "x"
	}
	if _, err = renderNotificationTemplate(p.TitleTemplate, allowed, dummy); err != nil {
		return err
	}
	if _, err = renderNotificationTemplate(p.BodyTemplate, allowed, dummy); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE notification_templates SET title_template=$2,body_template=$3,is_active=$4,updated_at=NOW() WHERE id=$1`, id, p.TitleTemplate, p.BodyTemplate, p.IsActive)
	if err != nil {
		return err
	}
	if err = finishOperationTx(ctx, tx, actor, "NOTIFICATION_TEMPLATE_UPDATE", key, map[string]bool{"updated": true}); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *OperationsService) RetryNotificationDelivery(ctx context.Context, actor, id, key string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, "NOTIFICATION_RETRY", key, map[string]string{"id": id})
	if err != nil {
		return err
	}
	if claim.Existing {
		return tx.Commit()
	}
	r, err := tx.ExecContext(ctx, `UPDATE notification_outbox SET status='RETRY',next_attempt_at=NOW(),last_error=NULL WHERE id=$1 AND status IN ('FAILED','CANCELLED')`, id)
	if err != nil {
		return err
	}
	if n, _ := r.RowsAffected(); n == 0 {
		return conflict("INVALID_NOTIFICATION_TRANSITION", "delivery cannot be retried")
	}
	if err = finishOperationTx(ctx, tx, actor, "NOTIFICATION_RETRY", key, map[string]bool{"queued": true}); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *OperationsService) ListNotificationDeliveries(ctx context.Context, status string, limit int) ([]map[string]any, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	status = normalizeCode(status)
	rows, err := s.db.QueryContext(ctx, `SELECT id,user_id,channel,status,attempt_count,next_attempt_at,provider_message_id,COALESCE(last_error,''),created_at,sent_at FROM notification_outbox WHERE ($1='' OR status=$1) ORDER BY created_at DESC LIMIT $2`, status, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, user, channel, deliveryStatus, lastError string
		var provider sql.NullString
		var attempts int
		var next, created time.Time
		var sent sql.NullTime
		if err = rows.Scan(&id, &user, &channel, &deliveryStatus, &attempts, &next, &provider, &lastError, &created, &sent); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"id": id, "user_id": user, "channel": channel, "status": deliveryStatus, "attempt_count": attempts, "next_attempt_at": next, "provider_message_id": scanNullableString(provider), "last_error": lastError, "created_at": created, "sent_at": nullableTime(sent)})
	}
	return out, rows.Err()
}

type WorkerResult struct {
	Job      string `json:"job"`
	Affected int    `json:"affected"`
	Error    string `json:"error,omitempty"`
}

func (s *OperationsService) RunWorkerOnce(ctx context.Context, retry []time.Duration) ([]WorkerResult, error) {
	jobs := []struct {
		name string
		fn   func(context.Context) (int, error)
	}{{"payment_due", s.runPaymentDueJob}, {"workflow_delay", s.runWorkflowDelayJob}, {"shipment_eta", s.runShipmentETAJob}, {"sales_followup", s.runSalesFollowupJob}, {"operations_report", s.refreshOperationsReportJob}, {"integrity_detection", s.DetectIntegrityFindings}, {"notification_outbox", func(ctx context.Context) (int, error) { return s.processNotificationOutbox(ctx, retry) }}}
	out := make([]WorkerResult, 0, len(jobs))
	for _, job := range jobs {
		n, err := s.withWorkerLock(ctx, job.name, job.fn)
		r := WorkerResult{Job: job.name, Affected: n}
		if err != nil {
			r.Error = err.Error()
		}
		out = append(out, r)
	}
	return out, nil
}

func (s *OperationsService) withWorkerLock(ctx context.Context, code string, fn func(context.Context) (int, error)) (int, error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	var locked bool
	if err = conn.QueryRowContext(ctx, `SELECT pg_try_advisory_lock(hashtext($1))`, "operations-worker:"+code).Scan(&locked); err != nil {
		return 0, err
	}
	if !locked {
		return 0, nil
	}
	defer conn.ExecContext(context.Background(), `SELECT pg_advisory_unlock(hashtext($1))`, "operations-worker:"+code)
	bucket := time.Now().UTC().Truncate(time.Minute)
	var runID string
	err = s.db.QueryRowContext(ctx, `INSERT INTO scheduled_job_runs(job_code,scheduled_bucket,status) VALUES($1,$2,'RUNNING') ON CONFLICT(job_code,scheduled_bucket) DO NOTHING RETURNING id`, code, bucket).Scan(&runID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	n, runErr := fn(ctx)
	status := "COMPLETED"
	var message any
	if runErr != nil {
		status = "FAILED"
		message = runErr.Error()
	}
	_, _ = s.db.ExecContext(ctx, `UPDATE scheduled_job_runs SET status=$2,affected_count=$3,error_text=$4,completed_at=NOW() WHERE id=$1`, runID, status, n, message)
	return n, runErr
}

func (s *OperationsService) processNotificationOutbox(ctx context.Context, retry []time.Duration) (int, error) {
	if !s.FeatureEnabled(ctx, "sms_enabled") {
		result, err := s.db.ExecContext(ctx, `UPDATE notification_outbox SET status='CANCELLED',last_error='provider-disabled' WHERE channel='SMS' AND status IN ('PENDING','RETRY','PROCESSING')`)
		if err != nil {
			return 0, err
		}
		affected, _ := result.RowsAffected()
		return int(affected), nil
	}
	count := 0
	for i := 0; i < 100; i++ {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return count, err
		}
		var id, recipient, body string
		var attempts int
		err = tx.QueryRowContext(ctx, `SELECT id,recipient,message_body,attempt_count FROM notification_outbox WHERE (status IN ('PENDING','RETRY') AND next_attempt_at<=NOW()) OR (status='PROCESSING' AND locked_at<NOW()-INTERVAL '5 minutes') ORDER BY created_at FOR UPDATE SKIP LOCKED LIMIT 1`).Scan(&id, &recipient, &body, &attempts)
		if errors.Is(err, sql.ErrNoRows) {
			tx.Rollback()
			break
		}
		if err != nil {
			tx.Rollback()
			return count, err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE notification_outbox SET status='PROCESSING',locked_at=NOW(),attempt_count=attempt_count+1 WHERE id=$1`, id); err != nil {
			tx.Rollback()
			return count, err
		}
		if err = tx.Commit(); err != nil {
			return count, err
		}
		providerID, sendErr := s.smsProvider.SendMessage(ctx, recipient, body)
		if errors.Is(sendErr, ErrSMSProviderDisabled) {
			_, err = s.db.ExecContext(ctx, `UPDATE notification_outbox SET status='CANCELLED',last_error='provider-disabled' WHERE id=$1`, id)
		} else if sendErr == nil {
			_, err = s.db.ExecContext(ctx, `UPDATE notification_outbox SET status='SENT',provider_message_id=$2,sent_at=NOW(),last_error=NULL WHERE id=$1`, id, providerID)
		} else {
			if attempts >= len(retry) {
				_, err = s.db.ExecContext(ctx, `UPDATE notification_outbox SET status='FAILED',last_error=$2 WHERE id=$1`, id, sendErr.Error())
			} else {
				delay := retry[attempts]
				_, err = s.db.ExecContext(ctx, `UPDATE notification_outbox SET status='RETRY',next_attempt_at=NOW()+($2::bigint*INTERVAL '1 millisecond'),last_error=$3 WHERE id=$1`, id, delay.Milliseconds(), sendErr.Error())
			}
		}
		if err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (s *OperationsService) runPaymentDueJob(ctx context.Context) (int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, `SELECT id,order_id,(amount-paid_amount)::text,currency,'PAYMENT_DUE_SOON' FROM order_payment_schedule WHERE customer_visible AND due_at BETWEEN NOW() AND NOW()+INTERVAL '24 hours' AND paid_amount<amount AND status NOT IN ('PAID','WAIVED','CANCELLED') UNION ALL SELECT id,order_id,(amount-paid_amount)::text,currency,'PAYMENT_OVERDUE' FROM order_payment_schedule WHERE customer_visible AND due_at<NOW() AND paid_amount<amount AND status NOT IN ('PAID','WAIVED','CANCELLED')`)
	if err != nil {
		return 0, err
	}
	type due struct{ id, order, amount, currency, event string }
	items := []due{}
	for rows.Next() {
		var x due
		if err = rows.Scan(&x.id, &x.order, &x.amount, &x.currency, &x.event); err != nil {
			rows.Close()
			return 0, err
		}
		items = append(items, x)
	}
	if err = rows.Close(); err != nil {
		return 0, err
	}
	for _, x := range items {
		if x.event == "PAYMENT_OVERDUE" {
			if _, err = tx.ExecContext(ctx, `UPDATE order_payment_schedule SET status=CASE WHEN paid_amount>0 THEN 'PARTIALLY_PAID' ELSE 'OVERDUE' END,updated_at=NOW() WHERE id=$1`, x.id); err != nil {
				return 0, err
			}
		}
		var user, number string
		if err = tx.QueryRowContext(ctx, `SELECT customer_user_id,order_number FROM orders WHERE id=$1`, x.order).Scan(&user, &number); err != nil {
			return 0, err
		}
		if err = emitNotificationTx(ctx, tx, user, x.event, strings.ToLower(x.event)+":"+x.id+":"+time.Now().UTC().Format("2006-01-02"), "ORDER", x.order, "/account", map[string]string{"order_number": number, "amount": x.amount, "currency": x.currency}); err != nil {
			return 0, err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO action_items(order_id,customer_user_id,title_fa,description_fa,status,priority,assigned_user_id,required_permission_code,due_at,deduplication_key,source_trigger_type) SELECT o.id,o.customer_user_id,CASE WHEN $3='PAYMENT_OVERDUE' THEN 'پیگیری پرداخت عقب‌افتاده' ELSE 'پیگیری سررسید پرداخت' END,o.order_number,'OPEN',CASE WHEN $3='PAYMENT_OVERDUE' THEN 'HIGH' ELSE 'NORMAL' END,o.sales_owner_user_id,'customers.contacts.record',NOW(),'payment:due:'||$1||':'||$3||':'||TO_CHAR(CURRENT_DATE,'YYYY-MM-DD'),$3 FROM orders o WHERE o.id=$2 AND o.sales_owner_user_id IS NOT NULL ON CONFLICT(deduplication_key) WHERE deduplication_key IS NOT NULL DO NOTHING`, x.id, x.order, x.event)
		if err != nil {
			return 0, err
		}
	}
	return len(items), tx.Commit()
}
func (s *OperationsService) runWorkflowDelayJob(ctx context.Context) (int, error) {
	r, err := s.db.ExecContext(ctx, `INSERT INTO action_items(workflow_instance_id,workflow_step_instance_id,order_id,customer_user_id,title_fa,description_fa,status,priority,assigned_role_id,assigned_user_id,required_permission_code,due_at,deduplication_key,source_trigger_type) SELECT si.workflow_instance_id,si.id,wi.order_id,wi.customer_user_id,'پیگیری تأخیر مرحله',COALESCE(si.internal_title_fa,si.step_code),'OPEN','HIGH',si.responsible_role_id,si.assigned_user_id,'workflow_steps.submit',NOW(), 'workflow:delay:'||si.id||':'||TO_CHAR(CURRENT_DATE,'YYYY-MM-DD'),'SCHEDULED_WORKFLOW_DELAY' FROM workflow_step_instances si JOIN workflow_instances wi ON wi.id=si.workflow_instance_id WHERE si.estimated_end_at<NOW() AND si.status IN ('WAITING_FOR_ASSIGNEE','IN_PROGRESS','BLOCKED','NEEDS_CORRECTION') ON CONFLICT(deduplication_key) WHERE deduplication_key IS NOT NULL DO NOTHING`)
	if err != nil {
		return 0, err
	}
	n, _ := r.RowsAffected()
	return int(n), nil
}
func (s *OperationsService) runShipmentETAJob(ctx context.Context) (int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, `INSERT INTO action_items(order_id,shipment_id,customer_user_id,title_fa,description_fa,status,priority,assigned_user_id,required_permission_code,due_at,deduplication_key,source_trigger_type) SELECT s.order_id,s.id,o.customer_user_id,'پیگیری زمان تحویل محموله',s.shipment_number,'OPEN','NORMAL',o.sales_owner_user_id,'shipments.view_assigned',s.estimated_arrival_at,'shipment:eta:'||s.id||':'||TO_CHAR(CURRENT_DATE,'YYYY-MM-DD'),'SCHEDULED_SHIPMENT_ETA' FROM shipments s JOIN orders o ON o.id=s.order_id WHERE s.estimated_arrival_at BETWEEN NOW() AND NOW()+INTERVAL '24 hours' AND s.status NOT IN ('DELIVERED','CANCELLED') ON CONFLICT(deduplication_key) WHERE deduplication_key IS NOT NULL DO NOTHING RETURNING shipment_id,order_id,customer_user_id,description_fa,due_at`)
	if err != nil {
		return 0, err
	}
	count := 0
	for rows.Next() {
		var shipment, order, user, number string
		var eta time.Time
		if err = rows.Scan(&shipment, &order, &user, &number, &eta); err != nil {
			rows.Close()
			return 0, err
		}
		if err = emitNotificationTx(ctx, tx, user, "SHIPMENT_ETA", "shipment-eta:"+shipment+":"+time.Now().UTC().Format("2006-01-02"), "SHIPMENT", shipment, "/account", map[string]string{"shipment_number": number, "eta": eta.In(time.FixedZone("Tehran", 12600)).Format("2006-01-02 15:04")}); err != nil {
			rows.Close()
			return 0, err
		}
		count++
	}
	if err = rows.Close(); err != nil {
		return 0, err
	}
	return count, tx.Commit()
}
func (s *OperationsService) runSalesFollowupJob(ctx context.Context) (int, error) {
	r, err := s.db.ExecContext(ctx, `INSERT INTO action_items(order_id,customer_user_id,title_fa,description_fa,status,priority,assigned_user_id,required_permission_code,due_at,deduplication_key,source_trigger_type) SELECT o.id,o.customer_user_id,'پیگیری فروش',o.order_number,'OPEN','NORMAL',o.sales_owner_user_id,'customers.contacts.record',NOW(),'sales:followup:'||o.id||':'||TO_CHAR(CURRENT_DATE,'IYYY-IW'),'SCHEDULED_SALES_FOLLOWUP' FROM orders o WHERE o.sales_owner_user_id IS NOT NULL AND o.status IN ('PROFORMA_ISSUED','CONFIRMED','IN_PROGRESS') AND NOT EXISTS(SELECT 1 FROM customer_contact_logs l WHERE l.order_id=o.id AND l.contacted_at>NOW()-INTERVAL '7 days') ON CONFLICT(deduplication_key) WHERE deduplication_key IS NOT NULL DO NOTHING`)
	if err != nil {
		return 0, err
	}
	n, _ := r.RowsAffected()
	return int(n), nil
}
func (s *OperationsService) refreshOperationsReportJob(ctx context.Context) (int, error) {
	_, err := s.db.ExecContext(ctx, `INSERT INTO daily_operations_report(report_date,active_workflows,overdue_steps,open_discrepancies,shipments_dispatched,shipments_delivered,on_time_deliveries,average_step_duration_hours,rework_iterations,refreshed_at) SELECT CURRENT_DATE,(SELECT COUNT(*) FROM workflow_instances WHERE status='IN_PROGRESS'),(SELECT COUNT(*) FROM workflow_step_instances WHERE estimated_end_at<NOW() AND status NOT IN ('COMPLETED','SKIPPED','CANCELLED')),(SELECT COUNT(*) FROM workflow_discrepancies WHERE status NOT IN ('RESOLVED','ACCEPTED')),(SELECT COUNT(*) FROM shipments WHERE actual_departure_at::date=CURRENT_DATE),(SELECT COUNT(*) FROM shipments WHERE actual_arrival_at::date=CURRENT_DATE AND status='DELIVERED'),(SELECT COUNT(*) FROM shipments WHERE actual_arrival_at::date=CURRENT_DATE AND status='DELIVERED' AND (estimated_arrival_at IS NULL OR actual_arrival_at<=estimated_arrival_at)),COALESCE((SELECT AVG(EXTRACT(EPOCH FROM (actual_end_at-actual_start_at))/3600) FROM workflow_step_instances WHERE actual_end_at::date=CURRENT_DATE AND actual_start_at IS NOT NULL),0),(SELECT COUNT(*) FROM workflow_step_instances WHERE iteration_number>1 AND created_at::date=CURRENT_DATE),NOW() ON CONFLICT(report_date) DO UPDATE SET active_workflows=EXCLUDED.active_workflows,overdue_steps=EXCLUDED.overdue_steps,open_discrepancies=EXCLUDED.open_discrepancies,shipments_dispatched=EXCLUDED.shipments_dispatched,shipments_delivered=EXCLUDED.shipments_delivered,on_time_deliveries=EXCLUDED.on_time_deliveries,average_step_duration_hours=EXCLUDED.average_step_duration_hours,rework_iterations=EXCLUDED.rework_iterations,refreshed_at=NOW()`)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func sortedKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

var _ = sortedKeys
