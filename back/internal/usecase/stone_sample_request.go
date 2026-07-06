package usecase

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/ports"
)

var (
	ErrStoneSampleUserRequired     = errors.New("user is required")
	ErrStoneSampleBoxesRequired    = errors.New("at least one complete sample box is required")
	ErrStoneSampleTooManyBoxes     = errors.New("maximum 4 sample boxes are allowed")
	ErrStoneSampleIncompleteBox    = errors.New("each sample box must contain exactly 4 stones")
	ErrStoneSampleDuplicateProduct = errors.New("each stone can only be selected once")
	ErrStoneSampleProductInvalid   = errors.New("selected stone is unavailable for samples")
	ErrStoneSampleAddressRequired  = errors.New("address is required")
	ErrStoneSamplePhoneRequired    = errors.New("phone is required")
	ErrStoneSampleShippingInvalid  = errors.New("shipping method is invalid")
	ErrStoneSampleStatusInvalid    = errors.New("status is invalid")
)

type StoneSampleRequestService struct {
	repo  ports.StoneSampleRequestRepository
	users ports.UserRepository
}

func NewStoneSampleRequestService(repo ports.StoneSampleRequestRepository, users ports.UserRepository) *StoneSampleRequestService {
	return &StoneSampleRequestService{repo: repo, users: users}
}

func (s *StoneSampleRequestService) Options(ctx context.Context, userID string) (domain.StoneSampleRequestOptions, error) {
	if strings.TrimSpace(userID) == "" {
		return domain.StoneSampleRequestOptions{}, ErrStoneSampleUserRequired
	}
	addresses, err := s.repo.ListAddresses(ctx, userID)
	if err != nil {
		return domain.StoneSampleRequestOptions{}, err
	}
	phones, err := s.repo.ListContactNumbers(ctx, userID)
	if err != nil {
		return domain.StoneSampleRequestOptions{}, err
	}
	if s.users != nil {
		if user, err := s.users.GetByID(ctx, userID); err == nil && user.Phone != nil {
			normalized := normalizePhone(*user.Phone)
			if normalized != "" && !contactListHasPhone(phones, normalized) {
				phones = append([]domain.UserContactNumber{{
					Phone:     normalized,
					Source:    "profile",
					IsDefault: len(phones) == 0,
				}}, phones...)
			}
		}
	}
	return domain.StoneSampleRequestOptions{
		Addresses:        addresses,
		Phones:           phones,
		ShippingMethods:  stoneSampleShippingMethodOptions(),
		PricePerBoxToman: domain.StoneSamplePricePerBox,
		SamplesPerBox:    domain.StoneSampleSamplesPerBox,
		MaxBoxes:         domain.StoneSampleMaxBoxes,
	}, nil
}

func (s *StoneSampleRequestService) ListSampleCategories(ctx context.Context) ([]domain.CatalogCategory, error) {
	return s.repo.ListSampleCategories(ctx)
}

func (s *StoneSampleRequestService) ListSampleProducts(ctx context.Context, filter domain.StoneSampleCatalogFilter) ([]domain.StoneSampleProduct, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 9
	}
	if filter.Limit > 30 {
		filter.Limit = 30
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	filter.Query = strings.TrimSpace(filter.Query)
	return s.repo.ListSampleProducts(ctx, filter)
}

