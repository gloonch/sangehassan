package usecase

import (
	"errors"
	"testing"

	"sangehassan/back/internal/domain"
)

func TestValidateProductOrder(t *testing.T) {
	tests := []struct {
		name    string
		ids     []int64
		wantErr bool
	}{
		{name: "valid", ids: []int64{3, 1, 2}},
		{name: "empty", ids: nil, wantErr: true},
		{name: "duplicate", ids: []int64{1, 2, 1}, wantErr: true},
		{name: "non-positive", ids: []int64{1, 0}, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateProductOrder(test.ids)
			if test.wantErr && !errors.Is(err, domain.ErrInvalidProductOrder) {
				t.Fatalf("expected ErrInvalidProductOrder, got %v", err)
			}
			if !test.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
