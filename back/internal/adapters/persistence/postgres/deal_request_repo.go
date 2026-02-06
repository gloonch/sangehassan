package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/ports"
)

type DealRequestRepository struct {
	db *sql.DB
}

func NewDealRequestRepository(db *sql.DB) *DealRequestRepository {
	return &DealRequestRepository{db: db}
}

func (r *DealRequestRepository) Create(ctx context.Context, req domain.DealRequest) (domain.DealRequest, error) {
	if req.Status == "" {
		req.Status = domain.DealStatusPendingReview
	}

	var (
		buyerID  sql.NullString
		sellerID sql.NullString
		meeting  sql.NullTime
	)
	if req.BuyerID != nil {
		buyerID = sql.NullString{String: *req.BuyerID, Valid: true}
	}
	if req.SellerID != nil {
		sellerID = sql.NullString{String: *req.SellerID, Valid: true}
	}
	if req.MeetingAt != nil {
		meeting = sql.NullTime{Time: *req.MeetingAt, Valid: true}
	}

	row := r.db.QueryRowContext(ctx, `
		INSERT INTO deal_requests
			(listing_id, buyer_id, seller_id, request_type, buyer_note, status, meeting_at, admin_note)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`,
		req.ListingID, buyerID, sellerID, req.RequestType, req.BuyerNote, req.Status, meeting, req.AdminNote,
	)

	if err := row.Scan(&req.ID, &req.CreatedAt, &req.UpdatedAt); err != nil {
		return domain.DealRequest{}, err
	}

	return req, nil
}

func (r *DealRequestRepository) GetByID(ctx context.Context, id int64) (domain.DealRequest, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, listing_id, buyer_id, seller_id, request_type, buyer_note,
		       status, meeting_at, admin_note, created_at, updated_at
		FROM deal_requests
		WHERE id = $1
	`, id)

	req, err := scanDealRequest(row)
	if err != nil {
		return domain.DealRequest{}, err
	}

	history, err := r.loadStatusHistory(ctx, []int64{req.ID})
	if err != nil {
		return domain.DealRequest{}, err
	}
	req.StatusHistory = history[req.ID]

	return req, nil
}

func (r *DealRequestRepository) ListByBuyer(ctx context.Context, buyerID string, limit, offset int) ([]domain.DealRequest, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, listing_id, buyer_id, seller_id, request_type, buyer_note,
		       status, meeting_at, admin_note, created_at, updated_at
		FROM deal_requests
		WHERE buyer_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, buyerID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		requests []domain.DealRequest
		ids      []int64
	)
	for rows.Next() {
		req, err := scanDealRequest(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
		ids = append(ids, req.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	history, err := r.loadStatusHistory(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range requests {
		requests[i].StatusHistory = history[requests[i].ID]
	}

	return requests, nil
}

func (r *DealRequestRepository) ListAdmin(ctx context.Context, filter ports.DealRequestFilter) ([]domain.DealRequest, error) {
	var (
		conditions []string
		args       []any
		argPos     = 1
	)

	if filter.BuyerID != nil {
		conditions = append(conditions, fmt.Sprintf("buyer_id = $%d", argPos))
		args = append(args, *filter.BuyerID)
		argPos++
	}
	if filter.ListingID != nil {
		conditions = append(conditions, fmt.Sprintf("listing_id = $%d", argPos))
		args = append(args, *filter.ListingID)
		argPos++
	}
	if len(filter.Status) > 0 {
		conditions = append(conditions, fmt.Sprintf("status = ANY($%d)", argPos))
		args = append(args, pq.Array(filter.Status))
		argPos++
	}

	var sb strings.Builder
	sb.WriteString(`
		SELECT id, listing_id, buyer_id, seller_id, request_type, buyer_note,
		       status, meeting_at, admin_note, created_at, updated_at
		FROM deal_requests
	`)
	if len(conditions) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(conditions, " AND "))
	}
	sb.WriteString(" ORDER BY created_at DESC")

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	sb.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1))
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, sb.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		requests []domain.DealRequest
		ids      []int64
	)
	for rows.Next() {
		req, err := scanDealRequest(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
		ids = append(ids, req.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	history, err := r.loadStatusHistory(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range requests {
		requests[i].StatusHistory = history[requests[i].ID]
	}

	return requests, nil
}

func (r *DealRequestRepository) UpdateStatus(ctx context.Context, requestID int64, toStatus domain.DealRequestStatus, meetingAt *time.Time, adminNote *string) error {
	var (
		meeting sql.NullTime
		note    sql.NullString
	)
	if meetingAt != nil {
		meeting = sql.NullTime{Time: *meetingAt, Valid: true}
	}
	if adminNote != nil {
		note = sql.NullString{String: *adminNote, Valid: true}
	}

	_, err := r.db.ExecContext(ctx, `
		UPDATE deal_requests
		SET status = $1,
		    meeting_at = COALESCE($2, meeting_at),
		    admin_note = COALESCE($3, admin_note),
		    updated_at = NOW()
		WHERE id = $4
	`, toStatus, meeting, note, requestID)
	return err
}

func (r *DealRequestRepository) AddStatusHistory(ctx context.Context, history domain.DealStatusHistory) error {
	var createdBy sql.NullString
	if history.CreatedBy != nil {
		createdBy = sql.NullString{String: *history.CreatedBy, Valid: true}
	}
	var fromStatus sql.NullString
	if history.FromStatus != nil {
		fromStatus = sql.NullString{String: string(*history.FromStatus), Valid: true}
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO deal_status_history (request_id, from_status, to_status, created_by)
		VALUES ($1, $2, $3, $4)
	`, history.RequestID, fromStatus, history.ToStatus, createdBy)
	return err
}

func (r *DealRequestRepository) loadStatusHistory(ctx context.Context, requestIDs []int64) (map[int64][]domain.DealStatusHistory, error) {
	result := make(map[int64][]domain.DealStatusHistory)
	if len(requestIDs) == 0 {
		return result, nil
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, request_id, from_status, to_status, created_by, created_at
		FROM deal_status_history
		WHERE request_id = ANY($1)
		ORDER BY created_at ASC, id ASC
	`, pq.Array(requestIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			item       domain.DealStatusHistory
			requestID  int64
			fromStatus sql.NullString
			createdBy  sql.NullString
		)
		if err := rows.Scan(&item.ID, &requestID, &fromStatus, &item.ToStatus, &createdBy, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.RequestID = requestID
		if fromStatus.Valid {
			status := domain.DealRequestStatus(fromStatus.String)
			item.FromStatus = &status
		}
		if createdBy.Valid {
			item.CreatedBy = &createdBy.String
		}
		result[requestID] = append(result[requestID], item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func scanDealRequest(scanner interface {
	Scan(dest ...any) error
}) (domain.DealRequest, error) {
	var (
		req       domain.DealRequest
		buyerID   sql.NullString
		sellerID  sql.NullString
		meetingAt sql.NullTime
	)

	if err := scanner.Scan(
		&req.ID,
		&req.ListingID,
		&buyerID,
		&sellerID,
		&req.RequestType,
		&req.BuyerNote,
		&req.Status,
		&meetingAt,
		&req.AdminNote,
		&req.CreatedAt,
		&req.UpdatedAt,
	); err != nil {
		return domain.DealRequest{}, err
	}

	if buyerID.Valid {
		req.BuyerID = &buyerID.String
	}
	if sellerID.Valid {
		req.SellerID = &sellerID.String
	}
	if meetingAt.Valid {
		req.MeetingAt = &meetingAt.Time
	}

	return req, nil
}
