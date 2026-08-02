package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lib/pq"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/ports"
)

const defaultListingLimit = 20

type ListingRepository struct {
	db *sql.DB
}

func NewListingRepository(db *sql.DB) *ListingRepository {
	return &ListingRepository{db: db}
}

func (r *ListingRepository) List(ctx context.Context, filter ports.ListingFilter) ([]domain.Listing, error) {
	base := `
		SELECT l.id, l.created_by, l.product_id, l.title, l.stone_type, l.form, l.tonnage,
		       l.province, l.city, l.price_amount, l.price_unit,
		       l.description, l.extra_props, l.status,
		       l.created_at, l.updated_at,
		       p.id, p.title_en, p.title_fa, p.title_ar, p.slug,
		       COALESCE(product_cover.image_url, p.image_url, '') AS product_image_url
		FROM listings l
		LEFT JOIN products p ON p.id = l.product_id
		LEFT JOIN LATERAL (
			SELECT image_url
			FROM product_images pi
			WHERE pi.product_id = p.id
			ORDER BY pi.position ASC, pi.id ASC
			LIMIT 1
		) product_cover ON TRUE
	`

	var (
		conditions []string
		args       []any
		argPos     = 1
	)

	if len(filter.Status) > 0 {
		conditions = append(conditions, fmt.Sprintf("l.status = ANY($%d)", argPos))
		args = append(args, pq.Array(filter.Status))
		argPos++
	}
	if filter.StoneType != "" {
		conditions = append(conditions, fmt.Sprintf("l.stone_type = $%d", argPos))
		args = append(args, filter.StoneType)
		argPos++
	}
	if filter.OwnerID != nil {
		conditions = append(conditions, fmt.Sprintf("l.created_by = $%d", argPos))
		args = append(args, *filter.OwnerID)
		argPos++
	}
	if filter.Form != "" {
		conditions = append(conditions, fmt.Sprintf("l.form = $%d", argPos))
		args = append(args, filter.Form)
		argPos++
	}
	if filter.Province != "" {
		conditions = append(conditions, fmt.Sprintf("l.province = $%d", argPos))
		args = append(args, filter.Province)
		argPos++
	}
	if filter.City != "" {
		conditions = append(conditions, fmt.Sprintf("l.city = $%d", argPos))
		args = append(args, filter.City)
		argPos++
	}
	if filter.MinTonnage != nil {
		conditions = append(conditions, fmt.Sprintf("l.tonnage >= $%d", argPos))
		args = append(args, *filter.MinTonnage)
		argPos++
	}
	if filter.MaxTonnage != nil {
		conditions = append(conditions, fmt.Sprintf("l.tonnage <= $%d", argPos))
		args = append(args, *filter.MaxTonnage)
		argPos++
	}
	if filter.MinPrice != nil {
		conditions = append(conditions, fmt.Sprintf("l.price_amount >= $%d", argPos))
		args = append(args, *filter.MinPrice)
		argPos++
	}
	if filter.MaxPrice != nil {
		conditions = append(conditions, fmt.Sprintf("l.price_amount <= $%d", argPos))
		args = append(args, *filter.MaxPrice)
		argPos++
	}
	if filter.Query != "" {
		conditions = append(conditions, fmt.Sprintf("(l.title ILIKE $%[1]d OR l.description ILIKE $%[1]d OR p.title_en ILIKE $%[1]d OR p.title_fa ILIKE $%[1]d OR p.title_ar ILIKE $%[1]d)", argPos))
		args = append(args, "%"+filter.Query+"%")
		argPos++
	}

	var sb strings.Builder
	sb.WriteString(base)
	if len(conditions) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(conditions, " AND "))
	}

	sb.WriteString(" ORDER BY ")
	sb.WriteString(orderByFromSort(filter.Sort))

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = defaultListingLimit
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

	var listings []domain.Listing
	for rows.Next() {
		listing, err := scanListing(rows)
		if err != nil {
			return nil, err
		}
		listings = append(listings, listing)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return listings, nil
}

