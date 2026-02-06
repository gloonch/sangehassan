package ports

import (
	"context"

	"sangehassan/back/internal/domain"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, token domain.RefreshToken) (domain.RefreshToken, error)
	GetByHash(ctx context.Context, hash string) (domain.RefreshToken, error)
	Revoke(ctx context.Context, id string) error
	DeleteByUser(ctx context.Context, userID string) error
}
