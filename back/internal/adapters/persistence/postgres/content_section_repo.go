package postgres

import (
	"context"
	"database/sql"

	"sangehassan/back/internal/domain"
)

type ContentSectionRepository struct {
	db *sql.DB
}

func NewContentSectionRepository(db *sql.DB) *ContentSectionRepository {
	return &ContentSectionRepository{db: db}
}

func (r *ContentSectionRepository) List(ctx context.Context, page string) ([]domain.ContentSection, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, page, section_key,
		       title_en, title_fa, title_ar,
		       COALESCE(subtitle_en, ''), COALESCE(subtitle_fa, ''), COALESCE(subtitle_ar, ''),
		       COALESCE(description_en, ''), COALESCE(description_fa, ''), COALESCE(description_ar, ''),
		       COALESCE(cta_label_en, ''), COALESCE(cta_label_fa, ''), COALESCE(cta_label_ar, ''),
		       COALESCE(cta_href, ''), order_index, is_active,
		       (SELECT COUNT(*) FROM content_section_images csi WHERE csi.section_id = cs.id) AS image_count,
		       created_at, COALESCE(updated_at, created_at)
		FROM content_sections cs
		WHERE ($1 = '' OR cs.page = $1)
		ORDER BY cs.page, cs.order_index, cs.id
	`, page)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sections []domain.ContentSection
	for rows.Next() {
		var section domain.ContentSection
		var imageCount int64
		if err := rows.Scan(
			&section.ID,
			&section.Page,
			&section.Key,
			&section.TitleEN,
			&section.TitleFA,
			&section.TitleAR,
			&section.SubtitleEN,
			&section.SubtitleFA,
			&section.SubtitleAR,
			&section.DescriptionEN,
			&section.DescriptionFA,
			&section.DescriptionAR,
			&section.CTALabelEN,
			&section.CTALabelFA,
			&section.CTALabelAR,
			&section.CTAHref,
			&section.OrderIndex,
			&section.IsActive,
			&imageCount,
			&section.CreatedAt,
			&section.UpdatedAt,
		); err != nil {
			return nil, err
		}
		section.ImageCount = int(imageCount)
		sections = append(sections, section)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range sections {
		images, err := r.loadImages(ctx, sections[i].ID)
		if err != nil {
			return nil, err
		}
		sections[i].Images = images
		if sections[i].ImageCount == 0 {
			sections[i].ImageCount = len(images)
		}
	}
	return sections, nil
}

func (r *ContentSectionRepository) GetByID(ctx context.Context, id int64) (domain.ContentSection, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, page, section_key,
		       title_en, title_fa, title_ar,
		       COALESCE(subtitle_en, ''), COALESCE(subtitle_fa, ''), COALESCE(subtitle_ar, ''),
		       COALESCE(description_en, ''), COALESCE(description_fa, ''), COALESCE(description_ar, ''),
		       COALESCE(cta_label_en, ''), COALESCE(cta_label_fa, ''), COALESCE(cta_label_ar, ''),
		       COALESCE(cta_href, ''), order_index, is_active,
		       (SELECT COUNT(*) FROM content_section_images csi WHERE csi.section_id = cs.id) AS image_count,
		       created_at, COALESCE(updated_at, created_at)
		FROM content_sections cs
		WHERE id = $1
	`, id)

	var section domain.ContentSection
	var imageCount int64
	if err := row.Scan(
		&section.ID,
		&section.Page,
		&section.Key,
		&section.TitleEN,
		&section.TitleFA,
		&section.TitleAR,
		&section.SubtitleEN,
		&section.SubtitleFA,
		&section.SubtitleAR,
		&section.DescriptionEN,
		&section.DescriptionFA,
		&section.DescriptionAR,
		&section.CTALabelEN,
		&section.CTALabelFA,
		&section.CTALabelAR,
		&section.CTAHref,
		&section.OrderIndex,
		&section.IsActive,
		&imageCount,
		&section.CreatedAt,
		&section.UpdatedAt,
	); err != nil {
		return domain.ContentSection{}, err
	}
	section.ImageCount = int(imageCount)
	images, err := r.loadImages(ctx, section.ID)
	if err != nil {
		return domain.ContentSection{}, err
	}
	section.Images = images
	if section.ImageCount == 0 {
		section.ImageCount = len(images)
	}
	return section, nil
}

