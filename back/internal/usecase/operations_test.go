package usecase

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNormalizePhone(t *testing.T) {
	tests := map[string]string{
		"0912 123 4567":      "+989121234567",
		"00989121234567":     "+989121234567",
		"+98 (912) 123-4567": "+989121234567",
		"989121234567":       "+989121234567",
		"123":                "",
		"not-a-phone":        "",
	}
	for input, want := range tests {
		if got := NormalizePhone(input); got != want {
			t.Errorf("NormalizePhone(%q)=%q, want %q", input, got, want)
		}
	}
}

func TestValidateNewPasswordUsesCharacterCountAndBcryptLimit(t *testing.T) {
	if !errors.Is(validateNewPassword("Ab1!xyz"), ErrPasswordTooShort) {
		t.Fatal("seven-character password must be rejected")
	}
	if err := validateNewPassword("Ab1!xyza"); err != nil {
		t.Fatalf("valid password rejected: %v", err)
	}
	if err := validateNewPassword("رمزعبور۱"); err != nil {
		t.Fatalf("eight-character Unicode password rejected: %v", err)
	}
	if !errors.Is(validateNewPassword(strings.Repeat("a", 73)), ErrPasswordTooLong) {
		t.Fatal("password above bcrypt's 72-byte limit must be rejected")
	}
}

func TestRandomDigits(t *testing.T) {
	value := randomDigits(6)
	if len(value) != 6 {
		t.Fatalf("expected 6 digits, got %q", value)
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			t.Fatalf("non-digit in %q", value)
		}
	}
}

func TestAuthorizationUnionsMultipleRoles(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT DISTINCT r.code").WithArgs("user-1").WillReturnRows(
		sqlmock.NewRows([]string{"code", "permission"}).
			AddRow("OPERATOR", "orders.create").AddRow("SALES", "sales.sale_price.view").AddRow("SALES", "orders.create"),
	)
	roles, permissions, err := NewOperationsService(db).authorization(context.Background(), "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(roles) != 2 {
		t.Fatalf("expected two roles, got %v", roles)
	}
	if len(permissions) != 2 {
		t.Fatalf("expected de-duplicated union, got %v", permissions)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSuperAdminGetsFuturePermissions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT DISTINCT r.code").WithArgs("super-1").WillReturnRows(sqlmock.NewRows([]string{"code", "permission"}).AddRow("SUPER_ADMIN", "dashboard.internal.view"))
	mock.ExpectQuery("SELECT code FROM permissions").WillReturnRows(sqlmock.NewRows([]string{"code"}).AddRow("dashboard.internal.view").AddRow("future.permission"))
	_, permissions, err := NewOperationsService(db).authorization(context.Background(), "super-1")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, p := range permissions {
		if p == "future.permission" {
			found = true
		}
	}
	if !found {
		t.Fatalf("future permission missing from %v", permissions)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestActivationCodeCannotBeReusedAfterInvalidAttemptLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	sum := sha256.Sum256([]byte("123456"))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT t.id,t.user_id,t.token_hash,t.attempt_count").WithArgs("+989121234567").WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "token_hash", "attempt_count"}).AddRow(1, "11111111-1111-1111-1111-111111111111", hex.EncodeToString(sum[:]), 0))
	mock.ExpectExec("UPDATE customer_activation_tokens SET attempt_count").WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	err = NewOperationsService(db).ActivateCustomer(context.Background(), "+989121234567", "654321", "ValidPass!1")
	if !errors.Is(err, ErrActivationInvalid) {
		t.Fatalf("expected invalid activation, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestExpiredOrUsedActivationCodeIsRejected(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT t.id,t.user_id,t.token_hash,t.attempt_count").WithArgs("+989121234567").WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()
	err = NewOperationsService(db).ActivateCustomer(context.Background(), "+989121234567", "123456", "ValidPass!1")
	if !errors.Is(err, ErrActivationInvalid) {
		t.Fatalf("expected invalid activation, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestProtectedRolePermissionsCannotBeReplaced(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT is_protected FROM roles").WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"is_protected"}).AddRow(true))
	err = NewOperationsService(db).UpdateRolePermissions(context.Background(), "11111111-1111-1111-1111-111111111111", 1, []string{"users.view"})
	if !errors.Is(err, ErrProtectedRole) {
		t.Fatalf("expected protected role error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