func (s *StoneSampleRequestService) Create(ctx context.Context, input domain.CreateStoneSampleRequestInput) (domain.StoneSampleRequest, error) {
	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return domain.StoneSampleRequest{}, ErrStoneSampleUserRequired
	}
	if err := validateStoneSampleBoxes(input.Boxes); err != nil {
		return domain.StoneSampleRequest{}, err
	}
	if !isValidStoneSampleShipping(input.ShippingMethod) {
		return domain.StoneSampleRequest{}, ErrStoneSampleShippingInvalid
	}

	productIDs := flattenSampleBoxes(input.Boxes)
	products, err := s.repo.GetSampleProductsByID(ctx, productIDs)
	if err != nil {
		return domain.StoneSampleRequest{}, err
	}
	if len(products) != len(productIDs) {
		return domain.StoneSampleRequest{}, ErrStoneSampleProductInvalid
	}

	address, err := s.resolveAddress(ctx, userID, input.AddressID, input.AddressText)
	if err != nil {
		return domain.StoneSampleRequest{}, err
	}
	phone, err := s.resolvePhone(ctx, userID, input.PhoneID, input.Phone)
	if err != nil {
		return domain.StoneSampleRequest{}, err
	}

	items := make([]domain.StoneSampleRequestItem, 0, len(productIDs))
	for boxIndex, box := range input.Boxes {
		for slotIndex, productID := range box {
			product := products[productID]
			id := product.ID
			items = append(items, domain.StoneSampleRequestItem{
				ProductID:       &id,
				BoxIndex:        boxIndex + 1,
				SlotIndex:       slotIndex + 1,
				ProductTitleEN:  product.TitleEN,
				ProductTitleFA:  product.TitleFA,
				ProductTitleAR:  product.TitleAR,
				ProductSlug:     product.Slug,
				ProductImageURL: product.ImageURL,
				CategoryTitleEN: product.CategoryTitleEN,
				CategoryTitleFA: product.CategoryTitleFA,
				CategoryTitleAR: product.CategoryTitleAR,
			})
		}
	}

	boxCount := len(input.Boxes)
	req := domain.StoneSampleRequest{
		UserID:           userID,
		Status:           domain.StoneSampleStatusPending,
		BoxCount:         boxCount,
		StoneCount:       len(productIDs),
		PricePerBoxToman: domain.StoneSamplePricePerBox,
		TotalPriceToman:  int64(boxCount) * domain.StoneSamplePricePerBox,
		ShippingMethod:   strings.ToUpper(strings.TrimSpace(input.ShippingMethod)),
		AddressID:        &address.ID,
		PhoneID:          &phone.ID,
		AddressSnapshot:  address.AddressText,
		PhoneSnapshot:    phone.Phone,
	}
	return s.repo.Create(ctx, req, items)
}

func (s *StoneSampleRequestService) ListByUser(ctx context.Context, userID string, limit, offset int) ([]domain.StoneSampleRequest, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrStoneSampleUserRequired
	}
	return s.repo.ListByUser(ctx, userID, limit, offset)
}

func (s *StoneSampleRequestService) GetByUser(ctx context.Context, userID string, id int64) (domain.StoneSampleRequest, error) {
	if strings.TrimSpace(userID) == "" {
		return domain.StoneSampleRequest{}, ErrStoneSampleUserRequired
	}
	return s.repo.GetByUser(ctx, userID, id)
}

func (s *StoneSampleRequestService) ListAdmin(ctx context.Context, filter ports.StoneSampleRequestFilter) ([]domain.StoneSampleRequest, error) {
	return s.repo.ListAdmin(ctx, filter)
}

func (s *StoneSampleRequestService) GetAdmin(ctx context.Context, id int64) (domain.StoneSampleRequest, error) {
	return s.repo.GetAdmin(ctx, id)
}

func (s *StoneSampleRequestService) UpdateStatus(ctx context.Context, requestID int64, status domain.StoneSampleRequestStatus, adminNote *string, actorID *string) error {
	if !isValidStoneSampleStatus(status) {
		return ErrStoneSampleStatusInvalid
	}
	current, err := s.repo.GetAdmin(ctx, requestID)
	if err != nil {
		return err
	}
	if err := s.repo.UpdateStatus(ctx, requestID, status, adminNote); err != nil {
		return err
	}
	from := current.Status
	note := ""
	if adminNote != nil {
		note = *adminNote
	}
	return s.repo.AddStatusHistory(ctx, domain.StoneSampleStatusHistory{
		RequestID:  requestID,
		FromStatus: &from,
		ToStatus:   status,
		CreatedBy:  actorID,
		AdminNote:  note,
	})
}

func (s *StoneSampleRequestService) EnrichUsers(ctx context.Context, requests []domain.StoneSampleRequest) []domain.StoneSampleRequest {
	if s.users == nil {
		return requests
	}
	for i := range requests {
		if requests[i].UserID == "" {
			continue
		}
		user, err := s.users.GetByID(ctx, requests[i].UserID)
		if err != nil {
			continue
		}
		info := user.SafeInfo()
		requests[i].User = &info
	}
	return requests
}

