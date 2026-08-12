package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"
	"unicode"
)

var containerPattern = regexp.MustCompile(`^[A-Z]{4}[0-9]{7}$`)

func validPositiveDecimal(value string) bool {
	r, ok := new(big.Rat).SetString(strings.TrimSpace(value))
	return ok && r.Sign() > 0
}

func validNonNegativeDecimal(value string) bool {
	r, ok := new(big.Rat).SetString(strings.TrimSpace(value))
	return ok && r.Sign() >= 0
}

func decimalCmp(a, b string) (int, bool) {
	ar, aok := new(big.Rat).SetString(strings.TrimSpace(a))
	br, bok := new(big.Rat).SetString(strings.TrimSpace(b))
	if !aok || !bok {
		return 0, false
	}
	return ar.Cmp(br), true
}

func normalizeCode(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func normalizePlate(value string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(strings.TrimSpace(value)) {
		switch {
		case r == 'ـ':
			continue
		case r >= '۰' && r <= '۹':
			b.WriteRune('0' + (r - '۰'))
		case r >= '٠' && r <= '٩':
			b.WriteRune('0' + (r - '٠'))
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
		}
	}
	return b.String()
}

func requireReason(reason string) error {
	if strings.TrimSpace(reason) == "" {
		return errors.New("reason required")
	}
	return nil
}

func payloadHash(payload any) (string, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

func claimOperationTx(ctx context.Context, tx *sql.Tx, actor, operation, key string, payload any) (IdempotentResult, error) {
	if strings.TrimSpace(key) == "" {
		return IdempotentResult{}, errors.New("Idempotency-Key required")
	}
	hash, err := payloadHash(payload)
	if err != nil {
		return IdempotentResult{}, err
	}
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, actor+"\x1f"+operation+"\x1f"+key); err != nil {
		return IdempotentResult{}, err
	}
	var existingHash, status string
	var response []byte
	err = tx.QueryRowContext(ctx, `SELECT payload_hash,status,response_json FROM inventory_operation_requests WHERE actor_user_id=$1 AND operation_type=$2 AND idempotency_key=$3 FOR UPDATE`, actor, operation, key).Scan(&existingHash, &status, &response)
	if err == nil {
		if existingHash != hash {
			return IdempotentResult{}, conflict("IDEMPOTENCY_CONFLICT", "idempotency key was used with another payload")
		}
		if status != "COMPLETED" {
			return IdempotentResult{}, conflict("OPERATION_IN_PROGRESS", "operation is already in progress")
		}
		return IdempotentResult{Existing: true, Response: response}, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return IdempotentResult{}, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO inventory_operation_requests(actor_user_id,operation_type,idempotency_key,payload_hash) VALUES($1,$2,$3,$4)`, actor, operation, key, hash)
	return IdempotentResult{}, err
}

func finishOperationTx(ctx context.Context, tx *sql.Tx, actor, operation, key string, response any) error {
	b, err := json.Marshal(response)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE inventory_operation_requests SET status='COMPLETED',response_json=$4,completed_at=NOW() WHERE actor_user_id=$1 AND operation_type=$2 AND idempotency_key=$3`, actor, operation, key, b)
	return err
}

func nextReadableNumberTx(ctx context.Context, tx *sql.Tx, prefix string) (string, error) {
	var value string
	err := tx.QueryRowContext(ctx, `SELECT $1||'-'||TO_CHAR(NOW(),'YYYY')||'-'||LPAD(nextval('operations_number_seq')::text,6,'0')`, prefix).Scan(&value)
	return value, err
}

func scanNullableString(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	x := v.String
	return &x
}

func validateUnit(unit string) error {
	if !quantityUnits[normalizeCode(unit)] {
		return conflict("INCOMPATIBLE_UNIT", fmt.Sprintf("unsupported quantity unit %q", unit))
	}
	return nil
}

func nullableString(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return strings.TrimSpace(v)
}

func randomUUIDText() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("00000000-0000-4000-8000-%012d", time.Now().UnixNano()%1_000_000_000_000)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func decimalMath(a, b string, subtract bool) string {
	ar, aok := new(big.Rat).SetString(a)
	br, bok := new(big.Rat).SetString(b)
	if !aok || !bok {
		return "0.0000"
	}
	if subtract {
		ar.Sub(ar, br)
	} else {
		ar.Add(ar, br)
	}
	return ar.FloatString(quantityScale)
}
func addDecimal(a, b string) string { return decimalMath(a, b, false) }
func subDecimal(a, b string) string { return decimalMath(a, b, true) }

func (s *OperationsService) markDomainOperationTx(ctx context.Context, tx *sql.Tx, actor string, stepID *string, event, entityType, entityID, group string) error {
	if stepID == nil || strings.TrimSpace(*stepID) == "" {
		return nil
	}
	var expected, scopeType, scopeID, status, workflowID string
	var current sql.NullString
	err := tx.QueryRowContext(ctx, `SELECT COALESCE(si.domain_event_code,''),wi.scope_type,wi.scope_id,si.status,wi.id,wi.current_step_instance_id FROM workflow_step_instances si JOIN workflow_instances wi ON wi.id=si.workflow_instance_id WHERE si.id=$1 FOR UPDATE OF si,wi`, *stepID).Scan(&expected, &scopeType, &scopeID, &status, &workflowID, &current)
	if err != nil {
		return err
	}
	if expected != event {
		return conflict("DOMAIN_EVENT_MISMATCH", "workflow step does not accept this domain operation")
	}
	if !current.Valid || current.String != *stepID || status != "IN_PROGRESS" {
		return conflict("INVALID_TRANSITION", "domain operation requires the current in-progress workflow step")
	}
	if scopeType != entityType || scopeID != entityID {
		return conflict("SCOPE_MISMATCH", "domain operation belongs to another workflow scope")
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO workflow_domain_operation_completions(workflow_step_instance_id,domain_event_code,entity_type,entity_id,operation_group_id,completed_by_user_id) VALUES($1,$2,$3,$4,NULLIF($5,'')::uuid,$6) ON CONFLICT DO NOTHING`, *stepID, event, entityType, entityID, group, actor)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return nil
	}
	if err = s.runStepTriggersTx(ctx, tx, workflowID, *stepID, "ON_STEP_SUBMIT"); err != nil {
		return err
	}
	paymentEvent := map[string]string{"SHIPMENT_LOADED": "LOADING", "SHIPMENT_DISPATCHED": "DISPATCH", "SHIPMENT_DELIVERED": "DELIVERY"}[event]
	if paymentEvent != "" {
		var orderID string
		if err = tx.QueryRowContext(ctx, `SELECT order_id FROM workflow_instances WHERE id=$1`, workflowID).Scan(&orderID); err != nil {
			return err
		}
		var blocked int
		blocked, err = s.evaluatePaymentTriggerTx(ctx, tx, orderID, paymentEvent, *stepID)
		if err != nil {
			return err
		}
		if blocked > 0 {
			return nil
		}
	}
	var requiresApproval bool
	if err = tx.QueryRowContext(ctx, `SELECT requires_approval FROM workflow_step_instances WHERE id=$1`, *stepID).Scan(&requiresApproval); err != nil {
		return err
	}
	if requiresApproval {
		if _, err = tx.ExecContext(ctx, `UPDATE workflow_step_instances SET status='WAITING_FOR_APPROVAL',submitted_at=NOW(),customer_status_text='در حال بررسی',updated_at=NOW() WHERE id=$1`, *stepID); err != nil {
			return err
		}
		return s.createApprovalActionTx(ctx, tx, *stepID)
	}
	return s.completeStepTx(ctx, tx, actor, workflowID, *stepID, "COMPLETED")
}
