package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/ports"
)

type StoneSampleRequestRepository struct {
	db *sql.DB
}

func NewStoneSampleRequestRepository(db *sql.DB) *StoneSampleRequestRepository {
	return &StoneSampleRequestRepository{db: db}
}

func (r *StoneSampleRequestRepository) ListSampleCategories(ctx context.Context) ([]domain.CatalogCategory, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.title_en, c.title_fa, c.title_ar, c.slug, c.parent_id,
		       COALESCE(c.image_url, ''), COALESCE(c.intro_en, ''), COALESCE(c.intro_fa, ''), COALESCE(c.intro_ar, ''),
		       COALESCE(c.seo_title_en, ''), COALESCE(c.seo_title_fa, ''), COALESCE(c.seo_title_ar, ''),
		       COALESCE(c.seo_description_en, ''), COALESCE(c.seo_description_fa, ''), COALESCE(c.seo_description_ar, ''),
		       c.is_active, c.is_indexable, c.created_at, COALESCE(c.updated_at, c.created_at),
		       COUNT(DISTINCT p.id)::int,
		       COALESCE(NULLIF(c.image_url, ''), (ARRAY_AGG(p.image_url ORDER BY p.is_popular DESC, p.id) FILTER (WHERE COALESCE(p.image_url, '') <> ''))[1], '')
		FROM categories c
		JOIN products p ON p.is_active = TRUE AND p.is_indexable = TRUE AND p.sample_available = TRUE AND (
			p.main_category_id = c.id OR EXISTS (
				SELECT 1 FROM product_categories pc WHERE pc.product_id = p.id AND pc.category_id = c.id
			)
		)
		WHERE c.is_active = TRUE
		  AND c.slug NOT IN ('accessories', 'finishings')
		GROUP BY c.id
		HAVING COUNT(DISTINCT p.id) > 0
		ORDER BY c.parent_id NULLS FIRST, c.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]domain.CatalogCategory, 0)
	for rows.Next() {
		var category domain.CatalogCategory
		if err := scanCatalogCategory(rows, &category); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, rows.Err()
}

