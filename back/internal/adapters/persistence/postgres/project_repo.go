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
	return project, nil
}

func (r *ProjectRepository) Create(ctx context.Context, project domain.Project) (domain.Project, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO projects (description, description_en, description_fa, description_ar, cover_image_url, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, COALESCE(updated_at, created_at)
	`,
		nullableString(project.DescriptionEN),
		nullableString(project.DescriptionEN),
		nullableString(project.DescriptionFA),
		nullableString(project.DescriptionAR),
		project.CoverImageURL,
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
		    sort_order = $6,
		    updated_at = NOW()
		WHERE id = $7
		RETURNING created_at, COALESCE(updated_at, created_at)
	`,
		nullableString(project.DescriptionEN),
		nullableString(project.DescriptionEN),
		nullableString(project.DescriptionFA),
		nullableString(project.DescriptionAR),
		project.CoverImageURL,
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
