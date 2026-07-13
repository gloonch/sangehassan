package ports

import (
	"context"

	"sangehassan/back/internal/domain"
)

type ContactSubmissionRepository interface {
	Create(ctx context.Context, submission domain.ContactSubmission) (domain.ContactSubmission, error)
	List(ctx context.Context, limit, offset int) ([]domain.ContactSubmission, error)
	GetByID(ctx context.Context, id int64) (domain.ContactSubmission, error)
}
