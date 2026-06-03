package postgres

import (
	"context"
	"database/sql"

	"sangehassan/back/internal/domain"
)

type ProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) ListPublic(ctx context.Context) ([]domain.ProjectCard, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, COALESCE(cover_image_url, ''), sort_order
		FROM projects
		ORDER BY sort_order ASC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cards := make([]domain.ProjectCard, 0)
	for rows.Next() {
		var card domain.ProjectCard
		if err := rows.Scan(&card.ID, &card.CoverImageURL, &card.SortOrder); err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}
	return cards, rows.Err()
}

func (r *ProjectRepository) List(ctx context.Context) ([]domain.Project, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id,
		       COALESCE(p.description_en, p.description, ''),
		       COALESCE(p.description_fa, ''),
		       COALESCE(p.description_ar, ''),
		       COALESCE(p.cover_image_url, ''),
		       COALESCE(p.video_url, ''),
		       p.sort_order,
		       (SELECT COUNT(*) FROM project_images pi WHERE pi.project_id = p.id) AS gallery_count,
		       p.created_at,
		       COALESCE(p.updated_at, p.created_at)
		FROM projects p
		ORDER BY p.sort_order ASC, p.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := make([]domain.Project, 0)
	for rows.Next() {
		var project domain.Project
		var galleryCount int64
		if err := rows.Scan(
			&project.ID,
			&project.DescriptionEN,
			&project.DescriptionFA,
			&project.DescriptionAR,
			&project.CoverImageURL,
			&project.VideoURL,
			&project.SortOrder,
			&galleryCount,
			&project.CreatedAt,
			&project.UpdatedAt,
		); err != nil {
			return nil, err
		}
		project.Description = project.DescriptionEN
		if project.Description == "" {
			project.Description = project.DescriptionFA
		}
		if project.Description == "" {
			project.Description = project.DescriptionAR
		}
		project.GalleryCount = int(galleryCount)
		projects = append(projects, project)
	}
	return projects, rows.Err()
}

func (r *ProjectRepository) GetByID(ctx context.Context, id int64) (domain.Project, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id,
		       COALESCE(description_en, description, ''),
		       COALESCE(description_fa, ''),
		       COALESCE(description_ar, ''),
		       COALESCE(cover_image_url, ''),
		       COALESCE(video_url, ''),
		       sort_order,
		       created_at,
		       COALESCE(updated_at, created_at)
		FROM projects
		WHERE id = $1
	`, id)

	var project domain.Project
	if err := row.Scan(
		&project.ID,
		&project.DescriptionEN,
		&project.DescriptionFA,
		&project.DescriptionAR,
		&project.CoverImageURL,
		&project.VideoURL,
		&project.SortOrder,
		&project.CreatedAt,
		&project.UpdatedAt,
	); err != nil {
		return domain.Project{}, err
	}
	project.Description = project.DescriptionEN
	if project.Description == "" {
		project.Description = project.DescriptionFA
	}
	if project.Description == "" {
		project.Description = project.DescriptionAR
	}

	if err := r.loadGalleryImages(ctx, &project); err != nil {
		return domain.Project{}, err
	}
	if err := r.loadUsedProducts(ctx, &project); err != nil {
		return domain.Project{}, err
	}
	return project, nil
}

func (r *ProjectRepository) Create(ctx context.Context, project domain.Project) (domain.Project, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO projects (description, description_en, description_fa, description_ar, cover_image_url, video_url, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, COALESCE(updated_at, created_at)
	`,
		nullableString(project.DescriptionEN),
		nullableString(project.DescriptionEN),
		nullableString(project.DescriptionFA),
		nullableString(project.DescriptionAR),
		project.CoverImageURL,
		nullableString(project.VideoURL),
		project.SortOrder,
	)

	if err := row.Scan(&project.ID, &project.CreatedAt, &project.UpdatedAt); err != nil {
		return domain.Project{}, err
	}
	return project, nil
}

