package usecase

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/ports"
)

func TestStoneSampleCreateValidBoxCounts(t *testing.T) {
	for _, boxCount := range []int{1, 2, 3, 4} {
		t.Run("boxes", func(t *testing.T) {
			repo := newStoneSampleFakeRepo(16)
			service := NewStoneSampleRequestService(repo, nil)
			boxes := make([][]int64, 0, boxCount)
			var id int64 = 1
			for boxIndex := 0; boxIndex < boxCount; boxIndex++ {
				box := make([]int64, 0, domain.StoneSampleSamplesPerBox)
				for slotIndex := 0; slotIndex < domain.StoneSampleSamplesPerBox; slotIndex++ {
					box = append(box, id)
					id++
				}
				boxes = append(boxes, box)
			}

			created, err := service.Create(context.Background(), validStoneSampleInput(boxes))
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}
			if created.BoxCount != boxCount {
				t.Fatalf("BoxCount = %d, want %d", created.BoxCount, boxCount)
			}
			wantTotal := int64(boxCount) * domain.StoneSamplePricePerBox
			if created.TotalPriceToman != wantTotal {
				t.Fatalf("TotalPriceToman = %d, want %d", created.TotalPriceToman, wantTotal)
			}
		})
	}
}

func TestStoneSampleCreateRejectsDuplicateProduct(t *testing.T) {
	service := NewStoneSampleRequestService(newStoneSampleFakeRepo(4), nil)
	_, err := service.Create(context.Background(), validStoneSampleInput([][]int64{{1, 2, 3, 1}}))
	if !errors.Is(err, ErrStoneSampleDuplicateProduct) {
		t.Fatalf("Create() error = %v, want %v", err, ErrStoneSampleDuplicateProduct)
	}
}

func TestStoneSampleCreateRejectsIncompleteBox(t *testing.T) {
	service := NewStoneSampleRequestService(newStoneSampleFakeRepo(5), nil)
	_, err := service.Create(context.Background(), validStoneSampleInput([][]int64{{1, 2, 3, 4}, {5}}))
	if !errors.Is(err, ErrStoneSampleIncompleteBox) {
		t.Fatalf("Create() error = %v, want %v", err, ErrStoneSampleIncompleteBox)
	}
}

func TestStoneSampleCreateRejectsTooManyBoxes(t *testing.T) {
	service := NewStoneSampleRequestService(newStoneSampleFakeRepo(20), nil)
	_, err := service.Create(context.Background(), validStoneSampleInput([][]int64{
		{1, 2, 3, 4},
		{5, 6, 7, 8},
		{9, 10, 11, 12},
		{13, 14, 15, 16},
		{17, 18, 19, 20},
	}))
	if !errors.Is(err, ErrStoneSampleTooManyBoxes) {
		t.Fatalf("Create() error = %v, want %v", err, ErrStoneSampleTooManyBoxes)
	}
}

func TestStoneSampleCreateRejectsUnavailableProduct(t *testing.T) {
	service := NewStoneSampleRequestService(newStoneSampleFakeRepo(3), nil)
	_, err := service.Create(context.Background(), validStoneSampleInput([][]int64{{1, 2, 3, 4}}))
	if !errors.Is(err, ErrStoneSampleProductInvalid) {
		t.Fatalf("Create() error = %v, want %v", err, ErrStoneSampleProductInvalid)
	}
}

func TestStoneSampleCreateRequiresContactAndShipping(t *testing.T) {
	service := NewStoneSampleRequestService(newStoneSampleFakeRepo(4), nil)
	input := validStoneSampleInput([][]int64{{1, 2, 3, 4}})
	input.AddressText = ""
	if _, err := service.Create(context.Background(), input); !errors.Is(err, ErrStoneSampleAddressRequired) {
		t.Fatalf("Create() address error = %v", err)
	}
	input = validStoneSampleInput([][]int64{{1, 2, 3, 4}})
	input.Phone = ""
	if _, err := service.Create(context.Background(), input); !errors.Is(err, ErrStoneSamplePhoneRequired) {
		t.Fatalf("Create() phone error = %v", err)
	}
	input = validStoneSampleInput([][]int64{{1, 2, 3, 4}})
	input.ShippingMethod = "DRONE"
	if _, err := service.Create(context.Background(), input); !errors.Is(err, ErrStoneSampleShippingInvalid) {
		t.Fatalf("Create() shipping error = %v", err)
	}
}

