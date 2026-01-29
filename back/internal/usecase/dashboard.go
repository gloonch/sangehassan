package usecase

import (
	"context"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/ports"
)

type DashboardService struct {
	repo ports.DashboardRepository
}

func NewDashboardService(repo ports.DashboardRepository) *DashboardService {
	return &DashboardService{repo: repo}
}

func (s *DashboardService) GetStats(ctx context.Context) (domain.DashboardStats, error) {
	return s.repo.GetStats(ctx)
}