func (r *ListingRepository) ListLiveFeed(ctx context.Context, limit int) ([]domain.ListingLiveFeedItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT l.id, COALESCE(l.title, ''), COALESCE(l.stone_type, ''), COALESCE(l.form, ''),
		       l.tonnage, l.created_at,
		       p.id, COALESCE(p.title_en, ''), COALESCE(p.title_fa, ''),
		       COALESCE(p.title_ar, ''), COALESCE(p.slug, '')
		FROM listings l
		JOIN products p ON p.id = l.product_id
		WHERE l.status = $1
		  AND p.is_active = TRUE
		  AND COALESCE(
		        NULLIF(BTRIM(l.title), ''),
		        NULLIF(BTRIM(p.title_en), ''),
		        NULLIF(BTRIM(p.title_fa), ''),
		        NULLIF(BTRIM(p.title_ar), '')
		      ) IS NOT NULL
		ORDER BY l.created_at DESC, l.id DESC
		LIMIT $2
	`, domain.ListingStatusActive, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.ListingLiveFeedItem, 0, limit)
	for rows.Next() {
		var (
			item     domain.ListingLiveFeedItem
			quantity sql.NullFloat64
		)
		if err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.StoneType,
			&item.Form,
			&quantity,
			&item.PublishedAt,
			&item.Product.ID,
			&item.Product.TitleEN,
			&item.Product.TitleFA,
			&item.Product.TitleAR,
			&item.Product.Slug,
		); err != nil {
			return nil, err
		}
		if quantity.Valid {
			item.Quantity = &quantity.Float64
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ListingRepository) GetByID(ctx context.Context, id int64) (domain.Listing, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT l.id, l.created_by, l.product_id, l.title, l.stone_type, l.form, l.tonnage,
		       l.province, l.city, l.price_amount, l.price_unit,
		       l.description, l.extra_props, l.status,
		       l.created_at, l.updated_at,
		       p.id, p.title_en, p.title_fa, p.title_ar, p.slug,
		       COALESCE(product_cover.image_url, p.image_url, '') AS product_image_url
		FROM listings l
		LEFT JOIN products p ON p.id = l.product_id
		LEFT JOIN LATERAL (
			SELECT image_url
			FROM product_images pi
			WHERE pi.product_id = p.id
			ORDER BY pi.position ASC, pi.id ASC
			LIMIT 1
		) product_cover ON TRUE
		WHERE l.id = $1
	`, id)

	return scanListing(row)
}

func (r *ListingRepository) Create(ctx context.Context, listing domain.Listing) (domain.Listing, error) {
	if listing.Status == "" {
		listing.Status = domain.ListingStatusActive
	}
	if listing.PriceUnit == "" {
		listing.PriceUnit = domain.PriceUnitNegotiable
	}
	if listing.ExtraProps == nil {
		listing.ExtraProps = map[string]any{}
	}

	extraPropsJSON, err := json.Marshal(listing.ExtraProps)
	if err != nil {
		return domain.Listing{}, err
	}

	var (
		createdBy   sql.NullString
		tonnage     sql.NullFloat64
		priceAmount sql.NullFloat64
	)
	if listing.CreatedBy != nil {
		createdBy = sql.NullString{String: *listing.CreatedBy, Valid: true}
	}
	if listing.Tonnage != nil {
		tonnage = sql.NullFloat64{Float64: *listing.Tonnage, Valid: true}
	}
	if listing.PriceAmount != nil {
		priceAmount = sql.NullFloat64{Float64: *listing.PriceAmount, Valid: true}
	}

	row := r.db.QueryRowContext(ctx, `
		INSERT INTO listings
			(created_by, product_id, title, stone_type, form, tonnage, province, city, price_amount, price_unit, description, extra_props, status)
		SELECT
			$1, p.id, $2, COALESCE(NULLIF($3, ''), c.slug, ''), $4, $5, $6, $7, $8, $9, $10, $11, $12
		FROM products p
		LEFT JOIN categories c ON c.id = p.main_category_id
		WHERE p.id = $13 AND p.is_active = TRUE
		RETURNING id, created_at, updated_at
	`,
		createdBy, listing.Title, listing.StoneType, listing.Form, tonnage, listing.Province, listing.City,
		priceAmount, listing.PriceUnit, listing.Description, extraPropsJSON, listing.Status, listing.ProductID,
	)
	if err := row.Scan(&listing.ID, &listing.CreatedAt, &listing.UpdatedAt); err != nil {
		return domain.Listing{}, err
	}
	return listing, nil
}

