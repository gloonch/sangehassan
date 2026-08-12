package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

func decimalMul(a, b string) (string, bool) {
	ar, aok := new(big.Rat).SetString(strings.TrimSpace(a))
	br, bok := new(big.Rat).SetString(strings.TrimSpace(b))
	if !aok || !bok {
		return "", false
	}
	return ar.Mul(ar, br).FloatString(4), true
}

func decimalDiv(a, b string) (string, bool) {
	ar, aok := new(big.Rat).SetString(strings.TrimSpace(a))
	br, bok := new(big.Rat).SetString(strings.TrimSpace(b))
	if !aok || !bok || br.Sign() == 0 {
		return "", false
	}
	return ar.Quo(ar, br).FloatString(4), true
}

func validateCommercialTerms(p CommercialTermsPayload) error {
	p.TermsType, p.Currency = normalizeCode(p.TermsType), normalizeCode(p.Currency)
	validTermsType := map[string]bool{"FULL_PREPAYMENT": true, "DEPOSIT_AND_BALANCE": true, "INSTALLMENTS": true, "PAY_ON_DELIVERY": true, "CREDIT": true, "CUSTOM": true}
	if !validNonNegativeDecimal(p.Subtotal) || !validNonNegativeDecimal(p.DiscountAmount) ||
		!validNonNegativeDecimal(p.TaxAmount) || !validNonNegativeDecimal(p.AdditionalCharge) ||
		!validNonNegativeDecimal(p.FinalCustomerAmount) || len(p.Currency) != 3 || !validTermsType[p.TermsType] {
		return ErrValidation
	}
	expected := addDecimal(subDecimal(p.Subtotal, p.DiscountAmount), addDecimal(p.TaxAmount, p.AdditionalCharge))
	if cmp, ok := decimalCmp(expected, p.FinalCustomerAmount); !ok || cmp != 0 {
		return errors.New("final_customer_amount does not match subtotal - discount + tax + charges")
	}
	if cmp, _ := decimalCmp(p.DiscountAmount, p.Subtotal); cmp > 0 {
		return errors.New("discount cannot exceed subtotal")
	}
	if p.DepositPercentage != nil {
		if !validNonNegativeDecimal(*p.DepositPercentage) {
			return ErrValidation
		}
		if cmp, _ := decimalCmp(*p.DepositPercentage, "100"); cmp > 0 {
			return ErrValidation
		}
	}
	if p.DepositAmount != nil && !validNonNegativeDecimal(*p.DepositAmount) {
		return ErrValidation
	}
	if p.TermsType == "DEPOSIT_AND_BALANCE" {
		deposit := ""
		if p.DepositAmount != nil {
			deposit = *p.DepositAmount
		} else if p.DepositPercentage != nil {
			product, ok := decimalMul(p.FinalCustomerAmount, *p.DepositPercentage)
			if !ok {
				return ErrValidation
			}
			deposit, ok = decimalDiv(product, "100")
			if !ok {
				return ErrValidation
			}
		}
		if deposit == "" {
			return errors.New("deposit amount or percentage is required")
		}
		if cmp, _ := decimalCmp(deposit, "0"); cmp <= 0 {
			return errors.New("deposit must be positive")
		}
		if cmp, _ := decimalCmp(deposit, p.FinalCustomerAmount); cmp >= 0 {
			return errors.New("deposit must be less than final amount")
		}
		if p.DepositAmount != nil && p.DepositPercentage != nil {
			product, _ := decimalMul(p.FinalCustomerAmount, *p.DepositPercentage)
			percentageAmount, _ := decimalDiv(product, "100")
			if cmp, _ := decimalCmp(*p.DepositAmount, percentageAmount); cmp != 0 {
				return errors.New("deposit amount does not match percentage")
			}
		}
	}
	return nil
}

func (s *OperationsService) GetCommercialTerms(ctx context.Context, actor, orderID string, customer bool) (CommercialTerms, error) {
	var out CommercialTerms
	var dp, da sql.NullString
	if customer {
		var owner bool
		if err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM orders WHERE id=$1 AND customer_user_id=$2)`, orderID, actor).Scan(&owner); err != nil {
			return out, err
		}
		if !owner {
			return out, ErrForbidden
		}
	}
	err := s.db.QueryRowContext(ctx, `SELECT order_id,terms_type,currency,subtotal::text,discount_amount::text,tax_amount::text,additional_charge_amount::text,final_customer_amount::text,deposit_percentage::text,deposit_amount::text,COALESCE(payment_terms_text,''),COALESCE(delivery_terms_text,''),version_number,updated_at FROM order_commercial_terms WHERE order_id=$1`, orderID).Scan(&out.OrderID, &out.TermsType, &out.Currency, &out.Subtotal, &out.DiscountAmount, &out.TaxAmount, &out.AdditionalChargeAmount, &out.FinalCustomerAmount, &dp, &da, &out.PaymentTermsText, &out.DeliveryTermsText, &out.VersionNumber, &out.UpdatedAt)
	out.DepositPercentage = scanNullableString(dp)
	out.DepositAmount = scanNullableString(da)
	return out, err
}

func (s *OperationsService) SaveCommercialTerms(ctx context.Context, actor, orderID, key string, p CommercialTermsPayload) (CommercialTerms, error) {
	var out CommercialTerms
	if err := validateCommercialTerms(p); err != nil {
		return out, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return out, err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, "COMMERCIAL_TERMS_SAVE", key, map[string]any{"order_id": orderID, "payload": p})
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
	if err = tx.QueryRowContext(ctx, `SELECT status FROM orders WHERE id=$1 FOR UPDATE`, orderID).Scan(&status); err != nil {
		return out, err
	}
	if status != "DRAFT" && status != "PROFORMA_ISSUED" && strings.TrimSpace(p.Reason) == "" {
		return out, conflict(ErrInvalidFinancialTransition, "reason is required after order confirmation")
	}
	var dp, da any
	if p.DepositPercentage != nil {
		dp = *p.DepositPercentage
	}
	if p.DepositAmount != nil {
		da = *p.DepositAmount
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO order_commercial_terms(order_id,terms_type,currency,subtotal,discount_amount,tax_amount,additional_charge_amount,final_customer_amount,deposit_percentage,deposit_amount,payment_terms_text,delivery_terms_text,last_change_reason,updated_by_user_id) VALUES($1,$2,$3,$4::numeric,$5::numeric,$6::numeric,$7::numeric,$8::numeric,$9::numeric,$10::numeric,NULLIF($11,''),NULLIF($12,''),NULLIF($13,''),$14) ON CONFLICT(order_id) DO UPDATE SET terms_type=EXCLUDED.terms_type,currency=EXCLUDED.currency,subtotal=EXCLUDED.subtotal,discount_amount=EXCLUDED.discount_amount,tax_amount=EXCLUDED.tax_amount,additional_charge_amount=EXCLUDED.additional_charge_amount,final_customer_amount=EXCLUDED.final_customer_amount,deposit_percentage=EXCLUDED.deposit_percentage,deposit_amount=EXCLUDED.deposit_amount,payment_terms_text=EXCLUDED.payment_terms_text,delivery_terms_text=EXCLUDED.delivery_terms_text,last_change_reason=EXCLUDED.last_change_reason,updated_by_user_id=EXCLUDED.updated_by_user_id,version_number=order_commercial_terms.version_number+1,updated_at=NOW()`, orderID, normalizeCode(p.TermsType), normalizeCode(p.Currency), p.Subtotal, p.DiscountAmount, p.TaxAmount, p.AdditionalCharge, p.FinalCustomerAmount, dp, da, p.PaymentTermsText, p.DeliveryTermsText, p.Reason, actor)
	if err != nil {
		return out, err
	}
	if normalizeCode(p.TermsType) == "DEPOSIT_AND_BALANCE" {
		var count int
		if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM order_payment_schedule WHERE order_id=$1`, orderID).Scan(&count); err != nil {
			return out, err
		}
		if count == 0 {
			deposit := "0.0000"
			if p.DepositAmount != nil {
				deposit = *p.DepositAmount
			} else if p.DepositPercentage != nil {
				x, _ := decimalMul(p.FinalCustomerAmount, *p.DepositPercentage)
				deposit, _ = decimalDiv(x, "100")
			}
			if cmp, _ := decimalCmp(deposit, "0"); cmp > 0 {
				_, err = tx.ExecContext(ctx, `INSERT INTO order_payment_schedule(order_id,sequence_number,title_fa,payment_type,amount,percentage_of_order,currency,status,trigger_type) VALUES($1,1,'پیش‌پرداخت','DEPOSIT',$2::numeric,$5::numeric,$3,'DUE','ORDER_CONFIRMATION'),($1,2,'تسویه مانده','FINAL_PAYMENT',($4::numeric-$2::numeric),NULL,$3,'UPCOMING','DELIVERY')`, orderID, deposit, normalizeCode(p.Currency), p.FinalCustomerAmount, p.DepositPercentage)
				if err != nil {
					return out, err
				}
			}
		}
	}
	if err = refreshFinancialSummaryTx(ctx, tx, orderID); err != nil {
		return out, err
	}
	if err = scanCommercialTermsTx(ctx, tx, orderID, &out); err != nil {
		return out, err
	}
	if err = auditTx(ctx, tx, actor, "finance.commercial_terms.update", "order", orderID, map[string]any{"reason": p.Reason}, p); err != nil {
		return out, err
	}
	if err = finishOperationTx(ctx, tx, actor, "COMMERCIAL_TERMS_SAVE", key, out); err != nil {
		return out, err
	}
	return out, tx.Commit()
}

func scanCommercialTermsTx(ctx context.Context, tx *sql.Tx, orderID string, out *CommercialTerms) error {
	var dp, da sql.NullString
	err := tx.QueryRowContext(ctx, `SELECT order_id,terms_type,currency,subtotal::text,discount_amount::text,tax_amount::text,additional_charge_amount::text,final_customer_amount::text,deposit_percentage::text,deposit_amount::text,COALESCE(payment_terms_text,''),COALESCE(delivery_terms_text,''),version_number,updated_at FROM order_commercial_terms WHERE order_id=$1`, orderID).Scan(&out.OrderID, &out.TermsType, &out.Currency, &out.Subtotal, &out.DiscountAmount, &out.TaxAmount, &out.AdditionalChargeAmount, &out.FinalCustomerAmount, &dp, &da, &out.PaymentTermsText, &out.DeliveryTermsText, &out.VersionNumber, &out.UpdatedAt)
	out.DepositPercentage = scanNullableString(dp)
	out.DepositAmount = scanNullableString(da)
	return err
}

func auditTx(ctx context.Context, tx *sql.Tx, actor, action, entity, id string, before, after any) error {
	b, _ := json.Marshal(before)
	a, _ := json.Marshal(after)
	if requestID := RequestIDFromContext(ctx); requestID != "" {
		_, err := tx.ExecContext(ctx, `INSERT INTO audit_logs(actor_user_id,action_code,entity_type,entity_id,before_data,after_data,request_id) VALUES($1,$2,$3,$4,$5::jsonb,$6::jsonb,$7)`, actor, action, entity, id, string(b), string(a), requestID)
		return err
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO audit_logs(actor_user_id,action_code,entity_type,entity_id,before_data,after_data) VALUES($1,$2,$3,$4,$5::jsonb,$6::jsonb)`, actor, action, entity, id, string(b), string(a))
	return err
}

