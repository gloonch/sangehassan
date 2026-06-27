package ports

import (
	"context"

	"sangehassan/back/internal/domain"
)

type TeamMemberRepository interface {
	List(ctx context.Context, activeOnly bool) ([]domain.TeamMember, error)
	GetByID(ctx context.Context, id int64) (domain.TeamMember, error)
	Create(ctx context.Context, member domain.TeamMember) (domain.TeamMember, error)
	Update(ctx context.Context, member domain.TeamMember) (domain.TeamMember, error)
	Delete(ctx context.Context, id int64) error
}