func (r *StoneSampleRequestRepository) ListSampleProducts(ctx context.Context, filter domain.StoneSampleCatalogFilter) ([]domain.StoneSampleProduct, int, error) {
	where, args := sampleProductWhere(filter)
	var total int
	countQuery := `SELECT COUNT(DISTINCT p.id) FROM products p WHERE ` + where
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 || limit > 30 {
		limit = 9
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	args = append(args, limit, offset)
	query := fmt.Sprintf(`
		SELECT p.id, p.title_en, p.title_fa, p.title_ar, p.slug, COALESCE(p.image_url, ''),
		       p.main_category_id, COALESCE(c.slug, ''), COALESCE(c.title_en, ''), COALESCE(c.title_fa, ''), COALESCE(c.title_ar, '')
		FROM products p
		LEFT JOIN categories c ON c.id = p.main_category_id
		WHERE %s
		ORDER BY p.is_popular DESC, p.id
		LIMIT $%d OFFSET $%d
	`, where, len(args)-1, len(args))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	products := make([]domain.StoneSampleProduct, 0, limit)
	for rows.Next() {
		product, err := scanStoneSampleProduct(rows)
		if err != nil {
			return nil, 0, err
		}
		products = append(products, product)
	}
	return products, total, rows.Err()
}

func (r *StoneSampleRequestRepository) GetSampleProductsByID(ctx context.Context, ids []int64) (map[int64]domain.StoneSampleProduct, error) {
	result := make(map[int64]domain.StoneSampleProduct)
	if len(ids) == 0 {
		return result, nil
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id, p.title_en, p.title_fa, p.title_ar, p.slug, COALESCE(p.image_url, ''),
		       p.main_category_id, COALESCE(c.slug, ''), COALESCE(c.title_en, ''), COALESCE(c.title_fa, ''), COALESCE(c.title_ar, '')
		FROM products p
		LEFT JOIN categories c ON c.id = p.main_category_id
		WHERE p.id = ANY($1) AND p.is_active = TRUE AND p.is_indexable = TRUE AND p.sample_available = TRUE
	`, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		product, err := scanStoneSampleProduct(rows)
		if err != nil {
			return nil, err
		}
		result[product.ID] = product
	}
	return result, rows.Err()
}

func (r *StoneSampleRequestRepository) ListAddresses(ctx context.Context, userID string) ([]domain.UserAddress, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, COALESCE(label, ''), address_text, COALESCE(city, ''), COALESCE(province, ''),
		       COALESCE(postal_code, ''), is_default, created_at, updated_at
		FROM user_addresses
		WHERE user_id = $1
		ORDER BY is_default DESC, updated_at DESC, id DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.UserAddress, 0)
	for rows.Next() {
		item, err := scanUserAddress(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *StoneSampleRequestRepository) GetAddress(ctx context.Context, userID string, id int64) (domain.UserAddress, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, COALESCE(label, ''), address_text, COALESCE(city, ''), COALESCE(province, ''),
		       COALESCE(postal_code, ''), is_default, created_at, updated_at
		FROM user_addresses
		WHERE user_id = $1 AND id = $2
	`, userID, id)
	return scanUserAddress(row)
}

func (r *StoneSampleRequestRepository) SaveAddress(ctx context.Context, address domain.UserAddress) (domain.UserAddress, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO user_addresses (user_id, label, address_text, city, province, postal_code, is_default)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, address_text) DO UPDATE SET
		  label = COALESCE(EXCLUDED.label, user_addresses.label),
		  city = COALESCE(EXCLUDED.city, user_addresses.city),
		  province = COALESCE(EXCLUDED.province, user_addresses.province),
		  postal_code = COALESCE(EXCLUDED.postal_code, user_addresses.postal_code),
		  updated_at = NOW()
		RETURNING id, user_id, COALESCE(label, ''), address_text, COALESCE(city, ''), COALESCE(province, ''),
		          COALESCE(postal_code, ''), is_default, created_at, updated_at
	`, address.UserID, nullableString(address.Label), address.AddressText, nullableString(address.City), nullableString(address.Province), nullableString(address.PostalCode), address.IsDefault)
	return scanUserAddress(row)
}

func (r *StoneSampleRequestRepository) ListContactNumbers(ctx context.Context, userID string) ([]domain.UserContactNumber, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, COALESCE(label, ''), phone, is_default, created_at, updated_at
		FROM user_contact_numbers
		WHERE user_id = $1
		ORDER BY is_default DESC, updated_at DESC, id DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.UserContactNumber, 0)
	for rows.Next() {
		item, err := scanUserContactNumber(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *StoneSampleRequestRepository) GetContactNumber(ctx context.Context, userID string, id int64) (domain.UserContactNumber, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, COALESCE(label, ''), phone, is_default, created_at, updated_at
		FROM user_contact_numbers
		WHERE user_id = $1 AND id = $2
	`, userID, id)
	return scanUserContactNumber(row)
}