func (s *OperationsService) ListPaymentSchedule(ctx context.Context, actor, orderID string, customer bool) ([]PaymentScheduleItem, error) {
	if customer {
		var ok bool
		if err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM orders WHERE id=$1 AND customer_user_id=$2)`, orderID, actor).Scan(&ok); err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrForbidden
		}
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,sequence_number,title_fa,payment_type,due_at,amount::text,percentage_of_order::text,currency,paid_amount::text,status,trigger_type,trigger_step_code,customer_visible FROM order_payment_schedule WHERE order_id=$1 AND status<>'CANCELLED' AND ($2::boolean=FALSE OR customer_visible) ORDER BY sequence_number`, orderID, customer)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []PaymentScheduleItem{}
	for rows.Next() {
		var x PaymentScheduleItem
		var due sql.NullTime
		var code, percentage sql.NullString
		if err = rows.Scan(&x.ID, &x.Sequence, &x.TitleFA, &x.PaymentType, &due, &x.Amount, &percentage, &x.Currency, &x.PaidAmount, &x.Status, &x.TriggerType, &code, &x.CustomerVisible); err != nil {
			return nil, err
		}
		if due.Valid {
			x.DueAt = &due.Time
		}
		x.TriggerStepCode = scanNullableString(code)
		x.Percentage = scanNullableString(percentage)
		out = append(out, x)
	}
	return out, rows.Err()
}

