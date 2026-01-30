package usecase

import (
	"context"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/ports"
)

type BlockService struct {
	repo ports.BlockRepository
}

func NewBlockService(repo ports.BlockRepository) *BlockService {
	return &BlockService{repo: repo}
}

func (s *BlockService) List(ctx context.Context) ([]domain.Block, error) {
	return s.repo.List(ctx)
}

func (s *BlockService) ListFeatured(ctx context.Context) ([]domain.Block, error) {
	return s.repo.ListFeatured(ctx)
}

func (s *BlockService) GetByID(ctx context.Context, id int64) (domain.Block, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *BlockService) GetBySlug(ctx context.Context, slug string) (domain.Block, error) {
	return s.repo.GetBySlug(ctx, slug)
}

func (s *BlockService) Create(ctx context.Context, block domain.Block) (domain.Block, error) {
	if block.Slug == "" {
		block.Slug = slugify(block.TitleEN)
	}
	images := normalizeBlockImages(block.ImageURL, block.Images)
	if block.ImageURL == "" && len(images) > 0 {
		block.ImageURL = images[0]
	}

	created, err := s.repo.Create(ctx, block)
	if err != nil {
		return domain.Block{}, err
	}
	if len(images) > 0 {
		if err := s.repo.ReplaceImages(ctx, created.ID, images); err != nil {
			return domain.Block{}, err
		}
		created.Images = images
		created.ImageCount = len(images)
	}
	return created, nil
}

func (s *BlockService) Update(ctx context.Context, block domain.Block) (domain.Block, error) {
	if block.Slug == "" {
		block.Slug = slugify(block.TitleEN)
	}
	images := normalizeBlockImages(block.ImageURL, block.Images)
	if block.ImageURL == "" && len(images) > 0 {
		block.ImageURL = images[0]
	}

	updated, err := s.repo.Update(ctx, block)
	if err != nil {
		return domain.Block{}, err
	}
	if err := s.repo.ReplaceImages(ctx, updated.ID, images); err != nil {
		return domain.Block{}, err
	}
	updated.Images = images
	updated.ImageCount = len(images)
	return updated, nil
}

func (s *BlockService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func normalizeBlockImages(primary string, images []string) []string {
	if len(images) == 0 && primary != "" {
		return []string{primary}
	}
	unique := make([]string, 0, len(images))
	seen := make(map[string]struct{})
	for _, img := range images {
		if img == "" {
			continue
		}
		if _, ok := seen[img]; ok {
			continue
		}
		seen[img] = struct{}{}
		unique = append(unique, img)
	}
	return unique
}
