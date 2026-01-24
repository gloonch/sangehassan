package ports

import (
	"context"

	"sangehassan/back/internal/domain"
)

type AdminRepository interface {
	GetByUsername(ctx context.Context, username string) (domain.AdminUser, error)
}
