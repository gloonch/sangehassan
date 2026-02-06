package usecase

import (
	"context"
	"errors"
	"time"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/ports"
)

type DealRequestService struct {
	repo        ports.DealRequestRepository
	listingRepo ports.ListingRepository
}

func NewDealRequestService(repo ports.DealRequestRepository, listingRepo ports.ListingRepository) *DealRequestService {
	return &DealRequestService{
		repo:        repo,
		listingRepo: listingRepo,
	}
}

func (s *DealRequestService) Create(ctx context.Context, req domain.DealRequest) (domain.DealRequest, error) {
	if req.ListingID == 0 {
		return domain.DealRequest{}, errors.New("listing_id is required")
	}
	if req.RequestType == "" {
		return domain.DealRequest{}, errors.New("request_type is required")
	}

	listing, err := s.listingRepo.GetByID(ctx, req.ListingID)
	if err != nil {
		return domain.DealRequest{}, err
	}
	if listing.CreatedBy != nil {
		req.SellerID = listing.CreatedBy
	}
	if req.Status == "" {
		req.Status = domain.DealStatusPendingReview
	}

	created, err := s.repo.Create(ctx, req)
	if err != nil {
		return domain.DealRequest{}, err
	}

	_ = s.repo.AddStatusHistory(ctx, domain.DealStatusHistory{
		RequestID: created.ID,
		ToStatus:  created.Status,
		CreatedBy: req.BuyerID,
	})

	return created, nil
}

func (s *DealRequestService) GetByID(ctx context.Context, id int64) (domain.DealRequest, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *DealRequestService) ListByBuyer(ctx context.Context, buyerID string, limit, offset int) ([]domain.DealRequest, error) {
	return s.repo.ListByBuyer(ctx, buyerID, limit, offset)
}

func (s *DealRequestService) ListAdmin(ctx context.Context, filter ports.DealRequestFilter) ([]domain.DealRequest, error) {
	return s.repo.ListAdmin(ctx, filter)
}

func (s *DealRequestService) UpdateStatus(ctx context.Context, requestID int64, toStatus domain.DealRequestStatus, meetingAt *time.Time, adminNote *string, actorID *string) error {
	if err := s.repo.UpdateStatus(ctx, requestID, toStatus, meetingAt, adminNote); err != nil {
		return err
	}
	return s.repo.AddStatusHistory(ctx, domain.DealStatusHistory{
		RequestID: requestID,
		ToStatus:  toStatus,
		CreatedBy: actorID,
	})
}