func (s *StoneSampleRequestService) EnrichUser(ctx context.Context, request domain.StoneSampleRequest) domain.StoneSampleRequest {
	items := s.EnrichUsers(ctx, []domain.StoneSampleRequest{request})
	if len(items) == 0 {
		return request
	}
	return items[0]
}

func (s *StoneSampleRequestService) resolveAddress(ctx context.Context, userID string, addressID *int64, rawAddress string) (domain.UserAddress, error) {
	if addressID != nil && *addressID > 0 {
		address, err := s.repo.GetAddress(ctx, userID, *addressID)
		if err == nil {
			return address, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return domain.UserAddress{}, err
		}
	}
	addressText := strings.TrimSpace(rawAddress)
	if addressText == "" {
		return domain.UserAddress{}, ErrStoneSampleAddressRequired
	}
	return s.repo.SaveAddress(ctx, domain.UserAddress{
		UserID:      userID,
		AddressText: addressText,
		IsDefault:   false,
	})
}

func (s *StoneSampleRequestService) resolvePhone(ctx context.Context, userID string, phoneID *int64, rawPhone string) (domain.UserContactNumber, error) {
	if phoneID != nil && *phoneID > 0 {
		phone, err := s.repo.GetContactNumber(ctx, userID, *phoneID)
		if err == nil {
			return phone, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return domain.UserContactNumber{}, err
		}
	}
	normalized := normalizePhone(rawPhone)
	if normalized == "" {
		return domain.UserContactNumber{}, ErrStoneSamplePhoneRequired
	}
	return s.repo.SaveContactNumber(ctx, domain.UserContactNumber{
		UserID:    userID,
		Phone:     normalized,
		IsDefault: false,
	})
}

func validateStoneSampleBoxes(boxes [][]int64) error {
	if len(boxes) == 0 {
		return ErrStoneSampleBoxesRequired
	}
	if len(boxes) > domain.StoneSampleMaxBoxes {
		return ErrStoneSampleTooManyBoxes
	}
	seen := make(map[int64]struct{})
	for _, box := range boxes {
		if len(box) != domain.StoneSampleSamplesPerBox {
			return ErrStoneSampleIncompleteBox
		}
		for _, id := range box {
			if id <= 0 {
				return ErrStoneSampleProductInvalid
			}
			if _, ok := seen[id]; ok {
				return ErrStoneSampleDuplicateProduct
			}
			seen[id] = struct{}{}
		}
	}
	return nil
}

func flattenSampleBoxes(boxes [][]int64) []int64 {
	ids := make([]int64, 0, len(boxes)*domain.StoneSampleSamplesPerBox)
	for _, box := range boxes {
		ids = append(ids, box...)
	}
	return ids
}

func contactListHasPhone(phones []domain.UserContactNumber, phone string) bool {
	for _, item := range phones {
		if item.Phone == phone {
			return true
		}
	}
	return false
}

func stoneSampleShippingMethodOptions() []domain.StoneSampleShippingMethodOption {
	return []domain.StoneSampleShippingMethodOption{
		{Value: string(domain.StoneSampleShippingCourier), Label: "Courier"},
		{Value: string(domain.StoneSampleShippingFreight), Label: "Freight"},
		{Value: string(domain.StoneSampleShippingOperatorCoordination), Label: "Operator coordination"},
	}
}

func isValidStoneSampleShipping(value string) bool {
	switch domain.StoneSampleShippingMethod(strings.ToUpper(strings.TrimSpace(value))) {
	case domain.StoneSampleShippingCourier, domain.StoneSampleShippingFreight, domain.StoneSampleShippingOperatorCoordination:
		return true
	default:
		return false
	}
}

func isValidStoneSampleStatus(status domain.StoneSampleRequestStatus) bool {
	switch status {
	case domain.StoneSampleStatusPending,
		domain.StoneSampleStatusConfirmed,
		domain.StoneSampleStatusRejected,
		domain.StoneSampleStatusShipped,
		domain.StoneSampleStatusDelivered,
		domain.StoneSampleStatusCanceled:
		return true
	default:
		return false
	}
}
