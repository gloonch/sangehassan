package ports

import (
	"context"

	"sangehassan/back/internal/domain"
)

type DashboardRepository interface {
	GetStats(ctx context.Context) (domain.DashboardStats, error)
}