func (r *StoneSampleRequestRepository) SaveContactNumber(ctx context.Context, phone domain.UserContactNumber) (domain.UserContactNumber, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO user_contact_numbers (user_id, label, phone, is_default)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, phone) DO UPDATE SET
		  label = COALESCE(EXCLUDED.label, user_contact_numbers.label),
		  updated_at = NOW()
		RETURNING id, user_id, COALESCE(label, ''), phone, is_default, created_at, updated_at
	`, phone.UserID, nullableString(phone.Label), phone.Phone, phone.IsDefault)
	return scanUserContactNumber(row)
}

func (r *StoneSampleRequestRepository) Create(ctx context.Context, req domain.StoneSampleRequest, items []domain.StoneSampleRequestItem) (domain.StoneSampleRequest, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.StoneSampleRequest{}, err
	}
	defer tx.Rollback()

	var addressID, phoneID sql.NullInt64
	if req.AddressID != nil {
		addressID = sql.NullInt64{Int64: *req.AddressID, Valid: true}
	}
	if req.PhoneID != nil {
		phoneID = sql.NullInt64{Int64: *req.PhoneID, Valid: true}
	}

	row := tx.QueryRowContext(ctx, `
		INSERT INTO stone_sample_requests (
			user_id, status, box_count, stone_count, price_per_box_toman, total_price_toman,
			shipping_method, address_id, phone_id, address_snapshot, phone_snapshot, admin_note
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`, req.UserID, req.Status, req.BoxCount, req.StoneCount, req.PricePerBoxToman, req.TotalPriceToman,
		req.ShippingMethod, addressID, phoneID, req.AddressSnapshot, req.PhoneSnapshot, nullableString(req.AdminNote))
	if err := row.Scan(&req.ID, &req.CreatedAt, &req.UpdatedAt); err != nil {
		return domain.StoneSampleRequest{}, err
	}

	for i := range items {
		items[i].RequestID = req.ID
		var productID sql.NullInt64
		if items[i].ProductID != nil {
			productID = sql.NullInt64{Int64: *items[i].ProductID, Valid: true}
		}
		itemRow := tx.QueryRowContext(ctx, `
			INSERT INTO stone_sample_request_items (
				request_id, product_id, box_index, slot_index,
				product_title_en, product_title_fa, product_title_ar, product_slug, product_image_url,
				category_title_en, category_title_fa, category_title_ar
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			RETURNING id
		`, items[i].RequestID, productID, items[i].BoxIndex, items[i].SlotIndex,
			items[i].ProductTitleEN, items[i].ProductTitleFA, items[i].ProductTitleAR, items[i].ProductSlug, nullableString(items[i].ProductImageURL),
			nullableString(items[i].CategoryTitleEN), nullableString(items[i].CategoryTitleFA), nullableString(items[i].CategoryTitleAR))
		if err := itemRow.Scan(&items[i].ID); err != nil {
			return domain.StoneSampleRequest{}, err
		}
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO stone_sample_request_status_history (request_id, to_status)
		VALUES ($1, $2)
	`, req.ID, req.Status)
	if err != nil {
		return domain.StoneSampleRequest{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.StoneSampleRequest{}, err
	}
	req.Items = items
	return req, nil
}

func (r *StoneSampleRequestRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]domain.StoneSampleRequest, error) {
	filter := ports.StoneSampleRequestFilter{UserID: &userID, Limit: limit, Offset: offset}
	return r.list(ctx, filter, false)
}

func (r *StoneSampleRequestRepository) GetByUser(ctx context.Context, userID string, id int64) (domain.StoneSampleRequest, error) {
	row := r.db.QueryRowContext(ctx, requestSelectSQL()+` WHERE r.user_id = $1 AND r.id = $2`, userID, id)
	req, err := scanStoneSampleRequest(row)
	if err != nil {
		return domain.StoneSampleRequest{}, err
	}
	if err := r.attachRequestChildren(ctx, []*domain.StoneSampleRequest{&req}, true); err != nil {
		return domain.StoneSampleRequest{}, err
	}
	return req, nil
}

func (r *StoneSampleRequestRepository) ListAdmin(ctx context.Context, filter ports.StoneSampleRequestFilter) ([]domain.StoneSampleRequest, error) {
	return r.list(ctx, filter, false)
}

func (r *StoneSampleRequestRepository) GetAdmin(ctx context.Context, id int64) (domain.StoneSampleRequest, error) {
	row := r.db.QueryRowContext(ctx, requestSelectSQL()+` WHERE r.id = $1`, id)
	req, err := scanStoneSampleRequest(row)
	if err != nil {
		return domain.StoneSampleRequest{}, err
	}
	if err := r.attachRequestChildren(ctx, []*domain.StoneSampleRequest{&req}, true); err != nil {
		return domain.StoneSampleRequest{}, err
	}
	return req, nil
}

func (r *StoneSampleRequestRepository) UpdateStatus(ctx context.Context, requestID int64, status domain.StoneSampleRequestStatus, adminNote *string) error {
	if adminNote != nil {
		_, err := r.db.ExecContext(ctx, `
			UPDATE stone_sample_requests
			SET status = $1, admin_note = $2, updated_at = NOW()
			WHERE id = $3
		`, status, *adminNote, requestID)
		return err
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE stone_sample_requests
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`, status, requestID)
	return err
}

func (r *StoneSampleRequestRepository) AddStatusHistory(ctx context.Context, history domain.StoneSampleStatusHistory) error {
	var fromStatus, createdBy sql.NullString
	if history.FromStatus != nil {
		fromStatus = sql.NullString{String: string(*history.FromStatus), Valid: true}
	}
	if history.CreatedBy != nil {
		createdBy = sql.NullString{String: *history.CreatedBy, Valid: true}
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO stone_sample_request_status_history (request_id, from_status, to_status, created_by, admin_note)
		VALUES ($1, $2, $3, $4, $5)
	`, history.RequestID, fromStatus, history.ToStatus, createdBy, nullableString(history.AdminNote))
	return err
}

