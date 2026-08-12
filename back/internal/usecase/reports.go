package usecase

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

func (s *OperationsService) ReportOverview(ctx context.Context) (map[string]any, error) {
	out := map[string]any{}
	for key, q := range map[string]string{"active_orders": `SELECT COUNT(*) FROM orders WHERE status IN ('CONFIRMED','IN_PROGRESS')`, "overdue_receivables": `SELECT COUNT(*) FROM order_payment_schedule WHERE status='OVERDUE'`, "pending_payments": `SELECT COUNT(*) FROM customer_payments WHERE status='PENDING_CONFIRMATION'`, "pending_cost_approvals": `SELECT COUNT(*) FROM operational_cost_entries WHERE status='PENDING_APPROVAL'`, "open_workflows": `SELECT COUNT(*) FROM workflow_instances WHERE status='IN_PROGRESS'`, "unissued_documents": `SELECT COUNT(*) FROM workflow_instance_document_requirements WHERE is_required AND status='PENDING'`} {
		var n int
		if err := s.db.QueryRowContext(ctx, q).Scan(&n); err != nil {
			return nil, err
		}
		out[key] = n
	}
	rows, err := s.db.QueryContext(ctx, `WITH financial_rows AS (
		SELECT t.currency,CASE WHEN o.status IN ('CONFIRMED','IN_PROGRESS','COMPLETED','CLOSED') THEN t.final_customer_amount ELSE 0 END revenue,0::numeric collected,0::numeric cost FROM order_commercial_terms t JOIN orders o ON o.id=t.order_id
		UNION ALL SELECT p.currency,0,SUM(p.amount)-COALESCE((SELECT SUM(r.amount) FROM payment_refunds r WHERE r.payment_id=p.id),0),0 FROM customer_payments p WHERE p.status IN ('CONFIRMED','PARTIALLY_REFUNDED','REFUNDED') GROUP BY p.id,p.currency
		UNION ALL SELECT c.currency,0,0,c.amount FROM operational_cost_entries c WHERE c.status IN ('APPROVED','PAID')
	) SELECT currency,SUM(revenue)::text,SUM(collected)::text,SUM(cost)::text,(SUM(revenue)-SUM(cost))::text FROM financial_rows GROUP BY currency ORDER BY currency`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	money := []map[string]string{}
	for rows.Next() {
		var c, r, p, cost, profit string
		if err = rows.Scan(&c, &r, &p, &cost, &profit); err != nil {
			return nil, err
		}
		money = append(money, map[string]string{"currency": c, "revenue": r, "collected": p, "cost": cost, "profit": profit})
	}
	out["by_currency"] = money
	return out, rows.Err()
}

func (s *OperationsService) ReportReceivables(ctx context.Context) (map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT currency,CASE WHEN due_at IS NULL OR due_at>=NOW() THEN 'upcoming' WHEN due_at>=NOW()-INTERVAL '30 days' THEN '1_30' WHEN due_at>=NOW()-INTERVAL '60 days' THEN '31_60' WHEN due_at>=NOW()-INTERVAL '90 days' THEN '61_90' ELSE 'over_90' END bucket,SUM(amount-paid_amount)::text,COUNT(*) FROM order_payment_schedule WHERE status NOT IN ('PAID','CANCELLED') GROUP BY currency,bucket ORDER BY currency,bucket`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var c, b, a string
		var n int
		if err = rows.Scan(&c, &b, &a, &n); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{"currency": c, "bucket": b, "amount": a, "count": n})
	}
	return map[string]any{"buckets": items}, rows.Err()
}

func (s *OperationsService) ReportCosts(ctx context.Context) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT currency,cost_type,status,SUM(amount)::text,COUNT(*) FROM operational_cost_entries GROUP BY currency,cost_type,status ORDER BY currency,cost_type,status`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var c, t, st, a string
		var n int
		if err = rows.Scan(&c, &t, &st, &a, &n); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"currency": c, "cost_type": t, "status": st, "amount": a, "count": n})
	}
	return out, rows.Err()
}

