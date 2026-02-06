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
		SELECT id, created_by, title, stone_type, form, tonnage,
		       province, city, price_amount, price_unit,
		       description, extra_props, status,
		       created_at, updated_at
		FROM listings
	`

	var (
		conditions []string
		args       []any
		argPos     = 1
	)

	if len(filter.Status) > 0 {
		conditions = append(conditions, fmt.Sprintf("status = ANY($%d)", argPos))
		args = append(args, pq.Array(filter.Status))
		argPos++
	}
	if filter.StoneType != "" {
		conditions = append(conditions, fmt.Sprintf("stone_type = $%d", argPos))
		args = append(args, filter.StoneType)
		argPos++
	}
	if filter.Form != "" {
		conditions = append(conditions, fmt.Sprintf("form = $%d", argPos))
		args = append(args, filter.Form)
		argPos++
	}
	if filter.Province != "" {
		conditions = append(conditions, fmt.Sprintf("province = $%d", argPos))
		args = append(args, filter.Province)
		argPos++
	}
	if filter.City != "" {
		conditions = append(conditions, fmt.Sprintf("city = $%d", argPos))
		args = append(args, filter.City)
		argPos++
	}
	if filter.MinTonnage != nil {
		conditions = append(conditions, fmt.Sprintf("tonnage >= $%d", argPos))
		args = append(args, *filter.MinTonnage)
		argPos++
	}
	if filter.MaxTonnage != nil {
		conditions = append(conditions, fmt.Sprintf("tonnage <= $%d", argPos))
		args = append(args, *filter.MaxTonnage)
		argPos++
	}
	if filter.MinPrice != nil {
		conditions = append(conditions, fmt.Sprintf("price_amount >= $%d", argPos))
		args = append(args, *filter.MinPrice)
		argPos++
	}
	if filter.MaxPrice != nil {
		conditions = append(conditions, fmt.Sprintf("price_amount <= $%d", argPos))
		args = append(args, *filter.MaxPrice)
		argPos++
	}
	if filter.Query != "" {
		conditions = append(conditions, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", argPos, argPos))
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

	var (
		listings []domain.Listing
		ids      []int64
	)
	for rows.Next() {
		listing, err := scanListing(rows)
		if err != nil {
			return nil, err
		}
		listings = append(listings, listing)
		ids = append(ids, listing.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	imagesByListing, err := r.loadImages(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range listings {
		listings[i].Images = imagesByListing[listings[i].ID]
	}

	return listings, nil
}

func (r *ListingRepository) GetByID(ctx context.Context, id int64) (domain.Listing, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, created_by, title, stone_type, form, tonnage,
		       province, city, price_amount, price_unit,
		       description, extra_props, status,
		       created_at, updated_at
		FROM listings
		WHERE id = $1
	`, id)

	listing, err := scanListing(row)
	if err != nil {
		return domain.Listing{}, err
	}

	images, err := r.loadImages(ctx, []int64{id})
	if err != nil {
		return domain.Listing{}, err
	}
	listing.Images = images[id]

	return listing, nil
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
			(created_by, title, stone_type, form, tonnage, province, city, price_amount, price_unit, description, extra_props, status)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`,
		createdBy, listing.Title, listing.StoneType, listing.Form, tonnage, listing.Province, listing.City,
		priceAmount, listing.PriceUnit, listing.Description, extraPropsJSON, listing.Status,
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
		UPDATE listings SET
			created_by = $1,
			title = $2,
			stone_type = $3,
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
		WHERE id = $13
	`
	args := []any{
		createdBy, listing.Title, listing.StoneType, listing.Form, tonnage, listing.Province, listing.City,
		priceAmount, listing.PriceUnit, listing.Description, extraPropsJSON, listing.Status, listing.ID,
	}
	if ownerID != nil {
		query += " AND created_by = $14"
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
		listing     domain.Listing
		createdBy   sql.NullString
		tonnage     sql.NullFloat64
		priceAmount sql.NullFloat64
		priceUnit   sql.NullString
		extraProps  []byte
	)

	if err := scanner.Scan(
		&listing.ID,
		&createdBy,
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
	); err != nil {
		return domain.Listing{}, err
	}

	if createdBy.Valid {
		listing.CreatedBy = &createdBy.String
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

	return listing, nil
}

func orderByFromSort(sort string) string {
	switch sort {
	case "price_asc":
		return "price_amount ASC NULLS LAST"
	case "price_desc":
		return "price_amount DESC NULLS LAST"
	case "tonnage_asc":
		return "tonnage ASC NULLS LAST"
	case "tonnage_desc":
		return "tonnage DESC NULLS LAST"
	default:
		return "created_at DESC"
	}
}