func validStoneSampleInput(boxes [][]int64) domain.CreateStoneSampleRequestInput {
	return domain.CreateStoneSampleRequestInput{
		UserID:         "user-1",
		Boxes:          boxes,
		AddressText:    "Test address",
		Phone:          "+989121111111",
		ShippingMethod: string(domain.StoneSampleShippingOperatorCoordination),
	}
}

type stoneSampleFakeRepo struct {
	products map[int64]domain.StoneSampleProduct
}

func newStoneSampleFakeRepo(productCount int64) *stoneSampleFakeRepo {
	products := make(map[int64]domain.StoneSampleProduct)
	for id := int64(1); id <= productCount; id++ {
		products[id] = domain.StoneSampleProduct{
			ID:      id,
			TitleEN: "Stone",
			TitleFA: "Stone",
			TitleAR: "Stone",
			Slug:    "stone",
		}
	}
	return &stoneSampleFakeRepo{products: products}
}

func (r *stoneSampleFakeRepo) ListSampleCategories(context.Context) ([]domain.CatalogCategory, error) {
	return nil, nil
}

func (r *stoneSampleFakeRepo) ListSampleProducts(context.Context, domain.StoneSampleCatalogFilter) ([]domain.StoneSampleProduct, int, error) {
	return nil, 0, nil
}

func (r *stoneSampleFakeRepo) GetSampleProductsByID(_ context.Context, ids []int64) (map[int64]domain.StoneSampleProduct, error) {
	result := make(map[int64]domain.StoneSampleProduct)
	for _, id := range ids {
		if product, ok := r.products[id]; ok {
			result[id] = product
		}
	}
	return result, nil
}

func (r *stoneSampleFakeRepo) ListAddresses(context.Context, string) ([]domain.UserAddress, error) {
	return nil, nil
}

func (r *stoneSampleFakeRepo) GetAddress(context.Context, string, int64) (domain.UserAddress, error) {
	return domain.UserAddress{}, sql.ErrNoRows
}

func (r *stoneSampleFakeRepo) SaveAddress(_ context.Context, address domain.UserAddress) (domain.UserAddress, error) {
	address.ID = 1
	return address, nil
}

func (r *stoneSampleFakeRepo) ListContactNumbers(context.Context, string) ([]domain.UserContactNumber, error) {
	return nil, nil
}

func (r *stoneSampleFakeRepo) GetContactNumber(context.Context, string, int64) (domain.UserContactNumber, error) {
	return domain.UserContactNumber{}, sql.ErrNoRows
}

func (r *stoneSampleFakeRepo) SaveContactNumber(_ context.Context, phone domain.UserContactNumber) (domain.UserContactNumber, error) {
	phone.ID = 1
	return phone, nil
}

func (r *stoneSampleFakeRepo) Create(_ context.Context, req domain.StoneSampleRequest, items []domain.StoneSampleRequestItem) (domain.StoneSampleRequest, error) {
	req.ID = 1
	req.Items = items
	return req, nil
}

func (r *stoneSampleFakeRepo) ListByUser(context.Context, string, int, int) ([]domain.StoneSampleRequest, error) {
	return nil, nil
}

func (r *stoneSampleFakeRepo) GetByUser(context.Context, string, int64) (domain.StoneSampleRequest, error) {
	return domain.StoneSampleRequest{}, nil
}

func (r *stoneSampleFakeRepo) ListAdmin(context.Context, ports.StoneSampleRequestFilter) ([]domain.StoneSampleRequest, error) {
	return nil, nil
}

func (r *stoneSampleFakeRepo) GetAdmin(context.Context, int64) (domain.StoneSampleRequest, error) {
	return domain.StoneSampleRequest{}, nil
}

func (r *stoneSampleFakeRepo) UpdateStatus(context.Context, int64, domain.StoneSampleRequestStatus, *string) error {
	return nil
}

func (r *stoneSampleFakeRepo) AddStatusHistory(context.Context, domain.StoneSampleStatusHistory) error {
	return nil
}
