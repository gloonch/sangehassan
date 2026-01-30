package usecase

import (
	"context"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/ports"
)

type ContentSectionService struct {
	repo ports.ContentSectionRepository
}

func NewContentSectionService(repo ports.ContentSectionRepository) *ContentSectionService {
	return &ContentSectionService{repo: repo}
}

func (s *ContentSectionService) List(ctx context.Context, page string) ([]domain.ContentSection, error) {
	return s.repo.List(ctx, page)
}

func (s *ContentSectionService) GetByID(ctx context.Context, id int64) (domain.ContentSection, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ContentSectionService) Create(ctx context.Context, section domain.ContentSection) (domain.ContentSection, error) {
	images := normalizeBlockImages("", section.Images)
	section.Images = images
	created, err := s.repo.Create(ctx, section)
	if err != nil {
		return domain.ContentSection{}, err
	}
	if len(images) > 0 {
		if err := s.repo.ReplaceImages(ctx, created.ID, images); err != nil {
			return domain.ContentSection{}, err
		}
		created.Images = images
		created.ImageCount = len(images)
	}
	return created, nil
}

func (s *ContentSectionService) Update(ctx context.Context, section domain.ContentSection) (domain.ContentSection, error) {
	images := normalizeBlockImages("", section.Images)
	section.Images = images
	updated, err := s.repo.Update(ctx, section)
	if err != nil {
		return domain.ContentSection{}, err
	}
	if err := s.repo.ReplaceImages(ctx, updated.ID, images); err != nil {
		return domain.ContentSection{}, err
	}
	updated.Images = images
	updated.ImageCount = len(images)
	return updated, nil
}

func (s *ContentSectionService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
