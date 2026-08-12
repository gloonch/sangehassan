package usecase

import (
	"encoding/json"
	"testing"
)

func TestValidateApplicationSettingRegistry(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   string
		wantErr bool
	}{
		{name: "known currency", key: "default_currency", value: `"IRR"`},
		{name: "known timezone", key: "default_timezone", value: `"Asia/Tehran"`},
		{name: "boolean", key: "inventory_module_enabled", value: `false`},
		{name: "upload lower boundary", key: "max_upload_size_mb", value: `1`},
		{name: "upload upper boundary", key: "max_upload_size_mb", value: `100`},
		{name: "unknown setting", key: "database_password", value: `"secret"`, wantErr: true},
		{name: "invalid currency", key: "default_currency", value: `"BTC"`, wantErr: true},
		{name: "invalid bool type", key: "sms_enabled", value: `"true"`, wantErr: true},
		{name: "upload too large", key: "max_upload_size_mb", value: `101`, wantErr: true},
		{name: "invalid timezone", key: "default_timezone", value: `"Mars/Olympus"`, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateSettingValue(tc.key, json.RawMessage(tc.value))
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateSettingValue() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestValidateSavedViewRejectsSecretsAndNestedObjects(t *testing.T) {
	if err := validateSavedView("orders", "سفارش‌های من", map[string]any{"status": "OPEN", "overdue": true}); err != nil {
		t.Fatalf("expected valid saved view: %v", err)
	}
	if err := validateSavedView("orders", "unsafe", map[string]any{"refresh_token": "secret"}); err == nil {
		t.Fatal("expected token filter to be rejected")
	}
	if err := validateSavedView("orders", "nested", map[string]any{"query": map[string]any{"raw": "sql"}}); err == nil {
		t.Fatal("expected nested arbitrary structure to be rejected")
	}
}

func TestParsePageBounds(t *testing.T) {
	tests := []struct {
		page, size string
		wantPage   int
		wantSize   int
	}{
		{"", "", 1, 25},
		{"0", "0", 1, 25},
		{"2", "100", 2, 100},
		{"3", "1000", 3, 100},
	}
	for _, tc := range tests {
		got := ParsePage(tc.page, tc.size)
		if got.Page != tc.wantPage || got.PageSize != tc.wantSize {
			t.Fatalf("ParsePage(%q,%q) = %+v, want page=%d size=%d", tc.page, tc.size, got, tc.wantPage, tc.wantSize)
		}
	}
}