func (s *OperationsService) ReportProfitability(ctx context.Context, reportingCurrency string) (map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `WITH order_currencies AS (
		SELECT o.id order_id,t.currency FROM orders o JOIN order_commercial_terms t ON t.order_id=o.id WHERE o.status IN ('CONFIRMED','IN_PROGRESS','COMPLETED','CLOSED')
		UNION
		SELECT o.id,c.currency FROM orders o JOIN operational_cost_entries c ON c.order_id=o.id WHERE o.status IN ('CONFIRMED','IN_PROGRESS','COMPLETED','CLOSED') AND c.status IN ('APPROVED','PAID')
	) SELECT oc.order_id,o.order_number,COALESCE(NULLIF(TRIM(CONCAT_WS(' ',u.first_name,u.last_name)),''),u.phone_normalized),oc.currency,
		CASE WHEN oc.currency=t.currency THEN t.final_customer_amount ELSE 0 END::text,
		COALESCE((SELECT SUM(c.amount) FROM operational_cost_entries c WHERE c.order_id=oc.order_id AND c.currency=oc.currency AND c.status IN ('APPROVED','PAID')),0)::text,
		(CASE WHEN oc.currency=t.currency THEN t.final_customer_amount ELSE 0 END-COALESCE((SELECT SUM(c.amount) FROM operational_cost_entries c WHERE c.order_id=oc.order_id AND c.currency=oc.currency AND c.status IN ('APPROVED','PAID')),0))::text,
		COALESCE((SELECT SUM(c.amount) FROM operational_cost_entries c WHERE c.order_id=oc.order_id AND c.currency=oc.currency AND c.status='ESTIMATED'),0)::text,
		COALESCE((SELECT SUM(c.amount) FROM operational_cost_entries c WHERE c.order_id=oc.order_id AND c.currency=oc.currency AND c.status IN ('REPORTED','PENDING_APPROVAL')),0)::text,
		CASE WHEN oc.currency=t.currency THEN COALESCE(fs.outstanding_amount,0) ELSE 0 END::text
	FROM order_currencies oc JOIN orders o ON o.id=oc.order_id JOIN users u ON u.id=o.customer_user_id JOIN order_commercial_terms t ON t.order_id=o.id LEFT JOIN order_financial_summaries fs ON fs.order_id=o.id ORDER BY o.created_at DESC,oc.currency`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	missing := []string{}
	reportingCurrency = normalizeCode(reportingCurrency)
	for rows.Next() {
		var id, number, customer, c, revenue, cost, profit, estimated, reported, outstanding string
		if err = rows.Scan(&id, &number, &customer, &c, &revenue, &cost, &profit, &estimated, &reported, &outstanding); err != nil {
			return nil, err
		}
		item := map[string]any{"order_id": id, "order_number": number, "customer_name": customer, "currency": c, "revenue": revenue, "approved_cost": cost, "estimated_cost": estimated, "reported_cost": reported, "profit": profit, "outstanding_amount": outstanding}
		if margin, ok := decimalDiv(profit, revenue); ok {
			item["margin_percentage"], _ = decimalMul(margin, "100")
		}
		if reportingCurrency != "" && reportingCurrency != c {
			rate, e := s.directExchangeRate(ctx, c, reportingCurrency, time.Now())
			if e != nil {
				item["conversion_missing"] = true
				missing = append(missing, c+"/"+reportingCurrency)
			} else {
				item["reporting_currency"] = reportingCurrency
				item["exchange_rate"] = rate
				item["revenue_converted"], _ = decimalMul(revenue, rate)
				item["approved_cost_converted"], _ = decimalMul(cost, rate)
				item["profit_converted"], _ = decimalMul(profit, rate)
			}
		}
		items = append(items, item)
	}
	return map[string]any{"items": items, "missing_exchange_rates": uniqueReportStrings(missing)}, rows.Err()
}

func uniqueReportStrings(in []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, x := range in {
		if !seen[x] {
			seen[x] = true
			out = append(out, x)
		}
	}
	return out
}