func (r *ProjectRepository) Update(ctx context.Context, project domain.Project) (domain.Project, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE projects
		SET description = $1,
		    description_en = $2,
		    description_fa = $3,
		    description_ar = $4,
		    cover_image_url = $5,
		    video_url = $6,
		    sort_order = $7,
		    updated_at = NOW()
		WHERE id = $8
		RETURNING created_at, COALESCE(updated_at, created_at)
	`,
		nullableString(project.DescriptionEN),
		nullableString(project.DescriptionEN),
		nullableString(project.DescriptionFA),
		nullableString(project.DescriptionAR),
		project.CoverImageURL,
		nullableString(project.VideoURL),
		project.SortOrder,
		project.ID,
	)

	if err := row.Scan(&project.CreatedAt, &project.UpdatedAt); err != nil {
		return domain.Project{}, err
	}
	return project, nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM projects WHERE id = $1`, id)
	return err
}

func (r *ProjectRepository) ReplaceGalleryImages(ctx context.Context, projectID int64, images []string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM project_images WHERE project_id = $1`, projectID); err != nil {
		return err
	}

	for index, image := range images {
		if image == "" {
			continue
		}
		_, err := r.db.ExecContext(ctx, `
			INSERT INTO project_images (project_id, image_url, position)
			VALUES ($1, $2, $3)
			ON CONFLICT (project_id, image_url)
			DO UPDATE SET position = EXCLUDED.position
		`, projectID, image, index)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ProjectRepository) ReplaceProducts(ctx context.Context, projectID int64, productIDs []int64) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM project_products WHERE project_id = $1`, projectID); err != nil {
		return err
	}

	seen := make(map[int64]struct{}, len(productIDs))
	for _, productID := range productIDs {
		if productID <= 0 {
			continue
		}
		if _, exists := seen[productID]; exists {
			continue
		}
		seen[productID] = struct{}{}
		if _, err := r.db.ExecContext(ctx, `
			INSERT INTO project_products (project_id, product_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, projectID, productID); err != nil {
			return err
		}
	}
	return nil
}

func (r *ProjectRepository) loadGalleryImages(ctx context.Context, project *domain.Project) error {
	rows, err := r.db.QueryContext(ctx, `
		SELECT image_url
		FROM project_images
		WHERE project_id = $1
		ORDER BY position, id
	`, project.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	images := make([]string, 0)
	for rows.Next() {
		var image string
		if err := rows.Scan(&image); err != nil {
			return err
		}
		images = append(images, image)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	project.GalleryImages = images
	project.GalleryCount = len(images)
	return nil
}

func (r *ProjectRepository) loadUsedProducts(ctx context.Context, project *domain.Project) error {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id,
		       p.title_en,
		       p.title_fa,
		       p.title_ar,
		       p.slug,
		       COALESCE(p.image_url, ''),
		       p.main_category_id,
		       c.id,
		       c.title_en,
		       c.title_fa,
		       c.title_ar,
		       c.slug,
		       c.parent_id
		FROM project_products pp
		JOIN products p ON p.id = pp.product_id
		LEFT JOIN categories c ON c.id = p.main_category_id
		WHERE pp.project_id = $1
		ORDER BY pp.created_at ASC, p.id ASC
	`, project.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	products := make([]domain.Product, 0)
	productIDs := make([]int64, 0)
	for rows.Next() {
		var product domain.Product
		var mainCategoryID sql.NullInt64
		var categoryID sql.NullInt64
		var categoryTitleEN sql.NullString
		var categoryTitleFA sql.NullString
		var categoryTitleAR sql.NullString
		var categorySlug sql.NullString
		var categoryParentID sql.NullInt64

		if err := rows.Scan(
			&product.ID,
			&product.TitleEN,
			&product.TitleFA,
			&product.TitleAR,
			&product.Slug,
			&product.ImageURL,
			&mainCategoryID,
			&categoryID,
			&categoryTitleEN,
			&categoryTitleFA,
			&categoryTitleAR,
			&categorySlug,
			&categoryParentID,
		); err != nil {
			return err
		}

		if mainCategoryID.Valid {
			product.MainCategoryID = &mainCategoryID.Int64
		}
		if categoryID.Valid {
			category := domain.Category{
				ID:      categoryID.Int64,
				TitleEN: categoryTitleEN.String,
				TitleFA: categoryTitleFA.String,
				TitleAR: categoryTitleAR.String,
				Slug:    categorySlug.String,
			}
			if categoryParentID.Valid {
				category.ParentID = &categoryParentID.Int64
			}
			product.Category = &category
		}
		productIDs = append(productIDs, product.ID)
		products = append(products, product)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	project.ProductIDs = productIDs
	project.UsedProducts = products
	return nil
}