func (s *OperationsService) SavePaymentSchedule(ctx context.Context, actor, orderID, key string, p PaymentSchedulePayload) ([]PaymentScheduleItem, error) {
	if len(p.Items) == 0 {
		return nil, ErrValidation
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, "PAYMENT_SCHEDULE_SAVE", key, map[string]any{"order_id": orderID, "payload": p})
	if err != nil {
		return nil, err
	}
	if claim.Existing {
		var out []PaymentScheduleItem
		if err = json.Unmarshal(claim.Response, &out); err != nil {
			return nil, err
		}
		return out, tx.Commit()
	}
	var confirmed int
	var termsCurrency, finalAmount string
	if err = tx.QueryRowContext(ctx, `SELECT currency,final_customer_amount::text FROM order_commercial_terms WHERE order_id=$1 FOR UPDATE`, orderID).Scan(&termsCurrency, &finalAmount); err != nil {
		return nil, err
	}
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM payment_allocations a JOIN customer_payments p ON p.id=a.payment_id WHERE p.order_id=$1 AND p.status IN ('CONFIRMED','PARTIALLY_REFUNDED','REFUNDED')`, orderID).Scan(&confirmed); err != nil {
		return nil, err
	}
	if confirmed > 0 {
		return nil, conflict(ErrInvalidFinancialTransition, "payment schedule with confirmed allocations cannot be rebuilt")
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM order_payment_schedule WHERE order_id=$1`, orderID); err != nil {
		return nil, err
	}
	total := "0.0000"
	validTriggers := map[string]bool{"DATE": true, "ORDER_CONFIRMATION": true, "STEP_OPEN": true, "STEP_COMPLETE": true, "LOADING": true, "DISPATCH": true, "DELIVERY": true, "MANUAL": true}
	for i, item := range p.Items {
		item.Currency = normalizeCode(item.Currency)
		item.TriggerType = normalizeCode(item.TriggerType)
		item.PaymentType = normalizeCode(item.PaymentType)
		if item.PaymentType == "" {
			item.PaymentType = "CUSTOM"
		}
		validPaymentTypes := map[string]bool{"DEPOSIT": true, "PROGRESS_PAYMENT": true, "BEFORE_LOADING": true, "BEFORE_SHIPMENT": true, "ON_DELIVERY": true, "POST_DELIVERY": true, "INSTALLMENT": true, "FINAL_PAYMENT": true, "CUSTOM": true}
		if !validPositiveDecimal(item.Amount) || item.Currency != termsCurrency || strings.TrimSpace(item.TitleFA) == "" || !validTriggers[item.TriggerType] || !validPaymentTypes[item.PaymentType] {
			return nil, ErrValidation
		}
		if item.Percentage != nil && (!validPositiveDecimal(*item.Percentage) || decimalCmpMust(*item.Percentage, "100") > 0) {
			return nil, ErrValidation
		}
		if (item.TriggerType == "STEP_OPEN" || item.TriggerType == "STEP_COMPLETE") && strings.TrimSpace(item.TriggerStepCode) == "" {
			return nil, ErrValidation
		}
		total = addDecimal(total, item.Amount)
		visible := true
		if item.CustomerVisible != nil {
			visible = *item.CustomerVisible
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO order_payment_schedule(order_id,sequence_number,title_fa,payment_type,due_at,amount,percentage_of_order,currency,status,trigger_type,trigger_step_code,customer_visible) VALUES($1,$2,$3,$4,$5,$6::numeric,$7::numeric,$8,CASE WHEN $5::timestamptz<=NOW() THEN 'DUE' ELSE 'UPCOMING' END,$9,NULLIF($10,''),$11)`, orderID, i+1, item.TitleFA, item.PaymentType, item.DueAt, item.Amount, item.Percentage, item.Currency, item.TriggerType, item.TriggerStepCode, visible)
		if err != nil {
			return nil, err
		}
	}
	if cmp, _ := decimalCmp(total, finalAmount); cmp != 0 {
		return nil, conflict(ErrInvalidFinancialTransition, "payment schedule total must equal final customer amount")
	}
	if err = refreshFinancialSummaryTx(ctx, tx, orderID); err != nil {
		return nil, err
	}
	out, err := listPaymentScheduleTx(ctx, tx, orderID)
	if err != nil {
		return nil, err
	}
	if err = finishOperationTx(ctx, tx, actor, "PAYMENT_SCHEDULE_SAVE", key, out); err != nil {
		return nil, err
	}
	return out, tx.Commit()
}

func listPaymentScheduleTx(ctx context.Context, tx *sql.Tx, orderID string) ([]PaymentScheduleItem, error) {
	rows, err := tx.QueryContext(ctx, `SELECT id,sequence_number,title_fa,payment_type,due_at,amount::text,percentage_of_order::text,currency,paid_amount::text,status,trigger_type,trigger_step_code,customer_visible FROM order_payment_schedule WHERE order_id=$1 ORDER BY sequence_number`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []PaymentScheduleItem{}
	for rows.Next() {
		var x PaymentScheduleItem
		var d sql.NullTime
		var c, percentage sql.NullString
		if err = rows.Scan(&x.ID, &x.Sequence, &x.TitleFA, &x.PaymentType, &d, &x.Amount, &percentage, &x.Currency, &x.PaidAmount, &x.Status, &x.TriggerType, &c, &x.CustomerVisible); err != nil {
			return nil, err
		}
		if d.Valid {
			x.DueAt = &d.Time
		}
		x.TriggerStepCode = scanNullableString(c)
		x.Percentage = scanNullableString(percentage)
		out = append(out, x)
	}
	return out, rows.Err()
}

func (s *OperationsService) ConfirmOrder(ctx context.Context, actor, orderID, key, reason string) (map[string]any, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	payload := map[string]string{"order_id": orderID, "reason": reason}
	claim, err := claimOperationTx(ctx, tx, actor, "ORDER_CONFIRM", key, payload)
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
	var status string
	var amount, currency, customer, orderNumber string
	var salesOwner sql.NullString
	if err = tx.QueryRowContext(ctx, `SELECT o.status,t.final_customer_amount::text,t.currency,o.customer_user_id,o.order_number,o.sales_owner_user_id FROM orders o JOIN order_commercial_terms t ON t.order_id=o.id WHERE o.id=$1 FOR UPDATE OF o,t`, orderID).Scan(&status, &amount, &currency, &customer, &orderNumber, &salesOwner); err != nil {
		return nil, err
	}
	if status != "DRAFT" && status != "PROFORMA_ISSUED" {
		return nil, conflict(ErrInvalidFinancialTransition, "order cannot be confirmed from current state")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE orders SET status='CONFIRMED',confirmed_at=NOW(),updated_at=NOW() WHERE id=$1`, orderID); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE order_payment_schedule SET status=CASE WHEN trigger_type='ORDER_CONFIRMATION' OR due_at<=NOW() THEN 'DUE' ELSE status END,updated_at=NOW() WHERE order_id=$1`, orderID); err != nil {
		return nil, err
	}
	if _, err = s.evaluatePaymentTriggerTx(ctx, tx, orderID, "ORDER_CONFIRMATION", ""); err != nil {
		return nil, err
	}
	if err = refreshFinancialSummaryTx(ctx, tx, orderID); err != nil {
		return nil, err
	}
	if err = emitNotificationTx(ctx, tx, customer, "ORDER_CONFIRMED", "order-confirmed:"+orderID, "ORDER", orderID, "/account", map[string]string{}); err != nil {
		return nil, err
	}
	if salesOwner.Valid {
		if err = emitNotificationTx(ctx, tx, salesOwner.String, "ORDER_CONFIRMED", "order-confirmed:"+orderID, "ORDER", orderID, "/panel/dashboard/orders/"+orderID, map[string]string{"order_number": orderNumber}); err != nil {
			return nil, err
		}
	}
	out := map[string]any{"order_id": orderID, "status": "CONFIRMED", "final_customer_amount": amount, "currency": currency}
	if err = auditTx(ctx, tx, actor, "orders.confirm", "order", orderID, map[string]string{"status": status}, out); err != nil {
		return nil, err
	}
	if err = finishOperationTx(ctx, tx, actor, "ORDER_CONFIRM", key, out); err != nil {
		return nil, err
	}
	return out, tx.Commit()
}

func (s *OperationsService) RecordPayment(ctx context.Context, actor, orderID, key string, p PaymentPayload) (Payment, error) {
	var out Payment
	p.Currency = normalizeCode(p.Currency)
	p.PaymentMethod = normalizeCode(p.PaymentMethod)
	validMethods := map[string]bool{"BANK_TRANSFER": true, "CARD": true, "CARD_TO_CARD": true, "CASH": true, "CHEQUE": true, "SWIFT": true, "WIRE_TRANSFER": true, "POS": true, "OTHER": true}
	if !validPositiveDecimal(p.Amount) || len(p.Currency) != 3 || !validMethods[p.PaymentMethod] {
		return out, ErrValidation
	}
	paid := time.Now()
	if p.PaidAt != nil {
		paid = *p.PaidAt
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return out, err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, "PAYMENT_RECORD", key, map[string]any{"order_id": orderID, "payload": p})
	if err != nil {
		return out, err
	}
	if claim.Existing {
		if err = json.Unmarshal(claim.Response, &out); err != nil {
			return out, err
		}
		return out, tx.Commit()
	}
	var customer string
	if err = tx.QueryRowContext(ctx, `SELECT customer_user_id FROM orders WHERE id=$1 FOR SHARE`, orderID).Scan(&customer); err != nil {
		return out, err
	}
	status := "REPORTED"
	if strings.TrimSpace(p.ReferenceNumber) != "" || p.ReceiptFileID != nil {
		status = "PENDING_CONFIRMATION"
	}
	number, err := nextReadableNumberTx(ctx, tx, "PAY")
	if err != nil {
		return out, err
	}
	var ref sql.NullString
	err = tx.QueryRowContext(ctx, `INSERT INTO customer_payments(payment_number,order_id,customer_user_id,amount,currency,payment_method,status,reference_number,bank_name,receipt_file_id,paid_at,notes,reported_by_user_id) VALUES($1,$2,$3,$4::numeric,$5,$6,$7,NULLIF($8,''),NULLIF($9,''),$10,$11,NULLIF($12,''),$13) RETURNING id,reference_number`, number, orderID, customer, p.Amount, p.Currency, p.PaymentMethod, status, p.ReferenceNumber, p.BankName, p.ReceiptFileID, paid, p.Notes, actor).Scan(&out.ID, &ref)
	if err != nil {
		return out, err
	}
	out.PaymentNumber = number
	out.OrderID = orderID
	out.Amount = p.Amount
	out.Currency = p.Currency
	out.PaymentMethod = p.PaymentMethod
	out.Status = status
	out.ReferenceNumber = scanNullableString(ref)
	out.PaidAt = paid
	out.RefundedAmount = "0.0000"
	_, err = tx.ExecContext(ctx, `INSERT INTO action_items(order_id,customer_user_id,title_fa,description_fa,status,priority,assigned_role_id,required_permission_code,due_at,deduplication_key,source_trigger_type) SELECT $1,$2,'بررسی پرداخت مشتری',$3,'OPEN','URGENT',r.id,'finance.customer_payments.confirm',NOW(),'payment:confirm:'||$4,'PAYMENT_CONFIRMATION_REQUIRED' FROM roles r WHERE r.code='ACCOUNTANT' ON CONFLICT(deduplication_key) WHERE deduplication_key IS NOT NULL DO NOTHING`, orderID, customer, number, out.ID)
	if err != nil {
		return out, err
	}
	if err = emitNotificationToRoleTx(ctx, tx, "ACCOUNTANT", "PAYMENT_CONFIRMATION_REQUIRED", "payment-confirmation:"+out.ID, "PAYMENT", out.ID, "/panel/dashboard/orders/"+orderID, map[string]string{}); err != nil {
		return out, err
	}
	if err = auditTx(ctx, tx, actor, "finance.payments.record", "payment", out.ID, nil, out); err != nil {
		return out, err
	}
	if err = finishOperationTx(ctx, tx, actor, "PAYMENT_RECORD", key, out); err != nil {
		return out, err
	}
	return out, tx.Commit()
}

func (s *OperationsService) ListPayments(ctx context.Context, actor, orderID string, customer bool) ([]Payment, error) {
	if customer {
		var ok bool
		if err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM orders WHERE id=$1 AND customer_user_id=$2)`, orderID, actor).Scan(&ok); err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrForbidden
		}
	}
	rows, err := s.db.QueryContext(ctx, `SELECT p.id,p.payment_number,p.order_id,p.amount::text,p.currency,p.payment_method,p.status,p.reference_number,p.paid_at,p.confirmed_at,COALESCE((SELECT SUM(r.amount) FROM payment_refunds r WHERE r.payment_id=p.id),0)::text FROM customer_payments p WHERE p.order_id=$1 AND ($2::boolean=FALSE OR p.status IN ('CONFIRMED','PARTIALLY_REFUNDED','REFUNDED')) ORDER BY p.paid_at DESC`, orderID, customer)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Payment{}
	for rows.Next() {
		var x Payment
		var ref sql.NullString
		var confirmed sql.NullTime
		if err = rows.Scan(&x.ID, &x.PaymentNumber, &x.OrderID, &x.Amount, &x.Currency, &x.PaymentMethod, &x.Status, &ref, &x.PaidAt, &confirmed, &x.RefundedAmount); err != nil {
			return nil, err
		}
		x.ReferenceNumber = scanNullableString(ref)
		if confirmed.Valid {
			x.ConfirmedAt = &confirmed.Time
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (s *OperationsService) PaymentAllocations(ctx context.Context, paymentID string) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT a.id,a.schedule_id,a.amount::text,s.title_fa,s.currency,COALESCE((SELECT SUM(ra.amount) FROM payment_refund_allocations ra WHERE ra.payment_allocation_id=a.id),0)::text FROM payment_allocations a JOIN order_payment_schedule s ON s.id=a.schedule_id WHERE a.payment_id=$1 ORDER BY s.sequence_number`, paymentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, schedule, amount, title, currency, refunded string
		if err = rows.Scan(&id, &schedule, &amount, &title, &currency, &refunded); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"id": id, "schedule_id": schedule, "amount": amount, "title_fa": title, "currency": currency, "refunded_amount": refunded})
	}
	return out, rows.Err()
}

func (s *OperationsService) ConfirmPayment(ctx context.Context, actor, paymentID, key string, p PaymentConfirmPayload) (Payment, error) {
	var out Payment
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return out, err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, "PAYMENT_CONFIRM", key, map[string]any{"payment_id": paymentID, "payload": p})
	if err != nil {
		return out, err
	}
	if claim.Existing {
		if err = json.Unmarshal(claim.Response, &out); err != nil {
			return out, err
		}
		return out, tx.Commit()
	}
	var customer, status string
	if err = tx.QueryRowContext(ctx, `SELECT payment_number,order_id,customer_user_id,amount::text,currency,payment_method,status,paid_at FROM customer_payments WHERE id=$1 FOR UPDATE`, paymentID).Scan(&out.PaymentNumber, &out.OrderID, &customer, &out.Amount, &out.Currency, &out.PaymentMethod, &status, &out.PaidAt); err != nil {
		return out, err
	}
	out.ID = paymentID
	if status != "REPORTED" && status != "PENDING_CONFIRMATION" {
		return out, conflict(ErrInvalidFinancialTransition, "payment cannot be confirmed")
	}
	allocs := p.Allocations
	if len(allocs) == 0 {
		allocs, err = autoPaymentAllocationsTx(ctx, tx, out.OrderID, out.Currency, out.Amount)
		if err != nil {
			return out, err
		}
	}
	total := "0.0000"
	seenSchedules := map[string]bool{}
	for _, a := range allocs {
		if !validPositiveDecimal(a.Amount) {
			return out, ErrValidation
		}
		if seenSchedules[a.ScheduleID] {
			return out, conflict(ErrPaymentOverAllocation, "duplicate schedule allocation")
		}
		seenSchedules[a.ScheduleID] = true
		var currency, remaining string
		if err = tx.QueryRowContext(ctx, `SELECT currency,(amount-paid_amount)::text FROM order_payment_schedule WHERE id=$1 AND order_id=$2 FOR UPDATE`, a.ScheduleID, out.OrderID).Scan(&currency, &remaining); err != nil {
			return out, err
		}
		if currency != out.Currency {
			return out, conflict(ErrCurrencyMismatch, "allocation currency does not match payment")
		}
		if cmp, _ := decimalCmp(a.Amount, remaining); cmp > 0 {
			return out, conflict(ErrPaymentOverAllocation, "allocation exceeds schedule balance")
		}
		total = addDecimal(total, a.Amount)
		_, err = tx.ExecContext(ctx, `INSERT INTO payment_allocations(payment_id,schedule_id,amount,allocated_by_user_id) VALUES($1,$2,$3::numeric,$4)`, paymentID, a.ScheduleID, a.Amount, actor)
		if err != nil {
			return out, err
		}
	}
	if cmp, _ := decimalCmp(total, out.Amount); cmp != 0 {
		return out, conflict(ErrPaymentOverAllocation, "allocations must equal payment amount")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE customer_payments SET status='CONFIRMED',confirmed_by_user_id=$2,confirmed_at=NOW(),updated_at=NOW() WHERE id=$1`, paymentID, actor); err != nil {
		return out, err
	}
	if err = recalculatePaymentScheduleTx(ctx, tx, out.OrderID); err != nil {
		return out, err
	}
	if err = s.resolvePaymentBlocksTx(ctx, tx, out.OrderID, actor); err != nil {
		return out, err
	}
	if err = refreshFinancialSummaryTx(ctx, tx, out.OrderID); err != nil {
		return out, err
	}
	out.Status = "CONFIRMED"
	now := time.Now()
	out.ConfirmedAt = &now
	out.RefundedAmount = "0.0000"
	var orderNumber string
	var salesOwner sql.NullString
	if err = tx.QueryRowContext(ctx, `SELECT order_number,sales_owner_user_id FROM orders WHERE id=$1`, out.OrderID).Scan(&orderNumber, &salesOwner); err != nil {
		return out, err
	}
	if err = emitNotificationTx(ctx, tx, customer, "PAYMENT_CONFIRMED", "payment-confirmed:"+paymentID, "PAYMENT", paymentID, "/account", map[string]string{"order_number": orderNumber, "amount": out.Amount, "currency": out.Currency}); err != nil {
		return out, err
	}
	if salesOwner.Valid {
		if err = emitNotificationTx(ctx, tx, salesOwner.String, "PAYMENT_CONFIRMED", "payment-confirmed:"+paymentID, "PAYMENT", paymentID, "/panel/dashboard/orders/"+out.OrderID, map[string]string{"order_number": orderNumber, "amount": out.Amount, "currency": out.Currency}); err != nil {
			return out, err
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE action_items SET status='COMPLETED',completed_at=NOW(),completed_by_user_id=$2,updated_at=NOW() WHERE deduplication_key='payment:confirm:'||$1 AND status='OPEN'`, paymentID, actor); err != nil {
		return out, err
	}
	if err = auditTx(ctx, tx, actor, "finance.payments.confirm", "payment", paymentID, map[string]string{"status": status}, out); err != nil {
		return out, err
	}
	if err = finishOperationTx(ctx, tx, actor, "PAYMENT_CONFIRM", key, out); err != nil {
		return out, err
	}
	return out, tx.Commit()
}

func autoPaymentAllocationsTx(ctx context.Context, tx *sql.Tx, orderID, currency, amount string) ([]PaymentAllocationPayload, error) {
	rows, err := tx.QueryContext(ctx, `SELECT id,(amount-paid_amount)::text FROM order_payment_schedule WHERE order_id=$1 AND currency=$2 AND status NOT IN ('PAID','CANCELLED') ORDER BY due_at NULLS LAST,sequence_number FOR UPDATE`, orderID, currency)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	remaining := amount
	out := []PaymentAllocationPayload{}
	for rows.Next() {
		var id, balance string
		if err = rows.Scan(&id, &balance); err != nil {
			return nil, err
		}
		if cmp, _ := decimalCmp(remaining, "0"); cmp <= 0 {
			break
		}
		take := remaining
		if cmp, _ := decimalCmp(take, balance); cmp > 0 {
			take = balance
		}
		out = append(out, PaymentAllocationPayload{ScheduleID: id, Amount: take})
		remaining = subDecimal(remaining, take)
	}
	if cmp, _ := decimalCmp(remaining, "0"); cmp > 0 {
		return nil, conflict(ErrPaymentOverAllocation, "payment exceeds outstanding schedule")
	}
	return out, rows.Err()
}

func (s *OperationsService) RejectPayment(ctx context.Context, actor, id, key, reason string) error {
	if requireReason(reason) != nil {
		return ErrValidation
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, "PAYMENT_REJECT", key, map[string]string{"id": id, "reason": reason})
	if err != nil {
		return err
	}
	if claim.Existing {
		return tx.Commit()
	}
	var status, customer, orderID string
	if err = tx.QueryRowContext(ctx, `SELECT status,customer_user_id,order_id FROM customer_payments WHERE id=$1 FOR UPDATE`, id).Scan(&status, &customer, &orderID); err != nil {
		return err
	}
	if status != "REPORTED" && status != "PENDING_CONFIRMATION" {
		return conflict(ErrInvalidFinancialTransition, "payment cannot be rejected")
	}
	r, err := tx.ExecContext(ctx, `UPDATE customer_payments SET status='REJECTED',rejected_by_user_id=$2,rejected_at=NOW(),rejection_reason=$3,updated_at=NOW() WHERE id=$1`, id, actor, reason)
	if err != nil {
		return err
	}
	if n, _ := r.RowsAffected(); n == 0 {
		return conflict(ErrInvalidFinancialTransition, "payment cannot be rejected")
	}
	if err = emitNotificationTx(ctx, tx, customer, "PAYMENT_REJECTED", "payment-rejected:"+id, "PAYMENT", id, "/account", map[string]string{}); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE action_items SET status='COMPLETED',completed_at=NOW(),completed_by_user_id=$2,updated_at=NOW() WHERE deduplication_key='payment:confirm:'||$1 AND status='OPEN'`, id, actor); err != nil {
		return err
	}
	if err = auditTx(ctx, tx, actor, "finance.payments.reject", "payment", id, nil, map[string]string{"reason": reason}); err != nil {
		return err
	}
	if err = finishOperationTx(ctx, tx, actor, "PAYMENT_REJECT", key, map[string]bool{"rejected": true}); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *OperationsService) RefundPayment(ctx context.Context, actor, id, key string, p PaymentRefundPayload) (map[string]any, error) {
	if !validPositiveDecimal(p.Amount) || requireReason(p.Reason) != nil {
		return nil, ErrValidation
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, "PAYMENT_REFUND", key, map[string]any{"id": id, "payload": p})
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
	var orderID, currency, status, total, already string
	if err = tx.QueryRowContext(ctx, `SELECT p.order_id,p.currency,p.status,p.amount::text,COALESCE((SELECT SUM(amount) FROM payment_refunds WHERE payment_id=p.id),0)::text FROM customer_payments p WHERE p.id=$1 FOR UPDATE`, id).Scan(&orderID, &currency, &status, &total, &already); err != nil {
		return nil, err
	}
	if status != "CONFIRMED" && status != "PARTIALLY_REFUNDED" {
		return nil, conflict(ErrInvalidFinancialTransition, "payment cannot be refunded")
	}
	if cmp, _ := decimalCmp(p.Amount, subDecimal(total, already)); cmp > 0 {
		return nil, conflict(ErrPaymentOverAllocation, "refund exceeds refundable amount")
	}
	number, err := nextReadableNumberTx(ctx, tx, "REF")
	if err != nil {
		return nil, err
	}
	var refundID string
	if err = tx.QueryRowContext(ctx, `INSERT INTO payment_refunds(refund_number,payment_id,amount,currency,reason,reference_number,refunded_by_user_id) VALUES($1,$2,$3::numeric,$4,$5,NULLIF($6,''),$7) RETURNING id`, number, id, p.Amount, currency, p.Reason, p.ReferenceNumber, actor).Scan(&refundID); err != nil {
		return nil, err
	}
	remaining := p.Amount
	if len(p.Allocations) > 0 {
		for _, x := range p.Allocations {
			if !validPositiveDecimal(x.Amount) {
				return nil, ErrValidation
			}
			var allocationID, available string
			err = tx.QueryRowContext(ctx, `SELECT a.id,(a.amount-COALESCE((SELECT SUM(ra.amount) FROM payment_refund_allocations ra WHERE ra.payment_allocation_id=a.id),0))::text FROM payment_allocations a WHERE a.payment_id=$1 AND a.schedule_id=$2 FOR UPDATE`, id, x.ScheduleID).Scan(&allocationID, &available)
			if err != nil {
				return nil, err
			}
			if cmp, _ := decimalCmp(x.Amount, available); cmp > 0 {
				return nil, conflict(ErrPaymentOverAllocation, "refund allocation exceeds original allocation")
			}
			_, err = tx.ExecContext(ctx, `INSERT INTO payment_refund_allocations(refund_id,payment_allocation_id,amount) VALUES($1,$2,$3::numeric)`, refundID, allocationID, x.Amount)
			if err != nil {
				return nil, err
			}
			remaining = subDecimal(remaining, x.Amount)
		}
	} else {
		rows, e := tx.QueryContext(ctx, `SELECT a.id,(a.amount-COALESCE((SELECT SUM(ra.amount) FROM payment_refund_allocations ra WHERE ra.payment_allocation_id=a.id),0))::text FROM payment_allocations a WHERE a.payment_id=$1 ORDER BY a.created_at DESC FOR UPDATE`, id)
		if e != nil {
			return nil, e
		}
		defer rows.Close()
		for rows.Next() {
			var aid, available string
			if e = rows.Scan(&aid, &available); e != nil {
				return nil, e
			}
			if cmp, _ := decimalCmp(remaining, "0"); cmp <= 0 {
				break
			}
			take := remaining
			if cmp, _ := decimalCmp(take, available); cmp > 0 {
				take = available
			}
			if _, e = tx.ExecContext(ctx, `INSERT INTO payment_refund_allocations(refund_id,payment_allocation_id,amount) VALUES($1,$2,$3::numeric)`, refundID, aid, take); e != nil {
				return nil, e
			}
			remaining = subDecimal(remaining, take)
		}
	}
	if cmp, _ := decimalCmp(remaining, "0"); cmp != 0 {
		return nil, conflict(ErrPaymentOverAllocation, "refund allocations must equal refund amount")
	}
	newRefunded := addDecimal(already, p.Amount)
	newStatus := "PARTIALLY_REFUNDED"
	if cmp, _ := decimalCmp(newRefunded, total); cmp == 0 {
		newStatus = "REFUNDED"
	}
	if _, err = tx.ExecContext(ctx, `UPDATE customer_payments SET status=$2,updated_at=NOW() WHERE id=$1`, id, newStatus); err != nil {
		return nil, err
	}
	if err = recalculatePaymentScheduleTx(ctx, tx, orderID); err != nil {
		return nil, err
	}
	if err = refreshFinancialSummaryTx(ctx, tx, orderID); err != nil {
		return nil, err
	}
	out := map[string]any{"id": refundID, "refund_number": number, "payment_id": id, "amount": p.Amount, "currency": currency, "payment_status": newStatus}
	if err = auditTx(ctx, tx, actor, "finance.payments.refund", "payment", id, nil, out); err != nil {
		return nil, err
	}
	if err = finishOperationTx(ctx, tx, actor, "PAYMENT_REFUND", key, out); err != nil {
		return nil, err
	}
	return out, tx.Commit()
}

func recalculatePaymentScheduleTx(ctx context.Context, tx *sql.Tx, orderID string) error {
	_, err := tx.ExecContext(ctx, `UPDATE order_payment_schedule s SET paid_amount=x.paid,status=CASE WHEN x.paid>=s.amount THEN 'PAID' WHEN x.paid>0 THEN 'PARTIALLY_PAID' WHEN s.due_at<NOW() THEN 'OVERDUE' WHEN s.due_at IS NOT NULL AND s.due_at<=NOW() THEN 'DUE' ELSE 'UPCOMING' END,updated_at=NOW() FROM (SELECT s2.id,GREATEST(0,COALESCE(SUM(CASE WHEN p.id IS NOT NULL THEN a.amount ELSE 0 END),0)-COALESCE(SUM(CASE WHEN p.id IS NOT NULL THEN (SELECT COALESCE(SUM(ra.amount),0) FROM payment_refund_allocations ra WHERE ra.payment_allocation_id=a.id) ELSE 0 END),0)) AS paid FROM order_payment_schedule s2 LEFT JOIN payment_allocations a ON a.schedule_id=s2.id LEFT JOIN customer_payments p ON p.id=a.payment_id AND p.status IN ('CONFIRMED','PARTIALLY_REFUNDED','REFUNDED') WHERE s2.order_id=$1 GROUP BY s2.id) x WHERE s.id=x.id`, orderID)
	return err
}

func refreshFinancialSummaryTx(ctx context.Context, tx *sql.Tx, orderID string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO order_financial_summaries(order_id,currency,revenue_amount,confirmed_payment_amount,refunded_amount,approved_cost_amount,outstanding_amount,updated_at) SELECT o.id,t.currency,CASE WHEN o.status IN ('CONFIRMED','IN_PROGRESS','COMPLETED','CLOSED') THEN t.final_customer_amount ELSE 0 END,COALESCE((SELECT SUM(p.amount) FROM customer_payments p WHERE p.order_id=o.id AND p.currency=t.currency AND p.status IN ('CONFIRMED','PARTIALLY_REFUNDED','REFUNDED')),0),COALESCE((SELECT SUM(r.amount) FROM payment_refunds r JOIN customer_payments p ON p.id=r.payment_id WHERE p.order_id=o.id AND r.currency=t.currency),0),COALESCE((SELECT SUM(c.amount) FROM operational_cost_entries c WHERE c.order_id=o.id AND c.currency=t.currency AND c.status IN ('APPROVED','PAID')),0),GREATEST(0,CASE WHEN o.status IN ('CONFIRMED','IN_PROGRESS','COMPLETED','CLOSED') THEN t.final_customer_amount ELSE 0 END-COALESCE((SELECT SUM(p.amount) FROM customer_payments p WHERE p.order_id=o.id AND p.currency=t.currency AND p.status IN ('CONFIRMED','PARTIALLY_REFUNDED','REFUNDED')),0)+COALESCE((SELECT SUM(r.amount) FROM payment_refunds r JOIN customer_payments p ON p.id=r.payment_id WHERE p.order_id=o.id AND r.currency=t.currency),0)),NOW() FROM orders o JOIN order_commercial_terms t ON t.order_id=o.id WHERE o.id=$1 ON CONFLICT(order_id) DO UPDATE SET currency=EXCLUDED.currency,revenue_amount=EXCLUDED.revenue_amount,confirmed_payment_amount=EXCLUDED.confirmed_payment_amount,refunded_amount=EXCLUDED.refunded_amount,approved_cost_amount=EXCLUDED.approved_cost_amount,outstanding_amount=EXCLUDED.outstanding_amount,updated_at=NOW()`, orderID)
	return err
}

func (s *OperationsService) resolvePaymentBlocksTx(ctx context.Context, tx *sql.Tx, orderID, actor string) error {
	rows, err := tx.QueryContext(ctx, `SELECT b.id,b.workflow_instance_id,b.workflow_step_instance_id,b.schedule_id,COALESCE(b.previous_step_status,'IN_PROGRESS') FROM workflow_payment_blocks b JOIN workflow_instances wi ON wi.id=b.workflow_instance_id JOIN order_payment_schedule ps ON ps.id=b.schedule_id WHERE wi.order_id=$1 AND b.status='OPEN' AND ps.paid_amount>=b.required_amount FOR UPDATE OF b`, orderID)
	if err != nil {
		return err
	}
	type block struct {
		id, wf             string
		step               sql.NullString
		schedule, previous string
	}
	blocks := []block{}
	for rows.Next() {
		var b block
		if err = rows.Scan(&b.id, &b.wf, &b.step, &b.schedule, &b.previous); err != nil {
			rows.Close()
			return err
		}
		blocks = append(blocks, b)
	}
	if err = rows.Close(); err != nil {
		return err
	}
	for _, b := range blocks {
		if _, err = tx.ExecContext(ctx, `UPDATE workflow_payment_blocks SET status='RESOLVED',resolved_at=NOW() WHERE id=$1`, b.id); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE action_items SET status='COMPLETED',completed_at=NOW(),completed_by_user_id=$2,updated_at=NOW() WHERE workflow_instance_id=$1 AND deduplication_key LIKE 'payment:'||$3||':%' AND status NOT IN ('COMPLETED','CANCELLED')`, b.wf, actor, b.schedule); err != nil {
			return err
		}
		if b.step.Valid {
			if b.previous == "COMPLETED" {
				if _, err = tx.ExecContext(ctx, `UPDATE workflow_step_instances SET status='IN_PROGRESS',updated_at=NOW() WHERE id=$1`, b.step.String); err != nil {
					return err
				}
				if err = s.completeStepTx(ctx, tx, actor, b.wf, b.step.String, "COMPLETED"); err != nil {
					return err
				}
			} else {
				if _, err = tx.ExecContext(ctx, `UPDATE workflow_step_instances SET status=$2,updated_at=NOW() WHERE id=$1 AND status='BLOCKED'`, b.step.String, b.previous); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *OperationsService) evaluatePaymentTriggerTx(ctx context.Context, tx *sql.Tx, orderID, event, stepID string) (int, error) {
	var workflowID, currentStep string
	var status, stepCode sql.NullString
	err := tx.QueryRowContext(ctx, `SELECT wi.id,COALESCE(NULLIF($2,'')::uuid,wi.current_step_instance_id)::text,si.status,si.step_code FROM workflow_instances wi LEFT JOIN workflow_step_instances si ON si.id=COALESCE(NULLIF($2,'')::uuid,wi.current_step_instance_id) WHERE wi.order_id=$1 AND wi.status='IN_PROGRESS' ORDER BY wi.started_at DESC LIMIT 1 FOR UPDATE OF wi`, orderID, stepID).Scan(&workflowID, &currentStep, &status, &stepCode)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT id,title_fa,(amount-paid_amount)::text,currency FROM order_payment_schedule WHERE order_id=$1 AND trigger_type=$2 AND status NOT IN ('PAID','CANCELLED') AND amount>paid_amount AND (trigger_step_code IS NULL OR trigger_step_code='' OR trigger_step_code=$3) FOR UPDATE`, orderID, event, stepCode.String)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var schedule, title, amount, currency string
		if err = rows.Scan(&schedule, &title, &amount, &currency); err != nil {
			return count, err
		}
		var blockID string
		err = tx.QueryRowContext(ctx, `INSERT INTO workflow_payment_blocks(workflow_instance_id,workflow_step_instance_id,schedule_id,trigger_type,required_amount,currency,previous_step_status) VALUES($1,$2,$3,$4,$5::numeric,$6,$7) ON CONFLICT(workflow_instance_id,workflow_step_instance_id,schedule_id,trigger_type) DO UPDATE SET required_amount=EXCLUDED.required_amount,currency=EXCLUDED.currency RETURNING id`, workflowID, currentStep, schedule, event, amount, currency, status.String).Scan(&blockID)
		if err != nil {
			return count, err
		}
		for _, role := range []string{"ACCOUNTANT", "SALES"} {
			permission := "finance.payments.view"
			priority := "HIGH"
			if role == "ACCOUNTANT" {
				permission = "finance.payments.confirm"
				priority = "URGENT"
			}
			_, err = tx.ExecContext(ctx, `INSERT INTO action_items(workflow_instance_id,workflow_step_instance_id,order_id,customer_user_id,title_fa,description_fa,status,priority,assigned_role_id,required_permission_code,due_at,deduplication_key,source_trigger_type,is_blocking) SELECT $1,$2,$3,o.customer_user_id,$4,'پرداخت موردنیاز برای ادامه فرایند','OPEN',$5,r.id,$6,NOW(),'payment:'||$7||':'||$8,'PAYMENT_BLOCK',FALSE FROM orders o JOIN roles r ON r.code=$8 WHERE o.id=$3 ON CONFLICT(deduplication_key) WHERE deduplication_key IS NOT NULL DO NOTHING`, workflowID, currentStep, orderID, title, priority, permission, schedule, role)
			if err != nil {
				return count, err
			}
		}
		if _, err = tx.ExecContext(ctx, `UPDATE workflow_step_instances SET status='BLOCKED',updated_at=NOW() WHERE id=$1 AND status NOT IN ('CANCELLED','SKIPPED')`, currentStep); err != nil {
			return count, err
		}
		count++
	}
	return count, rows.Err()
}

func (s *OperationsService) FinancialSummary(ctx context.Context, actor, orderID, reportingCurrency string, customer bool) (map[string]any, error) {
	if customer {
		var ok bool
		if err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM orders WHERE id=$1 AND customer_user_id=$2)`, orderID, actor).Scan(&ok); err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrForbidden
		}
	}
	var currency, revenue, paid, refund, cost, outstanding string
	err := s.db.QueryRowContext(ctx, `SELECT currency,revenue_amount::text,confirmed_payment_amount::text,refunded_amount::text,approved_cost_amount::text,outstanding_amount::text FROM order_financial_summaries WHERE order_id=$1`, orderID).Scan(&currency, &revenue, &paid, &refund, &cost, &outstanding)
	if err != nil {
		return nil, err
	}
	out := map[string]any{"order_id": orderID, "currency": currency, "revenue_amount": revenue, "confirmed_payment_amount": paid, "refunded_amount": refund, "outstanding_amount": outstanding}
	var pendingAmount, overdueAmount string
	var nextAmount sql.NullString
	var nextDue sql.NullTime
	err = s.db.QueryRowContext(ctx, `SELECT
		COALESCE((SELECT SUM(amount) FROM customer_payments WHERE order_id=$1 AND currency=$2 AND status IN ('REPORTED','PENDING_CONFIRMATION')),0)::text,
		COALESCE((SELECT SUM(amount-paid_amount) FROM order_payment_schedule WHERE order_id=$1 AND currency=$2 AND due_at<NOW() AND status NOT IN ('PAID','WAIVED','CANCELLED')),0)::text,
		(SELECT (amount-paid_amount)::text FROM order_payment_schedule WHERE order_id=$1 AND currency=$2 AND status NOT IN ('PAID','WAIVED','CANCELLED') ORDER BY due_at NULLS LAST,sequence_number LIMIT 1),
		(SELECT due_at FROM order_payment_schedule WHERE order_id=$1 AND currency=$2 AND status NOT IN ('PAID','WAIVED','CANCELLED') ORDER BY due_at NULLS LAST,sequence_number LIMIT 1)`, orderID, currency).Scan(&pendingAmount, &overdueAmount, &nextAmount, &nextDue)
	if err != nil {
		return nil, err
	}
	out["pending_payment_amount"] = pendingAmount
	out["overdue_amount"] = overdueAmount
	out["next_payment_amount"] = scanNullableString(nextAmount)
	out["next_payment_due_at"] = nullableTime(nextDue)
	canViewProfit := !customer && s.HasPermission(ctx, actor, "finance.profit.view")
	if canViewProfit {
		out["approved_cost_amount"] = cost
		out["profit_amount"] = subDecimal(revenue, cost)
	}
	rows, err := s.db.QueryContext(ctx, `WITH currency_set AS (
		SELECT currency FROM order_commercial_terms WHERE order_id=$1
		UNION SELECT currency FROM customer_payments WHERE order_id=$1 AND status IN ('CONFIRMED','PARTIALLY_REFUNDED','REFUNDED')
		UNION SELECT currency FROM operational_cost_entries WHERE order_id=$1 AND status IN ('APPROVED','PAID')
	) SELECT cs.currency,
		CASE WHEN cs.currency=t.currency AND o.status IN ('CONFIRMED','IN_PROGRESS','COMPLETED','CLOSED') THEN t.final_customer_amount ELSE 0 END::text,
		COALESCE((SELECT SUM(p.amount) FROM customer_payments p WHERE p.order_id=o.id AND p.currency=cs.currency AND p.status IN ('CONFIRMED','PARTIALLY_REFUNDED','REFUNDED')),0)::text,
		COALESCE((SELECT SUM(r.amount) FROM payment_refunds r JOIN customer_payments p ON p.id=r.payment_id WHERE p.order_id=o.id AND r.currency=cs.currency),0)::text,
		COALESCE((SELECT SUM(c.amount) FROM operational_cost_entries c WHERE c.order_id=o.id AND c.currency=cs.currency AND c.status IN ('APPROVED','PAID')),0)::text
	FROM currency_set cs CROSS JOIN orders o JOIN order_commercial_terms t ON t.order_id=o.id WHERE o.id=$1 ORDER BY cs.currency`, orderID)
	if err != nil {
		return nil, err
	}
	byCurrency := []map[string]any{}
	for rows.Next() {
		var code, currencyRevenue, currencyPaid, currencyRefund, currencyCost string
		if err = rows.Scan(&code, &currencyRevenue, &currencyPaid, &currencyRefund, &currencyCost); err != nil {
			rows.Close()
			return nil, err
		}
		item := map[string]any{
			"currency": code, "revenue_amount": currencyRevenue, "confirmed_payment_amount": currencyPaid,
			"refunded_amount": currencyRefund, "outstanding_amount": "0.0000",
		}
		if code == currency {
			item["outstanding_amount"] = outstanding
		}
		if canViewProfit {
			item["approved_cost_amount"] = currencyCost
			item["profit_amount"] = subDecimal(currencyRevenue, currencyCost)
		}
		byCurrency = append(byCurrency, item)
	}
	if err = rows.Close(); err != nil {
		return nil, err
	}
	out["by_currency"] = byCurrency
	reportingCurrency = normalizeCode(reportingCurrency)
	if reportingCurrency != "" && reportingCurrency != currency {
		rate, err := s.directExchangeRate(ctx, currency, reportingCurrency, time.Now())
		if err != nil {
			out["conversion_missing"] = true
			out["missing_exchange_rate"] = currency + "/" + reportingCurrency
		} else {
			out["reporting_currency"] = reportingCurrency
			out["exchange_rate"] = rate
			for _, k := range []string{"revenue_amount", "confirmed_payment_amount", "refunded_amount", "outstanding_amount", "approved_cost_amount", "profit_amount"} {
				if v, ok := out[k].(string); ok {
					x, _ := decimalMul(v, rate)
					out[k+"_converted"] = x
				}
			}
		}
	}
	if reportingCurrency != "" {
		missing := []string{}
		converted := []map[string]any{}
		for _, item := range byCurrency {
			code := item["currency"].(string)
			rate := "1.0000"
			if code != reportingCurrency {
				var rateErr error
				rate, rateErr = s.directExchangeRate(ctx, code, reportingCurrency, time.Now())
				if rateErr != nil {
					missing = append(missing, code+"/"+reportingCurrency)
					continue
				}
			}
			row := map[string]any{"source_currency": code, "reporting_currency": reportingCurrency, "exchange_rate": rate}
			for _, key := range []string{"revenue_amount", "confirmed_payment_amount", "refunded_amount", "outstanding_amount", "approved_cost_amount", "profit_amount"} {
				if value, ok := item[key].(string); ok {
					row[key], _ = decimalMul(value, rate)
				}
			}
			converted = append(converted, row)
		}
		out["converted_by_currency"] = converted
		out["missing_exchange_rates"] = uniqueReportStrings(missing)
	}
	return out, nil
}