func (r *ListingRepository) Update(ctx context.Context, listing domain.Listing, ownerID *string) (domain.Listing, error) {
	if listing.Status == "" {
		listing.Status = domain.ListingStatusActive
	}
	if listing.PriceUnit == "" {
		listing.PriceUnit = domain.PriceUnitNegotiable
	}
	if listing.ExtraProps == nil {
		listing.ExtraProps = map[string]any{}
	}

	extraPropsJSON, err := json.Marshal(listing.ExtraProps)
	if err != nil {
		return domain.Listing{}, err
	}

	var (
		createdBy   sql.NullString
		tonnage     sql.NullFloat64
		priceAmount sql.NullFloat64
	)
	if listing.CreatedBy != nil {
		createdBy = sql.NullString{String: *listing.CreatedBy, Valid: true}
	}
	if listing.Tonnage != nil {
		tonnage = sql.NullFloat64{Float64: *listing.Tonnage, Valid: true}
	}
	if listing.PriceAmount != nil {
		priceAmount = sql.NullFloat64{Float64: *listing.PriceAmount, Valid: true}
	}

	query := `
		UPDATE listings l SET
			created_by = $1,
			product_id = p.id,
			title = $2,
			stone_type = COALESCE(NULLIF($3, ''), c.slug, ''),
			form = $4,
			tonnage = $5,
			province = $6,
			city = $7,
			price_amount = $8,
			price_unit = $9,
			description = $10,
			extra_props = $11,
			status = $12,
			updated_at = NOW()
		FROM products p
		LEFT JOIN categories c ON c.id = p.main_category_id
		WHERE l.id = $13 AND p.id = $14 AND p.is_active = TRUE
	`
	args := []any{
		createdBy, listing.Title, listing.StoneType, listing.Form, tonnage, listing.Province, listing.City,
		priceAmount, listing.PriceUnit, listing.Description, extraPropsJSON, listing.Status, listing.ID, listing.ProductID,
	}
	if ownerID != nil {
		query += " AND l.created_by = $15"
		args = append(args, *ownerID)
	}

	row := r.db.QueryRowContext(ctx, query+` RETURNING id, created_at, updated_at`, args...)
	if err := row.Scan(&listing.ID, &listing.CreatedAt, &listing.UpdatedAt); err != nil {
		return domain.Listing{}, err
	}
	return listing, nil
}

