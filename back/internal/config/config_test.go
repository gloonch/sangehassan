package config

import (
	"strings"
	"testing"
)

func setProductionBaseline(t *testing.T) {
	t.Helper()
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", strings.Repeat("x", 32))
	t.Setenv("DB_PASSWORD", "test-only")
	t.Setenv("COOKIE_SECURE", "true")
	t.Setenv("ALLOWED_ORIGINS", "https://example.test")
	t.Setenv("SMS_PROVIDER", "disabled")
}

func TestProductionConfigurationValidation(t *testing.T) {
	t.Run("valid baseline", func(t *testing.T) {
		setProductionBaseline(t)
		if _, err := Load(); err != nil {
			t.Fatalf("Load() returned error: %v", err)
		}
	})
	t.Run("short JWT", func(t *testing.T) {
		setProductionBaseline(t)
		t.Setenv("JWT_SECRET", "short")
		if _, err := Load(); err == nil {
			t.Fatal("expected short JWT secret to be rejected")
		}
	})
	t.Run("insecure cookie", func(t *testing.T) {
		setProductionBaseline(t)
		t.Setenv("COOKIE_SECURE", "false")
		if _, err := Load(); err == nil {
			t.Fatal("expected insecure production cookies to be rejected")
		}
	})
	t.Run("fake SMS", func(t *testing.T) {
		setProductionBaseline(t)
		t.Setenv("SMS_PROVIDER", "fake")
		if _, err := Load(); err == nil {
			t.Fatal("expected fake production SMS provider to be rejected")
		}
	})
	t.Run("insecure origin", func(t *testing.T) {
		setProductionBaseline(t)
		t.Setenv("ALLOWED_ORIGINS", "http://example.test")
		if _, err := Load(); err == nil {
			t.Fatal("expected a non-HTTPS production origin to be rejected")
		}
	})
}