func (s *OperationsService) directExchangeRate(ctx context.Context, base, quote string, at time.Time) (string, error) {
	var rate string
	err := s.db.QueryRowContext(ctx, `SELECT rate::text FROM exchange_rates WHERE base_currency=$1 AND quote_currency=$2 AND effective_at<=$3 ORDER BY effective_at DESC LIMIT 1`, base, quote, at).Scan(&rate)
	if err == nil {
		return rate, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	var inverse string
	if err = s.db.QueryRowContext(ctx, `SELECT rate::text FROM exchange_rates WHERE base_currency=$1 AND quote_currency=$2 AND effective_at<=$3 ORDER BY effective_at DESC LIMIT 1`, quote, base, at).Scan(&inverse); err != nil {
		return "", err
	}
	rate, ok := decimalDiv("1", inverse)
	if !ok {
		return "", fmt.Errorf("invalid exchange rate")
	}
	return rate, nil
}

func (s *OperationsService) UpsertExchangeRate(ctx context.Context, actor, key string, p ExchangeRatePayload) (map[string]any, error) {
	p.BaseCurrency = normalizeCode(p.BaseCurrency)
	p.QuoteCurrency = normalizeCode(p.QuoteCurrency)
	if p.BaseCurrency == p.QuoteCurrency || !validPositiveDecimal(p.Rate) || p.EffectiveAt.IsZero() {
		return nil, ErrValidation
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, "EXCHANGE_RATE_SAVE", key, p)
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
	var id string
	err = tx.QueryRowContext(ctx, `INSERT INTO exchange_rates(base_currency,quote_currency,rate,effective_at,notes,created_by_user_id) VALUES($1,$2,$3::numeric,$4,NULLIF($5,''),$6) ON CONFLICT(base_currency,quote_currency,effective_at) DO UPDATE SET rate=EXCLUDED.rate,notes=EXCLUDED.notes,created_by_user_id=EXCLUDED.created_by_user_id RETURNING id`, p.BaseCurrency, p.QuoteCurrency, p.Rate, p.EffectiveAt, p.Notes, actor).Scan(&id)
	if err != nil {
		return nil, err
	}
	out := map[string]any{"id": id, "base_currency": p.BaseCurrency, "quote_currency": p.QuoteCurrency, "rate": p.Rate, "effective_at": p.EffectiveAt}
	if err = finishOperationTx(ctx, tx, actor, "EXCHANGE_RATE_SAVE", key, out); err != nil {
		return nil, err
	}
	return out, tx.Commit()
}

func (s *OperationsService) ListExchangeRates(ctx context.Context) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,base_currency,quote_currency,rate::text,effective_at,COALESCE(notes,'') FROM exchange_rates ORDER BY effective_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, b, q, r, n string
		var at time.Time
		if err = rows.Scan(&id, &b, &q, &r, &at, &n); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"id": id, "base_currency": b, "quote_currency": q, "rate": r, "effective_at": at, "notes": n})
	}
	return out, rows.Err()
}

