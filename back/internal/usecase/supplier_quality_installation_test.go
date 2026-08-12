package usecase

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestValidateRuntimeValueQCCheck(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"pass", `{"result":"PASS"}`, false},
		{"measurement and file", `{"result":"FAIL","note":"لب پریدگی","measuredValue":"1.25","unit":"mm","fileId":"2c3f44b3-188e-4d08-83f1-1fb854700d0e"}`, false},
		{"not applicable", `{"result":"NOT_APPLICABLE"}`, false},
		{"unknown result", `{"result":"MAYBE"}`, true},
		{"non object", `"PASS"`, true},
		{"invalid measurement", `{"result":"FAIL","measuredValue":"x"}`, true},
		{"long note", `{"result":"FAIL","note":"` + string(make([]byte, 2001)) + `"}`, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateRuntimeValue("QC_CHECK", json.RawMessage(test.value), nil, nil, true, "", "")
			if (err != nil) != test.wantErr {
				t.Fatalf("validateRuntimeValue() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestPhaseFivePayloadValidation(t *testing.T) {
	if _, err := validateSupplierPayload(SupplierPayload{Name: "معدن تست", SupplierType: "mine", CountryCode: "ir"}); err != nil {
		t.Fatalf("valid supplier rejected: %v", err)
	}
	if _, err := validateSupplierPayload(SupplierPayload{Name: "", SupplierType: "MINE"}); !errors.Is(err, ErrValidation) {
		t.Fatalf("empty supplier name error = %v", err)
	}
	if _, err := validatePurchasePayload(PurchasePayload{SupplierID: "supplier", StoneName: "تراورتن", Quantity: "10.5000", QuantityUnit: "ton", UnitPrice: "120", CurrencyCode: "irr"}); err != nil {
		t.Fatalf("valid purchase rejected: %v", err)
	}
	if _, err := validatePurchasePayload(PurchasePayload{SupplierID: "supplier", StoneName: "تراورتن", Quantity: "-1", QuantityUnit: "TON", UnitPrice: "0", CurrencyCode: "IRR"}); !errors.Is(err, ErrValidation) {
		t.Fatalf("negative purchase quantity error = %v", err)
	}
	if payload, err := validateInstallationPayload(InstallationPayload{AreaUnit: "square_meter", PlannedArea: phaseFiveStringPointer("35.2500")}); err != nil || payload.AreaUnit != "SQUARE_METER" {
		t.Fatalf("valid installation rejected: payload=%+v err=%v", payload, err)
	}
}

func phaseFiveStringPointer(value string) *string { return &value }