func (r *ContentSectionRepository) Create(ctx context.Context, section domain.ContentSection) (domain.ContentSection, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO content_sections (
			page, section_key,
			title_en, title_fa, title_ar,
			subtitle_en, subtitle_fa, subtitle_ar,
			description_en, description_fa, description_ar,
			cta_label_en, cta_label_fa, cta_label_ar,
			cta_href, order_index, is_active
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id, created_at, COALESCE(updated_at, created_at)
	`, section.Page, section.Key,
		section.TitleEN, section.TitleFA, section.TitleAR,
		nullableString(section.SubtitleEN), nullableString(section.SubtitleFA), nullableString(section.SubtitleAR),
		nullableString(section.DescriptionEN), nullableString(section.DescriptionFA), nullableString(section.DescriptionAR),
		nullableString(section.CTALabelEN), nullableString(section.CTALabelFA), nullableString(section.CTALabelAR),
		nullableString(section.CTAHref), section.OrderIndex, section.IsActive)

	if err := row.Scan(&section.ID, &section.CreatedAt, &section.UpdatedAt); err != nil {
		return domain.ContentSection{}, err
	}
	return section, nil
}

func (r *ContentSectionRepository) Update(ctx context.Context, section domain.ContentSection) (domain.ContentSection, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE content_sections
		SET page = $1,
		    section_key = $2,
		    title_en = $3,
		    title_fa = $4,
		    title_ar = $5,
		    subtitle_en = $6,
		    subtitle_fa = $7,
		    subtitle_ar = $8,
		    description_en = $9,
		    description_fa = $10,
		    description_ar = $11,
		    cta_label_en = $12,
		    cta_label_fa = $13,
		    cta_label_ar = $14,
		    cta_href = $15,
		    order_index = $16,
		    is_active = $17,
		    updated_at = NOW()
		WHERE id = $18
		RETURNING created_at, COALESCE(updated_at, created_at)
	`, section.Page, section.Key,
		section.TitleEN, section.TitleFA, section.TitleAR,
		nullableString(section.SubtitleEN), nullableString(section.SubtitleFA), nullableString(section.SubtitleAR),
		nullableString(section.DescriptionEN), nullableString(section.DescriptionFA), nullableString(section.DescriptionAR),
		nullableString(section.CTALabelEN), nullableString(section.CTALabelFA), nullableString(section.CTALabelAR),
		nullableString(section.CTAHref), section.OrderIndex, section.IsActive, section.ID)

	if err := row.Scan(&section.CreatedAt, &section.UpdatedAt); err != nil {
		return domain.ContentSection{}, err
	}
	return section, nil
}

func (r *ContentSectionRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM content_sections WHERE id = $1`, id)
	return err
}

func (r *ContentSectionRepository) ReplaceImages(ctx context.Context, sectionID int64, images []string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM content_section_images WHERE section_id = $1`, sectionID); err != nil {
		return err
	}
	for index, url := range images {
		if url == "" {
			continue
		}
		_, err := r.db.ExecContext(ctx, `
			INSERT INTO content_section_images (section_id, image_url, position)
			VALUES ($1, $2, $3)
			ON CONFLICT (section_id, image_url) DO UPDATE SET position = EXCLUDED.position
		`, sectionID, url, index)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ContentSectionRepository) loadImages(ctx context.Context, sectionID int64) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT image_url
		FROM content_section_images
		WHERE section_id = $1
		ORDER BY position, id
	`, sectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return nil, err
		}
		images = append(images, url)
	}
	return images, rows.Err()
}