var _ = fmt.Sprintf

func (s *OperationsService) CreateProformaWithKey(ctx context.Context, actor, orderID, key, currency, subtotal, discount, notes string) (Proforma, error) {
	var out Proforma
	currency = normalizeCode(currency)
	if currency == "" {
		currency = "IRR"
	}
	if !validNonNegativeDecimal(subtotal) || !validNonNegativeDecimal(discount) {
		return out, ErrValidation
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return out, err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, "PROFORMA_CREATE", key, map[string]string{"order_id": orderID, "currency": currency, "subtotal": subtotal, "discount": discount, "notes": notes})
	if err != nil {
		return out, err
	}
	if claim.Existing {
		if err = json.Unmarshal(claim.Response, &out); err != nil {
			return out, err
		}
		return out, tx.Commit()
	}
	number, err := nextReadableNumberTx(ctx, tx, "PF")
	if err != nil {
		return out, err
	}
	snapshot := []byte("{}")
	_ = tx.QueryRowContext(ctx, `SELECT COALESCE(TO_JSONB(t),'{}'::jsonb) FROM order_commercial_terms t WHERE order_id=$1`, orderID).Scan(&snapshot)
	err = tx.QueryRowContext(ctx, `INSERT INTO proformas(proforma_number,order_id,customer_user_id,currency,subtotal,discount_amount,total_amount,notes,commercial_terms_snapshot,version_number) SELECT $2,o.id,o.customer_user_id,$3,$4::numeric,$5::numeric,GREATEST($4::numeric-$5::numeric,0),NULLIF($6,''),COALESCE($7::jsonb,'{}'::jsonb),COALESCE((SELECT MAX(version_number)+1 FROM proformas WHERE order_id=$1),1) FROM orders o WHERE o.id=$1 RETURNING id,proforma_number,order_id,status,currency,subtotal::text,discount_amount::text,total_amount::text,notes,issued_at,created_at`, orderID, number, currency, subtotal, discount, notes, string(snapshot)).Scan(&out.ID, &out.ProformaNumber, &out.OrderID, &out.Status, &out.Currency, &out.Subtotal, &out.DiscountAmount, &out.TotalAmount, &out.Notes, &out.IssuedAt, &out.CreatedAt)
	if err != nil {
		return out, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO proforma_items(proforma_id,order_item_id,line_number,description_fa,stone_snapshot,quantity,quantity_unit,unit_price,discount_amount,line_amount,currency) SELECT $1,i.id,ROW_NUMBER() OVER(ORDER BY i.created_at,i.id),i.stone_name,JSONB_BUILD_OBJECT('stone_category',i.stone_category,'stone_name',i.stone_name,'stone_variant',i.stone_variant,'finish_type',i.finish_type,'cut_type',i.cut_type,'quality_grade',i.quality_grade),i.ordered_quantity,i.quantity_unit,i.unit_price,i.discount_amount,i.line_amount,i.currency FROM order_items i WHERE i.order_id=$2`, out.ID, orderID)
	if err != nil {
		return out, err
	}
	if err = auditTx(ctx, tx, actor, "proformas.create", "proforma", out.ID, nil, map[string]any{"order_id": orderID}); err != nil {
		return out, err
	}
	if err = finishOperationTx(ctx, tx, actor, "PROFORMA_CREATE", key, out); err != nil {
		return out, err
	}
	return out, tx.Commit()
}

func (s *OperationsService) IssueProformaWithKey(ctx context.Context, actor, proformaID, key string) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	claim, err := claimOperationTx(ctx, tx, actor, "PROFORMA_ISSUE", key, map[string]string{"id": proformaID})
	if err != nil {
		return err
	}
	if claim.Existing {
		return tx.Commit()
	}
	var orderID string
	err = tx.QueryRowContext(ctx, `UPDATE proformas SET status='ISSUED',issued_by_user_id=$2,issued_at=NOW(),updated_at=NOW() WHERE id=$1 AND status='DRAFT' RETURNING order_id`, proformaID, actor).Scan(&orderID)
	if errors.Is(err, sql.ErrNoRows) {
		return conflict(ErrInvalidFinancialTransition, "proforma cannot be issued")
	}
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE orders SET status='PROFORMA_ISSUED',proforma_issued_at=NOW(),updated_at=NOW() WHERE id=$1`, orderID); err != nil {
		return err
	}
	var workflowID, stepID, stepStatus string
	err = tx.QueryRowContext(ctx, `SELECT wi.id,si.id,si.status FROM workflow_instances wi JOIN workflow_step_instances si ON si.workflow_instance_id=wi.id WHERE wi.order_id=$1 AND si.step_code='ISSUE_PROFORMA' FOR UPDATE OF si`, orderID).Scan(&workflowID, &stepID, &stepStatus)
	if err == nil {
		if stepStatus != "WAITING_FOR_ASSIGNEE" && stepStatus != "IN_PROGRESS" && stepStatus != "NEEDS_CORRECTION" {
			return ErrInvalidTransition
		}
		if stepStatus != "IN_PROGRESS" {
			if _, err = tx.ExecContext(ctx, `UPDATE workflow_step_instances SET status='IN_PROGRESS',assigned_user_id=COALESCE(assigned_user_id,$2),actual_start_at=COALESCE(actual_start_at,NOW()),updated_at=NOW() WHERE id=$1`, stepID, actor); err != nil {
				return err
			}
			if err = s.runStepTriggersTx(ctx, tx, workflowID, stepID, "ON_STEP_START"); err != nil {
				return err
			}
		}
		if err = s.runStepTriggersTx(ctx, tx, workflowID, stepID, "ON_STEP_SUBMIT"); err != nil {
			return err
		}
		if err = s.completeStepTx(ctx, tx, actor, workflowID, stepID, "COMPLETED"); err != nil {
			return err
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO order_timeline(order_id,title_fa,status_code) VALUES($1,'پیش‌فاکتور صادر شد','PROFORMA_ISSUED')`, orderID); err != nil {
		return err
	}
	var customer string
	if err = tx.QueryRowContext(ctx, `SELECT customer_user_id FROM orders WHERE id=$1`, orderID).Scan(&customer); err != nil {
		return err
	}
	if err = emitNotificationTx(ctx, tx, customer, "PROFORMA_ISSUED", "proforma-issued:"+proformaID, "PROFORMA", proformaID, "/account", map[string]string{}); err != nil {
		return err
	}
	if err = auditTx(ctx, tx, actor, "proformas.issue", "proforma", proformaID, nil, map[string]any{"order_id": orderID}); err != nil {
		return err
	}
	if err = finishOperationTx(ctx, tx, actor, "PROFORMA_ISSUE", key, map[string]bool{"issued": true}); err != nil {
		return err
	}
	return tx.Commit()
}
