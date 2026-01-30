package postgres

import (
	"context"
	"database/sql"

	"sangehassan/back/internal/domain"
)

type BlockRepository struct {
	db *sql.DB
}

func NewBlockRepository(db *sql.DB) *BlockRepository {
	return &BlockRepository{db: db}
}

func (r *BlockRepository) List(ctx context.Context) ([]domain.Block, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT b.id, b.title_en, b.title_fa, b.title_ar, b.slug,
		       COALESCE(b.stone_type, ''), COALESCE(b.quarry, ''), COALESCE(b.dimensions, ''),
		       b.weight_ton, b.status, COALESCE(b.description, ''), COALESCE(b.image_url, ''), b.is_featured,
		       (SELECT COUNT(*) FROM block_images bi WHERE bi.block_id = b.id) AS image_count,
		       b.created_at, COALESCE(b.updated_at, b.created_at)
		FROM blocks b
		ORDER BY b.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []domain.Block
	for rows.Next() {
		var block domain.Block
		var weight sql.NullFloat64
		var imageCount int64
		if err := rows.Scan(
			&block.ID,
			&block.TitleEN,
			&block.TitleFA,
			&block.TitleAR,
			&block.Slug,
			&block.StoneType,
			&block.Quarry,
			&block.Dimensions,
			&weight,
			&block.Status,
			&block.Description,
			&block.ImageURL,
			&block.IsFeatured,
			&imageCount,
			&block.CreatedAt,
			&block.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if weight.Valid {
			block.WeightTon = weight.Float64
		}
		block.ImageCount = int(imageCount)
		if block.ImageCount == 0 && block.ImageURL != "" {
			block.ImageCount = 1
		}
		blocks = append(blocks, block)
	}
	return blocks, rows.Err()
}

func (r *BlockRepository) ListFeatured(ctx context.Context) ([]domain.Block, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT b.id, b.title_en, b.title_fa, b.title_ar, b.slug,
		       COALESCE(b.stone_type, ''), COALESCE(b.quarry, ''), COALESCE(b.dimensions, ''),
		       b.weight_ton, b.status, COALESCE(b.description, ''), COALESCE(b.image_url, ''), b.is_featured,
		       (SELECT COUNT(*) FROM block_images bi WHERE bi.block_id = b.id) AS image_count,
		       b.created_at, COALESCE(b.updated_at, b.created_at)
		FROM blocks b
		WHERE b.is_featured = TRUE
		ORDER BY b.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []domain.Block
	for rows.Next() {
		var block domain.Block
		var weight sql.NullFloat64
		var imageCount int64
		if err := rows.Scan(
			&block.ID,
			&block.TitleEN,
			&block.TitleFA,
			&block.TitleAR,
			&block.Slug,
			&block.StoneType,
			&block.Quarry,
			&block.Dimensions,
			&weight,
			&block.Status,
			&block.Description,
			&block.ImageURL,
			&block.IsFeatured,
			&imageCount,
			&block.CreatedAt,
			&block.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if weight.Valid {
			block.WeightTon = weight.Float64
		}
		block.ImageCount = int(imageCount)
		if block.ImageCount == 0 && block.ImageURL != "" {
			block.ImageCount = 1
		}
		blocks = append(blocks, block)
	}
	return blocks, rows.Err()
}

func (r *BlockRepository) GetByID(ctx context.Context, id int64) (domain.Block, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT b.id, b.title_en, b.title_fa, b.title_ar, b.slug,
		       COALESCE(b.stone_type, ''), COALESCE(b.quarry, ''), COALESCE(b.dimensions, ''),
		       b.weight_ton, b.status, COALESCE(b.description, ''), COALESCE(b.image_url, ''), b.is_featured,
		       (SELECT COUNT(*) FROM block_images bi WHERE bi.block_id = b.id) AS image_count,
		       b.created_at, COALESCE(b.updated_at, b.created_at)
		FROM blocks b
		WHERE b.id = $1
	`, id)

	var block domain.Block
	var weight sql.NullFloat64
	var imageCount int64
	if err := row.Scan(
		&block.ID,
		&block.TitleEN,
		&block.TitleFA,
		&block.TitleAR,
		&block.Slug,
		&block.StoneType,
		&block.Quarry,
		&block.Dimensions,
		&weight,
		&block.Status,
		&block.Description,
		&block.ImageURL,
		&block.IsFeatured,
		&imageCount,
		&block.CreatedAt,
		&block.UpdatedAt,
	); err != nil {
		return domain.Block{}, err
	}
	if weight.Valid {
		block.WeightTon = weight.Float64
	}
	block.ImageCount = int(imageCount)
	if err := r.loadImages(ctx, &block); err != nil {
		return domain.Block{}, err
	}
	return block, nil
}

func (r *BlockRepository) GetBySlug(ctx context.Context, slug string) (domain.Block, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT b.id, b.title_en, b.title_fa, b.title_ar, b.slug,
		       COALESCE(b.stone_type, ''), COALESCE(b.quarry, ''), COALESCE(b.dimensions, ''),
		       b.weight_ton, b.status, COALESCE(b.description, ''), COALESCE(b.image_url, ''), b.is_featured,
		       (SELECT COUNT(*) FROM block_images bi WHERE bi.block_id = b.id) AS image_count,
		       b.created_at, COALESCE(b.updated_at, b.created_at)
		FROM blocks b
		WHERE b.slug = $1
	`, slug)

	var block domain.Block
	var weight sql.NullFloat64
	var imageCount int64
	if err := row.Scan(
		&block.ID,
		&block.TitleEN,
		&block.TitleFA,
		&block.TitleAR,
		&block.Slug,
		&block.StoneType,
		&block.Quarry,
		&block.Dimensions,
		&weight,
		&block.Status,
		&block.Description,
		&block.ImageURL,
		&block.IsFeatured,
		&imageCount,
		&block.CreatedAt,
		&block.UpdatedAt,
	); err != nil {
		return domain.Block{}, err
	}
	if weight.Valid {
		block.WeightTon = weight.Float64
	}
	block.ImageCount = int(imageCount)
	if err := r.loadImages(ctx, &block); err != nil {
		return domain.Block{}, err
	}
	return block, nil
}

func (r *BlockRepository) Create(ctx context.Context, block domain.Block) (domain.Block, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO blocks (
			title_en, title_fa, title_ar, slug, stone_type, quarry, dimensions,
			weight_ton, status, description, image_url, is_featured
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, COALESCE(updated_at, created_at)
	`, block.TitleEN, block.TitleFA, block.TitleAR, block.Slug, nullableString(block.StoneType),
		nullableString(block.Quarry), nullableString(block.Dimensions), nullableFloat(block.WeightTon),
		block.Status, nullableString(block.Description), nullableString(block.ImageURL), block.IsFeatured)

	if err := row.Scan(&block.ID, &block.CreatedAt, &block.UpdatedAt); err != nil {
		return domain.Block{}, err
	}
	return block, nil
}

func (r *BlockRepository) Update(ctx context.Context, block domain.Block) (domain.Block, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE blocks
		SET title_en = $1,
		    title_fa = $2,
		    title_ar = $3,
		    slug = $4,
		    stone_type = $5,
		    quarry = $6,
		    dimensions = $7,
		    weight_ton = $8,
		    status = $9,
		    description = $10,
		    image_url = $11,
		    is_featured = $12,
		    updated_at = NOW()
		WHERE id = $13
		RETURNING created_at, COALESCE(updated_at, created_at)
	`, block.TitleEN, block.TitleFA, block.TitleAR, block.Slug, nullableString(block.StoneType),
		nullableString(block.Quarry), nullableString(block.Dimensions), nullableFloat(block.WeightTon),
		block.Status, nullableString(block.Description), nullableString(block.ImageURL), block.IsFeatured, block.ID)

	if err := row.Scan(&block.CreatedAt, &block.UpdatedAt); err != nil {
		return domain.Block{}, err
	}
	return block, nil
}

func (r *BlockRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM blocks WHERE id = $1`, id)
	return err
}

func (r *BlockRepository) ReplaceImages(ctx context.Context, blockID int64, images []string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM block_images WHERE block_id = $1`, blockID); err != nil {
		return err
	}
	for index, url := range images {
		if url == "" {
			continue
		}
		_, err := r.db.ExecContext(ctx, `
			INSERT INTO block_images (block_id, image_url, position)
			VALUES ($1, $2, $3)
			ON CONFLICT (block_id, image_url) DO UPDATE SET position = EXCLUDED.position
		`, blockID, url, index)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *BlockRepository) loadImages(ctx context.Context, block *domain.Block) error {
	rows, err := r.db.QueryContext(ctx, `
		SELECT image_url
		FROM block_images
		WHERE block_id = $1
		ORDER BY position, id
	`, block.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var images []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return err
		}
		images = append(images, url)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	block.Images = images
	if block.ImageCount == 0 {
		block.ImageCount = len(images)
	}
	if block.ImageURL == "" && len(images) > 0 {
		block.ImageURL = images[0]
	}
	return nil
}