func (r *StoneSampleRequestRepository) list(ctx context.Context, filter ports.StoneSampleRequestFilter, includeHistory bool) ([]domain.StoneSampleRequest, error) {
	conditions := make([]string, 0, 2)
	args := make([]any, 0, 4)
	argPos := 1
	if filter.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("r.user_id = $%d", argPos))
		args = append(args, *filter.UserID)
		argPos++
	}
	if len(filter.Status) > 0 {
		conditions = append(conditions, fmt.Sprintf("r.status = ANY($%d)", argPos))
		args = append(args, pq.Array(filter.Status))
		argPos++
	}

	var sb strings.Builder
	sb.WriteString(requestSelectSQL())
	if len(conditions) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(conditions, " AND "))
	}
	sb.WriteString(" ORDER BY r.created_at DESC")

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
	requests := make([]domain.StoneSampleRequest, 0, limit)
	for rows.Next() {
		req, err := scanStoneSampleRequest(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	ptrs := make([]*domain.StoneSampleRequest, 0, len(requests))
	for i := range requests {
		ptrs = append(ptrs, &requests[i])
	}
	if err := r.attachRequestChildren(ctx, ptrs, includeHistory); err != nil {
		return nil, err
	}
	return requests, nil
}

func (r *StoneSampleRequestRepository) attachRequestChildren(ctx context.Context, requests []*domain.StoneSampleRequest, includeHistory bool) error {
	if len(requests) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(requests))
	byID := make(map[int64]*domain.StoneSampleRequest, len(requests))
	for _, req := range requests {
		ids = append(ids, req.ID)
		byID[req.ID] = req
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, request_id, product_id, box_index, slot_index,
		       product_title_en, product_title_fa, product_title_ar, product_slug, COALESCE(product_image_url, ''),
		       COALESCE(category_title_en, ''), COALESCE(category_title_fa, ''), COALESCE(category_title_ar, '')
		FROM stone_sample_request_items
		WHERE request_id = ANY($1)
		ORDER BY request_id, box_index, slot_index
	`, pq.Array(ids))
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item domain.StoneSampleRequestItem
		var requestID int64
		var productID sql.NullInt64
		if err := rows.Scan(
			&item.ID, &requestID, &productID, &item.BoxIndex, &item.SlotIndex,
			&item.ProductTitleEN, &item.ProductTitleFA, &item.ProductTitleAR, &item.ProductSlug, &item.ProductImageURL,
			&item.CategoryTitleEN, &item.CategoryTitleFA, &item.CategoryTitleAR,
		); err != nil {
			return err
		}
		item.RequestID = requestID
		if productID.Valid {
			item.ProductID = &productID.Int64
		}
		if req := byID[requestID]; req != nil {
			req.Items = append(req.Items, item)
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	if !includeHistory {
		return nil
	}
	historyRows, err := r.db.QueryContext(ctx, `
		SELECT id, request_id, from_status, to_status, created_by, COALESCE(admin_note, ''), created_at
		FROM stone_sample_request_status_history
		WHERE request_id = ANY($1)
		ORDER BY request_id, created_at ASC, id ASC
	`, pq.Array(ids))
	if err != nil {
		return err
	}
	defer historyRows.Close()
	for historyRows.Next() {
		var history domain.StoneSampleStatusHistory
		var requestID int64
		var fromStatus sql.NullString
		var createdBy sql.NullString
		if err := historyRows.Scan(&history.ID, &requestID, &fromStatus, &history.ToStatus, &createdBy, &history.AdminNote, &history.CreatedAt); err != nil {
			return err
		}
		history.RequestID = requestID
		if fromStatus.Valid {
			status := domain.StoneSampleRequestStatus(fromStatus.String)
			history.FromStatus = &status
		}
		if createdBy.Valid {
			history.CreatedBy = &createdBy.String
		}
		if req := byID[requestID]; req != nil {
			req.StatusHistory = append(req.StatusHistory, history)
		}
	}
	return historyRows.Err()
}

func sampleProductWhere(filter domain.StoneSampleCatalogFilter) (string, []any) {
	args := make([]any, 0, 3)
	parts := []string{
		"p.is_active = TRUE",
		"p.is_indexable = TRUE",
		"p.sample_available = TRUE",
		"NOT EXISTS (SELECT 1 FROM categories hidden_category WHERE hidden_category.id = p.main_category_id AND hidden_category.slug IN ('accessories', 'finishings'))",
	}
	if filter.CategoryID != nil {
		args = append(args, *filter.CategoryID)
		parts = append(parts, fmt.Sprintf("EXISTS (SELECT 1 FROM categories requested_category WHERE requested_category.id = $%[1]d AND requested_category.slug NOT IN ('accessories', 'finishings'))", len(args)))
		parts = append(parts, fmt.Sprintf("(p.main_category_id = $%d OR EXISTS (SELECT 1 FROM product_categories pc WHERE pc.product_id = p.id AND pc.category_id = $%d))", len(args), len(args)))
	}
	if strings.TrimSpace(filter.Query) != "" {
		args = append(args, "%"+strings.TrimSpace(filter.Query)+"%")
		parts = append(parts, fmt.Sprintf("(p.title_en ILIKE $%d OR p.title_fa ILIKE $%d OR p.title_ar ILIKE $%d OR p.slug ILIKE $%d)", len(args), len(args), len(args), len(args)))
	}
	return strings.Join(parts, " AND "), args
}

func requestSelectSQL() string {
	return `
		SELECT r.id, r.user_id, r.status, r.box_count, r.stone_count, r.price_per_box_toman, r.total_price_toman,
		       r.shipping_method, r.address_id, r.phone_id, r.address_snapshot, r.phone_snapshot, COALESCE(r.admin_note, ''),
		       r.created_at, r.updated_at
		FROM stone_sample_requests r
	`
}

func scanStoneSampleProduct(scanner interface{ Scan(dest ...any) error }) (domain.StoneSampleProduct, error) {
	var product domain.StoneSampleProduct
	var categoryID sql.NullInt64
	if err := scanner.Scan(
		&product.ID, &product.TitleEN, &product.TitleFA, &product.TitleAR, &product.Slug, &product.ImageURL,
		&categoryID, &product.CategorySlug, &product.CategoryTitleEN, &product.CategoryTitleFA, &product.CategoryTitleAR,
	); err != nil {
		return domain.StoneSampleProduct{}, err
	}
	if categoryID.Valid {
		product.CategoryID = &categoryID.Int64
	}
	return product, nil
}

func scanUserAddress(scanner interface{ Scan(dest ...any) error }) (domain.UserAddress, error) {
	var item domain.UserAddress
	if err := scanner.Scan(&item.ID, &item.UserID, &item.Label, &item.AddressText, &item.City, &item.Province, &item.PostalCode, &item.IsDefault, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return domain.UserAddress{}, err
	}
	return item, nil
}

func scanUserContactNumber(scanner interface{ Scan(dest ...any) error }) (domain.UserContactNumber, error) {
	var item domain.UserContactNumber
	if err := scanner.Scan(&item.ID, &item.UserID, &item.Label, &item.Phone, &item.IsDefault, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return domain.UserContactNumber{}, err
	}
	return item, nil
}

func scanStoneSampleRequest(scanner interface{ Scan(dest ...any) error }) (domain.StoneSampleRequest, error) {
	var req domain.StoneSampleRequest
	var addressID, phoneID sql.NullInt64
	if err := scanner.Scan(
		&req.ID, &req.UserID, &req.Status, &req.BoxCount, &req.StoneCount, &req.PricePerBoxToman, &req.TotalPriceToman,
		&req.ShippingMethod, &addressID, &phoneID, &req.AddressSnapshot, &req.PhoneSnapshot, &req.AdminNote,
		&req.CreatedAt, &req.UpdatedAt,
	); err != nil {
		return domain.StoneSampleRequest{}, err
	}
	if addressID.Valid {
		req.AddressID = &addressID.Int64
	}
	if phoneID.Valid {
		req.PhoneID = &phoneID.Int64
	}
	return req, nil
}
