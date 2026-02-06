package ports

import (
	"context"
	"time"

	"sangehassan/back/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	GetByID(ctx context.Context, id string) (domain.User, error)
	UpdateProfile(ctx context.Context, user domain.User) (domain.User, error)
	UpdateLastLogin(ctx context.Context, id string, at time.Time) error
}