func (r *ListingRepository) ReplaceImages(ctx context.Context, listingID int64, images []domain.ListingImage) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, `DELETE FROM listing_images WHERE listing_id = $1`, listingID); err != nil {
		return err
	}

	for i, img := range images {
		position := img.Position
		if position == 0 {
			position = i
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO listing_images (listing_id, image_url, position)
			VALUES ($1, $2, $3)
		`, listingID, img.ImageURL, position); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *ListingRepository) Delete(ctx context.Context, id int64, ownerID *string) error {
	if ownerID == nil {
		_, err := r.db.ExecContext(ctx, `DELETE FROM listings WHERE id = $1`, id)
		return err
	}
	res, err := r.db.ExecContext(ctx, `DELETE FROM listings WHERE id = $1 AND created_by = $2`, id, *ownerID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *ListingRepository) CreateProductRequest(ctx context.Context, request domain.ListingProductRequest) (domain.ListingProductRequest, error) {
	var userID sql.NullString
	if request.UserID != nil && strings.TrimSpace(*request.UserID) != "" {
		userID = sql.NullString{String: *request.UserID, Valid: true}
	}
	if request.Status == "" {
		request.Status = "NEW"
	}

	row := r.db.QueryRowContext(ctx, `
		INSERT INTO listing_product_requests (user_id, query, status)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`, userID, request.Query, request.Status)
	if err := row.Scan(&request.ID, &request.CreatedAt); err != nil {
		return domain.ListingProductRequest{}, err
	}
	return request, nil
}

func (r *ListingRepository) ListProductRequests(ctx context.Context, limit, offset int) ([]domain.ListingProductRequest, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT r.id, r.user_id, r.query, r.status, r.created_at,
		       u.id, u.email, u.full_name, u.phone, u.role, u.created_at
		FROM listing_product_requests r
		LEFT JOIN users u ON u.id = r.user_id
		ORDER BY r.created_at DESC, r.id DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := make([]domain.ListingProductRequest, 0)
	for rows.Next() {
		var (
			request       domain.ListingProductRequest
			userID        sql.NullString
			userInfo      domain.UserInfo
			joinedUserID  sql.NullString
			userEmail     sql.NullString
			userName      sql.NullString
			userPhone     sql.NullString
			userRole      sql.NullString
			userCreatedAt sql.NullTime
		)
		if err := rows.Scan(
			&request.ID,
			&userID,
			&request.Query,
			&request.Status,
			&request.CreatedAt,
			&joinedUserID,
			&userEmail,
			&userName,
			&userPhone,
			&userRole,
			&userCreatedAt,
		); err != nil {
			return nil, err
		}
		if userID.Valid {
			request.UserID = &userID.String
		}
		if joinedUserID.Valid {
			userInfo.ID = joinedUserID.String
			userInfo.Email = userEmail.String
			if userName.Valid {
				userInfo.FullName = &userName.String
			}
			if userPhone.Valid {
				userInfo.Phone = &userPhone.String
			}
			userInfo.Role = userRole.String
			if userCreatedAt.Valid {
				userInfo.CreatedAt = userCreatedAt.Time
			}
			request.User = &userInfo
		}
		requests = append(requests, request)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return requests, nil
}

func (r *ListingRepository) loadImages(ctx context.Context, listingIDs []int64) (map[int64][]domain.ListingImage, error) {
	result := make(map[int64][]domain.ListingImage)
	if len(listingIDs) == 0 {
		return result, nil
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, listing_id, image_url, position, created_at
		FROM listing_images
		WHERE listing_id = ANY($1)
		ORDER BY position ASC, id ASC
	`, pq.Array(listingIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			img       domain.ListingImage
			listingID int64
		)
		if err := rows.Scan(&img.ID, &listingID, &img.ImageURL, &img.Position, &img.CreatedAt); err != nil {
			return nil, err
		}
		img.ListingID = listingID
		result[listingID] = append(result[listingID], img)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func scanListing(scanner interface {
	Scan(dest ...any) error
}) (domain.Listing, error) {
	var (
		listing         domain.Listing
		createdBy       sql.NullString
		productID       sql.NullInt64
		tonnage         sql.NullFloat64
		priceAmount     sql.NullFloat64
		priceUnit       sql.NullString
		extraProps      []byte
		product         domain.ListingProduct
		joinedProductID sql.NullInt64
		productTitleEN  sql.NullString
		productTitleFA  sql.NullString
		productTitleAR  sql.NullString
		productSlug     sql.NullString
		productImageURL sql.NullString
	)

	if err := scanner.Scan(
		&listing.ID,
		&createdBy,
		&productID,
		&listing.Title,
		&listing.StoneType,
		&listing.Form,
		&tonnage,
		&listing.Province,
		&listing.City,
		&priceAmount,
		&priceUnit,
		&listing.Description,
		&extraProps,
		&listing.Status,
		&listing.CreatedAt,
		&listing.UpdatedAt,
		&joinedProductID,
		&productTitleEN,
		&productTitleFA,
		&productTitleAR,
		&productSlug,
		&productImageURL,
	); err != nil {
		return domain.Listing{}, err
	}

	if createdBy.Valid {
		listing.CreatedBy = &createdBy.String
	}
	if productID.Valid {
		listing.ProductID = productID.Int64
	}
	if tonnage.Valid {
		listing.Tonnage = &tonnage.Float64
	}
	if priceAmount.Valid {
		listing.PriceAmount = &priceAmount.Float64
	}
	if priceUnit.Valid {
		listing.PriceUnit = priceUnit.String
	}
	if len(extraProps) > 0 {
		if err := json.Unmarshal(extraProps, &listing.ExtraProps); err != nil {
			return domain.Listing{}, err
		}
	}
	if listing.ExtraProps == nil {
		listing.ExtraProps = map[string]any{}
	}
	if joinedProductID.Valid {
		product.ID = joinedProductID.Int64
		product.TitleEN = productTitleEN.String
		product.TitleFA = productTitleFA.String
		product.TitleAR = productTitleAR.String
		product.Slug = productSlug.String
		product.ImageURL = productImageURL.String
		listing.Product = &product
		if product.ImageURL != "" {
			listing.Images = []domain.ListingImage{{
				ListingID: listing.ID,
				ImageURL:  product.ImageURL,
				Position:  0,
			}}
		}
	}

	return listing, nil
}

func orderByFromSort(sort string) string {
	switch sort {
	case "price_asc":
		return "l.price_amount ASC NULLS LAST"
	case "price_desc":
		return "l.price_amount DESC NULLS LAST"
	case "tonnage_asc":
		return "l.tonnage ASC NULLS LAST"
	case "tonnage_desc":
		return "l.tonnage DESC NULLS LAST"
	default:
		return "l.created_at DESC"
	}
}
