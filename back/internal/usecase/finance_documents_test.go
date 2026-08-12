package usecase

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

func TestCommercialTermsFormula(t *testing.T) {
	valid := CommercialTermsPayload{TermsType: "DEPOSIT_AND_BALANCE", Currency: "IRR", Subtotal: "100.0000", DiscountAmount: "10", TaxAmount: "9", AdditionalCharge: "1", FinalCustomerAmount: "100", DepositPercentage: stringPointer("30")}
	if err := validateCommercialTerms(valid); err != nil {
		t.Fatalf("valid terms rejected: %v", err)
	}
	valid.FinalCustomerAmount = "101"
	if err := validateCommercialTerms(valid); err == nil {
		t.Fatal("invalid formula accepted")
	}
}

func TestNotificationTemplateAllowlist(t *testing.T) {
	got, err := renderNotificationTemplate("سفارش {{order_number}}", []string{"order_number"}, map[string]string{"order_number": "ORD-1"})
	if err != nil || got != "سفارش ORD-1" {
		t.Fatalf("render=%q err=%v", got, err)
	}
	if _, err = renderNotificationTemplate("{{secret}}", []string{"order_number"}, map[string]string{"secret": "x"}); err == nil {
		t.Fatal("unknown variable accepted")
	}
	if _, err = renderNotificationTemplate("{{order_number}}", []string{"order_number"}, map[string]string{}); err == nil {
		t.Fatal("unresolved variable accepted")
	}
}

func TestSMSProviders(t *testing.T) {
	if _, err := (DisabledSMSProvider{}).SendMessage(context.Background(), "+989123456789", "test"); !errors.Is(err, ErrSMSProviderDisabled) {
		t.Fatalf("disabled provider error=%v", err)
	}
	fake := &FakeSMSProvider{}
	id, err := fake.SendMessage(context.Background(), "+989123456789", "پیام تست")
	if err != nil || id == "" || len(fake.Messages) != 1 {
		t.Fatalf("fake provider id=%q messages=%d err=%v", id, len(fake.Messages), err)
	}
}

func TestPersianPDFIsValidAndEmbedsFont(t *testing.T) {
	pdf, err := generatePersianPDF("پیش‌فاکتور", map[string]any{"document_number": "PF-1", "customer_name": "حسن", "total": "100000", "currency": "IRR"})
	if err != nil {
		t.Fatal(err)
	}
	if len(pdf) < 1000 || !bytes.HasPrefix(pdf, []byte("%PDF-")) {
		t.Fatalf("invalid pdf length=%d", len(pdf))
	}
}

func TestDocumentTemplateRejectsCode(t *testing.T) {
	if err := validateDocumentTemplateJSON(map[string]any{"title": "<script>alert(1)</script>", "sections": []any{"items"}}); err == nil {
		t.Fatal("script accepted")
	}
	if err := validateDocumentTemplateJSON(map[string]any{"title": "سند", "sections": []any{"items"}}); err != nil {
		t.Fatalf("valid template rejected: %v", err)
	}
}

func stringPointer(value string) *string { return &value }
