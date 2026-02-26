package usecase

import (
	"context"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/ports"
)

type ProjectService struct {
	repo ports.ProjectRepository
}

func NewProjectService(repo ports.ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) ListPublic(ctx context.Context) ([]domain.ProjectCard, error) {
	return s.repo.ListPublic(ctx)
}

func (s *ProjectService) List(ctx context.Context) ([]domain.Project, error) {
	return s.repo.List(ctx)
}

func (s *ProjectService) GetByID(ctx context.Context, id int64) (domain.Project, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProjectService) Create(ctx context.Context, project domain.Project) (domain.Project, error) {
	project = normalizeProjectDescriptions(project)
	gallery := normalizeProjectGallery(project.CoverImageURL, project.GalleryImages)
	created, err := s.repo.Create(ctx, project)
	if err != nil {
		return domain.Project{}, err
	}
	if err := s.repo.ReplaceGalleryImages(ctx, created.ID, gallery); err != nil {
		return domain.Project{}, err
	}
	created.GalleryImages = gallery
	created.GalleryCount = len(gallery)
	return created, nil
}

func (s *ProjectService) Update(ctx context.Context, project domain.Project) (domain.Project, error) {
	project = normalizeProjectDescriptions(project)
	gallery := normalizeProjectGallery(project.CoverImageURL, project.GalleryImages)
	updated, err := s.repo.Update(ctx, project)
	if err != nil {
		return domain.Project{}, err
	}
	if err := s.repo.ReplaceGalleryImages(ctx, updated.ID, gallery); err != nil {
		return domain.Project{}, err
	}
	updated.GalleryImages = gallery
	updated.GalleryCount = len(gallery)
	return updated, nil
}

func (s *ProjectService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func normalizeProjectGallery(cover string, images []string) []string {
	if len(images) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(images))
	normalized := make([]string, 0, len(images))
	for _, image := range images {
		if image == "" || image == cover {
			continue
		}
		if _, exists := seen[image]; exists {
			continue
		}
		seen[image] = struct{}{}
		normalized = append(normalized, image)
		if len(normalized) == 5 {
			break
		}
	}
	return normalized
}

func normalizeProjectDescriptions(project domain.Project) domain.Project {
	if project.DescriptionEN == "" {
		if project.Description != "" {
			project.DescriptionEN = project.Description
		} else if project.DescriptionFA != "" {
			project.DescriptionEN = project.DescriptionFA
		} else if project.DescriptionAR != "" {
			project.DescriptionEN = project.DescriptionAR
		}
	}
	project.Description = project.DescriptionEN
	return project
}
