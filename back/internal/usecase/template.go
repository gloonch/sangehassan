package usecase

import (
	"context"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/ports"
)

type TemplateService struct {
	repo ports.TemplateRepository
}

func NewTemplateService(repo ports.TemplateRepository) *TemplateService {
	return &TemplateService{repo: repo}
}

func (s *TemplateService) List(ctx context.Context) ([]domain.Template, error) {
	return s.repo.List(ctx)
}

func (s *TemplateService) ListActive(ctx context.Context) ([]domain.Template, error) {
	return s.repo.ListActive(ctx)
}

func (s *TemplateService) GetByID(ctx context.Context, id int64) (domain.Template, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TemplateService) Create(ctx context.Context, template domain.Template) (domain.Template, error) {
	return s.repo.Create(ctx, template)
}

func (s *TemplateService) Update(ctx context.Context, template domain.Template) (domain.Template, error) {
	return s.repo.Update(ctx, template)
}

func (s *TemplateService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