func (s *OperationsService) ReportOperations(ctx context.Context) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT report_date,active_workflows,overdue_steps,open_discrepancies,shipments_dispatched,shipments_delivered,on_time_deliveries,average_step_duration_hours::text,rework_iterations,refreshed_at FROM daily_operations_report ORDER BY report_date DESC LIMIT 90`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var date time.Time
		var active, overdue, disc, dispatch, delivered, onTime, rework int
		var duration string
		var refreshed time.Time
		if err = rows.Scan(&date, &active, &overdue, &disc, &dispatch, &delivered, &onTime, &duration, &rework, &refreshed); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"date": date, "active_workflows": active, "overdue_steps": overdue, "open_discrepancies": disc, "shipments_dispatched": dispatch, "shipments_delivered": delivered, "on_time_deliveries": onTime, "average_step_duration_hours": duration, "rework_iterations": rework, "refreshed_at": refreshed})
	}
	return out, rows.Err()
}

func (s *OperationsService) ReportSales(ctx context.Context) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT o.id,o.order_number,o.status,o.sales_owner_user_id,COALESCE(NULLIF(TRIM(CONCAT_WS(' ',u.first_name,u.last_name)),''),u.phone_normalized),MAX(l.contacted_at),MAX(l.follow_up_at),s.outstanding_amount::text,s.currency FROM orders o JOIN users u ON u.id=o.customer_user_id LEFT JOIN customer_contact_logs l ON l.order_id=o.id LEFT JOIN order_financial_summaries s ON s.order_id=o.id WHERE o.status NOT IN ('COMPLETED','CANCELLED') GROUP BY o.id,u.id,s.order_id ORDER BY MAX(l.follow_up_at) NULLS FIRST,o.created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, num, status, customer string
		var owner sql.NullString
		var last, follow sql.NullTime
		var outstanding, currency sql.NullString
		if err = rows.Scan(&id, &num, &status, &owner, &customer, &last, &follow, &outstanding, &currency); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"order_id": id, "order_number": num, "status": status, "sales_owner_user_id": scanNullableString(owner), "customer_name": customer, "last_contact_at": nullableTime(last), "follow_up_at": nullableTime(follow), "outstanding_amount": outstanding.String, "currency": currency.String})
	}
	return out, rows.Err()
}

func (s *OperationsService) ListContactLogs(ctx context.Context, customerID, orderID string) ([]map[string]any, error) {
	query := `SELECT id,customer_user_id,order_id,contact_type,direction,COALESCE(reason_code,''),COALESCE(result_code,''),COALESCE(subject,''),summary,follow_up_at,contacted_at,created_by_user_id,created_at FROM customer_contact_logs WHERE ($1='' OR customer_user_id=$1::uuid) AND ($2='' OR order_id=$2::uuid) ORDER BY contacted_at DESC`
	rows, err := s.db.QueryContext(ctx, query, customerID, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, cid, ct, dir, reason, result, sub, sum, createdBy string
		var oid sql.NullString
		var follow sql.NullTime
		var contacted, created time.Time
		if err = rows.Scan(&id, &cid, &oid, &ct, &dir, &reason, &result, &sub, &sum, &follow, &contacted, &createdBy, &created); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"id": id, "customer_user_id": cid, "order_id": scanNullableString(oid), "contact_type": ct, "direction": dir, "reason_code": reason, "result_code": result, "subject": sub, "summary": sum, "follow_up_at": nullableTime(follow), "contacted_at": contacted, "created_by_user_id": createdBy, "created_at": created})
	}
	return out, rows.Err()
}

func (s *OperationsService) CreateContactLog(ctx context.Context, actor, customerID, orderID string, p ContactLogPayload) (map[string]any, error) {
	p.ContactType = normalizeCode(p.ContactType)
	p.Direction = normalizeCode(p.Direction)
	p.ReasonCode = normalizeCode(p.ReasonCode)
	p.ResultCode = normalizeCode(p.ResultCode)
	validContact := map[string]bool{"PHONE": true, "SMS": true, "EMAIL": true, "WHATSAPP": true, "IN_PERSON": true, "OTHER": true}
	validResult := map[string]bool{"": true, "ANSWERED": true, "NO_ANSWER": true, "CUSTOMER_CONFIRMED": true, "PAYMENT_PROMISED": true, "FOLLOW_UP_REQUIRED": true, "ISSUE_REPORTED": true, "OTHER": true}
	if strings.TrimSpace(p.Summary) == "" || !validContact[p.ContactType] || (p.Direction != "INBOUND" && p.Direction != "OUTBOUND") || !validResult[p.ResultCode] {
		return nil, ErrValidation
	}
	if orderID != "" {
		var actual string
		if err := s.db.QueryRowContext(ctx, `SELECT customer_user_id FROM orders WHERE id=$1`, orderID).Scan(&actual); err != nil {
			return nil, err
		}
		if customerID == "" {
			customerID = actual
		}
		if customerID != actual {
			return nil, ErrValidation
		}
	}
	at := time.Now()
	if p.ContactedAt != nil {
		at = *p.ContactedAt
	}
	var id string
	err := s.db.QueryRowContext(ctx, `INSERT INTO customer_contact_logs(customer_user_id,order_id,contact_type,direction,reason_code,result_code,subject,summary,follow_up_at,contacted_at,created_by_user_id) VALUES($1,NULLIF($2,'')::uuid,$3,$4,NULLIF($5,''),NULLIF($6,''),NULLIF($7,''),$8,$9,$10,$11) RETURNING id`, customerID, orderID, p.ContactType, p.Direction, p.ReasonCode, p.ResultCode, p.Subject, p.Summary, p.FollowUpAt, at, actor).Scan(&id)
	if err != nil {
		return nil, err
	}
	s.audit(ctx, actor, "customers.contacts.record", "customer", customerID, map[string]any{"contact_log_id": id, "order_id": orderID})
	return map[string]any{"id": id, "contacted_at": at}, nil
}
